package cache

import (
	"testing"
	"time"

	"ebike-device-openapi-go/internal/api/dto"
)

func TestEcuLoginCachePutGet(t *testing.T) {
	c := NewEcuLoginCache(10, time.Minute)
	login := &dto.EcuLogin{Host: "10.0.0.1", Port: 8080, Type: "xiaoan"}

	c.Put("862551059858406", login)
	got := c.Get("862551059858406")
	if got == nil || got.Host != "10.0.0.1" || got.Port != 8080 {
		t.Fatalf("unexpected cache value: %+v", got)
	}
}

func TestEcuLoginCacheExpiry(t *testing.T) {
	c := NewEcuLoginCache(10, 10*time.Millisecond)
	login := &dto.EcuLogin{Host: "10.0.0.2", Port: 8080, Type: "xiaoan"}
	c.Put("imei-expire", login)

	time.Sleep(20 * time.Millisecond)
	if got := c.Get("imei-expire"); got != nil {
		t.Fatalf("expected expired entry to be nil, got %+v", got)
	}
}
