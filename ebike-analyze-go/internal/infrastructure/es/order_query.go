package es

import (
	"ebike-analyze-go/internal/api/dto"
	"ebike-analyze-go/internal/common/dateutil"
)

// BuildOrderQuery mirrors Java OrderQueryServiceImpl.buildQueryBuilder.
func BuildOrderQuery(cmd *dto.OrderQueryCmd) M {
	var must []M

	if cmd.IzPaid != nil {
		must = append(must, Term("izPaid", *cmd.IzPaid))
	} else {
		must = append(must, Terms("izPaid", []int{3, 4}))
	}
	if cmd.ServiceID != nil {
		must = append(must, Term("serviceId", *cmd.ServiceID))
	}
	if cmd.EndMaintainArea != nil {
		must = append(must, Term("endMaintainArea", *cmd.EndMaintainArea))
	}
	if cmd.OrderID != nil {
		must = append(must, Term("orderId", *cmd.OrderID))
	}
	if cmd.CarID != "" {
		must = append(must, Term("carId", cmd.CarID))
	}
	if cmd.Imei != "" {
		must = append(must, Term("imei", cmd.Imei))
	}
	if len(cmd.UserPinList) > 0 {
		must = append(must, Terms("userPin", cmd.UserPinList))
	}
	if cmd.UserPin != "" {
		must = append(must, Term("userPin", cmd.UserPin))
	}
	if cmd.IzComplained != nil {
		if *cmd.IzComplained == 0 {
			must = append(must, Term("izComplained", -1))
		} else {
			must = append(must, Terms("izComplained", []int{0, 1}))
		}
	}
	if cmd.IzStandardReturn != nil {
		if *cmd.IzStandardReturn == 0 {
			must = append(must, BoolMustNot(Term("penaltyType", 1101)))
		} else {
			must = append(must, Term("penaltyType", 1101))
		}
	}
	if len(cmd.StartTime) == 2 {
		must = append(must, RangeGteLte("startTime",
			dateutil.TimestampFormat(cmd.StartTime[0]),
			dateutil.TimestampFormat(cmd.StartTime[1])))
	}
	if len(cmd.EndTime) == 2 {
		must = append(must, RangeGteLte("endTime",
			dateutil.TimestampFormat(cmd.EndTime[0]),
			dateutil.TimestampFormat(cmd.EndTime[1])))
	}
	if len(cmd.StartTime) == 1 && len(cmd.EndTime) == 1 {
		must = append(must, RangeGteLte("endTime",
			dateutil.TimestampFormat(cmd.StartTime[0]),
			dateutil.TimestampFormat(cmd.EndTime[0])))
	}
	if cmd.Equal != nil && cmd.Mile != nil {
		if *cmd.Equal == 1 {
			must = append(must, RangeGte("ridingDistance", *cmd.Mile))
		} else if *cmd.Equal == 2 {
			must = append(must, RangeLte("ridingDistance", *cmd.Mile))
		}
	}
	if cmd.CostEqual != nil && cmd.PayCost != nil {
		if *cmd.CostEqual == 1 {
			must = append(must, RangeGte("payCost", *cmd.PayCost))
		} else if *cmd.CostEqual == 2 {
			must = append(must, RangeLte("payCost", *cmd.PayCost))
		}
	}
	if cmd.TimeEqual != nil && cmd.RidingTime != nil {
		if *cmd.TimeEqual == 1 {
			must = append(must, RangeGte("ridingTime", *cmd.RidingTime))
		} else if *cmd.TimeEqual == 2 {
			must = append(must, RangeLte("ridingTime", *cmd.RidingTime))
		}
	}

	return ConstantScore(BoolQuery(must, nil, nil))
}

// BuildUserQuery mirrors Java UserQueryServiceImpl.buildQueryBuilder.
// Note: Java's buildQueryBuilder does NOT call cmd.setByType() itself - that
// expansion (qualificationType -> izByDefault/izByDeposit/...) happens upstream
// in UserGatewayImpl (a different microservice) before the RPC call reaches
// this service. Calling it here would diverge from direct-API-call behavior.
func BuildUserQuery(cmd *dto.UserQueryCmd) M {
	var must, should []M

	if cmd.Phone != "" {
		must = append(must, Prefix("phone", cmd.Phone))
	}
	if cmd.AuthName != "" {
		must = append(must, Term("authName", cmd.AuthName))
	}
	if cmd.IzRiding != nil {
		must = append(must, Term("izRiding", *cmd.IzRiding))
	}
	if cmd.DepositState != nil {
		must = append(must, Term("depositState", *cmd.DepositState))
	}
	if cmd.RidingState != nil {
		switch *cmd.RidingState {
		case 1:
			must = append(must, Term("izAuth", 0))
		case 2:
			must = append(must, Term("izAuth", 1))
		case 8:
			now := dateutil.NowFormatted()
			must = append(must, RangeGt("blacklistExpire", now))
			must = append(must, Exists("blacklistExpire"))
		case 3, 4, 5, 6:
			must = append(must, Term("ridingState", *cmd.RidingState))
		}
	}
	if cmd.IzByType != nil && *cmd.IzByType == 1 {
		should = append(should, Term("tenantId", "-1"))
		if cmd.IzByDefault != nil && *cmd.IzByDefault == 1 {
			should = append(should, Term("izRiding", 1))
		}
		if cmd.IzByDeposit != nil && *cmd.IzByDeposit == 1 {
			should = append(should, Term("depositState", 1))
		}
		if cmd.IzByDepositCard != nil && *cmd.IzByDepositCard == 1 {
			now := dateutil.NowFormatted()
			should = append(should, RangeGt("depositCardExpire", now))
			// Java: exists("depositCardExpire") is added to the outer `must`, not `should`,
			// so it is always enforced regardless of other should clauses.
			must = append(must, Exists("depositCardExpire"))
		}
		if cmd.IzByCareer != nil && *cmd.IzByCareer == 1 {
			should = append(should, Term("careerState", 1))
		}
		if cmd.IzByWePayScore != nil && *cmd.IzByWePayScore == 1 {
			should = append(should, Term("izWxScorePay", 1))
		}
	}
	if cmd.StartTime != nil && cmd.EndTime != nil {
		must = append(must, RangeGteLte("createdAt",
			dateutil.LocalDateTimeFormat(*cmd.StartTime),
			dateutil.LocalDateTimeFormat(*cmd.EndTime)))
	}
	if cmd.LoginStartTime != nil && cmd.LoginEndTime != nil {
		must = append(must, Exists("lastLoginTime"))
		must = append(must, RangeGteLte("lastLoginTime",
			dateutil.LocalDateTimeFormat(*cmd.LoginStartTime),
			dateutil.LocalDateTimeFormat(*cmd.LoginEndTime)))
	}
	if len(cmd.ServiceID) > 0 {
		must = append(must, Terms("serviceId", cmd.ServiceID))
		must = append(must, Exists("serviceId"))
	}

	return ConstantScore(BoolQuery(must, should, nil))
}

func OrderIndex(tenantID string) string {
	return "order_" + tenantID
}

func UserIndex(tenantID string) string {
	return "user_" + tenantID
}
