package controller

import (
	"github.com/gin-gonic/gin"

	"ebike-fence-go/internal/api/dto"
)

func siteApplicationHandlers() map[string]gin.HandlerFunc {
	initFenceAdmin()
	return map[string]gin.HandlerFunc{
		"/SiteApplication/SiteApplication":     handleSiteApplicationCreate,
		"/SiteApplication/dealSiteApplication": handleSiteApplicationDeal,
		"/SiteApplication/pageSiteApplication": handleSiteApplicationPage,
	}
}

func handleSiteApplicationCreate(c *gin.Context) {
	var req dto.SiteApplicationCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	res, err := siteAppAdmin.Create(c.Request.Context(), tenantID, cmdPin(c), req)
	respondAdmin(c, res, err)
}

func handleSiteApplicationDeal(c *gin.Context) {
	var req dto.SiteApplicationCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	err := siteAppAdmin.Deal(c.Request.Context(), tenantID, cmdPin(c), req)
	respondAdmin(c, nil, err)
}

func handleSiteApplicationPage(c *gin.Context) {
	var req dto.ApplicationQuery
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	res, err := siteAppAdmin.Page(c.Request.Context(), tenantID, req)
	respondAdmin(c, res, err)
}
