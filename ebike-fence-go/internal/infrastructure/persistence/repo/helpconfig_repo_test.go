package repo

import (
	"context"
	"testing"
	"time"

	"ebike-fence-go/internal/infrastructure/persistence/model"
	"ebike-fence-go/internal/pkg/idgen"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestApplyAuditSetsIzDelFalse(t *testing.T) {
	base := &model.ConfigBaseDO{}
	applyAudit(base, "1007", "@system", time.Now())
	if !base.IzDel.Valid || base.IzDel.Bool {
		t.Fatalf("expected iz_del=false valid, got Valid=%v Bool=%v", base.IzDel.Valid, base.IzDel.Bool)
	}
}

// Mirrors Java homenav_bug_fix_report: duplicate-name check is scoped to the edited row's service_id.
func TestEnsureHomeNavDefaultsDoesNotDuplicateExistingRows(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.TConfigHomeNav{}); err != nil {
		t.Fatal(err)
	}
	repo := NewHelpConfigRepo(db)
	ctx := context.Background()
	tenantID := "1"
	serviceID := int64(362296827341443870)
	trueVal := true
	for _, def := range homeNavDefaults {
		row := model.TConfigHomeNav{
			ServiceID: serviceID,
			Name:      def.Name,
			Icon:      def.Icon,
			JumpPage:  `{"chainType":0,"linkUrl":"` + def.LinkUrl + `"}`,
			IzOn:      &trueVal,
		}
		applyAudit(&row.ConfigBaseDO, tenantID, "pin", time.Now())
		row.ID = idgen.NextID()
		if err := db.Create(&row).Error; err != nil {
			t.Fatal(err)
		}
	}
	list, err := repo.ensureHomeNavDefaults(ctx, tenantID, "pin", serviceID)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 4 {
		t.Fatalf("expected 4 rows, got %d", len(list))
	}
	var dbCount int64
	if err := db.Model(&model.TConfigHomeNav{}).Where("service_id = ?", serviceID).Count(&dbCount).Error; err != nil {
		t.Fatal(err)
	}
	if dbCount != 4 {
		t.Fatalf("expected 4 rows in db, got %d", dbCount)
	}
}

func TestHomeNavHasSameNameScopedByEditedRowServiceID(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.TConfigHomeNav{}); err != nil {
		t.Fatal(err)
	}
	tenantID := "1003"
	serviceA := int64(338362359727786125)
	serviceB := int64(999999999999999999)
	name := "联系客服"

	rows := []model.TConfigHomeNav{
		{ConfigBaseDO: model.ConfigBaseDO{TenantID: tenantID}, ID: 1, ServiceID: serviceA, Name: name},
		{ConfigBaseDO: model.ConfigBaseDO{TenantID: tenantID}, ID: 2, ServiceID: serviceB, Name: name},
		{ConfigBaseDO: model.ConfigBaseDO{TenantID: tenantID}, ID: 3, ServiceID: serviceA, Name: "车辆报修"},
	}
	if err := db.Create(&rows).Error; err != nil {
		t.Fatal(err)
	}

	repo := NewHelpConfigRepo(db)
	ctx := context.Background()

	same, err := repo.HomeNavHasSameName(ctx, tenantID, 2, name)
	if err != nil {
		t.Fatal(err)
	}
	if same {
		t.Fatalf("expected no conflict across service areas for name %q", name)
	}

	same, err = repo.HomeNavHasSameName(ctx, tenantID, 3, name)
	if err != nil {
		t.Fatal(err)
	}
	if !same {
		t.Fatal("expected conflict within same service area when renaming to existing name")
	}
}

func TestInsertGuidePageAssignsServerID(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.TConfigGuidePage{}); err != nil {
		t.Fatal(err)
	}
	repo := NewHelpConfigRepo(db)
	row := &model.TConfigGuidePage{
		ServiceID:  364881328207300695,
		GuidePages: `[{"chainType":3}]`,
	}
	if err := repo.InsertGuidePage(context.Background(), "1007", "pin1", row); err != nil {
		t.Fatal(err)
	}
	if row.ID <= 0 {
		t.Fatalf("expected server-assigned id, got %d", row.ID)
	}
	if row.OrderWeights != 9 {
		t.Fatalf("expected orderWeights=9, got %d", row.OrderWeights)
	}
}

func TestUpdateGuidePageWritesZeroFrequencyAndVisibleRange(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.TConfigGuidePage{}); err != nil {
		t.Fatal(err)
	}
	repo := NewHelpConfigRepo(db)
	falseVal := false
	existing := &model.TConfigGuidePage{
		ConfigBaseDO:  model.ConfigBaseDO{TenantID: "1007"},
		ID:            373641672081806660,
		ServiceID:     364848372922716182,
		GuidePages:    `[]`,
		Frequency:     1,
		VisibleRange:  1,
		PageNumEsc:    1,
		AllowSuperEsc: &falseVal,
		IzOn:          &falseVal,
	}
	if err := db.Create(existing).Error; err != nil {
		t.Fatal(err)
	}

	update := &model.TConfigGuidePage{
		ID:            existing.ID,
		ServiceID:     existing.ServiceID,
		GuidePages:    `[{"chainType":3,"picUrl":"https://example.com/a.png"}]`,
		Frequency:     0,
		VisibleRange:  0,
		PageNumEsc:    1,
		AllowSuperEsc: &falseVal,
		IzOn:          &falseVal,
		ByRegister:    &falseVal,
		ByTags:        &falseVal,
	}
	if err := repo.UpdateGuidePage(context.Background(), "1007", "pin1", update, false); err != nil {
		t.Fatal(err)
	}
	var stored model.TConfigGuidePage
	if err := db.First(&stored, existing.ID).Error; err != nil {
		t.Fatal(err)
	}
	if stored.Frequency != 0 {
		t.Fatalf("expected frequency=0, got %d", stored.Frequency)
	}
	if stored.VisibleRange != 0 {
		t.Fatalf("expected visibleRange=0, got %d", stored.VisibleRange)
	}
	if stored.GuidePages == "" || stored.GuidePages == "[]" {
		t.Fatalf("expected guidePages updated, got %q", stored.GuidePages)
	}
}

func TestUpdateSpecialTipsWritesZeroFrequency(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.TConfigSpecialTips{}); err != nil {
		t.Fatal(err)
	}
	repo := NewHelpConfigRepo(db)
	falseVal := false
	existing := &model.TConfigSpecialTips{
		ConfigBaseDO: model.ConfigBaseDO{TenantID: "1007"},
		ID:           1,
		ServiceID:    100,
		Title:        "t",
		Frequency:    1,
		VisibleRange: 1,
		PopUpTime:    2,
		IzOn:         &falseVal,
	}
	if err := db.Create(existing).Error; err != nil {
		t.Fatal(err)
	}
	update := &model.TConfigSpecialTips{
		ID:           1,
		ServiceID:    100,
		Title:        "t2",
		Frequency:    0,
		VisibleRange: 0,
		PopUpTime:    0,
		ClickEvent:   0,
		IzOn:         &falseVal,
	}
	if err := repo.UpdateSpecialTips(context.Background(), "1007", "pin", update); err != nil {
		t.Fatal(err)
	}
	var stored model.TConfigSpecialTips
	if err := db.First(&stored, 1).Error; err != nil {
		t.Fatal(err)
	}
	if stored.Frequency != 0 || stored.VisibleRange != 0 || stored.PopUpTime != 0 {
		t.Fatalf("got frequency=%d visibleRange=%d popUpTime=%d", stored.Frequency, stored.VisibleRange, stored.PopUpTime)
	}
}

func TestInsertSpecialTipsAssignsServerID(t *testing.T) {
	testHelpConfigInsertAssignsID(t, "special_tips", func(repo *HelpConfigRepo) (*int64, error) {
		row := &model.TConfigSpecialTips{
			ServiceID: 364848372922716182,
			Title:     "test",
			PopUpTime: 2,
		}
		err := repo.InsertSpecialTips(context.Background(), "1007", "pin1", row)
		return &row.ID, err
	})
}

func TestApplyCustomerServiceInsertDefaults(t *testing.T) {
	row := &model.TConfigCustomerService{ServiceID: 1, Tel: "123"}
	applyCustomerServiceInsertDefaults(row)
	for _, name := range []struct {
		label string
		ptr   **bool
	}{
		{"izOnlineEntrance", &row.IzOnlineEntrance},
		{"izArtificialEntrance", &row.IzArtificialEntrance},
		{"izWorkTime", &row.IzWorkTime},
	} {
		if name.ptr == nil || *name.ptr == nil {
			t.Fatalf("%s should be set to false", name.label)
		}
		if **name.ptr {
			t.Fatalf("%s should be false", name.label)
		}
	}
}

func TestHelpConfigInsertAssignsID(t *testing.T) {
	cases := []struct {
		name string
		run  func(*HelpConfigRepo) (*int64, error)
	}{
		{"home_scroll", func(repo *HelpConfigRepo) (*int64, error) {
			row := &model.TConfigHomeScrollMsg{ServiceID: 1, Content: "c", Type: 1}
			return &row.ID, repo.InsertHomeScrollMsg(context.Background(), "1007", "pin", row)
		}},
		{"faq", func(repo *HelpConfigRepo) (*int64, error) {
			row := &model.TConfigFaq{ServiceID: 1, Title: "t"}
			return &row.ID, repo.InsertFaq(context.Background(), "1007", "pin", row)
		}},
		{"home_activity", func(repo *HelpConfigRepo) (*int64, error) {
			row := &model.TConfigHomeActivityEntrance{ServiceID: 1, LinkUrl: "u"}
			return &row.ID, repo.InsertHomeActivityEntrance(context.Background(), "1007", "pin", row)
		}},
		{"customer_service", func(repo *HelpConfigRepo) (*int64, error) {
			row := &model.TConfigCustomerService{ServiceID: 1, Tel: "123"}
			return &row.ID, repo.InsertCustomerService(context.Background(), "1007", "pin", row)
		}},
		{"home_nav", func(repo *HelpConfigRepo) (*int64, error) {
			row := &model.TConfigHomeNav{ServiceID: 1, Name: "nav"}
			return &row.ID, repo.CreateHomeNav(context.Background(), "1007", "pin", row)
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			testHelpConfigInsertAssignsID(t, tc.name, tc.run)
		})
	}
}

func testHelpConfigInsertAssignsID(t *testing.T, table string, insert func(*HelpConfigRepo) (*int64, error)) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	switch table {
	case "guide_page":
		if err := db.AutoMigrate(&model.TConfigGuidePage{}); err != nil {
			t.Fatal(err)
		}
	case "special_tips":
		if err := db.AutoMigrate(&model.TConfigSpecialTips{}); err != nil {
			t.Fatal(err)
		}
	default:
		if err := db.AutoMigrate(
			&model.TConfigHomeScrollMsg{},
			&model.TConfigFaq{},
			&model.TConfigHomeActivityEntrance{},
			&model.TConfigCustomerService{},
			&model.TConfigHomeNav{},
		); err != nil {
			t.Fatal(err)
		}
	}
	repo := NewHelpConfigRepo(db)
	id, err := insert(repo)
	if err != nil {
		t.Fatal(err)
	}
	if id == nil || *id <= 0 {
		t.Fatalf("expected server-assigned id, got %v", id)
	}
}
