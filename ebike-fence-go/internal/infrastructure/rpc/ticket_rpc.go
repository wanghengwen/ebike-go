package rpc

import (
	"context"
	"fmt"
	"log"

	"ebike-fence-go/internal/api/dto"
	pkg_rpc "ebike-fence-go/internal/pkg/rpc"
	"ebike-fence-go/internal/pkg/shadow"
)

type TicketRPC struct {
}

func NewTicketRPC() *TicketRPC {
	return &TicketRPC{}
}

type HelmetTicket struct {
	TicketNo string `json:"ticket_no"`
}

type carIdRepairQry struct {
	CommandContext *dto.CommandContext `json:"commandContext,omitempty"`
	CarId          string              `json:"carId"`
	RepairId       int64               `json:"repairId"`
}

type alarmTicketQry struct {
	CommandContext *dto.CommandContext `json:"commandContext,omitempty"`
	Imei           string              `json:"imei"`
	IzFinish       bool                `json:"izFinish"`
	ServiceId      int64               `json:"serviceId"`
	Types          []int               `json:"types"`
}

type repairCmd struct {
	CommandContext *dto.CommandContext `json:"commandContext,omitempty"`
	CarId          string              `json:"carId"`
	Types          []int64             `json:"types"`
	RepairSource   int                 `json:"repairSource"`
}

func (r *TicketRPC) QueryTicket(ctx context.Context, cmdCtx *dto.CommandContext, carId string, repairId int64) (bool, error) {
	if cmdCtx == nil {
		cmdCtx = &dto.CommandContext{}
	}
	var res Result[bool]
	body := carIdRepairQry{CommandContext: cmdCtx, CarId: carId, RepairId: repairId}
	err := pkg_rpc.PostToService(ctx, "ebike-ticket", "/ticket/queryTicket", body, &res)
	if err != nil {
		return false, err
	}

	if !res.OK() {
		return false, fmt.Errorf("ticket api returned error: code=%d msg=%s", res.Code, res.Msg)
	}

	return res.Data, nil
}

func (r *TicketRPC) GetHelmetTicket(ctx context.Context, cmdCtx *dto.CommandContext, imei string, isHistory bool, serviceId int64, types []int) ([]HelmetTicket, error) {
	if cmdCtx == nil {
		cmdCtx = &dto.CommandContext{}
	}
	var res Result[[]HelmetTicket]
	body := alarmTicketQry{
		CommandContext: cmdCtx,
		Imei:           imei,
		IzFinish:       isHistory,
		ServiceId:      serviceId,
		Types:          types,
	}
	err := pkg_rpc.PostToService(ctx, "ebike-ticket", "/ticket/getHelmetTicket", body, &res)
	if err != nil {
		return nil, err
	}

	if !res.OK() {
		return nil, fmt.Errorf("ticket api returned error: code=%d msg=%s", res.Code, res.Msg)
	}

	return res.Data, nil
}

// CreateTicket mirrors Java TicketApiRpcImpl.createTicket (repairSource=4).
func (r *TicketRPC) CreateTicket(ctx context.Context, cmdCtx *dto.CommandContext, carId string, repairType int64) error {
	if shadow.IsShadowTest(ctx) {
		log.Printf("[SHADOW TEST] Skipped CreateTicket: carId=%s, repairType=%d", carId, repairType)
		return nil
	}
	if cmdCtx == nil {
		cmdCtx = &dto.CommandContext{}
	}

	cmd := repairCmd{
		CommandContext: cmdCtx,
		CarId:          carId,
		Types:          []int64{repairType},
		RepairSource:   4,
	}
	var res Result[bool]
	if err := pkg_rpc.PostToService(ctx, "ebike-ticket", "/repair/add", cmd, &res); err != nil {
		return err
	}
	return nil
}
