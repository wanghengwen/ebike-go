package service

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"ebike-open-paas-go/internal/client"
	"ebike-open-paas-go/internal/contract"
)

// DeviceDetail is the Xiaoan cached deviceInfo payload.
//
// Everything except imei is a pointer with omitempty, because the spec builds
// this response out of whatever the device last reported and the shadow returns
// JSON null for a field no report has filled in. A zero value would be a
// statement of fact we cannot make: batteryLock defaulting to 1 tells a caller
// the battery is locked on a device that has never reported its lock at all, and
// callers act on that.
//
// objectType and lastLoginTimestamp stay absent by construction: neither is in
// the device shadow (/device/paas/device/detail), and the login timestamp only
// exists in ebike-management, which this service deliberately does not query on
// the read path.
type DeviceDetail struct {
	Imei               string   `json:"imei"`
	IsOnline           *int     `json:"isOnline,omitempty"`
	GsmSignal          *int     `json:"gsmSignal,omitempty"`
	Defend             *int     `json:"defend,omitempty"`
	Acc                *int     `json:"acc,omitempty"`
	Lat                *float64 `json:"lat,omitempty"`
	Lng                *float64 `json:"lng,omitempty"`
	BatteryLock        *int     `json:"batteryLock,omitempty"`
	BackWheelLock      *int     `json:"backWheelLock,omitempty"`
	ObjectType         *int     `json:"objectType,omitempty"`
	Voltage            *int     `json:"voltage,omitempty"`
	Version            *int     `json:"version,omitempty"`
	IsMoving           *int     `json:"isMoving,omitempty"`
	Imsi               string   `json:"imsi,omitempty"`
	LastGPSTimestamp   *int64   `json:"lastGPSTimestamp,omitempty"`
	LastLoginTimestamp *int64   `json:"lastLoginTimestamp,omitempty"`
}

// GetDeviceInfo returns the cached device shadow, mapped to Xiaoan fields.
func GetDeviceInfo(tenantID, imei string) (*DeviceDetail, error) {
	env, err := client.PostPaas("/device/paas/device/detail", map[string]interface{}{
		"imei":           imei,
		"commandContext": client.CommandContext(tenantID, ""),
	})
	if err != nil {
		return nil, err
	}
	if !env.Success || len(env.Data) == 0 || string(env.Data) == "null" {
		return nil, fmt.Errorf("%s", contract.ErrDeviceInfoIsNull)
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(env.Data, &raw); err != nil {
		return nil, err
	}
	if AsString(raw["imei"]) == "" && !deviceRegisteredToTenant(tenantID, imei) {
		return nil, fmt.Errorf("%s", contract.ErrDeviceInfoIsNull)
	}

	out := &DeviceDetail{
		Imei:      imei,
		IsOnline:  OptInt(raw, "isOnline"),
		GsmSignal: OptInt(raw, "gsmSignal"),
		Defend:    OptInt(raw, "defend"),
		Acc:       OptInt(raw, "acc"),
		// Xiaoan documents deviceInfo's lat/lng as GCJ02, which is what the
		// shadow's lat/lng already are (wgs84Lat/wgs84Lng are the other pair).
		Lat:           OptFloat(raw, "lat"),
		Lng:           OptFloat(raw, "lng"),
		BatteryLock:   OptInt(raw, "batteryLock"),
		BackWheelLock: OptInt(raw, "backWheelLock"),
		Voltage:       OptInt(raw, "voltage"),
		IsMoving:      OptInt(raw, "isMoving"),
		Imsi:          AsString(raw["imsi"]),
		Version:       parseVersion(raw["version"]),
	}
	if ts := AsInt64(raw["timestamp"], 0); ts > 0 {
		sec := ts
		if ts > 1e12 {
			sec = ts / 1000
		}
		out.LastGPSTimestamp = &sec
	}
	return out, nil
}

// Coordinates returns the cached position, and whether the device has one. A
// device that has never reported a fix has neither.
func (d *DeviceDetail) Coordinates() (lng, lat float64, ok bool) {
	if d == nil || d.Lat == nil || d.Lng == nil {
		return 0, 0, false
	}
	return *d.Lng, *d.Lat, true
}

func normalizeCacheTenant(s string) string {
	return strings.ReplaceAll(strings.TrimSpace(s), "\"", "")
}

// deviceRegisteredToTenant reports whether anvelink registered the IMEI under
// this tenant (Redis hash device_ebike_{imei}.tenantId).
func deviceRegisteredToTenant(tenantID, imei string) bool {
	return DiagnoseDeviceBelongs(tenantID, imei).OK
}

// AsFloat coerces a loosely-typed JSON value to float64, 0 when it cannot.
func AsFloat(v interface{}) float64 {
	switch t := v.(type) {
	case float64:
		return t
	case int:
		return float64(t)
	case int64:
		return float64(t)
	case json.Number:
		f, _ := t.Float64()
		return f
	case string:
		f, _ := strconv.ParseFloat(t, 64)
		return f
	default:
		return 0
	}
}

// parseVersion reads the shadow's firmware version, which the protocol stores as
// a numeric string ("460905"). A value that is not a plain number is omitted
// rather than reported as 0: version 0 reads as a real firmware revision, and a
// caller comparing against it would draw the wrong conclusion.
func parseVersion(v interface{}) *int {
	s := AsString(v)
	if s == "" {
		return nil
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}
	return &n
}
