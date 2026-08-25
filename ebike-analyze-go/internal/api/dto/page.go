package dto

// PageDTO mirrors Java com.xyy.dto.PageDTO<T>.
type PageDTO[T any] struct {
	Count       int64       `json:"count"`
	PageNum     int         `json:"pageNum"`
	PageSize    int         `json:"pageSize"`
	Orders      []OrderItem `json:"orders"`
	SearchCount bool        `json:"searchCount"`
	List        []T         `json:"list"`
}

func NewPageDTO[T any](pageNum, pageSize int, count int64, list []T) PageDTO[T] {
	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if list == nil {
		list = []T{}
	}
	return PageDTO[T]{
		Count:       count,
		PageNum:     pageNum,
		PageSize:    pageSize,
		SearchCount: true,
		List:        list,
	}
}
