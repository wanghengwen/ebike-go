package cache

import (
	"database/sql"
	"testing"

	"ebike-fence-go/internal/infrastructure/persistence/model"
)

func TestUnmarshalHomeNavList_JavaRedisJSON(t *testing.T) {
	raw := `[{"id":1,"serviceId":2,"carType":0,"name":"n","icon":"i","jumpPage":"{}","izOn":false,"izDel":false,"updatedAt":"2024-07-15T16:19:08","updatedPin":"p","tenantId":"1004","version":0}]`
	rows, err := UnmarshalHomeNavList(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("len=%d", len(rows))
	}
	if !rows[0].IzDel.Valid || rows[0].IzDel.Bool {
		t.Fatalf("izDel=%+v", rows[0].IzDel)
	}
	if rows[0].UpdatedAt.IsZero() {
		t.Fatal("updatedAt not parsed")
	}
}

func TestUnmarshalSpecialTipsList_JavaAndGoIzDelFormats(t *testing.T) {
	javaRaw := `[{"id":1,"serviceId":364848372922716182,"popUpType":0,"bgUrl":"[]","title":"t","popUpTime":1,"izOn":false,"izDel":false,"tenantId":"1007","version":0}]`
	rows, err := UnmarshalSpecialTipsList(javaRaw)
	if err != nil {
		t.Fatalf("java format: %v", err)
	}
	if len(rows) != 1 || rows[0].ID != 1 {
		t.Fatalf("java format rows=%+v", rows)
	}

	goRaw := `[{"id":2,"serviceId":364848372922716182,"popUpType":0,"title":"t2","izDel":{"Bool":false,"Valid":true}}]`
	rows, err = UnmarshalSpecialTipsList(goRaw)
	if err != nil {
		t.Fatalf("go sql.NullBool format: %v", err)
	}
	if len(rows) != 1 || rows[0].ID != 2 {
		t.Fatalf("go format rows=%+v", rows)
	}
}

func TestMarshalSpecialTipsListRoundTrip(t *testing.T) {
	falseVal := false
	rows := []model.TConfigSpecialTips{{
		ConfigBaseDO: model.ConfigBaseDO{TenantID: "1007", IzDel: sql.NullBool{Bool: false, Valid: true}},
		ID:           99,
		ServiceID:    364848372922716182,
		Title:        "t",
		PopUpTime:    1,
		IzOn:         &falseVal,
	}}
	raw, err := MarshalSpecialTipsList(rows)
	if err != nil {
		t.Fatal(err)
	}
	got, err := UnmarshalSpecialTipsList(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].ID != 99 {
		t.Fatalf("got=%+v", got)
	}
}
