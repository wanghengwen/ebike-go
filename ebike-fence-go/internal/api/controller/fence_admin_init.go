package controller

import (
	"net/http"
	"sync"

	"ebike-fence-go/internal/api/dto"
	"ebike-fence-go/internal/domain/service"
	"ebike-fence-go/internal/domain/service/fenceadmin"
	"ebike-fence-go/internal/middleware"
	"ebike-fence-go/internal/pkg/web"

	"github.com/gin-gonic/gin"
)

var (
	fenceAdminOnce sync.Once
	svcAreaAdmin   *fenceadmin.ServiceAreaAdmin
	parkingAdmin   *fenceadmin.ParkingAdmin
	noParkingAdmin *fenceadmin.NoParkingAdmin
	banRidingAdmin *fenceadmin.BanRidingAdmin
	maintainAdmin  *fenceadmin.MaintainAreaAdmin
	customAdmin    *fenceadmin.FenceCustomAdmin
	rfidAdmin      *fenceadmin.FenceRfidAdmin
	businessAdmin  *fenceadmin.FenceBusinessAdmin
	siteAppAdmin   *fenceadmin.SiteApplicationAdmin
)

func initFenceAdmin() {
	fenceAdminOnce.Do(func() {
		svcAreaAdmin = fenceadmin.NewServiceAreaAdmin()
		parkingAdmin = fenceadmin.NewParkingAdmin()
		noParkingAdmin = fenceadmin.NewNoParkingAdmin()
		banRidingAdmin = fenceadmin.NewBanRidingAdmin()
		maintainAdmin = fenceadmin.NewMaintainAreaAdmin()
		customAdmin = fenceadmin.NewFenceCustomAdmin()
		rfidAdmin = fenceadmin.NewFenceRfidAdmin()
		businessAdmin = fenceadmin.NewFenceBusinessAdmin()
		siteAppAdmin = fenceadmin.NewSiteApplicationAdmin()
	})
}

func respondAdmin(c *gin.Context, data interface{}, err error) {
	if err != nil {
		if biz, ok := err.(*service.BizError); ok {
			web.WriteBizError(c, biz.Code, biz.Msg)
			return
		}
		if hasCode, ok := err.(interface{ Code() string }); ok {
			web.WriteBizError(c, hasCode.Code(), err.Error())
			return
		}
		web.WriteException(c, err.Error())
		return
	}
	result := dto.NewSuccessResult(data)
	web.RespondResult(c, &result, nil)
}

func respondRaw(c *gin.Context, data interface{}, err error) {
	if err != nil {
		web.WriteException(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, data)
}

func cmdPin(c *gin.Context) string {
	if ctx := middleware.GetCommandContext(c); ctx != nil {
		return ctx.Pin
	}
	return ""
}
