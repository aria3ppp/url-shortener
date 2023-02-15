package helper

import (
	"net/http"

	"github.com/aria3ppp/url-shortener/internal/handler"
	"github.com/labstack/echo/v4"
)

func HandleRoutes(router *echo.Echo, handler *handler.Handler) http.Handler {
	router.POST("/link", handler.HandleCreateLink)
	router.GET("/link/:shortened_string", handler.HandleGetLink)
	router.GET("/link/:shortened_string/user", handler.HandleGetLinkUser)
	router.POST("/user", handler.HandleCreateUser)

	// define an endpoint for self redirection used in end2end tests
	router.GET(
		"/test/redirection-destination",
		func(c echo.Context) error { return c.JSON(http.StatusOK, echo.Map{"redirection": "success"}) },
	)
	return router.Server.Handler
}
