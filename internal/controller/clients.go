package controller

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/BingyanStudio/is-hust-online/internal/controller/param"
	"github.com/BingyanStudio/is-hust-online/internal/dao"
	"github.com/BingyanStudio/is-hust-online/internal/model"
	"github.com/labstack/echo/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type CreateClientRequest struct {
	Name         string   `json:"name" validate:"required"`
	Location     string   `json:"location"`
	Capabilities int32    `json:"capabilities"`
	Labels       []string `json:"labels"`
}

type UpdateClientRequest struct {
	Name         *string  `json:"name"`
	Location     *string  `json:"location"`
	Capabilities *int32   `json:"capabilities"`
	Labels       []string `json:"labels"`
	Status       *int     `json:"status"`
}

func ListClients(c *echo.Context) error {
	var pageParam param.PageParam
	if err := c.Bind(&pageParam); err != nil {
		return param.Error(c, http.StatusBadRequest, "invalid params", err)
	}

	clients, total, err := dao.FindClients(c.Request().Context(), bson.M{}, pageParam.Page, pageParam.PageSize)
	if err != nil {
		return param.Error(c, http.StatusInternalServerError, "failed to list clients", err)
	}

	return param.SuccessWithPaging(c, clients, &pageParam, total)
}

func GetClient(c *echo.Context) error {
	id, err := bson.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		return param.Error(c, http.StatusBadRequest, "invalid id", err)
	}

	client, err := dao.FindClientByID(c.Request().Context(), id)
	if err != nil {
		return param.Error(c, http.StatusNotFound, "client not found", err)
	}

	return param.Success(c, client)
}

func generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func CreateClient(c *echo.Context) error {
	var req CreateClientRequest
	if err := c.Bind(&req); err != nil {
		return param.Error(c, http.StatusBadRequest, "invalid request body", err)
	}
	if err := c.Validate(req); err != nil {
		return param.Error(c, http.StatusBadRequest, "validation failed", err)
	}

	client := &model.Client{
		Token:        generateToken(),
		Name:         req.Name,
		Location:     req.Location,
		Capabilities: req.Capabilities,
		Labels:       req.Labels,
		Status:       model.CLIENT_STATUS_OFFLINE,
		CreatedAt:    time.Now().Unix(),
	}

	if err := dao.InsertClient(c.Request().Context(), client); err != nil {
		return param.Error(c, http.StatusInternalServerError, "failed to create client", err)
	}

	return param.Success(c, client)
}

func UpdateClient(c *echo.Context) error {
	id, err := bson.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		return param.Error(c, http.StatusBadRequest, "invalid id", err)
	}

	var req UpdateClientRequest
	if err := c.Bind(&req); err != nil {
		return param.Error(c, http.StatusBadRequest, "invalid request body", err)
	}

	update := bson.M{}
	if req.Name != nil {
		update["name"] = *req.Name
	}
	if req.Location != nil {
		update["location"] = *req.Location
	}
	if req.Capabilities != nil {
		update["capabilities"] = *req.Capabilities
	}
	if req.Labels != nil {
		update["labels"] = req.Labels
	}
	if req.Status != nil {
		update["status"] = *req.Status
	}

	if len(update) == 0 {
		return param.Error(c, http.StatusBadRequest, "no fields to update", nil)
	}

	if err := dao.UpdateClient(c.Request().Context(), id, update); err != nil {
		return param.Error(c, http.StatusInternalServerError, "failed to update client", err)
	}

	return param.Success(c, nil)
}

func DeleteClient(c *echo.Context) error {
	id, err := bson.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		return param.Error(c, http.StatusBadRequest, "invalid id", err)
	}

	if err := dao.DeleteClient(c.Request().Context(), id); err != nil {
		return param.Error(c, http.StatusInternalServerError, "failed to delete client", err)
	}

	return param.Success(c, nil)
}
