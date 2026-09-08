package service

import (
	"fmt"

	"ebike-open-paas-go/internal/contract"
	"ebike-open-paas-go/internal/pkg/logger"
	"ebike-open-paas-go/internal/pkg/redis"
	"ebike-open-paas-go/internal/pkg/rediskey"

	"go.uber.org/zap"
)

// BelongReason classifies why device ownership passed or failed.
type BelongReason string

const (
	BelongOK                 BelongReason = "OK"
	BelongMissingTenant      BelongReason = "MISSING_TENANT"
	BelongMissingIMEI        BelongReason = "MISSING_IMEI"
	BelongRedisUnavailable   BelongReason = "REDIS_UNAVAILABLE"
	BelongKeyNotFound        BelongReason = "KEY_NOT_FOUND"
	BelongTenantFieldEmpty   BelongReason = "TENANT_FIELD_EMPTY"
	BelongTenantMismatch     BelongReason = "TENANT_MISMATCH"
	BelongRegistryReadFailed BelongReason = "REGISTRY_READ_FAILED"
)

// BelongDiagnostics captures one ownership check for logs and /internal/debug.
type BelongDiagnostics struct {
	OK               bool         `json:"ok"`
	Reason           BelongReason `json:"reason"`
	ExpectedTenant   string       `json:"expectedTenant"`
	CachedTenant     string       `json:"cachedTenant,omitempty"`
	IMEI             string       `json:"imei"`
	RedisKey         string       `json:"redisKey"`
	RedisAddr        string       `json:"redisAddr"`
	RegistryDB       int          `json:"registryDb"`
	ShadowDB         int          `json:"shadowDb"`
	KeyExists        bool         `json:"keyExists"`
	TenantFieldSet   bool         `json:"tenantFieldSet"`
	SeparateRegistry bool         `json:"separateRegistryClient"`
	Hint             string       `json:"hint,omitempty"`
}

// DiagnoseDeviceBelongs inspects device_ebike_{imei}.tenantId against the
// agent's tenantId without side effects.
func DiagnoseDeviceBelongs(expectedTenant, imei string) BelongDiagnostics {
	d := BelongDiagnostics{
		IMEI:             imei,
		ExpectedTenant:   stringsTrim(expectedTenant),
		RedisKey:         rediskey.DeviceEbike(imei),
		RedisAddr:        redis.RegistryAddr(),
		RegistryDB:       redis.RegistryDB(),
		ShadowDB:         redis.ShadowDB(),
		SeparateRegistry: redis.UsesSeparateRegistryDB(),
	}

	if d.ExpectedTenant == "" {
		d.Reason = BelongMissingTenant
		d.Hint = "agent 未映射 tenantId，检查 Nacos open.agents"
		return d
	}
	if imei == "" {
		d.Reason = BelongMissingIMEI
		return d
	}
	if !redis.Available() {
		d.Reason = BelongRedisUnavailable
		d.Hint = "Redis 未初始化或 host 为空，检查 Nacos redis.yaml 与 Pod 日志 [redis]"
		return d
	}

	lookup := redis.RegistryLookup(d.RedisKey, "tenantId")
	if lookup.Err != nil {
		d.Reason = BelongRegistryReadFailed
		d.Hint = fmt.Sprintf("读 Redis 失败: %v；确认 REDIS_REGISTRY_DATABASE=%d 与 anvelink 登记库一致",
			lookup.Err, d.RegistryDB)
		return d
	}
	d.KeyExists = lookup.KeyExists
	d.TenantFieldSet = lookup.FieldExists
	d.CachedTenant = normalizeCacheTenant(lookup.Value)

	if !lookup.KeyExists {
		d.Reason = BelongKeyNotFound
		d.Hint = fmt.Sprintf("在 %s db=%d 未找到 %s；用 redis-cli -n %d HGETALL %s 核对",
			d.RedisAddr, d.RegistryDB, d.RedisKey, d.RegistryDB, d.RedisKey)
		return d
	}
	if d.CachedTenant == "" {
		d.Reason = BelongTenantFieldEmpty
		d.Hint = "hash 存在但 tenantId 为空；重新执行 saveBatchTenantDevice"
		return d
	}
	if d.CachedTenant != d.ExpectedTenant {
		d.Reason = BelongTenantMismatch
		d.Hint = fmt.Sprintf("设备登记 tenantId=%s，但 agent 映射 tenantId=%s；改 Nacos open.agents 或重新导入设备",
			d.CachedTenant, d.ExpectedTenant)
		return d
	}

	d.OK = true
	d.Reason = BelongOK
	return d
}

// LogBelongFailure emits a structured WARN when ownership fails (1001).
func LogBelongFailure(agentID, expectedTenant, imei string) {
	d := DiagnoseDeviceBelongs(expectedTenant, imei)
	logger.Log.Warn("device ownership check failed (1001)",
		zap.String("agentId", agentID),
		zap.String("reason", string(d.Reason)),
		zap.String("expectedTenant", d.ExpectedTenant),
		zap.String("cachedTenant", d.CachedTenant),
		zap.String("imei", d.IMEI),
		zap.String("redisKey", d.RedisKey),
		zap.String("redisAddr", d.RedisAddr),
		zap.Int("registryDb", d.RegistryDB),
		zap.Int("shadowDb", d.ShadowDB),
		zap.Bool("keyExists", d.KeyExists),
		zap.Bool("tenantFieldSet", d.TenantFieldSet),
		zap.Bool("separateRegistryClient", d.SeparateRegistry),
		zap.String("hint", d.Hint),
	)
}

func stringsTrim(s string) string {
	return normalizeCacheTenant(s)
}

// EnsureDeviceBelongs returns an error containing "1001" when the imei is not
// registered to the agent's tenant in device_ebike_{imei}.
func EnsureDeviceBelongs(agentID, tenantID, imei string) error {
	if tenantID == "" || imei == "" {
		return fmt.Errorf("%s", contract.ErrImeiIllegal)
	}
	d := DiagnoseDeviceBelongs(tenantID, imei)
	if d.OK {
		return nil
	}
	LogBelongFailure(agentID, tenantID, imei)
	return fmt.Errorf("%d", contract.CodeNotBelongAgent)
}
