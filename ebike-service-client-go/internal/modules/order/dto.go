package order

import (
	"encoding/json"

	"ebike-service-client-go/internal/api/dto"
)

// ---------------------------------------------------------------------------
// Validation message maps (JSON field name -> hibernate-validator English
// default message of the Java annotation on that field).
// ---------------------------------------------------------------------------

var pageMsgs = map[string]string{
	"traceId":  "must not be empty", // @NotEmpty (ClientDTO)
	"tenantId": "must not be empty", // @NotEmpty (ClientDTO)
}

// calculateMsgs covers CCalculateCmd (used by /order/calculateCost and /order/frozen).
var calculateMsgs = map[string]string{
	"traceId":   "must not be empty",
	"tenantId":  "must not be empty",
	"orderId":   "must not be null", // @NotNull
	"userPin":   "must not be null", // @NotNull (String)
	"imei":      "must not be null", // @NotNull (String)
	"serviceId": "must not be null", // @NotNull
	"type":      "must not be null", // @NotNull
}

// closeOrderMsgs covers CCloseOrderCmd.
var closeOrderMsgs = map[string]string{
	"traceId":  "must not be empty",
	"tenantId": "must not be empty",
	"orderId":  "must not be null",
	"payType":  "must not be null",
	"izPaid":   "must not be null",
}

// updateItineraryMsgs covers CUpdateItineraryCmd. The Java annotations carry
// custom message templates ({updateItinerary.imei} / {updateItinerary.itinerary})
// whose keys are NOT defined in the i18n bundles, so hibernate-validator keeps
// the literal template text as the message.
var updateItineraryMsgs = map[string]string{
	"traceId":   "must not be empty",
	"tenantId":  "must not be empty",
	"imei":      "{updateItinerary.imei}",
	"itinerary": "{updateItinerary.itinerary}",
}

// orderIdMsgs covers ClientOrderIdQuery.
var orderIdMsgs = map[string]string{
	"traceId":  "must not be empty",
	"tenantId": "must not be empty",
	"orderId":  "must not be null",
}

// billConfigMsgs covers UserBillConfigQuery.
var billConfigMsgs = map[string]string{
	"traceId":   "must not be empty",
	"tenantId":  "must not be empty",
	"serviceId": "must not be null",  // @NotNull
	"userPin":   "must not be blank", // @NotBlank
}

// returnBikeAuditMsgs covers ReturnBikeAuditDTO (default validation group).
var returnBikeAuditMsgs = map[string]string{
	"traceId":    "must not be empty",
	"tenantId":   "must not be empty",
	"orderId":    "must not be null",               // @NotNull
	"applyType":  "must not be null",               // @NotNull
	"photoUrl":   "must not be blank",              // @NotBlank
	"userReason": "size must be between 0 and 50",  // @Size(max = 50)
}

// cameraAuditMsgs covers CameraAuditDTO.
var cameraAuditMsgs = map[string]string{
	"traceId":   "must not be empty",
	"tenantId":  "must not be empty",
	"serviceId": "must not be null",
	"orderId":   "must not be null",
	"carId":     "must not be blank", // @NotBlank
}

// deviceLocationMsgs covers ClientDeviceLocationQry.
var deviceLocationMsgs = map[string]string{
	"traceId":   "must not be empty",
	"tenantId":  "must not be empty",
	"serviceId": "must not be null",
	"lat":       "must not be null",
	"lng":       "must not be null",
}

// imeiMsgs covers ClientBluetoothTokenQry / ClientCarSearchVoiceCmd.
var imeiMsgs = map[string]string{
	"traceId":  "must not be empty",
	"tenantId": "must not be empty",
	"imei":     "must not be blank", // @NotBlank
}

// ---------------------------------------------------------------------------
// Unvalidated common DTOs. Java handlers whose @RequestBody parameter carries
// no @Validated perform NO bean validation, so these mirror ClientDTO /
// PageClientDTO without binding tags.
// ---------------------------------------------------------------------------

type clientDTONoValid struct {
	TraceId       string   `json:"traceId"`
	TenantId      string   `json:"tenantId"`
	Platform      string   `json:"platform"`
	DeviceId      string   `json:"deviceId"`
	Version       string   `json:"version"`
	Ip            string   `json:"ip"`
	Longitude     *float64 `json:"longitude"`
	Latitude      *float64 `json:"latitude"`
	Source        string   `json:"source"`
	StressTesting bool     `json:"stressTesting"`
}

// clientDTOPtr converts the unvalidated body into a dto.ClientDTO for
// middleware.CompleteCommandContext.
func (s *clientDTONoValid) clientDTOPtr() *dto.ClientDTO {
	return &dto.ClientDTO{
		TraceId:       s.TraceId,
		TenantId:      s.TenantId,
		Platform:      s.Platform,
		DeviceId:      s.DeviceId,
		Version:       s.Version,
		Ip:            s.Ip,
		Longitude:     s.Longitude,
		Latitude:      s.Latitude,
		Source:        s.Source,
		StressTesting: s.StressTesting,
	}
}

// pageClientNoValid mirrors PageClientDTO without validation
// (UserTicketController.getUserTicket binds it with no @Validated).
// Java field defaults: pageNum=1, pageSize=10, searchCount=true.
type pageClientNoValid struct {
	clientDTONoValid
	PageNum      int             `json:"pageNum"`
	PageSize     int             `json:"pageSize"`
	Orders       []dto.OrderItem `json:"orders"`
	SearchCount  *bool           `json:"searchCount"`
	LastRecordId string          `json:"lastRecordId"`
}

// normalizeSearchCount materializes the Java field default searchCount=true.
func (p *pageClientNoValid) normalizeSearchCount() {
	if p.SearchCount == nil {
		t := true
		p.SearchCount = &t
	}
}

// ---------------------------------------------------------------------------
// OrderController request DTOs
// ---------------------------------------------------------------------------

// cCalculateCmd matches Java CCalculateCmd extends ClientDTO (@Validated).
type cCalculateCmd struct {
	dto.ClientDTO
	OrderId          *int64  `json:"orderId" binding:"required"`
	UserPin          *string `json:"userPin" binding:"required"`
	Imei             *string `json:"imei" binding:"required"`
	ServiceId        *int64  `json:"serviceId" binding:"required"`
	ParkingFenceId   *int    `json:"parkingFenceId"`
	NoParkingFenceId *int    `json:"noParkingFenceId"`
	DeviceLocation   *int    `json:"deviceLocation"`
	Distance         *int    `json:"distance"`
	ApplyExemption   *bool   `json:"applyExemption"`
	Type             *int    `json:"type" binding:"required"`
}

// calculateCmdBody matches the downstream com.xyy.ebike.order.api.dto.CalculateCmd
// as Java actually populates it via BeanUtils.copyProperties:
//   - parkingFenceId / noParkingFenceId are Integer on the client DTO but Long
//     on the Cmd, so BeanUtils silently SKIPS them -> always null downstream;
//   - applyExemption does not exist on the Cmd -> dropped;
//   - izHelmetReign is never set -> null.
type calculateCmdBody struct {
	OrderId          *int64  `json:"orderId"`
	UserPin          *string `json:"userPin"`
	Imei             *string `json:"imei"`
	ServiceId        *int64  `json:"serviceId"`
	ParkingFenceId   *int64  `json:"parkingFenceId"`
	NoParkingFenceId *int64  `json:"noParkingFenceId"`
	DeviceLocation   *int    `json:"deviceLocation"`
	Distance         *int    `json:"distance"`
	Type             *int    `json:"type"`
	IzHelmetReign    *bool   `json:"izHelmetReign"`
}

func (r *cCalculateCmd) toCalculateBody() calculateCmdBody {
	return calculateCmdBody{
		OrderId:        r.OrderId,
		UserPin:        r.UserPin,
		Imei:           r.Imei,
		ServiceId:      r.ServiceId,
		DeviceLocation: r.DeviceLocation,
		Distance:       r.Distance,
		Type:           r.Type,
	}
}

// orderPageQuery matches Java OrderPageQuery extends PageClientDTO
// (NOT validated on /order/list and /order/listInvoicedOrders).
// It is forwarded directly: every same-named field of the downstream
// com.xyy.ebike.order.api.dto.OrderPageQuery has an identical type (including
// tenantId, which Java copies from the ClientDTO base!), and the extra
// client-only fields are ignored by the downstream Jackson.
type orderPageQuery struct {
	clientDTONoValid
	PageNum       int             `json:"pageNum"`
	PageSize      int             `json:"pageSize"`
	Orders        []dto.OrderItem `json:"orders"`
	SearchCount   *bool           `json:"searchCount"`
	LastRecordId  string          `json:"lastRecordId"`
	OrderId       *int64          `json:"orderId"`
	UserPin       *string         `json:"userPin"`
	ServiceId     *int64          `json:"serviceId"`
	Mile          *int            `json:"mile"`
	IzPaid        *int            `json:"izPaid"`
	IzComplained  *int            `json:"izComplained"`
	InvoiceId     *int64          `json:"invoiceId"`
	Imei          *string         `json:"imei"`
	CarId         *string         `json:"carId"`
	RidingTime    *int64          `json:"ridingTime"`
	Type          *int            `json:"type"`
	StartTime     []int64         `json:"startTime"`
	EndTime       []int64         `json:"endTime"`
	MinCost       *int            `json:"minCost"`
	MaxCost       *int            `json:"maxCost"`
	IzCanInvoiced *bool           `json:"izCanInvoiced"`
}

func (p *orderPageQuery) normalizeSearchCount() {
	if p.SearchCount == nil {
		t := true
		p.SearchCount = &t
	}
}

// cCloseOrderCmd matches Java CCloseOrderCmd extends ClientDTO (@Validated).
type cCloseOrderCmd struct {
	dto.ClientDTO
	OrderId      *int64 `json:"orderId" binding:"required"`
	RechargeCost *int   `json:"rechargeCost"`
	PresentCost  *int   `json:"presentCost"`
	PayType      *int   `json:"payType" binding:"required"`
	IzPaid       *int   `json:"izPaid" binding:"required"`
}

// closeOrderCmdBody matches the downstream CloseOrderCmd: only orderId and
// payType are copied (userPin does not exist on the client DTO -> null;
// rechargeCost/presentCost/izPaid do not exist downstream -> dropped).
type closeOrderCmdBody struct {
	OrderId *int64  `json:"orderId"`
	UserPin *string `json:"userPin"`
	PayType *int    `json:"payType"`
}

// cOrderDetailQuery matches Java COrderDetailQuery extends ClientDTO
// (NOT validated on /order/detailLast and /order/detail).
type cOrderDetailQuery struct {
	clientDTONoValid
	OrderId  *int64  `json:"orderId"`
	UserPin  *string `json:"userPin"`
	Imei     *string `json:"imei"`
	CarId    *string `json:"carId"`
	IzNewApp *bool   `json:"izNewApp"` // Java default false; Optional.orElse(false) handles null
}

// orderDetailQueryBody matches the downstream OrderDetailQuery
// (izNewApp is not part of it).
type orderDetailQueryBody struct {
	OrderId *int64  `json:"orderId"`
	UserPin *string `json:"userPin"`
	Imei    *string `json:"imei"`
	CarId   *string `json:"carId"`
}

// cUpdateItineraryCmd matches Java CUpdateItineraryCmd extends ClientDTO (@Validated).
type cUpdateItineraryCmd struct {
	dto.ClientDTO
	Imei      *string `json:"imei" binding:"required"`      // @NotNull
	Itinerary *int    `json:"itinerary" binding:"required"` // @NotNull
}

// updateItineraryCmdBody matches the downstream UpdateItineraryCmd: only imei
// is copied by BeanUtils; the client "itinerary" has no same-named downstream
// field (the downstream field is "distance") so distance is sent null,
// exactly like Java does.
type updateItineraryCmdBody struct {
	Imei     *string `json:"imei"`
	Distance *int    `json:"distance"`
	CarId    *string `json:"carId"`
}

// clientOrderIdQuery matches Java ClientOrderIdQuery extends ClientDTO (@Validated).
type clientOrderIdQuery struct {
	dto.ClientDTO
	OrderId *int64 `json:"orderId" binding:"required"`
}

// ---------------------------------------------------------------------------
// BillConfig request DTOs
// ---------------------------------------------------------------------------

// userBillConfigQuery matches Java UserBillConfigQuery extends ClientDTO (@Validated).
// scene is a primitive int (default 0). It is forwarded directly: the
// downstream UserBillQuery copies serviceId/userPin and ignores the rest.
type userBillConfigQuery struct {
	dto.ClientDTO
	ServiceId *int64 `json:"serviceId" binding:"required"` // @NotNull
	UserPin   string `json:"userPin" binding:"required"`   // @NotBlank
	Scene     int    `json:"scene"`
}

// idCmdBody matches com.xyy.ebike.fence.api.dto.packing.IdCmd as built by
// SystemConfigGatewayImpl.getConfigBackCar (only id is set).
type idCmdBody struct {
	Id        *int64          `json:"id"`
	Name      *string         `json:"name"`
	AreaSize  *float64        `json:"areaSize"`
	FieldList json.RawMessage `json:"fieldList"`
}

// popupCmdBody matches com.xyy.ebike.order.api.dto.PopupCmd.
type popupCmdBody struct {
	UserPin      string `json:"userPin"`
	BillConfigId int64  `json:"billConfigId"`
}

// ---------------------------------------------------------------------------
// Invoice request DTOs
// ---------------------------------------------------------------------------

// invoiceDTO matches Java InvoiceDTO extends ClientDTO with
// @Validated(CreateGroup.class): ONLY the CreateGroup-annotated fields are
// validated (notably the ClientDTO traceId/tenantId @NotEmpty are in the
// Default group and therefore NOT enforced here) -> manual validation.
// orderIds is bound through the Java setter setOrderIds(List<Long>) and
// stored as a fastjson string "[1,2,3]".
type invoiceDTO struct {
	clientDTONoValid
	ServiceId      *int64          `json:"serviceId"`
	Type           *int            `json:"type"`
	Title          *string         `json:"title"`
	CompanyEin     *string         `json:"companyEin"`
	Email          *string         `json:"email"`
	Content        *string         `json:"content"`
	Bank           *string         `json:"bank"`
	CompanyAddress *string         `json:"companyAddress"`
	BankAccount    *string         `json:"bankAccount"`
	CompanyPhone   *string         `json:"companyPhone"`
	Phone          *string         `json:"phone"`
	OrderIds       json.RawMessage `json:"orderIds"`
}

// invoiceCmdBody matches the downstream com.xyy.ebike.order.api.dto.invoice.cmd.InvoiceCmd.
type invoiceCmdBody struct {
	Id             *int64          `json:"id"`
	ServiceId      *int64          `json:"serviceId"`
	Type           *int            `json:"type"`
	UserPin        *string         `json:"userPin"`
	Title          *string         `json:"title"`
	CompanyEin     *string         `json:"companyEin"`
	Amount         *int            `json:"amount"`
	Email          *string         `json:"email"`
	Content        *string         `json:"content"`
	State          *int            `json:"state"`
	OpManPin       *string         `json:"opManPin"`
	Bank           *string         `json:"bank"`
	CompanyAddress *string         `json:"companyAddress"`
	BankAccount    *string         `json:"bankAccount"`
	CompanyPhone   *string         `json:"companyPhone"`
	Phone          *string         `json:"phone"`
	OpResult       *int            `json:"opResult"`
	Mark           *string         `json:"mark"`
	OrderIds       *string         `json:"orderIds"`
	RemindWay      json.RawMessage `json:"remindWay"`
}

// invoiceQueryBody matches the downstream InvoiceQuery as built by
// InvoiceGatewayImpl.page: page fields copied from PageClientDTO plus
// userPin from the command context.
type invoiceQueryBody struct {
	PageNum      int             `json:"pageNum"`
	PageSize     int             `json:"pageSize"`
	Orders       []dto.OrderItem `json:"orders"`
	SearchCount  bool            `json:"searchCount"`
	LastRecordId string          `json:"lastRecordId"`
	UserPin      string          `json:"userPin"`
}

// ---------------------------------------------------------------------------
// ReturnBikeAudit request DTOs
// ---------------------------------------------------------------------------

// returnBikeAuditDTO matches Java ReturnBikeAuditDTO extends ClientDTO
// (validated on createReturnBikeAudit). It is forwarded directly: the
// downstream ReturnBikeAuditCmd shares orderId/applyType/photoUrl/userReason
// with identical types and ignores the client-only fields.
type returnBikeAuditDTO struct {
	dto.ClientDTO
	OrderId    *int64 `json:"orderId" binding:"required"`            // @NotNull
	ApplyType  *int   `json:"applyType" binding:"required"`          // @NotNull
	PhotoUrl   string `json:"photoUrl" binding:"required"`           // @NotBlank
	UserReason string `json:"userReason" binding:"omitempty,max=50"` // @Size(max = 50)
}

// returnBikeAuditNoValid is the same body without validation
// (izCapable has no @Validated).
type returnBikeAuditNoValid struct {
	clientDTONoValid
	OrderId    *int64  `json:"orderId"`
	ApplyType  *int    `json:"applyType"`
	PhotoUrl   *string `json:"photoUrl"`
	UserReason *string `json:"userReason"`
}

// cameraAuditDTO matches Java CameraAuditDTO extends ClientDTO (@Validated).
// Forwarded directly: the downstream CheckingPartCmd shares
// serviceId/orderId/carId with identical types.
type cameraAuditDTO struct {
	dto.ClientDTO
	ServiceId *int64 `json:"serviceId" binding:"required"` // @NotNull
	OrderId   *int64 `json:"orderId" binding:"required"`   // @NotNull
	CarId     string `json:"carId" binding:"required"`     // @NotBlank
}

// ---------------------------------------------------------------------------
// UserTicket request DTOs
// ---------------------------------------------------------------------------

// userTicketDTO matches Java UserTicketDTO extends ClientDTO with
// @Validated(CreateGroup.class) -> manual validation (see createUserTicket).
// photoUrl is bound through the Java setter setPhotoUrl(String[]) and joined
// with "," (length 1..3) or set to "" otherwise.
type userTicketDTO struct {
	clientDTONoValid
	OrderId    *int64          `json:"orderId"`
	Initiator  *int            `json:"initiator"`
	PhotoUrl   json.RawMessage `json:"photoUrl"`
	UserReason *string         `json:"userReason"`
}

// userTicketCmdBody matches the downstream com.xyy.ebike.order.api.dto.userticket.cmd.UserTicketCmd.
type userTicketCmdBody struct {
	Id         *int64          `json:"id"`
	OrderId    *int64          `json:"orderId"`
	Initiator  *int            `json:"initiator"`
	PhotoUrl   *string         `json:"photoUrl"`
	UserReason *string         `json:"userReason"`
	IzAccepted *bool           `json:"izAccepted"`
	OpReason   *string         `json:"opReason"`
	RefundCost *int            `json:"refundCost"`
	RemindWay  json.RawMessage `json:"remindWay"`
}

// userTicketListQueryBody matches the downstream UserTicketListQuery as built
// by UserTicketGatewayImpl.getUserTicket: page fields from PageClientDTO plus
// phone from the user profile (everything else stays null).
type userTicketListQueryBody struct {
	PageNum      int             `json:"pageNum"`
	PageSize     int             `json:"pageSize"`
	Orders       []dto.OrderItem `json:"orders"`
	SearchCount  bool            `json:"searchCount"`
	LastRecordId string          `json:"lastRecordId"`
	Phone        json.RawMessage `json:"phone"`
}

// userDetailQueryBody matches com.xyy.ebike.user.api.dto.user.UserDetailQuery.
type userDetailQueryBody struct {
	Pin       string `json:"pin"`
	ServiceId *int64 `json:"serviceId"`
}

// ---------------------------------------------------------------------------
// DeviceInfo request DTOs
// ---------------------------------------------------------------------------

// clientDeviceLocationQry matches Java ClientDeviceLocationQry extends
// ClientDTO (@Validated). Forwarded directly: the downstream DeviceLocationQry
// shares serviceId/lat/lng/radius/limit with identical types.
type clientDeviceLocationQry struct {
	dto.ClientDTO
	ServiceId *int64   `json:"serviceId" binding:"required"` // @NotNull
	Lat       *float64 `json:"lat" binding:"required"`       // @NotNull
	Lng       *float64 `json:"lng" binding:"required"`       // @NotNull
	Radius    *float64 `json:"radius"`
	Limit     *int     `json:"limit"`
}

// clientImeiDTO matches both ClientBluetoothTokenQry and ClientCarSearchVoiceCmd
// (each declares a single @NotBlank imei).
type clientImeiDTO struct {
	dto.ClientDTO
	Imei string `json:"imei" binding:"required"` // @NotBlank
}

// voiceCmdBody matches com.xyy.ebike.device.paas.api.dto.cmd.VoiceCmd as built
// by DeviceInfoGatewayImpl.carSearchVoice (imei copied, async=false, idx=9).
type voiceCmdBody struct {
	Async   bool    `json:"async"`
	CarId   *string `json:"carId"`
	Imei    string  `json:"imei"`
	Payload *string `json:"payload"`
	Idx     int     `json:"idx"`
	Volume  *int    `json:"volume"`
}

// carImeiCmdBody matches com.xyy.ebike.management.api.dto.CarImeiCmd
// (BaseCmd adds nothing). The element type is *string because Java propagates
// possible null imeis into the list.
type carImeiCmdBody struct {
	ImeiList []*string `json:"imeiList"`
}

// carInfoCmdBody matches com.xyy.ebike.management.api.dto.CarInfoCmd as built
// from a CarInfoDto carrying only carId (num is a primitive int copied as 0;
// the remaining null fields are omitted, which is equivalent for the
// downstream Jackson).
type carInfoCmdBody struct {
	CarId *string `json:"carId"`
	Num   int     `json:"num"`
}

// userBaseCmdBody matches com.xyy.ebike.account.api.cmd.UserBaseCmd
// (izUserService has the Java field default false).
type userBaseCmdBody struct {
	Pin           *string `json:"pin"`
	IzUserService bool    `json:"izUserService"`
}

// ---------------------------------------------------------------------------
// Trajectory query bodies (ebike-device-paas)
// ---------------------------------------------------------------------------

type trajectoryHistoryQryBody struct {
	OrderId   *int64 `json:"orderId"`
	StartTime *int64 `json:"startTime"`
	EndTime   *int64 `json:"endTime"`
}

type trajectoryBatchQryBody struct {
	OrderBatch []trajectoryHistoryQryBody `json:"orderBatch"`
}

type trajectoryRealTimeQryBody struct {
	Imei      *string `json:"imei"`
	StartTime *int64  `json:"startTime"`
	EndTime   *int64  `json:"endTime"`
}
