package web

import (
	"bytes"
	"errors"
	"io"
	"log"
	"net/http"
	"unicode"

	"ebike-device-worker-go/internal/api/dto"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	jsoniter "github.com/json-iterator/go"
	"github.com/json-iterator/go/extra"
)

func init() {
	extra.RegisterFuzzyDecoders()
	binding.JSON = &jsoniterBinding{}
}

type jsoniterBinding struct{}

func (jsoniterBinding) Name() string { return "json" }

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

func BindJSON(c *gin.Context, obj interface{}, msgs map[string]string) bool {
	if err := c.ShouldBindJSON(obj); err != nil {
		writeBindError(c, err, msgs, true)
		return false
	}
	return true
}

func writeBindError(c *gin.Context, err error, msgs map[string]string, jsonBody bool) {
	var vErrs validator.ValidationErrors
	if errors.As(err, &vErrs) && len(vErrs) > 0 {
		field := lowerFirst(vErrs[0].Field())
		msg, ok := msgs[field]
		if !ok {
			msg = "must not be null"
		}
		log.Printf("[BINDING_ERROR] Path: %s | field '%s': %v", c.Request.URL.Path, field, err)
		WriteParamError(c, field+" "+msg)
		return
	}
	if jsonBody {
		log.Printf("[BINDING_ERROR] Path: %s | JSON parse error: %v", c.Request.URL.Path, err)
		c.JSON(http.StatusOK, dto.NewErrorResult(dto.CodeParamException, "http message not readable"))
		return
	}
	WriteParamError(c, err.Error())
}

func WriteParamError(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, dto.NewErrorResult(dto.CodeIllegalArgument, msg))
}

func WriteException(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, dto.NewErrorResult(dto.CodeException, msg))
}

func WriteBizError(c *gin.Context, code, msg string) {
	c.JSON(http.StatusOK, dto.NewErrorResult(code, msg))
}

func RespondSuccess(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, dto.NewSuccessResult(data))
}

func HandleServiceError(c *gin.Context, err error) {
	if err == nil {
		return
	}
	if biz, ok := err.(*BizError); ok {
		WriteBizError(c, biz.Code, biz.Msg)
		return
	}
	WriteException(c, err.Error())
}

func lowerFirst(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}

// BizError mirrors Java BizException with EbikeMsgCode.
type BizError struct {
	Code string
	Msg  string
}

func (e *BizError) Error() string { return e.Msg }
