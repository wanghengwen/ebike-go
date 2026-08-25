package dto

type MemberStatisticQuery struct {
	CommandContext *CommandContext `json:"commandContext"`
	ServiceID      []int64         `json:"serviceId"`
}

type AgeStatisticCmd struct {
	CommandContext *CommandContext `json:"commandContext"`
	ServiceIds     []int64         `json:"serviceIds" binding:"required"`
}

type AllUserStatisticCmd struct {
	CommandContext *CommandContext `json:"commandContext"`
	ServiceID      []int64         `json:"serviceId" binding:"required"`
	Begin          *DateOnlyValue  `json:"begin"`
	End            *DateOnlyValue  `json:"end"`
}

type UserStatisticHourQry struct {
	CommandContext *CommandContext `json:"commandContext"`
	ServiceID      int64           `json:"serviceId" binding:"required"`
	QueryDate      *DateOnlyValue  `json:"queryDate" binding:"required"`
}

// TodayUserStatisticCo mirrors Java TodayUserStatisticCo: field JSON names are
// remapped via @JsonProperty ("new"/"active"/"date"), not the Java field names.
type TodayUserStatisticCo struct {
	NewCou    int           `json:"new"`
	ActiveCou int           `json:"active"`
	QueryDate DateOnlyValue `json:"date"`
	Total     int           `json:"total"`
}

// AllUserStatisticCo mirrors Java AllUserStatisticCo @JsonProperty remaps.
type AllUserStatisticCo struct {
	List       []TodayUserStatisticCo `json:"new_data"`
	ActiveUser int                    `json:"active_user"`
	NewUser    int                    `json:"new_user"`
	TotalUser  int64                  `json:"total_user"`
}

type AgeMap struct {
	OneAgeScope   int `json:"16_to_22"`
	TwoAgeScope   int `json:"23_to_28"`
	ThreeAgeScope int `json:"29_to_40"`
	FourAgeScope  int `json:"41_to_60"`
	Total         int `json:"total"`
}

type RidingQualification struct {
	Male    AgeMap `json:"male"`
	Female  AgeMap `json:"female"`
	Unknown AgeMap `json:"unknown"`
	Total   int    `json:"total"`
}

type AgeStatisticCO struct {
	HaveRidingQualification RidingQualification `json:"have_riding_qualification"`
	NoRidingQualification   RidingQualification `json:"no_riding_qualification"`
}

type GenderMapV2 struct {
	Male    int `json:"male"`
	Female  int `json:"female"`
	Unknown int `json:"unknown"`
}

type AgeStatisticCOV2 struct {
	OneAgeScope   GenderMapV2 `json:"oneAgeScope"`
	TwoAgeScope   GenderMapV2 `json:"twoAgeScope"`
	ThreeAgeScope GenderMapV2 `json:"threeAgeScope"`
	FourAgeScope  GenderMapV2 `json:"fourAgeScope"`
	FiveAgeScope  GenderMapV2 `json:"fiveAgeScope"`
	SixAgeScope   GenderMapV2 `json:"sixAgeScope"`
}

type UserStatisticHourCo struct {
	StatisticHour int `json:"statisticHour"`
	ActiveNum     int `json:"activeNum"`
	CreateNum     int `json:"createNum"`
}
