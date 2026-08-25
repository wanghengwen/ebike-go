package repository

import (
	"fmt"
	"strings"

	"ebike-open-paas-go/internal/pkg/redis"
	"ebike-open-paas-go/internal/pkg/rediskey"
)

// Valid callback event types (Xiaoan event 1..5).
const (
	EventPing   = 1
	EventGPS    = 2
	EventNotify = 3
	EventBMS    = 4
	EventUART   = 5
)

// IsValidEvent reports whether event is in 1..5.
func IsValidEvent(event int) bool {
	return event >= EventPing && event <= EventUART
}

// IsDeliverableEvent reports whether this platform has a source for the event.
//
// Events 1-4 come off saas_0. Event 5 (UART passthrough) does not: no decoder
// produces UART frames and nothing publishes them, so a subscription to it would
// be silently inert. Registration refuses it rather than accepting a promise
// nothing can keep.
func IsDeliverableEvent(event int) bool {
	return IsValidEvent(event) && event != EventUART
}

// RegisterCallback adds url to the agent's event set. Returns the number of
// newly added URLs (0 means already present), matching Xiaoan's data semantics.
func RegisterCallback(agentID string, event int, url string) (int64, error) {
	url = strings.TrimSpace(url)
	if agentID == "" || !IsValidEvent(event) {
		return 0, fmt.Errorf("invalid callback params")
	}
	if err := ValidateCallbackURL(url); err != nil {
		return 0, err
	}
	n, err := redis.SAdd(rediskey.Callback(agentID, event), url)
	if err == nil {
		// Deliver to the new URL without waiting out the refresh interval.
		RefreshSubscriptions()
	}
	return n, err
}

// ListCallbacks returns all URLs registered for agent+event.
func ListCallbacks(agentID string, event int) ([]string, error) {
	if agentID == "" || !IsValidEvent(event) {
		return nil, fmt.Errorf("invalid callback params")
	}
	return redis.SMembers(rediskey.Callback(agentID, event))
}

// UnregisterCallback removes url from the agent's event set.
func UnregisterCallback(agentID string, event int, url string) (int64, error) {
	url = strings.TrimSpace(url)
	if agentID == "" || url == "" || !IsValidEvent(event) {
		return 0, fmt.Errorf("invalid callback params")
	}
	n, err := redis.SRem(rediskey.Callback(agentID, event), url)
	if err == nil {
		// Stop delivering immediately; an unsubscribed URL receiving events is
		// worse than one that waits for its first.
		RefreshSubscriptions()
	}
	return n, err
}
