package javalog

import "testing"

func TestParseJavaCOInt64FieldPreservesSnowflakeID(t *testing.T) {
	rep := `{"success":true,"code":"0","msg":"成功","data":BindParkingCO(parkingId=185442561702239423, noParkingId=null, banRidingId=null, maintainAreaId=null, nearParkingId=null, fenceCustomId=null)}`
	got, ok := ParseJavaCOInt64Field(rep, "BindParkingCO", "parkingId")
	if !ok {
		t.Fatal("expected parkingId")
	}
	const want int64 = 185442561702239423
	if got != want {
		t.Fatalf("parkingId=%d want %d", got, want)
	}
}
