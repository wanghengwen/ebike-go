package fenceadmin

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"ebike-fence-go/internal/api/dto"
	"ebike-fence-go/internal/infrastructure/mq"
	"ebike-fence-go/internal/infrastructure/persistence"
	"ebike-fence-go/internal/infrastructure/persistence/model"
	"ebike-fence-go/internal/infrastructure/persistence/repo"
	"ebike-fence-go/internal/infrastructure/rpc"
	"ebike-fence-go/internal/pkg/shadow"
	"ebike-fence-go/internal/pkg/timefmt"
)

type SiteApplicationAdmin struct{}

func NewSiteApplicationAdmin() *SiteApplicationAdmin { return &SiteApplicationAdmin{} }

func (s *SiteApplicationAdmin) Create(ctx context.Context, tenantID, pin string, cmd dto.SiteApplicationCmd) (int64, error) {
	if persistence.SiteApplication == nil {
		return 0, fmt.Errorf("mysql not initialized")
	}
	row := &model.TSiteApplication{
		ServiceID:       cmd.ServiceId,
		Location:        sql.NullString{String: cmd.Location, Valid: cmd.Location != ""},
		Lat:             sql.NullFloat64{Float64: cmd.Lat, Valid: true},
		Lng:             sql.NullFloat64{Float64: cmd.Lng, Valid: true},
		PhotoURL:        sql.NullString{String: cmd.PhotoUrl, Valid: cmd.PhotoUrl != ""},
		ApplicantRemark: sql.NullString{String: cmd.ApplicantRemark, Valid: cmd.ApplicantRemark != ""},
		State:           sql.NullInt32{Int32: 0, Valid: true},
	}
	return persistence.SiteApplication.Insert(ctx, tenantID, pin, row)
}

func (s *SiteApplicationAdmin) Deal(ctx context.Context, tenantID, pin string, cmd dto.SiteApplicationCmd) error {
	if persistence.SiteApplication == nil || cmd.Id == nil {
		return fmt.Errorf("mysql not initialized")
	}
	row, err := persistence.SiteApplication.GetByID(ctx, *cmd.Id)
	if err != nil || row == nil {
		return newBizError("NOT_FOUND", "site application not found")
	}
	if cmd.State != nil {
		row.State = sql.NullInt32{Int32: int32(*cmd.State), Valid: true}
	}
	row.OpManPin = sql.NullString{String: pin, Valid: pin != ""}
	row.OpManRemark = sql.NullString{String: cmd.OpManRemark, Valid: cmd.OpManRemark != ""}
	row.UpdatedPin = sql.NullString{String: pin, Valid: pin != ""}
	if err := persistence.SiteApplication.Update(ctx, row); err != nil {
		return err
	}
	s.dealSideEffects(ctx, tenantID, cmd, row)
	return nil
}

// dealSideEffects mirrors Java SiteApplicationServiceImpl.dealSiteApplication after the
// status update: notify the applicant (sendMsg) when remindWay is provided, and publish
// the "申请通过" user-action event (applyPass) when the application is approved.
func (s *SiteApplicationAdmin) dealSideEffects(ctx context.Context, tenantID string, cmd dto.SiteApplicationCmd, row *model.TSiteApplication) {
	if shadow.IsShadowTest(ctx) {
		return
	}
	if !row.UserPin.Valid || row.UserPin.String == "" {
		return
	}
	cc := dto.EnsureCommandContext(cmd.CommandContext, tenantID)
	user, err := rpc.NewUserRPC().GetUserDetailByPin(ctx, cc, row.UserPin.String)
	if err != nil || user == nil {
		return
	}

	if len(cmd.RemindWay) > 0 {
		applyResult := "未通过"
		if cmd.State != nil && *cmd.State == 1 {
			applyResult = "通过"
		}
		applyTime := ""
		if row.CreatedAt.Valid {
			applyTime = timefmt.FormatJavaLocalSpace(row.CreatedAt.Time)
		}
		recv := rpc.MsgReceivedUser{
			ReceivePin: user.Pin,
			Phone:      user.Phone,
			TemplateParams: map[string]string{
				"username":    user.AuthName,
				"applyTime":   applyTime,
				"applyResult": applyResult,
				"mark":        cmd.OpManRemark,
			},
		}
		if containsInt(cmd.RemindWay, 3) {
			appType := 1
			recv.AppType = &appType
		}
		_ = rpc.NewManagementRPC().SendMsg(ctx, rpc.MsgSendCmd{
			CommandContext: cc,
			Type:           14,
			RemindTypes:    cmd.RemindWay,
			ReceivedUser:   []rpc.MsgReceivedUser{recv},
		})
	}

	if cmd.State != nil && *cmd.State == 1 {
		mq.PublishUserAction(2, user.Pin, cc)
	}
}

func containsInt(s []int, v int) bool {
	for _, n := range s {
		if n == v {
			return true
		}
	}
	return false
}

func (s *SiteApplicationAdmin) Page(ctx context.Context, tenantID string, q dto.ApplicationQuery) (dto.JavaPageDTO[dto.SiteApplicationCO], error) {
	if persistence.SiteApplication == nil {
		return dto.JavaPageDTO[dto.SiteApplicationCO]{}, fmt.Errorf("mysql not initialized")
	}
	pq := repo.SiteApplicationPageQuery{
		ServiceID: q.ServiceId,
		PageNum:   q.PageNum,
		PageSize:  q.PageSize,
	}
	if q.State != nil {
		pq.State = q.State
	}
	if q.CreatedTimeStart != "" {
		if t, err := time.Parse(time.RFC3339, q.CreatedTimeStart); err == nil {
			pq.CreatedTimeStart = &t
		}
	}
	if q.CreatedTimeEnd != "" {
		if t, err := time.Parse(time.RFC3339, q.CreatedTimeEnd); err == nil {
			pq.CreatedTimeEnd = &t
		}
	}
	rows, total, err := persistence.SiteApplication.Page(ctx, tenantID, pq)
	if err != nil {
		return dto.JavaPageDTO[dto.SiteApplicationCO]{}, err
	}

	pins := make([]string, 0, len(rows))
	seen := map[string]struct{}{}
	for _, row := range rows {
		if row.UserPin.Valid && row.UserPin.String != "" {
			if _, ok := seen[row.UserPin.String]; !ok {
				pins = append(pins, row.UserPin.String)
				seen[row.UserPin.String] = struct{}{}
			}
		}
	}
	userMap := map[string]rpc.AppUser{}
	if len(pins) > 0 {
		userRPC := rpc.NewUserRPC()
		if m, err := userRPC.BatchGetByPins(ctx, q.CommandContext, pins); err == nil {
			userMap = m
		}
	}

	out := make([]dto.SiteApplicationCO, 0, len(rows))
	for _, row := range rows {
		co := siteAppToCO(row)
		if row.UserPin.Valid {
			if u, ok := userMap[row.UserPin.String]; ok {
				co.ApplicantPhone = u.Phone
			}
		}
		out = append(out, co)
	}
	return dto.JavaPageDTO[dto.SiteApplicationCO]{
		Count:       total,
		PageNum:     q.PageNum,
		PageSize:    q.PageSize,
		Orders:      nil,
		SearchCount: q.SearchCount,
		List:        out,
	}, nil
}

func siteAppToCO(row model.TSiteApplication) dto.SiteApplicationCO {
	co := dto.SiteApplicationCO{Id: row.ID, ServiceId: row.ServiceID, OpManRemark: ""}
	if row.Location.Valid {
		co.Location = row.Location.String
	}
	if row.Lat.Valid {
		co.Lat = row.Lat.Float64
	}
	if row.Lng.Valid {
		co.Lng = row.Lng.Float64
	}
	if row.UserPin.Valid {
		co.UserPin = row.UserPin.String
	}
	if row.PhotoURL.Valid {
		co.PhotoUrl = row.PhotoURL.String
	}
	if row.ApplicantRemark.Valid {
		co.ApplicantRemark = row.ApplicantRemark.String
	} else {
		co.ApplicantRemark = ""
	}
	if row.State.Valid {
		co.State = int(row.State.Int32)
	}
	if row.OpManPin.Valid {
		co.OpManPin = row.OpManPin.String
	}
	if row.OpManRemark.Valid {
		co.OpManRemark = row.OpManRemark.String
	}
	if row.CreatedAt.Valid {
		co.CreatedAt = timefmt.FormatJavaLocal(row.CreatedAt.Time)
	}
	return co
}
