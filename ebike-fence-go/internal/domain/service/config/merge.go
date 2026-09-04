package configsvc

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"ebike-fence-go/internal/pkg/timefmt"
)

// mergeJSON copies present fields from src onto dst.
//
// Unlike a plain json.Marshal/Unmarshal round-trip, this understands the
// sql.Null* column types used by config models: a JSON scalar such as
// `true` / `7` / `"foo"` becomes {Valid:true, …} instead of being silently
// dropped. Null JSON values and missing keys are left untouched (Java
// ConvertorHelper.copyProperties skip-null semantics).
//
// src/dst may be structs or pointers; embedded structs are walked.
func mergeJSON(dst, src interface{}) {
	raw, err := json.Marshal(src)
	if err != nil || len(raw) == 0 || string(raw) == "null" {
		return
	}
	var values map[string]interface{}
	if err := json.Unmarshal(raw, &values); err != nil {
		return
	}
	applyJSONMap(reflect.ValueOf(dst), values)
}

func applyJSONMap(dst reflect.Value, values map[string]interface{}) {
	if dst.Kind() == reflect.Ptr {
		if dst.IsNil() {
			return
		}
		dst = dst.Elem()
	}
	if dst.Kind() != reflect.Struct {
		return
	}
	t := dst.Type()
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		if sf.PkgPath != "" { // unexported
			continue
		}
		fv := dst.Field(i)
		if sf.Anonymous && fv.Kind() == reflect.Struct {
			applyJSONMap(fv, values)
			continue
		}
		if sf.Anonymous && fv.Kind() == reflect.Ptr && fv.Type().Elem().Kind() == reflect.Struct {
			if fv.IsNil() {
				continue
			}
			applyJSONMap(fv, values)
			continue
		}
		key := jsonFieldName(sf)
		if key == "-" || key == "" {
			continue
		}
		val, ok := lookupJSON(values, key)
		if !ok || val == nil {
			continue
		}
		_ = setFieldFromJSON(fv, val)
	}
}

func jsonFieldName(sf reflect.StructField) string {
	tag := sf.Tag.Get("json")
	if tag == "" {
		return sf.Name
	}
	name, _, _ := strings.Cut(tag, ",")
	return name
}

func lookupJSON(values map[string]interface{}, key string) (interface{}, bool) {
	if v, ok := values[key]; ok {
		return v, true
	}
	// Models have no json tags; match PascalCase field names case-insensitively
	// against camelCase JSON keys from the DTO.
	lower := strings.ToLower(key)
	for k, v := range values {
		if strings.ToLower(k) == lower {
			return v, true
		}
	}
	return nil, false
}

func setFieldFromJSON(field reflect.Value, raw interface{}) error {
	if !field.CanSet() {
		return nil
	}
	switch field.Type() {
	case reflect.TypeOf(sql.NullBool{}):
		b, ok := asBool(raw)
		if !ok {
			return nil
		}
		field.Set(reflect.ValueOf(sql.NullBool{Bool: b, Valid: true}))
		return nil
	case reflect.TypeOf(sql.NullInt32{}):
		n, ok := asInt64(raw)
		if !ok {
			return nil
		}
		field.Set(reflect.ValueOf(sql.NullInt32{Int32: int32(n), Valid: true}))
		return nil
	case reflect.TypeOf(sql.NullInt64{}):
		n, ok := asInt64(raw)
		if !ok {
			return nil
		}
		field.Set(reflect.ValueOf(sql.NullInt64{Int64: n, Valid: true}))
		return nil
	case reflect.TypeOf(sql.NullFloat64{}):
		f, ok := asFloat64(raw)
		if !ok {
			return nil
		}
		field.Set(reflect.ValueOf(sql.NullFloat64{Float64: f, Valid: true}))
		return nil
	case reflect.TypeOf(sql.NullString{}):
		s, ok := asString(raw)
		if !ok {
			return nil
		}
		field.Set(reflect.ValueOf(sql.NullString{String: s, Valid: true}))
		return nil
	case reflect.TypeOf(sql.NullTime{}):
		t, ok := asTime(raw)
		if !ok {
			return nil
		}
		field.Set(reflect.ValueOf(sql.NullTime{Time: t, Valid: true}))
		return nil
	}

	switch field.Kind() {
	case reflect.Ptr:
		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}
		return setFieldFromJSON(field.Elem(), raw)
	case reflect.Bool:
		b, ok := asBool(raw)
		if ok {
			field.SetBool(b)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, ok := asInt64(raw)
		if ok {
			field.SetInt(n)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, ok := asInt64(raw)
		if ok && n >= 0 {
			field.SetUint(uint64(n))
		}
	case reflect.Float32, reflect.Float64:
		f, ok := asFloat64(raw)
		if ok {
			field.SetFloat(f)
		}
	case reflect.String:
		s, ok := asString(raw)
		if ok {
			field.SetString(s)
		}
	case reflect.Slice:
		// DTO↔DTO copies (e.g. []int userTicketPhotoWays) keep working via JSON.
		b, err := json.Marshal(raw)
		if err != nil {
			return err
		}
		tmp := reflect.New(field.Type())
		if err := json.Unmarshal(b, tmp.Interface()); err != nil {
			return err
		}
		field.Set(tmp.Elem())
	case reflect.Struct:
		b, err := json.Marshal(raw)
		if err != nil {
			return err
		}
		tmp := reflect.New(field.Type())
		if err := json.Unmarshal(b, tmp.Interface()); err != nil {
			return err
		}
		field.Set(tmp.Elem())
	}
	return nil
}

func asBool(v interface{}) (bool, bool) {
	switch x := v.(type) {
	case bool:
		return x, true
	case float64:
		return x != 0, true
	case string:
		b, err := strconv.ParseBool(x)
		return b, err == nil
	case json.Number:
		n, err := x.Int64()
		return n != 0, err == nil
	default:
		return false, false
	}
}

func asInt64(v interface{}) (int64, bool) {
	switch x := v.(type) {
	case float64:
		return int64(x), true
	case int:
		return int64(x), true
	case int64:
		return x, true
	case json.Number:
		n, err := x.Int64()
		return n, err == nil
	case string:
		n, err := strconv.ParseInt(x, 10, 64)
		return n, err == nil
	case bool:
		if x {
			return 1, true
		}
		return 0, true
	default:
		return 0, false
	}
}

func asFloat64(v interface{}) (float64, bool) {
	switch x := v.(type) {
	case float64:
		return x, true
	case int:
		return float64(x), true
	case int64:
		return float64(x), true
	case json.Number:
		f, err := x.Float64()
		return f, err == nil
	case string:
		f, err := strconv.ParseFloat(x, 64)
		return f, err == nil
	default:
		return 0, false
	}
}

func asString(v interface{}) (string, bool) {
	switch x := v.(type) {
	case string:
		return x, true
	case float64:
		// Prefer integer formatting when the JSON number is whole.
		if x == float64(int64(x)) {
			return strconv.FormatInt(int64(x), 10), true
		}
		return strconv.FormatFloat(x, 'f', -1, 64), true
	case bool:
		return strconv.FormatBool(x), true
	case json.Number:
		return x.String(), true
	case []interface{}:
		// ConfigBaseItem.userTicketPhotoWays is List<Integer> in the Cmd but a
		// comma-separated string column on the DO — mirror Java StringUtils.join.
		parts := make([]string, 0, len(x))
		for _, item := range x {
			n, ok := asInt64(item)
			if !ok {
				return "", false
			}
			parts = append(parts, strconv.FormatInt(n, 10))
		}
		return strings.Join(parts, ","), true
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return "", false
		}
		s := string(b)
		if len(s) >= 2 && s[0] == '"' {
			return s[1 : len(s)-1], true
		}
		return s, true
	}
}

func asTime(v interface{}) (time.Time, bool) {
	switch x := v.(type) {
	case string:
		t, err := timefmt.ParseJavaLocalString(x)
		return t, err == nil && !t.IsZero()
	case float64:
		// Unix millis are uncommon here; reject to avoid accidental epoch writes.
		return time.Time{}, false
	default:
		s := fmt.Sprint(x)
		t, err := timefmt.ParseJavaLocalString(s)
		return t, err == nil && !t.IsZero()
	}
}
