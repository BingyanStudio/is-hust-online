package controller

import (
	"net/http"
	"strconv"

	"github.com/BingyanStudio/is-hust-online/internal/controller/param"
	"github.com/BingyanStudio/is-hust-online/internal/dao"
	"github.com/BingyanStudio/is-hust-online/internal/model"
	"github.com/labstack/echo/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
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

	siteIDObj, err := bson.ObjectIDFromHex(siteID)
	if err != nil {
		return param.Error(c, http.StatusBadRequest, "invalid site_id", err)
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

	var checkConfigID *bson.ObjectID
	if ccIDStr := c.QueryParam("check_config_id"); ccIDStr != "" {
		ccID, err := bson.ObjectIDFromHex(ccIDStr)
		if err != nil {
			return param.Error(c, http.StatusBadRequest, "invalid check_config_id", err)
		}
		checkConfigID = &ccID
	}

	reports, err := dao.FindReportsBySiteID(c.Request().Context(), siteIDObj, checkConfigID, reportType, pageParam.Page, pageParam.PageSize)
	if err != nil {
		return param.Error(c, http.StatusInternalServerError, "failed to list reports", err)
	}

	return param.Success(c, reports)
}
