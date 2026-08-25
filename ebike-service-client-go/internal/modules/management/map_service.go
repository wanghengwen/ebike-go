package management

// Port of infrastructure/gateway/map/MapServiceRpcImpl (RestTemplate call to the
// external map-service, configured via xyy.mapServiceConfig url/secret) and of
// GaodeMapServiceImpl.selectNavigationPath / getPolyline.

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"ebike-service-client-go/internal/api/dto"
	"ebike-service-client-go/internal/pkg/config"
	"ebike-service-client-go/internal/pkg/rpc"
	"ebike-service-client-go/internal/pkg/web"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

// navigateCmd matches Java infrastructure.gateway.map.dto.NavigateCmd.
// All fields are serialized (Java sends nulls too; type stays a number).
type navigateCmd struct {
	TraceId     string `json:"traceId"`
	TenantId    string `json:"tenantId"`
	Api         string `json:"api"`
	Origin      string `json:"origin"`
	Destination string `json:"destination"`
	Type        *int   `json:"type"`
}

// navigationDO / routeDO / pathDO / stepDO match the map-service response DTOs
// (parsed leniently like fastjson does).
type navigationDO struct {
	Count *int `json:"count"`
	Route *routeDO `json:"route"`
}

type routeDO struct {
	Origin      *string `json:"origin"`
	Destination *string `json:"destination"`
	Paths       []*pathDO   `json:"paths"`
}

type pathDO struct {
	Distance *string `json:"distance"`
	Duration *string `json:"duration"`
	Steps    []*stepDO   `json:"steps"`
}

type stepDO struct {
	Instruction  *string `json:"instruction"`
	Orientation  *string `json:"orientation"`
	RoadName     *string `json:"road_name"`
	StepDistance *int    `json:"step_distance"`
	Polyline     *string `json:"polyline"`
}

// rpcResultNavigation matches common/dto/RpcResult<NavigationDO>.
type rpcResultNavigation struct {
	Success bool          `json:"success"`
	Code    *string       `json:"code"`
	Msg     *string       `json:"msg"`
	Data    *navigationDO `json:"data"`
}

// locationVO matches clientobject.LocationVO {lng, lat}.
type locationVO struct {
	Lng float64 `json:"lng"`
	Lat float64 `json:"lat"`
}

// selectNavigationPath ports GaodeMapServiceImpl.selectNavigationPath +
// MapServiceRpcImpl.navigate. On failure it writes the exact Java error
// response and returns ok=false:
//   - transport/HTTP/empty response -> code "00017" "查询导航路径响应为null"
//     (RestTemplateUtils.postJson swallows errors and returns null)
//   - success=false                 -> code "00017" "查询导航路径响失败:<msg>"
//   - missing route/paths           -> code "00018" "查询地图路径规划信息为空"
func selectNavigationPath(c *gin.Context, origin, destination, mapType string, pathType *int, tenantId string) (*pathDO, bool) {
	mapCfg := config.GlobalConfig.Xyy.MapServiceConfig
	baseURL, urlSource := resolveMapServiceBaseURL()
	if baseURL == "" {
		log.Printf("[WARN] gaode navigate: map-service url unavailable (config empty, nacos discovery failed)")
		writeBizError(c, "00017", "查询导航路径响应为null")
		return nil, false
	}
	url := baseURL + "/map/mapNavigate"

	cmd := navigateCmd{
		TraceId:     randomUUID(),
		TenantId:    tenantId,
		Api:         mapType,
		Type:        pathType,
		Origin:      origin,
		Destination: destination,
	}

	r := rpc.RestyClient.R().
		SetContext(c.Request.Context()).
		SetHeader("Content-Type", "application/json").
		SetBody(&cmd)
	// Java RestTemplateUtils skips empty header values
	if mapCfg.Secret != "" {
		r.SetHeader("secret", mapCfg.Secret)
	}

	resp, err := r.Post(url)
	if err != nil || resp.IsError() || strings.TrimSpace(string(resp.Body())) == "" {
		status := 0
		bodyLen := 0
		if resp != nil {
			status = resp.StatusCode()
			bodyLen = len(resp.Body())
		}
		log.Printf("[WARN] gaode navigate: map-service POST failed urlSource=%s status=%d bodyLen=%d err=%v secretConfigured=%t",
			urlSource, status, bodyLen, err, mapCfg.Secret != "")
		writeBizError(c, "00017", "查询导航路径响应为null")
		return nil, false
	}

	var rpcResult rpcResultNavigation
	if err := json.Unmarshal(resp.Body(), &rpcResult); err != nil {
		// fastjson parse failure -> uncaught JSONException -> code "00001"
		web.WriteException(c, fmt.Sprintf("JSONException:%v", err))
		return nil, false
	}

	if !rpcResult.Success {
		// Java string concat with a null msg yields "null"
		msg := "null"
		if rpcResult.Msg != nil {
			msg = string(*rpcResult.Msg)
		}
		writeBizError(c, "00017", "查询导航路径响失败:"+msg)
		return nil, false
	}

	legal := rpcResult.Data != nil && rpcResult.Data.Route != nil && len(rpcResult.Data.Route.Paths) > 0
	if !legal {
		// Assert.isTrue(legal, MsgCodeEnum.NOT_NULL, "查询地图路径规划信息为空")
		writeBizError(c, "00018", "查询地图路径规划信息为空")
		return nil, false
	}
	return rpcResult.Data.Route.Paths[0], true
}

// resolveMapServiceBaseURL returns the map-service base URL from config, or via Nacos
// service discovery when xyy.mapServiceConfig.url is empty (Java uses the explicit URL).
func resolveMapServiceBaseURL() (string, string) {
	if u := strings.TrimSpace(config.GlobalConfig.Xyy.MapServiceConfig.Url); u != "" {
		return strings.TrimRight(u, "/"), "config"
	}
	if rpc.NamingClient == nil {
		return "", ""
	}
	addr, err := rpc.SelectOneHealthyInstance(rpc.ServiceMapService)
	if err != nil {
		log.Printf("[WARN] map-service discovery failed: %v", err)
		return "", ""
	}
	return "http://" + addr, "nacos-discovery"
}

// parsePolylineLatFirst ports GaodeMapServiceImpl.getPolyline (lat parsed before
// lng, mirroring potential error ordering). Blank input yields an empty list.
func parsePolylineLatFirst(c *gin.Context, polyline string) ([]locationVO, bool) {
	points := []locationVO{}
	if strings.TrimSpace(polyline) == "" {
		return points, true
	}
	for _, pointStr := range javaSplit(polyline, ";") {
		parts := javaSplit(pointStr, ",")
		if len(parts) < 2 {
			// Java: points[1] -> ArrayIndexOutOfBoundsException -> code "00001"
			web.WriteException(c, fmt.Sprintf("ArrayIndexOutOfBoundsException:Index 1 out of bounds for length %d", len(parts)))
			return nil, false
		}
		lat, ok := parseDoubleOrFail(c, parts[1])
		if !ok {
			return nil, false
		}
		lng, ok := parseDoubleOrFail(c, parts[0])
		if !ok {
			return nil, false
		}
		points = append(points, locationVO{Lng: lng, Lat: lat})
	}
	return points, true
}

// writeBizError mirrors GlobalExceptionHandler.handleBizException:
// {success:false, code:<msgCode>, msg:<override message>, data:null}, HTTP 200.
func writeBizError(c *gin.Context, code, msg string) {
	c.JSON(200, dto.NewErrorResult(code, msg))
}

// ---------------------------------------------------------------------------
// gaode.navigate.url (Java @Value("${gaode.navigate.url}") from Nacos config).
// The shared Go config struct does not expose this key, and this module must
// not modify files outside its own directory, so the value is read lazily from
// the GAODE_NAVIGATE_URL env var or the conf/application.yml used by main.go.
// ---------------------------------------------------------------------------

var (
	gaodeNavigateURLOnce sync.Once
	gaodeNavigateURL     string
)

// getGaodeNavigateURL returns gaode.navigate.url from Nacos config (v1 navigate).
func getGaodeNavigateURL() string {
	if u := strings.TrimSpace(config.GlobalConfig.Gaode.Navigate.Url); u != "" {
		return u
	}
	gaodeNavigateURLOnce.Do(func() {
		if env := os.Getenv("GAODE_NAVIGATE_URL"); env != "" {
			gaodeNavigateURL = env
			return
		}
		data, err := os.ReadFile("conf/application.yml")
		if err != nil {
			return
		}
		var cfg struct {
			Gaode struct {
				Navigate struct {
					Url string `yaml:"url"`
				} `yaml:"navigate"`
			} `yaml:"gaode"`
		}
		if yaml.Unmarshal(data, &cfg) == nil {
			gaodeNavigateURL = cfg.Gaode.Navigate.Url
		}
	})
	return gaodeNavigateURL
}
