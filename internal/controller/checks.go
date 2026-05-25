package controller

import (
	"net/http"

	"github.com/BingyanStudio/is-hust-online/internal/controller/param"
	"github.com/BingyanStudio/is-hust-online/internal/dao"
	"github.com/labstack/echo/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func ListChecks(c *echo.Context) error {
	var pageParam param.PageParam
	if err := c.Bind(&pageParam); err != nil {
		return param.Error(c, http.StatusBadRequest, "invalid params", err)
	}

	filter := bson.M{}
	if siteID := c.QueryParam("site_id"); siteID != "" {
		filter["site_id"] = siteID
	}
	if clientID := c.QueryParam("client_id"); clientID != "" {
		filter["client_id"] = clientID
	}

	checks, total, err := dao.FindChecks(c.Request().Context(), filter, pageParam.Page, pageParam.PageSize)
	if err != nil {
		return param.Error(c, http.StatusInternalServerError, "failed to list checks", err)
	}

	return param.SuccessWithPaging(c, checks, &pageParam, total)
}
