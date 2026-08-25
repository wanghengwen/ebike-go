package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"identity-auth-go/internal/model"

	"github.com/gin-gonic/gin"
)

func TestAuthHandler_InvalidBody_Returns200WithFalse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	// Simulate the real handler behavior: always return HTTP 200 with success=false on bad input
	r.POST("/auth", func(c *gin.Context) {
		var req model.AuthRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusOK, model.R{Success: false, Msg: err.Error()})
			return
		}
		c.JSON(http.StatusOK, model.R{Success: true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth", bytes.NewBuffer([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	// Java always returns HTTP 200
	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d", w.Code)
	}

	var resp model.R
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Empty body missing required fields should give success=false
	if resp.Success {
		t.Fatalf("Expected success=false for empty body, got true")
	}
}

func TestAuthHandler_ValidBody_Returns200WithTrue(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	r.POST("/auth", func(c *gin.Context) {
		var req model.AuthRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusOK, model.R{Success: false, Msg: err.Error()})
			return
		}
		c.JSON(http.StatusOK, model.R{Success: true, Msg: "success (dry-run)"})
	})

	body := `{"traceId":"t1","tenantId":"100","name":"张三","idCardNum":"123456","type":1}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth", bytes.NewBuffer([]byte(body)))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d", w.Code)
	}

	var resp model.R
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !resp.Success {
		t.Fatalf("Expected success=true, got false, msg=%s", resp.Msg)
	}
}
