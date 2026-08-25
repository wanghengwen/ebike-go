package operation

import (
	"encoding/json"

	"ebike-service-client-go/internal/pkg/javacompat"
)

type repairRecordCO struct {
	CarId         *string              `json:"carId"`
	State         *int                 `json:"state"`
	BikeType      *int                 `json:"bikeType"`
	ReportName    *string              `json:"reportName"`
	ReportPhone   *string              `json:"reportPhone"`
	ReportTime    *javacompat.DateTime `json:"reportTime"`
	ReportDesc    *string              `json:"reportDesc"`
	RepairPart    []javacompat.LongStr `json:"repairPart"`
	RepairName    *string              `json:"repairName"`
	ReportPhoto   []string             `json:"reportPhoto"`
	ReportLat     *float64             `json:"reportLat"`
	ReportLng     *float64             `json:"reportLng"`
	ReportAddress *string              `json:"reportAddress"`
}

type repairRecordPageDTO struct {
	Count       *javacompat.LongStr `json:"count"`
	PageNum     *int                `json:"pageNum"`
	PageSize    *int                `json:"pageSize"`
	Orders      []json.RawMessage   `json:"orders"`
	SearchCount *bool               `json:"searchCount"`
	List        []repairRecordCO    `json:"list"`
}

func convertRepairRecordPageDTO(raw json.RawMessage) json.RawMessage {
	if javacompat.IsNullJSON(raw) {
		return raw
	}
	var page repairRecordPageDTO
	if json.Unmarshal(raw, &page) != nil {
		return raw
	}
	b, err := json.Marshal(page)
	if err != nil {
		return raw
	}
	return b
}
