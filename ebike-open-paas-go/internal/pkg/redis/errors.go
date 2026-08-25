package redis

import "errors"

// ErrUnavailable is returned by the write/read helpers that cannot degrade
// silently. Rate-limit counters may fail open, but callback subscriptions must
// not: silently accepting a registration we did not persist would leave the
// third party believing it will receive events.
var ErrUnavailable = errors.New("redis unavailable")
