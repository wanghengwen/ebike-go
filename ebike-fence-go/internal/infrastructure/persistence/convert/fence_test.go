package convert

import (
	"database/sql"
	"testing"

	"ebike-fence-go/internal/infrastructure/persistence/model"
)

func TestFenceToEntityMapsParkingFlags(t *testing.T) {
	tbeacon := sql.NullBool{Bool: true, Valid: true}
	izFull := sql.NullInt32{Int32: 1, Valid: true}
	row := model.TFence{
		ID:               1,
		Name:             "P1",
		PointList:        `[[116.4,39.9],[116.41,39.9],[116.41,39.91],[116.4,39.91]]`,
		Tbeacon:          tbeacon,
		IzFullPileNoStop: izFull,
		MaxParkingNumber: 10,
	}
	fe := FenceToEntity(row)
	if fe.Tbeacon == nil || !*fe.Tbeacon {
		t.Fatal("expected tbeacon true")
	}
	if fe.IzFullPileNoStop == nil || *fe.IzFullPileNoStop != 1 {
		t.Fatal("expected izFullPileNoStop=1")
	}
	if len(fe.ParsedPolygon) == 0 {
		t.Fatal("expected parsed polygon")
	}
}
