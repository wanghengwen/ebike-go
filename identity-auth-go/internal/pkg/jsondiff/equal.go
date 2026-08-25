// Package jsondiff compares JSON documents for shadow-traffic diffing, tolerating
// Java-vs-Go representation differences (e.g. Long as number vs string).
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
	return ValuesEqual(va, vb)
}

// ValuesEqual reports whether two parsed JSON values are semantically equal.
func ValuesEqual(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}

	switch av := a.(type) {
	case map[string]interface{}:
		bv, ok := b.(map[string]interface{})
		if !ok || len(av) != len(bv) {
			return false
		}
		for k, v := range av {
			if !ValuesEqual(v, bv[k]) {
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
			if !ValuesEqual(av[i], bv[i]) {
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
	default:
		return reflect.DeepEqual(a, b)
	}
}

func numbersEqual(a float64, b interface{}) bool {
	switch bv := b.(type) {
	case float64:
		if a == bv {
			return true
		}
		return isWholeNumber(a) && isWholeNumber(bv) && int64(a) == int64(bv)
	case string:
		return numericStringEqual(a, bv)
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
