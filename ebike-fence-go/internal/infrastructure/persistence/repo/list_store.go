package repo

import (
	"context"
	"encoding/json"
	"errors"

	"ebike-fence-go/internal/domain/gateway"
	"ebike-fence-go/internal/domain/rediskeys"

	"gorm.io/gorm"
)

// ListStore provides cache-aside list read and double-delete write for service-scoped rows.
type ListStore[T any] struct {
	DB       *gorm.DB
	Store    *gateway.ConfigStore
	CacheKey func(tenantID string, serviceID int64) string
	OrderBy  string
}

func (s *ListStore[T]) ListByServiceID(ctx context.Context, tenantID string, serviceID int64) ([]T, error) {
	if s.DB == nil {
		return nil, errors.New("database not configured")
	}
	key := s.CacheKey(tenantID, serviceID)
	raw, err := s.Store.GetList(ctx, key, func(db *gorm.DB) (string, error) {
		var rows []T
		q := db.Where("service_id = ?", serviceID)
		if s.OrderBy != "" {
			q = q.Order(s.OrderBy)
		}
		if err := q.Find(&rows).Error; err != nil {
			return "", err
		}
		b, err := json.Marshal(rows)
		return string(b), err
	})
	if err != nil {
		return nil, err
	}
	if rediskeys.IsCacheMiss(raw) {
		return nil, nil
	}
	var rows []T
	if err := json.Unmarshal([]byte(raw), &rows); err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *ListStore[T]) GetByID(ctx context.Context, id int64) (*T, error) {
	if s.DB == nil {
		return nil, errors.New("database not configured")
	}
	var row T
	if err := s.DB.WithContext(ctx).First(&row, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (s *ListStore[T]) Insert(ctx context.Context, tenantID string, serviceID int64, row *T) error {
	key := s.CacheKey(tenantID, serviceID)
	return s.Store.WriteWithInvalidate(ctx, key, func(db *gorm.DB) error {
		return db.Create(row).Error
	})
}

func (s *ListStore[T]) Update(ctx context.Context, tenantID string, serviceID int64, row *T) error {
	key := s.CacheKey(tenantID, serviceID)
	return s.Store.WriteWithInvalidate(ctx, key, func(db *gorm.DB) error {
		return db.Save(row).Error
	})
}

func (s *ListStore[T]) DeleteByID(ctx context.Context, tenantID string, serviceID int64, id int64) error {
	key := s.CacheKey(tenantID, serviceID)
	return s.Store.WriteWithInvalidate(ctx, key, func(db *gorm.DB) error {
		var zero T
		return db.Delete(&zero, id).Error
	})
}

func (s *ListStore[T]) Invalidate(ctx context.Context, tenantID string, serviceID int64) {
	s.Store.Invalidate(ctx, s.CacheKey(tenantID, serviceID))
}

// ObjectStore provides cache-aside single-object read/write (CustomerService, GuidePage izOn).
type ObjectStore[T any] struct {
	DB       *gorm.DB
	Store    *gateway.ConfigStore
	CacheKey func(tenantID string, serviceID int64) string
	Load     func(db *gorm.DB, serviceID int64) (*T, error)
}

func (s *ObjectStore[T]) Get(ctx context.Context, tenantID string, serviceID int64) (*T, error) {
	if s.DB == nil {
		return nil, errors.New("database not configured")
	}
	key := s.CacheKey(tenantID, serviceID)
	raw, err := s.Store.GetJSON(ctx, key, func(db *gorm.DB) (string, error) {
		row, err := s.Load(db, serviceID)
		if err != nil || row == nil {
			return "", err
		}
		b, err := json.Marshal(row)
		return string(b), err
	})
	if err != nil {
		return nil, err
	}
	if rediskeys.IsCacheMiss(raw) {
		return s.Load(s.DB.WithContext(ctx), serviceID)
	}
	var out T
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *ListStore[T]) Write(ctx context.Context, tenantID string, serviceID int64, fn func(*gorm.DB) error) error {
	key := s.CacheKey(tenantID, serviceID)
	return s.Store.WriteWithInvalidate(ctx, key, fn)
}

func (s *ObjectStore[T]) Write(ctx context.Context, tenantID string, serviceID int64, write func(db *gorm.DB) error) error {
	key := s.CacheKey(tenantID, serviceID)
	return s.Store.WriteWithInvalidate(ctx, key, write)
}
