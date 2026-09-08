package order

import (
	"context"
	"encoding/json"
	"errors"

	"ebike-service-client-go/internal/api/dto"
	"ebike-service-client-go/internal/pkg/javacompat"
	"ebike-service-client-go/internal/pkg/rpc"
)

// ---------------------------------------------------------------------------
// Downstream client objects (parsed) and locally constructed response objects
// for the OrderController endpoints.
//
// Conversion rules mirror Java exactly:
//   - ConvertorHelper.copyProperties(OrderCO, CLastOrderDetailCO/COrderDetailCO):
//     Spring BeanUtils copies only same-named properties of assignable types
//     (e.g. OrderCO.carType is NOT copied into bikeType -> null).
//   - BeanUtils.copyProperties(OrderItemCO, detail): same-named fields are
//     overwritten UNCONDITIONALLY, including nulls (the item value wins).
//   - Long fields serialize as strings (Java global ToStringSerializer),
//     LocalDateTime as "yyyy-MM-dd HH:mm:ss".
// ---------------------------------------------------------------------------

// orderCO parses the downstream com.xyy.ebike.order.api.dto.OrderCO, keeping
// every field raw for passthrough.
type orderCO struct {
	Id                  json.RawMessage `json:"id"`
	UserPin             json.RawMessage `json:"userPin"`
	Phone               json.RawMessage `json:"phone"`
	ServiceId           json.RawMessage `json:"serviceId"`
	OriginCost          json.RawMessage `json:"originCost"`
	PayCost             json.RawMessage `json:"payCost"`
	ModifyCost          json.RawMessage `json:"modifyCost"`
	ModifyPayCost       json.RawMessage `json:"modifyPayCost"`
	ModifyDispatchCost  json.RawMessage `json:"modifyDispatchCost"`
	ModifyHelmetPenalty json.RawMessage `json:"modifyHelmetPenalty"`
	Mile                json.RawMessage `json:"mile"`
	RidingTime          json.RawMessage `json:"ridingTime"`
	IzPaid              json.RawMessage `json:"izPaid"`
	IzComplained        json.RawMessage `json:"izComplained"`
	IzModified          json.RawMessage `json:"izModified"`
	IzRepair            json.RawMessage `json:"izRepair"`
	InvoiceId           json.RawMessage `json:"invoiceId"`
	Imei                json.RawMessage `json:"imei"`
	CarId               json.RawMessage `json:"carId"`
	PaidInfo            json.RawMessage `json:"paidInfo"`
	StartLat            json.RawMessage `json:"startLat"`
	StartLng            json.RawMessage `json:"startLng"`
	EndLat              json.RawMessage `json:"endLat"`
	EndLng              json.RawMessage `json:"endLng"`
	PayTime             json.RawMessage `json:"payTime"`
	StartTime           json.RawMessage `json:"startTime"`
	EndTime             json.RawMessage `json:"endTime"`
	Penalty             json.RawMessage `json:"penalty"`
	PenaltyType         json.RawMessage `json:"penaltyType"`
	Discount            json.RawMessage `json:"discount"`
	IzRidingCard        json.RawMessage `json:"izRidingCard"`
	RechargeCost        json.RawMessage `json:"rechargeCost"`
	PresentCost         json.RawMessage `json:"presentCost"`
	PayType             json.RawMessage `json:"payType"`
	IzFreeN             json.RawMessage `json:"izFreeN"`
	StartPrice          json.RawMessage `json:"startPrice"`
	HelmetPenalty       json.RawMessage `json:"helmetPenalty"`
	IzDiscount          json.RawMessage `json:"izDiscount"`
	IzActityFree        json.RawMessage `json:"izActityFree"`
	IzFavorable         json.RawMessage `json:"izFavorable"`
}

// orderItemCO parses the downstream com.xyy.ebike.order.api.dto.OrderItemCO.
type orderItemCO struct {
	Id                    json.RawMessage `json:"id"`
	BillingConfigId       json.RawMessage `json:"billingConfigId"`
	Penalty               json.RawMessage `json:"penalty"`
	PenaltyType           json.RawMessage `json:"penaltyType"`
	IzImpunity            json.RawMessage `json:"izImpunity"`
	DiscountCount         json.RawMessage `json:"discountCount"`
	Discount              json.RawMessage `json:"discount"`
	IzDiscount            json.RawMessage `json:"izDiscount"`
	Deduction             json.RawMessage `json:"deduction"`
	IzRidingCard          json.RawMessage `json:"izRidingCard"`
	RidingCardCount       json.RawMessage `json:"ridingCardCount"`
	IzActityFree          json.RawMessage `json:"izActityFree"`
	ActityFreeCount       json.RawMessage `json:"actityFreeCount"`
	IzFavorable           json.RawMessage `json:"izFavorable"`
	RechargeCost          json.RawMessage `json:"rechargeCost"`
	PresentCost           json.RawMessage `json:"presentCost"`
	TimeCost              json.RawMessage `json:"timeCost"`
	MileCost              json.RawMessage `json:"mileCost"`
	HasPaid               json.RawMessage `json:"hasPaid"`
	DispatchCost          json.RawMessage `json:"dispatchCost"`
	HelmetPenalty         json.RawMessage `json:"helmetPenalty"`
	DiscountType          json.RawMessage `json:"discountType"`
	IzRandomDeduct        json.RawMessage `json:"izRandomDeduct"`
	RandomCount           json.RawMessage `json:"randomCount"`
	IzFreeN               json.RawMessage `json:"izFreeN"`
	StartPrice            json.RawMessage `json:"startPrice"`
	IzRedPaket            json.RawMessage `json:"izRedPaket"`
	RedPaketCarDeduct     json.RawMessage `json:"redPaketCarDeduct"`
	RedPaketParkingDeduct json.RawMessage `json:"redPaketParkingDeduct"`
	CarActivityId         json.RawMessage `json:"carActivityId"`
	ParkingActivityId     json.RawMessage `json:"parkingActivityId"`
}

// orderDetailIn parses the downstream OrderDetailCO wrapper.
type orderDetailIn struct {
	OrderCO     *orderCO     `json:"orderCO"`
	OrderItemCO *orderItemCO `json:"orderItemCO"`
}

// deviceTrajectoryIn parses the downstream DeviceTrajectoryCo.
type deviceTrajectoryIn struct {
	Lng       json.RawMessage `json:"lng"`
	Lat       json.RawMessage `json:"lat"`
	Speed     json.RawMessage `json:"speed"`
	Course    json.RawMessage `json:"course"`
	Timestamp json.RawMessage `json:"timestamp"`
}

// deviceTrajectoryCO mirrors the client-facing DeviceTrajectoryCO
// (timestamp is a Java Long -> string).
type deviceTrajectoryCO struct {
	Lng       json.RawMessage `json:"lng"`
	Lat       json.RawMessage `json:"lat"`
	Speed     json.RawMessage `json:"speed"`
	Course    json.RawMessage `json:"course"`
	Timestamp *string         `json:"timestamp"`
}

func convertTrajectories(items []deviceTrajectoryIn) []deviceTrajectoryCO {
	out := make([]deviceTrajectoryCO, 0, len(items))
	for _, it := range items {
		out = append(out, deviceTrajectoryCO{
			Lng:       it.Lng,
			Lat:       it.Lat,
			Speed:     it.Speed,
			Course:    it.Course,
			Timestamp: longStr(it.Timestamp),
		})
	}
	return out
}

// cLastOrderDetailCO mirrors the Java CLastOrderDetailCO field declaration
// order. Long fields (id/serviceId/invoiceId/billingConfigId via
// @JsonSerialize, ridingTime/carActivityId/parkingActivityId via the global
// serializer) are emitted as strings.
type cLastOrderDetailCO struct {
	Id                    *string              `json:"id"`
	UserPin               json.RawMessage      `json:"userPin"`
	Phone                 json.RawMessage      `json:"phone"`
	ServiceId             *string              `json:"serviceId"`
	BikeType              json.RawMessage      `json:"bikeType"`
	OriginCost            json.RawMessage      `json:"originCost"`
	PayCost               json.RawMessage      `json:"payCost"`
	ModifyCost            json.RawMessage      `json:"modifyCost"`
	ModifyPayCost         json.RawMessage      `json:"modifyPayCost"`
	ModifyDispatchCost    json.RawMessage      `json:"modifyDispatchCost"`
	ModifyHelmetPenalty   json.RawMessage      `json:"modifyHelmetPenalty"`
	Mile                  json.RawMessage      `json:"mile"`
	RidingTime            *string              `json:"ridingTime"`
	IzPaid                json.RawMessage      `json:"izPaid"`
	IzComplained          json.RawMessage      `json:"izComplained"`
	IzModified            json.RawMessage      `json:"izModified"`
	IzRepair              json.RawMessage      `json:"izRepair"`
	InvoiceId             *string              `json:"invoiceId"`
	Imei                  json.RawMessage      `json:"imei"`
	CarId                 json.RawMessage      `json:"carId"`
	PaidInfo              json.RawMessage      `json:"paidInfo"`
	StartLat              json.RawMessage      `json:"startLat"`
	StartLng              json.RawMessage      `json:"startLng"`
	EndLat                json.RawMessage      `json:"endLat"`
	EndLng                json.RawMessage      `json:"endLng"`
	PayTime               json.RawMessage      `json:"payTime"`
	StartTime             json.RawMessage      `json:"startTime"`
	EndTime               json.RawMessage      `json:"endTime"`
	BillingConfigId       *string              `json:"billingConfigId"`
	Penalty               json.RawMessage      `json:"penalty"`
	PenaltyType           json.RawMessage      `json:"penaltyType"`
	Discount              json.RawMessage      `json:"discount"`
	IzRidingCard          json.RawMessage      `json:"izRidingCard"`
	Deduction             json.RawMessage      `json:"deduction"`
	FreeOrderTime         json.RawMessage      `json:"freeOrderTime"`
	RechargeCost          json.RawMessage      `json:"rechargeCost"`
	PresentCost           json.RawMessage      `json:"presentCost"`
	PayType               json.RawMessage      `json:"payType"`
	Balance               json.RawMessage      `json:"balance"`
	Recharge              json.RawMessage      `json:"recharge"`
	Present               json.RawMessage      `json:"present"`
	TimeCost              json.RawMessage      `json:"timeCost"`
	MileCost              json.RawMessage      `json:"mileCost"`
	DeviceTrajectory      []deviceTrajectoryCO `json:"deviceTrajectory"`
	HasPaid               json.RawMessage      `json:"hasPaid"`
	DiscountCount         json.RawMessage      `json:"discountCount"`
	ActityFreeCount       json.RawMessage      `json:"actityFreeCount"`
	RidingCardCount       json.RawMessage      `json:"ridingCardCount"`
	IzFavorable           json.RawMessage      `json:"izFavorable"`
	IzDiscount            json.RawMessage      `json:"izDiscount"`
	IzActityFree          json.RawMessage      `json:"izActityFree"`
	DiscountType          json.RawMessage      `json:"discountType"`
	IzRandomDeduct        json.RawMessage      `json:"izRandomDeduct"`
	RandomCount           json.RawMessage      `json:"randomCount"`
	IzFreeN               json.RawMessage      `json:"izFreeN"`
	DispatchCost          json.RawMessage      `json:"dispatchCost"`
	HelmetPenalty         json.RawMessage      `json:"helmetPenalty"`
	StartPrice            json.RawMessage      `json:"startPrice"`
	IzRedPaket            json.RawMessage      `json:"izRedPaket"`
	RedPaketCarDeduct     json.RawMessage      `json:"redPaketCarDeduct"`
	RedPaketParkingDeduct json.RawMessage      `json:"redPaketParkingDeduct"`
	CarActivityId         *string              `json:"carActivityId"`
	ParkingActivityId     *string              `json:"parkingActivityId"`
}

// orderCOToLastDetail mirrors ConvertorHelper.copyProperties(OrderCO,
// CLastOrderDetailCO.class). Fields that exist only on the target stay null.
func orderCOToLastDetail(o *orderCO) cLastOrderDetailCO {
	return cLastOrderDetailCO{
		Id:                  longStr(o.Id),
		UserPin:             o.UserPin,
		Phone:               o.Phone,
		ServiceId:           longStr(o.ServiceId),
		OriginCost:          o.OriginCost,
		PayCost:             o.PayCost,
		ModifyCost:          o.ModifyCost,
		ModifyPayCost:       o.ModifyPayCost,
		ModifyDispatchCost:  o.ModifyDispatchCost,
		ModifyHelmetPenalty: o.ModifyHelmetPenalty,
		Mile:                o.Mile,
		RidingTime:          longStr(o.RidingTime),
		IzPaid:              o.IzPaid,
		IzComplained:        o.IzComplained,
		IzModified:          o.IzModified,
		IzRepair:            o.IzRepair,
		InvoiceId:           longStr(o.InvoiceId),
		Imei:                o.Imei,
		CarId:               o.CarId,
		PaidInfo:            o.PaidInfo,
		StartLat:            o.StartLat,
		StartLng:            o.StartLng,
		EndLat:              o.EndLat,
		EndLng:              o.EndLng,
		PayTime:             normalizeLDT(o.PayTime),
		StartTime:           normalizeLDT(o.StartTime),
		EndTime:             normalizeLDT(o.EndTime),
		Penalty:             o.Penalty,
		PenaltyType:         o.PenaltyType,
		Discount:            o.Discount,
		IzRidingCard:        o.IzRidingCard,
		RechargeCost:        o.RechargeCost,
		PresentCost:         o.PresentCost,
		PayType:             o.PayType,
		IzFavorable:         o.IzFavorable,
		IzDiscount:          o.IzDiscount,
		IzActityFree:        o.IzActityFree,
		IzFreeN:             o.IzFreeN,
		HelmetPenalty:       o.HelmetPenalty,
		StartPrice:          o.StartPrice,
	}
}

// applyItemToLastDetail mirrors BeanUtils.copyProperties(OrderItemCO, detail):
// every same-named field is overwritten unconditionally (nulls included);
// the caller restores the order id afterwards, exactly like Java.
func applyItemToLastDetail(d *cLastOrderDetailCO, it *orderItemCO) {
	d.Id = longStr(it.Id)
	d.BillingConfigId = longStr(it.BillingConfigId)
	d.Penalty = it.Penalty
	d.PenaltyType = it.PenaltyType
	d.DiscountCount = it.DiscountCount
	d.Discount = it.Discount
	d.IzDiscount = it.IzDiscount
	d.Deduction = it.Deduction
	d.IzRidingCard = it.IzRidingCard
	d.RidingCardCount = it.RidingCardCount
	d.IzActityFree = it.IzActityFree
	d.ActityFreeCount = it.ActityFreeCount
	d.IzFavorable = it.IzFavorable
	d.RechargeCost = it.RechargeCost
	d.PresentCost = it.PresentCost
	d.TimeCost = it.TimeCost
	d.MileCost = it.MileCost
	d.HasPaid = it.HasPaid
	d.DispatchCost = it.DispatchCost
	d.HelmetPenalty = it.HelmetPenalty
	d.DiscountType = it.DiscountType
	d.IzRandomDeduct = it.IzRandomDeduct
	d.RandomCount = it.RandomCount
	d.IzFreeN = it.IzFreeN
	d.StartPrice = it.StartPrice
	d.IzRedPaket = it.IzRedPaket
	d.RedPaketCarDeduct = it.RedPaketCarDeduct
	d.RedPaketParkingDeduct = it.RedPaketParkingDeduct
	d.CarActivityId = longStr(it.CarActivityId)
	d.ParkingActivityId = longStr(it.ParkingActivityId)
}

// cOrderDetailCO mirrors the Java COrderDetailCO field declaration order.
type cOrderDetailCO struct {
	Id                    *string              `json:"id"`
	UserPin               json.RawMessage      `json:"userPin"`
	Phone                 json.RawMessage      `json:"phone"`
	ServiceId             *string              `json:"serviceId"`
	BikeType              json.RawMessage      `json:"bikeType"`
	OriginCost            json.RawMessage      `json:"originCost"`
	PayCost               json.RawMessage      `json:"payCost"`
	ModifyCost            json.RawMessage      `json:"modifyCost"`
	ModifyPayCost         json.RawMessage      `json:"modifyPayCost"`
	ModifyDispatchCost    json.RawMessage      `json:"modifyDispatchCost"`
	ModifyHelmetPenalty   json.RawMessage      `json:"modifyHelmetPenalty"`
	Mile                  json.RawMessage      `json:"mile"`
	RidingTime            *string              `json:"ridingTime"`
	IzPaid                json.RawMessage      `json:"izPaid"`
	IzComplained          json.RawMessage      `json:"izComplained"`
	IzModified            json.RawMessage      `json:"izModified"`
	IzRepair              json.RawMessage      `json:"izRepair"`
	InvoiceId             *string              `json:"invoiceId"`
	Imei                  json.RawMessage      `json:"imei"`
	CarId                 json.RawMessage      `json:"carId"`
	PaidInfo              json.RawMessage      `json:"paidInfo"`
	StartLat              json.RawMessage      `json:"startLat"`
	StartLng              json.RawMessage      `json:"startLng"`
	EndLat                json.RawMessage      `json:"endLat"`
	EndLng                json.RawMessage      `json:"endLng"`
	PayTime               json.RawMessage      `json:"payTime"`
	StartTime             json.RawMessage      `json:"startTime"`
	EndTime               json.RawMessage      `json:"endTime"`
	BillingConfigId       *string              `json:"billingConfigId"`
	Penalty               json.RawMessage      `json:"penalty"`
	PenaltyType           json.RawMessage      `json:"penaltyType"`
	Discount              json.RawMessage      `json:"discount"`
	Deduction             json.RawMessage      `json:"deduction"`
	FreeOrderTime         json.RawMessage      `json:"freeOrderTime"`
	RechargeCost          json.RawMessage      `json:"rechargeCost"`
	PresentCost           json.RawMessage      `json:"presentCost"`
	PayType               json.RawMessage      `json:"payType"`
	Balance               json.RawMessage      `json:"balance"`
	Recharge              json.RawMessage      `json:"recharge"`
	Present               json.RawMessage      `json:"present"`
	TimeCost              json.RawMessage      `json:"timeCost"`
	MileCost              json.RawMessage      `json:"mileCost"`
	HasPaid               json.RawMessage      `json:"hasPaid"`
	DiscountCount         json.RawMessage      `json:"discountCount"`
	IzImpunity            json.RawMessage      `json:"izImpunity"`
	DeviceTrajectory      []deviceTrajectoryCO `json:"deviceTrajectory"`
	IzDiscount            json.RawMessage      `json:"izDiscount"`
	IzRidingCard          json.RawMessage      `json:"izRidingCard"`
	IzActityFree          json.RawMessage      `json:"izActityFree"`
	IzFavorable           json.RawMessage      `json:"izFavorable"`
	RidingCardCount       json.RawMessage      `json:"ridingCardCount"`
	ActityFreeCount       json.RawMessage      `json:"actityFreeCount"`
	IzRandomDeduct        json.RawMessage      `json:"izRandomDeduct"`
	RandomCount           json.RawMessage      `json:"randomCount"`
	IzFreeN               json.RawMessage      `json:"izFreeN"`
	DispatchCost          json.RawMessage      `json:"dispatchCost"`
	HelmetPenalty         json.RawMessage      `json:"helmetPenalty"`
	StartPrice            json.RawMessage      `json:"startPrice"`
	IzRedPaket            json.RawMessage      `json:"izRedPaket"`
	RedPaketCarDeduct     json.RawMessage      `json:"redPaketCarDeduct"`
	RedPaketParkingDeduct json.RawMessage      `json:"redPaketParkingDeduct"`
	CarActivityId         *string              `json:"carActivityId"`
	ParkingActivityId     *string              `json:"parkingActivityId"`
}

// orderCOToOrderDetail mirrors ConvertorHelper.copyProperties(OrderCO, COrderDetailCO.class).
func orderCOToOrderDetail(o *orderCO) cOrderDetailCO {
	return cOrderDetailCO{
		Id:                  longStr(o.Id),
		UserPin:             o.UserPin,
		Phone:               o.Phone,
		ServiceId:           longStr(o.ServiceId),
		OriginCost:          o.OriginCost,
		PayCost:             o.PayCost,
		ModifyCost:          o.ModifyCost,
		ModifyPayCost:       o.ModifyPayCost,
		ModifyDispatchCost:  o.ModifyDispatchCost,
		ModifyHelmetPenalty: o.ModifyHelmetPenalty,
		Mile:                o.Mile,
		RidingTime:          longStr(o.RidingTime),
		IzPaid:              o.IzPaid,
		IzComplained:        o.IzComplained,
		IzModified:          o.IzModified,
		IzRepair:            o.IzRepair,
		InvoiceId:           longStr(o.InvoiceId),
		Imei:                o.Imei,
		CarId:               o.CarId,
		PaidInfo:            o.PaidInfo,
		StartLat:            o.StartLat,
		StartLng:            o.StartLng,
		EndLat:              o.EndLat,
		EndLng:              o.EndLng,
		PayTime:             normalizeLDT(o.PayTime),
		StartTime:           normalizeLDT(o.StartTime),
		EndTime:             normalizeLDT(o.EndTime),
		Penalty:             o.Penalty,
		PenaltyType:         o.PenaltyType,
		Discount:            o.Discount,
		RechargeCost:        o.RechargeCost,
		PresentCost:         o.PresentCost,
		PayType:             o.PayType,
		IzDiscount:          o.IzDiscount,
		IzRidingCard:        o.IzRidingCard,
		IzActityFree:        o.IzActityFree,
		IzFavorable:         o.IzFavorable,
		IzFreeN:             o.IzFreeN,
		HelmetPenalty:       o.HelmetPenalty,
		StartPrice:          o.StartPrice,
	}
}

// applyItemToOrderDetail mirrors BeanUtils.copyProperties(OrderItemCO, detail)
// for COrderDetailCO (which additionally has izImpunity).
func applyItemToOrderDetail(d *cOrderDetailCO, it *orderItemCO) {
	d.Id = longStr(it.Id)
	d.BillingConfigId = longStr(it.BillingConfigId)
	d.Penalty = it.Penalty
	d.PenaltyType = it.PenaltyType
	d.IzImpunity = it.IzImpunity
	d.DiscountCount = it.DiscountCount
	d.Discount = it.Discount
	d.IzDiscount = it.IzDiscount
	d.Deduction = it.Deduction
	d.IzRidingCard = it.IzRidingCard
	d.RidingCardCount = it.RidingCardCount
	d.IzActityFree = it.IzActityFree
	d.ActityFreeCount = it.ActityFreeCount
	d.IzFavorable = it.IzFavorable
	d.RechargeCost = it.RechargeCost
	d.PresentCost = it.PresentCost
	d.TimeCost = it.TimeCost
	d.MileCost = it.MileCost
	d.HasPaid = it.HasPaid
	d.DispatchCost = it.DispatchCost
	d.HelmetPenalty = it.HelmetPenalty
	d.IzRandomDeduct = it.IzRandomDeduct
	d.RandomCount = it.RandomCount
	d.IzFreeN = it.IzFreeN
	d.StartPrice = it.StartPrice
	d.IzRedPaket = it.IzRedPaket
	d.RedPaketCarDeduct = it.RedPaketCarDeduct
	d.RedPaketParkingDeduct = it.RedPaketParkingDeduct
	d.CarActivityId = longStr(it.CarActivityId)
	d.ParkingActivityId = longStr(it.ParkingActivityId)
}

// discountType field exists on cOrderDetailCO via applyItemToOrderDetail;
// declared after opType-like fields in Java? -> kept inside struct above.

// ---------------------------------------------------------------------------
// Trajectory + wallet helpers shared by the detail endpoints
// ---------------------------------------------------------------------------

// errNPE marks a spot where Java would throw a NullPointerException.
var errNPE = errors.New("NullPointerException:null")

// fetchTrajectory performs one trajectory call with the Java try/catch
// semantics of OrderGatewayImpl.trajectoryHistory/trajectoryRealTime: any
// transport error, failed Result or undecodable body yields null.
func fetchTrajectory(ctx context.Context, cmdCtx *dto.CommandContext, path string, body interface{}) []deviceTrajectoryCO {
	result, err := rpc.ForwardCommand(ctx, rpc.ServiceDevicePaas, path, body, cmdCtx)
	if err != nil || result == nil || !result.Success || isNullRaw(result.Data) {
		return nil
	}
	var items []deviceTrajectoryIn
	if json.Unmarshal(result.Data, &items) != nil {
		return nil
	}
	return convertTrajectories(items)
}

// getOrderTrajectory mirrors OrderGatewayImpl.getOrderTrajectory.
// A null izPaid causes the Java NPE (orderCO.getIzPaid().intValue()) which is
// NOT caught -> surfaced as an error.
func getOrderTrajectory(ctx context.Context, cmdCtx *dto.CommandContext, o *orderCO) ([]deviceTrajectoryCO, error) {
	izPaid, ok := rawInt(o.IzPaid)
	if !ok {
		return nil, errNPE
	}
	if izPaid == 3 || izPaid == 4 {
		// historic trajectory; Java returns null without calling when endTime is null
		if isNullRaw(o.EndTime) {
			return nil, nil
		}
		qry := trajectoryHistoryQryBody{
			OrderId:   rawInt64Ptr(o.Id),
			StartTime: ldtToMillisPtr(o.StartTime),
			EndTime:   ldtToMillisPtr(o.EndTime),
		}
		return fetchTrajectory(ctx, cmdCtx, "/device/trajectory/history", &qry), nil
	}
	// realtime trajectory
	now := nowMillis()
	qry := trajectoryRealTimeQryBody{
		Imei:      rawStringPtr(o.Imei),
		StartTime: ldtToMillisPtr(o.StartTime),
		EndTime:   &now,
	}
	return fetchTrajectory(ctx, cmdCtx, "/device/trajectory/realTime", &qry), nil
}

// getOrdersTrajectory mirrors OrderGatewayImpl.getOrdersTrajectory:
// it builds the batch query (a null startTime/endTime on ANY order causes the
// uncaught Java NPE -> error) and swallows every downstream problem (try/catch
// around the API call) by returning nil.
func getOrdersTrajectory(ctx context.Context, cmdCtx *dto.CommandContext, orders []orderCO) (map[string]json.RawMessage, error) {
	batch := make([]trajectoryHistoryQryBody, 0, len(orders))
	for i := range orders {
		o := &orders[i]
		start, okS := parseLDTRaw(o.StartTime)
		end, okE := parseLDTRaw(o.EndTime)
		if !okS || !okE {
			// Java: orderCO.getStartTime()/getEndTime().toInstant(...) -> NPE
			return nil, errNPE
		}
		s, e := start.UnixMilli(), end.UnixMilli()
		batch = append(batch, trajectoryHistoryQryBody{OrderId: rawInt64Ptr(o.Id), StartTime: &s, EndTime: &e})
	}
	qry := trajectoryBatchQryBody{OrderBatch: batch}
	result, err := rpc.ForwardCommand(ctx, rpc.ServiceDevicePaas, "/device/trajectory/history/batch", &qry, cmdCtx)
	if err != nil || result == nil || !result.Success || isNullRaw(result.Data) {
		return nil, nil
	}
	var m map[string]json.RawMessage
	if json.Unmarshal(result.Data, &m) != nil {
		return nil, nil
	}
	return m, nil
}

// walletInfo mirrors UserWalletGatewayImpl.getUserWallet: any error, failed
// Result or null data yields an empty wallet (all-null fields).
type walletInfo struct {
	Balance  json.RawMessage `json:"balance"`
	Recharge json.RawMessage `json:"recharge"`
	Present  json.RawMessage `json:"present"`
}

func getUserWallet(ctx context.Context, cmdCtx *dto.CommandContext, pin *string) walletInfo {
	body := userBaseCmdBody{Pin: pin, IzUserService: false}
	result, err := rpc.ForwardCommand(ctx, rpc.ServiceAccount, "/wallet/info", &body, cmdCtx)
	w := walletInfo{}
	if err == nil && result != nil && result.Success && !isNullRaw(result.Data) {
		_ = json.Unmarshal(result.Data, &w)
	}
	return w
}

type orderFrozenCO struct {
	OrderId        *javacompat.LongStr `json:"orderId"`
	Mile           *int                `json:"mile"`
	FrozenTime     *javacompat.LongStr `json:"frozenTime"`
	IzOverTime     *bool               `json:"izOverTime"`
	IzOverDistance *bool               `json:"izOverDistance"`
	IzFrozen       *bool               `json:"izFrozen"`
	IzPaid         *int                `json:"izPaid"`
	PayCost        *int                `json:"payCost"`
}

func convertOrderFrozenCO(raw json.RawMessage) json.RawMessage {
	if javacompat.IsNullJSON(raw) {
		return raw
	}
	var co orderFrozenCO
	if json.Unmarshal(raw, &co) != nil {
		return raw
	}
	b, _ := json.Marshal(co)
	return b
}

type deviceLocationCO struct {
	CarId          *string              `json:"carId"`
	Imei           *string              `json:"imei"`
	Lat            *float64             `json:"lat"`
	Lng            *float64             `json:"lng"`
	Distance       *float64             `json:"distance"`
	RestBattery    *int                 `json:"restBattery"`
	RestMileage    *int                 `json:"restMileage"`
	ActivityId     *javacompat.LongStr  `json:"activityId"`
	ExpirationTime *javacompat.DateTime `json:"expirationTime"`
}

func convertDeviceLocationCOList(raw json.RawMessage) json.RawMessage {
	if javacompat.IsNullJSON(raw) {
		return raw
	}
	var list []deviceLocationCO
	if json.Unmarshal(raw, &list) != nil {
		return raw
	}
	if list == nil {
		list = []deviceLocationCO{}
	}
	b, _ := json.Marshal(list)
	return b
}

type blueToothTokenCo struct {
	Token *int `json:"token"`
}

func convertBlueToothTokenCo(raw json.RawMessage) json.RawMessage {
	if javacompat.IsNullJSON(raw) {
		return raw
	}
	var co blueToothTokenCo
	if json.Unmarshal(raw, &co) != nil {
		return raw
	}
	b, _ := json.Marshal(co)
	return b
}

func convertBoolean(raw json.RawMessage) json.RawMessage {
	if javacompat.IsNullJSON(raw) {
		return raw
	}
	var b bool
	if json.Unmarshal(raw, &b) != nil {
		return raw
	}
	out, _ := json.Marshal(b)
	return out
}

type cCalculateResult struct {
	OriginCost  *int                `json:"originCost"`
	Penalty     *int                `json:"penalty"`
	PenaltyType *int                `json:"penaltyType"`
	RidingTime  *javacompat.LongStr `json:"ridingTime"`
	Mile        *int                `json:"mile"`
}

// invoiceCOIn parses downstream InvoiceCO where orderIds is a JSON string
// (Java InvoiceCO.orderIds), not a native array.
type invoiceCOIn struct {
	ServiceId      *javacompat.LongStr  `json:"serviceId"`
	Id             *javacompat.LongStr  `json:"id"`
	CreatedAt      *javacompat.DateTime `json:"createdAt"`
	Phone          *string              `json:"phone"`
	Title          *string              `json:"title"`
	CompanyEin     *string              `json:"companyEin"`
	Type           *int                 `json:"type"`
	Amount         *int                 `json:"amount"`
	Email          *string              `json:"email"`
	Content        *string              `json:"content"`
	State          *int                 `json:"state"`
	Bank           *string              `json:"bank"`
	CompanyAddress *string              `json:"companyAddress"`
	BankAccount    *string              `json:"bankAccount"`
	CompanyPhone   *string              `json:"companyPhone"`
	OpResult       *int                 `json:"opResult"`
	Mark           *string              `json:"mark"`
	OpManPhone     *string              `json:"opManPhone"`
	UserPin        *string              `json:"userPin"`
	OpManPin       *string              `json:"opManPin"`
	DealTime       *javacompat.DateTime `json:"dealTime"`
	OrderIds       json.RawMessage      `json:"orderIds"`
}

func convertCInvoiceCO(raw json.RawMessage) (cInvoiceCO, bool) {
	var in invoiceCOIn
	if json.Unmarshal(raw, &in) != nil {
		return cInvoiceCO{}, false
	}
	orderIds := javacompat.ParseStringArrayField(in.OrderIds)
	return cInvoiceCO{
		ServiceId:      in.ServiceId,
		Id:             in.Id,
		CreatedAt:      in.CreatedAt,
		Phone:          in.Phone,
		Title:          in.Title,
		CompanyEin:     in.CompanyEin,
		Type:           in.Type,
		Amount:         in.Amount,
		Email:          in.Email,
		Content:        in.Content,
		State:          in.State,
		Bank:           in.Bank,
		CompanyAddress: in.CompanyAddress,
		BankAccount:    in.BankAccount,
		CompanyPhone:   in.CompanyPhone,
		OpResult:       in.OpResult,
		Mark:           in.Mark,
		OpManPhone:     in.OpManPhone,
		UserPin:        in.UserPin,
		OpManPin:       in.OpManPin,
		DealTime:       in.DealTime,
		OrderIds:       orderIds,
	}, true
}

type cInvoiceCO struct {
	CommandContext *struct{}            `json:"commandContext"`
	ServiceId      *javacompat.LongStr  `json:"serviceId"`
	Id             *javacompat.LongStr  `json:"id"`
	CreatedAt      *javacompat.DateTime `json:"createdAt"`
	Phone          *string              `json:"phone"`
	Title          *string              `json:"title"`
	CompanyEin     *string              `json:"companyEin"`
	Type           *int                 `json:"type"`
	Amount         *int                 `json:"amount"`
	Email          *string              `json:"email"`
	Content        *string              `json:"content"`
	State          *int                 `json:"state"`
	Bank           *string              `json:"bank"`
	CompanyAddress *string              `json:"companyAddress"`
	BankAccount    *string              `json:"bankAccount"`
	CompanyPhone   *string              `json:"companyPhone"`
	OpResult       *int                 `json:"opResult"`
	Mark           *string              `json:"mark"`
	OpManPhone     *string              `json:"opManPhone"`
	UserPin        *string              `json:"userPin"`
	OpManPin       *string              `json:"opManPin"`
	DealTime       *javacompat.DateTime `json:"dealTime"`
	OrderIds       []string             `json:"orderIds"`
}

type userTicketListCO struct {
	Id                 *string              `json:"id"`
	OrderId            *string              `json:"orderId"`
	UserPin            *string              `json:"userPin"`
	OpManPin           *string              `json:"opManPin"`
	CreatedAt          *javacompat.DateTime `json:"createdAt"`
	UpdatedAt          *javacompat.DateTime `json:"updatedAt"`
	DealAt             *javacompat.DateTime `json:"dealAt"`
	CarType            *int                 `json:"carType"`
	CarId              *string              `json:"carId"`
	Phone              *string              `json:"phone"`
	PayCost            *int                 `json:"payCost"`
	RechargeCost       *int                 `json:"rechargeCost"`
	PresentCost        *int                 `json:"presentCost"`
	RidingTime         *javacompat.LongStr  `json:"ridingTime"`
	StartTime          *javacompat.DateTime `json:"startTime"`
	EndTime            *javacompat.DateTime `json:"endTime"`
	Mile               *int                 `json:"mile"`
	Initiator          *int                 `json:"initiator"`
	State              *int                 `json:"state"`
	IzPaid             *int                 `json:"izPaid"`
	OpManPhone         *string              `json:"opManPhone"`
	OpManName          *string              `json:"opManName"`
	RefundCost         *int                 `json:"refundCost"`
	RefundRechargeCost *int                 `json:"refundRechargeCost"`
	RefundPresentCost  *int                 `json:"refundPresentCost"`
	OpType             *int                 `json:"opType"`
	UserReason         *string              `json:"userReason"`
}

type ladderItemCO struct {
	Count     *int `json:"count"`
	FromIndex *int `json:"fromIndex"`
	EndIndex  *int `json:"endIndex"`
	Price     *int `json:"price"`
}

type cBillingConfigCmd struct {
	javacompat.ClientDTOResponse
	Id                      *string              `json:"id"`
	ServiceId               *string              `json:"serviceId"`
	FreeDistance            *int                 `json:"freeDistance"`
	FreeTime                *int                 `json:"freeTime"`
	FreeTimes               *int                 `json:"freeTimes"`
	StartingDistance        *int                 `json:"startingDistance"`
	StartingTime            *int                 `json:"startingTime"`
	StartingPrice           *int                 `json:"startingPrice"`
	Discount                *float64             `json:"discount"`
	OverDistanceCostPerMter *int                 `json:"overDistanceCostPerMter"`
	TimeOutCostPerMin       *int                 `json:"timeOutCostPerMin"`
	AllowOutofService       *bool                `json:"allowOutofService"`
	AllowInNostop           *bool                `json:"allowInNostop"`
	AllowOutofParking       *bool                `json:"allowOutofParking"`
	TimeUnit                *int                 `json:"timeUnit"`
	DispatchCost            *int                 `json:"dispatchCost"`
	MaxPenaltyOutofParking  *int                 `json:"maxPenaltyOutofParking"`
	PenaltyInNostop         *int                 `json:"penaltyInNostop"`
	PenaltyOutofService     *int                 `json:"penaltyOutofService"`
	MostMoneyOneday         *int                 `json:"mostMoneyOneday"`
	Type                    *int                 `json:"type"`
	AllowInBanRiding        *bool                `json:"allowInBanRiding"`
	PenaltyInBanRiding      *int                 `json:"penaltyInBanRiding"`
	IzPopup                 *bool                `json:"izPopup"`
	IzEnable                *bool                `json:"izEnable"`
	IzAccumulate            *bool                `json:"izAccumulate"`
	LadderItem              []ladderItemCO       `json:"ladderItem"`
	UpdatedAt               *javacompat.DateTime `json:"updatedAt"`
	UpdatedPin              *string              `json:"updatedPin"`
}
