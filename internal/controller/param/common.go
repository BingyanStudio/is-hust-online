package param

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
)

type PageParam struct {
	Page     int64 `query:"page" validate:"omitzero,gte=1"`
	PageSize int64 `query:"page_size" validate:"omitzero,gte=1,lte=20"`
}

type PagingMeta struct {
	Page       int64 `json:"page"`
	PageSize   int64 `json:"page_size"`
	TotalPages int64 `json:"total_pages"`
	HasMore    bool  `json:"has_more"`
}

type PagedData struct {
	Items  any        `json:"items"`
	Paging PagingMeta `json:"paging"`
}

type Validator struct {
	Validator *validator.Validate
}

// Validate 验证参数
func (cv *Validator) Validate(i any) error {
	return cv.Validator.Struct(i)
}

// GetValidator 获得验证器
func GetValidator() *Validator {
	return &Validator{
		Validator: validator.New(),
	}
}

type Resp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitzero"`
}

func Success(c *echo.Context, data any) error {

	if data == nil {
		data = []struct{}{}
	}

	return c.JSON(http.StatusOK, Resp{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

func normalizePageParam(pageParam *PageParam) (int64, int64) {
	page := int64(1)
	pageSize := int64(10)
	if pageParam == nil {
		return page, pageSize
	}
	if pageParam.Page > 0 {
		page = pageParam.Page
	}
	if pageParam.PageSize > 0 {
		pageSize = pageParam.PageSize
	}
	return page, pageSize
}

func BuildPagingMeta(pageParam *PageParam, totalItems int64) PagingMeta {
	page, pageSize := normalizePageParam(pageParam)

	totalPages := int64(0)
	if totalItems > 0 {
		totalPages = (totalItems + pageSize - 1) / pageSize
	}

	return PagingMeta{
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
		HasMore:    page < totalPages,
	}
}

func SuccessWithPaging(c *echo.Context, items any, pageParam *PageParam, totalItems int64) error {
	return Success(c, PagedData{
		Items:  items,
		Paging: BuildPagingMeta(pageParam, totalItems),
	})
}

func Error(c *echo.Context, code int, message string, err error) error {

	if err != nil {
		c.Logger().Error(err.Error(), "request", c.Response().Header().Get(echo.HeaderXRequestID))
	}

	return c.JSON(code, Resp{
		Code:    code,
		Message: message,
	})
}
