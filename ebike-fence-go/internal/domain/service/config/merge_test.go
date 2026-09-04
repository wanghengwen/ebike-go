package configsvc

import (
	"database/sql"
	"testing"

	"ebike-fence-go/internal/api/dto"
	"ebike-fence-go/internal/infrastructure/persistence/model"
)

func TestMergeJSONPopulatesSQLNullFields(t *testing.T) {
	yes := true
	n := 7
	way := "0,1"
	cmd := dto.ConfigPayCmd{
		IzBalanceEnoughReturnBike: &yes,
		IzMakeupBalancePay:        &yes,
		IzAutoRefundAfterOrder:    &yes,
		NotifyInterval:            &n,
		RemindWay:                 &way,
	}
	row := model.ConfigPay{ID: 1, ServiceID: 99}
	mergeJSON(&row, cmd)

	assertNullBool(t, "IzBalanceEnoughReturnBike", row.IzBalanceEnoughReturnBike, true)
	assertNullBool(t, "IzMakeupBalancePay", row.IzMakeupBalancePay, true)
	assertNullBool(t, "IzAutoRefundAfterOrder", row.IzAutoRefundAfterOrder, true)
	assertNullInt32(t, "NotifyInterval", row.NotifyInterval, 7)
	assertNullString(t, "RemindWay", row.RemindWay, "0,1")
	if row.IzNotifyUnpaidOrder.Valid {
		t.Fatalf("omitted field must stay invalid, got %+v", row.IzNotifyUnpaidOrder)
	}
	if row.ServiceID != 99 || row.ID != 1 {
		t.Fatalf("pre-set identity fields must be preserved, got id=%d serviceId=%d", row.ID, row.ServiceID)
	}
}

func TestMergeJSONCreditScoreFloatAndInt(t *testing.T) {
	on := true
	score := 100.5
	days := 3
	cmd := dto.CreditScoreConfigCmd{
		IzCreditScore:      &on,
		Score:              &score,
		FirstNoRiddingDays: &days,
	}
	var row model.CreditScoreConfig
	mergeJSON(&row, cmd)
	assertNullBool(t, "IzCreditScore", row.IzCreditScore, true)
	if !row.Score.Valid || row.Score.Float64 != 100.5 {
		t.Fatalf("Score=%+v", row.Score)
	}
	assertNullInt32(t, "FirstNoRiddingDays", row.FirstNoRiddingDays, 3)
}

func TestMergeJSONUserTicketPhotoWaysJoinsInts(t *testing.T) {
	cmd := dto.ConfigBaseItemCmd{
		UserTicketPhotoWays: []int{0, 1},
		IzOpenInvoice:       boolPtr(true),
	}
	var row model.ConfigBaseItem
	mergeJSON(&row, cmd)
	assertNullString(t, "UserTicketPhotoWays", row.UserTicketPhotoWays, "0,1")
	assertNullBool(t, "IzOpenInvoice", row.IzOpenInvoice, true)
}

func TestMergeJSONUseCarRecoveryData(t *testing.T) {
	raw := "2024-07-15 16:19:08"
	cmd := dto.ConfigUseCarCmd{
		RecoveryData:      &raw,
		RechargeBeforeUse: boolPtr(true),
		RechargeCost:      intPtr(100),
		MinAge:            intPtr(16),
	}
	var row model.ConfigUseCar
	mergeJSON(&row, cmd)
	assertNullBool(t, "RechargeBeforeUse", row.RechargeBeforeUse, true)
	assertNullInt32(t, "RechargeCost", row.RechargeCost, 100)
	assertNullInt32(t, "MinAge", row.MinAge, 16)
	if !row.RecoveryData.Valid {
		t.Fatal("RecoveryData must be set")
	}
	got := row.RecoveryData.Time.Format("2006-01-02 15:04:05")
	if got != raw {
		t.Fatalf("RecoveryData=%q want %q", got, raw)
	}
}

func TestMergeJSONNullSkipsField(t *testing.T) {
	row := model.ConfigPay{
		IzBalanceEnoughReturnBike: sql.NullBool{Bool: true, Valid: true},
	}
	// Explicit JSON null must not clear an already-populated column (skip-null).
	mergeJSON(&row, map[string]interface{}{
		"izBalanceEnoughReturnBike": nil,
		"notifyInterval":            5,
	})
	assertNullBool(t, "IzBalanceEnoughReturnBike", row.IzBalanceEnoughReturnBike, true)
	assertNullInt32(t, "NotifyInterval", row.NotifyInterval, 5)
}

func TestMergeJSONDTOToDTOKeepsPointers(t *testing.T) {
	yes := true
	cost := 50
	src := dto.ConfigUseCarCO{
		RechargeBeforeUse: &yes,
		RechargeCost:      &cost,
		StopServiceNotice: strPtr("停运"),
	}
	var dst dto.ConfigUseCarCmd
	mergeJSON(&dst, src)
	if dst.RechargeBeforeUse == nil || !*dst.RechargeBeforeUse {
		t.Fatalf("RechargeBeforeUse=%v", dst.RechargeBeforeUse)
	}
	if dst.RechargeCost == nil || *dst.RechargeCost != 50 {
		t.Fatalf("RechargeCost=%v", dst.RechargeCost)
	}
	if dst.StopServiceNotice == nil || *dst.StopServiceNotice != "停运" {
		t.Fatalf("StopServiceNotice=%v", dst.StopServiceNotice)
	}
}

func TestMergeJSONRidingPermission(t *testing.T) {
	cmd := dto.RidingPermissionCmd{
		IzDeposit: boolPtr(true),
		Deposit:   intPtr(19900),
		Career:    strPtr("学生"),
	}
	var row model.RidingPermission
	mergeJSON(&row, cmd)
	assertNullBool(t, "IzDeposit", row.IzDeposit, true)
	assertNullInt32(t, "Deposit", row.Deposit, 19900)
	assertNullString(t, "Career", row.Career, "学生")
}

func assertNullBool(t *testing.T, name string, got sql.NullBool, want bool) {
	t.Helper()
	if !got.Valid || got.Bool != want {
		t.Fatalf("%s=%+v want Valid=true Bool=%v", name, got, want)
	}
}

func assertNullInt32(t *testing.T, name string, got sql.NullInt32, want int32) {
	t.Helper()
	if !got.Valid || got.Int32 != want {
		t.Fatalf("%s=%+v want Valid=true Int32=%d", name, got, want)
	}
}

func assertNullString(t *testing.T, name string, got sql.NullString, want string) {
	t.Helper()
	if !got.Valid || got.String != want {
		t.Fatalf("%s=%+v want Valid=true String=%q", name, got, want)
	}
}

func boolPtr(v bool) *bool    { return &v }
func intPtr(v int) *int       { return &v }
func strPtr(v string) *string { return &v }
