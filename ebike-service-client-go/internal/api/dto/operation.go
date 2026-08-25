package dto

// UserRepairDTO represents a user repair submission request.
// Matches Java: com.xyy.ebike.service.client.client.dto.operation.UserRepairDTO extends ClientDTO
// (carId/imei are @NotBlank, serviceId is @NotNull in Java)
type UserRepairDTO struct {
	ClientDTO
	CarId       string   `json:"carId" binding:"required"`
	Imei        string   `json:"imei" binding:"required"`
	ServiceId   *int64   `json:"serviceId" binding:"required"` // Java Long (boxed) -> Go *int64
	Address     string   `json:"address,omitempty"`
	RepairPart  []int64  `json:"repairPart,omitempty"`  // Java List<Long>
	ReportDesc  string   `json:"reportDesc,omitempty"`
	ReportPhoto []string `json:"reportPhoto,omitempty"` // Java List<String>
}

// UserRepairCmd is the downstream Command sent to ebike-operation.
// Matches Java: com.xyy.ebike.operation.api.cmd.UserRepairCmd extends LocationBaseCmd extends Command
// Java's CommandContextHolder.genParam(UserRepairCmd.class, dto) copies DTO fields + injects commandContext.
type UserRepairCmd struct {
	CommandContext *CommandContext `json:"commandContext,omitempty"`
	Longitude      *float64        `json:"longitude,omitempty"` // from LocationBaseCmd
	Latitude       *float64        `json:"latitude,omitempty"`  // from LocationBaseCmd
	CarId          string          `json:"carId"`
	Imei           string          `json:"imei"`
	ServiceId      *int64          `json:"serviceId,omitempty"`
	Address        string          `json:"address,omitempty"`
	RepairPart     []int64         `json:"repairPart,omitempty"`
	ReportDesc     string          `json:"reportDesc,omitempty"`
	ReportPhoto    []string        `json:"reportPhoto,omitempty"`
}

// BasePageCmd is the downstream Command for pagination queries.
// Matches Java: com.xyy.ebike.operation.api.cmd.BasePageCmd extends PageQuery extends Query
// PageQuery declares pageNum=1, pageSize=10, searchCount=true defaults and a
// lastRecordId field — all of which BeanUtils.copyProperties carries over from
// PageClientDTO in the Java gateway.
type BasePageCmd struct {
	CommandContext *CommandContext `json:"commandContext,omitempty"`
	PageNum        int             `json:"pageNum"`
	PageSize       int             `json:"pageSize"`
	Orders         []OrderItem     `json:"orders,omitempty"`
	SearchCount    bool            `json:"searchCount"`
	LastRecordId   string          `json:"lastRecordId,omitempty"`
}

// ConvertToRepairCmd converts the client-facing UserRepairDTO into the downstream UserRepairCmd.
// The CommandContext comes from the middleware (parsed from "authorities" header)
// completed with ClientDTO fields, mirroring Java's CommandContextHolder/CommandContextAspect.
func ConvertToRepairCmd(dto *UserRepairDTO, ctx *CommandContext) *UserRepairCmd {
	return &UserRepairCmd{
		CommandContext: ctx,
		Longitude:      dto.Longitude,
		Latitude:       dto.Latitude,
		CarId:          dto.CarId,
		Imei:           dto.Imei,
		ServiceId:      dto.ServiceId,
		Address:        dto.Address,
		RepairPart:     dto.RepairPart,
		ReportDesc:     dto.ReportDesc,
		ReportPhoto:    dto.ReportPhoto,
	}
}

// ConvertToBasePageCmd converts the client-facing PageClientDTO into the downstream BasePageCmd.
func ConvertToBasePageCmd(dto *PageClientDTO, ctx *CommandContext) *BasePageCmd {
	// Java PageClientDTO declares `private boolean searchCount = true;`
	searchCount := true
	if dto.SearchCount != nil {
		searchCount = *dto.SearchCount
	}
	return &BasePageCmd{
		CommandContext: ctx,
		PageNum:        dto.PageNum,
		PageSize:       dto.PageSize,
		Orders:         dto.Orders,
		SearchCount:    searchCount,
		LastRecordId:   dto.LastRecordId,
	}
}
