package mq

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"sync"
	"time"

	appconfig "ebike-fence-go/internal/pkg/config"
	"ebike-fence-go/internal/pkg/timefmt"

	"github.com/segmentio/kafka-go"
)

// ConfigUpdateDTO mirrors Java com.xyy.ebike.fence.app.mq.dto.ConfigUpdateDTO.
type ConfigUpdateDTO struct {
	Config    int    `json:"config"`
	TenantId  string `json:"tenantId"`
	ServiceId int64  `json:"serviceId"`
}

var (
	writerOnce sync.Once
	writer     *kafka.Writer
)

func configUpdateTopic() string {
	prefix := appconfig.GlobalConfig.Kafka.TopicPrefix
	if prefix == "" {
		prefix = "ebike"
	}
	return prefix + "_config_update"
}

func getWriter() *kafka.Writer {
	writerOnce.Do(func() {
		brokers := appconfig.GlobalConfig.Kafka.Brokers
		if len(brokers) == 0 {
			return
		}
		writer = &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Topic:    configUpdateTopic(),
			Balancer: &kafka.LeastBytes{},
		}
	})
	return writer
}

// SynchronizeConfig mirrors Java ConfigUpdateKafkaProducer.synchronizeConfig.
func SynchronizeConfig(ctx context.Context, dto ConfigUpdateDTO) {
	if dto.ServiceId == 0 {
		return
	}
	w := getWriter()
	if w == nil {
		log.Printf("[kafka] skip config_update: brokers not configured")
		return
	}
	body, err := json.Marshal(dto)
	if err != nil {
		log.Printf("[kafka] marshal config_update failed: %v", err)
		return
	}
	msg := kafka.Message{
		Key:   []byte(strconv.FormatInt(dto.ServiceId, 10)),
		Value: body,
	}
	if err := w.WriteMessages(ctx, msg); err != nil {
		log.Printf("[kafka] publish config_update failed topic=%s payload=%s err=%v", configUpdateTopic(), string(body), err)
		return
	}
	log.Printf("[kafka] publish config_update success topic=%s payload=%s", configUpdateTopic(), string(body))
}

var (
	userActionOnce   sync.Once
	userActionWriter *kafka.Writer
)

func userActionTopic() string {
	prefix := appconfig.GlobalConfig.Kafka.TopicPrefix
	if prefix == "" {
		prefix = "ebike"
	}
	return prefix + "_user_action"
}

func getUserActionWriter() *kafka.Writer {
	userActionOnce.Do(func() {
		brokers := appconfig.GlobalConfig.Kafka.Brokers
		if len(brokers) == 0 {
			return
		}
		userActionWriter = &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Topic:    userActionTopic(),
			Balancer: &kafka.LeastBytes{},
		}
	})
	return userActionWriter
}

// PublishUserAction mirrors Java UserActionServiceImpl.applyPass. It publishes a
// user-action event ({type, pin, actionTime, CommandContext}) to {prefix}_user_action.
// Java publishes asynchronously and best-effort, so failures are logged, not returned.
// The actionType values follow Java's contract (2 = 站点申请通过).
func PublishUserAction(actionType int, pin string, cmdCtx interface{}) {
	w := getUserActionWriter()
	if w == nil {
		log.Printf("[kafka] skip user_action: brokers not configured")
		return
	}
	payload := map[string]interface{}{
		"type":           actionType,
		"pin":            pin,
		"actionTime":     timefmt.FormatJavaLocalSpace(time.Now()),
		"CommandContext": cmdCtx,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[kafka] marshal user_action failed: %v", err)
		return
	}
	// Use a background context so the publish is not cancelled when the HTTP request
	// completes (Java executes this asynchronously).
	if err := w.WriteMessages(context.Background(), kafka.Message{Value: body}); err != nil {
		log.Printf("[kafka] publish user_action failed topic=%s payload=%s err=%v", userActionTopic(), string(body), err)
		return
	}
	log.Printf("[kafka] publish user_action success topic=%s payload=%s", userActionTopic(), string(body))
}

// Close releases the kafka writers.
func Close() error {
	var firstErr error
	if writer != nil {
		if err := writer.Close(); err != nil {
			firstErr = err
		}
	}
	if userActionWriter != nil {
		if err := userActionWriter.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}
