package dto

type TrajectoryCmd struct {
	Imei      string `json:"imei" binding:"required,numeric,min=10,max=15"`
	StartTime int64  `json:"startTime" binding:"required,gt=0"`
	EndTime   int64  `json:"endTime" binding:"required,gt=0"`
	Type      *int   `json:"type" binding:"required"` // Pointer to allow 0 (WGS84) value
}

type OrderTrajectoryCmd struct {
	OrderId   string `json:"orderId" binding:"required,max=32"`
	Imei      string `json:"imei" binding:"omitempty,numeric,min=10,max=15"`
	StartTime int64  `json:"startTime"`
	EndTime   int64  `json:"endTime"`
	Type      *int   `json:"type" binding:"required"`
}

// BatchOrderTrajectoryItemCmd mirrors Java batch nested item: type is not @Valid-cascaded on batch API.
type BatchOrderTrajectoryItemCmd struct {
	OrderId   string `json:"orderId" binding:"required,max=32"`
	Imei      string `json:"imei" binding:"omitempty,numeric,min=10,max=15"`
	StartTime int64  `json:"startTime"`
	EndTime   int64  `json:"endTime"`
	Type      *int   `json:"type"`
}

type BatchOrderTrajectoryCmd struct {
	Type                    *int                          `json:"type" binding:"required"`
	OrderTrajectoryRequests []BatchOrderTrajectoryItemCmd `json:"orderTrajectoryRequests" binding:"required,dive"`
}

type MetricCmd struct {
	Imei      string `json:"imei" binding:"required,numeric,len=15"` // mirrors Java @NullableDigit(length=15)
	StartTime int64  `json:"startTime" binding:"required,gt=0"`
	EndTime   int64  `json:"endTime" binding:"required,gt=0"`
	Type      *int   `json:"type" binding:"required"`
}

type TrajectoryCo struct {
	Lng        float64 `json:"lng"`
	Lat        float64 `json:"lat"`
	Timestamp  int64   `json:"timestamp"`
	Speed      float32 `json:"speed"`
	Course     float32 `json:"course"`
	TotalMiles *int    `json:"totalMiles"` // GPS: 0 or value; itinerary: null (Java never sets on getOrderTrajectory)
}

type TrajectoryDistanceCo struct {
	Distance float64 `json:"distance"`
}

type MetricCo struct {
	Lng       float64 `json:"lng"`
	Lat       float64 `json:"lat"`
	Timestamp int64   `json:"timestamp"`
	Speed     int     `json:"speed"`
	Course    int     `json:"course"`
	Voltage   int     `json:"voltage"`
	Gsm       int     `json:"gsm"`
}

type CmdQueryCmd struct {
	Imei      string `json:"imei" binding:"required,numeric,len=15"`
	StartTime int64  `json:"startTime" binding:"required,gt=0"`
	EndTime   int64  `json:"endTime" binding:"required,gt=0"`
	Dir       string `json:"dir"` // mirrors Java CmdQueryCmd.dir (optional filter)
}

type TrajectoryPoint struct {
	Lng       float64 `json:"lng"`
	Lat       float64 `json:"lat"`
	Timestamp int64   `json:"timestamp"`
	Speed     float32 `json:"speed"`
	Course     float32 `json:"course"`
}

type SaveOrderTrajectoryCmd struct {
	OrderId   string           `json:"orderId" binding:"required,max=32"`
	Imei      string           `json:"imei" binding:"required,numeric,min=10,max=15"`
	StartTime int64            `json:"startTime" binding:"required,gt=0"`
	EndTime   int64            `json:"endTime" binding:"required,gt=0"`
	Gps       *TrajectoryPoint `json:"gps"`
}

type SaveMovingTrajectoryCmd struct {
	MovingEBikeTaskId string `json:"movingEBikeTaskId" binding:"required,max=32"`
	Imei              string `json:"imei" binding:"required,numeric,len=15"`
	StartTime         int64  `json:"startTime" binding:"required,gt=0"`
	EndTime           int64  `json:"endTime" binding:"required,gt=0"`
}

type DeviceCommandCmd struct {
	Timestamp int64       `json:"timestamp" binding:"required,gt=0"`
	Imei      string      `json:"imei" binding:"required,numeric,min=10,max=15"`
	Cmd       interface{} `json:"cmd"`
	Dir       string      `json:"dir"`
	MsgId     string      `json:"msgId"`
}
