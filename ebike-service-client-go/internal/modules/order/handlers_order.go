package order

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

func calculateCost(c *gin.Context) {
	var req cCalculateCmd
	if !web.BindJSON(c, &req, calculateMsgs) {
		return
	}
	if req.UserPin != nil && !web.NotBlank(c, "userPin", *req.UserPin) {
		return
	}
	if req.Imei != nil && !web.NotBlank(c, "imei", *req.Imei) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	body := req.toCalculateBody()
	raw, ok := callForData(c, rpc.ServiceOrder, "/order/calculateCost", body, cmdCtx)
	if !ok {
		return
	}
	var out cCalculateResult
	if json.Unmarshal(raw, &out) != nil {
		web.WriteException(c, "failed to decode calculate result")
		return
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(out))
}

func listOrders(c *gin.Context) {
	req := orderPageQuery{PageNum: 1, PageSize: 10}
	if !web.BindJSON(c, &req, nil) {
		return
	}
	req.normalizeSearchCount()
	if len(req.EndTime) == 0 {
		now := time.Now().In(cst8)
		start := javaMinusMonths(now, 6).UnixMilli()
		end := now.UnixMilli()
		req.EndTime = []int64{start, end}
	}
	cmdCtx := middleware.CompleteCommandContext(c, req.clientDTOPtr())
	raw, ok := callForData(c, rpc.ServiceOrder, "/order/list", &req, cmdCtx)
	if !ok {
		return
	}
	var page pageDTOIn
	if json.Unmarshal(raw, &page) != nil {
		web.WriteException(c, "failed to decode order page")
		return
	}
	out := pageOutFrom(&page)
	if len(page.List) == 0 {
		out.List = nil
		c.JSON(http.StatusOK, dto.NewSuccessResult(out))
		return
	}
	orders := make([]orderCO, 0, len(page.List))
	for _, item := range page.List {
		var o orderCO
		if json.Unmarshal(item, &o) == nil {
			orders = append(orders, o)
		}
	}
	if len(page.List) > 0 && len(orders) == 0 {
		web.WriteException(c, "failed to decode order page list")
		return
	}
	trajMap, err := getOrdersTrajectory(c.Request.Context(), cmdCtx, orders)
	if err != nil {
		writeNPE(c)
		return
	}
	list := make([]cLastOrderDetailCO, 0, len(orders))
	for _, o := range orders {
		d := orderCOToLastDetail(&o)
		if trajMap != nil {
			if id := longStr(o.Id); id != nil {
				if traj, ok := trajMap[*id]; ok {
					var items []deviceTrajectoryIn
					if json.Unmarshal(traj, &items) == nil {
						d.DeviceTrajectory = convertTrajectories(items)
					}
				}
			}
		}
		list = append(list, d)
	}
	out.List = list
	c.JSON(http.StatusOK, dto.NewSuccessResult(out))
}

func closeOrder(c *gin.Context) {
	var req cCloseOrderCmd
	if !web.BindJSON(c, &req, closeOrderMsgs) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	body := closeOrderCmdBody{OrderId: req.OrderId, PayType: req.PayType}
	raw, ok := callForData(c, rpc.ServiceOrder, "/order/closeOrder", body, cmdCtx)
	if !ok {
		return
	}
	var n int
	_ = json.Unmarshal(raw, &n)
	c.JSON(http.StatusOK, dto.NewSuccessResult(n))
}

func detailLast(c *gin.Context) {
	var req cOrderDetailQuery
	if !web.BindJSON(c, &req, nil) {
		return
	}
	detail, ok := buildLastDetail(c, &req)
	if !ok {
		return
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(detail))
}

func orderDetail(c *gin.Context) {
	var req cOrderDetailQuery
	if !web.BindJSON(c, &req, nil) {
		return
	}
	detail, ok := buildOrderDetail(c, &req)
	if !ok {
		return
	}
	if detail == nil {
		c.JSON(http.StatusOK, dto.NewSuccessResult(nil))
		return
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(detail))
}

func updateItinerary(c *gin.Context) {
	var req cUpdateItineraryCmd
	if !web.BindJSON(c, &req, updateItineraryMsgs) {
		return
	}
	if req.Imei != nil && !web.NotBlank(c, "imei", *req.Imei) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	body := updateItineraryCmdBody{Imei: req.Imei}
	raw, ok := callForData(c, rpc.ServiceOrder, "/order/updateItinerary", body, cmdCtx)
	if !ok {
		return
	}
	var s string
	_ = json.Unmarshal(raw, &s)
	c.JSON(http.StatusOK, dto.NewSuccessResult(s))
}

func frozenItinerary(c *gin.Context) {
	var req cCalculateCmd
	if !web.BindJSON(c, &req, calculateMsgs) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	body := req.toCalculateBody()
	if cmdCtx != nil && cmdCtx.Pin != "" {
		body.UserPin = &cmdCtx.Pin
	}
	if _, ok := callForData(c, rpc.ServiceOrder, "/order/frozen", body, cmdCtx); !ok {
		return
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(1))
}

func queryFrozen(c *gin.Context) {
	var req clientOrderIdQuery
	if !web.BindJSON(c, &req, orderIdMsgs) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	body := orderDetailQueryBody{OrderId: req.OrderId}
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceOrder, "/order/queryFrozen", body, cmdCtx)
	if err == nil && result != nil && result.Success {
		result.Data = convertOrderFrozenCO(result.Data)
	}
	web.RespondResult(c, result, err)
}

func listCanInvoiced(c *gin.Context) {
	req := orderPageQuery{PageNum: 1, PageSize: 10}
	if !web.BindJSON(c, &req, nil) {
		return
	}
	req.normalizeSearchCount()
	cmdCtx := middleware.CompleteCommandContext(c, req.clientDTOPtr())
	raw, ok := callForData(c, rpc.ServiceOrder, "/order/listInvoicedOrders", &req, cmdCtx)
	if !ok {
		return
	}
	var orders []orderCO
	if json.Unmarshal(raw, &orders) != nil {
		web.WriteException(c, "failed to decode invoiced orders")
		return
	}
	list := make([]cLastOrderDetailCO, 0, len(orders))
	for i := range orders {
		list = append(list, orderCOToLastDetail(&orders[i]))
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(list))
}

func buildLastDetail(c *gin.Context, req *cOrderDetailQuery) (*cLastOrderDetailCO, bool) {
	if req.CarId != nil {
		cmdCtx := middleware.CompleteCommandContext(c, req.clientDTOPtr())
		carResult, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceManagement, "/car-info/detail",
			carInfoCmdBody{CarId: req.CarId, Num: 0}, cmdCtx)
		if err != nil {
			web.WriteException(c, err.Error())
			return nil, false
		}
		if !carResult.Success {
			web.WriteBizResult(c, carResult)
			return nil, false
		}
		var car struct {
			Imei *string `json:"imei"`
		}
		_ = json.Unmarshal(carResult.Data, &car)
		req.Imei = car.Imei
	}
	cmdCtx := middleware.CompleteCommandContext(c, req.clientDTOPtr())
	body := orderDetailQueryBody{OrderId: req.OrderId, UserPin: req.UserPin, Imei: req.Imei, CarId: req.CarId}
	raw, ok := callForData(c, rpc.ServiceOrder, "/order/detailLast", body, cmdCtx)
	if !ok {
		return nil, false
	}
	if isNullRaw(raw) {
		empty := cLastOrderDetailCO{}
		return &empty, true
	}
	return enrichLastDetail(c, cmdCtx, raw, req.izNewApp())
}

func buildOrderDetail(c *gin.Context, req *cOrderDetailQuery) (*cOrderDetailCO, bool) {
	cmdCtx := middleware.CompleteCommandContext(c, req.clientDTOPtr())
	body := orderDetailQueryBody{OrderId: req.OrderId, UserPin: req.UserPin, Imei: req.Imei, CarId: req.CarId}
	raw, ok := callForData(c, rpc.ServiceOrder, "/order/detail", body, cmdCtx)
	if !ok {
		return nil, false
	}
	if isNullRaw(raw) {
		return nil, true
	}
	var in orderDetailIn
	if json.Unmarshal(raw, &in) != nil || in.OrderCO == nil {
		web.WriteException(c, "failed to decode order detail")
		return nil, false
	}
	d := orderCOToOrderDetail(in.OrderCO)
	if in.OrderItemCO != nil {
		applyItemToOrderDetail(&d, in.OrderItemCO)
		d.Id = longStr(in.OrderCO.Id)
	}
	w := getUserWallet(c.Request.Context(), cmdCtx, rawStringPtr(in.OrderCO.UserPin))
	d.Balance, d.Recharge, d.Present = w.Balance, w.Recharge, w.Present
	traj, err := getOrderTrajectory(c.Request.Context(), cmdCtx, in.OrderCO)
	if err != nil {
		writeNPE(c)
		return nil, false
	}
	d.DeviceTrajectory = traj
	applyMileCostMergeOrder(&d, req.izNewApp())
	return &d, true
}

func enrichLastDetail(c *gin.Context, cmdCtx *dto.CommandContext, raw json.RawMessage, izNewApp bool) (*cLastOrderDetailCO, bool) {
	var in orderDetailIn
	if json.Unmarshal(raw, &in) != nil || in.OrderCO == nil {
		web.WriteException(c, "failed to decode order detail")
		return nil, false
	}
	d := orderCOToLastDetail(in.OrderCO)
	if in.OrderItemCO != nil {
		applyItemToLastDetail(&d, in.OrderItemCO)
		d.Id = longStr(in.OrderCO.Id)
	}
	w := getUserWallet(c.Request.Context(), cmdCtx, rawStringPtr(in.OrderCO.UserPin))
	d.Balance, d.Recharge, d.Present = w.Balance, w.Recharge, w.Present
	traj, err := getOrderTrajectory(c.Request.Context(), cmdCtx, in.OrderCO)
	if err != nil {
		writeNPE(c)
		return nil, false
	}
	d.DeviceTrajectory = traj
	applyMileCostMergeLast(&d, izNewApp)
	return &d, true
}

func (q *cOrderDetailQuery) izNewApp() bool {
	if q.IzNewApp == nil {
		return false
	}
	return *q.IzNewApp
}

func applyMileCostMergeLast(d *cLastOrderDetailCO, izNewApp bool) {
	if izNewApp {
		return
	}
	sp, ok := rawInt(d.StartPrice)
	if !ok || sp == 0 {
		return
	}
	mc, ok := rawInt(d.MileCost)
	if !ok {
		return
	}
	sum := mc + sp
	b, _ := json.Marshal(sum)
	d.MileCost = b
}

func applyMileCostMergeOrder(d *cOrderDetailCO, izNewApp bool) {
	if izNewApp {
		return
	}
	sp, ok := rawInt(d.StartPrice)
	if !ok || sp == 0 {
		return
	}
	mc, ok := rawInt(d.MileCost)
	if !ok {
		return
	}
	sum := mc + sp
	b, _ := json.Marshal(sum)
	d.MileCost = b
}
