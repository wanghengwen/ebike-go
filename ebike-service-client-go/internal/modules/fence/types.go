// Package fence ports the Java fence-related controllers of ebike-service-client:
// FenceController, CreditScoreConfigController, ProtocolConfigController,
// ResourceManagementController, SystemConfigController, HelmetController,
// SiteApplicationController, AdConfigController, HelpConfigController and
// RidingPermissionController.
//
// Java class-level mappings resolved from Nacos config (ebike-service-client.yml):
//
//	${ebike.fence.name} = client/fence
package fence

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"ebike-service-client-go/internal/api/dto"
	"ebike-service-client-go/internal/pkg/javacompat"
	"github.com/gin-gonic/gin"
)

// ---------------------------------------------------------------------------
// Shared helper types
// ---------------------------------------------------------------------------

// jLong mirrors a Java Long that is globally serialized with ToStringSerializer:
// it accepts a JSON number or string on input and always marshals as a string.
// Use *jLong so that null stays null.
type jLong string

func (l *jLong) UnmarshalJSON(b []byte) error {
	s := bytes.TrimSpace(b)
	if string(s) == "null" {
		return nil
	}
	if len(s) > 0 && s[0] == '"' {
		var str string
		if err := json.Unmarshal(s, &str); err != nil {
			return err
		}
		*l = jLong(str)
		return nil
	}
	*l = jLong(s)
	return nil
}

func (l jLong) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(l))
}

// int64Value parses the jLong for numeric comparisons (Java Long.equals).
func (l *jLong) int64Value() (int64, bool) {
	if l == nil {
		return 0, false
	}
	v, err := strconv.ParseInt(string(*l), 10, 64)
	if err != nil {
		return 0, false
	}
	return v, true
}

// jDateTime mirrors a Java LocalDateTime bound/serialized with the
// "yyyy-MM-dd HH:mm:ss" pattern. An invalid value fails unmarshalling, which
// maps to Java's HttpMessageNotReadableException (code 00002).
type jDateTime string

const javaDateTimeLayout = "2006-01-02 15:04:05"

func (t *jDateTime) UnmarshalJSON(b []byte) error {
	var s *string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	if s == nil {
		return nil
	}
	if _, err := time.ParseInLocation(javaDateTimeLayout, *s, time.Local); err != nil {
		return fmt.Errorf("cannot parse %q as LocalDateTime (yyyy-MM-dd HH:mm:ss)", *s)
	}
	*t = jDateTime(*s)
	return nil
}

func (t jDateTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(t))
}

// remindWayStr mirrors Java ConfigPayDTO.setRemindWay(Integer[]): the request
// field is an int array which is stored (and forwarded) as a comma-joined
// string, e.g. [0,1,2] -> "0,1,2", [] -> "".
type remindWayStr string

func (r *remindWayStr) UnmarshalJSON(b []byte) error {
	var arr *[]int64
	if err := json.Unmarshal(b, &arr); err != nil {
		return err
	}
	if arr == nil {
		return nil
	}
	parts := make([]string, 0, len(*arr))
	for _, n := range *arr {
		parts = append(parts, strconv.FormatInt(n, 10))
	}
	*r = remindWayStr(strings.Join(parts, ","))
	return nil
}

func (r remindWayStr) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(r))
}

// errNPE mirrors a java.lang.NullPointerException raised inside the Java
// service/controller layer; it is rendered by the GlobalExceptionHandler as
// code 00001.
var errNPE = errors.New("NullPointerException")

// ---------------------------------------------------------------------------
// Loose (non-validated) ClientDTO
// ---------------------------------------------------------------------------

// clientLoose mirrors com.xyy.dto.ClientDTO for Java handlers WITHOUT
// @Validated/@Valid (or with group-only validation): the @NotEmpty
// constraints on traceId/tenantId are NOT enforced.
type clientLoose struct {
	TraceId       string   `json:"traceId"`
	TenantId      string   `json:"tenantId"`
	Platform      string   `json:"platform,omitempty"`
	DeviceId      string   `json:"deviceId,omitempty"`
	Version       string   `json:"version,omitempty"`
	Ip            string   `json:"ip,omitempty"`
	Longitude     *float64 `json:"longitude,omitempty"`
	Latitude      *float64 `json:"latitude,omitempty"`
	Source        string   `json:"source,omitempty"`
	StressTesting bool     `json:"stressTesting"`
}

// client converts the loose DTO to the shared ClientDTO so that
// middleware.CompleteCommandContext can be reused.
func (l *clientLoose) client() *dto.ClientDTO {
	return &dto.ClientDTO{
		TraceId:       l.TraceId,
		TenantId:      l.TenantId,
		Platform:      l.Platform,
		DeviceId:      l.DeviceId,
		Version:       l.Version,
		Ip:            l.Ip,
		Longitude:     l.Longitude,
		Latitude:      l.Latitude,
		Source:        l.Source,
		StressTesting: l.StressTesting,
	}
}

// ---------------------------------------------------------------------------
// Response helpers (Java exception-path parity)
// ---------------------------------------------------------------------------

// failResult mirrors ResultHelper.getResultData throwing
// BizException(result.code, result.msg) for an unsuccessful downstream Result,
// which the GlobalExceptionHandler renders as
// {success:false, code:<downstream code>, msg:<downstream msg>, data:null}.
func failResult(r *dto.Result) *dto.Result {
	return &dto.Result{Success: false, Code: r.Code, Msg: r.Msg, Data: nil}
}

// writeRpcAdviceError mirrors RpcAdviceAspect: any non-BizException thrown by
// a @RpcAdvice gateway method (e.g. a Feign transport error) is rewrapped as
// BizException(EXCEPTION, "RPC调用失败").
func writeRpcAdviceError(c *gin.Context) {
	c.JSON(http.StatusOK, dto.NewErrorResult(dto.CodeException, "RPC调用失败"))
}

// writeNPE mirrors an uncaught java.lang.NullPointerException reaching the
// GlobalExceptionHandler (code 00001).
var writeNPE = javacompat.WriteNPE

// cCreditScoreConfigCO mirrors client CCreditScoreConfigCO (null fields emitted).
type cCreditScoreConfigCO struct {
	Id                  *string  `json:"id"`
	IzCreditScore       *bool    `json:"izCreditScore"`
	Score               *float64 `json:"score"`
	WarnScore           *float64 `json:"warnScore"`
	NoRiddingScore      *float64 `json:"noRiddingScore"`
	FirstNoRiddingDays  *int     `json:"firstNoRiddingDays"`
	SecondNoRiddingDays *int     `json:"secondNoRiddingDays"`
	MoreNoRiddingDays   *int     `json:"moreNoRiddingDays"`
	AddScore            *float64 `json:"addScore"`
	AddScoreUpperLimit  *float64 `json:"addScoreUpperLimit"`
	RemindWay           *string  `json:"remindWay"`
	IzLowScoreOn        *int     `json:"izLowScoreOn"`
}

// isJSONNull reports whether a downstream data payload is JSON null/absent.
func isJSONNull(raw json.RawMessage) bool {
	t := bytes.TrimSpace(raw)
	return len(t) == 0 || string(t) == "null"
}

// requireNotNull writes the Java bean-validation error for @NotNull and
// returns false when ok is false. Used for group-based validation
// (@Validated(CreateGroup)/(UpdateGroup)) which gin tags cannot express.
func requireNotNull(c *gin.Context, field string, ok bool) bool {
	if !ok {
		c.JSON(http.StatusOK, dto.NewErrorResult(dto.CodeIllegalArgument, field+" must not be null"))
		return false
	}
	return true
}

// ---------------------------------------------------------------------------
// Java utility ports
// ---------------------------------------------------------------------------

// javaSplit mirrors Java String.split(","): trailing empty strings are
// removed, and a zero-length input yields [""].
func javaSplit(s string) []string {
	if !strings.Contains(s, ",") {
		return []string{s}
	}
	parts := strings.Split(s, ",")
	n := len(parts)
	for n > 0 && parts[n-1] == "" {
		n--
	}
	return parts[:n]
}

// containsAny mirrors hutool CollectionUtil.containsAny.
func containsAny(a, b []string) bool {
	set := make(map[string]struct{}, len(a))
	for _, s := range a {
		set[s] = struct{}{}
	}
	for _, s := range b {
		if _, ok := set[s]; ok {
			return true
		}
	}
	return false
}

// isShowActivity ports com.xyy.ebike.service.client.common.utils.ActivityShowUtils.isShow.
// byRegister/byTags are unboxed in Java (NPE when null) and must be checked by
// the caller before invoking this function.
func isShowActivity(byRegister, byTags bool, visible *int, isNewUser bool, userTags, activeTags *string) bool {
	registerShow := false
	tagsShow := false
	if byRegister {
		registerShow = visible != nil && (*visible == 0 || (*visible == 1 && isNewUser))
	}
	if byTags {
		at, ut := "", ""
		if activeTags != nil {
			at = *activeTags
		}
		if userTags != nil {
			ut = *userTags
		}
		tagsShow = containsAny(javaSplit(at), javaSplit(ut))
	}
	return registerShow || tagsShow
}
