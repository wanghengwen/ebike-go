package xiaoan

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"ebike-device-openapi-go/internal/api/dto"
)

// javaCase is one real production request captured from the Java
// ebike-device-openapi service log. hexBody + cmd are the decode input,
// expected is the JSON that the Java decoder produced for that same payload.
type javaCase struct {
	Cmd      int16  `json:"cmd"`
	Imei     string `json:"imei"`
	HexBody  string `json:"hexBody"`
	Expected string `json:"expected"`
}

func loadJavaCases(t *testing.T) []javaCase {
	t.Helper()
	path := filepath.Join("testdata", "java_cases.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var cases []javaCase
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatalf("unmarshal %s: %v", path, err)
	}
	return cases
}

// TestGoDecodeMatchesJavaGolden replays real production request payloads
// (grouped by command, multiple devices per command) captured from the Java
// service log, and asserts the Go decoder produces a byte-equivalent decodeBody
// (compared field-by-field, order-independent).
//
// This is the regression guard for the Go rewrite of ebike-device-openapi:
// any drift from the Java decoder (e.g. a wrong battery/alarm field that could
// cause a false "battery removed" alert) will fail here with a field-level diff.
func TestGoDecodeMatchesJavaGolden(t *testing.T) {
	cases := loadJavaCases(t)
	if len(cases) == 0 {
		t.Fatal("no java cases loaded from testdata/java_cases.json")
	}

	for i, c := range cases {
		c := c
		t.Run(fmt.Sprintf("cmd%d_%s_#%d", c.Cmd, c.Imei, i), func(t *testing.T) {
			body, err := hex.DecodeString(c.HexBody)
			if err != nil {
				t.Fatalf("hex decode %q: %v", c.HexBody, err)
			}

			msg, err := DecodeHex(&dto.MessageHeader{Cmd: c.Cmd}, NewByteBuf(body))
			if err != nil {
				t.Fatalf("Go decode failed while Java succeeded (cmd=%d hex=%s): %v", c.Cmd, c.HexBody, err)
			}

			actualBytes, err := json.Marshal(msg)
			if err != nil {
				t.Fatalf("marshal go result: %v", err)
			}

			actual := map[string]interface{}{}
			if err := json.Unmarshal(actualBytes, &actual); err != nil {
				t.Fatalf("unmarshal go json: %v", err)
			}
			expected := map[string]interface{}{}
			if err := json.Unmarshal([]byte(c.Expected), &expected); err != nil {
				t.Fatalf("unmarshal java expected: %v", err)
			}

			// imei is attached by the outer DeviceReportMessage, not the decodeBody;
			// drop it on both sides so it never affects the comparison.
			delete(actual, "imei")
			delete(expected, "imei")

			// fault is the raw BMS fault bitfield added to both the Java and Go
			// Bin66 decoders so the open-platform callback can pass it through;
			// these golden payloads were captured before the Java change shipped,
			// so the snapshot has no such key. Recapture the cmd=66 cases from a
			// Java build that includes it and delete this exemption — the derived
			// bit fields below still cover the value in the meantime.
			delete(actual, "fault")
			delete(expected, "fault")

			if !reflect.DeepEqual(actual, expected) {
				t.Errorf("decode mismatch for cmd=%d imei=%s hex=%s", c.Cmd, c.Imei, c.HexBody)
				diffMaps(t, expected, actual)
			}
		})
	}
}

// diffMaps reports per-field differences between the Java (expected) and Go
// (actual) decode outputs to make any divergence easy to spot.
func diffMaps(t *testing.T, expected, actual map[string]interface{}) {
	t.Helper()
	seen := map[string]struct{}{}
	for k := range expected {
		seen[k] = struct{}{}
	}
	for k := range actual {
		seen[k] = struct{}{}
	}
	keys := make([]string, 0, len(seen))
	for k := range seen {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		ev, eok := expected[k]
		av, aok := actual[k]
		switch {
		case eok && !aok:
			t.Errorf("  field %q: java=%v (%T)  go=<MISSING>", k, ev, ev)
		case !eok && aok:
			t.Errorf("  field %q: java=<MISSING>  go=%v (%T)", k, av, av)
		case !reflect.DeepEqual(ev, av):
			t.Errorf("  field %q: java=%v (%T)  go=%v (%T)", k, ev, ev, av, av)
		}
	}
}
