package rpc

import (
	"errors"
	"testing"
)

func TestIsServiceDiscoveryError(t *testing.T) {
	tests := []struct {
		err  error
		want bool
	}{
		{nil, false},
		{errors.New("failed to discover service ebike-device: instance list is empty!"), true},
		{errors.New("RPC failed for http://1.2.3.4/x: timeout"), false},
	}
	for _, tt := range tests {
		if got := IsServiceDiscoveryError(tt.err); got != tt.want {
			t.Fatalf("IsServiceDiscoveryError(%v) = %v want %v", tt.err, got, tt.want)
		}
	}
}
