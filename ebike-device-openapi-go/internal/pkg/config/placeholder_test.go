package config

import "testing"

func TestResolvePlaceholders_MQTTBrokerURL(t *testing.T) {
	in := `${MQTT_BROKER:tcp://dev.luopingtech.com:1883}`
	got := resolvePlaceholders(in, nil)
	want := "tcp://dev.luopingtech.com:1883"
	if got != want {
		t.Fatalf("resolvePlaceholders(%q) = %q, want %q", in, got, want)
	}
}

func TestResolveConfigString_LiteralPlaceholder(t *testing.T) {
	got := resolveConfigString("${MQTT_BROKER:tcp://dev.luopingtech.com:1883}")
	want := "tcp://dev.luopingtech.com:1883"
	if got != want {
		t.Fatalf("resolveConfigString = %q, want %q", got, want)
	}
	if got := resolveConfigString("tcp://emqx:1883"); got != "tcp://emqx:1883" {
		t.Fatalf("plain broker mutated: %q", got)
	}
}
