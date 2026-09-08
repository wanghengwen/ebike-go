package javacompat

// ClientDTOResponse mirrors com.xyy.dto.ClientDTO on HTTP responses.
// Pointer fields emit JSON null when unset (Java Jackson without NON_NULL).
type ClientDTOResponse struct {
	TraceId       *string  `json:"traceId"`
	TenantId      *string  `json:"tenantId"`
	Platform      *string  `json:"platform"`
	DeviceId      *string  `json:"deviceId"`
	Version       *string  `json:"version"`
	Ip            *string  `json:"ip"`
	Longitude     *float64 `json:"longitude"`
	Latitude      *float64 `json:"latitude"`
	Source        *string  `json:"source"`
	StressTesting bool     `json:"stressTesting"`
}

// ClientDTONullMap returns ClientDTO parent fields as they appear on Java
// response objects when unset (stressTesting defaults to false).
func ClientDTONullMap() map[string]interface{} {
	return map[string]interface{}{
		"traceId":       nil,
		"tenantId":      nil,
		"platform":      nil,
		"deviceId":      nil,
		"version":       nil,
		"ip":            nil,
		"longitude":     nil,
		"latitude":      nil,
		"source":        nil,
		"stressTesting": false,
	}
}
