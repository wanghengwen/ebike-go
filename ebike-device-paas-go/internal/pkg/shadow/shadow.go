// Package shadow supports the fence-style proxy gateway:
//
//   - liveList: Go only
//   - recordList: Java proxy + [RECORD]
//   - neither + DryRun: dual-run shadow compare (Go captured, client gets Java)
//
// Report / jsondiff remain available for offline log replay. CompareWithJava is
// kept as a no-op compatibility shim (middleware owns live shadowing).
package shadow

import (
	"context"
	"log"

	"ebike-device-paas-go/internal/pkg/config"
	"ebike-device-paas-go/internal/pkg/jsondiff"
)

type shadowKeyType struct{}

var shadowKey = shadowKeyType{}

// WithShadowTest marks the request context as shadow dual-run. Handlers / repos
// may skip local side effects (Redis writes, Kafka) when IsShadowTest is true.
func WithShadowTest(ctx context.Context) context.Context {
	return context.WithValue(ctx, shadowKey, true)
}

// IsShadowTest reports whether ctx is in middleware shadow dual-run mode.
func IsShadowTest(ctx context.Context) bool {
	v, _ := ctx.Value(shadowKey).(bool)
	return v
}

// diffOptions builds jsondiff options from current config, merging any per-call
// order-insensitive keys (e.g. a top-level "data" list for a specific endpoint).
func diffOptions(extraUnordered ...string) jsondiff.Options {
	unordered := config.UnorderedFieldSet()
	for _, k := range extraUnordered {
		unordered[k] = true
	}
	return jsondiff.Options{
		IgnoreKeys:     config.IgnoreFieldSet(),
		FloatTolerance: config.GlobalConfig.Shadow.FloatTolerance,
		UnorderedKeys:  unordered,
	}
}

// CompareWithJava is a no-op. Shadow compare is owned by middleware.ProxyGateway
// (fence-go style). Kept so older call sites / tests compile.
func CompareWithJava(method, path string, reqBody, goResponseJSON []byte, extraUnordered ...string) {
	_ = method
	_ = path
	_ = reqBody
	_ = goResponseJSON
	_ = extraUnordered
}

// Report compares two responses and logs MATCH or DIFF. Used by ProxyGateway
// and offline log-replay tests.
func Report(path string, reqBody, javaJSON, goJSON []byte, extraUnordered ...string) bool {
	if jsondiff.EqualOpts(goJSON, javaJSON, diffOptions(extraUnordered...)) {
		log.Printf("[SHADOW MATCH] Path: %s", path)
		return true
	}
	log.Printf("[SHADOW DIFF] Path: %s\nReq: %s\nJava: %s\nGo: %s",
		path, string(reqBody), string(javaJSON), string(goJSON))
	return false
}
