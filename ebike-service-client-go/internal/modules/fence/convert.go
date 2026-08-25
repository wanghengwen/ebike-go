package fence

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// ---------------------------------------------------------------------------
// Downstream fence CO (com.xyy.ebike.fence.api.dto.packing.FenceCO subclasses)
// One parse struct covers ServiceAreaCO / ParkingCO / NoParkingCO / BanRidingCO;
// fields that the target ClientO lacks are simply not declared here.
// ---------------------------------------------------------------------------

type fenceCOIn struct {
	Type      json.RawMessage `json:"type"`
	Id        *jLong          `json:"id"`
	Name      json.RawMessage `json:"name"`
	ShapeType json.RawMessage `json:"shapeType"`
	CenterLat json.RawMessage `json:"centerLat"`
	CenterLng json.RawMessage `json:"centerLng"`
	PointList *string         `json:"pointList"`
	Pics      json.RawMessage `json:"pics"`

	// ParkingCO / NoParkingCO / BanRidingCO extras
	MaxParkingNumber       json.RawMessage `json:"maxParkingNumber"`
	ServiceId              *jLong          `json:"serviceId"`
	Tbeacon                json.RawMessage `json:"tbeacon"`
	Directional            json.RawMessage `json:"directional"`
	Direction              json.RawMessage `json:"direction"`
	Rfid                   json.RawMessage `json:"rfid"`
	IzEnable               json.RawMessage `json:"izEnable"`
	CoefficientOfDifficult json.RawMessage `json:"coefficientOfDifficult"`
	BufferDistance         json.RawMessage `json:"bufferDistance"`
	Camera                 json.RawMessage `json:"camera"`
	CarCount               json.RawMessage `json:"carCount"`
	FullCar                json.RawMessage `json:"fullCar"`
	ActivityId             *jLong          `json:"activityId"`
	ExpirationTime         json.RawMessage `json:"expirationTime"`
}

// parsePointList ports FenceClientO.setPointList(String): the downstream
// pointList is a JSON string which fastjson parses into List<List<Double>>.
// Inner arrays are kept as raw tokens to avoid float re-formatting.
// A null/empty string keeps the field null; invalid JSON throws in Java
// (uncaught -> 00001).
func parsePointList(s *string) ([]json.RawMessage, error) {
	if s == nil || *s == "" {
		return nil, nil
	}
	var arr []json.RawMessage
	if err := json.Unmarshal([]byte(*s), &arr); err != nil {
		return nil, fmt.Errorf("JSONException:%s", err.Error())
	}
	out := make([]json.RawMessage, 0, len(arr))
	for _, el := range arr {
		t := bytes.TrimSpace(el)
		if len(t) > 0 && t[0] == '"' {
			// fastjson coerces each element to String first, then parses it.
			var inner string
			if err := json.Unmarshal(t, &inner); err != nil {
				return nil, fmt.Errorf("JSONException:%s", err.Error())
			}
			t = bytes.TrimSpace([]byte(inner))
		}
		var nums []json.Number
		if err := json.Unmarshal(t, &nums); err != nil {
			return nil, fmt.Errorf("JSONException:%s", err.Error())
		}
		out = append(out, json.RawMessage(t))
	}
	return out, nil
}

// ---------------------------------------------------------------------------
// Client-facing ClientO structs (com.xyy.ebike.service.client.client.dto.fence.*)
// Java does not enable NON_NULL, so every field is always emitted.
// ---------------------------------------------------------------------------

type serviceAreaClientO struct {
	Type      json.RawMessage   `json:"type"`
	Id        *jLong            `json:"id"`
	Name      json.RawMessage   `json:"name"`
	ShapeType json.RawMessage   `json:"shapeType"`
	CenterLat json.RawMessage   `json:"centerLat"`
	CenterLng json.RawMessage   `json:"centerLng"`
	PointList []json.RawMessage `json:"pointList"`
	Pics      json.RawMessage   `json:"pics"`
	// ServiceAreaClientO.agentId has no counterpart in the downstream
	// ServiceAreaCO, so BeanUtils never fills it -> always null.
	AgentId *jLong `json:"agentId"`
}

type parkingClientO struct {
	Type                   json.RawMessage   `json:"type"`
	Id                     *jLong            `json:"id"`
	Name                   json.RawMessage   `json:"name"`
	ShapeType              json.RawMessage   `json:"shapeType"`
	CenterLat              json.RawMessage   `json:"centerLat"`
	CenterLng              json.RawMessage   `json:"centerLng"`
	PointList              []json.RawMessage `json:"pointList"`
	Pics                   json.RawMessage   `json:"pics"`
	MaxParkingNumber       json.RawMessage   `json:"maxParkingNumber"`
	ServiceId              *jLong            `json:"serviceId"`
	Tbeacon                json.RawMessage   `json:"tbeacon"`
	Directional            json.RawMessage   `json:"directional"`
	Direction              json.RawMessage   `json:"direction"`
	Rfid                   json.RawMessage   `json:"rfid"`
	IzEnable               json.RawMessage   `json:"izEnable"`
	CoefficientOfDifficult json.RawMessage   `json:"coefficientOfDifficult"`
	BufferDistance         json.RawMessage   `json:"bufferDistance"`
	Camera                 json.RawMessage   `json:"camera"`
	CarCount               json.RawMessage   `json:"carCount"`
	FullCar                json.RawMessage   `json:"fullCar"`
	ActivityId             *jLong            `json:"activityId"`
	ExpirationTime         json.RawMessage   `json:"expirationTime"`
}

type noParkingClientO struct {
	Type      json.RawMessage   `json:"type"`
	Id        *jLong            `json:"id"`
	Name      json.RawMessage   `json:"name"`
	ShapeType json.RawMessage   `json:"shapeType"`
	CenterLat json.RawMessage   `json:"centerLat"`
	CenterLng json.RawMessage   `json:"centerLng"`
	PointList []json.RawMessage `json:"pointList"`
	Pics      json.RawMessage   `json:"pics"`
}

type banRidingClientO struct {
	Type      json.RawMessage   `json:"type"`
	Id        *jLong            `json:"id"`
	Name      json.RawMessage   `json:"name"`
	ShapeType json.RawMessage   `json:"shapeType"`
	CenterLat json.RawMessage   `json:"centerLat"`
	CenterLng json.RawMessage   `json:"centerLng"`
	PointList []json.RawMessage `json:"pointList"`
	Pics      json.RawMessage   `json:"pics"`
	ServiceId *jLong            `json:"serviceId"`
}

// fencesCO mirrors com.xyy...fence.FencesCO; unset lists stay null.
type fencesCO struct {
	ServiceAreas []*serviceAreaClientO `json:"serviceAreas"`
	Parkings     []*parkingClientO     `json:"parkings"`
	NoParkings   []*noParkingClientO   `json:"noParkings"`
	BanRidings   []*banRidingClientO   `json:"banRidings"`
}

// nearParkingClientO mirrors NearParkingClientO (copyProperties from NearParkingCO).
type nearParkingClientO struct {
	IzOutService json.RawMessage `json:"izOutService"`
	ParkingNum   json.RawMessage `json:"parkingNum"`
	Distance     json.RawMessage `json:"distance"`
}

// ---------------------------------------------------------------------------
// Convertor ports (com.xyy...infrastructure.convertor.Convertor.copyToClientO)
// A nil source mirrors Java's NPE inside Convertor (errNPE).
// ---------------------------------------------------------------------------

func toServiceAreaClientO(co *fenceCOIn) (*serviceAreaClientO, error) {
	if co == nil {
		return nil, errNPE
	}
	pl, err := parsePointList(co.PointList)
	if err != nil {
		return nil, err
	}
	return &serviceAreaClientO{
		Type: co.Type, Id: co.Id, Name: co.Name, ShapeType: co.ShapeType,
		CenterLat: co.CenterLat, CenterLng: co.CenterLng, PointList: pl, Pics: co.Pics,
	}, nil
}

func toParkingClientO(co *fenceCOIn) (*parkingClientO, error) {
	if co == nil {
		return nil, errNPE
	}
	pl, err := parsePointList(co.PointList)
	if err != nil {
		return nil, err
	}
	return &parkingClientO{
		Type: co.Type, Id: co.Id, Name: co.Name, ShapeType: co.ShapeType,
		CenterLat: co.CenterLat, CenterLng: co.CenterLng, PointList: pl, Pics: co.Pics,
		MaxParkingNumber: co.MaxParkingNumber, ServiceId: co.ServiceId,
		Tbeacon: co.Tbeacon, Directional: co.Directional, Direction: co.Direction,
		Rfid: co.Rfid, IzEnable: co.IzEnable,
		CoefficientOfDifficult: co.CoefficientOfDifficult, BufferDistance: co.BufferDistance,
		Camera: co.Camera, CarCount: co.CarCount, FullCar: co.FullCar,
		ActivityId: co.ActivityId, ExpirationTime: co.ExpirationTime,
	}, nil
}

func toNoParkingClientO(co *fenceCOIn) (*noParkingClientO, error) {
	if co == nil {
		return nil, errNPE
	}
	pl, err := parsePointList(co.PointList)
	if err != nil {
		return nil, err
	}
	return &noParkingClientO{
		Type: co.Type, Id: co.Id, Name: co.Name, ShapeType: co.ShapeType,
		CenterLat: co.CenterLat, CenterLng: co.CenterLng, PointList: pl, Pics: co.Pics,
	}, nil
}

func toBanRidingClientO(co *fenceCOIn) (*banRidingClientO, error) {
	if co == nil {
		return nil, errNPE
	}
	pl, err := parsePointList(co.PointList)
	if err != nil {
		return nil, err
	}
	return &banRidingClientO{
		Type: co.Type, Id: co.Id, Name: co.Name, ShapeType: co.ShapeType,
		CenterLat: co.CenterLat, CenterLng: co.CenterLng, PointList: pl, Pics: co.Pics,
		ServiceId: co.ServiceId,
	}, nil
}
