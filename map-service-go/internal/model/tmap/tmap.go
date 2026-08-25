package tmap

import "encoding/json"

type TMapRouteResponse struct {
	Status  int             `json:"status"`
	Message string          `json:"message"`
	Result  TMapRouteResult `json:"result"`
}

type TMapRouteResult struct {
	Routes []TRoute `json:"routes"`
}

type TRoute struct {
	Distance int       `json:"distance"`
	Duration int       `json:"duration"` // in minutes in tmap
	Polyline []float64 `json:"polyline"`
	Steps    []TStep   `json:"steps"`
}

type TStep struct {
	Instruction string `json:"instruction"`
	Distance    int    `json:"distance"`
	RoadName    string `json:"road_name"`
	DirDesc     string `json:"dir_desc"`
	ActDesc     string `json:"act_desc"`
	PolylineIdx []int  `json:"polyline_idx"`
}

type TMapLocationResponse struct {
	Status  int            `json:"status"`
	Message string         `json:"message"`
	Result  LocationResult `json:"result"`
}

type LocationResult struct {
	Location TLocation `json:"location"`
}

type TLocation struct {
	Lat json.Number `json:"lat"`
	Lng json.Number `json:"lng"`
}

type TMapAddressResponse struct {
	Status  int           `json:"status"`
	Message string        `json:"message"`
	Result  AddressResult `json:"result"`
}

type AddressResult struct {
	Address            string             `json:"address"`
	FormattedAddresses FormattedAddresses `json:"formatted_addresses"`
	AdInfo             AdInfo             `json:"ad_info"`
}

type FormattedAddresses struct {
	Recommend string `json:"recommend"`
}

type AdInfo struct {
	Province string `json:"province"`
	City     string `json:"city"`
	District string `json:"district"`
	Adcode   string `json:"adcode"`
}
