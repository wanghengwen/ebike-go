// Package protocol implements the fixed-length Redis device-info codec, ported
// 1:1 from the Java DeviceProtocol / ProtocolConvertor.
//
// Device info is stored in Redis as one fixed-length string. Each field occupies
// a fixed [offset, offset+length) slice. Updates are applied incrementally via
// Redis SETRANGE. This package must stay byte-compatible with the Java side so
// that Go and Java can read/write the same Redis values during shadow migration.
package protocol

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Kind describes how a field's raw fixed-length string maps to a typed value.
type Kind int

const (
	KString     Kind = iota // raw string (trimmed)
	KInt                    // integer
	KLong                   // 64-bit integer
	KDouble                 // floating point
	KBitmapList             // hex string -> list of set-bit indexes (operationState/alarmState)
	KCsvIntList             // comma-separated ints (carTagTypeIds)
)

// Field is a single entry in the protocol layout.
type Field struct {
	Name   string
	Length int
	Offset int
	Kind   Kind
}

// layout mirrors the Java static initializer in DeviceProtocol exactly (order +
// lengths). Offsets are auto-computed in init(). DO NOT reorder.
var layout = []Field{
	{Name: "imei", Length: 15, Kind: KString},
	{Name: "imsi", Length: 15, Kind: KString},
	{Name: "version", Length: 11, Kind: KLong},
	{Name: "deviceType", Length: 3, Kind: KInt},
	{Name: "carId", Length: 15, Kind: KString},
	{Name: "serviceId", Length: 19, Kind: KLong},
	{Name: "tenantId", Length: 20, Kind: KString},
	{Name: "reportTime", Length: 13, Kind: KLong},
	{Name: "gsmSignal", Length: 3, Kind: KInt},
	{Name: "defend", Length: 1, Kind: KInt},
	{Name: "acc", Length: 1, Kind: KInt},
	{Name: "wgs84Lat", Length: 10, Kind: KDouble},
	{Name: "wgs84Lng", Length: 11, Kind: KDouble},
	{Name: "lat", Length: 10, Kind: KDouble},
	{Name: "lng", Length: 11, Kind: KDouble},
	{Name: "timestamp", Length: 13, Kind: KLong},
	{Name: "speed", Length: 6, Kind: KDouble},
	{Name: "course", Length: 6, Kind: KDouble},
	{Name: "hdop", Length: 7, Kind: KDouble},
	{Name: "batteryLock", Length: 1, Kind: KInt},
	{Name: "batteryConnect", Length: 1, Kind: KInt},
	{Name: "backWheelLock", Length: 1, Kind: KInt},
	{Name: "isAutoLock", Length: 1, Kind: KInt},
	{Name: "voltage", Length: 6, Kind: KInt},
	{Name: "isMoving", Length: 1, Kind: KInt},
	{Name: "totalMiles", Length: 11, Kind: KDouble},
	{Name: "fenceVersion", Length: 11, Kind: KLong},
	{Name: "moveAlarmOn", Length: 1, Kind: KInt},
	{Name: "overSpeedOn", Length: 1, Kind: KInt},
	{Name: "rfidCarId", Length: 16, Kind: KString},
	{Name: "headingAngle", Length: 6, Kind: KInt},
	{Name: "helmetType", Length: 1, Kind: KInt},
	{Name: "helmetLock", Length: 1, Kind: KInt},
	{Name: "helmetReact", Length: 1, Kind: KInt},
	{Name: "isWheelSpan", Length: 1, Kind: KInt},
	{Name: "isFenceEnable", Length: 1, Kind: KInt},
	{Name: "isOutofServAera", Length: 1, Kind: KInt},
	{Name: "noParkId", Length: 19, Kind: KLong},
	{Name: "forParkId", Length: 19, Kind: KLong},
	{Name: "batteryId", Length: 19, Kind: KLong},
	{Name: "maxMileage", Length: 3, Kind: KInt},
	{Name: "restMileage", Length: 3, Kind: KInt},
	{Name: "isPowerExist", Length: 1, Kind: KInt},
	{Name: "isOnline", Length: 1, Kind: KInt},
	{Name: "bmsSN", Length: 20, Kind: KString},
	{Name: "soc", Length: 3, Kind: KInt},
	{Name: "bmsTimeStamp", Length: 13, Kind: KLong},
	{Name: "restBattery", Length: 3, Kind: KInt},
	{Name: "lockTime", Length: 13, Kind: KLong},
	{Name: "unlockTime", Length: 13, Kind: KLong},
	{Name: "ridingState", Length: 2, Kind: KInt},
	{Name: "operationState", Length: 16, Kind: KBitmapList},
	{Name: "alarmState", Length: 16, Kind: KBitmapList},
	{Name: "helmetBind", Length: 1, Kind: KInt},
	{Name: "helmetSOC", Length: 3, Kind: KInt},
	{Name: "helmetState", Length: 1, Kind: KInt},
	{Name: "helmetAngleFault", Length: 1, Kind: KInt},
	{Name: "helmetCapacitanceFault", Length: 1, Kind: KInt},
	{Name: "helmetTinfraredFault", Length: 1, Kind: KInt},
	{Name: "helmetPressureFault", Length: 1, Kind: KInt},
	{Name: "noRideParkId", Length: 19, Kind: KLong},
	{Name: "isDisconnect", Length: 1, Kind: KInt},
	{Name: "etcFault", Length: 6, Kind: KInt},
	{Name: "bmsFault", Length: 6, Kind: KInt},
	{Name: "ecuFault", Length: 6, Kind: KInt},
	{Name: "rfidAck", Length: 2, Kind: KInt},
	{Name: "rfidTimestamp", Length: 13, Kind: KLong},
	{Name: "maintainAreaId", Length: 19, Kind: KLong},
	{Name: "carTagTypeIds", Length: 100, Kind: KCsvIntList},
	{Name: "overload", Length: 1, Kind: KInt},
	{Name: "overloadThreshold", Length: 1, Kind: KInt},
	{Name: "isSupportOverload", Length: 1, Kind: KInt},
	{Name: "isCarBindHelmet", Length: 1, Kind: KInt},
	{Name: "scanLat", Length: 10, Kind: KDouble},
	{Name: "scanLng", Length: 11, Kind: KDouble},
}

// byName indexes the layout for quick lookups.
var byName = map[string]Field{}

// TotalLength is the full fixed-record length (sum of all field lengths).
var TotalLength int

func init() {
	off := 0
	for i := range layout {
		layout[i].Offset = off
		off += layout[i].Length
		byName[layout[i].Name] = layout[i]
	}
	TotalLength = off
}

// Layout returns the ordered field descriptors (read-only).
func Layout() []Field { return layout }

// SetRange is a single Redis SETRANGE op: write Value at Offset.
type SetRange struct {
	Offset int
	Value  string // already padded to the field length
}

// genBlank returns n spaces (Java ProtocolConvertor.genBlankByLength).
func genBlank(n int) string { return strings.Repeat(" ", n) }

// fixedLength right-pads value with spaces up to length (Java getFixedLengthValue).
// If value is already >= length it is returned unchanged.
func fixedLength(value string, length int) string {
	if len(value) >= length {
		return value
	}
	return value + genBlank(length-len(value))
}

// getOneIndexes parses a hex string and returns the indexes of set bits (LSB=0).
// Mirrors Java ProtocolConvertor.getOneIndexes.
func getOneIndexes(s string) []int {
	out := []int{}
	s = strings.TrimSpace(s)
	if s == "" {
		return out
	}
	v, err := strconv.ParseUint(s, 16, 64)
	if err != nil || v == 0 {
		return out
	}
	idx := 0
	for {
		if v&1 == 1 {
			out = append(out, idx)
		}
		idx++
		v >>= 1
		if v == 0 {
			break
		}
	}
	return out
}

// getOneIndexHexString ORs 1<<idx for each index and returns lowercase hex.
// Mirrors Java getOneIndexHexString exactly: Java does `1 << type` where 1 is a
// 32-bit int (so the shift count wraps mod 32 and bit 31 sign-extends into the
// accumulating Long), then renders via Long.toHexString (unsigned). We replicate
// both the `t & 31` wrap and the unsigned hex rendering so bit indices >= 31
// produce identical strings (e.g. bit 31 -> "ffffffff80000000", not "-80000000").
func getOneIndexHexString(indexes []int) string {
	var value int64
	for _, t := range indexes {
		value = int64(int32(1)<<uint(t&31)) | value
	}
	return strconv.FormatUint(uint64(value), 16)
}

// toStringValue converts a typed value to its protocol string form, matching
// Java DeviceProtocol.toStringValue. Returns "" for nil.
//
// Java stores values in DeviceInfoDO with concrete boxed types and calls
// object.toString(): Integer/Long render plain decimal, Double renders via
// Double.toString (always a decimal point, ".0" for whole numbers, "E" notation
// outside [1e-3,1e7)). Go's fmt "%v" on a float64 uses %g and would drop the
// ".0" (e.g. 27.0 -> "27", 0.0 -> "0"), producing different Redis bytes than
// Java for the Double fields (wgs84Lat/Lng, lat/lng, speed, course, hdop,
// totalMiles, scanLat/Lng). We therefore format strictly by the field Kind so
// the byte output matches Java regardless of the concrete Go numeric type.
func toStringValue(f Field, v interface{}) string {
	if v == nil {
		return ""
	}
	switch f.Kind {
	case KCsvIntList:
		ints := toIntSlice(v)
		parts := make([]string, len(ints))
		for i, n := range ints {
			parts[i] = strconv.Itoa(n)
		}
		return strings.Join(parts, ",")
	case KBitmapList:
		return getOneIndexHexString(toIntSlice(v))
	case KInt, KLong:
		// Java Integer/Long.toString -> plain decimal. Accept int/int64 (the
		// normal case) and coerce a float64 (e.g. a JSON-sourced map) to its
		// integral decimal so it never renders as "1.23e+09".
		switch n := v.(type) {
		case int:
			return strconv.FormatInt(int64(n), 10)
		case int32:
			return strconv.FormatInt(int64(n), 10)
		case int64:
			return strconv.FormatInt(n, 10)
		case float64:
			return strconv.FormatInt(int64(n), 10)
		case float32:
			return strconv.FormatInt(int64(n), 10)
		case string:
			return n
		default:
			return fmt.Sprintf("%v", v)
		}
	case KDouble:
		switch n := v.(type) {
		case float64:
			return javaDoubleString(n)
		case float32:
			return javaDoubleString(float64(n))
		case int:
			return javaDoubleString(float64(n))
		case int64:
			return javaDoubleString(float64(n))
		case string:
			return n
		default:
			return fmt.Sprintf("%v", v)
		}
	default: // KString
		if s, ok := v.(string); ok {
			return s
		}
		return fmt.Sprintf("%v", v)
	}
}

// javaDoubleString reproduces java.lang.Double.toString(double) so the
// fixed-length Double fields are byte-identical to the Java encoder. Java uses
// plain decimal notation (with a mandatory fractional digit) when the magnitude
// is in [1e-3, 1e7) and computerized scientific notation ("d.dEx") otherwise.
func javaDoubleString(v float64) string {
	switch {
	case math.IsNaN(v):
		return "NaN"
	case math.IsInf(v, 1):
		return "Infinity"
	case math.IsInf(v, -1):
		return "-Infinity"
	case v == 0:
		if math.Signbit(v) {
			return "-0.0"
		}
		return "0.0"
	}
	abs := math.Abs(v)
	if abs >= 1e-3 && abs < 1e7 {
		s := strconv.FormatFloat(v, 'f', -1, 64)
		if !strings.Contains(s, ".") {
			s += ".0"
		}
		return s
	}
	return javaScientific(strconv.FormatFloat(v, 'e', -1, 64))
}

// javaScientific rewrites Go's shortest 'e' output (e.g. "1e+20", "1.234e-04")
// into Java's Double.toString scientific form ("1.0E20", "1.234E-4"): mantissa
// keeps a mandatory fractional digit and the exponent drops its '+' sign and
// leading zeros.
func javaScientific(s string) string {
	i := strings.IndexAny(s, "eE")
	if i < 0 {
		return s
	}
	mant, exp := s[:i], s[i+1:]
	if !strings.Contains(mant, ".") {
		mant += ".0"
	}
	sign := ""
	if len(exp) > 0 && (exp[0] == '+' || exp[0] == '-') {
		if exp[0] == '-' {
			sign = "-"
		}
		exp = exp[1:]
	}
	exp = strings.TrimLeft(exp, "0")
	if exp == "" {
		exp = "0"
	}
	return mant + "E" + sign + exp
}

// toFieldValue converts a trimmed protocol string back to a typed value,
// matching Java DeviceProtocol.toFieldTypeValue.
func toFieldValue(f Field, s string) (interface{}, error) {
	switch f.Kind {
	case KString:
		return s, nil
	case KInt:
		n, err := strconv.Atoi(s)
		return n, err
	case KLong:
		n, err := strconv.ParseInt(s, 10, 64)
		return n, err
	case KDouble:
		n, err := strconv.ParseFloat(s, 64)
		return n, err
	case KBitmapList:
		return getOneIndexes(s), nil
	case KCsvIntList:
		out := []int{}
		for _, p := range strings.Split(s, ",") {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			n, err := strconv.Atoi(p)
			if err != nil {
				return nil, err
			}
			out = append(out, n)
		}
		return out, nil
	default:
		return s, nil
	}
}

// InitProtocolData builds the full fixed-length string for a device record,
// padding missing/over-long fields with blanks. Mirrors Java initProtocolData.
func InitProtocolData(device map[string]interface{}) string {
	var sb strings.Builder
	for _, f := range layout {
		v, ok := device[f.Name]
		if !ok || v == nil {
			sb.WriteString(genBlank(f.Length))
			continue
		}
		valueStr := toStringValue(f, v)
		if len(valueStr) > f.Length {
			sb.WriteString(genBlank(f.Length))
			continue
		}
		sb.WriteString(fixedLength(valueStr, f.Length))
	}
	return sb.String()
}

// EncodeNotNull returns SETRANGE ops for every non-empty field (Java encodeNotNull).
func EncodeNotNull(device map[string]interface{}) []SetRange {
	out := []SetRange{}
	for _, f := range layout {
		v, ok := device[f.Name]
		if !ok {
			continue
		}
		valueStr := toStringValue(f, v)
		if valueStr == "" {
			continue
		}
		out = append(out, SetRange{Offset: f.Offset, Value: fixedLength(valueStr, f.Length)})
	}
	return out
}

// Encode returns SETRANGE ops only for the named changed fields (Java encode).
func Encode(device map[string]interface{}, changed map[string]bool) []SetRange {
	out := []SetRange{}
	for _, f := range layout {
		if !changed[f.Name] {
			continue
		}
		v := device[f.Name]
		valueStr := toStringValue(f, v)
		if len(valueStr) > f.Length {
			continue
		}
		out = append(out, SetRange{Offset: f.Offset, Value: fixedLength(valueStr, f.Length)})
	}
	return out
}

// Decode parses a fixed-length record string into a typed map. Blank fields are
// omitted (except list fields, which become empty lists). Mirrors Java decode.
func Decode(s string) map[string]interface{} {
	out := map[string]interface{}{}
	if strings.TrimSpace(s) == "" {
		return out
	}
	runes := s
	for _, f := range layout {
		if len(runes) < f.Offset+f.Length {
			break
		}
		raw := strings.TrimSpace(runes[f.Offset : f.Offset+f.Length])
		if raw == "" {
			if f.Kind == KBitmapList || f.Kind == KCsvIntList {
				out[f.Name] = []int{}
			}
			continue
		}
		v, err := toFieldValue(f, raw)
		if err != nil {
			// Java logs and skips on conversion failure.
			continue
		}
		out[f.Name] = v
	}
	return out
}

func toIntSlice(v interface{}) []int {
	switch t := v.(type) {
	case []int:
		return t
	case []interface{}:
		out := make([]int, 0, len(t))
		for _, e := range t {
			switch n := e.(type) {
			case int:
				out = append(out, n)
			case float64:
				out = append(out, int(n))
			}
		}
		return out
	default:
		return nil
	}
}
