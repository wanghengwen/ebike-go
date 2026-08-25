package api

import (
	"net/http"
	"sync"

	"ebike-device-worker-go/internal/api/dto"
	"ebike-device-worker-go/internal/pkg/apilog"
	"ebike-device-worker-go/internal/pkg/web"

	"github.com/gin-gonic/gin"
)

var (
	handlerRegistry   = map[string]gin.HandlerFunc{}
	registryOnce      sync.Once
	nativeHandlerKeys map[string]struct{}
)

func RegisterHandler(path string, h gin.HandlerFunc) {
	handlerRegistry[path] = h
}

func RegisterHandlers(m map[string]gin.HandlerFunc) {
	for p, h := range m {
		RegisterHandler(p, h)
	}
}

func NativeHandlerPaths() map[string]struct{} {
	registryOnce.Do(buildNativeIndex)
	out := make(map[string]struct{}, len(nativeHandlerKeys))
	for k := range nativeHandlerKeys {
		out[k] = struct{}{}
	}
	return out
}

func IsRegistered(path string) bool {
	_, ok := handlerRegistry[path]
	return ok
}

func buildNativeIndex() {
	nativeHandlerKeys = make(map[string]struct{}, len(handlerRegistry))
	for p := range handlerRegistry {
		nativeHandlerKeys[p] = struct{}{}
	}
}

func Dispatch(path string) gin.HandlerFunc {
	if h, ok := handlerRegistry[path]; ok {
		return h
	}
	return NotImplementedHandler(path)
}

func NotImplementedHandler(path string) gin.HandlerFunc {
	msg := "endpoint not implemented: POST " + path
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, dto.NewErrorResult(dto.CodeException, msg))
	}
}

func BindAndServe(c *gin.Context, obj interface{}, msgs map[string]string, fn func() (interface{}, error)) {
	if !web.BindJSON(c, obj, msgs) {
		return
	}
	apilog.Object(c.Request.URL.Path, obj)
	res, err := fn()
	if err != nil {
		web.HandleServiceError(c, err)
		return
	}
	web.RespondSuccess(c, res)
}
