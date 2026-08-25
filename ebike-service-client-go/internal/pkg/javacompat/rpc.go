package javacompat

import (
	"context"
	"encoding/json"

	"ebike-service-client-go/internal/api/dto"
	"ebike-service-client-go/internal/pkg/rpc"
	"ebike-service-client-go/internal/pkg/web"
	"github.com/gin-gonic/gin"
)

// CallData mirrors Java ResultHelper.getResultData(api.xxx(...)):
// transport error -> 00001; downstream failure -> passthrough code/msg with data null.
func CallData(c *gin.Context, service, path string, body interface{}, cmdCtx *dto.CommandContext) (json.RawMessage, bool) {
	result, err := rpc.ForwardCommand(c.Request.Context(), service, path, body, cmdCtx)
	if err != nil {
		web.WriteException(c, err.Error())
		return nil, false
	}
	if !result.Success {
		web.WriteBizResult(c, result)
		return nil, false
	}
	return result.Data, true
}

// CallDataQuiet mirrors ResultHelper.getResultDataWithoutException: errors yield nil.
func CallDataQuiet(ctx context.Context, service, path string, body interface{}, cmdCtx *dto.CommandContext) json.RawMessage {
	result, err := rpc.ForwardCommand(ctx, service, path, body, cmdCtx)
	if err != nil || result == nil || !result.Success || IsNullJSON(result.Data) {
		return nil
	}
	return result.Data
}

// WriteNPE mirrors GlobalExceptionHandler for uncaught NullPointerException.
func WriteNPE(c *gin.Context) {
	web.WriteException(c, "NullPointerException:null")
}
