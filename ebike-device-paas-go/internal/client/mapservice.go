package client

import (
	"encoding/json"
	"log"

	"ebike-device-paas-go/internal/pkg/config"
)

// GeoPoint mirrors Java GeoQryDto ({longitude, latitude}) — the per-location
// request element for the batch reverse-geocode.
type GeoPoint struct {
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

// MapBatchRegeo mirrors MultiApiRpcImpl.aMapBatch: POST map-service /map/batchRegeo
// with {tenantId, api, traceId, locations:[{longitude, latitude}]} and returns the
// RegeoDo.name list in request order. The api parameter defaults to "aMap"
// (Java ${xyy.mapServiceConfig.api:aMap}). On any failure it returns nil so the
// caller blanks the addresses, matching Java's try/catch around the RPC.
func MapBatchRegeo(tenantID, traceID string, locations []GeoPoint) []string {
	base := config.GlobalConfig.Xyy.MapServiceConfig.URL
	if base == "" {
		return nil
	}
	api := config.GlobalConfig.Xyy.MapServiceConfig.API
	if api == "" {
		api = "aMap"
	}
	// FeignClientConfig adds a "secret" header to every map-service request.
	var extra map[string]string
	if s := config.GlobalConfig.Xyy.MapServiceConfig.Secret; s != "" {
		extra = map[string]string{"secret": s}
	}
	env, err := doJSON("POST", base+"/map/batchRegeo", map[string]interface{}{
		"tenantId":  tenantID,
		"api":       api,
		"traceId":   traceID,
		"locations": locations,
	}, "", extra)
	if err != nil || !env.Success || len(env.Data) == 0 {
		log.Printf("[client] map/batchRegeo failed: %v", err)
		return nil
	}
	var regeo []struct {
		Name string `json:"name"`
	}
	if json.Unmarshal(env.Data, &regeo) != nil {
		return nil
	}
	out := make([]string, len(regeo))
	for i, r := range regeo {
		out[i] = r.Name
	}
	return out
}
