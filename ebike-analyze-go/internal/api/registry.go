package api

import (
	"net/http"
	"sync"

	"ebike-analyze-go/internal/api/dto"
	"ebike-analyze-go/internal/common/bizerror"
	"ebike-analyze-go/internal/pkg/web"

	"github.com/gin-gonic/gin"
)

var (
	handlerRegistry   = map[string]gin.HandlerFunc{}
	registryOnce      sync.Once
	nativeHandlerKeys map[string]struct{}
)

// RegisterHandler binds a native implementation for an exact POST path.
func RegisterHandler(path string, h gin.HandlerFunc) {
	handlerRegistry[path] = h
}

// RegisterHandlers registers multiple path handlers.
func RegisterHandlers(m map[string]gin.HandlerFunc) {
	for p, h := range m {
		RegisterHandler(p, h)
	}
}

// NativeHandlerPaths returns paths with registered native handlers.
func NativeHandlerPaths() map[string]struct{} {
	registryOnce.Do(buildNativeIndex)
	out := make(map[string]struct{}, len(nativeHandlerKeys))
	for k := range nativeHandlerKeys {
		out[k] = struct{}{}
	}
	return out
}

// IsRegistered reports whether a native handler is registered for path.
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

// Dispatch returns the handler for path, or NotImplementedHandler.
func Dispatch(path string) gin.HandlerFunc {
	if h, ok := handlerRegistry[path]; ok {
		return h
	}
	return NotImplementedHandler(path)
}

// NotImplementedHandler responds when no native Go handler is registered (native_only policy).
func NotImplementedHandler(path string) gin.HandlerFunc {
	msg := "endpoint not implemented: POST " + path
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, dto.NewErrorResult(dto.CodeException, msg))
	}
}

// BindAndDispatch is a helper for generated-style handlers: bind JSON body to obj, call fn.
func BindAndDispatch(c *gin.Context, obj interface{}, msgs map[string]string, fn func(*gin.Context) (interface{}, error)) {
	if !web.BindJSON(c, obj, msgs) {
		return
	}
	res, err := fn(c)
	if err != nil {
		if biz, ok := err.(*bizerror.BizError); ok {
			web.WriteBizError(c, biz.Code(), biz.FormattedMsg())
			return
		}
		web.WriteException(c, bizerror.FormatException(err))
		return
	}
	result := dto.NewSuccessResult(res)
	web.RespondResult(c, &result, nil)
}
