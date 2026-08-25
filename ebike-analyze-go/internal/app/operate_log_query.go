package app

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"ebike-analyze-go/internal/api/dto"
	"ebike-analyze-go/internal/infrastructure/es"
	"ebike-analyze-go/internal/pkg/config"

	"go.uber.org/zap"
)

type OperateLogService struct{}

// Page mirrors Java OperateLogServiceImpl.page.
func (s *OperateLogService) Page(ctx context.Context, cmd *dto.OperateLogCmd, tenantID string) *dto.PageDTO[dto.OperationLogCo] {
	if cmd == nil || es.Client == nil {
		return nil
	}

	pageNum, pageSize := cmd.PageNum, cmd.PageSize
	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	q := es.BuildOperateLogQuery(cmd, tenantID)
	sort := []es.M{{"@timestamp": es.M{"order": "desc"}}}
	res, err := es.SearchAll(ctx, []string{es.OperateLogIndex}, q, pageNum, pageSize, sort, config.GlobalConfig.ES.TrackTotalHitsUpTo)
	if err != nil || res == nil {
		zap.L().Info("ES返回空", zap.String("indexName", es.OperateLogIndex), zap.Error(err))
		return nil
	}

	list := make([]dto.OperationLogCo, 0, len(res.Hits))
	for _, hit := range res.Hits {
		co, mapErr := mapHitToOperationLogCo(hit.Source)
		if mapErr != nil {
			zap.L().Warn("operate log hit map failed", zap.Error(mapErr), zap.String("id", hit.ID))
			continue
		}
		list = append(list, co)
	}

	page := dto.NewPageDTO(pageNum, pageSize, res.TotalHits, list)
	return &page
}

func mapHitToOperationLogCo(src map[string]interface{}) (dto.OperationLogCo, error) {
	var co dto.OperationLogCo
	if src == nil {
		return co, nil
	}

	// Prefer explicit time; fall back to @timestamp used by ES/filebeat indices.
	normalized := make(map[string]interface{}, len(src)+1)
	for k, v := range src {
		normalized[k] = v
	}
	if _, ok := normalized["time"]; !ok {
		if ts, ok := normalized["@timestamp"]; ok {
			normalized["time"] = ts
		}
	}
	if t, ok := normalized["time"]; ok {
		if converted, err := coerceLocalDateTime(t); err == nil {
			normalized["time"] = converted
		} else {
			// Keep other fields; leave time unset (closer to Java Fastjson partial parse).
			zap.L().Warn("operate log time coerce failed", zap.Error(err), zap.Any("time", t))
			delete(normalized, "time")
		}
	}

	b, err := json.Marshal(normalized)
	if err != nil {
		return co, err
	}
	if err := json.Unmarshal(b, &co); err != nil {
		return co, err
	}
	return co, nil
}

func coerceLocalDateTime(v interface{}) (string, error) {
	switch t := v.(type) {
	case string:
		if t == "" {
			return "", fmt.Errorf("empty time")
		}
		if parsed, err := parseFlexibleTime(t); err == nil {
			return parsed.In(time.Local).Format("2006-01-02T15:04:05"), nil
		}
		return t, nil
	case float64:
		return formatEpoch(int64(t)), nil
	case int64:
		return formatEpoch(t), nil
	case json.Number:
		n, err := t.Int64()
		if err != nil {
			return "", err
		}
		return formatEpoch(n), nil
	default:
		return "", fmt.Errorf("unsupported time type %T", v)
	}
}

func parseFlexibleTime(str string) (time.Time, error) {
	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02 15:04:05.000",
	}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, str, time.Local); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid datetime: %s", str)
}

func formatEpoch(n int64) string {
	// Heuristic: seconds vs millis (OperateLogCmd documents second timestamps).
	if n < 1_000_000_000_000 {
		return time.Unix(n, 0).In(time.Local).Format("2006-01-02T15:04:05")
	}
	sec := n / 1000
	nsec := (n % 1000) * int64(time.Millisecond)
	return time.Unix(sec, nsec).In(time.Local).Format("2006-01-02T15:04:05")
}
