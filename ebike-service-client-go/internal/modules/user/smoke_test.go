package user

import (
	"encoding/json"
	"testing"
)

func TestSmoke(t *testing.T) {
	// regex compile + basic matches
	if !grayPhonePattern.MatchString("+86-15571260681") {
		t.Fatal("gray phone pattern should match")
	}
	if grayPhonePattern.MatchString("15571260681") {
		t.Fatal("gray phone pattern should not match without prefix")
	}
	if !authNoPattern.MatchString("110101199003077758") {
		t.Fatal("authNo pattern should match a normal id")
	}
	// desensitization (mirrors DesensitizedUtil#main expectations)
	cases := map[string]string{"张三": "*三", "张三丰": "**丰", "张": "*", "zhang san": "***** san", "zhangsan": "*******n"}
	for in, want := range cases {
		if got := *desensitizedName(&in); got != want {
			t.Fatalf("name(%q) = %q, want %q", in, got, want)
		}
	}
	phone := "+86-15571260681"
	if got := *desensitizedIdCard(&phone, 7, 2); got != "+86-155******81" {
		t.Fatalf("idCardNum = %q", got)
	}
	// LongStr round-trip
	var l LongStr
	if err := json.Unmarshal([]byte(`123`), &l); err != nil || l != 123 {
		t.Fatal("LongStr number")
	}
	if err := json.Unmarshal([]byte(`"456"`), &l); err != nil || l != 456 {
		t.Fatal("LongStr string")
	}
	b, _ := json.Marshal(l)
	if string(b) != `"456"` {
		t.Fatalf("LongStr marshal = %s", b)
	}
	// DateTime round-trip
	var d DateTime
	if err := json.Unmarshal([]byte(`"2022-01-02 03:04:05"`), &d); err != nil {
		t.Fatal(err)
	}
	if out, _ := json.Marshal(d); string(out) != `"2022-01-02 03:04:05"` {
		t.Fatalf("DateTime marshal = %s", out)
	}
	if err := json.Unmarshal([]byte(`"2022-01-02T03:04:05"`), &d); err != nil {
		t.Fatal(err)
	}
	// depositedInfo key order
	p := "13800000000"
	want := `{"transaction_id":"13800000000","time_expire":"","gmt_payment":"","out_trade_no":"","time_start":"","total_fee":"","channel":"","prepay_id":""}`
	if got := buildDepositedInfo(&p); got != want {
		t.Fatalf("depositedInfo = %s", got)
	}
	// javaDoubleString
	v := 113.0
	if javaDoubleString(&v) != "113.0" {
		t.Fatal("javaDoubleString 113.0")
	}
	if javaDoubleString(nil) != "null" {
		t.Fatal("javaDoubleString null")
	}
}
