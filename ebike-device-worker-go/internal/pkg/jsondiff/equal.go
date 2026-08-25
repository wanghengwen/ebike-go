// Package jsondiff compares JSON for shadow diffing, aligned with ebike-fence-go:
// stripNil, map key union, float/geo tolerance (±1e-5 and ≤3), and Long number vs string.
package jsondiff

import (
	"encoding/json"
	"math"
	"reflect"
	"strconv"
)

// Equal reports whether two JSON byte slices are semantically equal for shadow diff.
func Equal(a, b []byte) bool {
	var va, vb interface{}
	if err := json.Unmarshal(a, &va); err != nil {
		return string(a) == string(b)
	}
	if err := json.Unmarshal(b, &vb); err != nil {
		return false
	}
	if aMap, ok := va.(map[string]interface{}); ok {
		if bMap, ok := vb.(map[string]interface{}); ok {
			stripNil(aMap)
			stripNil(bMap)
		}
	}
	return valuesEqual(va, vb)
}

func valuesEqual(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	switch av := a.(type) {
	case map[string]interface{}:
		bv, ok := b.(map[string]interface{})
		if !ok {
			return false
		}
		return mapJSONEqual(av, bv)
	case []interface{}:
		bv, ok := b.([]interface{})
		if !ok || len(av) != len(bv) {
			return false
		}
		for i := range av {
			if !valuesEqual(av[i], bv[i]) {
				return false
			}
		}
		return true
	case float64:
		return numbersEqual(av, b)
	case string:
		return stringsEqual(av, b)
	case bool:
		bv, ok := b.(bool)
		return ok && av == bv
	case json.Number:
		bf, err := av.Float64()
		if err != nil {
			return false
		}
		return numbersEqual(bf, b)
	default:
		return reflect.DeepEqual(a, b)
	}
}

func mapJSONEqual(a, b map[string]interface{}) bool {
	seen := map[string]struct{}{}
	for k := range a {
		seen[k] = struct{}{}
	}
	for k := range b {
		seen[k] = struct{}{}
	}
	for k := range seen {
		av, aOk := a[k]
		bv, bOk := b[k]
		if !aOk {
			av = nil
		}
		if !bOk {
			bv = nil
		}
		if av == nil && bv == nil {
			continue
		}
		if !valuesEqual(av, bv) {
			return false
		}
	}
	return true
}

func numbersEqual(a float64, b interface{}) bool {
	switch bv := b.(type) {
	case float64:
		if floatJSONEqual(a, bv) {
			return true
		}
		return isWholeNumber(a) && isWholeNumber(bv) && int64(a) == int64(bv)
	case string:
		return numericStringEqual(a, bv)
	case json.Number:
		bf, err := bv.Float64()
		return err == nil && numbersEqual(a, bf)
	default:
		return false
	}
}

func stringsEqual(a string, b interface{}) bool {
	switch bv := b.(type) {
	case string:
		return a == bv
	case float64:
		return numericStringEqual(bv, a)
	case json.Number:
		bf, err := bv.Float64()
		return err == nil && numericStringEqual(bf, a)
	default:
		return false
	}
}

func numericStringEqual(n float64, s string) bool {
	if isWholeNumber(n) {
		i, err := strconv.ParseInt(s, 10, 64)
		return err == nil && int64(n) == i
	}
	f, err := strconv.ParseFloat(s, 64)
	return err == nil && f == n
}

func isWholeNumber(f float64) bool {
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return false
	}
	return f == math.Trunc(f)
}

// floatJSONEqual tolerates tiny floating-point drift and small geo distance deltas vs Java JTS.
func floatJSONEqual(a, b float64) bool {
	if a == b {
		return true
	}
	diff := math.Abs(a - b)
	if diff <= 1e-5 {
		return true
	}
	if diff <= 3 {
		return true
	}
	return false
}

func stripNil(m map[string]interface{}) {
	for k, v := range m {
		if v == nil {
			delete(m, k)
		} else if vMap, ok := v.(map[string]interface{}); ok {
			stripNil(vMap)
		} else if vSlice, ok := v.([]interface{}); ok {
			for _, item := range vSlice {
				if itemMap, ok := item.(map[string]interface{}); ok {
					stripNil(itemMap)
				}
			}
		}
	}
}
