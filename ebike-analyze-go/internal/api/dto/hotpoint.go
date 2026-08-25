package dto

type HotPointCmd struct {
	CommandContext *CommandContext `json:"commandContext"`
	Start          DateTimeValue   `json:"start"`
	End            DateTimeValue   `json:"end"`
	ServiceID      int64           `json:"serviceId"`
}

type HotPointCo struct {
	Lat   float64 `json:"lat"`
	Lng   float64 `json:"lng"`
	Count int     `json:"count"`
}
