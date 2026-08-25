// Package event turns saas_0 records into Xiaoan callback events.
//
// ebike-device-worker publishes one flattened JSON object per device report to
// saas_0: the full openapi decoder output plus appId/appName/deviceDataType/imei
// (see worker's push.buildSendData, which merges rather than projects, so every
// decoded field survives). This package is the only place that knows the shape
// of that record and the Xiaoan wire format.
package event

import "encoding/json"

// Device data types worker assigns via deviceDataTypeByCmd.
const (
	typePing   = "ping"
	typeGPS    = "gps"
	typeBMS    = "bms"
	typeAlarm  = "alarm"
	typeLogin  = "login"
	typeLogout = "logout"
)

// Record is the subset of a saas_0 record this service reads.
//
// Every numeric field is *float64 on purpose. saas_0 carries GPS reports from
// several device protocols (worker maps cmd 3/29/68/81 all to "gps") whose
// decoders do not agree on integer vs fractional encoding for speed, course and
// voltage, and a single `0.0` where we expected an int would fail the whole
// unmarshal and drop the record. float64 represents every value these fields can
// hold exactly — the widest is a uint32 — and Go marshals whole floats back
// without a decimal point, so the outbound payload stays integral.
type Record struct {
	// AppID is the numeric tenantId; json.Number avoids a float round-trip on
	// what is really an identifier.
	AppID          json.Number `json:"appId"`
	DeviceDataType string      `json:"deviceDataType"`
	Imei           string      `json:"imei"`

	// Ping (Bin2). Note the field is gsmSignal here but gsm on GPS reports.
	GsmSignal *float64 `json:"gsmSignal"`

	// GPS (Bin68).
	Sw        *float64 `json:"sw"`
	Gsm       *float64 `json:"gsm"`
	Voltage   *float64 `json:"voltage"`
	Timestamp *float64 `json:"timestamp"`
	Wgs84Lng  *float64 `json:"wgs84Lng"`
	Wgs84Lat  *float64 `json:"wgs84Lat"`
	Speed     *float64 `json:"speed"`
	Course    *float64 `json:"course"`
	Hdop      *float64 `json:"hdop"`
	Satellite *float64 `json:"satellite"`

	// GPS state bits used for derived notifies.
	IsOutofServAera *float64 `json:"isOutofServAera"`
	IsFenceEnable   *float64 `json:"isFenceEnable"`
	Defend          *float64 `json:"defend"`
	// BmsSoc is the TLV-carried state of charge on a GPS report; the standalone
	// BMS report uses Soc instead.
	BmsSoc *float64 `json:"bmsSoc"`

	// BMS (Bin66).
	Sn               *string  `json:"sn"`
	HardVersion      *float64 `json:"hardVersion"`
	SoftVersion      *float64 `json:"softVersion"`
	MosTemperature   *float64 `json:"mosTemperature"`
	MosState         *float64 `json:"mosState"`
	MaxVoltage       *float64 `json:"maxVoltage"`
	MinVoltage       *float64 `json:"minVoltage"`
	Soh              *float64 `json:"soh"`
	Fault            *float64 `json:"fault"`
	Capacity         *float64 `json:"capacity"`
	RemainCapacity   *float64 `json:"remainCapacity"`
	Soc              *float64 `json:"soc"`
	CycleLifeCounter *float64 `json:"cycleLifeCounter"`
	Current          *float64 `json:"current"`

	// Alarm (Bin5) carries only the alarm type.
	Type *float64 `json:"type"`

	// Login/logout. "1" is online, "0" offline.
	LineState *string `json:"lineState"`
}

// TenantID returns the tenant this record belongs to, as the string form used by
// the agent registry and Redis keys.
func (r *Record) TenantID() string {
	return r.AppID.String()
}

// Parse decodes one saas_0 record.
func Parse(value []byte) (*Record, error) {
	var r Record
	if err := json.Unmarshal(value, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// bit reports whether bit n of sw is set. Absent sw reads as all-zero.
func (r *Record) bit(n uint) bool {
	if r.Sw == nil {
		return false
	}
	return (uint64(*r.Sw)>>n)&1 == 1
}

func intOf(v *float64) (int, bool) {
	if v == nil {
		return 0, false
	}
	return int(*v), true
}
