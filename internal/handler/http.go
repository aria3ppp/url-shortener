package handler

import (
	"errors"
	"net/http"

	"github.com/aria3ppp/url-shortener/internal/core/domain"
	domain_errors "github.com/aria3ppp/url-shortener/internal/core/errors"
	"github.com/aria3ppp/url-shortener/internal/core/port"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	serviceUseCases port.ServiceUseCases
}

func NewHandler(serviceUseCases port.ServiceUseCases) *Handler {
	return &Handler{serviceUseCases: serviceUseCases}
}

func (h *Handler) HandleGetLink(c echo.Context) error {
	shortenedString := c.Param("shortened_string")

	link, err := h.serviceUseCases.GetLink(shortenedString)
	if err != nil {
		if errors.Is(err, domain_errors.ErrLinkNotFound) {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(err)
	}

	return c.Redirect(http.StatusPermanentRedirect, link.URL)
}

func (h *Handler) HandleCreateLink(c echo.Context) error {
	// parse and validate url and shortened string
	var body domain.Link
	if httpError := (&echo.DefaultBinder{}).BindBody(c, &body); httpError != nil {
		return httpError
	}
	if err := body.Validate(); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// fetch username:password off the basic authorization
	username, password, ok := c.Request().BasicAuth()
	if !ok {
		return echo.NewHTTPError(
			http.StatusUnauthorized,
			"basic authorization not provided",
		)
	}

	link, err := h.serviceUseCases.CreateLink(
		body.URL,
		body.ShortenedString,
		&domain.User{Username: username, Password: password},
	)
	if err != nil {
		if errors.Is(err, domain_errors.ErrUserNotFound) ||
			errors.Is(err, domain_errors.ErrIncorrectPassword) {
			return echo.NewHTTPError(
				http.StatusUnauthorized,
				"invalid username or password",
			)
		}
		if errors.Is(err, domain_errors.ErrUsedShortenedString) {
			return echo.NewHTTPError(
				http.StatusConflict,
				"shortened string have used",
			)
		}
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(err)
	}

	return c.JSON(http.StatusOK, link)
}

func (h *Handler) HandleGetLinkUser(c echo.Context) error {
	shortenedString := c.Param("shortened_string")

	user, err := h.serviceUseCases.GetLinkUser(shortenedString)
	if err != nil {
		if errors.Is(err, domain_errors.ErrLinkNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "link not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(err)
	}

	return c.JSON(http.StatusOK, user)
}

func (h *Handler) HandleCreateUser(c echo.Context) error {
	// parse and validate username and password
	var user domain.User
	if httpError := (&echo.DefaultBinder{}).BindBody(c, &user); httpError != nil {
		return httpError
	}
	if err := user.Validate(); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	err := h.serviceUseCases.CreateUser(&user)
	if err != nil {
		if errors.Is(err, domain_errors.ErrUsernameTaken) {
			return echo.NewHTTPError(
				http.StatusConflict,
				"username have taken",
			)
		}
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(err)
	}

	return c.NoContent(http.StatusOK)
}
