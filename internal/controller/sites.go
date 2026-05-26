package controller

import (
	"net/http"
	"time"

	"github.com/BingyanStudio/is-hust-online/internal/controller/param"
	"github.com/BingyanStudio/is-hust-online/internal/dao"
	"github.com/BingyanStudio/is-hust-online/internal/model"
	"github.com/labstack/echo/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type CreateSiteRequest struct {
	Name        string `json:"name" validate:"required"`
	URL         string `json:"url" validate:"required"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

type UpdateSiteRequest struct {
	Name        *string `json:"name"`
	URL         *string `json:"url"`
	Type        *string `json:"type"`
	Description *string `json:"description"`
	Status      *int    `json:"status"`
}

func ListSites(c *echo.Context) error {
	var pageParam param.PageParam
	if err := c.Bind(&pageParam); err != nil {
		return param.Error(c, http.StatusBadRequest, "invalid params", err)
	}

	sites, total, err := dao.FindSites(c.Request().Context(), bson.M{}, pageParam.Page, pageParam.PageSize)
	if err != nil {
		return param.Error(c, http.StatusInternalServerError, "failed to list sites", err)
	}

	return param.SuccessWithPaging(c, sites, &pageParam, total)
}

func GetSite(c *echo.Context) error {
	id, err := bson.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		return param.Error(c, http.StatusBadRequest, "invalid id", err)
	}

	site, err := dao.FindSiteByID(c.Request().Context(), id)
	if err != nil {
		return param.Error(c, http.StatusNotFound, "site not found", err)
	}

	return param.Success(c, site)
}

func CreateSite(c *echo.Context) error {
	var req CreateSiteRequest
	if err := c.Bind(&req); err != nil {
		return param.Error(c, http.StatusBadRequest, "invalid request body", err)
	}
	if err := c.Validate(req); err != nil {
		return param.Error(c, http.StatusBadRequest, "validation failed", err)
	}

	site := &model.Site{
		Name:        req.Name,
		URL:         req.URL,
		Type:        req.Type,
		Description: req.Description,
		Status:      model.SITE_STATUS_ENABLED,
		CreatedAt:   time.Now().Unix(),
	}

	if err := dao.InsertSite(c.Request().Context(), site); err != nil {
		return param.Error(c, http.StatusInternalServerError, "failed to create site", err)
	}

	return param.Success(c, site)
}

func UpdateSite(c *echo.Context) error {
	id, err := bson.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		return param.Error(c, http.StatusBadRequest, "invalid id", err)
	}

	var req UpdateSiteRequest
	if err := c.Bind(&req); err != nil {
		return param.Error(c, http.StatusBadRequest, "invalid request body", err)
	}

	update := bson.M{}
	if req.Name != nil {
		update["name"] = *req.Name
	}
	if req.URL != nil {
		update["url"] = *req.URL
	}
	if req.Type != nil {
		update["type"] = *req.Type
	}
	if req.Description != nil {
		update["description"] = *req.Description
	}
	if req.Status != nil {
		update["status"] = *req.Status
	}

	if len(update) == 0 {
		return param.Error(c, http.StatusBadRequest, "no fields to update", nil)
	}

	if err := dao.UpdateSite(c.Request().Context(), id, update); err != nil {
		return param.Error(c, http.StatusInternalServerError, "failed to update site", err)
	}

	return param.Success(c, nil)
}

func DeleteSite(c *echo.Context) error {
	id, err := bson.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		return param.Error(c, http.StatusBadRequest, "invalid id", err)
	}

	if err := dao.DeleteSite(c.Request().Context(), id); err != nil {
		return param.Error(c, http.StatusInternalServerError, "failed to delete site", err)
	}

	return param.Success(c, nil)
}
