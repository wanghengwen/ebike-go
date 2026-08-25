package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"map-service-go/internal/config"
	"map-service-go/internal/model"
	"map-service-go/internal/model/tmap"
	"map-service-go/internal/pkg/polyline"

	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

type TMapGatewayImpl struct {
	client *resty.Client
}

func NewTMapGateway() *TMapGatewayImpl {
	return &TMapGatewayImpl{
		client: NewMapHTTPClient(),
	}
}

func (g *TMapGatewayImpl) GetNavigate(ctx context.Context, origin, destination string, navType int, tenantId string) (*model.Navigate, error) {
	url := g.getRouteUrl(origin, destination, navType, tenantId)

	var response tmap.TMapRouteResponse
	_, err := g.client.R().
		SetContext(ctx).
		SetResult(&response).
		Get(url)

	if err != nil {
		return nil, fmt.Errorf("API_INVOKE_FAILED: %v", err)
	}

	if response.Status == 0 {
		b, _ := json.Marshal(response)
		logrus.WithContext(ctx).Infof("tmap getNavigate: %s", string(b))
		return g.transRouteNavigate(&response, origin, destination), nil
	}

	return nil, fmt.Errorf("API_INVOKE_FAILED: %s", response.Message)
}

func (g *TMapGatewayImpl) GetLocationByAddress(ctx context.Context, address, city, tenantId string) (*model.Location, error) {
	url := g.getLocationsUrl(address, city, tenantId)

	var response tmap.TMapLocationResponse
	_, err := g.client.R().
		SetContext(ctx).
		SetResult(&response).
		Get(url)

	if err != nil {
		return nil, fmt.Errorf("API_INVOKE_FAILED: %v", err)
	}

	if response.Status == 0 {
		b, _ := json.Marshal(response)
		logrus.WithContext(ctx).Infof("tmap getLocationByAddress: %s", string(b))
		return g.transLocation(&response.Result), nil
	}

	return nil, fmt.Errorf("API_INVOKE_FAILED: %s", response.Message)
}

func (g *TMapGatewayImpl) GetAddressByLocation(ctx context.Context, longitude, latitude, tenantId string) (*model.AddressCo, error) {
	url := g.getAddressUrl(longitude, latitude, tenantId)

	var response tmap.TMapAddressResponse
	_, err := g.client.R().
		SetContext(ctx).
		SetResult(&response).
		Get(url)

	if err != nil {
		return nil, fmt.Errorf("API_INVOKE_FAILED: %v", err)
	}

	if response.Status == 0 {
		b, _ := json.Marshal(response)
		logrus.WithContext(ctx).Infof("tmap getAddressByLocation: %s", string(b))
		return g.transAddress(&response.Result), nil
	}

	return nil, fmt.Errorf("API_INVOKE_FAILED: %s", response.Message)
}

func (g *TMapGatewayImpl) GetBatchAddressByLocation(ctx context.Context, locations []model.Location, tenantId string) ([]model.AddressCo, error) {
	return nil, nil // Return null as in Java
}

func (g *TMapGatewayImpl) getLocationsUrl(address, city, tenantId string) string {
	tMapConf := config.GetConfig().Xyy.TMap
	// FIXME: 这是一个已知BUG的还原。Java原版中TMap错误地使用了AMap的Key（大概率是因为代码复制遗留问题）。
	// 为了确保重构后的Go版本与Java版本在行为和报错输出上100%黑盒对齐，这里故意保持了使用AMap Key的行为。
	// 后续如果业务确认需要修复此BUG，应将此处的 aMapConf 替换为 tMapConf，并使用 g.getKey(tenantId, tMapConf)。
	aMapConf := config.GetConfig().Xyy.AMap
	url := fmt.Sprintf("%s?address=%s&key=%s", tMapConf.GeoUrl, address, g.getKey(tenantId, aMapConf))
	if city != "" {
		url += "&region=" + city
	}
	return url
}

func (g *TMapGatewayImpl) getRouteUrl(origin, destination string, navType int, tenantId string) string {
	origin = polyline.ExchangeLngLat(origin)
	destination = polyline.ExchangeLngLat(destination)

	tMapConf := config.GetConfig().Xyy.TMap
	baseUrl := tMapConf.WalkingUrl
	if navType != 1 {
		baseUrl = tMapConf.BicyclingUrl
	}

	// FIXME: 这是一个已知BUG的还原。Java原版中TMap错误地使用了AMap的Key（大概率是因为代码复制遗留问题）。
	// 为了确保重构后的Go版本与Java版本在行为和报错输出上100%黑盒对齐，这里故意保持了使用AMap Key的行为。
	// 后续如果业务确认需要修复此BUG，应将此处的 aMapConf 替换为 tMapConf，并使用 g.getKey(tenantId, tMapConf)。
	aMapConf := config.GetConfig().Xyy.AMap
	return fmt.Sprintf("%s?key=%s&from=%s&to=%s&output=json", baseUrl, g.getKey(tenantId, aMapConf), origin, destination)
}

func (g *TMapGatewayImpl) getAddressUrl(lng, lat, tenantId string) string {
	tMapConf := config.GetConfig().Xyy.TMap
	// FIXME: 这是一个已知BUG的还原。Java原版中TMap错误地使用了AMap的Key（大概率是因为代码复制遗留问题）。
	// 为了确保重构后的Go版本与Java版本在行为和报错输出上100%黑盒对齐，这里故意保持了使用AMap Key的行为。
	// 后续如果业务确认需要修复此BUG，应将此处的 aMapConf 替换为 tMapConf，并使用 g.getKey(tenantId, tMapConf)。
	aMapConf := config.GetConfig().Xyy.AMap
	return fmt.Sprintf("%s?location=%s,%s&key=%s&get_poi=0&poi_options=address_format=short;radius=3000;policy=1&output=json",
		tMapConf.RegeoUrl, lat, lng, g.getKey(tenantId, aMapConf))
}

func (g *TMapGatewayImpl) transRouteNavigate(response *tmap.TMapRouteResponse, origin, destination string) *model.Navigate {
	navigate := &model.Navigate{
		Count: len(response.Result.Routes),
	}

	aRoute := model.ARoute{
		Origin:      origin,
		Destination: destination,
	}

	var paths []model.Path
	for _, tRoute := range response.Result.Routes {
		dur := strconv.Itoa(tRoute.Duration * 60)
		path := model.Path{
			Distance: strconv.Itoa(tRoute.Distance),
			Duration: &dur,
		}

		uncompressed := polyline.UnCompressPolyline(tRoute.Polyline)
		navigate.Polyline = polyline.PolylineLocation(uncompressed)

		var aSteps []model.Step
		for _, tStep := range tRoute.Steps {
			aStep := model.Step{
				Instruction:  tStep.Instruction,
				StepDistance: model.IntString(tStep.Distance),
				RoadName:     tStep.RoadName,
				Orientation:  tStep.DirDesc,
				Navi: model.Navi{
					Action:          tStep.ActDesc,
					AssistantAction: "",
					WalkType:        "0",
				},
			}

			if len(tStep.PolylineIdx) == 2 {
				start := tStep.PolylineIdx[0]
				end := tStep.PolylineIdx[1] + 1
				if end <= len(uncompressed) {
					sub := uncompressed[start:end]
					aStep.Polyline = polyline.PolylineString(sub)
				}
			}

			aSteps = append(aSteps, aStep)
		}
		path.Steps = aSteps
		paths = append(paths, path)
	}
	aRoute.Paths = paths
	route := aRoute
	navigate.Route = &route

	return navigate
}

func (g *TMapGatewayImpl) transAddress(result *tmap.AddressResult) *model.AddressCo {
	province := result.AdInfo.Province
	city := result.AdInfo.City
	district := result.AdInfo.District

	area := ""
	if province == city {
		area = province + district
	} else {
		area = province + city + district
	}

	address := result.Address
	recommend := result.FormattedAddresses.Recommend
	if strings.HasPrefix(recommend, district) {
		recommend = strings.Replace(recommend, district, "", 1)
	}

	return &model.AddressCo{
		Name:   address + recommend,
		Area:   area,
		Adcode: result.AdInfo.Adcode,
	}
}

func (g *TMapGatewayImpl) transLocation(result *tmap.LocationResult) *model.Location {
	return &model.Location{
		Latitude:  result.Location.Lat.String(),
		Longitude: result.Location.Lng.String(),
	}
}

func (g *TMapGatewayImpl) getKey(tenantId string, mapConf config.MapUrl) string {
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
