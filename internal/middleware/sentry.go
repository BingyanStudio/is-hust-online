package middleware

import (
	"strings"

	"github.com/getsentry/sentry-go"
	sentryecho "github.com/getsentry/sentry-go/echo"
	"github.com/labstack/echo/v5"
)

func SentryUserMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			// Get Sentry hub from context
			hub := sentryecho.GetHubFromContext(c)
			if hub == nil {
				return next(c)
			}

			// Configure scope with user information
			hub.ConfigureScope(func(scope *sentry.Scope) {

				// Check if user is authenticated
				usr, ok := c.Get("oauth_id").(string)
				if ok && usr != "" { // This assumes JWT middleware sets "user" in context
					scope.SetUser(sentry.User{
						ID:        usr,
						Username:  c.Get("name").(string),
						IPAddress: c.RealIP(),
					})

				} else {
					// For unauthenticated requests, just set IP
					scope.SetUser(sentry.User{
						IPAddress: c.RealIP(),
					})
					scope.SetTag("user.authenticated", "false")
				}

				// Set request information
				scope.SetTag("request.path", c.Path())
				scope.SetTag("request.method", c.Request().Method)
				scope.SetTag("request.user_agent", c.Request().UserAgent())

				// Set additional request context
				scope.SetContext("request", map[string]interface{}{
					"url":         c.Request().URL.String(),
					"method":      c.Request().Method,
					"headers":     getFilteredHeaders(c.Request().Header),
					"remote_addr": c.Request().RemoteAddr,
				})
			})

			return next(c)
		}
	}
}

// getFilteredHeaders returns request headers with sensitive information filtered out
func getFilteredHeaders(headers map[string][]string) map[string]string {
	filtered := make(map[string]string)
	sensitiveHeaders := map[string]bool{
		"authorization": true,
		"cookie":        true,
		"x-api-key":     true,
	}

	for key, values := range headers {
		lowerKey := strings.ToLower(key)
		if sensitiveHeaders[lowerKey] {
			filtered[key] = "[Filtered]"
		} else if len(values) > 0 {
			filtered[key] = values[0]
		}
	}
	return filtered
}
