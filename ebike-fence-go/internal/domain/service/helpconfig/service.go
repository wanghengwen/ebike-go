package helpconfig

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"ebike-fence-go/internal/api/dto"
	"ebike-fence-go/internal/infrastructure/persistence/model"
	"ebike-fence-go/internal/infrastructure/persistence/repo"
	"ebike-fence-go/internal/infrastructure/rpc"
	"ebike-fence-go/internal/pkg/mysql"
	pkg_rpc "ebike-fence-go/internal/pkg/rpc"
)

const (
	codeConfigInsertFail = "13008"
	codeConfigUpdateFail = "13009"
	codeConfigDeleteFail = "13100"
	codeRepeatAdd        = "13018"
	codeFenceCustomExist = "13043"
	codeDisableOpen      = "13045"
)

// BizError mirrors Java BizException for help-config handlers.
type BizError struct {
	Code string
	Msg  string
}

func (e *BizError) Error() string { return e.Msg }

func bizErr(code, msg string) error {
	return &BizError{Code: code, Msg: msg}
}

func boolVal(v *bool) bool { return v != nil && *v }

type Service struct {
	repo *repo.HelpConfigRepo
}

func NewService() *Service {
	return &Service{repo: repo.NewHelpConfigRepo(mysql.DB)}
}

// HomeScrollerMsg

func (s *Service) ListHomeScrollerMsg(ctx context.Context, tenantID string, serviceID int64) ([]model.TConfigHomeScrollMsg, error) {
	return s.repo.ListHomeScrollMsg(ctx, tenantID, serviceID)
}

func (s *Service) FirstHomeScrollerMsg(ctx context.Context, tenantID string, serviceID int64) (*model.TConfigHomeScrollMsg, error) {
	list, err := s.ListHomeScrollerMsg(ctx, tenantID, serviceID)
	if err != nil || len(list) == 0 {
		return nil, err
	}
	return &list[0], nil
}

func (s *Service) GetHomeScrollerMsgByID(ctx context.Context, id int64) (*model.TConfigHomeScrollMsg, error) {
	return s.repo.GetHomeScrollMsgByID(ctx, id)
}

func (s *Service) AddHomeScrollerMsg(ctx context.Context, tenantID, pin string, row *model.TConfigHomeScrollMsg) error {
	return s.repo.InsertHomeScrollMsg(ctx, tenantID, pin, row)
}

func (s *Service) EditHomeScrollerMsg(ctx context.Context, tenantID, pin string, row *model.TConfigHomeScrollMsg) error {
	return s.repo.UpdateHomeScrollMsg(ctx, tenantID, pin, row)
}

func (s *Service) DelHomeScrollerMsg(ctx context.Context, tenantID string, serviceID, id int64) error {
	return s.repo.DeleteHomeScrollMsg(ctx, tenantID, serviceID, id)
}

// Faq

func (s *Service) ListFaq(ctx context.Context, tenantID string, serviceID int64) ([]model.TConfigFaq, error) {
	return s.repo.ListFaq(ctx, tenantID, serviceID)
}

func (s *Service) GetFaqByID(ctx context.Context, id int64) (*model.TConfigFaq, error) {
	return s.repo.GetFaqByID(ctx, id)
}

func (s *Service) AddFaq(ctx context.Context, tenantID, pin string, row *model.TConfigFaq) error {
	return s.repo.InsertFaq(ctx, tenantID, pin, row)
}

func (s *Service) EditFaq(ctx context.Context, tenantID, pin string, row *model.TConfigFaq) error {
	return s.repo.UpdateFaq(ctx, tenantID, pin, row)
}

func (s *Service) DelFaq(ctx context.Context, tenantID string, serviceID, id int64) error {
	return s.repo.DeleteFaq(ctx, tenantID, serviceID, id)
}

// GuidePage

func (s *Service) ListGuidePage(ctx context.Context, tenantID string, serviceID int64) ([]model.TConfigGuidePage, error) {
	return s.repo.ListGuidePage(ctx, tenantID, serviceID)
}

func (s *Service) GetGuidePageIzOn(ctx context.Context, tenantID, pin string, cmdCtx *dto.CommandContext, serviceID int64) (*model.TConfigGuidePage, error) {
	row, err := s.repo.GetGuidePageIzOnCached(ctx, tenantID, serviceID)
	if err != nil || row == nil {
		return nil, err
	}
	isNewUser, _ := s.isNewUser(ctx, pin, cmdCtx)
	if !shouldShow(row.ByRegister, row.ByTags, row.VisibleRange, isNewUser, "", row.TagIds) {
		return nil, nil
	}
	return row, nil
}

func (s *Service) AddGuidePage(ctx context.Context, tenantID, pin string, row *model.TConfigGuidePage) error {
	return s.repo.InsertGuidePage(ctx, tenantID, pin, row)
}

func (s *Service) EditGuidePage(ctx context.Context, tenantID, pin string, row *model.TConfigGuidePage) error {
	disable := boolVal(row.IzOn)
	return s.repo.UpdateGuidePage(ctx, tenantID, pin, row, disable)
}

func (s *Service) DelGuidePage(ctx context.Context, tenantID string, serviceID, id int64) error {
	return s.repo.DeleteGuidePage(ctx, tenantID, serviceID, id)
}

func (s *Service) SortGuidePage(ctx context.Context, tenantID, pin string, serviceID int64, ids []int64) error {
	return s.repo.SortGuidePage(ctx, tenantID, pin, serviceID, ids)
}

func (s *Service) isNewUser(ctx context.Context, pin string, cmdCtx *dto.CommandContext) (bool, error) {
	if pin == "" || cmdCtx == nil {
		return false, nil
	}
	body := map[string]interface{}{
		"commandContext": cmdCtx,
		"userPin":        pin,
	}
	var res rpc.Result[bool]
	if err := pkg_rpc.PostToService(ctx, "ebike-order", "/order/hasOrder", body, &res); err != nil {
		return false, err
	}
	if !res.OK() {
		return false, fmt.Errorf("order api: code=%d msg=%s", res.Code, res.Msg)
	}
	return !res.Data, nil
}

func shouldShow(byRegister, byTags *bool, visible int, isNewUser bool, userTags, activeTags string) bool {
	registerShow := false
	tagsShow := false
	if boolVal(byRegister) {
		registerShow = visible == 0 || (visible == 1 && isNewUser)
	}
	if boolVal(byTags) {
		tagsShow = tagsIntersect(activeTags, userTags)
	}
	return registerShow || tagsShow
}

func tagsIntersect(active, user string) bool {
	if active == "" || user == "" {
		return false
	}
	a := strings.Split(active, ",")
	u := strings.Split(user, ",")
	for _, x := range a {
		for _, y := range u {
			if x != "" && x == y {
				return true
			}
		}
	}
	return false
}

// HomeActivity

func (s *Service) ListHomeActivity(ctx context.Context, tenantID string, serviceID int64) ([]model.TConfigHomeActivityEntrance, error) {
	list, err := s.repo.ListHomeActivityEntrance(ctx, tenantID, serviceID)
	if err != nil {
		return nil, err
	}
	if list == nil {
		return []model.TConfigHomeActivityEntrance{}, nil
	}
	return list, nil
}

func (s *Service) GetHomeActivityByID(ctx context.Context, id int64) (*model.TConfigHomeActivityEntrance, error) {
	return s.repo.GetHomeActivityEntranceByID(ctx, id)
}

func (s *Service) AddHomeActivity(ctx context.Context, tenantID, pin string, row *model.TConfigHomeActivityEntrance) error {
	list, err := s.repo.ListHomeActivityEntrance(ctx, tenantID, row.ServiceID)
	if err != nil {
		return err
	}
	for _, n := range list {
		if n.Position == row.Position {
			return bizErr(codeRepeatAdd, "同一位置不能添加多个")
		}
	}
	if len(list) > 2 {
		return bizErr(codeRepeatAdd, "同一位置不能添加多个")
	}
	return s.repo.InsertHomeActivityEntrance(ctx, tenantID, pin, row)
}

func (s *Service) EditHomeActivity(ctx context.Context, tenantID, pin string, row *model.TConfigHomeActivityEntrance) error {
	list, err := s.repo.ListHomeActivityEntrance(ctx, tenantID, row.ServiceID)
	if err != nil {
		return err
	}
	for _, n := range list {
		if n.ID != row.ID && n.Position == row.Position {
			return bizErr(codeRepeatAdd, "同一位置不能添加多个")
		}
	}
	var old *model.TConfigHomeActivityEntrance
	for _, n := range list {
		if n.ID == row.ID {
			old = &n
			break
		}
	}
	if old == nil {
		return bizErr(codeConfigUpdateFail, "配置修改失败")
	}
	if boolVal(old.IzOn) && boolVal(row.IzOn) {
		return bizErr(codeConfigDeleteFail, "请先禁用配置再操作")
	}
	if boolVal(old.IzOn) && (!boolVal(row.IzOn) || boolVal(row.Unlimited)) {
		row.StartTime = nil
		row.EndTime = nil
	}
	if !boolVal(row.IzOn) || boolVal(row.Unlimited) {
		row.StartTime = nil
		row.EndTime = nil
	}
	return s.repo.UpdateHomeActivityEntrance(ctx, tenantID, pin, row)
}

func (s *Service) DelHomeActivity(ctx context.Context, tenantID string, serviceID, id int64) error {
	row, err := s.repo.GetHomeActivityEntranceByID(ctx, id)
	if err != nil {
		return err
	}
	if row != nil && boolVal(row.IzOn) {
		return bizErr(codeConfigDeleteFail, "请先禁用配置再操作")
	}
	return s.repo.DeleteHomeActivityEntrance(ctx, tenantID, serviceID, id)
}

// SpecialTips

func (s *Service) ListSpecialTips(ctx context.Context, tenantID string, serviceID int64) ([]model.TConfigSpecialTips, error) {
	return s.repo.ListSpecialTips(ctx, tenantID, serviceID)
}

func (s *Service) GetSpecialTipsByID(ctx context.Context, id int64) (*model.TConfigSpecialTips, error) {
	return s.repo.GetSpecialTipsByID(ctx, id)
}

func (s *Service) AddSpecialTips(ctx context.Context, tenantID, pin string, row *model.TConfigSpecialTips) error {
	guideList, err := s.repo.ListGuidePage(ctx, tenantID, row.ServiceID)
	if err != nil {
		return err
	}
	// Java SpecialTipsServiceImpl compares List == null; MyBatis selectList never returns nil,
	// so this guard is effectively disabled in production Java.
	if guideList == nil && (row.PopUpTime == 0 || row.PopUpTime == 1) {
		return bizErr(codeConfigInsertFail, "配置新增失败")
	}
	return s.repo.InsertSpecialTips(ctx, tenantID, pin, row)
}

func (s *Service) EditSpecialTips(ctx context.Context, tenantID, pin string, row *model.TConfigSpecialTips) error {
	old, err := s.repo.GetSpecialTipsByID(ctx, row.ID)
	if err != nil {
		return err
	}
	if old != nil && boolVal(old.IzOn) {
		return bizErr(codeConfigUpdateFail, "配置修改失败")
	}
	return s.repo.UpdateSpecialTips(ctx, tenantID, pin, row)
}

func (s *Service) DelSpecialTips(ctx context.Context, tenantID string, serviceID, id int64) error {
	old, err := s.repo.GetSpecialTipsByID(ctx, id)
	if err != nil {
		return err
	}
	if old != nil && boolVal(old.IzOn) {
		return bizErr(codeConfigDeleteFail, "请先禁用配置再操作")
	}
	return s.repo.DeleteSpecialTips(ctx, tenantID, serviceID, id)
}

func (s *Service) OnOffSpecialTips(ctx context.Context, tenantID, pin string, row *model.TConfigSpecialTips) error {
	return s.repo.OnOffSpecialTips(ctx, tenantID, pin, row)
}

// CustomerService

func (s *Service) GetCustomerService(ctx context.Context, tenantID, pin string, cmdCtx *dto.CommandContext, serviceID int64) (*dto.CustomerServiceCO, error) {
	row, err := s.repo.GetCustomerService(ctx, tenantID, serviceID)
	if err != nil {
		return nil, err
	}
	if row == nil {
		def := &model.TConfigCustomerService{ServiceID: serviceID}
		if err := s.repo.InsertCustomerService(ctx, tenantID, pin, def); err != nil {
			return nil, err
		}
		row, err = s.repo.GetCustomerService(ctx, tenantID, serviceID)
		if err != nil {
			return nil, err
		}
	}
	co := ToCustomerServiceCO(row, serviceID)
	if co.UpdatedPin != "" && cmdCtx != nil {
		name, err := s.fetchUserName(ctx, co.UpdatedPin, cmdCtx)
		if err != nil || name == "" {
			name = "--"
		}
		co.UpdatedName = &name
	}
	return &co, nil
}

func (s *Service) InsertCustomerService(ctx context.Context, tenantID, pin string, row *model.TConfigCustomerService) error {
	return s.repo.InsertCustomerService(ctx, tenantID, pin, row)
}

func (s *Service) InsertCustomerServiceBatch(ctx context.Context, tenantID, pin string, row *model.TConfigCustomerService, copyIDs []int64) error {
	for _, sid := range copyIDs {
		cp := *row
		cp.ID = 0
		cp.ServiceID = sid
		if err := s.repo.InsertCustomerService(ctx, tenantID, pin, &cp); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) fetchUserName(ctx context.Context, pin string, cmdCtx *dto.CommandContext) (string, error) {
	body := map[string]interface{}{
		"commandContext": cmdCtx,
		"pin":            pin,
	}
	var res rpc.Result[*struct {
		Name string `json:"name"`
	}]
	if err := pkg_rpc.PostToService(ctx, "ebike-management", "/user/getUserByPin", body, &res); err != nil {
		return "", err
	}
	if !res.OK() || res.Data == nil {
		return "", fmt.Errorf("management api: code=%d", res.Code)
	}
	return res.Data.Name, nil
}

// HomeNav

func (s *Service) ListHomeNav(ctx context.Context, tenantID, pin string, serviceID int64) ([]model.TConfigHomeNav, error) {
	return s.repo.ListHomeNav(ctx, tenantID, pin, serviceID)
}

func (s *Service) AddHomeNav(ctx context.Context, tenantID, pin string, row *model.TConfigHomeNav) error {
	list, err := s.repo.ListHomeNav(ctx, tenantID, pin, row.ServiceID)
	if err != nil {
		return err
	}
	for _, n := range list {
		if n.Name == row.Name {
			return bizErr(codeFenceCustomExist, "名称重复")
		}
	}
	if len(list) >= 5 {
		return bizErr(codeDisableOpen, "该模块配置不可少许4项，最多5项")
	}
	if len(list) > 0 {
		sort.Slice(list, func(i, j int) bool { return list[i].UpdatedAt.After(list[j].UpdatedAt) })
		row.IzOn = list[0].IzOn
	}
	return s.repo.CreateHomeNav(ctx, tenantID, pin, row)
}

func (s *Service) IsOpenHomeNav(ctx context.Context, tenantID, pin string, serviceID int64, izOn bool) error {
	list, err := s.repo.ListHomeNav(ctx, tenantID, pin, serviceID)
	if err != nil {
		return err
	}
	ids := make([]int64, 0, len(list))
	for _, n := range list {
		ids = append(ids, n.ID)
	}
	if err := s.repo.HomeNavIsOpen(ctx, tenantID, pin, serviceID, ids, izOn); err != nil {
		if err.Error() == "DISABLE_OPEN" {
			return bizErr(codeDisableOpen, "该模块配置不可少许4项，最多5项")
		}
		return err
	}
	return nil
}

func (s *Service) EditHomeNav(ctx context.Context, tenantID, pin string, row *model.TConfigHomeNav, sortIDs []int64) error {
	if len(sortIDs) > 0 {
		return s.repo.SortHomeNav(ctx, tenantID, pin, row.ServiceID, sortIDs)
	}
	if row.ID > 0 && row.Name != "" {
		// Mirrors Java HomeNavQueryImpl.updateHaveSameName: scope duplicate check to the
		// service area of the row being edited (lookup by id), not the whole table.
		same, err := s.repo.HomeNavHasSameName(ctx, tenantID, row.ID, row.Name)
		if err != nil {
			return err
		}
		if same {
			return bizErr(codeFenceCustomExist, "名称重复")
		}
	}
	return s.repo.UpdateHomeNav(ctx, tenantID, pin, row)
}

func (s *Service) DelHomeNav(ctx context.Context, tenantID string, serviceID, id int64) error {
	if err := s.repo.DeleteHomeNav(ctx, tenantID, serviceID, id); err != nil {
		if err.Error() == "CONFIG_DELETE_FAIL" {
			return bizErr(codeConfigDeleteFail, "请先禁用配置再操作")
		}
		return err
	}
	return nil
}

func serviceIDRequired(id int64) error {
	if id == 0 {
		return errors.New("serviceId must not be null")
	}
	return nil
}
