package middleware

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v5"
)

func CustomHTTPErrorHandler(c *echo.Context, err error) {
	if resp, uErr := echo.UnwrapResponse(c.Response()); uErr == nil {
		if resp.Committed {
			return
		}
	}

	code := http.StatusInternalServerError
	msg := "服务器内部错误"

	var sc echo.HTTPStatusCoder
	if errors.As(err, &sc) {
		if tmp := sc.StatusCode(); tmp != 0 {
			code = tmp
			msg = http.StatusText(code)
		}
	}

	// 根据请求方法返回不同的响应
	if c.Request().Method == http.MethodHead {
		c.NoContent(code)
	} else {
		c.JSON(code, map[string]any{
			"code":    code,
			"message": msg,
		})
	}
}
