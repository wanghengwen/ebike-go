package persistence

import (
	"log"
	"strings"

	"ebike-device-worker-go/internal/config"
	"ebike-device-worker-go/internal/domain/device"
	"ebike-device-worker-go/internal/domain/message"
	"ebike-device-worker-go/internal/pkg/utils"
)

// JDBC mirrors Java JdbcMessagePersistence (PostgreSQL only).
type JDBC struct {
	registry *device.Registry
	cfg      *config.Config
}

func NewJDBC(registry *device.Registry, cfg *config.Config) *JDBC {
	return &JDBC{registry: registry, cfg: cfg}
}

func (j *JDBC) Persist(structs []message.MessageInfoDTO) {
	if len(structs) == 0 {
		return
	}

	grouped := make(map[string][]message.MessageInfoDTO)
	for _, s := range structs {
		bt := s.BussinessType
		if strings.EqualFold(bt, "bike") {
			bt = "ebike"
		}
		key := bt + "_" + utils.Itoa(s.Cmd)
		grouped[key] = append(grouped[key], s)
	}

	batchSize := j.cfg.PersistConfig.PersistBatchSize
	if batchSize <= 0 {
		batchSize = 20
	}
	for key, list := range grouped {
		svc, ok := j.registry.Get(key)
		if !ok {
			log.Printf("[persistence] no service for key=%s count=%d", key, len(list))
			continue
		}
		if err := svc.SaveBatch(list, batchSize); err != nil {
			log.Printf("[persistence] save failed key=%s: %v", key, err)
		}
	}
}
