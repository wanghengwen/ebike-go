package dto

// IdCmd matches Java com.xyy.ebike.fence.api.dto.packing.IdCmd.
type IdCmd struct {
	Command
	Id        *int64 `json:"id,omitempty"`
	ServiceId *int64 `json:"serviceId,omitempty"`
}

// IdsCmd matches Java IdsCmd.
type IdsCmd struct {
	Command
	Ids []int64 `json:"ids,omitempty"`
}

// ServiceIdCmd matches Java ServiceIdCmd.
type ServiceIdCmd struct {
	Command
	ServiceId *int64 `json:"serviceId" binding:"required"`
}

// PageCmd matches Java PageCmd.
type PageCmd struct {
	Command
	PageNum  int                    `json:"pageNum,omitempty"`
	PageSize int                    `json:"pageSize,omitempty"`
	Filter   map[string]interface{} `json:"filter,omitempty"`
}

// PageDTO matches Java com.xyy.dto.PageDTO.
type PageDTO[T any] struct {
	List     []T   `json:"list"`
	Total    int64 `json:"total"`
	PageNum  int   `json:"pageNum"`
	PageSize int   `json:"pageSize"`
}

// JavaPageDTO matches MyBatis-Plus page responses (count field).
type JavaPageDTO[T any] struct {
	Count       int64 `json:"count"`
	PageNum     int   `json:"pageNum"`
	PageSize    int   `json:"pageSize"`
	Orders      any   `json:"orders"`
	SearchCount bool  `json:"searchCount"`
	List        []T   `json:"list"`
}
