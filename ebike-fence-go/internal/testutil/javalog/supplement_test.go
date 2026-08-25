package javalog_test

import (
	"testing"

	"ebike-fence-go/internal/testutil/javalog"
)

func TestDeriveSupplementFixtures(t *testing.T) {
	entries := []javalog.Entry{
		{
			URL:     "/helpConfig/getHomeNav",
			Request: []byte(`{"commandContext":{"tenantId":"1003","traceId":"t1"},"serviceId":338362359727786125}`),
			Reply:   `ConfigBackcarCO(id=339660543657779464, serviceId=338362359727786125) HomeScrollerMsgCO(id=240355586884507544, serviceId=239659900966803395)`,
			Source:  "ebike-fence-1.log",
		},
		{
			URL:     "/helpConfig/getSpecialTipsByServiceId",
			Request: []byte(`{"commandContext":{"tenantId":"1004","traceId":"t2"},"serviceId":239659900966803395}`),
			Reply:   `SpecialTipsCO(id=246730273981991810, serviceId=239659900966803395)`,
			Source:  "ebike-fence-1.log",
		},
	}
	got := javalog.DeriveSupplementFixtures(entries)
	if len(got) == 0 {
		t.Fatal("expected supplement fixtures")
	}
	urls := map[string]int{}
	for _, e := range got {
		urls[e.URL]++
	}
	for _, want := range []string{
		"/helpConfig/getFaqByServiceId",
		"/helpConfig/getCustomerServiceByServiceId",
		"/config/backcar/getConfigById",
		"/helpConfig/getSpecialTipsById",
		"/fence/tags/getAll",
	} {
		if urls[want] == 0 {
			t.Fatalf("missing derived url %s: %#v", want, urls)
		}
	}
}
