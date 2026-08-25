package auth

import (
	"context"
	"net/http"

	"ebike-auth-go/internal/pkg/redis"

	"github.com/gin-gonic/gin"
)

func healthHandler(c *gin.Context) {
	redisStatus := "UP"
	ctx := context.Background()
	if redis.Rdb == nil {
		redisStatus = "DOWN"
	} else if err := redis.Rdb.Ping(ctx).Err(); err != nil {
		redisStatus = "DOWN"
	}

	status := "UP"
	if redisStatus != "UP" {
		status = "DOWN"
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": status,
			"components": gin.H{
				"redis": redisStatus,
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": status,
		"components": gin.H{
			"redis": redisStatus,
			"nacos": "UP",
		},
	})
}
