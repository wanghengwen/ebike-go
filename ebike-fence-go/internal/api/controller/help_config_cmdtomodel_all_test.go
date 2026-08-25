package controller

import (
	"testing"
	"time"

	"ebike-fence-go/internal/api/dto"
	"ebike-fence-go/internal/infrastructure/persistence/model"
)

func TestCmdToModelAllHelpConfigTypes(t *testing.T) {
	sid := int64(364881328207300695)
	id := int64(123)
	typ := 1
	on := false
	now := time.Now()

	cases := []struct {
		name string
		run  func() error
	}{
		{"HomeScroll", func() error {
			_, err := cmdToModel[model.TConfigHomeScrollMsg](dto.HomeScrollerMsgCmd{ServiceId: &sid, Title: "t", Content: "c", Type: &typ, IzOn: &on})
			return err
		}},
		{"Faq", func() error {
			_, err := cmdToModel[model.TConfigFaq](dto.FaqCmd{ServiceId: &sid, Title: "faq", IzOn: &on})
			return err
		}},
		{"HomeActivity", func() error {
			_, err := cmdToModel[model.TConfigHomeActivityEntrance](dto.HomeActivityEntranceCmd{ServiceId: &sid, LinkUrl: "u", IzOn: &on, StartTime: &now})
			return err
		}},
		{"SpecialTips", func() error {
			_, err := cmdToModel[model.TConfigSpecialTips](dto.SpecialTipsCmd{ServiceId: &sid, Title: "s", JumpPage: `{"chainType":1}`, IzOn: &on})
			return err
		}},
		{"CustomerService", func() error {
			_, err := cmdToModel[model.TConfigCustomerService](dto.CustomerServiceCmd{ServiceId: &sid, StartTime: "09:00:00", EndTime: "18:00:00", Tel: "123"})
			return err
		}},
		{"HomeNav", func() error {
			_, err := cmdToModel[model.TConfigHomeNav](dto.HomeNavCmd{ServiceId: &sid, Name: "nav", JumpPage: "{}", IzOn: &on, Ids: []int64{1, 2}})
			return err
		}},
		{"GuidePageWithGuidePagesArray", func() error {
			row, err := cmdToModel[model.TConfigGuidePage](dto.GuidePageConfigCmd{
				ServiceId:  &sid,
				Id:         &id,
				GuidePages: []dto.JumpPage{{ChainType: 3, PicUrl: "p"}},
			})
			if err != nil {
				return err
			}
			if row == nil {
				t.Fatal("nil row")
			}
			if row.GuidePages != "" {
				t.Fatalf("cmdToModel should not populate guidePages from array; got %q", row.GuidePages)
			}
			return nil
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.run(); err != nil {
				t.Fatalf("cmdToModel failed: %v", err)
			}
		})
	}
}
