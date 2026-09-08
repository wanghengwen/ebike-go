package middleware

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"ebike-gateway-go/config"
	"ebike-gateway-go/logger"
	"ebike-gateway-go/proxy"
	"ebike-gateway-go/response"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

var (
	Rdb   *redis.Client
	rdbMu sync.RWMutex
)

// InitRedis connects to Redis using current configuration.
func InitRedis() error {
	rdbMu.Lock()
	defer rdbMu.Unlock()

	cfg := config.GetRedisConfig()
	Rdb = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.Database,
	})
	return Rdb.Ping(context.Background()).Err()
}

// ReinitRedis reconnects Redis using the latest configuration (for Nacos hot reload).
func ReinitRedis() error {
	rdbMu.Lock()
	defer rdbMu.Unlock()

	if Rdb != nil {
		_ = Rdb.Close()
	}

	cfg := config.GetRedisConfig()
	Rdb = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.Database,
	})
	return Rdb.Ping(context.Background()).Err()
}

// InitRedisWithPolicy initializes Redis and enforces startup policy.
func InitRedisWithPolicy() error {
	err := InitRedis()
	if err == nil {
		return nil
	}
	required := config.AppConfig.GatewayMode == "business" || config.AppConfig.RedisRequired
	if required {
		return fmt.Errorf("redis connection required but failed: %w", err)
	}
	logger.Log.Error("Failed to init redis, continuing because REDIS_REQUIRED=false", zap.Error(err))
	return nil
}

func getRedisClient() *redis.Client {
	rdbMu.RLock()
	defer rdbMu.RUnlock()
	return Rdb
}

func SecurityMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}

		routeObj, exists := c.Get("matchedRoute")
		if !exists {
			c.Next()
			return
		}
		matchedRoute := routeObj.(*proxy.Route)

		if !matchedRoute.SecurityEnabled {
			c.Next()
			return
		}

		path := c.Request.URL.Path
		for _, ignore := range matchedRoute.SecurityIgnores {
			if ignore != "" && ignore == path {
				c.Next()
				return
			}
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			response.AbortUnauthorized(c)
			return
		}

		tokenString := strings.TrimSpace(authHeader[7:])

		secretToUse := matchedRoute.SecuritySecret
		if secretToUse == "" {
			secretToUse = config.GetJwtSecret()
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return []byte(secretToUse), nil
		})

		if err != nil || !token.Valid {
			response.AbortTokenInvalid(c, "token无效或过期")
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			response.AbortTokenInvalid(c, "token无效或过期")
			return
		}

		if exp, ok := claims["exp"].(float64); ok {
			if int64(exp) < time.Now().Unix() {
				response.AbortTokenInvalid(c, "token无效或过期")
				return
			}
		}

		deviceId, _ := claims["deviceId"].(string)
		if deviceId == "" {
			deviceId = "normal"
		}
		userName, _ := claims["user_name"].(string)
		clientId, _ := claims["client_id"].(string)

		if getRedisClient() != nil && userName != "" {
			if err := verifyDeviceLogin(claims, deviceId, userName, clientId); err != nil {
				response.Abort(c, err.status, err.code, err.msg)
				return
			}
		} else if getRedisClient() != nil && config.AppConfig.GatewayMode == "business" {
			response.Abort(c, http.StatusUnauthorized, response.CodeInvalidToken, "登录已过期，请重新登录(01)")
			return
		}

		if config.AppConfig.GatewayMode == "business" && config.AppConfig.EnableApiVerify {
			if !checkAPIPermission(claims, path) {
				logger.Log.Warn("API refuse access",
					zap.String("traceId", GetTraceID(c)),
					zap.String("tenantId", clientId),
					zap.String("path", path),
				)
				response.Abort(c, http.StatusForbidden, response.CodeAPIRefuseAccess, "API拒绝访问")
				return
			}
		}

		payloadBytes, _ := json.Marshal(claims)
		authorities := base64.URLEncoding.EncodeToString(payloadBytes)
		c.Request.Header.Set("authorities", authorities)

		c.Set("claims", claims)
		c.Next()
	}
}

type deviceCheckError struct {
	status int
	code   string
	msg    string
}

func verifyDeviceLogin(claims jwt.MapClaims, deviceId, userName, clientId string) *deviceCheckError {
	rdb := getRedisClient()
	if rdb == nil {
		return nil
	}

	if config.AppConfig.GatewayMode == "business" {
		platform, _ := claims["platform"].(string)
		redisKey := fmt.Sprintf("login_device_%s_%s", getPlatformName(platform), userName)
		cachedDeviceId, err := rdb.Get(context.Background(), redisKey).Result()
		if err == redis.Nil || cachedDeviceId == "" {
			return &deviceCheckError{http.StatusUnauthorized, response.CodeInvalidToken, "登录已过期，请重新登录(01)"}
		}
		if err != nil {
			logger.Log.Error("Redis check failed", zap.Error(err))
			return &deviceCheckError{http.StatusUnauthorized, response.CodeInvalidToken, "登录已过期，请重新登录(01)"}
		}
		if cachedDeviceId == "role-change" {
			return &deviceCheckError{http.StatusUnauthorized, response.CodeUserRoleChanged, "账号权限已变更，请重新登录(01)"}
		}
		if cachedDeviceId != deviceId {
			return &deviceCheckError{http.StatusUnauthorized, response.CodeOtherDeviceLogin, "账号已在别处登录(01)"}
		}
		return nil
	}

	// client mode
	if clientId == "" {
		return nil
	}
	newKey := fmt.Sprintf("login_device_%s_%s", clientId, userName)
	cachedDeviceId, errNew := rdb.Get(context.Background(), newKey).Result()
	oldKey := fmt.Sprintf("login_device_%s", userName)
	cachedDeviceIdOld, errOld := rdb.Get(context.Background(), oldKey).Result()

	newEmpty := errNew == redis.Nil || cachedDeviceId == ""
	oldEmpty := errOld == redis.Nil || cachedDeviceIdOld == ""
	if newEmpty && oldEmpty {
		return &deviceCheckError{http.StatusUnauthorized, response.CodeTokenInvalidOrExpired, "token无效或过期"}
	}
	if cachedDeviceId != deviceId {
		return &deviceCheckError{http.StatusUnauthorized, response.CodeOtherDeviceLogin, "账号已在别处登录"}
	}
	return nil
}

func getPlatformName(platform string) string {
	switch strings.ToLower(platform) {
	case "android", "ios":
		return "app"
	case "wechat", "pc", "other":
		return strings.ToLower(platform)
	default:
		return "other"
	}
}

func checkAPIPermission(claims jwt.MapClaims, path string) bool {
	var authorities []string
	if authObj, ok := claims["authorities"].([]interface{}); ok {
		for _, a := range authObj {
			if str, ok := a.(string); ok {
				authorities = append(authorities, str)
			}
		}
	}
	if len(authorities) == 0 {
		return false
	}

	tenantId, _ := claims["client_id"].(string)
	for _, authority := range authorities {
		if strings.EqualFold(authority, "root") {
			return true
		}
		if PermManager != nil && PermManager.CanAccessApi(tenantId, authority, path) {
			return true
		}
	}
	return false
}
