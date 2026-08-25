package user

// Ports Java CreditScoreController + CreditScoreServiceImpl + CreditScoreGatewayImpl.

import (
	"encoding/json"
	"net/http"
	"time"

	"ebike-service-client-go/internal/api/dto"
	"ebike-service-client-go/internal/middleware"
	"ebike-service-client-go/internal/pkg/rpc"
	"ebike-service-client-go/internal/pkg/web"
	"github.com/gin-gonic/gin"
)

type creditScorePageDTO struct {
	dto.PageClientDTO
	Pin string `json:"pin" binding:"required"` // @NotBlank
}

var creditScoreMsgs = withBase(map[string]string{"pin": "must not be blank"})

func bindCreditScorePageDTO(c *gin.Context) (*creditScorePageDTO, bool) {
	// Java PageClientDTO field defaults pageNum=1, pageSize=10
	req := creditScorePageDTO{PageClientDTO: dto.PageClientDTO{PageNum: 1, PageSize: 10}}
	if !web.BindJSON(c, &req, creditScoreMsgs) {
		return nil, false
	}
	if !web.NotBlank(c, "pin", req.Pin) {
		return nil, false
	}
	return &req, true
}

// ---------------------------------------------------------------------------
// POST /client/ebike-user/creditScore/noRidding/info
// ---------------------------------------------------------------------------

// creditScoreNoRidingCO mirrors CreditScoreNoRidingCO; score/maxScore are
// BigDecimal.valueOf(Integer) in Java, i.e. plain JSON numbers.
type creditScoreNoRidingCO struct {
	Status         *int      `json:"status"`
	Score          *int      `json:"score"`
	EndTime        *DateTime `json:"endTime"`
	NoRiddingTotal *int      `json:"noRiddingTotal"` // never populated
	MaxScore       *int      `json:"maxScore"`
	Days           *int      `json:"days"`
}

func creditScoreInfo(c *gin.Context) {
	req, ok := bindCreditScorePageDTO(c)
	if !ok {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)

	// UserDetailQuery copy from query: only pin matches.
	raw, ok := callData(c, rpc.ServiceUser, "/user/detail",
		map[string]interface{}{"pin": req.Pin}, cmdCtx)
	if !ok {
		return
	}
	if isNullJSON(raw) {
		// Java: user.getRidingState() on null -> NPE
		writeNPE(c)
		return
	}
	var detail userDetailCo
	if err := json.Unmarshal(raw, &detail); err != nil {
		web.WriteException(c, err.Error())
		return
	}
	if detail.RidingState == nil {
		// Java: getRidingState() == 9 unboxing -> NPE
		writeNPE(c)
		return
	}
	days, status := 0, 0
	if *detail.RidingState == 9 {
		days, status = 99999, 2
	}
	if detail.PlatformScore == nil {
		// Java: BigDecimal.valueOf(getPlatformScore()) unboxing -> NPE
		writeNPE(c)
		return
	}
	endTime := DateTime{time.Now().AddDate(0, 0, days)}

	// CreditScoreV2Api.getConfig(BaseCmd) — BaseCmd has no fields.
	rawCfg, ok := callData(c, rpc.ServiceUser, "/creditScore/v2/getConfig",
		map[string]interface{}{}, cmdCtx)
	if !ok {
		return
	}
	if isNullJSON(rawCfg) {
		// Java: configCO.getMaxScore() on null -> NPE
		writeNPE(c)
		return
	}
	var cfg struct {
		MaxScore *int `json:"maxScore"`
	}
	if err := json.Unmarshal(rawCfg, &cfg); err != nil {
		web.WriteException(c, err.Error())
		return
	}
	if cfg.MaxScore == nil {
		// Java: BigDecimal.valueOf(getMaxScore()) unboxing -> NPE
		writeNPE(c)
		return
	}
	out := creditScoreNoRidingCO{
		Status:   &status,
		Score:    detail.PlatformScore,
		EndTime:  &endTime,
		MaxScore: cfg.MaxScore,
		Days:     &days,
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(out))
}

// ---------------------------------------------------------------------------
// POST /client/ebike-user/creditScore/list
// ---------------------------------------------------------------------------

func convertCreditScoreCOList(raw json.RawMessage) json.RawMessage {
	if isNullJSON(raw) {
		return raw
	}
	var list []creditScoreCO
	if err := json.Unmarshal(raw, &list); err != nil {
		return raw
	}
	b, _ := json.Marshal(list)
	return b
}

func creditScoreList(c *gin.Context) {
	req, ok := bindCreditScorePageDTO(c)
	if !ok {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	// UserDetailQuery copy {pin} -> /creditScore/list, List<CreditScoreCO> passthrough
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceUser, "/creditScore/list",
		map[string]interface{}{"pin": req.Pin}, cmdCtx)
	if err == nil && result != nil && result.Success {
		result.Data = convertCreditScoreCOList(result.Data)
	}
	web.RespondResult(c, result, err)
}

// ---------------------------------------------------------------------------
// POST /client/ebike-user/creditScore/page
// ---------------------------------------------------------------------------

// orderItem mirrors com.xyy.dto.OrderItem (asc defaults to true on the Java side).
type orderItem struct {
	Column *string `json:"column"`
	Asc    bool    `json:"asc"`
}

func (o *orderItem) UnmarshalJSON(b []byte) error {
	type alias orderItem
	tmp := alias{Asc: true}
	if err := json.Unmarshal(b, &tmp); err != nil {
		return err
	}
	*o = orderItem(tmp)
	return nil
}

type creditScoreRecordCO struct {
	Remark    *string   `json:"remark"`
	CreatedAt *DateTime `json:"createdAt"`
	Change    *int      `json:"change"`
}

// creditScoreCO mirrors the downstream CreditScoreCO built by toCreditScoreCO.
type creditScoreCO struct {
	Pin       *string   `json:"pin"` // never populated
	Score     *float64  `json:"score"`
	Type      *int      `json:"type"`
	Reason    *string   `json:"reason"`
	CreatedAt *DateTime `json:"createdAt"`
}

// creditScorePage* mirror com.xyy.dto.PageDTO; count is a primitive long, which
// the Java service serializes as a string (global ToStringSerializer covers Long.TYPE).
type creditScorePageIn struct {
	Count       LongStr               `json:"count"`
	PageNum     int                   `json:"pageNum"`
	PageSize    int                   `json:"pageSize"`
	Orders      []orderItem           `json:"orders"`
	SearchCount bool                  `json:"searchCount"`
	List        []creditScoreRecordCO `json:"list"`
}

type creditScorePageOut struct {
	Count       LongStr         `json:"count"`
	PageNum     int             `json:"pageNum"`
	PageSize    int             `json:"pageSize"`
	Orders      []orderItem     `json:"orders"`
	SearchCount bool            `json:"searchCount"`
	List        []creditScoreCO `json:"list"`
}

func creditScorePage(c *gin.Context) {
	req, ok := bindCreditScorePageDTO(c)
	if !ok {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)

	// CreditScorePageQuery copy from page: pin + PageQuery fields.
	searchCount := true
	if req.SearchCount != nil {
		searchCount = *req.SearchCount
	}
	body := map[string]interface{}{
		"pin":          req.Pin,
		"pageNum":      req.PageNum,
		"pageSize":     req.PageSize,
		"orders":       req.Orders,
		"searchCount":  searchCount,
		"lastRecordId": req.LastRecordId,
	}
	raw, ok := callData(c, rpc.ServiceUser, "/creditScore/v2/page", body, cmdCtx)
	if !ok {
		return
	}
	if isNullJSON(raw) {
		// Java: ConvertorHelper.convert(null, fn) -> null -> success(null)
		c.JSON(http.StatusOK, dto.NewSuccessResult(nil))
		return
	}
	// Jackson leaves PageDTO field defaults in place for absent properties.
	in := creditScorePageIn{PageNum: 1, PageSize: 10, SearchCount: true}
	if err := json.Unmarshal(raw, &in); err != nil {
		web.WriteException(c, err.Error())
		return
	}
	out := creditScorePageOut{
		Count:       in.Count,
		PageNum:     in.PageNum,
		PageSize:    in.PageSize,
		Orders:      in.Orders,
		SearchCount: in.SearchCount,
	}
	if in.List != nil {
		list := make([]creditScoreCO, 0, len(in.List))
		for _, rec := range in.List {
			if rec.Change == nil {
				// Java: Math.abs(co.getChange()) unboxing -> NPE
				writeNPE(c)
				return
			}
			score := *rec.Change
			typ := 1
			if score < 0 {
				score = -score
				typ = 2
			}
			scoreF := float64(score)
			list = append(list, creditScoreCO{
				Score:     &scoreF,
				Type:      &typ,
				Reason:    rec.Remark,
				CreatedAt: rec.CreatedAt,
			})
		}
		out.List = list
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(out))
}
