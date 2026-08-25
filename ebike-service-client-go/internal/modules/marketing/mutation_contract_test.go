package marketing

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestWatchAdJudgementPinIsNull(t *testing.T) {
	svcID := int64(10)
	body := userWalletJudgementCmd{ServiceId: &svcID, Pin: nil, BuyTime: nil}

	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	s := string(raw)
	if !strings.Contains(s, `"pin":null`) {
		t.Fatalf("watch ad judgement must send pin:null like Java JudgementGatewayImpl, got %s", s)
	}
	if !strings.Contains(s, `"serviceId":10`) {
		t.Fatalf("serviceId missing: %s", s)
	}
}

func TestInviteAcceptDownstreamBody(t *testing.T) {
	svcID := int64(5)
	body := acceptInviteCmd{
		ServiceId: &svcID,
		InviteeId: "invitee-pin",
		InviterId: "inviter-pin",
	}
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	s := string(raw)
	for _, key := range []string{
		`"serviceId":5`,
		`"inviteeId":"invitee-pin"`,
		`"inviterId":"inviter-pin"`,
	} {
		if !strings.Contains(s, key) {
			t.Errorf("accept invite body missing %s in %s", key, s)
		}
	}
}
