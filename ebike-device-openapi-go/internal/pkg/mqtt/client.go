package mqtt

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"ebike-device-openapi-go/internal/pkg/config"
	"ebike-device-openapi-go/internal/pkg/logger"
	paho "github.com/eclipse/paho.mqtt.golang"
	"go.uber.org/zap"
)

// MessageHandler is invoked for every inbound MQTT message (topic + raw payload).
type MessageHandler func(topic string, payload []byte)

// Client wraps a paho MQTT client used for both uplink (shared subscription)
// and downlink (publish) with the luoping/EMQX broker.
type Client struct {
	c       paho.Client
	handler MessageHandler
	topics  map[string]byte // topics to (re)subscribe on every (re)connect
}

var (
	defaultClient *Client
	clientMu      sync.RWMutex
)

// InitClient connects to the MQTT broker and subscribes to the luoping uplink
// topics using shared subscriptions. Subscriptions are (re)established inside the
// OnConnect handler so they survive broker reconnects (clean session drops subs).
//
// It is a no-op when mqtt.enabled is false. handler may be nil (publish-only).
func InitClient(handler MessageHandler) error {
	cfg := config.GetConfig()
	if !cfg.Mqtt.Enabled {
		logger.Log.Info("mqtt client disabled (mqtt.enabled=false)")
		return nil
	}

	broker := config.GetMqttBroker()
	if broker == "" {
		return fmt.Errorf("mqtt.broker is empty")
	}

	clientID := config.GetMqttClientID()
	keepAlive := cfg.Mqtt.KeepAliveSeconds
	if keepAlive <= 0 {
		keepAlive = 30
	}
	subQos := byte(cfg.Mqtt.SubscribeQos)

	// Build the shared-subscription topic set.
	group := cfg.Mqtt.SharedGroup
	if group == "" {
		group = "openapi"
	}
	topics := map[string]byte{}
	if handler != nil {
		if cfg.Mqtt.RptTopic != "" {
			topics[sharedTopic(group, cfg.Mqtt.RptTopic)] = subQos
		}
		// RspTopic kept for yaml compat but V8 binary replies arrive on brpt;
		// leave empty in defaults so it is not subscribed.
		if cfg.Mqtt.RspTopic != "" {
			topics[sharedTopic(group, cfg.Mqtt.RspTopic)] = subQos
		}
		// Only disconnected presence is subscribed. Connected is ignored by design
		// (online comes from binary uplink / cmd=35). EMQX 5.x supports $share on $SYS;
		// ACL must allow this client to subscribe $SYS for disconnectedTopic.
		if cfg.Mqtt.DisconnectedTopic != "" {
			topics[sharedTopic(group, cfg.Mqtt.DisconnectedTopic)] = subQos
		}
	}

	cl := &Client{handler: handler, topics: topics}

	opts := paho.NewClientOptions()
	for _, b := range strings.Split(broker, ",") {
		if b = strings.TrimSpace(b); b != "" {
			opts.AddBroker(b)
		}
	}
	opts.SetClientID(clientID)
	if u := config.GetMqttUsername(); u != "" {
		opts.SetUsername(u)
	}
	if p := config.GetMqttPassword(); p != "" {
		opts.SetPassword(p)
	}
	opts.SetCleanSession(cfg.Mqtt.CleanSession)
	opts.SetKeepAlive(time.Duration(keepAlive) * time.Second)
	opts.SetConnectTimeout(10 * time.Second)
	opts.SetAutoReconnect(true)
	opts.SetConnectRetry(true)
	opts.SetConnectRetryInterval(5 * time.Second)
	opts.SetMaxReconnectInterval(30 * time.Second)
	opts.SetOrderMatters(false)
	opts.SetResumeSubs(false)

	opts.SetDefaultPublishHandler(func(_ paho.Client, m paho.Message) {
		cl.dispatch(m)
	})
	opts.SetOnConnectHandler(func(pc paho.Client) {
		logger.Log.Info("mqtt connected", zap.String("clientId", clientID))
		cl.subscribeAll(pc)
	})
	opts.SetConnectionLostHandler(func(_ paho.Client, err error) {
		logger.Log.Warn("mqtt connection lost", zap.Error(err))
	})

	pc := paho.NewClient(opts)
	cl.c = pc

	// ConnectRetry=true means Connect returns a token that completes once the
	// first connection succeeds; wait briefly but don't block startup forever.
	token := pc.Connect()
	if err := waitToken(token, 10*time.Second); err != nil {
		logger.Log.Warn("mqtt initial connect not ready (will retry in background)", zap.Error(err))
	}

	clientMu.Lock()
	defaultClient = cl
	clientMu.Unlock()

	logger.Log.Info("mqtt client initialized",
		zap.String("broker", broker),
		zap.String("clientId", clientID),
		zap.String("sharedGroup", group),
		zap.Any("subscriptions", topicKeys(topics)),
	)
	return nil
}

func (cl *Client) subscribeAll(pc paho.Client) {
	if len(cl.topics) == 0 {
		return
	}
	handler := func(_ paho.Client, m paho.Message) { cl.dispatch(m) }
	for topic, qos := range cl.topics {
		token := pc.Subscribe(topic, qos, handler)
		if err := waitToken(token, 5*time.Second); err != nil {
			logger.Log.Error("mqtt subscribe failed", zap.String("topic", topic), zap.Error(err))
			continue
		}
		logger.Log.Info("mqtt subscribed", zap.String("topic", topic), zap.Uint8("qos", qos))
	}
}

func (cl *Client) dispatch(m paho.Message) {
	if cl.handler == nil {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			logger.Log.Error("mqtt message handler panic", zap.Any("recover", r), zap.String("topic", m.Topic()))
		}
	}()
	cl.handler(m.Topic(), m.Payload())
}

// Publish serializes payload as JSON and publishes it to topic at the configured
// publish QoS.
func Publish(topic string, payload interface{}) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return PublishBytes(topic, body)
}

// PublishBytes publishes raw bytes (e.g. V8 encrypted MQTT payload) at configured QoS.
func PublishBytes(topic string, body []byte) error {
	clientMu.RLock()
	cl := defaultClient
	clientMu.RUnlock()
	if cl == nil || cl.c == nil {
		return fmt.Errorf("mqtt client not initialized")
	}
	if !cl.c.IsConnected() {
		return fmt.Errorf("mqtt client not connected")
	}

	qos := byte(config.GetConfig().Mqtt.PublishQos)

	logger.Log.Info("mqtt publish",
		zap.String("topic", topic),
		zap.Uint8("qos", qos),
		zap.Int("payloadBytes", len(body)),
	)
	token := cl.c.Publish(topic, qos, false, body)
	if err := waitToken(token, 10*time.Second); err != nil {
		logger.Log.Error("mqtt publish failed",
			zap.String("topic", topic),
			zap.Uint8("qos", qos),
			zap.Int("payloadBytes", len(body)),
			zap.Error(err),
		)
		return fmt.Errorf("mqtt publish failed: %w", err)
	}
	return nil
}

// waitToken waits for a paho token up to timeout, distinguishing a timeout from
// a completed-with-error result. Returns nil only when the operation actually
// completed successfully. (token.WaitTimeout returns false on timeout, in which
// case token.Error() may still be nil — that must NOT be treated as success.)
func waitToken(t paho.Token, timeout time.Duration) error {
	if !t.WaitTimeout(timeout) {
		return fmt.Errorf("operation timed out after %s", timeout)
	}
	return t.Error()
}

// Healthy reports whether the MQTT dependency is healthy for readiness checks.
// It returns true when MQTT is not configured (disabled or no broker) — nothing to
// check — and otherwise true only when the client exists and is connected.
func Healthy() bool {
	cfg := config.GetConfig()
	if !cfg.Mqtt.Enabled || config.GetMqttBroker() == "" {
		return true
	}
	clientMu.RLock()
	cl := defaultClient
	clientMu.RUnlock()
	return cl != nil && cl.c != nil && cl.c.IsConnected()
}

// Close disconnects the MQTT client, allowing in-flight work to drain.
func Close() {
	clientMu.RLock()
	cl := defaultClient
	clientMu.RUnlock()
	if cl != nil && cl.c != nil && cl.c.IsConnected() {
		cl.c.Disconnect(1000)
		logger.Log.Info("mqtt client disconnected")
	}
}

// CmdTopic builds V8 downlink topic ecu/bcmd/ebike/{deviceId}.
// groupName is ignored; the product segment is fixed to "ebike".
func CmdTopic(_groupName, deviceId string) string {
	return BcmdTopic(deviceId)
}

// BcmdTopic builds V8 downlink topic ecu/bcmd/ebike/{deviceId}.
func BcmdTopic(deviceId string) string {
	return fmt.Sprintf("ecu/bcmd/ebike/%s", deviceId)
}

// sharedTopic wraps a topic filter as an EMQX shared subscription: $share/{group}/{filter}.
func sharedTopic(group, filter string) string {
	return "$share/" + group + "/" + strings.TrimPrefix(filter, "/")
}

func topicKeys(m map[string]byte) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
