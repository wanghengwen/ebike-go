package dto

// OperateLogCmd mirrors Java com.xyy.ebike.analyze.api.dto.OperateLogCmd.
type OperateLogCmd struct {
	CommandContext *CommandContext `json:"commandContext"`
	PageNum        int             `json:"pageNum"`
	PageSize       int             `json:"pageSize"`
	StartTime      *int64          `json:"startTime"`
	EndTime        *int64          `json:"endTime"`
	Pins           []string        `json:"pins"`
	EventType      string          `json:"eventType"`
	Platform       string          `json:"platform"`
	TenantID       string          `json:"tenantId"`
	TraceID        string          `json:"traceId"`
	CarID          string          `json:"carId"`
	Imei           string          `json:"imei"`
}

// OperationLogCo mirrors Java com.xyy.ebike.management.api.clientobject.OperationLogCo.
// Nullable string fields use *string so missing/null ES values serialize as null (Java parity).
type OperationLogCo struct {
	TenantID  *string             `json:"tenantId"`
	TraceID   *string             `json:"traceId"`
	Pin       *string             `json:"pin"`
	CarID     *string             `json:"carId"`
	Imei      *string             `json:"imei"`
	Time      *LocalDateTimeValue `json:"time"`
	EventType *string             `json:"eventType"`
	EventName *string             `json:"eventName"`
	Name      *string             `json:"name"`
	Phone     *string             `json:"phone"`
	Content   *string             `json:"content"`
	Result    interface{}         `json:"result"`
}
