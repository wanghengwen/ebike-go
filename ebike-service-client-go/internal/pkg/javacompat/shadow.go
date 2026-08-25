package javacompat

import (
	"encoding/base64"
	"encoding/json"

	"github.com/gin-gonic/gin"
)

// ShadowJavaResult returns the gateway's base64 Java response body when present.
func ShadowJavaResult(c *gin.Context) ([]byte, bool) {
	if c == nil || c.Request == nil {
		return nil, false
	}
	header := c.Request.Header.Get("X-Shadow-Java-Result")
	if header == "" {
		return nil, false
	}
	b, err := base64.StdEncoding.DecodeString(header)
	return b, err == nil && len(b) > 0
}

// DataFieldBoolFromShadowResult reads data.<field> from the gateway's
// X-Shadow-Java-Result header (base64-encoded Java Result JSON).
// Used to avoid re-invoking write endpoints during SHADOW mirroring.
func DataFieldBoolFromShadowResult(javaResultBase64, field string) (bool, bool) {
	if javaResultBase64 == "" || field == "" {
		return false, false
	}
	javaResultBytes, err := base64.StdEncoding.DecodeString(javaResultBase64)
	if err != nil {
		return false, false
	}
	var javaResult map[string]interface{}
	if err := json.Unmarshal(javaResultBytes, &javaResult); err != nil {
		return false, false
	}
	data, ok := javaResult["data"].(map[string]interface{})
	if !ok {
		return false, false
	}
	v, ok := data[field].(bool)
	return v, ok
}
