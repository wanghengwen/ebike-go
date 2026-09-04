package dto

import (
	"encoding/json"
	"testing"
)

func marshalKeys(t *testing.T, v interface{}) map[string]json.RawMessage {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("unmarshal %s: %v", raw, err)
	}
	return m
}

func requireKeys(t *testing.T, m map[string]json.RawMessage, keys ...string) {
	t.Helper()
	for _, k := range keys {
		if _, ok := m[k]; !ok {
			t.Errorf("missing key %q", k)
		}
	}
}

func rejectKeys(t *testing.T, m map[string]json.RawMessage, keys ...string) {
	t.Helper()
	for _, k := range keys {
		if _, ok := m[k]; ok {
			t.Errorf("unexpected key %q", k)
		}
	}
}

// Java CreditScoreConfigCO and AlarmContactCO do not extend Command.
func TestConfigCOsCarryNoCommandContext(t *testing.T) {
	rejectKeys(t, marshalKeys(t, CreditScoreConfigCO{}), "commandContext")
	rejectKeys(t, marshalKeys(t, AlarmContactCO{}), "commandContext")
}

func TestConfigPayCOHasAutoRefundAfterOrder(t *testing.T) {
	requireKeys(t, marshalKeys(t, ConfigPayCO{}), "izAutoRefundAfterOrder")
}

func TestConfigBaseItemCOHasAlarmContactFields(t *testing.T) {
	requireKeys(t, marshalKeys(t, ConfigBaseItemCO{}),
		"abnormalMovementContact", "batteryRemovalContact")
}

// Java FenceTagCO exposes tagId/tagName/version plus audit pins, and has no izEnable.
func TestFenceTagCOMatchesJavaShape(t *testing.T) {
	m := marshalKeys(t, FenceTagCO{})
	requireKeys(t, m, "tagId", "tagName", "version", "createdPin", "createdAt", "updatedPin", "updatedAt")
	rejectKeys(t, m, "izEnable")
	if len(m) != 7 {
		t.Errorf("expected 7 keys, got %d: %v", len(m), m)
	}
}

// Java ParkingCO declares neither formulateDirection nor izOpenAllDay; refTags is a
// List<FenceTagCO> and RFIDs serializes as a "rfids" string array.
func TestParkingCOMatchesJavaShape(t *testing.T) {
	tagID := int64(7)
	tagName := "VIP"
	co := ParkingCO{
		RefTags: []FenceTagCO{{TagId: &tagID, TagName: &tagName}},
		Rfids:   []string{"A1", "B2"},
	}
	m := marshalKeys(t, co)
	rejectKeys(t, m, "formulateDirection", "izOpenAllDay")
	requireKeys(t, m, "refTags", "rfids")
	if string(m["rfids"]) != `["A1","B2"]` {
		t.Errorf("rfids should be a string array, got %s", m["rfids"])
	}
	var tags []map[string]json.RawMessage
	if err := json.Unmarshal(m["refTags"], &tags); err != nil {
		t.Fatalf("refTags should be an object array: %v", err)
	}
	if len(tags) != 1 || string(tags[0]["tagName"]) != `"VIP"` {
		t.Errorf("unexpected refTags payload: %s", m["refTags"])
	}
}

func TestParkingCOEmptyCollectionsAreNull(t *testing.T) {
	m := marshalKeys(t, ParkingCO{})
	if string(m["refTags"]) != "null" {
		t.Errorf("empty refTags must be null like Java's guarded setter, got %s", m["refTags"])
	}
	if string(m["rfids"]) != "null" {
		t.Errorf("empty rfids must be null like Java buildRfids, got %s", m["rfids"])
	}
}

// The request side keeps Java ParkingCmd's List<Long> refTags plus the two fields that
// only exist on the write path.
func TestParkingCmdBindsIDListRefTags(t *testing.T) {
	var cmd ParkingCmd
	body := `{"refTags":[11,12],"formulateDirection":20,"izOpenAllDay":true,"name":"p1"}`
	if err := json.Unmarshal([]byte(body), &cmd); err != nil {
		t.Fatal(err)
	}
	if len(cmd.RefTags) != 2 || cmd.RefTags[0] != 11 {
		t.Errorf("expected refTags [11 12], got %v", cmd.RefTags)
	}
	if cmd.FormulateDirection == nil || *cmd.FormulateDirection != 20 {
		t.Errorf("expected formulateDirection 20, got %v", cmd.FormulateDirection)
	}
	if cmd.IzOpenAllDay == nil || !*cmd.IzOpenAllDay {
		t.Errorf("expected izOpenAllDay true, got %v", cmd.IzOpenAllDay)
	}
	if cmd.Name != "p1" {
		t.Errorf("expected embedded ParkingCO fields to still bind, got %q", cmd.Name)
	}
}

// HomeNavDO supplies neither ids nor updatedName, so ConvertorHelper leaves both null.
func TestHomeNavCOHasNullIdsAndUpdatedName(t *testing.T) {
	m := marshalKeys(t, HomeNavCO{})
	requireKeys(t, m, "ids", "updatedName")
	if string(m["ids"]) != "null" {
		t.Errorf("expected ids null, got %s", m["ids"])
	}
	if string(m["updatedName"]) != "null" {
		t.Errorf("expected updatedName null, got %s", m["updatedName"])
	}
}

func TestCustomerServiceCOUpdatedNameIsNullWhenUnset(t *testing.T) {
	m := marshalKeys(t, CustomerServiceCO{})
	if string(m["updatedName"]) != "null" {
		t.Errorf("expected updatedName null, got %s", m["updatedName"])
	}
}

// Java uses @JsonFormat("yyyy-MM-dd HH:mm:ss") for these, never RFC3339.
func TestHelpConfigTimesUseSpaceSeparatedStrings(t *testing.T) {
	at := "2024-07-15 16:19:08"
	m := marshalKeys(t, HomeActivityEntranceCO{StartTime: &at, EndTime: &at})
	if string(m["startTime"]) != `"2024-07-15 16:19:08"` {
		t.Errorf("unexpected startTime: %s", m["startTime"])
	}
	faq := marshalKeys(t, FaqCO{CreatedAt: &at})
	if string(faq["createdAt"]) != `"2024-07-15 16:19:08"` {
		t.Errorf("unexpected createdAt: %s", faq["createdAt"])
	}
}

func TestHomeActivityEntranceCOCarriesNoAuditFields(t *testing.T) {
	rejectKeys(t, marshalKeys(t, HomeActivityEntranceCO{}),
		"tenantId", "createdPin", "createdAt", "updatedPin", "updatedAt", "version", "izDel", "orderWeights")
}
