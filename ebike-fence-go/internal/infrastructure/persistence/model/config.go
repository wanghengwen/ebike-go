package model

import (
	"database/sql"
	"time"
)

// BaseConfig mirrors Java BaseDO columns on config tables.
type BaseConfig struct {
	TenantID   string        `gorm:"column:tenant_id"`
	CreatedPin string        `gorm:"column:created_pin"`
	CreatedAt  time.Time     `gorm:"column:created_at"`
	UpdatedPin string        `gorm:"column:updated_pin"`
	UpdatedAt  time.Time     `gorm:"column:updated_at"`
	Version    sql.NullInt32 `gorm:"column:version"`
	IzDel      sql.NullBool  `gorm:"column:iz_del"`
}

// ConfigBackcar maps t_config_backcar.
type ConfigBackcar struct {
	ID int64 `gorm:"column:id;primaryKey"`
	BaseConfig
	AllowOutofService          sql.NullBool    `gorm:"column:allow_outof_service"`
	AllowInNostop              sql.NullBool    `gorm:"column:allow_in_nostop"`
	AllowOutofParking          sql.NullBool    `gorm:"column:allow_outof_parking"`
	AllowInBanRiding           sql.NullBool    `gorm:"column:allow_in_ban_riding"`
	DispatchCost               sql.NullInt32   `gorm:"column:dispatch_cost"`
	PenaltyInNostop            sql.NullInt32   `gorm:"column:penalty_in_nostop"`
	PenaltyOutofService        sql.NullInt32   `gorm:"column:penalty_outof_service"`
	PenaltyInBanRiding         sql.NullInt32   `gorm:"column:penalty_in_ban_riding"`
	IzCivilizationRemind       sql.NullBool    `gorm:"column:iz_civilization_remind"`
	CoefficientOfDifficult     sql.NullFloat64 `gorm:"column:coefficient_of_difficult"`
	BufferDistance             sql.NullFloat64 `gorm:"column:buffer_distance"`
	ParkingBackcar             sql.NullInt32   `gorm:"column:parking_backcar"`
	NoParkingBackcar           sql.NullInt32   `gorm:"column:no_parking_backcar"`
	BanRidingBackcar           sql.NullInt32   `gorm:"column:ban_riding_backcar"`
	IzHelmetLock               sql.NullBool    `gorm:"column:iz_helmet_lock"`
	IzHelmetPhoto              sql.NullBool    `gorm:"column:iz_helmet_photo"`
	IzBeacon                   sql.NullBool    `gorm:"column:iz_beacon"`
	ServiceID                  int64           `gorm:"column:service_id"`
	IzHelmetRemovalDetection   sql.NullBool    `gorm:"column:iz_helmet_removal_detection"`
	IzHelmetWearDetection      sql.NullBool    `gorm:"column:iz_helmet_wear_detection"`
	IzHelmetAutoRepair         sql.NullBool    `gorm:"column:iz_helmet_auto_repair"`
	PowerOff                   sql.NullInt32   `gorm:"column:power_off"`
	HelmetPenalty              sql.NullInt32   `gorm:"column:helmet_penalty"`
	IzHelmetReign              sql.NullBool    `gorm:"column:iz_helmet_reign"`
	IzCameraBackcar            sql.NullBool    `gorm:"column:iz_camera_backcar"`
	IzCameraDirectionalBackcar sql.NullBool    `gorm:"column:iz_camera_directional_backcar"`
	IzCameraPointBackcar       sql.NullBool    `gorm:"column:iz_camera_point_backcar"`
	CameraImpunityCount        sql.NullInt32   `gorm:"column:camera_impunity_count"`
	IzFullPileNoStop           sql.NullInt32   `gorm:"column:iz_full_pile_no_stop"`
	IzUseOtherParking          sql.NullBool    `gorm:"column:iz_use_other_parking"`
	BeaconEffect               sql.NullInt32   `gorm:"column:beacon_effect"`
	RfidEffect                 sql.NullInt32   `gorm:"column:rfid_effect"`
	DirectionEffect            sql.NullInt32   `gorm:"column:direction_effect"`
	FootSupportEffect          sql.NullInt32   `gorm:"column:foot_support_effect"`
	CameraEffect               sql.NullInt32   `gorm:"column:camera_effect"`
	CameraAngle                sql.NullInt32   `gorm:"column:camera_angle"`
	IzStandardReturn           sql.NullInt32   `gorm:"column:iz_standard_return"`
}

func (ConfigBackcar) TableName() string { return "t_config_backcar" }

// ConfigPay maps t_config_pay.
type ConfigPay struct {
	ID int64 `gorm:"column:id;primaryKey"`
	BaseConfig
	ServiceID                 int64          `gorm:"column:service_id"`
	IzBalanceEnoughReturnBike sql.NullBool   `gorm:"column:iz_balance_enough_return_bike"`
	IzMakeupBalancePay        sql.NullBool   `gorm:"column:iz_makeup_balance_pay"`
	IzNotifyUnpaidOrder       sql.NullBool   `gorm:"column:iz_notify_unpaid_order"`
	NotifyInterval            sql.NullInt32  `gorm:"column:notify_interval"`
	TimesUpperBound           sql.NullInt32  `gorm:"column:times_upper_bound"`
	RemindWay                 sql.NullString `gorm:"column:remind_way"`
	IzAutoRefundAfterOrder    sql.NullBool   `gorm:"column:iz_auto_refund_after_order"`
}

func (ConfigPay) TableName() string { return "t_config_pay" }

// ConfigBaseItem maps t_config_base_item.
type ConfigBaseItem struct {
	ID int64 `gorm:"column:id;primaryKey"`
	BaseConfig
	ServiceID              int64          `gorm:"column:service_id"`
	OfflineTicketJudgeTime sql.NullInt32  `gorm:"column:offline_ticket_judge_time"`
	MsgSign                sql.NullString `gorm:"column:msg_sign"`
	MsgCode                sql.NullString `gorm:"column:msg_code"`
	IzOnAbnormalMovement   sql.NullBool   `gorm:"column:iz_on_abnormal_movement"`
	IzOnBatteryRemoval     sql.NullBool   `gorm:"column:iz_on_battery_removal"`
	IzOfflineTicket        sql.NullBool   `gorm:"column:iz_offline_ticket"`
	IzAutoRepairTicket     sql.NullBool   `gorm:"column:iz_auto_repair_ticket"`
	IzAutoSwapBattery      sql.NullBool   `gorm:"column:iz_auto_swap_battery"`
	SwapBatteryThreshold   sql.NullInt32  `gorm:"column:swap_battery_threshold"`
	IzOpenInvoice          sql.NullBool   `gorm:"column:iz_open_invoice"`
	IzTempUnlock           sql.NullBool   `gorm:"column:iz_temp_unlock"`
	TempUnlockTime         sql.NullInt32  `gorm:"column:temp_unlock_time"`
	DashboardSwitch        sql.NullBool   `gorm:"column:dashboard_switch"`
	DashboardStartTime     sql.NullString `gorm:"column:dashboard_start_time"`
	DashboardEndTime       sql.NullString `gorm:"column:dashboard_end_time"`
	IzCanInitiativeRepair  sql.NullBool   `gorm:"column:iz_can_initiative_repair"`
	IzCanAfterRidingRepair sql.NullBool   `gorm:"column:iz_can_after_riding_repair"`
	CanRepairCon           sql.NullInt32  `gorm:"column:can_repair_con"`
	IzHelmetAbnormalRiding sql.NullBool   `gorm:"column:iz_helmet_abnormal_riding"`
	IzEnableMoveAlarm      sql.NullBool   `gorm:"column:iz_enable_move_alarm"`
	MoveAlarmTime          sql.NullInt32  `gorm:"column:move_alarm_time"`
	MoveAlarmCount         sql.NullInt32  `gorm:"column:move_alarm_count"`
	MoveAlarmDistance      sql.NullInt32  `gorm:"column:move_alarm_distance"`
	IzCreateShortOrder     sql.NullBool   `gorm:"column:iz_create_short_order"`
	ShortOrderDuration     sql.NullInt32  `gorm:"column:short_order_duration"`
	ShortOrderCon          sql.NullInt32  `gorm:"column:short_order_con"`
	UserTicketPhotoWays    sql.NullString `gorm:"column:user_ticket_photo_ways"`
	IzWithdraw             sql.NullBool   `gorm:"column:iz_withdraw"`
}

func (ConfigBaseItem) TableName() string { return "t_config_base_item" }

// ConfigUseCar maps t_config_use_car.
type ConfigUseCar struct {
	ID int64 `gorm:"column:id;primaryKey"`
	BaseConfig
	ServiceID                       int64          `gorm:"column:service_id"`
	RechargeBeforeUse               sql.NullBool   `gorm:"column:recharge_before_use"`
	RechargeVisibleRange            sql.NullInt32  `gorm:"column:recharge_visible_range"`
	RechargeByRegister              sql.NullBool   `gorm:"column:recharge_by_register"`
	RechargeByTags                  sql.NullBool   `gorm:"column:recharge_by_tags"`
	RechargeTagIds                  sql.NullString `gorm:"column:recharge_tag_ids"`
	RechargeCost                    sql.NullInt32  `gorm:"column:recharge_cost"`
	IzStopService                   sql.NullBool   `gorm:"column:iz_stop_service"`
	StopTimeStart                   sql.NullString `gorm:"column:stop_time_start"`
	StopTimeEnd                     sql.NullString `gorm:"column:stop_time_end"`
	IzAutoRecovery                  sql.NullBool   `gorm:"column:iz_auto_recovery"`
	RecoveryData                    sql.NullTime   `gorm:"column:recovery_data"`
	StopServiceNotice               sql.NullString `gorm:"column:stop_service_notice"`
	IzOnCertification               sql.NullBool   `gorm:"column:iz_on_certification"`
	IzOnUseCar                      sql.NullBool   `gorm:"column:iz_on_use_car"`
	RecognitionDegree               sql.NullInt32  `gorm:"column:recognition_degree"`
	EffectiveTime                   sql.NullInt32  `gorm:"column:effective_time"`
	IzRidingStopTrigger             sql.NullBool   `gorm:"column:iz_riding_stop_trigger"`
	IzParkingTriggerReturnBike      sql.NullBool   `gorm:"column:iz_parking_trigger_return_bike"`
	RidingStopTime                  sql.NullInt32  `gorm:"column:riding_stop_time"`
	RidingStopEvent                 sql.NullInt32  `gorm:"column:riding_stop_event"`
	ParkingTime                     sql.NullInt32  `gorm:"column:parking_time"`
	RemindWay                       sql.NullString `gorm:"column:remind_way"`
	IzBeacon                        sql.NullBool   `gorm:"column:iz_beacon"`
	OutServiceAreaAutoLock          sql.NullInt32  `gorm:"column:out_service_area_auto_lock"`
	OutServiceAreaAutoLockRemindWay sql.NullString `gorm:"column:out_service_area_auto_lock_remind_way"`
	NearLine                        sql.NullInt32  `gorm:"column:near_line"`
	IzOpenSaddleOverloadMonitor     sql.NullBool   `gorm:"column:iz_open_saddle_overload_monitor"`
	OverloadRemind                  sql.NullInt32  `gorm:"column:overload_remind"`
	IzRemoteUnlock                  sql.NullBool   `gorm:"column:iz_remote_unlock"`
	HelmetConfig                    sql.NullString `gorm:"column:helmet_config"`
	IzOrderNotice                   sql.NullBool   `gorm:"column:iz_order_notice"`
	NoticeRidingTime                sql.NullInt32  `gorm:"column:notice_riding_time"`
	OrderRemindWay                  sql.NullString `gorm:"column:order_remind_way"`
	MinAge                          sql.NullInt32  `gorm:"column:min_age"`
	MaxAge                          sql.NullInt32  `gorm:"column:max_age"`
	HideCarConfig                   sql.NullString `gorm:"column:hide_car_config"`
	IzAuth                          sql.NullBool   `gorm:"column:iz_auth"`
}

func (ConfigUseCar) TableName() string { return "t_config_use_car" }

// CreditScoreConfig maps t_config_credit_score.
type CreditScoreConfig struct {
	ID int64 `gorm:"column:id;primaryKey"`
	BaseConfig
	IzCreditScore       sql.NullBool    `gorm:"column:iz_credit_score"`
	Score               sql.NullFloat64 `gorm:"column:score"`
	WarnScore           sql.NullFloat64 `gorm:"column:warn_score"`
	NoRiddingScore      sql.NullFloat64 `gorm:"column:no_ridding_score"`
	FirstNoRiddingDays  sql.NullInt32   `gorm:"column:first_no_ridding_days"`
	SecondNoRiddingDays sql.NullInt32   `gorm:"column:second_no_ridding_days"`
	MoreNoRiddingDays   sql.NullInt32   `gorm:"column:more_no_ridding_days"`
	AddScore            sql.NullFloat64 `gorm:"column:add_score"`
	AddScoreUpperLimit  sql.NullFloat64 `gorm:"column:add_score_upper_limit"`
	RemindWay           sql.NullString  `gorm:"column:remind_way"`
}

func (CreditScoreConfig) TableName() string { return "t_config_credit_score" }

// ParkApplyConfig maps t_config_park_apply.
type ParkApplyConfig struct {
	ID int64 `gorm:"column:id;primaryKey"`
	BaseConfig
	IzApplyPark sql.NullInt32 `gorm:"column:iz_apply_park"`
	ServiceID   int64         `gorm:"column:service_id"`
}

func (ParkApplyConfig) TableName() string { return "t_config_park_apply" }

// PushRidingCardConfig maps t_config_push_riding_card.
type PushRidingCardConfig struct {
	ID int64 `gorm:"column:id;primaryKey"`
	BaseConfig
	IzOpen    sql.NullBool `gorm:"column:iz_open"`
	ServiceID int64        `gorm:"column:service_id"`
}

func (PushRidingCardConfig) TableName() string { return "t_config_push_riding_card" }

// AdConfig maps t_ad_config.
type AdConfig struct {
	ID int64 `gorm:"column:id;primaryKey"`
	BaseConfig
	IzOn      string `gorm:"column:iz_on"`
	ServiceID int64  `gorm:"column:service_id"`
}

func (AdConfig) TableName() string { return "t_ad_config" }

// AlarmContact maps t_alarm_contact.
type AlarmContact struct {
	ID int64 `gorm:"column:id;primaryKey"`
	BaseConfig
	ServiceID   int64          `gorm:"column:service_id"`
	Type        sql.NullInt32  `gorm:"column:type"`
	Name        sql.NullString `gorm:"column:name"`
	Phone       sql.NullString `gorm:"column:phone"`
	StartTime   sql.NullString `gorm:"column:start_time"`
	EndTime     sql.NullString `gorm:"column:end_time"`
	NotifyType  sql.NullString `gorm:"column:notify_type"`
	TriggerArea sql.NullInt32  `gorm:"column:trigger_area"`
}

func (AlarmContact) TableName() string { return "t_alarm_contact" }

// BigScreen maps t_config_bigscreen.
type BigScreen struct {
	ID int64 `gorm:"column:id;primaryKey"`
	BaseConfig
	DisplayCoefficient sql.NullFloat64 `gorm:"column:display_coefficient"`
}

func (BigScreen) TableName() string { return "t_config_bigscreen" }

// RidingPermission maps t_ride_bike_permission.
type RidingPermission struct {
	ID int64 `gorm:"column:id;primaryKey"`
	BaseConfig
	ServiceID                    int64          `gorm:"column:service_id"`
	IzDefaultPermit              sql.NullBool   `gorm:"column:iz_default_permit"`
	IzDeposit                    sql.NullBool   `gorm:"column:iz_deposit"`
	Deposit                      sql.NullInt32  `gorm:"column:deposit"`
	IzCareer                     sql.NullBool   `gorm:"column:iz_career"`
	Career                       sql.NullString `gorm:"column:career"`
	IzDepositCard                sql.NullBool   `gorm:"column:iz_deposit_card"`
	IzWxScorePayDeposited        sql.NullBool   `gorm:"column:iz_wx_score_pay_deposited"`
	IzWxScorePayNoPassword       sql.NullBool   `gorm:"column:iz_wx_score_pay_no_password"`
	IzFreeDeposit                sql.NullBool   `gorm:"column:iz_free_deposit"`
	IzZhimaPayAfterUseDeposited  sql.NullBool   `gorm:"column:iz_zhima_pay_after_use_deposited"`
	IzZhimaPayAfterUseNoPassword sql.NullBool   `gorm:"column:iz_zhima_pay_after_use_no_password"`
}

func (RidingPermission) TableName() string { return "t_ride_bike_permission" }

// ConfigProtocol maps t_config_protocol.
type ConfigProtocol struct {
	ID int64 `gorm:"column:id;primaryKey"`
	BaseConfig
	ServiceID int64  `gorm:"column:service_id"`
	Title     string `gorm:"column:title"`
	Type      int    `gorm:"column:type"`
	Content   string `gorm:"column:content"`
}

func (ConfigProtocol) TableName() string { return "t_config_protocol" }

// ResourceManagement maps t_config_resource_management.
type ResourceManagement struct {
	ID int64 `gorm:"column:id;primaryKey"`
	BaseConfig
	PageCode      sql.NullInt32  `gorm:"column:page_code"`
	Type          sql.NullInt32  `gorm:"column:type"`
	Status        sql.NullInt32  `gorm:"column:status"`
	Name          sql.NullString `gorm:"column:name"`
	StartTime     sql.NullTime   `gorm:"column:start_time"`
	EndTime       sql.NullTime   `gorm:"column:end_time"`
	IzLimitTime   sql.NullBool   `gorm:"column:iz_limit_time"`
	AdvID         sql.NullString `gorm:"column:adv_id"`
	Title         sql.NullString `gorm:"column:title"`
	AppID         sql.NullString `gorm:"column:appid"`
	SkipURL       sql.NullString `gorm:"column:skip_url"`
	Params        sql.NullString `gorm:"column:params"`
	ImgURL        sql.NullString `gorm:"column:img_url"`
	Sort          sql.NullInt32  `gorm:"column:sort"`
	ServiceID     int64          `gorm:"column:service_id"`
	ExposureCount sql.NullInt64  `gorm:"column:exposure_count"`
	ClickCount    sql.NullInt64  `gorm:"column:click_count"`
	ClickPerson   sql.NullInt64  `gorm:"column:click_person"`
}

func (ResourceManagement) TableName() string { return "t_config_resource_management" }

// FenceTag maps t_fence_tag (tagId column in Java).
type FenceTag struct {
	TagID int64 `gorm:"column:id;primaryKey"`
	BaseConfig
	TagName  string       `gorm:"column:tag_name"`
	IzEnable sql.NullBool `gorm:"column:iz_enable"`
}

func (FenceTag) TableName() string { return "t_fence_tag" }
