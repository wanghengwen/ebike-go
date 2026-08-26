package dto

// JobStateQry mirrors the Java JobStateQry (jobId + Query base/commandContext).
type JobStateQry struct {
	JobId          string          `json:"jobId"`
	CommandContext *CommandContext `json:"commandContext,omitempty"`
}

// BluetoothCo mirrors the Java BluetoothCo. Java serializes with Jackson ALWAYS,
// so both fields are emitted (null when absent).
type BluetoothCo struct {
	Token *int    `json:"token"`
	Name  *string `json:"name"`
}

// JobSucCo mirrors the Java JobSucCo.
type JobSucCo struct {
	JobSuc bool `json:"jobSuc"`
}

// DefendCo mirrors the Java DefendCo. The gateway result carries "lon" which the
// Java convertor copies into lng (see mapDefend).
type DefendCo struct {
	Defend      *int     `json:"defend"`
	Lng         *float64 `json:"lng"`
	Lat         *float64 `json:"lat"`
	Timestamp   *int64   `json:"timestamp"`
	Event       *int     `json:"event"`
	TBeaconAddr *string  `json:"tBeaconAddr"`
	TBeaconId   *string  `json:"tBeaconId"`
	TBeaconSOC  *int     `json:"tBeaconSOC"`
	TBeaconVsn  *string  `json:"tBeaconVsn"`
	// Output key is lowercase "rfid": ebike-device-paas serializes responses with
	// Spring's default Jackson (no FastJsonHttpMessageConverter is registered in
	// the app or the xyy starters), so the fastjson @JSONField(name="RFID") on the
	// Java DefendCo is ignored and Jackson emits the bean name "rfid". The upstream
	// gateway still sends "RFID"; Go's case-insensitive unmarshal matches it.
	Rfid      *DefendRfidCo      `json:"rfid"`
	KickStand *DefendKickStandCo `json:"kickStand"`
	Camera    *DefendCameraCo    `json:"camera"`
}

// DefendRfidCo mirrors the Java DefendRfidCo.
type DefendRfidCo struct {
	Event *int    `json:"event"`
	Id    *string `json:"id"`
}

// DefendKickStandCo mirrors the Java DefendKickStandCo.
type DefendKickStandCo struct {
	Event    *int    `json:"event"`
	Type     *int    `json:"type"`
	MagState *int    `json:"magState"`
	RfState  *int    `json:"rfState"`
	Version  *string `json:"version"`
	CardID   *string `json:"cardID"`
}

// DefendCameraCo mirrors the Java DefendCameraCo.
type DefendCameraCo struct {
	Event     *int    `json:"event"`
	Version   *string `json:"version"`
	ErrorCode *int    `json:"errorCode"`
}
