package model

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type IntString int

func (is *IntString) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		if s == "" {
			*is = 0
			return nil
		}
		val, err := strconv.Atoi(s)
		if err != nil {
			return err
		}
		*is = IntString(val)
		return nil
	}
	var i int
	if err := json.Unmarshal(data, &i); err == nil {
		*is = IntString(i)
		return nil
	}
	return fmt.Errorf("cannot unmarshal %s into IntString", string(data))
}

func (is IntString) MarshalJSON() ([]byte, error) {
	return json.Marshal(int(is))
}

type MapCmd struct {
	Api         string      `json:"api"`
	TraceId     string      `json:"traceId" binding:"required"`
	TenantId    string      `json:"tenantId" binding:"required"`
	Source      string      `json:"source"`
	Longitude   string      `json:"longitude"`
	Latitude    string      `json:"latitude"`
	Locations   *[]Location `json:"locations"`
	Address     string      `json:"address"`
	City        string      `json:"city"`
	Origin      string      `json:"origin"`
	Destination string      `json:"destination"`
	Type        int         `json:"type"` // 1: walk, 2: bicycle
}

type Location struct {
	Longitude string `json:"longitude"`
	Latitude  string `json:"latitude"`
}

type AddressCo struct {
	Name   string `json:"name"`
	Area   string `json:"area"`
	Adcode string `json:"adcode"`
}

type Navigate struct {
	Polyline    []Location `json:"polyline"`
	Count       int        `json:"count"`
	Duration    *string    `json:"duration"`
	Distance    *string    `json:"distance"`
	Duration2   *string    `json:"duration2"`
	Distance2   *string    `json:"distance2"`
	Duration3   *string    `json:"duration3"`
	Distance3   *string    `json:"distance3"`
	PolylineApi *string `json:"polylineApi"`
	Api         *string `json:"api"`
	Route       *ARoute `json:"route"`
}

type ARoute struct {
	Origin      string `json:"origin"`
	Destination string `json:"destination"`
	Paths       []Path `json:"paths"`
}

type Path struct {
	Distance string  `json:"distance"`
	Duration *string `json:"duration"`
	Steps    []Step  `json:"steps"`
}

type Step struct {
	Instruction  string    `json:"instruction"`
	Orientation  string    `json:"orientation"`
	RoadName     string    `json:"road_name"`
	StepDistance IntString `json:"step_distance"`
	Polyline     string    `json:"polyline"`
	Cost         *Cost     `json:"cost"`
	Navi         Navi      `json:"navi"`
}

type Cost struct {
	Duration string `json:"duration"`
}

type Navi struct {
	Action          string `json:"action"`
	AssistantAction string `json:"assistant_action"`
	WalkType        string `json:"walk_type"`
}
