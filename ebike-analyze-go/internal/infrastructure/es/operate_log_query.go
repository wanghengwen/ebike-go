package es

import (
	"strings"

	"ebike-analyze-go/internal/api/dto"
)

const OperateLogIndex = "ebike-operate-log"

// BuildOperateLogQuery mirrors Java OperateLogServiceImpl.buildQueryBuilder.
func BuildOperateLogQuery(cmd *dto.OperateLogCmd, tenantID string) M {
	var must []M

	must = append(must, Term("tenantId", tenantID))

	if len(cmd.Pins) > 0 {
		must = append(must, Terms("pin", cmd.Pins))
	}
	if values := splitCSV(cmd.EventType); len(values) > 0 {
		must = append(must, Terms("eventType", values))
	}
	if values := splitCSV(cmd.Platform); len(values) > 0 {
		must = append(must, Terms("platform", values))
	}
	if cmd.TraceID != "" {
		must = append(must, Term("traceId", cmd.TraceID))
	}
	if cmd.CarID != "" {
		must = append(must, Term("carId", cmd.CarID))
	}
	if cmd.Imei != "" {
		must = append(must, Term("imei", cmd.Imei))
	}

	// Match Java RangeQueryBuilder: omit null bounds instead of emitting "gte":null.
	if cmd.StartTime != nil || cmd.EndTime != nil {
		rangeClause := M{}
		if cmd.StartTime != nil {
			rangeClause["gte"] = *cmd.StartTime
		}
		if cmd.EndTime != nil {
			rangeClause["lte"] = *cmd.EndTime
		}
		must = append(must, M{"range": M{"@timestamp": rangeClause}})
	}

	return ConstantScore(BoolMust(must...))
}

func splitCSV(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
