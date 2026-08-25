package model

import (
	"time"

	"ebike-analyze-go/internal/infrastructure/persistence/tenant"
	"gorm.io/gorm"
)

type UserStatistic struct {
	ID            int64     `gorm:"column:id"`
	ServiceID     int64     `gorm:"column:service_id"`
	StatisticDate time.Time `gorm:"column:statistic_date;type:date"`
	StatisticHour int       `gorm:"column:statistic_hour"`
	TotalNum      int       `gorm:"column:total_num"`
	ActiveNum     int       `gorm:"column:active_num"`
	CreateNum     int       `gorm:"column:create_num"`
	HourActiveNum int       `gorm:"column:hour_active_num"`
	TenantID      string    `gorm:"column:tenant_id"`
}

func (UserStatistic) TableName() string { return "t_user_statistic" }

type AgeStatistic struct {
	ID          int64     `gorm:"column:id"`
	ServiceID   int64     `gorm:"column:service_id"`
	Gender      int       `gorm:"column:gender"`
	IzRiding    int       `gorm:"column:iz_riding"`
	Scope       int       `gorm:"column:scope"`
	TotalNum    int       `gorm:"column:total_num"`
	StatisticAt time.Time `gorm:"column:statistic_at"`
	TenantID    string    `gorm:"column:tenant_id"`
}

func (AgeStatistic) TableName() string { return "t_user_age_statistic" }

type MemberStatistic struct {
	ID                       int64     `gorm:"column:id"`
	ServiceID                int64     `gorm:"column:service_id"`
	StatisticAt              time.Time `gorm:"column:statistic_at"`
	Auth                     int       `gorm:"column:auth"`
	AuthMember               int       `gorm:"column:auth_member"`
	AuthNoMember             int       `gorm:"column:auth_no_member"`
	AuthValidMember          int       `gorm:"column:auth_valid_member"`
	AuthInvalidMember        int       `gorm:"column:auth_invalid_member"`
	AuthHistoricalMember     int       `gorm:"column:auth_historical_member"`
	AuthHistoricalNoMember   int       `gorm:"column:auth_historical_no_member"`
	AuthCareer               int       `gorm:"column:auth_career"`
	AuthDeposit              int       `gorm:"column:auth_deposit"`
	AuthDepositCard          int       `gorm:"column:auth_deposit_card"`
	AuthFreeDeposit          int       `gorm:"column:auth_free_deposit"`
	NoAuth                   int       `gorm:"column:no_auth"`
	NoAuthMember             int       `gorm:"column:no_auth_member"`
	NoAuthNoMember           int       `gorm:"column:no_auth_no_member"`
	NoAuthValidMember        int       `gorm:"column:no_auth_valid_member"`
	NoAuthInvalidMember      int       `gorm:"column:no_auth_invalid_member"`
	NoAuthHistoricalMember   int       `gorm:"column:no_auth_historical_member"`
	NoAuthHistoricalNoMember int       `gorm:"column:no_auth_historical_no_member"`
	NoAuthCareer             int       `gorm:"column:no_auth_career"`
	NoAuthDeposit            int       `gorm:"column:no_auth_deposit"`
	NoAuthDepositCard        int       `gorm:"column:no_auth_deposit_card"`
	NoAuthFreeDeposit        int       `gorm:"column:no_auth_free_deposit"`
	TenantID                 string    `gorm:"column:tenant_id"`
}

func (MemberStatistic) TableName() string { return "t_user_member_statistic" }

type CarStatistic struct {
	ID                      int64     `gorm:"column:id"`
	ServiceID               int64     `gorm:"column:service_id"`
	CarID                   string    `gorm:"column:car_id"`
	Imei                    string    `gorm:"column:imei"`
	TenantID                string    `gorm:"column:tenant_id"`
	OrderCount              int       `gorm:"column:order_count"`
	OrderCost               int64     `gorm:"column:order_cost"`      // Java: Long
	RidingDistance          int64     `gorm:"column:riding_distance"` // Java: Long
	RidingTime              int64     `gorm:"column:riding_time"`
	DdMissOrder             int       `gorm:"column:dd_miss_order"`
	OperationMissOrderCount int       `gorm:"column:operation_miss_order_count"`
	ChangeBatteryCount      int       `gorm:"column:change_battery_count"`
	RepairCount             int       `gorm:"column:repair_count"`
	MoveCarCount            int       `gorm:"column:move_car_count"`
	DataTime                time.Time `gorm:"column:data_time"`
}

func (CarStatistic) TableName() string { return "t_car_statistics" }
func (c CarStatistic) ResolvedTable(tenantID string) string {
	return tenant.ResolveTableName(c.TableName(), tenantID)
}

type ServiceCarStatistic struct {
	ID                        int64     `gorm:"column:id"`
	ServiceID                 int64     `gorm:"column:service_id"`
	CanRent                   int       `gorm:"column:can_rent"`
	Booking                   int       `gorm:"column:booking"`
	Riding                    int       `gorm:"column:riding"`
	Parking                   int       `gorm:"column:parking"`
	Operation                 int       `gorm:"column:operation"`
	LowBattery                int       `gorm:"column:low_battery"`
	FreeTimeOneToThree        int       `gorm:"column:free_time_one_to_three"`
	FreeTimeThreeToSix        int       `gorm:"column:free_time_three_to_six"`
	FreeTimeSixToTwelve       int       `gorm:"column:free_time_six_to_twelve"`
	FreeTimeHalfOrOneDay      int       `gorm:"column:free_time_half_or_one_day"`
	FreeTimeOneOrTowDay       int       `gorm:"column:free_time_one_or_tow_day"`
	FreeTimeTowDayMore        int       `gorm:"column:free_time_tow_day_more"`
	VoltageZero               int       `gorm:"column:voltage_zero"`
	VoltageZeroToTwenty       int       `gorm:"column:voltage_zero_to_twenty"`
	VoltageTwentyToThirtyFive int       `gorm:"column:voltage_twenty_to_thirty_five"`
	VoltageThirtyFiveMore     int       `gorm:"column:voltage_thirty_five_more"`
	TenantID                  string    `gorm:"column:tenant_id"`
	CreatedPin                string    `gorm:"column:created_pin"`
	CreatedAt                 time.Time `gorm:"column:created_at"`
	UpdatedPin                string    `gorm:"column:updated_pin"`
	UpdatedAt                 time.Time `gorm:"column:updated_at"`
	Version                   int       `gorm:"column:version"`
	IzDel                     bool      `gorm:"column:iz_del"`
}

func (ServiceCarStatistic) TableName() string { return "t_service_car_statistics" }

type SiteStatistic struct {
	ID         int64     `gorm:"column:id"`
	ParkingID  int64     `gorm:"column:parking_id"`
	OrderCount int       `gorm:"column:order_count"`
	OrderCost  int64     `gorm:"column:order_cost"`
	StartTime  time.Time `gorm:"column:start_time"`
	EndTime    time.Time `gorm:"column:end_time"`
	TenantID   string    `gorm:"column:tenant_id"`
}

func (SiteStatistic) TableName() string { return "t_site_statistics" }

type ParkingStationStatistic struct {
	ID                      int64     `gorm:"column:id"`
	TenantID                string    `gorm:"column:tenant_id"`
	ServiceID               int64     `gorm:"column:service_id"`
	ParkingID               int64     `gorm:"column:parking_id"`
	DataTime                time.Time `gorm:"column:data_time"`
	CanRent                 int       `gorm:"column:can_rent"`
	Idle                    int       `gorm:"column:idle"`
	SiteOut                 int       `gorm:"column:site_out"`
	DdMissOrder             int       `gorm:"column:dd_miss_order"`
	Idle13                  int       `gorm:"column:idle1_3"`
	Idle36                  int       `gorm:"column:idle3_6"`
	Idle612                 int       `gorm:"column:idle6_12"`
	Idle1224                int       `gorm:"column:idle12_24"`
	Idle2448                int       `gorm:"column:idle24_48"`
	Idle48                  int       `gorm:"column:idle48"`
	Booking                 int       `gorm:"column:booking"`
	Operation               int       `gorm:"column:operation"`
	Alarm                   int       `gorm:"column:alarm"`
	Fault                   int       `gorm:"column:fault"`
	RideOrderCount          int       `gorm:"column:ride_order_count"`
	ReturnOrderCount        int       `gorm:"column:return_order_count"`
	RideOrderCost           int       `gorm:"column:ride_order_cost"`
	OperationMissOrderCount int       `gorm:"column:operation_miss_order_count"`
}

func (ParkingStationStatistic) TableName() string { return "t_parking_site_statistics" }
func (p ParkingStationStatistic) ResolvedTable(tenantID string) string {
	return tenant.ResolveTableName(p.TableName(), tenantID)
}

type Fence struct {
	ID             int64   `gorm:"column:id"`
	Name           string  `gorm:"column:name"`
	Type           int     `gorm:"column:type"`
	ServiceID      int64   `gorm:"column:service_id"`
	AreaSize       float64 `gorm:"column:area_size"`
	MaintainAreaID int64   `gorm:"column:maintain_area_id"`
	TenantID       string  `gorm:"column:tenant_id"`
}

func (Fence) TableName() string { return "t_fence" }

type RidingCardDetail struct {
	ID              int64     `gorm:"column:id"`
	PinID           string    `gorm:"column:pin_id"`
	PinPhone        string    `gorm:"column:pin_phone"`
	PinName         string    `gorm:"column:pin_name"`
	ConfigID        int64     `gorm:"column:config_id"`
	ServiceID       int64     `gorm:"column:service_id"`
	Type            int       `gorm:"column:type"`
	Channel         int       `gorm:"column:channel"`
	SysTradeNo      string    `gorm:"column:sys_trade_no"`
	MerchantTradeNo string    `gorm:"column:merchant_trade_no"`
	Amount          int       `gorm:"column:amount"`
	Name            string    `gorm:"column:name"`
	Duration        int       `gorm:"column:duration"`
	PaidAt          time.Time `gorm:"column:paid_at"`
	IzRefund        bool      `gorm:"column:iz_refund"`
	TenantID        string    `gorm:"column:tenant_id"`
}

func (RidingCardDetail) TableName() string { return "t_ebike_visual_riding_card_detail" }
func (r RidingCardDetail) ResolvedTable(tenantID string) string {
	return tenant.ResolveTableName(r.TableName(), tenantID)
}

type WalletDetail struct {
	ID              int64     `gorm:"column:id"`
	PinID           string    `gorm:"column:pin_id"`
	PinPhone        string    `gorm:"column:pin_phone"`
	PinName         string    `gorm:"column:pin_name"`
	ServiceID       int64     `gorm:"column:service_id"`
	Type            int       `gorm:"column:type"`
	Channel         int       `gorm:"column:channel"`
	SysTradeNo      string    `gorm:"column:sys_trade_no"`
	MerchantTradeNo string    `gorm:"column:merchant_trade_no"`
	Amount          int       `gorm:"column:amount"`
	RechargeAmount  int       `gorm:"column:recharge_amount"`
	PresentAmount   int       `gorm:"column:present_amount"`
	PaidAt          time.Time `gorm:"column:paid_at"`
	IzRefund        bool      `gorm:"column:iz_refund"`
	TenantID        string    `gorm:"column:tenant_id"`
}

func (WalletDetail) TableName() string { return "t_ebike_visual_wallet_detail" }
func (w WalletDetail) ResolvedTable(tenantID string) string {
	return tenant.ResolveTableName(w.TableName(), tenantID)
}

type DepositDetail struct {
	ID              int64     `gorm:"column:id"`
	PinID           string    `gorm:"column:pin_id"`
	PinPhone        string    `gorm:"column:pin_phone"`
	PinName         string    `gorm:"column:pin_name"`
	ConfigID        int64     `gorm:"column:config_id"`
	ServiceID       int64     `gorm:"column:service_id"`
	Type            int       `gorm:"column:type"`
	Channel         int       `gorm:"column:channel"`
	SysTradeNo      string    `gorm:"column:sys_trade_no"`
	MerchantTradeNo string    `gorm:"column:merchant_trade_no"`
	Amount          int       `gorm:"column:amount"`
	Name            string    `gorm:"column:name"`
	Duration        int       `gorm:"column:duration"`
	IzCard          int       `gorm:"column:iz_card"`
	PaidAt          time.Time `gorm:"column:paid_at"`
	IzRefund        bool      `gorm:"column:iz_refund"`
	TenantID        string    `gorm:"column:tenant_id"`
}

func (DepositDetail) TableName() string { return "t_ebike_visual_deposit_card_detail" }
func (d DepositDetail) ResolvedTable(tenantID string) string {
	return tenant.ResolveTableName(d.TableName(), tenantID)
}

func ScopedDB(db *gorm.DB, tenantID string) *gorm.DB {
	return tenant.WithTenant(db, tenantID)
}
