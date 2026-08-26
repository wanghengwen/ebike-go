package dto

// TrajectoryRealTimeQry mirrors the Java TrajectoryRealTimeQry.
type TrajectoryRealTimeQry struct {
	Imei           string          `json:"imei"`
	StartTime      FlexInt64       `json:"startTime"`
	EndTime        FlexInt64       `json:"endTime"`
	CommandContext *CommandContext `json:"commandContext,omitempty"`
}

// TrajectoryHistoryQry mirrors the Java TrajectoryHistoryQry.
type TrajectoryHistoryQry struct {
	OrderId        FlexInt64       `json:"orderId"`
	StartTime      FlexInt64       `json:"startTime"`
	EndTime        FlexInt64       `json:"endTime"`
	CommandContext *CommandContext `json:"commandContext,omitempty"`
}

// TrajectoryHistoryBatchQry mirrors the Java TrajectoryHistoryBatchQry.
type TrajectoryHistoryBatchQry struct {
	OrderBatch     []TrajectoryHistoryQry `json:"orderBatch"`
	CommandContext *CommandContext        `json:"commandContext,omitempty"`
}

// MetricQry mirrors the Java MetricQry (imei may be a carId; type is coordinate type).
type MetricQry struct {
	Imei           string          `json:"imei"`
	StartTime      FlexInt64       `json:"startTime"`
	EndTime        FlexInt64       `json:"endTime"`
	Type           *int            `json:"type,omitempty"`
	CommandContext *CommandContext `json:"commandContext,omitempty"`
}

// TrajectoryDbCmd mirrors the Java TrajectoryDbCmd (trajectory save-to-db).
type TrajectoryDbCmd struct {
	OrderId        FlexInt64       `json:"orderId"`
	Imei           string          `json:"imei"`
	StartTime      FlexInt64       `json:"startTime"`
	EndTime        FlexInt64       `json:"endTime"`
	CommandContext *CommandContext `json:"commandContext,omitempty"`
}

// DeviceTrajectoryCo mirrors the Java DeviceTrajectoryCo. Fields are nullable
// (Double/Long) and serialized as-is to match the worker passthrough.
type DeviceTrajectoryCo struct {
	Lng       *float64 `json:"lng"`
	Lat       *float64 `json:"lat"`
	Speed     *float64 `json:"speed"`
	Course    *float64 `json:"course"`
	Timestamp *int64   `json:"timestamp"`
}

// TrajectoryDistanceCo mirrors the Java TrajectoryDistanceCo.
type TrajectoryDistanceCo struct {
	Distance float64 `json:"distance"`
}

// The metric reshape DTOs mirror com.xyy.ebike.device.paas.monitor.dto.*.

// MetricCourse mirrors monitor.dto.Course.
type MetricCourse struct {
	Course []int    `json:"course"`
	Time   []string `json:"time"`
}

// MetricSpeed mirrors monitor.dto.Speed.
type MetricSpeed struct {
	Speed []int    `json:"speed"`
	Time  []string `json:"time"`
}

// MetricGsm mirrors monitor.dto.Gsm.
type MetricGsm struct {
	Gsm  []int    `json:"gsm"`
	Time []string `json:"time"`
}

// MetricGps mirrors monitor.dto.Gps.
type MetricGps struct {
	Timestamp int64   `json:"timestamp"`
	Lng       float64 `json:"lng"`
	Lat       float64 `json:"lat"`
	Course    int     `json:"course"`
	Speed     int     `json:"speed"`
}

// MetricVoltage mirrors monitor.dto.Voltage (BigDecimal scaled to 2 decimals).
type MetricVoltage struct {
	Voltage []float64 `json:"voltage"`
	Time    []string  `json:"time"`
}
