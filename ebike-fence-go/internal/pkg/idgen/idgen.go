package idgen

import (
	"sync/atomic"
	"time"

	"ebike-fence-go/internal/pkg/fastid"
)

var seq atomic.Uint32

// NextID returns the next distributed id. Uses FastId when initialized (Java FenceKeyGenerator),
// otherwise a local fallback for unit tests.
func NextID() int64 {
	if fastid.Ready() {
		return fastid.Next()
	}
	ms := time.Now().UnixMilli()
	s := seq.Add(1) & 0x3FFFFF
	return (ms << 22) | int64(s)
}
