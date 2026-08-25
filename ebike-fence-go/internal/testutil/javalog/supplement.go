package javalog

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
)

var (
	backcarIDRe      = regexp.MustCompile(`ConfigBackcarCO\(id=(\d+)`)
	scrollerMsgIDRe  = regexp.MustCompile(`HomeScrollerMsgCO\(id=(\d+)`)
	specialTipsIDRe  = regexp.MustCompile(`SpecialTipsCO\(id=(\d+)`)
	homeActivityIDRe = regexp.MustCompile(`HomeActivityEntranceCO\(id=(\d+)`)
	faqIDRe          = regexp.MustCompile(`FaqCO\(id=(\d+)`)
)

type serviceContext struct {
	tenantID  string
	serviceID int64
	cmdCtx    map[string]interface{}
	source    string
}

type minedIDContext struct {
	tenantID string
	cmdCtx   map[string]interface{}
	source   string
}

// DeriveSupplementFixtures builds read-only HTTP cases for endpoints absent from Java logs,
// using commandContext/serviceId templates and entity ids mined from log replies.
func DeriveSupplementFixtures(entries []Entry) []Entry {
	ctxByKey := map[string]serviceContext{}
	backcarIDs := map[int64]minedIDContext{}
	scrollerIDs := map[int64]minedIDContext{}
	specialTipsIDs := map[int64]minedIDContext{}
	faqIDs := map[int64]minedIDContext{}
	homeActivityIDs := map[int64]minedIDContext{}

	for _, e := range entries {
		if len(e.Request) == 0 || !json.Valid(e.Request) {
			continue
		}
		var req map[string]interface{}
		if json.Unmarshal(e.Request, &req) != nil {
			continue
		}
		cmdCtx, tenantID := commandContextFromReq(req)
		if cmdCtx == nil || tenantID == "" {
			continue
		}
		if sid := serviceIDFromReq(req); sid > 0 {
			key := tenantID + ":" + strconv.FormatInt(sid, 10)
			if _, ok := ctxByKey[key]; !ok {
				ctxByKey[key] = serviceContext{
					tenantID:  tenantID,
					serviceID: sid,
					cmdCtx:    cmdCtx,
					source:    e.Source,
				}
			}
		}
		mineReplyIDs(e, tenantID, cmdCtx, backcarIDs, scrollerIDs, specialTipsIDs, faqIDs, homeActivityIDs)
	}

	contexts := make([]serviceContext, 0, len(ctxByKey))
	for _, c := range ctxByKey {
		contexts = append(contexts, c)
	}
	sort.Slice(contexts, func(i, j int) bool {
		if contexts[i].tenantID != contexts[j].tenantID {
			return contexts[i].tenantID < contexts[j].tenantID
		}
		return contexts[i].serviceID < contexts[j].serviceID
	})

	var out []Entry
	for _, ctx := range contexts {
		out = append(out, serviceScopedSupplements(ctx)...)
	}
	out = append(out, idScopedSupplements(backcarIDs, "/config/backcar/getConfigById", "id")...)
	out = append(out, idScopedSupplements(scrollerIDs, "/helpConfig/getHomeScrollerMsgById", "id")...)
	out = append(out, idScopedSupplements(specialTipsIDs, "/helpConfig/getSpecialTipsById", "id")...)
	out = append(out, idScopedSupplements(faqIDs, "/helpConfig/getFaqById", "id")...)
	out = append(out, idScopedSupplements(homeActivityIDs, "/helpConfig/getHomeActivityById", "id")...)
	return dedupeEntries(out)
}

func mineReplyIDs(e Entry, tenantID string, cmdCtx map[string]interface{}, backcar, scroller, tips, faq, activity map[int64]minedIDContext) {
	pair := minedIDContext{tenantID: tenantID, cmdCtx: cmdCtx, source: e.Source}
	for _, m := range backcarIDRe.FindAllStringSubmatch(e.Reply, -1) {
		if id, err := strconv.ParseInt(m[1], 10, 64); err == nil {
			backcar[id] = pair
		}
	}
	for _, m := range scrollerMsgIDRe.FindAllStringSubmatch(e.Reply, -1) {
		if id, err := strconv.ParseInt(m[1], 10, 64); err == nil {
			scroller[id] = pair
		}
	}
	for _, m := range specialTipsIDRe.FindAllStringSubmatch(e.Reply, -1) {
		if id, err := strconv.ParseInt(m[1], 10, 64); err == nil {
			tips[id] = pair
		}
	}
	for _, m := range faqIDRe.FindAllStringSubmatch(e.Reply, -1) {
		if id, err := strconv.ParseInt(m[1], 10, 64); err == nil {
			faq[id] = pair
		}
	}
	for _, m := range homeActivityIDRe.FindAllStringSubmatch(e.Reply, -1) {
		if id, err := strconv.ParseInt(m[1], 10, 64); err == nil {
			activity[id] = pair
		}
	}
}

func commandContextFromReq(req map[string]interface{}) (map[string]interface{}, string) {
	raw, ok := req["commandContext"]
	if !ok {
		return nil, ""
	}
	cmdCtx, ok := raw.(map[string]interface{})
	if !ok {
		return nil, ""
	}
	tenantID, _ := cmdCtx["tenantId"].(string)
	return cmdCtx, tenantID
}

func serviceIDFromReq(req map[string]interface{}) int64 {
	if v, ok := req["serviceId"]; ok {
		if id := asInt64(v); id > 0 {
			return id
		}
	}
	if v, ok := req["id"]; ok {
		return asInt64(v)
	}
	return 0
}

func asInt64(v interface{}) int64 {
	switch n := v.(type) {
	case float64:
		return int64(n)
	case json.Number:
		i, _ := n.Int64()
		return i
	case int64:
		return n
	case int:
		return int64(n)
	default:
		return 0
	}
}

func serviceScopedSupplements(ctx serviceContext) []Entry {
	type spec struct {
		url    string
		fields map[string]interface{}
	}
	specs := []spec{
		{"/helpConfig/getFaqByServiceId", map[string]interface{}{"serviceId": ctx.serviceID}},
		{"/helpConfig/getCustomerServiceByServiceId", map[string]interface{}{"serviceId": ctx.serviceID}},
		{"/helpConfig/getGuidePageConfigIzOn", map[string]interface{}{"serviceId": ctx.serviceID}},
		{"/helpConfig/getHomeScrollerMsgByServiceId", map[string]interface{}{"serviceId": ctx.serviceID}},
		{"/config/getAd", map[string]interface{}{"serviceId": ctx.serviceID}},
		{"/config/parkApply/getByServiceId", map[string]interface{}{"id": ctx.serviceID}},
		{"/config/getRidingCarConfig", map[string]interface{}{"id": ctx.serviceID}},
		{"/config/protocol/list", map[string]interface{}{"serviceId": ctx.serviceID}},
		{"/config/protocol/byType", map[string]interface{}{"serviceId": ctx.serviceID, "type": 1}},
	}
	var out []Entry
	for _, s := range specs {
		if e, ok := makeSupplementEntry(s.url, ctx, s.fields); ok {
			out = append(out, e)
		}
	}
	// Tenant/global scoped reads — one case per service context template.
	for _, url := range []string{
		"/config/bigSceen/getConfigByServiceId",
		"/config/creditScore/getConfigByServiceId",
		"/config/protocol/default",
		"/fence/tags/getAll",
	} {
		fields := map[string]interface{}{}
		if url == "/config/protocol/default" {
			fields["type"] = 1
		}
		if e, ok := makeSupplementEntry(url, ctx, fields); ok {
			out = append(out, e)
		}
	}
	return out
}

func idScopedSupplements(ids map[int64]minedIDContext, url, idField string) []Entry {
	keys := make([]int64, 0, len(ids))
	for id := range ids {
		keys = append(keys, id)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	var out []Entry
	for _, id := range keys {
		p := ids[id]
		ctx := serviceContext{tenantID: p.tenantID, cmdCtx: p.cmdCtx, source: p.source}
		if e, ok := makeSupplementEntry(url, ctx, map[string]interface{}{idField: id}); ok {
			out = append(out, e)
		}
	}
	return out
}

func makeSupplementEntry(url string, ctx serviceContext, fields map[string]interface{}) (Entry, bool) {
	if !IsReadOnlyURL(url) || ctx.cmdCtx == nil {
		return Entry{}, false
	}
	body := map[string]interface{}{
		"commandContext": ctx.cmdCtx,
	}
	for k, v := range fields {
		body[k] = v
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return Entry{}, false
	}
	return Entry{
		URL:     url,
		Request: raw,
		Source:  fmt.Sprintf("derived:%s", ctx.source),
	}, true
}

func dedupeEntries(in []Entry) []Entry {
	seen := map[string]struct{}{}
	out := make([]Entry, 0, len(in))
	for _, e := range in {
		key := e.URL + "\x00" + string(e.Request)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, e)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].URL != out[j].URL {
			return out[i].URL < out[j].URL
		}
		return string(out[i].Request) < string(out[j].Request)
	})
	return out
}

// MergeFixtures deduplicates log fixtures and derived supplements.
func MergeFixtures(base, extra []Entry) []Entry {
	return dedupeEntries(append(append([]Entry{}, base...), extra...))
}
