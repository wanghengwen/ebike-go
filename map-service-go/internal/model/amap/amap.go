package amap

import (
	"map-service-go/internal/model"
)

type AMapRouteResponse struct {
	InfoCode string       `json:"infocode"`
	Info     string       `json:"info"`
	Count    string       `json:"count"`
	Route    model.ARoute `json:"route"`
}

type AMapLocationResponse struct {
	InfoCode string     `json:"infocode"`
	Info     string     `json:"info"`
	Geocodes []AGeoCode `json:"geocodes"`
}

type AGeoCode struct {
	Location string `json:"location"`
}

type AMapAddressResponse struct {
	InfoCode  string     `json:"infocode"`
	Info      string     `json:"info"`
	Regeocode AReGeoCode `json:"regeocode"`
}

type AReGeoCode struct {
	FormattedAddress string            `json:"formatted_address"`
	AddressComponent AAddressComponent `json:"addressComponent"`
}

type AAddressComponent struct {
	Province interface{} `json:"province"` // Can be string or empty array in AMap
	City     interface{} `json:"city"`
	District interface{} `json:"district"`
	Adcode   string      `json:"adcode"`
}
