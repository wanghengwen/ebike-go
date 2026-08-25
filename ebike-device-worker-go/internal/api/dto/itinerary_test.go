package dto_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"ebike-device-worker-go/internal/api/dto"
	_ "ebike-device-worker-go/internal/pkg/web"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

func TestBatchOrderTrajectoryCmdAllowsNullNestedType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	body := []byte(`{"orderTrajectoryRequests":[{"type":null,"orderId":"369691897740925658","startTime":1781581606000,"endTime":1781581871000}],"type":1}`)

	var cmd dto.BatchOrderTrajectoryCmd
	req := httptest.NewRequest(http.MethodPost, "/ebike/gps/getBatchOrderTrajectory", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if err := binding.JSON.Bind(req, &cmd); err != nil {
		t.Fatalf("bind failed: %v", err)
	}
	if cmd.Type == nil || *cmd.Type != 1 {
		t.Fatalf("top-level type=%v", cmd.Type)
	}
	if len(cmd.OrderTrajectoryRequests) != 1 {
		t.Fatalf("requests len=%d", len(cmd.OrderTrajectoryRequests))
	}
	if cmd.OrderTrajectoryRequests[0].Type != nil {
		t.Fatalf("nested type should stay nil, got %v", cmd.OrderTrajectoryRequests[0].Type)
	}
}
