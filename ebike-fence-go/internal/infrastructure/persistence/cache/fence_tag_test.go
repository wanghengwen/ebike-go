package cache

import "testing"

func TestUnmarshalFenceTagList_JavaRedisJSON(t *testing.T) {
	raw := `[{"id":1,"tagName":"VIP","izEnable":true,"izDel":false,"updatedAt":"2024-07-15T16:19:08","tenantId":"1004"}]`
	rows, err := UnmarshalFenceTagList(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].TagName != "VIP" {
		t.Fatalf("unexpected rows: %+v", rows)
	}
}
