package fence

import (
	"encoding/json"

	"ebike-service-client-go/internal/pkg/javacompat"
)

type resourceManagementCO struct {
	Id            *javacompat.LongStr  `json:"id"`
	PageCode      *int                 `json:"pageCode"`
	Type          *int                 `json:"type"`
	Status        *int                 `json:"status"`
	Name          *string              `json:"name"`
	StartTime     *javacompat.DateTime `json:"startTime"`
	EndTime       *javacompat.DateTime `json:"endTime"`
	IzLimitTime   *bool                `json:"izLimitTime"`
	AdvId         *string              `json:"advId"`
	Title         *string              `json:"title"`
	Appid         *string              `json:"appid"`
	SkipUrl       *string              `json:"skipUrl"`
	Params        *string              `json:"params"`
	ImgUrl        *string              `json:"imgUrl"`
	Sort          *int                 `json:"sort"`
	ServiceId     *javacompat.LongStr  `json:"serviceId"`
	ExposureCount *javacompat.LongStr  `json:"exposureCount"`
	ClickCount    *javacompat.LongStr  `json:"clickCount"`
	ClickPerson   *javacompat.LongStr  `json:"clickPerson"`
	UpdatedPin    *string              `json:"updatedPin"`
	UpdatedAt     *javacompat.DateTime `json:"updatedAt"`
}

func convertResourceManagementList(raw json.RawMessage) json.RawMessage {
	if javacompat.IsNullJSON(raw) {
		return raw
	}
	var out []resourceManagementCO
	if json.Unmarshal(raw, &out) != nil {
		return raw
	}
	b, _ := json.Marshal(out)
	return b
}
