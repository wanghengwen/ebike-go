package management

// Ports Java MessageCenterController (@RequestMapping("/client/messageCenter")):
//   - POST /client/messageCenter/page/list            -> ebike-management /msg-record/list
//   - POST /client/messageCenter/getById              -> ebike-management /msg-record/getById
//   - POST /client/messageCenter/judgeIsHaveNoReadMsg -> ebike-management /msg-record/judgeIsHaveNoReadMsg
//
// BFF is passthrough; downstream getById marks izRead=true and updates version (side effect).
// Do not SHADOW mirror getById (see ebike-gateway-go/middleware/mirror.go).

import (
	"encoding/json"
	"net/http"

	"ebike-service-client-go/internal/api/dto"
	"ebike-service-client-go/internal/middleware"
	"ebike-service-client-go/internal/pkg/javacompat"
	"ebike-service-client-go/internal/pkg/rpc"
	"ebike-service-client-go/internal/pkg/web"
	"github.com/gin-gonic/gin"
)

// msgRecordPageQueryDTO matches Java MsgRecordPageQueryDto extends PageClientDTO
// (msgTypes is @NotEmpty Integer[] -> "required,min=1"; elements may be null).
type msgRecordPageQueryDTO struct {
	dto.PageClientDTO
	MsgTypes []*int `json:"msgTypes" binding:"required,min=1"`
}

// msgRecordQueryDTO matches Java MsgRecordQueryDto extends ClientDTO (id @NotNull).
type msgRecordQueryDTO struct {
	dto.ClientDTO
	Id *int64 `json:"id" binding:"required"`
}

// judgeIsReadMsgQueryDTO matches Java JudgeIsReadMsgQueryDto extends ClientDTO
// (msgTypes @NotEmpty).
type judgeIsReadMsgQueryDTO struct {
	dto.ClientDTO
	MsgTypes []*int `json:"msgTypes" binding:"required,min=1"`
}

var msgTypesMsgs = mergeMsgs(map[string]string{
	"msgTypes": "must not be empty", // @NotEmpty
})

var msgRecordIdMsgs = mergeMsgs(map[string]string{
	"id": "must not be null", // @NotNull
})

func msgRecordPageList(c *gin.Context) {
	// Java PageClientDTO field defaults pageNum=1, pageSize=10
	req := msgRecordPageQueryDTO{PageClientDTO: dto.PageClientDTO{PageNum: 1, PageSize: 10}}
	if !web.BindJSON(c, &req, msgTypesMsgs) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceManagement, "/msg-record/list", &req, cmdCtx)
	if err == nil && result != nil && result.Success {
		result.Data = convertMsgRecordPageDTO(result.Data)
	}
	web.RespondResult(c, result, err)
}

func msgRecordGetById(c *gin.Context) {
	if raw, ok := javacompat.ShadowJavaResult(c); ok {
		c.Data(http.StatusOK, "application/json; charset=utf-8", raw)
		return
	}
	var req msgRecordQueryDTO
	if !web.BindJSON(c, &req, msgRecordIdMsgs) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceManagement, "/msg-record/getById", &req, cmdCtx)
	if err == nil && result != nil && result.Success {
		result.Data = convertMsgRecordCO(result.Data)
	}
	web.RespondResult(c, result, err)
}

func judgeIsHaveNoReadMsg(c *gin.Context) {
	var req judgeIsReadMsgQueryDTO
	if !web.BindJSON(c, &req, msgTypesMsgs) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceManagement, "/msg-record/judgeIsHaveNoReadMsg", &req, cmdCtx)
	if err == nil && result != nil && result.Success {
		result.Data = convertMsgRecordCountCO(result.Data)
	}
	web.RespondResult(c, result, err)
}

type msgRecordCO struct {
	TenantId   *string              `json:"tenantId"`
	CreatedPin *string              `json:"createdPin"`
	CreatedAt  *javacompat.DateTime `json:"createdAt"`
	UpdatedPin *string              `json:"updatedPin"`
	UpdatedAt  *javacompat.DateTime `json:"updatedAt"`
	Version    *int                 `json:"version"`
	IzDel      *bool                `json:"izDel"`
	Id         *javacompat.LongStr  `json:"id"`
	MsgType    *int                 `json:"msgType"`
	Topic      *string              `json:"topic"`
	Content    *string              `json:"content"`
	IzRead     *bool                `json:"izRead"`
}

type msgRecordPageDTO struct {
	Count       *javacompat.LongStr `json:"count"`
	PageNum     *int                `json:"pageNum"`
	PageSize    *int                `json:"pageSize"`
	Orders      []json.RawMessage   `json:"orders"`
	SearchCount *bool               `json:"searchCount"`
	List        []msgRecordCO       `json:"list"`
}

func convertMsgRecordPageDTO(raw json.RawMessage) json.RawMessage {
	if javacompat.IsNullJSON(raw) {
		return raw
	}
	var page msgRecordPageDTO
	if json.Unmarshal(raw, &page) != nil {
		return raw
	}
	b, err := javacompat.MarshalJSONNoHTMLEscape(page)
	if err != nil {
		return raw
	}
	return b
}

func convertMsgRecordCO(raw json.RawMessage) json.RawMessage {
	if javacompat.IsNullJSON(raw) {
		return raw
	}
	var co msgRecordCO
	if json.Unmarshal(raw, &co) != nil {
		return raw
	}
	b, err := javacompat.MarshalJSONNoHTMLEscape(co)
	if err != nil {
		return raw
	}
	return b
}

type msgRecordCountCO struct {
	Count *int `json:"count"`
}

func convertMsgRecordCountCO(raw json.RawMessage) json.RawMessage {
	if javacompat.IsNullJSON(raw) {
		return raw
	}
	var co msgRecordCountCO
	if json.Unmarshal(raw, &co) != nil {
		return raw
	}
	b, _ := json.Marshal(co)
	return b
}
