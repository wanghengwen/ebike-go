package controller

import (
	"testing"

	"ebike-fence-go/internal/api/dto"
	"ebike-fence-go/internal/infrastructure/persistence/model"
)

func TestCmdToModelGuidePageMapping(t *testing.T) {
	// 1. Prepare DTO with ServiceId and other fields
	var serviceID int64 = 364881328207300695
	var pageNumEsc int = 1
	var visibleRange int = 0
	var frequency int = 0
	allowSuperEsc := false
	izOn := false
	byRegister := true
	byTags := false

	req := dto.GuidePageConfigCmd{
		ServiceId:     &serviceID,
		AllowSuperEsc: &allowSuperEsc,
		PageNumEsc:    &pageNumEsc,
		VisibleRange:  &visibleRange,
		Frequency:     &frequency,
		IzOn:          &izOn,
		ByRegister:    &byRegister,
		ByTags:        &byTags,
		TagIds:        "",
	}

	// 2. Call cmdToModel
	row, err := cmdToModel[model.TConfigGuidePage](req)
	if err != nil {
		t.Fatalf("cmdToModel failed: %v", err)
	}

	// 3. Verify mappings
	if row.ServiceID != serviceID {
		t.Errorf("expected ServiceID %d, got %d", serviceID, row.ServiceID)
	}
	if row.AllowSuperEsc == nil || *row.AllowSuperEsc != allowSuperEsc {
		t.Errorf("expected AllowSuperEsc %v, got %v", allowSuperEsc, row.AllowSuperEsc)
	}
	if row.PageNumEsc != pageNumEsc {
		t.Errorf("expected PageNumEsc %d, got %d", pageNumEsc, row.PageNumEsc)
	}
	if row.VisibleRange != visibleRange {
		t.Errorf("expected VisibleRange %d, got %d", visibleRange, row.VisibleRange)
	}
	if row.Frequency != frequency {
		t.Errorf("expected Frequency %d, got %d", frequency, row.Frequency)
	}
	if row.IzOn == nil || *row.IzOn != izOn {
		t.Errorf("expected IzOn %v, got %v", izOn, row.IzOn)
	}
	if row.ByRegister == nil || *row.ByRegister != byRegister {
		t.Errorf("expected ByRegister %v, got %v", byRegister, row.ByRegister)
	}
	if row.ByTags == nil || *row.ByTags != byTags {
		t.Errorf("expected ByTags %v, got %v", byTags, row.ByTags)
	}
}

func TestCmdToModelSpecialTipsMapping(t *testing.T) {
	// 1. Prepare DTO with ServiceId and other fields
	var serviceID int64 = 364881328207300695
	var popUpType int = 2
	izSubtitle := true
	izButton := false
	izOn := false

	req := dto.SpecialTipsCmd{
		ServiceId:  &serviceID,
		PopUpType:  &popUpType,
		BgUrl:      "https://example.com/bg.png",
		BgColor:    "#FFFFFF",
		Title:      "Special Tip",
		TitleColor: "#000000",
		IzSubtitle: &izSubtitle,
		Subtitle:   "Sub",
		IzButton:   &izButton,
		IzOn:       &izOn,
	}

	// 2. Call cmdToModel
	row, err := cmdToModel[model.TConfigSpecialTips](req)
	if err != nil {
		t.Fatalf("cmdToModel failed: %v", err)
	}

	// 3. Verify mappings
	if row.ServiceID != serviceID {
		t.Errorf("expected ServiceID %d, got %d", serviceID, row.ServiceID)
	}
	if row.PopUpType != popUpType {
		t.Errorf("expected PopUpType %d, got %d", popUpType, row.PopUpType)
	}
	if row.BgUrl != req.BgUrl {
		t.Errorf("expected BgUrl %s, got %s", req.BgUrl, row.BgUrl)
	}
	if row.Title != req.Title {
		t.Errorf("expected Title %s, got %s", req.Title, row.Title)
	}
	if row.IzSubtitle == nil || *row.IzSubtitle != izSubtitle {
		t.Errorf("expected IzSubtitle %v, got %v", izSubtitle, row.IzSubtitle)
	}
	if row.IzButton == nil || *row.IzButton != izButton {
		t.Errorf("expected IzButton %v, got %v", izButton, row.IzButton)
	}
}
