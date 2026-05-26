package views

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"
)

func InitStatic(e *echo.Echo, fsys fs.FS) {
	fileServer := http.FileServer(http.FS(fsys))

	e.Any("/*", func(c *echo.Context) error {
		path := strings.TrimPrefix(c.Path(), "/")

		if path == "" {
			path = "index.html"
		}

		// Try to serve the file directly
		if f, err := fs.ReadFile(fsys, path); err == nil {
			ct := detectContentType(path, f)
			return c.Blob(http.StatusOK, ct, f)
		}

		// Fallback to index.html for SPA routing
		if f, err := fs.ReadFile(fsys, "index.html"); err == nil {
			return c.Blob(http.StatusOK, "text/html", f)
		}

		fileServer.ServeHTTP(c.Response(), c.Request())
		return nil
	})
}

func detectContentType(path string, data []byte) string {
	switch {
	case strings.HasSuffix(path, ".js"):
		return "application/javascript"
	case strings.HasSuffix(path, ".css"):
		return "text/css"
	case strings.HasSuffix(path, ".html"):
		return "text/html"
	case strings.HasSuffix(path, ".json"):
		return "application/json"
	case strings.HasSuffix(path, ".svg"):
		return "image/svg+xml"
	case strings.HasSuffix(path, ".png"):
		return "image/png"
	case strings.HasSuffix(path, ".ico"):
		return "image/x-icon"
	case strings.HasSuffix(path, ".woff"):
		return "font/woff"
	case strings.HasSuffix(path, ".woff2"):
		return "font/woff2"
	default:
		return http.DetectContentType(data)
	}
}
