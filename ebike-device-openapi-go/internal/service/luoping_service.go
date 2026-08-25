package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"

	"ebike-device-openapi-go/internal/api/dto"
	"ebike-device-openapi-go/internal/decode/xiaoan"
	"ebike-device-openapi-go/internal/pkg/config"
	lpcrypto "ebike-device-openapi-go/internal/pkg/crypto"
	"ebike-device-openapi-go/internal/pkg/logger"
	"ebike-device-openapi-go/internal/pkg/mqtt"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	luopingProductGroup = "ebike"
	luopingPendingTTL   = 2 * time.Minute
	luopingSeqKeyPrefix = "mqtt:seq:"
	luopingPendingPrefix = "mqtt:pending:"
)

// LuopingService handles EMQX presence + MQTT message ingest for type=luoping.
type LuopingService struct {
	Register *RegisterService
	Mapping  *LuopingMappingService
	Sweeper  *LuopingPresenceSweeper
}

// HandleConnected is intentionally a no-op.
// Business online is driven by brpt uplink (cmd=35 / other frames with version rules).
func (s *LuopingService) HandleConnected(ctx context.Context, ev *dto.EmqxPresenceEvent) error {
	clientID := ""
	if ev != nil {
		clientID = strings.TrimSpace(ev.ClientID)
	}
	logger.Log.Debug("luoping connected ignored by design (online via uplink only)",
		zap.String("clientId", clientID))
	return nil
}

// HandleDisconnected mirrors gateway ConnectionCountHandler.channelInactive for 小安 TCP:
//
//	unbind session → DeviceOpenapiRpc.logout(imei, channelId) → Register.Unregister
//
// EMQX client.disconnected is the MQTT equivalent of Netty channelInactive.
func (s *LuopingService) HandleDisconnected(ctx context.Context, ev *dto.EmqxPresenceEvent) error {
	deviceId := strings.TrimSpace(ev.ClientID)
	if deviceId == "" {
		return fmt.Errorf("clientid is required")
	}
	if !isLuopingDeviceClientID(deviceId) {
		logger.Log.Debug("luoping disconnected ignored: non-device clientid",
			zap.String("clientId", deviceId))
		return nil
	}

	// MQTT race guard (TCP has no equivalent: session is bound to one channel).
	ts := presenceMillis(ev.DisconnectedAt, ev.Timestamp)
	applied, err := s.Mapping.TryAdvancePresence(ctx, deviceId, ts, 0)
	if err != nil {
		logger.Log.Warn("luoping presence epoch check failed (disconnected)",
			zap.String("deviceId", deviceId), zap.Error(err))
	}
	if !applied {
		logger.Log.Info("luoping stale disconnected event ignored",
			zap.String("deviceId", deviceId), zap.Int64("ts", ts))
		return nil
	}

	// ≈ session.getImei(); skip when session == null
	mapping, err := s.Mapping.GetByDeviceId(ctx, deviceId)
	if err != nil {
		logger.Log.Error("luoping disconnected mapping lookup failed",
			zap.String("deviceId", deviceId), zap.Error(err))
		return err
	}
	if mapping == nil || mapping.Imei == "" {
		logger.Log.Debug("luoping disconnected without imei mapping",
			zap.String("deviceId", deviceId))
		return nil
	}

	// Same as gateway DeviceOpenapiRpc.logout → POST /device-gateway/xiaoan/logout
	// channelId for luoping is deviceId (written at bindOnline).
	if s.Register == nil {
		return fmt.Errorf("luoping register service not initialized")
	}
	logger.Log.Info("luoping device logout",
		zap.String("deviceId", deviceId),
		zap.String("imei", mapping.Imei),
		zap.String("channelId", deviceId),
	)
	ok := s.Register.Unregister(&dto.LogoutCmd{
		Imei:      mapping.Imei,
		ChannelId: deviceId,
	}, false)
	if !ok {
		logger.Log.Error("luoping unregister failed on disconnect",
			zap.String("deviceId", deviceId), zap.String("imei", mapping.Imei))
		return fmt.Errorf("luoping unregister failed")
	}

	// MQTT bookkeeping (TCP session already unbound above). The mapping itself is
	// persistent and stays put; only the local record is dropped so the next session
	// re-establishes it.
	s.Mapping.forgetMappingWrite(deviceId)
	if s.Sweeper != nil {
		s.Sweeper.Remove(ctx, mapping.Imei)
	}
	return nil
}

// HandleMqtt routes a raw MQTT message by topic:
//   - EMQX …/disconnected (optional) → HandleDisconnected
//   - EMQX …/connected → ignored
//   - ecu/brpt/ebike/{deviceId}/{imei} → decrypt + 小安帧
//
// Luoping MQTT is a real production path: side effects always run (no shadowMode).
func (s *LuopingService) HandleMqtt(ctx context.Context, topic string, payload []byte) error {
	if strings.HasSuffix(topic, "/disconnected") {
		if config.GetConfig().Mqtt.DisconnectedTopic == "" {
			return nil
		}
		var ev dto.EmqxPresenceEvent
		if err := json.Unmarshal(payload, &ev); err != nil {
			return fmt.Errorf("parse disconnected event: %w", err)
		}
		return s.HandleDisconnected(ctx, &ev)
	}
	if strings.HasSuffix(topic, "/connected") {
		return s.HandleConnected(ctx, nil)
	}
	return s.handleBinaryUplink(ctx, topic, payload)
}

// luopingFrame is a decrypted 小安 plaintext frame.
type luopingFrame struct {
	Cmd      int16
	Sequence int16
	MsgType  string
	Map      map[string]interface{}
	// Typed holds the concrete decode result (e.g. *xiaoan.Bin35LoginMessage).
	Typed interface{}
}

// decodeXiaoanFrame parses an AA55 frame and decodes its body. Wild frames
// (header cmd=0) carry a JSON body that has no 小安 binary decoder, so they are
// unwrapped here — over TCP the gateway does this before calling openapi.
func decodeXiaoanFrame(raw []byte) (*luopingFrame, error) {
	header, body, err := xiaoan.ParseFrame(raw)
	if err != nil {
		return nil, err
	}
	out := &luopingFrame{
		Cmd:      header.Cmd,
		Sequence: header.Sequence,
	}

	if header.Cmd == xiaoan.WildCmdHeader {
		jsonBytes := body.ReadBytes(body.ReadableBytes())
		if body.Err() != nil {
			return nil, body.Err()
		}
		out.MsgType = "event"
		out.Map = map[string]interface{}{
			"msgType": "event",
			"cmd":     0,
			"json":    string(jsonBytes),
		}
		return out, nil
	}

	decoded, err := xiaoan.DecodeHex(header, body)
	if err != nil {
		return out, fmt.Errorf("xiaoan decode cmd=%d: %w", header.Cmd, err)
	}
	out.Typed = decoded
	out.Map, out.MsgType = toDecodedMap(decoded)
	return out, nil
}

// toDecodedMap normalizes a decoded 小安 struct into a map (via JSON round-trip)
// and extracts its msgType, mirroring the 小安 TCP decode path.
func toDecodedMap(decoded interface{}) (map[string]interface{}, string) {
	b, err := json.Marshal(decoded)
	if err != nil {
		return map[string]interface{}{}, "data"
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		return map[string]interface{}{}, "data"
	}
	msgType, _ := m["msgType"].(string)
	if msgType == "" {
		msgType = "data"
	}
	return m, msgType
}

// handleBinaryUplink processes V8 encrypted MQTT uplink on ecu/brpt/ebike/{deviceId}/{imei}.
func (s *LuopingService) handleBinaryUplink(ctx context.Context, topic string, payload []byte) error {
	deviceId, imei, err := parseBrptTopic(topic)
	if err != nil {
		return err
	}
	groupName := luopingProductGroup

	plain, err := lpcrypto.DecryptLuopingMQTT(payload, deviceId, imei)
	if err != nil {
		logger.Log.Warn("luoping decrypt failed",
			zap.String("deviceId", deviceId), zap.String("imei", imei), zap.Error(err))
		return fmt.Errorf("luoping decrypt: %w", err)
	}

	// Successful decryption alone proves liveness — even when the frame below fails
	// to decode, the device is online. The same round trip reports whether the
	// session still exists, which is what drives re-binding further down.
	sessionAlive := s.touchLastSeen(ctx, imei, deviceId, groupName)
	if s.Sweeper != nil {
		s.Sweeper.Touch(ctx, imei, time.Now().Unix())
	}

	frame, decodeErr := decodeXiaoanFrame(plain)
	if decodeErr != nil {
		logger.Log.Warn("luoping frame decode failed",
			zap.String("deviceId", deviceId),
			zap.String("imei", imei),
			zap.String("plainHex", fmt.Sprintf("%x", plain)),
			zap.Error(decodeErr))
		return decodeErr
	}
	xaCmd := frame.Cmd
	seq := frame.Sequence
	decodedMap := frame.Map
	msgType := frame.MsgType

	// The binding is established once per session and left alone until it changes.
	// A missing session means Redis may have lost the mapping along with it, so drop
	// the local record first and let it be re-persisted.
	if !sessionAlive {
		s.Mapping.forgetMappingWrite(deviceId)
	}
	if err := s.Mapping.SaveMapping(ctx, deviceId, imei, groupName); err != nil {
		return err
	}

	// Wild reply (header cmd=0): correlate by seq like gateway WildKey(imei, sequence).
	if xaCmd == xiaoan.WildCmdHeader {
		return s.handleWildReply(ctx, imei, seq, decodedMap)
	}

	login := extractLuopingLoginInfo(nil, decodedMap)
	// Prefer typed Bin35 fields — avoids JSON map round-trip dropping version=0 etc.
	if xaCmd == 35 {
		if msg, ok := frame.Typed.(*xiaoan.Bin35LoginMessage); ok && msg != nil {
			login = loginInfoFromBin35(msg)
			logger.Log.Info("luoping bin35 decoded",
				zap.String("deviceId", deviceId),
				zap.String("topicImei", imei),
				zap.String("bodyImei", msg.Imei),
				zap.String("imsi", msg.Imsi),
				zap.Int64("version", msg.Version),
				zap.Int("deviceType", msg.DeviceType),
				zap.Int16("seq", seq),
			)
			if msg.Imei != "" && msg.Imei != imei {
				logger.Log.Warn("luoping bin35 imei mismatch with topic",
					zap.String("topicImei", imei), zap.String("bodyImei", msg.Imei))
			}
		} else {
			logger.Log.Warn("luoping cmd=35 typed decode missing",
				zap.String("deviceId", deviceId),
				zap.String("imei", imei),
				zap.String("plainHex", fmt.Sprintf("%x", plain)),
				zap.String("typed", fmt.Sprintf("%T", frame.Typed)),
			)
		}
	}

	// cmd=35 is the login frame and always (re)binds. Any other frame binds only when
	// the session is gone, which is how a device force-offlined by the sweeper comes
	// back without a power cycle: the sweeper deletes ecu_login but keeps the mapping.
	needBind := xaCmd == 35
	if !needBind && s.Register != nil && s.Register.Rdb != nil {
		needBind = !sessionAlive
	}
	// Device ACKs (gateway-compatible), before bind / kafka push.
	switch xaCmd {
	case 35:
		// LoginHandler.replayLogin: AA55|35|seq|len=4|unix
		if ackErr := s.publishLoginAck(deviceId, imei, byte(seq)); ackErr != nil {
			logger.Log.Error("luoping login ack publish failed",
				zap.String("deviceId", deviceId),
				zap.String("imei", imei),
				zap.Int16("seq", seq),
				zap.Error(ackErr),
			)
		}
	case 2:
		// ReplayMessage autoReplay: AA55|02|seq|len=0
		if ackErr := s.publishAutoReplay(deviceId, imei, byte(xaCmd), byte(seq)); ackErr != nil {
			logger.Log.Error("luoping cmd2 ack publish failed",
				zap.String("deviceId", deviceId),
				zap.String("imei", imei),
				zap.Int16("seq", seq),
				zap.Error(ackErr),
			)
		}
	}

	if needBind {
		ver := int64(-1)
		if login != nil && login.Version != nil {
			ver = *login.Version
		}
		imsi := ""
		if login != nil {
			imsi = login.Imsi
		}
		logger.Log.Info("luoping brpt bind",
			zap.String("deviceId", deviceId),
			zap.String("imei", imei),
			zap.String("group", groupName),
			zap.String("imsi", imsi),
			zap.Int64("version", ver),
			zap.Int16("xiaoanCmd", xaCmd),
		)
		if err := s.bindOnline(ctx, imei, deviceId, groupName, "", login); err != nil {
			return err
		}
	}

	// cmd=35 is the login frame itself: bind + ACK already happened, nothing to push.
	if xaCmd == 35 {
		return nil
	}

	dataBytes, _ := json.Marshal(decodedMap)
	report := &dto.DeviceReportMessage{
		Imei:            imei,
		MsgType:         msgType,
		BussinessType:   "ebike",
		Data:            string(dataBytes),
		ReceiveDataTime: time.Now().UnixMilli(),
	}
	s.Register.Pusher.PushMessage(imei, report, false)
	return nil
}

// publishLoginAck encrypts gateway-compatible login ACK and publishes to bcmd.
func (s *LuopingService) publishLoginAck(deviceId, imei string, seq byte) error {
	frame := xiaoan.EncodeLoginAck(seq, time.Now().Unix())
	return s.publishEncryptedDownlink(deviceId, imei, frame, "login ack", seq)
}

// publishAutoReplay encrypts gateway ReplayMessage (empty body) and publishes to bcmd.
func (s *LuopingService) publishAutoReplay(deviceId, imei string, cmd, seq byte) error {
	frame := xiaoan.EncodeAutoReplay(cmd, seq)
	return s.publishEncryptedDownlink(deviceId, imei, frame, fmt.Sprintf("autoReplay cmd=%d", cmd), seq)
}

func (s *LuopingService) publishEncryptedDownlink(deviceId, imei string, frame []byte, kind string, seq byte) error {
	enc, err := lpcrypto.EncryptLuopingMQTT(frame, deviceId, imei)
	if err != nil {
		return err
	}
	topic := mqtt.BcmdTopic(deviceId)
	if err := mqtt.PublishBytes(topic, enc); err != nil {
		return err
	}
	logger.Log.Info("luoping downlink ack published",
		zap.String("kind", kind),
		zap.String("topic", topic),
		zap.String("deviceId", deviceId),
		zap.String("imei", imei),
		zap.Uint8("seq", seq),
		zap.String("plainHex", fmt.Sprintf("%x", frame)),
		zap.Int("payloadBytes", len(enc)),
	)
	return nil
}

func loginInfoFromBin35(msg *xiaoan.Bin35LoginMessage) *luopingLoginInfo {
	info := &luopingLoginInfo{
		Imsi: strings.TrimSpace(msg.Imsi),
	}
	v := msg.Version
	info.Version = &v
	dt := msg.DeviceType
	info.DeviceType = &dt
	return info
}

func (s *LuopingService) handleWildReply(ctx context.Context, imei string, seq int16, decodedMap map[string]interface{}) error {
	jsonBody := ""
	if decodedMap != nil {
		if v, ok := decodedMap["json"].(string); ok {
			jsonBody = v
		}
	}
	jobId := ""
	if s.Register != nil && s.Register.Rdb != nil {
		pendingKey := luopingPendingKey(imei, seq)
		if v, err := s.Register.Rdb.Get(ctx, pendingKey).Result(); err == nil && v != "" {
			jobId = v
			_ = s.Register.Rdb.Del(ctx, pendingKey).Err()
			replayKey := "ecu_replay_" + jobId
			_ = s.Register.Rdb.Set(ctx, replayKey, jsonBody, luopingPendingTTL).Err()
		}
	}

	replyData := map[string]interface{}{
		"msgType":       "event",
		"cmd":           0,
		"dir":           "u",
		"msgId":         jobId,
		"bussinessType": "ebike",
		"timestamp":     time.Now().Unix(),
		"sequence":      seq,
		"result":        jsonBody,
	}
	dataStr, _ := json.Marshal(replyData)
	msg := &dto.DeviceReplyMessage{
		Imei:            imei,
		MsgType:         "event",
		BussinessType:   "ebike",
		Data:            string(dataStr),
		ReceiveDataTime: time.Now().UnixMilli(),
		MsgId:           jobId,
		Payload:         jsonBody,
	}
	s.Register.Pusher.PushReplyMessage(imei, msg, false)
	return nil
}

// luopingLoginInfo carries 小安 cmd=35 login fields into RegisterLuoping.
type luopingLoginInfo struct {
	Imsi       string
	Version    *int64
	DeviceType *int
}

func (s *LuopingService) bindOnline(ctx context.Context, imei, deviceId, groupName, peername string, login *luopingLoginInfo) error {
	now := time.Now().Unix()
	ecuLogin := &dto.EcuLogin{
		Type:          dto.DeviceTypeLuoping,
		DeviceId:      deviceId,
		ClientId:      deviceId,
		GroupName:     groupName,
		ChannelId:     deviceId,
		RemoteAddress: peername,
		Timestamp:     &now,
		LastSeenAt:    &now,
	}
	cmd := &dto.LoginCmd{
		Imei:          imei,
		Host:          "emqx",
		Port:          1883,
		ChannelId:     deviceId,
		RemoteAddress: peername,
		Timestamp:     &now,
	}
	if login != nil {
		cmd.Imsi = login.Imsi
		cmd.DeviceType = login.DeviceType
		if login.Version != nil {
			cmd.Version = login.Version
			ecuLogin.Version = login.Version
		}
	}

	// Non-35 / missing version: restore from Redis; else default 0.0.0 (0).
	if cmd.Version == nil {
		key := ecuLoginKey(imei)
		if s.Register != nil && s.Register.Rdb != nil {
			if val, err := s.Register.Rdb.Get(ctx, key).Result(); err == nil && val != "" {
				var existing dto.EcuLogin
				if err := json.Unmarshal([]byte(val), &existing); err == nil && existing.Version != nil {
					cmd.Version = existing.Version
					ecuLogin.Version = existing.Version
					logger.Log.Info("luoping bindOnline restored version",
						zap.String("imei", imei), zap.Int64("version", *existing.Version))
				}
			}
		}
	}
	if cmd.Version == nil {
		zero := int64(0) // 0.0.0
		cmd.Version = &zero
		ecuLogin.Version = &zero
		logger.Log.Info("luoping bindOnline default version 0.0.0",
			zap.String("deviceId", deviceId), zap.String("imei", imei))
	}

	if err := s.Register.RegisterLuoping(cmd, ecuLogin); err != nil {
		return err
	}
	if s.Sweeper != nil {
		s.Sweeper.Touch(ctx, imei, now)
	}
	return nil
}

// extractLuopingLoginInfo prefers decoded Bin35 fields.
func extractLuopingLoginInfo(params, decoded map[string]interface{}) *luopingLoginInfo {
	info := &luopingLoginInfo{}
	if decoded != nil {
		info.Imsi = strings.TrimSpace(asLoginString(decoded["imsi"]))
		info.Version = resolveLoginVersion(decoded["version"], decoded["ver"])
		info.DeviceType = asLoginIntPtr(decoded["deviceType"])
		if info.DeviceType == nil {
			info.DeviceType = asLoginIntPtr(decoded["equipmentType"])
		}
	}
	if params != nil {
		if info.Imsi == "" {
			info.Imsi = paramString(params, "imsi", "IMSI")
		}
		if info.Version == nil {
			info.Version = resolveLoginVersion(params["version"], params["ver"], params["VER"])
		}
		if info.DeviceType == nil {
			info.DeviceType = asLoginIntPtr(params["deviceType"])
			if info.DeviceType == nil {
				info.DeviceType = asLoginIntPtr(params["equipmentType"])
			}
		}
	}
	return info
}

func resolveLoginVersion(candidates ...interface{}) *int64 {
	for _, c := range candidates {
		if c == nil {
			continue
		}
		if s := asLoginString(c); strings.Contains(s, ".") {
			if v := parseXiaoanFirmwareVer(s); v != nil {
				return v
			}
			continue
		}
		if v := asLoginInt64Ptr(c); v != nil {
			return v
		}
		if v := parseXiaoanFirmwareVer(asLoginString(c)); v != nil {
			return v
		}
	}
	return nil
}

// parseXiaoanFirmwareVer packs Major.Minor.Patch into 0x00MMNNPP.
func parseXiaoanFirmwareVer(s string) *int64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	if len(s) > 0 && (s[0] == 'v' || s[0] == 'V') {
		s = s[1:]
	}
	parts := strings.Split(s, ".")
	if len(parts) < 3 {
		return nil
	}
	var major, minor, patch int64
	if _, err := fmt.Sscanf(parts[0], "%d", &major); err != nil {
		return nil
	}
	if _, err := fmt.Sscanf(parts[1], "%d", &minor); err != nil {
		return nil
	}
	if _, err := fmt.Sscanf(parts[2], "%d", &patch); err != nil {
		return nil
	}
	if major < 0 || major > 255 || minor < 0 || minor > 255 || patch < 0 || patch > 255 {
		return nil
	}
	v := (major << 16) | (minor << 8) | patch
	return &v
}

func asLoginString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	default:
		s := strings.TrimSpace(fmt.Sprint(t))
		if s == "" || s == "<nil>" {
			return ""
		}
		return s
	}
}

func asLoginInt64Ptr(v interface{}) *int64 {
	if v == nil {
		return nil
	}
	switch t := v.(type) {
	case int64:
		return &t
	case int:
		n := int64(t)
		return &n
	case float64:
		n := int64(t)
		return &n
	case json.Number:
		n, err := t.Int64()
		if err != nil {
			return nil
		}
		return &n
	case string:
		var n int64
		if _, err := fmt.Sscan(strings.TrimSpace(t), &n); err != nil {
			return nil
		}
		return &n
	default:
		return nil
	}
}

func asLoginIntPtr(v interface{}) *int {
	n := asLoginInt64Ptr(v)
	if n == nil {
		return nil
	}
	i := int(*n)
	return &i
}

// touchLastSeen refreshes the presence epoch and the session's lastSeenAt, and
// reports whether Redis currently holds a luoping session for the IMEI. The GET it
// already performs is what answers that, so callers need no extra lookup.
//
// A transient Redis failure reports the session as alive: re-binding on a blip would
// emit a spurious online event, while skipping it only defers recovery by one frame.
func (s *LuopingService) touchLastSeen(ctx context.Context, imei, deviceId, groupName string) bool {
	if s.Register == nil || s.Register.Rdb == nil {
		return true
	}
	_, _ = s.Mapping.TryAdvancePresence(ctx, deviceId, time.Now().UnixMilli(), 0)
	key := ecuLoginKey(imei)
	val, err := s.Register.Rdb.Get(ctx, key).Result()
	switch {
	case errors.Is(err, redis.Nil):
		return false
	case err != nil:
		logger.Log.Warn("luoping session lookup failed",
			zap.String("imei", imei), zap.Error(err))
		return true
	case val == "":
		return false
	}
	var ecu dto.EcuLogin
	if err := json.Unmarshal([]byte(val), &ecu); err != nil {
		return false
	}
	if !strings.EqualFold(ecu.Type, dto.DeviceTypeLuoping) {
		return false
	}
	now := time.Now().Unix()
	ecu.LastSeenAt = &now
	if ecu.DeviceId == "" {
		ecu.DeviceId = deviceId
	}
	if ecu.GroupName == "" {
		ecu.GroupName = groupName
	}
	newVal, _ := json.Marshal(&ecu)
	if err := s.Register.Rdb.Set(ctx, key, string(newVal), 0).Err(); err != nil {
		logger.Log.Warn("luoping lastSeen update failed",
			zap.String("imei", imei), zap.Error(err))
	}
	return true
}

func isLuopingDeviceClientID(clientID string) bool {
	if len(clientID) != 10 {
		return false
	}
	for _, r := range clientID {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func paramString(params map[string]interface{}, keys ...string) string {
	for _, k := range keys {
		v, ok := params[k]
		if !ok || v == nil {
			continue
		}
		var s string
		if str, ok := v.(string); ok {
			s = str
		} else {
			s = fmt.Sprintf("%v", v)
		}
		if s = strings.TrimSpace(s); s != "" {
			return s
		}
	}
	return ""
}

func presenceMillis(candidates ...int64) int64 {
	for _, v := range candidates {
		if v > 0 {
			return normalizeToMillis(v)
		}
	}
	return time.Now().UnixMilli()
}

func normalizeToMillis(ts int64) int64 {
	if ts < 1_000_000_000_000 {
		return ts * 1000
	}
	return ts
}

// parseBrptTopic expects ecu/brpt/ebike/{deviceId}/{imei}.
func parseBrptTopic(topic string) (deviceId, imei string, err error) {
	parts := strings.Split(strings.Trim(topic, "/"), "/")
	if len(parts) != 5 || parts[0] != "ecu" || parts[1] != "brpt" || parts[2] != "ebike" {
		return "", "", fmt.Errorf("invalid luoping brpt topic: %s", topic)
	}
	deviceId = strings.TrimSpace(parts[3])
	imei = strings.TrimSpace(parts[4])
	if deviceId == "" || imei == "" {
		return "", "", fmt.Errorf("empty deviceId/imei in topic: %s", topic)
	}
	return deviceId, imei, nil
}

// parseLuopingTopic is kept for tests/compat; V8 uplink uses parseBrptTopic.
func parseLuopingTopic(topic string) (kind, groupName, deviceId string, err error) {
	if deviceId, imei, e := parseBrptTopic(topic); e == nil {
		_ = imei
		return "brpt", luopingProductGroup, deviceId, nil
	}
	parts := strings.Split(strings.Trim(topic, "/"), "/")
	if len(parts) != 4 || parts[0] != "ecu" {
		return "", "", "", fmt.Errorf("invalid luoping topic: %s", topic)
	}
	kind = parts[1]
	groupName = parts[2]
	deviceId = parts[3]
	if kind != "brpt" && kind != "bcmd" {
		return "", "", "", fmt.Errorf("invalid luoping topic kind: %s", topic)
	}
	if deviceId == "" {
		return "", "", "", fmt.Errorf("empty deviceId in topic: %s", topic)
	}
	return kind, groupName, deviceId, nil
}

func luopingPendingKey(imei string, seq int16) string {
	return fmt.Sprintf("%s%s:%d", luopingPendingPrefix, imei, seq)
}

func luopingSeqRedisKey(deviceId string) string {
	return luopingSeqKeyPrefix + deviceId
}

// nextLuopingSequence returns the next 1-byte sequence for a device (0..255).
func nextLuopingSequence(ctx context.Context, deviceId string) (byte, error) {
	if actionRdb == nil {
		return byte(time.Now().UnixNano() & 0xFF), nil
	}
	n, err := actionRdb.Incr(ctx, luopingSeqRedisKey(deviceId)).Result()
	if err != nil {
		return 0, err
	}
	return byte(n & 0xFF), nil
}

// PublishLuopingWildCmd builds 小安 wild frame (cmd=0 + JSON body), encrypts, and
// publishes to ecu/bcmd/ebike/{deviceId}. Registers seq→jobId for sync wait.
func PublishLuopingWildCmd(ctx context.Context, ecuLogin *dto.EcuLogin, wildCmd *dto.WildCmd) (string, error) {
	if ecuLogin == nil {
		return "", fmt.Errorf("ecuLogin is nil")
	}
	if wildCmd == nil {
		return "", fmt.Errorf("wildCmd is nil")
	}
	deviceId := ecuLogin.DeviceId
	if deviceId == "" {
		deviceId = ecuLogin.ClientId
	}
	if deviceId == "" {
		return "", fmt.Errorf("luoping deviceId missing on ecuLogin")
	}
	imei := strings.TrimSpace(wildCmd.Imei)
	if imei == "" {
		return "", fmt.Errorf("luoping imei missing on wildCmd")
	}
	jobId := wildCmd.JobId
	if jobId == "" {
		jobId = strings.ReplaceAll(uuid.New().String(), "-", "")
	}

	seq, err := nextLuopingSequence(ctx, deviceId)
	if err != nil {
		return "", fmt.Errorf("luoping seq: %w", err)
	}

	frame, err := xiaoan.EncodeWildCmdFrame(seq, wildCmd.Cmd, wildCmd.Params, wildCmd.Dt, wildCmd.Tm)
	if err != nil {
		return "", fmt.Errorf("encode wild frame: %w", err)
	}
	enc, err := lpcrypto.EncryptLuopingMQTT(frame, deviceId, imei)
	if err != nil {
		return "", fmt.Errorf("encrypt downlink: %w", err)
	}

	if actionRdb != nil {
		_ = actionRdb.Set(ctx, luopingPendingKey(imei, int16(seq)), jobId, luopingPendingTTL).Err()
	}

	topic := mqtt.BcmdTopic(deviceId)
	if err := mqtt.PublishBytes(topic, enc); err != nil {
		return jobId, err
	}
	logger.Log.Info("luoping bcmd published",
		zap.String("topic", topic),
		zap.String("deviceId", deviceId),
		zap.String("imei", imei),
		zap.Int16("bizCmd", wildCmd.Cmd),
		zap.Uint8("seq", seq),
		zap.String("jobId", jobId),
		zap.Int("payloadBytes", len(enc)),
	)
	return jobId, nil
}
