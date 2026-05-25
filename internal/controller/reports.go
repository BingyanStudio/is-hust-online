package controller

import (
	"net/http"
	"strconv"

	"github.com/BingyanStudio/is-hust-online/internal/controller/param"
	"github.com/BingyanStudio/is-hust-online/internal/dao"
	"github.com/labstack/echo/v5"
)

func ListReports(c *echo.Context) error {
	siteID := c.QueryParam("site_id")
	if siteID == "" {
		return param.Error(c, http.StatusBadRequest, "site_id is required", nil)
	}

	var reportType *int
	if typeStr := c.QueryParam("type"); typeStr != "" {
		t, err := strconv.Atoi(typeStr)
		if err != nil {
			return param.Error(c, http.StatusBadRequest, "invalid type", err)
		}
		reportType = &t
	}

	reports, err := dao.FindReportsBySiteID(c.Request().Context(), siteID, reportType)
	if err != nil {
		return param.Error(c, http.StatusInternalServerError, "failed to list reports", err)
	}

	return param.Success(c, reports)
}
