package app

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"ebike-analyze-go/internal/api/dto"
	"ebike-analyze-go/internal/infrastructure/es"
	"ebike-analyze-go/internal/pkg/config"
)

type OrderQueryService struct{}

func (s *OrderQueryService) SelectOrderCount(ctx context.Context, cmd *dto.OrderQueryCmd, tenantID string) int64 {
	if cmd == nil || tenantID == "" {
		return 0
	}
	if es.Client == nil {
		return 0
	}
	q := es.BuildOrderQuery(cmd)
	n, err := es.SearchCount(ctx, []string{es.OrderIndex(tenantID)}, q, config.GlobalConfig.ES.TrackTotalHitsUpTo)
	if err != nil {
		return 0
	}
	return n
}

func (s *OrderQueryService) SelectOrderList(ctx context.Context, cmd *dto.OrderQueryCmd, tenantID string) dto.PageDTO[dto.OrderCO] {
	if cmd == nil || tenantID == "" || es.Client == nil {
		return dto.NewPageDTO(1, 10, 0, []dto.OrderCO{})
	}
	pageNum, pageSize := cmd.PageNum, cmd.PageSize
	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	q := es.BuildOrderQuery(cmd)
	sort := []es.M{{"endTime": es.M{"order": "desc"}}}
	res, err := es.SearchAll(ctx, []string{es.OrderIndex(tenantID)}, q, pageNum, pageSize, sort, config.GlobalConfig.ES.TrackTotalHitsUpTo)
	if err != nil {
		return dto.NewPageDTO(pageNum, pageSize, 0, []dto.OrderCO{})
	}
	list := make([]dto.OrderCO, 0, len(res.Hits))
	for _, h := range res.Hits {
		co := mapHitToOrderCO(h.Source)
		list = append(list, co)
	}
	return dto.NewPageDTO(pageNum, pageSize, res.TotalHits, list)
}

func (s *OrderQueryService) SelectOrderAnalyze(ctx context.Context, cmd *dto.OrderQueryCmd, tenantID string) dto.OrderLocationAnalyzeCO {
	if cmd == nil || tenantID == "" || es.Client == nil {
		return dto.EmptyOrderLocationAnalyze()
	}
	q := es.BuildOrderQuery(cmd)
	var all []es.Hit
	var searchAfter []interface{}
	for {
		res, err := es.SearchLocation(ctx, []string{es.OrderIndex(tenantID)}, q, searchAfter)
		if err != nil || res == nil || len(res.Hits) == 0 {
			break
		}
		all = append(all, res.Hits...)
		last := res.Hits[len(res.Hits)-1]
		searchAfter = last.SortValues
	}
	oList := make([]dto.OrderLatLng, 0, len(all))
	pCount := 0
	for _, h := range all {
		lng := fmt.Sprint(h.Source["endLng"])
		lat := fmt.Sprint(h.Source["endLat"])
		endParkingID := int64(0)
		if v, ok := h.Source["endParkingId"]; ok && v != nil {
			switch n := v.(type) {
			case float64:
				endParkingID = int64(n)
			case string:
				endParkingID, _ = strconv.ParseInt(n, 10, 64)
			}
		}
		p := endParkingID == 0
		if p {
			pCount++
		}
		pVal := p
		oList = append(oList, dto.OrderLatLng{L: lng + "," + lat, P: &pVal})
	}
	return dto.OrderLocationAnalyzeCO{
		Total:  len(all),
		PCount: pCount,
		NCount: len(all) - pCount,
		OList:  oList,
	}
}

// DeleteByQuery mirrors Java OrderQueryServiceImpl.deleteByQuery +
// ElasticsearchServiceImpl.deleteByQuery: ES IO failures are caught in the ES
// layer (return null) and the service maps null → 0L. Go therefore swallows
// errors and returns 0 so API callers see data:0 rather than success:false.
func (s *OrderQueryService) DeleteByQuery(ctx context.Context, cmd *dto.OrderQueryCmd, tenantID string) int64 {
	if es.Client == nil || cmd == nil {
		return 0
	}
	q := es.BuildOrderQuery(cmd)
	n, err := es.DeleteByQuery(ctx, []string{es.OrderIndex(tenantID)}, q)
	if err != nil {
		return 0
	}
	return n
}

func mapHitToOrderCO(src map[string]interface{}) dto.OrderCO {
	var co dto.OrderCO
	b, _ := json.Marshal(src)
	_ = json.Unmarshal(b, &co)
	// Java: orderCO.setId(orderCO.getOrderId()); orderCO.setMile(orderCO.getRidingDistance());
	if n, ok := toInt64(src["orderId"]); ok {
		co.ID = n
		co.OrderID = n
	}
	if n, ok := toInt64(src["ridingDistance"]); ok {
		co.Mile = int(n)
	}
	return co
}

// toInt64 tolerates ES source values arriving as float64, json.Number, string, or int64.
func toInt64(v interface{}) (int64, bool) {
	switch n := v.(type) {
	case float64:
		return int64(n), true
	case int64:
		return n, true
	case int:
		return int64(n), true
	case json.Number:
		i, err := n.Int64()
		return i, err == nil
	case string:
		i, err := strconv.ParseInt(n, 10, 64)
		return i, err == nil
	default:
		return 0, false
	}
}

type UserQueryService struct{}

func (s *UserQueryService) SelectUserCount(ctx context.Context, cmd *dto.UserQueryCmd, tenantID string) int64 {
	if cmd == nil || tenantID == "" || es.Client == nil {
		return 0
	}
	q := es.BuildUserQuery(cmd)
	n, err := es.SearchCount(ctx, []string{es.UserIndex(tenantID)}, q, config.GlobalConfig.ES.TrackTotalHitsUpTo)
	if err != nil {
		return 0
	}
	return n
}
