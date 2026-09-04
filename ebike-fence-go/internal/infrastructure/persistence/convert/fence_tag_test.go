package convert

import (
	"database/sql"
	"testing"
	"time"

	"ebike-fence-go/internal/infrastructure/persistence/model"
)

// Java copies FenceTagDO/FenceTagE onto FenceTagCO, so version/createdPin/createdAt travel
// with the tag while updatedPin/updatedAt have no source and stay null.
func TestFenceTagToCOCarriesAuditFields(t *testing.T) {
	created := time.Date(2024, 7, 15, 16, 19, 8, 0, time.FixedZone("CST", 8*3600))
	row := model.FenceTag{TagID: 9, TagName: "VIP"}
	row.Version = sql.NullInt32{Int32: 3, Valid: true}
	row.CreatedPin = "ops001"
	row.CreatedAt = created
	row.IzEnable = sql.NullBool{Bool: true, Valid: true}

	co := FenceTagToCO(row)
	if co.TagId == nil || *co.TagId != 9 {
		t.Fatalf("tagId=%v", co.TagId)
	}
	if co.Version == nil || *co.Version != 3 {
		t.Fatalf("version=%v", co.Version)
	}
	if co.CreatedPin == nil || *co.CreatedPin != "ops001" {
		t.Fatalf("createdPin=%v", co.CreatedPin)
	}
	if co.CreatedAt == nil || *co.CreatedAt != "2024-07-15T16:19:08" {
		t.Fatalf("createdAt=%v", co.CreatedAt)
	}
	if co.UpdatedPin != nil || co.UpdatedAt != nil {
		t.Fatalf("updatedPin/updatedAt must stay null, got %v/%v", co.UpdatedPin, co.UpdatedAt)
	}
}

func TestFenceTagToCOLeavesMissingAuditFieldsNull(t *testing.T) {
	co := FenceTagToCO(model.FenceTag{TagID: 1, TagName: "x"})
	if co.CreatedAt != nil {
		t.Fatalf("expected null createdAt, got %v", *co.CreatedAt)
	}
	if co.CreatedPin != nil {
		t.Fatalf("expected null createdPin, got %v", *co.CreatedPin)
	}
}
