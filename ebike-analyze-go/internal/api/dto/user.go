package dto

import "time"

// UserQueryCmd mirrors Java UserQueryCmd. IzAuth/IzCareer/IzNoRidding/TagID
// are present in Java but unused by buildQueryBuilder; kept for API-contract
// completeness only.
type UserQueryCmd struct {
	CommandContext    *CommandContext `json:"commandContext"`
	ServiceID         []int64         `json:"serviceId"`
	Phone             string          `json:"phone"`
	AuthName          string          `json:"authName"`
	IzAuth            *int            `json:"izAuth"`
	IzRiding          *int            `json:"izRiding"`
	IzCareer          *int            `json:"izCareer"`
	StartTime         *time.Time      `json:"startTime"`
	EndTime           *time.Time      `json:"endTime"`
	IzNoRidding       *int            `json:"izNoRidding"`
	LoginStartTime    *time.Time      `json:"loginStartTime"`
	LoginEndTime      *time.Time      `json:"loginEndTime"`
	TagID             *int64          `json:"tagId"`
	DepositState      *int            `json:"depositState"`
	RidingState       *int            `json:"ridingState"`
	QualificationType []int           `json:"qualificationType"`
	IzByDefault       *int            `json:"izByDefault"`
	IzByDeposit       *int            `json:"izByDeposit"`
	IzByDepositCard   *int            `json:"izByDepositCard"`
	IzByCareer        *int            `json:"izByCareer"`
	IzByWePayScore    *int            `json:"izByWePayScore"`
	IzByType          *int            `json:"izByType"`
}

func (c *UserQueryCmd) SetByType() {
	if len(c.QualificationType) == 0 {
		return
	}
	for _, t := range c.QualificationType {
		switch t {
		case 0:
			v := 1
			c.IzByDefault = &v
		case 1:
			v := 1
			c.IzByDeposit = &v
		case 2:
			v := 1
			c.IzByDepositCard = &v
		case 3:
			v := 1
			c.IzByCareer = &v
		case 4:
			v := 1
			c.IzByWePayScore = &v
		}
	}
	v := 1
	c.IzByType = &v
}
