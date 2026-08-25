package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"map-service-go/internal/config"
	"map-service-go/internal/model"
	"map-service-go/internal/model/amap"

	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

type AMapGatewayImpl struct {
	client *resty.Client
}

func NewAMapGateway() *AMapGatewayImpl {
	return &AMapGatewayImpl{
		client: NewMapHTTPClient(),
	}
}

func (g *AMapGatewayImpl) GetNavigate(ctx context.Context, origin, destination string, navType int, tenantId string) (*model.Navigate, error) {
	url := g.getRouteUrl(origin, destination, navType, tenantId)

	var response amap.AMapRouteResponse
	_, err := g.client.R().
		SetContext(ctx).
		SetResult(&response).
		Get(url)

	if err != nil {
		return nil, fmt.Errorf("API_INVOKE_FAILED: %v", err)
	}

	b, _ := json.Marshal(response)
	logrus.WithContext(ctx).Infof("getNavigate amap :%s", string(b))

	if response.InfoCode == "10000" {
		return g.transARouteNavigate(response.Route, response.Count), nil
	}

	return nil, fmt.Errorf("API_INVOKE_FAILED: %s", response.Info)
}

func (g *AMapGatewayImpl) transARouteNavigate(aRoute model.ARoute, count string) *model.Navigate {
	c, _ := strconv.Atoi(count)
	route := aRoute
	return &model.Navigate{
		Count: c,
		Route: &route,
	}
}

func (g *AMapGatewayImpl) GetLocationByAddress(ctx context.Context, address, city, tenantId string) (*model.Location, error) {
	url := g.getLocationUrl(address, city, tenantId)

	var response amap.AMapLocationResponse
	_, err := g.client.R().
		SetContext(ctx).
		SetResult(&response).
		Get(url)

	if err != nil {
		return nil, fmt.Errorf("API_INVOKE_FAILED: %v", err)
	}

	b, _ := json.Marshal(response)
	logrus.WithContext(ctx).Infof("getLocationByAddress amap :%s", string(b))

	if response.InfoCode == "10000" && len(response.Geocodes) > 0 {
		return g.transLocation(&response.Geocodes[0]), nil
	}

	return nil, fmt.Errorf("API_INVOKE_FAILED: %s", response.Info)
}

func (g *AMapGatewayImpl) GetAddressByLocation(ctx context.Context, longitude, latitude, tenantId string) (*model.AddressCo, error) {
	url := g.getAddressUrl(longitude, latitude, tenantId)

	var response amap.AMapAddressResponse
	_, err := g.client.R().
		SetContext(ctx).
		SetResult(&response).
		Get(url)

	if err != nil {
		return nil, fmt.Errorf("API_INVOKE_FAILED: %v", err)
	}

	b, _ := json.Marshal(response)
	logrus.WithContext(ctx).Infof("getAddressByLocation amap :%s", string(b))

	if response.InfoCode == "10000" {
		return g.transAddress(&response.Regeocode), nil
	}

	return nil, fmt.Errorf("API_INVOKE_FAILED: %s", response.Info)
}

func (g *AMapGatewayImpl) GetBatchAddressByLocation(ctx context.Context, locations []model.Location, tenantId string) ([]model.AddressCo, error) {
	return nil, nil
}

func (g *AMapGatewayImpl) getRouteUrl(origin, destination string, navType int, tenantId string) string {
	aMapConf := config.GetConfig().Xyy.AMap
	baseUrl := aMapConf.WalkingUrl
	if navType != 1 {
		baseUrl = aMapConf.BicyclingUrl
	}

	// CoordinateUtils.keep6Bit2Str truncates origin/destination to 6 decimal places before sending to AMap.
	return fmt.Sprintf("%s?key=%s&isindoor=0&alternative_route=3&output=json&show_fields=cost,navi,polyline&origin=%s&destination=%s",
		baseUrl, g.getKey(tenantId, aMapConf), keep6Bit2Str(origin), keep6Bit2Str(destination))
}

func (g *AMapGatewayImpl) getAddressUrl(lng, lat, tenantId string) string {
	aMapConf := config.GetConfig().Xyy.AMap
	return fmt.Sprintf("%s?key=%s&location=%s,%s&radius=1000&extensions=base&batch=false",
		aMapConf.RegeoUrl, g.getKey(tenantId, aMapConf), lng, lat)
}

func (g *AMapGatewayImpl) getLocationUrl(address, city, tenantId string) string {
	aMapConf := config.GetConfig().Xyy.AMap
	url := fmt.Sprintf("%s?key=%s&address=%s", aMapConf.GeoUrl, g.getKey(tenantId, aMapConf), address)
	if city != "" {
		url += "&city=" + city
	}
	return url
}

func (g *AMapGatewayImpl) transAddress(regeocode *amap.AReGeoCode) *model.AddressCo {
	comp := regeocode.AddressComponent

	area := ""
	if p, ok := comp.Province.(string); ok {
		area += p
	}
	if c, ok := comp.City.(string); ok {
		area += c
	}
	if d, ok := comp.District.(string); ok {
		area += d
	}

	return &model.AddressCo{
		Name:   regeocode.FormattedAddress,
		Area:   area,
		Adcode: comp.Adcode,
	}
}

func (g *AMapGatewayImpl) transLocation(geocode *amap.AGeoCode) *model.Location {
	coor := strings.Split(geocode.Location, ",")
	if len(coor) >= 2 {
		return &model.Location{
			Longitude: coor[0],
			Latitude:  coor[1],
		}
	}
	return &model.Location{}
}

func (g *AMapGatewayImpl) getKey(tenantId string, mapConf config.MapUrl) string {
	keys := mapConf.Keys
	if keys == nil {
		keys = mapConf.Key
	}
	if keys == nil {
		return ""
	}
	var val interface{}
	if tenantId == "" {
		val = keys["default"]
	} else {
		var ok bool
		val, ok = keys[tenantId]
		if !ok || val == nil {
			val = keys["default"]
		}
	}
	if val == nil {
		return ""
	}
	if s, ok := val.(string); ok {
		return s
	}
	return ""
}
