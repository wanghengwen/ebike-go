package convert

import (
	"testing"
)

func TestUnmarshalBaseItemCO_JavaDOUserTicketPhotoWaysString(t *testing.T) {
	raw := `{"id":372170774628277787,"serviceId":364881328207300695,"izOpenInvoice":true,"userTicketPhotoWays":"0"}`
	co, err := UnmarshalBaseItemCO(raw)
	if err != nil {
		t.Fatal(err)
	}
	if co.Id == nil || *co.Id != 372170774628277787 {
		t.Fatalf("id=%v", co.Id)
	}
	if co.IzOpenInvoice == nil || !*co.IzOpenInvoice {
		t.Fatalf("izOpenInvoice=%v", co.IzOpenInvoice)
	}
	if len(co.UserTicketPhotoWays) != 1 || co.UserTicketPhotoWays[0] != 0 {
		t.Fatalf("userTicketPhotoWays=%v", co.UserTicketPhotoWays)
	}
}

func TestUnmarshalBaseItemCO_GoCOUserTicketPhotoWaysArray(t *testing.T) {
	raw := `{"id":1,"serviceId":2,"izOpenInvoice":false,"userTicketPhotoWays":[0,1]}`
	co, err := UnmarshalBaseItemCO(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(co.UserTicketPhotoWays) != 2 || co.UserTicketPhotoWays[1] != 1 {
		t.Fatalf("userTicketPhotoWays=%v", co.UserTicketPhotoWays)
	}
}
