package rediskeys

import "testing"

// Keys mirror Java FenceRedisKey.format placeholders {tenantId}/{id}/{serviceId}.

func TestFenceDetailKeysMatchJava(t *testing.T) {
	tests := []struct {
		fn       func(string, int64) string
		tenantID string
		id       int64
		want     string
	}{
		{ServiceArea, "1003", 1, "fence_serviceArea_1003_1"},
		{Parking, "1003", 192856886438004387, "fence_parking_1003_192856886438004387"},
		{NoParking, "1003", 2, "fence_noParking_1003_2"},
		{BanRiding, "1003", 3, "fence_banRiding_1003_3"},
		{MaintainArea, "1003", 4, "fence_maintainArea_1003_4"},
		{FenceCustom, "1003", 5, "fence_custom_1003_5"},
	}
	for _, tt := range tests {
		if got := tt.fn(tt.tenantID, tt.id); got != tt.want {
			t.Fatalf("got %q want %q", got, tt.want)
		}
	}
}

func TestFenceGeoKeysMatchJava(t *testing.T) {
	tests := []struct {
		fn       func(string) string
		tenantID string
		want     string
	}{
		{ServiceAreaGeo, "1003", "fence_serviceArea_geo_1003"},
		{ParkingGeo, "1003", "fence_parking_geo_1003"},
		{NoParkingGeo, "1003", "fence_noParking_geo_1003"},
		{BanRidingGeo, "1003", "fence_banRiding_geo_1003"},
		{MaintainAreaGeo, "1003", "fence_maintainArea_geo_1003"},
	}
	for _, tt := range tests {
		if got := tt.fn(tt.tenantID); got != tt.want {
			t.Fatalf("got %q want %q", got, tt.want)
		}
	}
}

func TestFenceCustomGeoSearchKeyMatchesJavaFormat(t *testing.T) {
	got := FenceCustomGeoSearchKey("1007")
	want := "fence_custom_geo_1007_{id}"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
