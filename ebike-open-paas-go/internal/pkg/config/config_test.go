package config

import (
	"sync"
	"testing"
)

// restoreConfig snapshots the live and baseline configs so a test can mutate them
// without leaking into the next one.
func restoreConfig(t *testing.T) {
	t.Helper()
	prevLive, prevBase := live.Load(), baseline.Load()
	prevIndex := agentIndex.Load()
	t.Cleanup(func() {
		live.Store(prevLive)
		baseline.Store(prevBase)
		agentIndex.Store(prevIndex)
	})
}

const twoAgents = `
open:
  agents:
    - agentId: "a1"
      agentToken: "t1"
      tenantId: "1000"
    - agentId: "a2"
      agentToken: "t2"
      tenantId: "2000"
`

// TestMergeNacosAppConfigReplacesAgents is the credential-revocation case:
// yaml.Unmarshal appends to slices, so without the reset a hot-reload that drops
// an agent would leave the revoked token still working.
func TestMergeNacosAppConfigReplacesAgents(t *testing.T) {
	restoreConfig(t)
	live.Store(defaultConfig())
	snapshotBaseline()

	MergeNacosAppConfig(twoAgents)
	if got := len(GlobalConfig().Open.Agents); got != 2 {
		t.Fatalf("after first merge there are %d agents, want 2", got)
	}

	MergeNacosAppConfig(`
open:
  agents:
    - agentId: "a1"
      agentToken: "t1"
      tenantId: "1000"
`)
	agents := GlobalConfig().Open.Agents
	if len(agents) != 1 || agents[0].AgentID != "a1" {
		t.Fatalf("after revoking a2 the agents are %#v, want only a1", agents)
	}
	if _, ok := Agent("a2"); ok {
		t.Error("Agent(\"a2\") still resolves after the credential was removed from Nacos")
	}
}

// TestMergeNacosAppConfigAllowsRevokingEverything: an empty remote list has to
// win, or there would be no way to turn the platform off.
func TestMergeNacosAppConfigAllowsRevokingEverything(t *testing.T) {
	restoreConfig(t)
	live.Store(defaultConfig())
	snapshotBaseline()
	MergeNacosAppConfig(twoAgents)

	MergeNacosAppConfig("open:\n  agents: []\n")
	if got := GlobalConfig().Open.Agents; len(got) != 0 {
		t.Errorf("agents = %#v, want none", got)
	}
	if _, ok := Agent("a1"); ok {
		t.Error("Agent(\"a1\") still resolves after every credential was revoked")
	}
}

// TestMergeNacosAppConfigDoesNotGrowRegexList: the same append behaviour turns
// each hot-reload into a duplicate entry, which is a slow leak rather than a
// visible failure.
func TestMergeNacosAppConfigDoesNotGrowRegexList(t *testing.T) {
	restoreConfig(t)
	live.Store(defaultConfig())
	snapshotBaseline()

	const remote = `
feign:
  client:
    auth-url-regex:
      - "/ebike/cmd/.*"
`
	MergeNacosAppConfig(remote)
	first := len(GlobalConfig().Feign.Client.AuthURLRegex)
	for i := 0; i < 5; i++ {
		MergeNacosAppConfig(remote)
	}
	if got := len(GlobalConfig().Feign.Client.AuthURLRegex); got != first {
		t.Errorf("auth-url-regex grew from %d to %d across reloads", first, got)
	}
}

// TestMergeNacosAppConfigFallsBackToBaseline: a remote file that only sets
// credentials must not blank the tables the local file provides, since a nil
// VoiceIndexMap silently changes which ringtone a device plays.
func TestMergeNacosAppConfigFallsBackToBaseline(t *testing.T) {
	restoreConfig(t)
	c := defaultConfig()
	c.Open.VoiceIndexMap = map[int]int{1: 7, 2: 8}
	c.Open.Notify.SocSteps = []SocStep{{Percent: 20, Notify: 8}, {Percent: 10, Notify: 9}}
	live.Store(c)
	snapshotBaseline()

	MergeNacosAppConfig(twoAgents)

	g := GlobalConfig()
	if len(g.Open.VoiceIndexMap) != 2 || g.Open.VoiceIndexMap[1] != 7 {
		t.Errorf("VoiceIndexMap = %#v, want the local table kept", g.Open.VoiceIndexMap)
	}
	if len(g.Open.Notify.SocSteps) != 2 {
		t.Errorf("SocSteps = %#v, want the local list kept", g.Open.Notify.SocSteps)
	}
}

// TestMergeNacosAppConfigKeepsPreviousOnParseError: a malformed push must not
// take the running credentials down with it.
func TestMergeNacosAppConfigKeepsPreviousOnParseError(t *testing.T) {
	restoreConfig(t)
	live.Store(defaultConfig())
	snapshotBaseline()
	MergeNacosAppConfig(twoAgents)

	MergeNacosAppConfig("open:\n  agents: [ this is not: valid: yaml\n")

	if got := len(GlobalConfig().Open.Agents); got != 2 {
		t.Errorf("after a malformed push there are %d agents, want the previous 2", got)
	}
}

// TestGlobalConfigSnapshotIsStableDuringReload is why the config is an atomic
// snapshot: a reader holding a *Config must keep seeing a consistent config even
// as a Nacos push replaces it, rather than observing a half-applied one.
func TestGlobalConfigSnapshotIsStableDuringReload(t *testing.T) {
	restoreConfig(t)
	live.Store(defaultConfig())
	snapshotBaseline()
	MergeNacosAppConfig(twoAgents)

	before := GlobalConfig()
	agentsBefore := len(before.Open.Agents)

	MergeNacosAppConfig("open:\n  agents: []\n")

	if got := len(before.Open.Agents); got != agentsBefore {
		t.Errorf("the snapshot a reader was holding changed from %d to %d agents", agentsBefore, got)
	}
	if got := len(GlobalConfig().Open.Agents); got != 0 {
		t.Errorf("the new snapshot has %d agents, want 0", got)
	}
}

// TestConcurrentReloadAndRead is the race this refactor exists to remove: Nacos
// pushes arrive on the SDK's goroutine while every request reads the config.
// Run with -race for it to mean anything.
func TestConcurrentReloadAndRead(t *testing.T) {
	restoreConfig(t)
	live.Store(defaultConfig())
	snapshotBaseline()

	var wg sync.WaitGroup
	stop := make(chan struct{})

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 200; i++ {
			MergeNacosAppConfig(twoAgents)
			MergeNacosKafkaConfig("spring:\n  kafka:\n    bootstrap-servers: kafka-0:9092\n    topic:\n      to-saas-topic: saas_0\n")
		}
		close(stop)
	}()

	for r := 0; r < 4; r++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
				}
				g := GlobalConfig()
				_ = len(g.Open.Agents)
				_ = len(g.Open.VoiceIndexMap)
				_ = len(g.Kafka.Topics)
				_, _ = Agent("a1")
				_ = AgentsByTenant("1000")
			}
		}()
	}
	wg.Wait()
}

// TestAgentsByTenantSkipsDisabled keeps a revoked-but-present credential from
// receiving callbacks.
func TestAgentsByTenantSkipsDisabled(t *testing.T) {
	restoreConfig(t)
	live.Store(defaultConfig())
	snapshotBaseline()
	MergeNacosAppConfig(`
open:
  agents:
    - agentId: "live"
      agentToken: "t1"
      tenantId: "1000"
    - agentId: "off"
      agentToken: "t2"
      tenantId: "1000"
      enabled: false
`)
	got := AgentsByTenant("1000")
	if len(got) != 1 || got[0] != "live" {
		t.Errorf("AgentsByTenant(\"1000\") = %v, want only the enabled agent", got)
	}
}

// TestRebuildAgentIndexSkipsIncompleteEntries: an entry without a tenantId cannot
// be resolved to devices, and one without an agentId cannot be looked up at all.
func TestRebuildAgentIndexSkipsIncompleteEntries(t *testing.T) {
	restoreConfig(t)
	live.Store(defaultConfig())
	snapshotBaseline()
	MergeNacosAppConfig(`
open:
  agents:
    - agentId: ""
      agentToken: "t1"
      tenantId: "1000"
    - agentId: "noTenant"
      agentToken: "t2"
    - agentId: "ok"
      agentToken: "t3"
      tenantId: "1000"
`)
	if _, ok := Agent("noTenant"); ok {
		t.Error("an agent with no tenantId was indexed")
	}
	if _, ok := Agent(""); ok {
		t.Error("an agent with an empty agentId was indexed")
	}
	if _, ok := Agent("ok"); !ok {
		t.Error("the complete entry was not indexed")
	}
}
