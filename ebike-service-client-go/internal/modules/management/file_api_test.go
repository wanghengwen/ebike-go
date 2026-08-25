package management

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"ebike-service-client-go/internal/api/dto"
	"github.com/gin-gonic/gin"
)

func TestBindFileUploadFormRejectsOversizedBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	part, err := w.CreateFormFile("file", "big.bin")
	if err != nil {
		t.Fatal(err)
	}
	// Body must exceed maxFileUploadBytes; use a repeating source larger than the limit.
	chunk := bytes.Repeat([]byte("x"), 64*1024)
	remaining := maxFileUploadBytes + 1024
	for remaining > 0 {
		n := len(chunk)
		if n > remaining {
			n = remaining
		}
		if _, err := part.Write(chunk[:n]); err != nil {
			t.Fatal(err)
		}
		remaining -= n
	}
	_ = w.WriteField("traceId", "t1")
	_ = w.WriteField("tenantId", "tenant1")
	_ = w.Close()

	req := httptest.NewRequest(http.MethodPost, "/client/file/upload", body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = req
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxFileUploadBytes)

	var got fileUploadDTO
	if bindFileUploadForm(c, &got) {
		t.Fatal("expected oversized upload to be rejected")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var result dto.Result
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if result.Success {
		t.Fatal("expected success=false for oversized upload")
	}
	if result.Code == nil || *result.Code != dto.CodeIllegalArgument {
		t.Fatalf("code = %v, want %s", result.Code, dto.CodeIllegalArgument)
	}
	if result.Msg == nil || *result.Msg != "Maximum upload size exceeded" {
		t.Fatalf("msg = %v", result.Msg)
	}
}

func TestBindFileUploadFormAcceptsSmallUpload(t *testing.T) {
	gin.SetMode(gin.TestMode)

	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	part, err := w.CreateFormFile("file", "ok.txt")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write([]byte("hello")); err != nil {
		t.Fatal(err)
	}
	_ = w.WriteField("traceId", "t1")
	_ = w.WriteField("tenantId", "tenant1")
	_ = w.Close()

	req := httptest.NewRequest(http.MethodPost, "/client/file/upload", body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = req
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxFileUploadBytes)

	var got fileUploadDTO
	if !bindFileUploadForm(c, &got) {
		t.Fatalf("expected small upload to pass binding, body=%s", rec.Body.String())
	}
	if got.File == nil || got.File.Filename != "ok.txt" {
		t.Fatalf("file header = %#v", got.File)
	}
}
