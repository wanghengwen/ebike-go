// Package jsondiff compares JSON documents for shadow-traffic diffing, tolerating
// Java-vs-Go representation differences (e.g. Long as number vs string) and
// volatile fields that legitimately differ between two reads (timestamps, GPS).
//
// Ported from identity-auth-go and extended for ebike-device-paas with:
//   - IgnoreKeys: drop volatile keys anywhere in the tree before comparing.
//   - FloatTolerance: treat near-equal floats as equal (GCJ-02 conversion drift).
package jsondiff

import (
	"encoding/json"
	"math"
	"reflect"
	"strconv"
)

// Options tunes the semantic comparison.
type Options struct {
	// IgnoreKeys are object keys ignored at any depth (e.g. reportTime, lat).
	IgnoreKeys map[string]bool
	// FloatTolerance is the max absolute difference for two floats to be equal.
	FloatTolerance float64
	// UnorderedKeys are object keys whose array values are compared as multisets
	// (order-insensitive). Use for lists that Java emits in non-deterministic
	// order (e.g. HashMap.values() results) — e.g. "data", "serviceStatistics".
	UnorderedKeys map[string]bool
}

// Equal reports semantic equality with default (strict) options.
func Equal(a, b []byte) bool {
	return EqualOpts(a, b, Options{})
}

// EqualOpts reports semantic equality under the given options.
func EqualOpts(a, b []byte, opts Options) bool {
	var va, vb interface{}
	if err := json.Unmarshal(a, &va); err != nil {
		return string(a) == string(b)
	}
	if err := json.Unmarshal(b, &vb); err != nil {
		return false
	}
	return valuesEqual(va, vb, opts)
}

func valuesEqual(a, b interface{}, opts Options) bool {
	if a == nil && b == nil {
		return true
	}

	switch av := a.(type) {
	case map[string]interface{}:
		bv, ok := b.(map[string]interface{})
		if !ok {
			return false
		}
		ak := keysExcluding(av, opts.IgnoreKeys)
		bk := keysExcluding(bv, opts.IgnoreKeys)
		if len(ak) != len(bk) {
			return false
		}
		for k := range ak {
			if opts.UnorderedKeys[k] {
				if al, aok := av[k].([]interface{}); aok {
					if bl, bok := bv[k].([]interface{}); bok {
						if !unorderedArrayEqual(al, bl, opts) {
							return false
						}
						continue
					}
				}
			}
			if !valuesEqual(av[k], bv[k], opts) {
				return false
			}
		}
		return true
	case []interface{}:
		bv, ok := b.([]interface{})
		if !ok || len(av) != len(bv) {
			return false
		}
		for i := range av {
			if !valuesEqual(av[i], bv[i], opts) {
				return false
			}
		}
		return true
	case float64:
		return numbersEqual(av, b, opts)
	case string:
		return stringsEqual(av, b)
	case bool:
		bv, ok := b.(bool)
		return ok && av == bv
	default:
		return reflect.DeepEqual(a, b)
	}
}

// unorderedArrayEqual compares two arrays as multisets: every element in a must
// have a distinct matching element in b (greedy O(n²), fine for small lists).
func unorderedArrayEqual(a, b []interface{}, opts Options) bool {
	if len(a) != len(b) {
		return false
	}
	used := make([]bool, len(b))
	for _, av := range a {
		found := false
		for j, bv := range b {
			if used[j] {
				continue
			}
			if valuesEqual(av, bv, opts) {
				used[j] = true
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// keysExcluding returns the comparable keys of m: ignored keys are dropped, and
// keys whose value is null are treated as absent. The latter reconciles a real
// Java-vs-Go representation difference: ebike-device-paas serializes with
// Jackson's default ALWAYS inclusion, so a null field is emitted as "field":null,
// whereas the Go DTOs use `omitempty` and drop the field entirely. Both encode
// the same "no value", so null and absent must compare equal. A field that
// carries a concrete value on one side and is null/absent on the other still
// diffs, because that side keeps the key while the other does not (count check).
func keysExcluding(m map[string]interface{}, ignore map[string]bool) map[string]struct{} {
	out := make(map[string]struct{}, len(m))
	for k, v := range m {
		if ignore[k] || v == nil {
			continue
		}
		out[k] = struct{}{}
	}
	return out
}

func numbersEqual(a float64, b interface{}, opts Options) bool {
	switch bv := b.(type) {
	case float64:
		if a == bv {
			return true
		}
		if opts.FloatTolerance > 0 && math.Abs(a-bv) <= opts.FloatTolerance {
			return true
		}
		return isWholeNumber(a) && isWholeNumber(bv) && int64(a) == int64(bv)
	case string:
		return numericStringEqual(a, bv, opts)
	default:
		return false
	}
}

func stringsEqual(a string, b interface{}) bool {
	switch bv := b.(type) {
	case string:
		return a == bv
	case float64:
		return numericStringEqual(bv, a, Options{})
	default:
		return false
	}
}

func numericStringEqual(n float64, s string, opts Options) bool {
	if isWholeNumber(n) {
		i, err := strconv.ParseInt(s, 10, 64)
		return err == nil && int64(n) == i
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return false
	}
	if f == n {
		return true
	}
	return opts.FloatTolerance > 0 && math.Abs(f-n) <= opts.FloatTolerance
}

func isWholeNumber(f float64) bool {
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return false
	}
	return f == math.Trunc(f)
}
