package controller

import (
	"net/http"

	"github.com/BingyanStudio/is-hust-online/internal/checktype"
	"github.com/BingyanStudio/is-hust-online/internal/controller/param"
	"github.com/BingyanStudio/is-hust-online/internal/dao"
	"github.com/BingyanStudio/is-hust-online/internal/model"
	myproto "github.com/BingyanStudio/is-hust-online/pkg/proto"
	"github.com/labstack/echo/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type CreateCheckConfigRequest struct {
	SiteID        string `json:"site_id" validate:"required"`
	ClientID      string `json:"client_id" validate:"required"`
	CheckType     int32  `json:"check_type" validate:"required"`
	CheckInterval string `json:"check_interval" validate:"required"`
	CheckExtra    any    `json:"check_extra"`
}

type UpdateCheckConfigRequest struct {
	ClientID      *string `json:"client_id"`
	CheckType     *int32  `json:"check_type"`
	CheckInterval *string `json:"check_interval"`
	CheckExtra    *any    `json:"check_extra"`
	Status        *int32  `json:"status"`
}

func ListCheckConfigs(c *echo.Context) error {
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

	ccs, total, err := dao.FindCheckConfigs(c.Request().Context(), filter, pageParam.Page, pageParam.PageSize)
	if err != nil {
		return param.Error(c, http.StatusInternalServerError, "failed to list check configs", err)
	}

	return param.SuccessWithPaging(c, ccs, &pageParam, total)
}

func GetCheckConfig(c *echo.Context) error {
	id, err := bson.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		return param.Error(c, http.StatusBadRequest, "invalid id", err)
	}

	cc, err := dao.FindCheckConfigByID(c.Request().Context(), id)
	if err != nil {
		return param.Error(c, http.StatusNotFound, "check config not found", err)
	}

	return param.Success(c, cc)
}

func CreateCheckConfig(c *echo.Context) error {
	var req CreateCheckConfigRequest
	if err := c.Bind(&req); err != nil {
		return param.Error(c, http.StatusBadRequest, "invalid request body", err)
	}
	if err := c.Validate(req); err != nil {
		return param.Error(c, http.StatusBadRequest, "validation failed", err)
	}

	siteID, err := bson.ObjectIDFromHex(req.SiteID)
	if err != nil {
		return param.Error(c, http.StatusBadRequest, "invalid site_id", err)
	}
	clientID, err := bson.ObjectIDFromHex(req.ClientID)
	if err != nil {
		return param.Error(c, http.StatusBadRequest, "invalid client_id", err)
	}

	// Validate site exists
	if _, err := dao.FindSiteByID(c.Request().Context(), siteID); err != nil {
		return param.Error(c, http.StatusBadRequest, "site not found", err)
	}

	// Validate client exists and has matching capability
	client, err := dao.FindClientByID(c.Request().Context(), clientID)
	if err != nil {
		return param.Error(c, http.StatusBadRequest, "client not found", err)
	}

	requiredBit := checktype.Bit(myproto.CheckType(req.CheckType))
	if client.Capabilities&requiredBit == 0 {
		return param.Error(c, http.StatusBadRequest, "client does not support this check type", nil)
	}

	cc := &model.CheckConfig{
		ID:            bson.NewObjectID(),
		SiteID:        siteID,
		ClientID:      clientID,
		Status:        model.CHECK_ENABLED,
		CheckType:     req.CheckType,
		CheckInterval: req.CheckInterval,
		CheckExtra:    req.CheckExtra,
	}

	if err := dao.InsertCheckConfig(c.Request().Context(), cc); err != nil {
		return param.Error(c, http.StatusInternalServerError, "failed to create check config", err)
	}

	return param.Success(c, cc)
}

func UpdateCheckConfig(c *echo.Context) error {
	id, err := bson.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		return param.Error(c, http.StatusBadRequest, "invalid id", err)
	}

	var req UpdateCheckConfigRequest
	if err := c.Bind(&req); err != nil {
		return param.Error(c, http.StatusBadRequest, "invalid request body", err)
	}

	update := bson.M{}
	if req.ClientID != nil {
		update["client_id"] = *req.ClientID
	}
	if req.CheckType != nil {
		update["check_type"] = *req.CheckType
	}
	if req.CheckInterval != nil {
		update["check_interval"] = *req.CheckInterval
	}
	if req.CheckExtra != nil {
		update["check_extra"] = *req.CheckExtra
	}
	if req.Status != nil {
		update["status"] = *req.Status
	}

	if len(update) == 0 {
		return param.Error(c, http.StatusBadRequest, "no fields to update", nil)
	}

	if err := dao.UpdateCheckConfig(c.Request().Context(), id, update); err != nil {
		return param.Error(c, http.StatusInternalServerError, "failed to update check config", err)
	}

	return param.Success(c, nil)
}

func DeleteCheckConfig(c *echo.Context) error {
	id, err := bson.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		return param.Error(c, http.StatusBadRequest, "invalid id", err)
	}

	if err := dao.DeleteCheckConfig(c.Request().Context(), id); err != nil {
		return param.Error(c, http.StatusInternalServerError, "failed to delete check config", err)
	}

	return param.Success(c, nil)
}
