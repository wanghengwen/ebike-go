// Package order ports the Java order-domain controllers to Go:
//
//   - OrderController          (/client/order/...)
//   - BillConfigController     (/client/order/config/get)
//   - InvoiceController        (/client/invoice/...)
//   - ReturnBikeAuditController(/client/returnBikeAudit/...)
//   - UserTicketController     (/client/userTicket/...)
//   - DeviceInfoController     (/client/paas/device/...)
package order

import (
	"encoding/json"
	"strconv"
	"time"

	"ebike-service-client-go/internal/pkg/javacompat"
	"github.com/gin-gonic/gin"
)

var (
	isNullRaw    = javacompat.IsNullJSON
	longStr      = javacompat.LongStrPtr
	rawInt       = javacompat.RawInt
	rawInt64Ptr  = javacompat.RawInt64Ptr
	rawStringPtr = javacompat.RawStringPtr
	callForData  = javacompat.CallData
	writeNPE     = javacompat.WriteNPE
)

// RegisterRoutes wires all order module routes.
// Every Java controller in this module is class-annotated @RequestMapping("/client")
// except DeviceInfoController (@RequestMapping("/client/paas/device")).
func RegisterRoutes(r *gin.RouterGroup) {
	client := r.Group("/client")
	{
		// OrderController
		client.POST("/order/calculateCost", calculateCost)
		client.POST("/order/list", listOrders)
		client.POST("/order/closeOrder", closeOrder)
		client.POST("/order/detailLast", detailLast)
		client.POST("/order/updateItinerary", updateItinerary)
		client.POST("/order/detail", orderDetail)
		client.POST("/order/frozen", frozenItinerary)
		client.POST("/order/queryFrozen", queryFrozen)
		client.POST("/order/listInvoicedOrders", listCanInvoiced)

		// BillConfigController
		client.POST("/order/config/get", getBillConfig)

		// InvoiceController
		client.POST("/invoice/create", invoiceCreate)
		client.POST("/invoice/page", invoicePage)

		// ReturnBikeAuditController
		client.POST("/returnBikeAudit/createReturnBikeAudit", createReturnBikeAudit)
		client.POST("/returnBikeAudit/izCanCameraAudit", izCanCameraAudit)
		client.POST("/returnBikeAudit/izCapable", izCapable)

		// UserTicketController
		client.POST("/userTicket/createUserTicket", createUserTicket)
		client.POST("/userTicket/getUserTicket", getUserTicket)

		// DeviceInfoController
		device := client.Group("/paas/device")
		{
			device.POST("/eBikeLocation", eBikeLocation)
			device.POST("/getBlueToothToken", getBlueToothToken)
			device.POST("/carSearchVoice", carSearchVoice)
		}
	}
}

// ---------------------------------------------------------------------------
// Shared helpers
// ---------------------------------------------------------------------------

// ldtLayout mirrors the Java global LocalDateTimeSerializer pattern.
const ldtLayout = "2006-01-02 15:04:05"

// cst8 mirrors Java's ZoneOffset.ofHours(8) used in OrderGatewayImpl
// (and the assumed system default zone of the Java deployment).
var cst8 = time.FixedZone("GMT+8", 8*3600)

func nowMillis() int64 {
	return time.Now().In(cst8).UnixMilli()
}

// ldtParseLayouts: the Java services may emit LocalDateTime either with the
// global "yyyy-MM-dd HH:mm:ss" serializer or in ISO-8601 (Jackson default).
var ldtParseLayouts = []string{
	ldtLayout,
	"2006-01-02 15:04:05.999999999",
	"2006-01-02T15:04:05",
	"2006-01-02T15:04:05.999999999",
}

// stringValueOf mirrors Java String.valueOf(Long): a null value becomes the
// literal string "null" (this really happens, e.g. CBillingConfigCmd.setId).
func stringValueOf(raw json.RawMessage) string {
	return javacompat.StringValueOf(raw)
}

// parseLDTRaw parses a raw JSON LocalDateTime string at +08:00.
func parseLDTRaw(raw json.RawMessage) (time.Time, bool) {
	s := rawStringPtr(raw)
	if s == nil {
		return time.Time{}, false
	}
	for _, layout := range ldtParseLayouts {
		if t, err := time.ParseInLocation(layout, *s, cst8); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

// ldtToMillisPtr mirrors DateUtils.toTimestamp(LocalDateTime): null -> null,
// otherwise the epoch millisecond at the (assumed +08:00) system zone.
func ldtToMillisPtr(raw json.RawMessage) *int64 {
	if t, ok := parseLDTRaw(raw); ok {
		ms := t.UnixMilli()
		return &ms
	}
	return nil
}

// normalizeLDT re-emits a LocalDateTime value the way the Java ObjectMapper
// would (global pattern "yyyy-MM-dd HH:mm:ss"). Unparseable values pass
// through unchanged.
func normalizeLDT(raw json.RawMessage) json.RawMessage {
	if isNullRaw(raw) {
		if raw == nil {
			return nil
		}
		return raw
	}
	if t, ok := parseLDTRaw(raw); ok {
		b, _ := json.Marshal(t.Format(ldtLayout))
		return b
	}
	return raw
}

// javaMinusMonths mirrors Java LocalDateTime.minus(n, ChronoUnit.MONTHS):
// calendar-month arithmetic with day-of-month clamping (Go's AddDate
// normalizes overflow instead, which would differ at month ends).
func javaMinusMonths(t time.Time, months int) time.Time {
	y, m, d := t.Date()
	total := y*12 + int(m) - 1 - months
	ny, nm := total/12, time.Month(total%12+1)
	lastDay := time.Date(ny, nm+1, 0, 0, 0, 0, 0, t.Location()).Day()
	if d > lastDay {
		d = lastDay
	}
	return time.Date(ny, nm, d, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
}

// utf16Len mirrors Java String.length() (UTF-16 code units) for @Size checks.
func utf16Len(s string) int {
	n := 0
	for _, r := range s {
		n++
		if r > 0xFFFF {
			n++
		}
	}
	return n
}

// pageDTOOut mirrors the JSON shape of com.xyy.dto.PageDTO with the Java
// global serializers applied: count is a primitive long -> string, pageNum /
// pageSize are ints, searchCount a boolean. Field order matches the Java
// class declaration.
type pageDTOOut struct {
	Count       string          `json:"count"`
	PageNum     int64           `json:"pageNum"`
	PageSize    int64           `json:"pageSize"`
	Orders      json.RawMessage `json:"orders"`
	SearchCount bool            `json:"searchCount"`
	List        interface{}     `json:"list"`
}

// pageDTOIn parses a downstream PageDTO while keeping records raw.
type pageDTOIn struct {
	Count       json.RawMessage   `json:"count"`
	PageNum     json.RawMessage   `json:"pageNum"`
	PageSize    json.RawMessage   `json:"pageSize"`
	Orders      json.RawMessage   `json:"orders"`
	SearchCount json.RawMessage   `json:"searchCount"`
	List        []json.RawMessage `json:"list"`
}

// pageOutFrom rebuilds the downstream page envelope exactly like
// PageConvertorUtils.convert(PageDTO, ...) does (count/pageNum/pageSize/
// searchCount/orders copied from the downstream page; Java field defaults
// apply when the downstream omitted them).
func pageOutFrom(in *pageDTOIn) pageDTOOut {
	out := pageDTOOut{Count: "0", PageNum: 1, PageSize: 10, SearchCount: true, Orders: in.Orders}
	if v, ok := rawInt(in.Count); ok {
		out.Count = strconv.FormatInt(v, 10)
	} else if p := longStr(in.Count); p != nil {
		out.Count = *p
	}
	if v, ok := rawInt(in.PageNum); ok {
		out.PageNum = v
	}
	if v, ok := rawInt(in.PageSize); ok {
		out.PageSize = v
	}
	if !isNullRaw(in.SearchCount) {
		var b bool
		if err := json.Unmarshal(in.SearchCount, &b); err == nil {
			out.SearchCount = b
		}
	}
	return out
}
