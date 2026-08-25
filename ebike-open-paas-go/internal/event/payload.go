package event

// hdopScale converts a centi-unit hdop to the value Xiaoan publishes (1.86).
const hdopScale = 100.0

// hdopMaxPlausible is the largest hdop that can be a real dilution-of-precision
// figure. HDOP is a ratio starting at ~0.5 for a good fix; above 20 the fix is
// meaningless, and GPS receivers cap their reported value well below that.
const hdopMaxPlausible = 20.0

// normalizeHdop returns hdop in the unit Xiaoan publishes.
//
// saas_0 mixes two encodings, because worker labels several commands "gps": the
// Bin68 decoder reads hdop as an unsigned short in hundredths (186 → 1.86) while
// the Bin41 decoder reads a float that is already scaled (1.86), and worker
// rewrites cmd to 3 for both, so the record itself no longer says which decoder
// produced it. Dividing unconditionally turned every Bin41 report into a
// hundredth of its real value.
//
// The magnitude is what separates them: a scaled hdop is below hdopMaxPlausible
// and a centi-unit one is above it (0.5 → 50 at the very best fix), so the two
// ranges do not overlap anywhere a receiver actually reports.
func normalizeHdop(raw float64) float64 {
	if raw > hdopMaxPlausible {
		return raw / hdopScale
	}
	return raw
}

// PingPayload builds the event=1 body.
//
// Bin2 names the signal field gsmSignal while Xiaoan expects gsm; voltage is
// already in volts here (Bin2 reads a single byte), unlike the millivolts on a
// GPS report.
func PingPayload(r *Record) map[string]interface{} {
	out := map[string]interface{}{}
	putNum(out, "gsm", r.GsmSignal)
	putNum(out, "voltage", r.Voltage)
	return out
}

// GPSPayload builds the event=2 body.
//
// Coordinates are the WGS84 pair, which Xiaoan requires: the decoder also emits
// a GCJ02 lng/lat, and sending those would offset every point by a few hundred
// metres. sw is forwarded untouched because it is the raw switch word read off
// the Xiaoan wire — reconstructing it from the decoder's per-bit booleans would
// zero the bits that have no dedicated field.
func GPSPayload(r *Record) map[string]interface{} {
	out := map[string]interface{}{}
	putNum(out, "gsm", r.Gsm)
	putNum(out, "voltage", r.Voltage)
	putNum(out, "timestamp", r.Timestamp)
	putNum(out, "longitude", r.Wgs84Lng)
	putNum(out, "latitude", r.Wgs84Lat)
	putNum(out, "speed", r.Speed)
	putNum(out, "course", r.Course)
	putNum(out, "satellite", r.Satellite)
	putNum(out, "sw", r.Sw)
	if r.Hdop != nil {
		out["hdop"] = normalizeHdop(*r.Hdop)
	}
	return out
}

// BMSPayload builds the event=4 body.
//
// fault is the raw bitfield added to the Bin66 decoders for this callback. The
// decoders' per-bit booleans cannot substitute for it: they replicate a Java bug
// where the discharge-overcurrent flag is overwritten by the charge-overcurrent
// bit and the charge flag is never set, so bits 2 and 3 would both be wrong.
func BMSPayload(r *Record) map[string]interface{} {
	out := map[string]interface{}{}
	if r.Sn != nil {
		out["sn"] = *r.Sn
	}
	putNum(out, "hardVersion", r.HardVersion)
	putNum(out, "softVersion", r.SoftVersion)
	putNum(out, "MOSTemp", r.MosTemperature)
	putNum(out, "MOSState", r.MosState)
	putNum(out, "maxVoltage", r.MaxVoltage)
	putNum(out, "minVoltage", r.MinVoltage)
	putNum(out, "healthState", r.Soh)
	putNum(out, "fault", r.Fault)
	putNum(out, "capacity", r.Capacity)
	putNum(out, "remainCapacity", r.RemainCapacity)
	putNum(out, "soc", r.Soc)
	putNum(out, "cycle", r.CycleLifeCounter)
	putNum(out, "voltage", r.Voltage)
	putNum(out, "current", r.Current)
	putNum(out, "timestamp", r.Timestamp)
	return out
}

// putNum copies a present value under the Xiaoan field name. Absent fields are
// omitted rather than sent as 0, so a third party can tell "device did not
// report this" from "device reported zero".
func putNum(out map[string]interface{}, key string, v *float64) {
	if v != nil {
		out[key] = *v
	}
}
