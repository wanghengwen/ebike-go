package controller

import (
	"github.com/gin-gonic/gin"

	"ebike-fence-go/internal/api/dto"
)

func fenceBusinessHandlers() map[string]gin.HandlerFunc {
	initFenceAdmin()
	return map[string]gin.HandlerFunc{
		"/fence/business/queryByFenceIds": handleFenceBusinessQueryByIDs,
		"/fence/business/queryByFenceId":  handleFenceBusinessQueryByID,
	}
}

func handleFenceBusinessQueryByIDs(c *gin.Context) {
	var ids []int64
	if !bindJSONCmd(c, &ids, nil) {
		return
	}
	res, err := businessAdmin.QueryByFenceIDs(c.Request.Context(), ids)
	respondRaw(c, res, err)
}

func handleFenceBusinessQueryByID(c *gin.Context) {
	var req dto.FenceCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	if _, ok := applyCmd(c, &req.Command); !ok {
		return
	}
	res, err := businessAdmin.QueryByFenceID(c.Request.Context(), req.Id)
	respondRaw(c, res, err)
}
