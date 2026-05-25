package middleware

import (
	"crypto/subtle"
	"net/http"

	"github.com/BingyanStudio/is-hust-online/internal/config"
	"github.com/labstack/echo/v5"
)

func BasicAuthForMutations() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			method := c.Request().Method
			if method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions {
				return next(c)
			}

			user := config.C.BasicAuthUser
			pass := config.C.BasicAuthPassword
			if user == "" && pass == "" {
				return next(c)
			}

			reqUser, reqPass, ok := c.Request().BasicAuth()
			if !ok ||
				subtle.ConstantTimeCompare([]byte(reqUser), []byte(user)) != 1 ||
				subtle.ConstantTimeCompare([]byte(reqPass), []byte(pass)) != 1 {
				c.Response().Header().Set("WWW-Authenticate", `Basic realm="api"`)
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
			}

			return next(c)
		}
	}
}
