package dto

import "encoding/json"

// Error codes matching Java com.xyy.common.constant.MsgCodeEnum
const (
	CodeSuccess         = "0"     // SUCCESS
	CodeException       = "00001" // EXCEPTION
	CodeParamException  = "00002" // PARAM_EXCEPTION
	CodeIllegalArgument = "00004" // ILLEGAL_ARGUMENT
)

// Result is the standard JSON response format for all C-end gateway APIs.
// Matches Java: com.xyy.dto.Result
//
// Field types are chosen for byte-level parity with the Java service
// (whose ObjectMapper does NOT enable NON_NULL, so null fields are emitted):
//   - Code/Msg are *string so a downstream null stays null (not "")
//   - Data is json.RawMessage so the downstream payload passes through verbatim,
//     preserving Long-as-string values and avoiding float64 precision loss;
//     a nil RawMessage marshals to "data":null exactly like Java
type Result struct {
	Success bool            `json:"success"`
	Code    *string         `json:"code"`
	Msg     *string         `json:"msg"`
	Data    json.RawMessage `json:"data"`
}

// NewErrorResult builds a failure Result, mirroring Java's ResultHelper
// (success=false, data=null).
func NewErrorResult(code, msg string) Result {
	return Result{Success: false, Code: &code, Msg: &msg}
}

// CommandContext matches Java: com.xyy.dto.CommandContext
// This is the context information that gets injected into every downstream Command.
type CommandContext struct {
	TenantId      string `json:"tenantId,omitempty"`
	TraceId       string `json:"traceId,omitempty"`
	Pin           string `json:"pin,omitempty"`
	Ip            string `json:"ip,omitempty"`
	Platform      string `json:"platform,omitempty"`
	DeviceId      string `json:"deviceId,omitempty"`
	Source        string `json:"source,omitempty"`
	Name          string `json:"name,omitempty"`
	StressTesting bool   `json:"stressTesting"`
}

// ClientDTO represents common client-side request parameters.
// Matches Java: com.xyy.dto.ClientDTO extends DTO
// (traceId/tenantId are @NotEmpty in Java)
// Form tags support Java handlers without @RequestBody (Spring form/query binding).
type ClientDTO struct {
	TraceId       string   `json:"traceId" form:"traceId" binding:"required"`
	TenantId      string   `json:"tenantId" form:"tenantId" binding:"required"`
	Platform      string   `json:"platform,omitempty" form:"platform"`
	DeviceId      string   `json:"deviceId,omitempty" form:"deviceId"`
	Version       string   `json:"version,omitempty" form:"version"`
	Ip            string   `json:"ip,omitempty" form:"ip"`
	Longitude     *float64 `json:"longitude,omitempty" form:"longitude"` // Java Double (boxed) -> Go *float64
	Latitude      *float64 `json:"latitude,omitempty" form:"latitude"`   // Java Double (boxed) -> Go *float64
	Source        string   `json:"source,omitempty" form:"source"`
	StressTesting bool     `json:"stressTesting" form:"stressTesting"`
}

// OrderItem matches Java: com.xyy.dto.OrderItem (used in pagination sorting)
// Java declares `private boolean asc = true;` so the default when the client
// omits "asc" must be true — handled by the custom UnmarshalJSON below.
type OrderItem struct {
	Column string `json:"column,omitempty"`
	Asc    bool   `json:"asc"`
}

// UnmarshalJSON applies the Java field default asc=true before decoding.
func (o *OrderItem) UnmarshalJSON(data []byte) error {
	type orderItemAlias OrderItem
	tmp := orderItemAlias{Asc: true}
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	*o = OrderItem(tmp)
	return nil
}

// PageClientDTO is the standard pagination request body.
// Matches Java: com.xyy.dto.PageClientDTO extends ClientDTO
// CRITICAL: Java field is "pageNum", NOT "pageNo"!
// Java field defaults (pageNum=1, pageSize=10, searchCount=true) must be applied
// by the caller before binding (see controller) / during conversion.
type PageClientDTO struct {
	ClientDTO
	PageNum      int         `json:"pageNum"`
	PageSize     int         `json:"pageSize"`
	Orders       []OrderItem `json:"orders,omitempty"`
	SearchCount  *bool       `json:"searchCount,omitempty"` // nil => Java default true
	LastRecordId string      `json:"lastRecordId,omitempty"`
}
