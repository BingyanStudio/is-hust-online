package views

import (
	"github.com/BingyanStudio/is-hust-online/internal/controller"
	mymw "github.com/BingyanStudio/is-hust-online/internal/middleware"
	"github.com/labstack/echo/v5"
)

func InitViews(e *echo.Echo) {
	api := e.Group("/api")
	api.Use(mymw.BasicAuthForMutations())

	// Sites
	api.GET("/sites", controller.ListSites)
	api.GET("/sites/:id", controller.GetSite)
	api.POST("/sites", controller.CreateSite)
	api.PUT("/sites/:id", controller.UpdateSite)
	api.DELETE("/sites/:id", controller.DeleteSite)

	// Clients
	api.GET("/clients", controller.ListClients)
	api.GET("/clients/:id", controller.GetClient)
	api.POST("/clients", controller.CreateClient)
	api.PUT("/clients/:id", controller.UpdateClient)
	api.DELETE("/clients/:id", controller.DeleteClient)

	// Checks
	api.GET("/checks", controller.ListChecks)

	// Reports
	api.GET("/reports", controller.ListReports)
}
