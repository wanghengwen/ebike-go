package management

// Ports Java FileController (@RequestMapping("/client/file")):
//   - POST /client/file/upload       -> ebike-management /file/upload        (multipart)
//   - POST /client/file/upload/async -> ebike-management /file/upload/async  (multipart)
//
// Java binds FileUploadDto WITHOUT @RequestBody (Spring form/multipart binding,
// @Valid enabled), then FileGatewayImpl builds a FileUploadCmd whose fields are
// sent as multipart parts by the Feign SpringFormEncoder. Field/part names must
// match FileUploadCmd exactly: file, type, tenantId, traceId, pin, ip, source,
// stressTesting (null Java fields are skipped by the form encoder).
//
// Spring servlet.multipart.max-request-size default is 10MB; the gateway limits
// the incoming body the same way before parsing multipart.

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"ebike-service-client-go/internal/api/dto"
	"ebike-service-client-go/internal/middleware"
	"ebike-service-client-go/internal/pkg/javacompat"
	"ebike-service-client-go/internal/pkg/rpc"
	"ebike-service-client-go/internal/pkg/web"
	"github.com/gin-gonic/gin"
)

// maxFileUploadBytes mirrors Spring Boot servlet.multipart.max-request-size
// default (10 MiB). Single-file max-file-size is often 1 MiB in Spring, but
// the gateway only sees the whole multipart body.
const maxFileUploadBytes = 10 << 20

// fileUploadDTO matches Java FileUploadDto extends ClientDTO
// (file is @NotNull MultipartFile; type is optional).
// Type is *string so an absent form field stays null (skipped downstream),
// matching the Java null-vs-empty distinction.
type fileUploadDTO struct {
	dto.ClientDTO
	File *multipart.FileHeader `form:"file" binding:"required"`
	Type *string               `form:"type"`
}

var fileUploadMsgs = mergeMsgs(map[string]string{
	"file": "must not be null", // @NotNull
})

func fileUpload(c *gin.Context) {
	result, err := forwardFileUpload(c, "/file/upload")
	if result != nil && err == nil && result.Success {
		result.Data = convertString(result.Data)
	}
	web.RespondResult(c, result, err)
}

func fileUploadAsync(c *gin.Context) {
	result, err := forwardFileUpload(c, "/file/upload/async")
	if result != nil && err == nil && result.Success {
		result.Data = convertLongStr(result.Data)
	}
	web.RespondResult(c, result, err)
}

func forwardFileUpload(c *gin.Context, path string) (*dto.Result, error) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxFileUploadBytes)

	var req fileUploadDTO
	if !bindFileUploadForm(c, &req) {
		return nil, nil
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)

	addr, err := rpc.SelectOneHealthyInstance(rpc.ServiceManagement)
	if err != nil {
		return nil, fmt.Errorf("failed to discover service %s: %v", rpc.ServiceManagement, err)
	}
	url := fmt.Sprintf("http://%s%s", addr, path)

	f, err := req.File.Open()
	if err != nil {
		return nil, fmt.Errorf("IOException:%v", err)
	}
	defer f.Close()

	// FileGatewayImpl.getFileUploadCmd: tenantId/traceId/pin/ip/source come from
	// the completed CommandContext, stressTesting is always sent, type from dto.
	// Null (empty) string fields are skipped, like Feign's form encoder does.
	fields := map[string]string{
		"stressTesting": strconv.FormatBool(cmdCtx.StressTesting),
	}
	if req.Type != nil {
		fields["type"] = *req.Type
	}
	if cmdCtx.TenantId != "" {
		fields["tenantId"] = cmdCtx.TenantId
	}
	if cmdCtx.TraceId != "" {
		fields["traceId"] = cmdCtx.TraceId
	}
	if cmdCtx.Pin != "" {
		fields["pin"] = cmdCtx.Pin
	}
	if cmdCtx.Ip != "" {
		fields["ip"] = cmdCtx.Ip
	}
	if cmdCtx.Source != "" {
		fields["source"] = cmdCtx.Source
	}

	r := rpc.RestyUploadClient.R().
		SetContext(c.Request.Context()).
		SetMultipartField("file", req.File.Filename, req.File.Header.Get("Content-Type"), f).
		SetMultipartFormData(fields)

	// Accept-Language passthrough, matching Java's FeignHeaderInterceptor
	if acceptLang := c.GetHeader("Accept-Language"); acceptLang != "" {
		r.SetHeader("Accept-Language", acceptLang)
	}

	resp, err := r.Post(url)
	if err != nil {
		return nil, fmt.Errorf("RPC failed for %s: %v", url, err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("RPC returned HTTP %d for %s", resp.StatusCode(), url)
	}

	var result dto.Result
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to decode response from %s: %v", url, err)
	}
	return &result, nil
}

func bindFileUploadForm(c *gin.Context, req *fileUploadDTO) bool {
	if err := c.ShouldBind(req); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			// Spring MaxUploadSizeExceededException surfaces as a size error.
			web.WriteParamError(c, "Maximum upload size exceeded")
			return false
		}
		if errors.Is(err, io.EOF) || strings.Contains(err.Error(), "request body too large") {
			web.WriteParamError(c, "Maximum upload size exceeded")
			return false
		}
		web.WriteParamError(c, err.Error())
		return false
	}
	if req.File == nil {
		web.WriteParamError(c, "file must not be null")
		return false
	}
	return true
}

func convertString(raw json.RawMessage) json.RawMessage {
	if javacompat.IsNullJSON(raw) {
		return raw
	}
	var s *string
	if json.Unmarshal(raw, &s) != nil {
		return raw
	}
	out, _ := json.Marshal(s)
	return out
}

func convertLongStr(raw json.RawMessage) json.RawMessage {
	if javacompat.IsNullJSON(raw) {
		return raw
	}
	var l *javacompat.LongStr
	if json.Unmarshal(raw, &l) != nil {
		return raw
	}
	out, _ := json.Marshal(l)
	return out
}
