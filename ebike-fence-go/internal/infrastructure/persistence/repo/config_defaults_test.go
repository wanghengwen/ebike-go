package repo

import (
	"testing"
)

func TestNewDefaultUseCarRowNOTNULLColumns(t *testing.T) {
	row := NewDefaultUseCarRow(1, 100)
	if !row.RechargeVisibleRange.Valid || row.RechargeVisibleRange.Int32 != 0 {
		t.Fatalf("rechargeVisibleRange want 0, got %+v", row.RechargeVisibleRange)
	}
	if !row.IzAuth.Valid || !row.IzAuth.Bool {
		t.Fatalf("izAuth want true")
	}
}

func TestNewDefaultPayRowNOTNULLColumns(t *testing.T) {
	row := NewDefaultPayRow(1, 100)
	if !row.IzBalanceEnoughReturnBike.Valid {
		t.Fatalf("izBalanceEnoughReturnBike must be set")
	}
}

func TestNewDefaultParkApplyRowNOTNULLColumns(t *testing.T) {
	row := NewDefaultParkApplyRow(1, 100)
	if !row.IzApplyPark.Valid || row.IzApplyPark.Int32 != 1 {
		t.Fatalf("izApplyPark want 1, got %+v", row.IzApplyPark)
	}
}
