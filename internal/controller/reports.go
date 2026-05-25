package controller

import (
	"net/http"
	"strconv"

	"github.com/BingyanStudio/is-hust-online/internal/controller/param"
	"github.com/BingyanStudio/is-hust-online/internal/dao"
	"github.com/BingyanStudio/is-hust-online/internal/model"
	"github.com/labstack/echo/v5"
)

func ListReports(c *echo.Context) error {
	var pageParam param.PageParam
	if err := c.Bind(&pageParam); err != nil {
		return param.Error(c, http.StatusBadRequest, "invalid params", err)
	}

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
		if t < model.REPORT_TYPE_HOURLY || t > model.REPORT_TYPE_MONTHLY {
			return param.Error(c, http.StatusBadRequest, "type must be 0 (hourly), 1 (daily), or 2 (monthly)", nil)
		}
		reportType = &t
	}

	reports, err := dao.FindReportsBySiteID(c.Request().Context(), siteID, reportType, pageParam.Page, pageParam.PageSize)
	if err != nil {
		return param.Error(c, http.StatusInternalServerError, "failed to list reports", err)
	}

	return param.Success(c, reports)
}
