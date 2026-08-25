package javalog

import (
	"encoding/json"
	"sort"
	"strings"
)

// ExtractReadOnlyFixtures deduplicates read-only log entries into golden test cases.
// Requires parseable reply envelope (for Go replay tests).
func ExtractReadOnlyFixtures(entries []Entry) []Entry {
	return extractReadOnly(entries, true)
}

// ExtractReadOnlyRequests deduplicates read-only log entries for HTTP replay scripts.
// Only requires a valid request JSON with commandContext.
func ExtractReadOnlyRequests(entries []Entry) []Entry {
	return extractReadOnly(entries, false)
}

func extractReadOnly(entries []Entry, requireEnvelope bool) []Entry {
	seen := map[string]struct{}{}
	out := make([]Entry, 0)
	for _, e := range entries {
		if !IsReadOnlyURL(e.URL) || len(e.Request) == 0 {
			continue
		}
		if !strings.Contains(string(e.Request), `"commandContext"`) {
			continue
		}
		if !json.Valid(e.Request) {
			continue
		}
		if requireEnvelope {
			if _, err := ParseReplyEnvelope(e.Reply); err != nil {
				continue
			}
		}
		key := e.URL + "\x00" + string(e.Request)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, e)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].URL != out[j].URL {
			return out[i].URL < out[j].URL
		}
		if out[i].Source != out[j].Source {
			return out[i].Source < out[j].Source
		}
		return string(out[i].Request) < string(out[j].Request)
	})
	return out
}

// ExtractBindParkingFixtures deduplicates bindParking log entries for parity tests.
func ExtractBindParkingFixtures(entries []Entry) []Entry {
	const url = "/parking/bindParking"
	seen := map[string]struct{}{}
	out := make([]Entry, 0)
	for _, e := range entries {
		if e.URL != url || len(e.Request) == 0 {
			continue
		}
		if !strings.Contains(string(e.Request), `"commandContext"`) {
			continue
		}
		if !json.Valid(e.Request) {
			continue
		}
		if _, err := ParseReplyEnvelope(e.Reply); err != nil {
			continue
		}
		key := string(e.Request)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, e)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Source != out[j].Source {
			return out[i].Source < out[j].Source
		}
		return string(out[i].Request) < string(out[j].Request)
	})
	return out
}

func ReadOnlyURLCounts(fixtures []Entry) map[string]int {
	counts := map[string]int{}
	for _, e := range fixtures {
		counts[e.URL]++
	}
	return counts
}
