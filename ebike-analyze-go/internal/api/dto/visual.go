package dto

import "time"

type OrderQueryPageCmd struct {
	CommandContext *CommandContext `json:"commandContext"`
	ServiceID      int64           `json:"serviceId"`
	Start          *time.Time      `json:"start"`
	End            *time.Time      `json:"end"`
	Name           string          `json:"name"`
	Phone          string          `json:"phone"`
	Type           *int            `json:"type"`
	PageNum        int             `json:"pageNum"`
	PageSize       int             `json:"pageSize"`
	SearchCount    *bool           `json:"searchCount"`
	Orders         []OrderItem     `json:"orders"`
	LastRecordID   string          `json:"lastRecordId"`
}

// RidingCardCO mirrors Java api.co.RidingCardCO. `start`/`end` are always nil
// in the Java response too, since RidingCardDetailDO/RidingCardEntity carry no
// such columns; kept here only for API-contract completeness.
type RidingCardCO struct {
	ID              int64          `json:"id"`
	PinID           string         `json:"pinId"`
	PinPhone        string         `json:"pinPhone"`
	PinName         string         `json:"pinName"`
	ConfigID        int64          `json:"configId"`
	ServiceID       int64          `json:"serviceId"`
	Type            int            `json:"type"`
	Channel         int            `json:"channel"`
	SysTradeNo      string         `json:"sysTradeNo"`
	MerchantTradeNo string         `json:"merchantTradeNo"`
	Amount          int            `json:"amount"`
	Name            string         `json:"name"`
	Duration        int            `json:"duration"`
	PaidAt          DateTimeValue  `json:"paidAt"`
	IzRefund        bool           `json:"izRefund"`
	Start           *DateTimeValue `json:"start"`
	End             *DateTimeValue `json:"end"`
}

// WalletCO mirrors Java api.co.WalletCO.
type WalletCO struct {
	ID              int64         `json:"id"`
	PinID           string        `json:"pinId"`
	PinPhone        string        `json:"pinPhone"`
	PinName         string        `json:"pinName"`
	ServiceID       int64         `json:"serviceId"`
	Type            int           `json:"type"`
	Channel         int           `json:"channel"`
	SysTradeNo      string        `json:"sysTradeNo"`
	MerchantTradeNo string        `json:"merchantTradeNo"`
	Amount          int           `json:"amount"`
	RechargeAmount  int           `json:"rechargeAmount"`
	PresentAmount   int           `json:"presentAmount"`
	PaidAt          DateTimeValue `json:"paidAt"`
	IzRefund        bool          `json:"izRefund"`
}

// DepositCO mirrors Java api.co.DepositCO.
type DepositCO struct {
	ID              int64         `json:"id"`
	PinID           string        `json:"pinId"`
	PinPhone        string        `json:"pinPhone"`
	PinName         string        `json:"pinName"`
	ConfigID        int64         `json:"configId"`
	ServiceID       int64         `json:"serviceId"`
	Type            int           `json:"type"`
	Channel         int           `json:"channel"`
	SysTradeNo      string        `json:"sysTradeNo"`
	MerchantTradeNo string        `json:"merchantTradeNo"`
	Amount          int           `json:"amount"`
	Name            string        `json:"name"`
	Duration        int           `json:"duration"`
	IzCard          int           `json:"izCard"`
	PaidAt          DateTimeValue `json:"paidAt"`
	IzRefund        bool          `json:"izRefund"`
}
