package model

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

// LocalTime wraps time.Time for JSON serialization compatible with Java's
// LocalDateTime (ISO-8601 without timezone: "2021-10-29T22:31:00").
type LocalTime struct {
	time.Time
}

const localTimeFormat = "2006-01-02T15:04:05"

func (t LocalTime) MarshalJSON() ([]byte, error) {
	return []byte(`"` + t.Time.Format(localTimeFormat) + `"`), nil
}

func (t *LocalTime) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	if s == "" || s == "null" {
		return nil
	}
	parsed, err := time.Parse(localTimeFormat, s)
	if err != nil {
		return fmt.Errorf("LocalTime.UnmarshalJSON: %w", err)
	}
	t.Time = parsed
	return nil
}

// Value implements the driver.Valuer interface for GORM.
func (t LocalTime) Value() (driver.Value, error) {
	if t.Time.IsZero() {
		return nil, nil
	}
	return t.Time, nil
}

// Scan implements the sql.Scanner interface for GORM.
func (t *LocalTime) Scan(v interface{}) error {
	if v == nil {
		t.Time = time.Time{}
		return nil
	}
	switch val := v.(type) {
	case time.Time:
		t.Time = val
		return nil
	case []byte:
		parsed, err := time.Parse("2006-01-02 15:04:05", string(val))
		if err != nil {
			return err
		}
		t.Time = parsed
		return nil
	case string:
		parsed, err := time.Parse("2006-01-02 15:04:05", val)
		if err != nil {
			return err
		}
		t.Time = parsed
		return nil
	default:
		return fmt.Errorf("unsupported Scan type for LocalTime: %T", v)
	}
}

// Now returns the current time as a LocalTime.
func Now() LocalTime {
	return LocalTime{Time: time.Now()}
}

// App maps to the t_app table.
type App struct {
	ID        uint   `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Namespace string `gorm:"column:namespace;type:varchar(64);not null;uniqueIndex:uk_ns_group_name" json:"namespace"`
	GroupID   string `gorm:"column:group_id;type:varchar(64);not null;uniqueIndex:uk_ns_group_name" json:"groupId"`
	Name      string `gorm:"column:name;type:varchar(64);not null;uniqueIndex:uk_ns_group_name" json:"name"`
	// default:CURRENT_TIMESTAMP matches production schema.
	CreateTime LocalTime `gorm:"column:create_time;type:timestamp;not null;default:CURRENT_TIMESTAMP" json:"createTime"`
}

func (App) TableName() string {
	return "t_app"
}

// Machine maps to the t_machine table.
type Machine struct {
	ID          uint   `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	AppID       uint   `gorm:"column:app_id;type:int unsigned;not null;uniqueIndex:uk_app_machine" json:"appId"`
	MachineUUID string `gorm:"column:machine_uuid;type:varchar(128);not null;uniqueIndex:uk_app_machine" json:"machineUuid"`
	// type:bigint matches production DB; Go int aligns with Java Integer for business logic.
	MachineID  int       `gorm:"column:machine_id;type:bigint;not null" json:"machineId"`
	CreateTime LocalTime `gorm:"column:create_time;type:timestamp;not null;default:CURRENT_TIMESTAMP" json:"createTime"`
	BeatTime   LocalTime `gorm:"column:beat_time;type:timestamp;not null;default:CURRENT_TIMESTAMP" json:"beatTime"`
}

func (Machine) TableName() string {
	return "t_machine"
}
