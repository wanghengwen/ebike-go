package repository

import (
	"strconv"
	"time"

	"ebike-device-paas-go/internal/pkg/redis"
)

// Voice find-car rate-limit parameters, mirroring the @RateLimiter on
// DeviceCommandApiRpcImpl.voiceCommand (xyy-redis RateLimiterAspect):
//
//	key            = rateLimiterFindVoice_{tenantId}_{pin}
//	timeUnit       = MINUTES        -> hash field "yyyyMMddHHmm", value = count
//	preUnitMaxCount= 10 (default)   -> per-minute cap
//	windowSize     = 5              -> 5-minute sliding window
//	windowMaxCount = 30 (default)   -> window cap; exceeding it sets a 24h ban
//	windowForbidden= 24h (default)
const (
	voicePreUnitMaxCount = 10
	voiceWindowSize      = 5
	voiceWindowMaxCount  = 30
	voiceWindowForbidden = 24 * time.Hour
	voiceTimeLayout      = "200601021504" // yyyyMMddHHmm

	// VoicePreUnitMsg / VoiceWindowMsg mirror the annotation's
	// preUnitMaxCountMsg / windowMaxCountMsg.
	VoicePreUnitMsg = "请勿频繁点击寻车铃"
	VoiceWindowMsg  = "寻车铃功能将禁用24小时"
)

// VoiceFindCarAllowed applies the voice find-car sliding-window limiter for
// (tenantId, pin), mirroring RateLimiterAspect.rateLimiter. It returns
// (allowed=true, "") when the request may proceed, or (false, msg) when limited
// (msg is the per-unit or window-ban message). On the window-cap breach it also
// sets the {key}_forbidden ban for 24h. The counter is only incremented when the
// request is allowed (matching the aspect, which increments after the checks).
//
// Like the Java aspect (which swallows non-BizException errors and proceeds),
// any Redis unavailability degrades to "allowed" because the redis wrappers
// return zero values when the client is nil.
func VoiceFindCarAllowed(tenantID, pin string) (bool, string) {
	key := "rateLimiterFindVoice_" + tenantID + "_" + pin

	if redis.Exists(key + "_forbidden") {
		return false, VoiceWindowMsg
	}

	now := time.Now()
	m := redis.HGetAll(key)

	cur := now.Format(voiceTimeLayout)
	curN, _ := strconv.ParseInt(cur, 10, 64)
	windowMinN, _ := strconv.ParseInt(now.Add(-time.Duration(voiceWindowSize-1)*time.Minute).Format(voiceTimeLayout), 10, 64)

	// Window check: sum the counts of the buckets within [windowMin, cur].
	var reqCount int64
	for field, v := range m {
		f, err := strconv.ParseInt(field, 10, 64)
		if err != nil {
			continue
		}
		if f >= windowMinN && f <= curN {
			c, _ := strconv.ParseInt(v, 10, 64)
			reqCount += c
		}
	}
	if reqCount > voiceWindowMaxCount {
		redis.SetEx(key+"_forbidden", "1", voiceWindowForbidden)
		return false, VoiceWindowMsg
	}

	// Per-minute check (uses >= like the aspect's preUnit comparison).
	if pre, _ := strconv.ParseInt(m[cur], 10, 64); pre >= voicePreUnitMaxCount {
		return false, VoicePreUnitMsg
	}

	// Allowed: record the hit, refresh TTL, and prune buckets older than the window.
	redis.HIncrBy(key, cur, 1)
	redis.Expire(key, time.Duration((voiceWindowSize+1)*2)*time.Minute)
	var stale []string
	for field := range m {
		if f, err := strconv.ParseInt(field, 10, 64); err == nil && f < windowMinN {
			stale = append(stale, field)
		}
	}
	redis.HDel(key, stale...)
	return true, ""
}
