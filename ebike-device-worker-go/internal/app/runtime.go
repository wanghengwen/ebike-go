package app

import (
	"log"

	"ebike-device-worker-go/internal/config"
	"ebike-device-worker-go/internal/domain/device"
	"ebike-device-worker-go/internal/domain/handler"
	"ebike-device-worker-go/internal/domain/persistence"
	"ebike-device-worker-go/internal/domain/push"
	"ebike-device-worker-go/internal/infrastructure/kafka"
	"ebike-device-worker-go/internal/infrastructure/persistence/pg"
	"ebike-device-worker-go/internal/infrastructure/rpc"
	redispkg "ebike-device-worker-go/internal/pkg/redis"
)

// Runtime holds core worker dependencies.
type Runtime struct {
	KafkaManager *kafka.Manager
	Producer     *kafka.Producer
	PushService  *push.Service
	DataHandler  *handler.DataHandler
	EventHandler *handler.EventHandler
	AlarmHandler *handler.AlarmHandler
}

func NewRuntime(cfg *config.Config) *Runtime {
	repo := pg.NewRepo()
	registry := device.NewRegistry(
		device.NewGPS(repo),
		device.NewPing(repo),
		device.NewBMS(repo),
		device.NewAlarm(repo),
		device.NewFault(repo),
		device.NewCmd(repo),
		device.NewOnline(repo),
		device.NewOffline(repo),
	)
	jdbc := persistence.NewJDBC(registry, cfg)
	deps := &handler.Deps{JDBC: jdbc, Cfg: cfg}

	producer := kafka.NewProducer(cfg)
	paas := rpc.NewPaasClient(cfg.Paas.BaseURL, cfg.Paas.Enabled)
	pushSvc := push.NewService(cfg, producer, paas)

	rt := &Runtime{
		KafkaManager: kafka.NewManager(cfg),
		Producer:     producer,
		PushService:  pushSvc,
		DataHandler:  handler.NewDataHandler(deps, pushSvc),
		EventHandler: handler.NewEventHandler(deps, pushSvc),
		AlarmHandler: handler.NewAlarmHandler(deps, pushSvc),
	}
	return rt
}

func (r *Runtime) Start() {
	if r == nil {
		return
	}
	redispkg.StartTenantSubscriber(r.PushService.HandleTenantChange)
	r.KafkaManager.Start(r.DataHandler, r.EventHandler, r.AlarmHandler)
}

func (r *Runtime) Stop() {
	if r == nil {
		return
	}
	r.KafkaManager.Stop()
	_ = r.Producer.Close()
}

// RestartKafka reloads producer brokers and restarts consumers after Nacos kafka.yaml changes.
func (r *Runtime) RestartKafka() {
	if r == nil || r.KafkaManager == nil {
		return
	}
	if err := r.Producer.Reinit(config.GlobalConfig); err != nil {
		log.Printf("[kafka] producer reinit failed: %v", err)
	}
	r.KafkaManager.Restart()
}
