// Package web provides request-binding and response helpers shared by all
// API modules, mirroring the Java service's GlobalExceptionHandler behavior:
//
//   - validation failure  -> {success:false, code:"00004", msg:"<field> <message>", data:null}
//   - unreadable JSON     -> {success:false, code:"00002", msg:"http message not readable", data:null}
//   - uncaught exception  -> {success:false, code:"00001", msg:"...", data:null}
//
// All error responses use HTTP 200, exactly like the Java @RestControllerAdvice.
package web

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"
	"unicode"

	"ebike-service-client-go/internal/api/dto"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	jsoniter "github.com/json-iterator/go"
	"github.com/json-iterator/go/extra"
)

func init() {
	// Enable fuzzy decoding to tolerate String to Int64 conversions (Jackson compatibility)
	extra.RegisterFuzzyDecoders()
	binding.JSON = &jsoniterBinding{}
}

type jsoniterBinding struct{}

func (jsoniterBinding) Name() string {
	return "json"
}

func (jsoniterBinding) Bind(req *http.Request, obj interface{}) error {
	if req == nil || req.Body == nil {
		return errors.New("invalid request")
	}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return err
	}
	req.Body = io.NopCloser(bytes.NewBuffer(body))
	if err := jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(body, obj); err != nil {
		return err
	}
	if binding.Validator != nil {
		return binding.Validator.ValidateStruct(obj)
	}
	return nil
}

func (jsoniterBinding) BindBody(body []byte, obj interface{}) error {
	if err := jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(body, obj); err != nil {
		return err
	}
	if binding.Validator != nil {
		return binding.Validator.ValidateStruct(obj)
	}
	return nil
}

// BindJSON binds a JSON request body (Java: @RequestBody @Validated).
// msgs maps JSON field names to the hibernate-validator English default message
// of the Java annotation on that field, e.g.
//
//	@NotBlank -> "must not be blank"
//	@NotEmpty -> "must not be empty"
//	@NotNull  -> "must not be null"
//
// On failure it writes the error response and returns false.
func BindJSON(c *gin.Context, obj interface{}, msgs map[string]string) bool {
	if err := c.ShouldBindJSON(obj); err != nil {
		writeBindError(c, err, msgs, true)
		return false
	}
	return true
}

// BindForm binds form/query parameters (Java: controller arg without @RequestBody,
// bound by Spring as form data -> BindException on validation failure).
func BindForm(c *gin.Context, obj interface{}, msgs map[string]string) bool {
	if err := c.ShouldBind(obj); err != nil {
		writeBindError(c, err, msgs, false)
		return false
	}
	return true
}

// BindQuery binds query parameters (Java: GET handler args).
func BindQuery(c *gin.Context, obj interface{}, msgs map[string]string) bool {
	if err := c.ShouldBindQuery(obj); err != nil {
		writeBindError(c, err, msgs, false)
		return false
	}
	return true
}

func writeBindError(c *gin.Context, err error, msgs map[string]string, jsonBody bool) {
	var vErrs validator.ValidationErrors
	if errors.As(err, &vErrs) && len(vErrs) > 0 {
		field := LowerFirst(vErrs[0].Field())
		msg, ok := msgs[field]
		if !ok {
			msg = "must not be null"
		}
		log.Printf("[BINDING_ERROR] Path: %s | Validation failed on field '%s': %v", c.Request.URL.Path, field, err)
		WriteParamError(c, field+" "+msg)
		return
	}
	if jsonBody {
		// Java: HttpMessageNotReadableException -> PARAM_EXCEPTION (00002)
		log.Printf("[BINDING_ERROR] Path: %s | JSON parse error (http message not readable): %v", c.Request.URL.Path, err)
		c.JSON(http.StatusOK, dto.NewErrorResult(dto.CodeParamException, "http message not readable"))
		return
	}
	// Java: BindException (type mismatch on form/query) -> paramError (00004)
	log.Printf("[BINDING_ERROR] Path: %s | Form/Query BindException: %v", c.Request.URL.Path, err)
	WriteParamError(c, err.Error())
}

// WriteParamError mirrors Java's ResultHelper.paramError (code "00004").
func WriteParamError(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, dto.NewErrorResult(dto.CodeIllegalArgument, msg))
}

// WriteException mirrors Java's GlobalExceptionHandler.handleException (code "00001").
func WriteException(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, dto.NewErrorResult(dto.CodeException, msg))
}

// WriteBizResult mirrors Java BizException handling: preserve code/msg, data null.
func WriteBizResult(c *gin.Context, result *dto.Result) {
	if result == nil {
		WriteException(c, "NullPointerException:null")
		return
	}
	c.JSON(http.StatusOK, dto.Result{Success: false, Code: result.Code, Msg: result.Msg})
}

// RespondResult writes the passthrough downstream Result, or the Java-equivalent
// exception response when the RPC call itself failed.
func RespondResult(c *gin.Context, result *dto.Result, err error) {
	if err != nil {
		WriteException(c, err.Error())
		return
	}
	if result != nil && !result.Success {
		WriteBizResult(c, result)
		return
	}
	c.JSON(http.StatusOK, result)
}

// NotBlank enforces Java's @NotBlank semantics (rejects whitespace-only values,
// which gin's "required" tag does not). Writes the error response and returns
// false when the value is blank.
func NotBlank(c *gin.Context, field, value string) bool {
	if strings.TrimSpace(value) == "" {
		WriteParamError(c, field+" must not be blank")
		return false
	}
	return true
}

// RequireJSONFieldNotNull rejects absent/null/empty-string JSON values for a
// top-level field before struct binding. jsoniter fuzzy decoders coerce "" to
// numeric zero and bypass gin's required tag on *int64, which diverges from
// Java @NotNull Long validation (null or "" -> "不能为null").
func RequireJSONFieldNotNull(c *gin.Context, field, nullMsg string) bool {
	if c.Request == nil || c.Request.Body == nil {
		WriteParamError(c, field+" "+nullMsg)
		return false
	}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		WriteParamError(c, "http message not readable")
		return false
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
	probe := map[string]json.RawMessage{}
	if err := json.Unmarshal(body, &probe); err != nil {
		log.Printf("[BINDING_ERROR] Path: %s | JSON probe error: %v", c.Request.URL.Path, err)
		c.JSON(http.StatusOK, dto.NewErrorResult(dto.CodeParamException, "http message not readable"))
		return false
	}
	raw, ok := probe[field]
	if !ok || isJSONNullOrEmpty(raw) {
		WriteParamError(c, field+" "+nullMsg)
		return false
	}
	return true
}

func isJSONNullOrEmpty(raw json.RawMessage) bool {
	s := strings.TrimSpace(string(raw))
	return s == "" || s == "null" || s == `""`
}

// LowerFirst converts a Go field name (CarId) to its JSON name (carId).
func LowerFirst(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}
