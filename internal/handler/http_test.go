package handler_test

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/aria3ppp/url-shortener/helper"
	"github.com/aria3ppp/url-shortener/internal/core/domain"
	"github.com/aria3ppp/url-shortener/internal/core/usecase"
	"github.com/aria3ppp/url-shortener/internal/generator"
	"github.com/aria3ppp/url-shortener/internal/handler"
	"github.com/aria3ppp/url-shortener/internal/repository"
	"github.com/gavv/httpexpect/v2"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/labstack/echo/v4"
)

// setup test cases
func setup() (serverURL string, teardown func()) {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Panicf(
			"handler_test.setup: postgres.WithInstance error: %s",
			err,
		)
	}
	migrator, err := migrate.NewWithDatabaseInstance(
		"file://../../migrations",
		"postgres", driver,
	)
	if err != nil {
		log.Panicf(
			"handler_test.setup: migrate.NewWithDatabaseInstance error: %s",
			err,
		)
	}

	// run migrations
	err = migrator.Up()
	if err != nil {
		log.Panicf("handler_test.setup: migrator.Up error: %s", err)
	}

	repository := repository.NewRepository(db)
	generator := generator.NewRandomStringGenerator(6)
	serviceUseCases := usecase.NewService(repository, generator)
	handler := handler.NewHandler(serviceUseCases)

	server := httptest.NewServer(helper.HandleRoutes(echo.New(), handler))

	// prepare teardown
	teardown = func() {
		// drop migrations
		err = migrator.Drop()
		if err != nil {
			log.Panicf("handler_test.teardown: migrator.Drop error: %s", err)
		}
		// close server
		server.Close()
	}

	return server.URL, teardown
}

var db *sql.DB

func TestMain(m *testing.M) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@localhost:%s/%s?sslmode=disable",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_DB"),
	)
	var err error
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Panicf(
			"handler_test.TestMain: could not connect to database %q: %s",
			dsn,
			err,
		)
	}
	err = db.Ping()
	if err != nil {
		log.Panicf(
			"handler_test.TestMain: could not ping database %q: %s",
			dsn,
			err,
		)
	}

	// Run tests
	code := m.Run()

	// close db
	err = db.Close()
	if err != nil {
		log.Panicf("handler_test.TestMain: db.Close error: %s", err)
	}

	os.Exit(code)
}

func TestHandleGetLink(t *testing.T) {
	serverURL, teardown := setup()
	t.Cleanup(teardown)

	e := httpexpect.Default(t, serverURL)

	// create helper user
	user := domain.User{
		Username: "username",
		Password: "password",
	}
	e.Request(http.MethodPost, "/user").
		WithJSON(user).
		Expect().
		Status(http.StatusOK).
		NoContent()

	linkShortenedString := "LaLiLuLeLo"
	url := serverURL + "/test/redirection-destination"

	// first there's no link
	e.Request(http.MethodGet, "/link/{shortened_string}").
		WithPath("shortened_string", linkShortenedString).
		Expect().
		Status(http.StatusNotFound)

	// create a new link
	e.Request(http.MethodPost, "/link").
		WithBasicAuth(user.Username, user.Password).
		WithJSON(domain.Link{
			ShortenedString: linkShortenedString,
			URL:             url,
		}).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		IsEqual(domain.Link{
			ShortenedString: linkShortenedString,
			URL:             url,
			Username:        user.Username,
		})

	// get link
	e.Request(http.MethodGet, "/link/{shortened_string}").
		WithPath("shortened_string", linkShortenedString).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		IsEqual(map[string]string{"redirection": "success"})
}

func TestHandleCreateLink(t *testing.T) {
	serverURL, teardown := setup()
	t.Cleanup(teardown)

	e := httpexpect.Default(t, serverURL)

	// create helper user
	user := domain.User{
		Username: "username",
		Password: "password",
	}
	e.Request(http.MethodPost, "/user").
		WithJSON(user).
		Expect().
		Status(http.StatusOK).
		NoContent()

	linkShortenedString := "LaLiLuLeLo"
	url := serverURL + "/test/redirection-destination"

	// invalid json body
	e.Request(http.MethodPost, "/link").
		WithBasicAuth(user.Username, user.Password).
		WithJSON("{invalid_json_body}").
		Expect().
		Status(http.StatusBadRequest).JSON()

	// empty url
	e.Request(http.MethodPost, "/link").
		WithBasicAuth(user.Username, user.Password).
		WithJSON(domain.Link{}).
		Expect().
		Status(http.StatusBadRequest).
		JSON().
		Object().
		IsEqual(map[string]string{
			"message": validation.Errors{
				"url": validation.ErrRequired,
			}.Error(),
		})

	// invalid url and shortened string
	e.Request(http.MethodPost, "/link").
		WithBasicAuth(user.Username, user.Password).
		WithJSON(domain.Link{
			ShortenedString: "non|alpha|numeric|shortened|string",
			URL:             "invalid|url",
		}).
		Expect().
		Status(http.StatusBadRequest).
		JSON().
		Object().
		IsEqual(map[string]string{
			"message": validation.Errors{
				"shortened_string": is.ErrAlphanumeric,
				"url":              is.ErrURL,
			}.Error(),
		})

	// unauthorized - username and password not provided
	e.Request(http.MethodPost, "/link").
		WithJSON(domain.Link{
			ShortenedString: linkShortenedString,
			URL:             url,
		}).
		Expect().
		Status(http.StatusUnauthorized).
		JSON().
		Object().
		IsEqual(map[string]string{"message": "basic authorization not provided"})

	// unauthorized - user not found
	e.Request(http.MethodPost, "/link").
		WithBasicAuth("undefined_username", user.Password).
		WithJSON(domain.Link{
			ShortenedString: linkShortenedString,
			URL:             url,
		}).
		Expect().
		Status(http.StatusUnauthorized).
		JSON().
		Object().
		IsEqual(map[string]string{"message": "invalid username or password"})

	// unauthorized - incorrect password
	e.Request(http.MethodPost, "/link").
		WithBasicAuth(user.Username, "incorrect_password").
		WithJSON(domain.Link{
			ShortenedString: linkShortenedString,
			URL:             url,
		}).
		Expect().
		Status(http.StatusUnauthorized).
		JSON().
		Object().
		IsEqual(map[string]string{"message": "invalid username or password"})

	// create link
	e.Request(http.MethodPost, "/link").
		WithBasicAuth(user.Username, user.Password).
		WithJSON(domain.Link{
			ShortenedString: linkShortenedString,
			URL:             url,
		}).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		IsEqual(domain.Link{
			ShortenedString: linkShortenedString,
			URL:             url,
			Username:        user.Username,
		})

	// assert link is redirected
	e.Request(http.MethodGet, "/link/{shortened_string}").
		WithPath("shortened_string", linkShortenedString).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		IsEqual(map[string]string{"redirection": "success"})

	// shortened string have used
	e.Request(http.MethodPost, "/link").
		WithBasicAuth(user.Username, user.Password).
		WithJSON(domain.Link{
			ShortenedString: linkShortenedString,
			URL:             url,
		}).
		Expect().
		Status(http.StatusConflict).
		JSON().
		Object().
		IsEqual(map[string]string{"message": "shortened string have used"})
}

func TestHandleGetLinkUser(t *testing.T) {
	serverURL, teardown := setup()
	t.Cleanup(teardown)

	e := httpexpect.Default(t, serverURL)

	// create the user
	user := domain.User{
		Username: "username",
		Password: "password",
	}
	e.Request(http.MethodPost, "/user").
		WithJSON(user).
		Expect().
		Status(http.StatusOK).
		NoContent()

	linkShortenedString := "LaLiLuLeLo"
	url := serverURL + "/test/redirection-destination"

	// create a new link
	e.Request(http.MethodPost, "/link").
		WithBasicAuth(user.Username, user.Password).
		WithJSON(domain.Link{
			ShortenedString: linkShortenedString,
			URL:             url,
		}).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		IsEqual(domain.Link{
			ShortenedString: linkShortenedString,
			URL:             url,
			Username:        user.Username,
		})

	// link not found
	e.Request(http.MethodGet, "/link/{shortened_string}/user").
		WithBasicAuth(user.Username, user.Password).
		WithPath("shortened_string", "undefined_link").
		Expect().
		Status(http.StatusNotFound).
		JSON().
		Object().
		IsEqual(map[string]string{"message": "link not found"})

	// get link user
	e.Request(http.MethodGet, "/link/{shortened_string}/user").
		WithBasicAuth(user.Username, user.Password).
		WithPath("shortened_string", linkShortenedString).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		IsEqual(domain.User{
			Username: user.Username,
			// password ommited from serialization
		})
}

func TestHandleCreateUser(t *testing.T) {
	serverURL, teardown := setup()
	t.Cleanup(teardown)

	e := httpexpect.Default(t, serverURL)

	// invalid json body
	e.Request(http.MethodPost, "/user").
		WithJSON("{invalid_json_body}").
		Expect().
		Status(http.StatusBadRequest).JSON()

	// empty username and password
	e.Request(http.MethodPost, "/user").
		WithJSON(domain.Link{}).
		Expect().
		Status(http.StatusBadRequest).
		JSON().
		Object().
		IsEqual(map[string]string{
			"message": validation.Errors{
				"username": validation.ErrRequired,
				"password": validation.ErrRequired,
			}.Error(),
		})

	// invalid username and password
	e.Request(http.MethodPost, "/user").
		WithJSON(domain.User{
			Username: "un",
			Password: "pw",
		}).
		Expect().
		Status(http.StatusBadRequest).
		JSON().
		Object().
		IsEqual(map[string]string{
			"message": validation.Errors{
				"username": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{"min": 8, "max": 40},
				),
				"password": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{"min": 8, "max": 40},
				),
			}.Error(),
		})

	// create user
	user := domain.User{
		Username: "username",
		Password: "password",
	}
	e.Request(http.MethodPost, "/user").
		WithJSON(user).
		Expect().
		Status(http.StatusOK).
		NoContent()

	// create a link
	linkShortenedString := "LaLiLuLeLo"
	url := serverURL + "/test/redirection-destination"

	e.Request(http.MethodPost, "/link").
		WithBasicAuth(user.Username, user.Password).
		WithJSON(domain.Link{
			ShortenedString: linkShortenedString,
			URL:             url,
		}).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		IsEqual(domain.Link{
			ShortenedString: linkShortenedString,
			URL:             url,
			Username:        user.Username,
		})

	// assert user is created
	e.Request(http.MethodGet, "/link/{shortened_string}/user").
		WithBasicAuth(user.Username, user.Password).
		WithPath("shortened_string", linkShortenedString).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		IsEqual(domain.User{
			Username: user.Username,
			// password ommited from serialization
		})

	// username have taken
	e.Request(http.MethodPost, "/user").
		WithJSON(user).
		Expect().
		Status(http.StatusConflict).
		JSON().
		Object().
		IsEqual(map[string]string{"message": "username have taken"})
}
