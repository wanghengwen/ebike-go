package service

import (
	"testing"
	"time"
)

func TestWaitLuopingReplyTimeout(t *testing.T) {
	SetActionRedis(nil)
	_, err := waitLuopingReply("tid-missing", 20*time.Millisecond)
	if err == nil {
		t.Fatal("expected error without redis")
	}
}
