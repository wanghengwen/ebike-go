package cache

import (
	"database/sql"

	"ebike-fence-go/internal/infrastructure/persistence/model"

	"github.com/guregu/null/v5"
)

type fenceTag struct {
	ConfigBase
	TagID    int64     `json:"id"`
	TagName  string    `json:"tagName"`
	IzEnable null.Bool `json:"izEnable"`
}

func (c fenceTag) toModel() model.FenceTag {
	var izEnable sql.NullBool
	if c.IzEnable.Valid {
		izEnable = sql.NullBool{Bool: c.IzEnable.Bool, Valid: true}
	}
	return model.FenceTag{
		BaseConfig: c.ConfigBase.toBaseConfig(),
		TagID:      c.TagID,
		TagName:    c.TagName,
		IzEnable:   izEnable,
	}
}

func UnmarshalFenceTagList(raw string) ([]model.FenceTag, error) {
	return unmarshalList(raw, func(c fenceTag) model.FenceTag { return c.toModel() })
}
