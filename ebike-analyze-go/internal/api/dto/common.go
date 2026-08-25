package dto

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// DateOnlyValue serializes a time.Time as "yyyy-MM-dd", mirroring Java's
// @JsonFormat(pattern = "yyyy-MM-dd") on LocalDate fields.
type DateOnlyValue time.Time

const dateOnlyLayout = "2006-01-02"

func (d DateOnlyValue) MarshalJSON() ([]byte, error) {
	t := time.Time(d)
	if t.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(t.Format(dateOnlyLayout))
}

func (d *DateOnlyValue) UnmarshalJSON(b []byte) error {
	var str string
	if err := json.Unmarshal(b, &str); err != nil {
		return err
	}
	if str == "" {
		*d = DateOnlyValue(time.Time{})
		return nil
	}
	t, err := time.ParseInLocation(dateOnlyLayout, str, time.Local)
	if err != nil {
		return err
	}
	*d = DateOnlyValue(t)
	return nil
}

// AsTime returns the underlying time.Time.
func (d DateOnlyValue) AsTime() time.Time { return time.Time(d) }

// LocalDate returns the calendar date at local midnight, mirroring Java LocalDate.
func (d DateOnlyValue) LocalDate() time.Time { return NormalizeLocalDate(time.Time(d)) }

// NormalizeLocalDate strips the time component in the local timezone.
func NormalizeLocalDate(t time.Time) time.Time {
	if t.IsZero() {
		return t
	}
	in := t.In(time.Local)
	return time.Date(in.Year(), in.Month(), in.Day(), 0, 0, 0, 0, time.Local)
}

// NewDateOnly wraps a time.Time as DateOnlyValue.
func NewDateOnly(t time.Time) DateOnlyValue { return DateOnlyValue(t) }

// DateTimeValue serializes as "yyyy-MM-dd HH:mm:ss", mirroring Java
// @JsonFormat(pattern = "yyyy-MM-dd HH:mm:ss") on LocalDateTime fields.
type DateTimeValue time.Time

const dateTimeLayout = "2006-01-02 15:04:05"

func (d DateTimeValue) MarshalJSON() ([]byte, error) {
	t := time.Time(d)
	if t.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(t.Format(dateTimeLayout))
}

func (d *DateTimeValue) UnmarshalJSON(b []byte) error {
	var str string
	if err := json.Unmarshal(b, &str); err != nil {
		return err
	}
	if str == "" {
		*d = DateTimeValue(time.Time{})
		return nil
	}
	t, err := parseDateTimeString(str)
	if err != nil {
		return err
	}
	*d = DateTimeValue(t)
	return nil
}

func (d DateTimeValue) AsTime() time.Time { return time.Time(d) }

func NewDateTime(t time.Time) DateTimeValue { return DateTimeValue(t) }

// LocalDateTimeValue serializes as "yyyy-MM-dd'T'HH:mm:ss" (no timezone),
// mirroring Jackson's default Java 8 LocalDateTime JSON format.
type LocalDateTimeValue time.Time

const localDateTimeLayout = "2006-01-02T15:04:05"

func (d LocalDateTimeValue) MarshalJSON() ([]byte, error) {
	t := time.Time(d)
	if t.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(t.Format(localDateTimeLayout))
}

func (d *LocalDateTimeValue) UnmarshalJSON(b []byte) error {
	var str string
	if err := json.Unmarshal(b, &str); err != nil {
		return err
	}
	if str == "" {
		*d = LocalDateTimeValue(time.Time{})
		return nil
	}
	t, err := parseDateTimeString(str)
	if err != nil {
		return err
	}
	*d = LocalDateTimeValue(t)
	return nil
}

func (d LocalDateTimeValue) AsTime() time.Time { return time.Time(d) }

func NewLocalDateTime(t time.Time) LocalDateTimeValue { return LocalDateTimeValue(t) }

func parseDateTimeString(str string) (time.Time, error) {
	layouts := []string{
		dateTimeLayout,
		localDateTimeLayout,
		"2006-01-02T15:04",
		time.RFC3339,
		"2006-01-02 15:04:05.000",
	}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, str, time.Local); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid datetime: %s", str)
}

// JSONInt64 accepts JSON numbers or quoted decimal strings, mirroring Jackson Long coercion.
type JSONInt64 int64

func (n *JSONInt64) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		*n = 0
		return nil
	}
	var num int64
	if err := json.Unmarshal(b, &num); err == nil {
		*n = JSONInt64(num)
		return nil
	}
	var str string
	if err := json.Unmarshal(b, &str); err != nil {
		return err
	}
	if str == "" {
		*n = 0
		return nil
	}
	v, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return err
	}
	*n = JSONInt64(v)
	return nil
}

func (n JSONInt64) Int64() int64 { return int64(n) }

func (n *JSONInt64) PtrInt64() *int64 {
	if n == nil {
		return nil
	}
	v := n.Int64()
	return &v
}

// JSONInt64Slice accepts JSON arrays of numbers or quoted decimal strings.
type JSONInt64Slice []int64

func (s *JSONInt64Slice) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		*s = nil
		return nil
	}
	var raw []json.RawMessage
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	out := make([]int64, 0, len(raw))
	for _, item := range raw {
		var n JSONInt64
		if err := json.Unmarshal(item, &n); err != nil {
			return err
		}
		out = append(out, n.Int64())
	}
	*s = out
	return nil
}

// FormatMsgCode substitutes {0}, {1}, ... placeholders in a message template,
// mirroring Java MessageSource / MsgCode argument interpolation fallback.
func FormatMsgCode(template string, args []interface{}) string {
	msg := template
	for i, arg := range args {
		msg = strings.ReplaceAll(msg, fmt.Sprintf("{%d}", i), fmt.Sprint(arg))
	}
	return msg
}

// Command is the standard request DTO for POST endpoints.
// Mirrors Java Command<T> with commandContext + data payload.
type Command struct {
	CommandContext *CommandContext `json:"commandContext"`
	Data           interface{}     `json:"data,omitempty"`
}

// CommandContext carries tenant/user/trace metadata through the request pipeline.
// Mirrors Java CommandContext.
type CommandContext struct {
	TenantId      string `json:"tenantId"`
	TraceId       string `json:"traceId"`
	Pin           string `json:"pin"`
	IP            string `json:"ip"`
	Platform      string `json:"platform"`
	DeviceId      string `json:"deviceId"`
	Source        string `json:"source,omitempty"`
	Name          string `json:"name"`
	StressTesting bool   `json:"stressTesting"`
}

// Result is the standard response DTO.
// Mirrors Java Result<T>. Code and Msg are pointers for JSON null compatibility.
type Result struct {
	Success bool        `json:"success"`
	Code    *string     `json:"code"`
	Msg     *string     `json:"msg"`
	Data    interface{} `json:"data"`
}

// Standard error codes mirroring Java com.xyy.common.constant.MsgCodeEnum.
const (
	CodeSuccess         = "0"     // MsgCodeEnum.SUCCESS
	CodeException       = "00001" // MsgCodeEnum.EXCEPTION
	CodeParamException  = "00002" // MsgCodeEnum.PARAM_EXCEPTION
	CodeBeanCopy        = "00003" // MsgCodeEnum.BEAN_COPY_EXCEPTION
	CodeIllegalArgument = "00004" // MsgCodeEnum.ILLEGAL_ARGUMENT
	CodeNotNull         = "00018" // MsgCodeEnum.NOT_NULL ("{0} 不能为null")
)

// NewSuccessResult creates a successful Result.
// Mirrors Java ResultHelper.success(data) -> Result(data, MsgCodeEnum.SUCCESS, true)
// where SUCCESS = ("0", "success", "成功").
func NewSuccessResult(data interface{}) Result {
	code := CodeSuccess
	msg := "成功"
	return Result{
		Success: true,
		Code:    &code,
		Msg:     &msg,
		Data:    data,
	}
}

// NewErrorResult creates an error Result.
func NewErrorResult(code, msg string) Result {
	return Result{
		Success: false,
		Code:    &code,
		Msg:     &msg,
	}
}
