package device

import (
	"ebike-device-worker-go/internal/domain/message"
)

// Service mirrors Java domain.device.DeviceService.
type Service interface {
	SaveBatch(list []message.MessageInfoDTO, batchSize int) error
	Key() string
}

type Registry struct {
	services map[string]Service
}

func NewRegistry(services ...Service) *Registry {
	r := &Registry{services: make(map[string]Service)}
	for _, s := range services {
		r.services[s.Key()] = s
	}
	return r
}

func (r *Registry) Get(key string) (Service, bool) {
	s, ok := r.services[key]
	return s, ok
}
