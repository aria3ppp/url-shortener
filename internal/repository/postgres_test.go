package repository_test

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/aria3ppp/url-shortener/internal/core/domain"
	domain_errors "github.com/aria3ppp/url-shortener/internal/core/errors"
	"github.com/aria3ppp/url-shortener/internal/repository"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/stretchr/testify/require"
)

var db *sql.DB

// setup test cases
func setup() (teardown func()) {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Panicf(
			"repository_test.setup: postgres.WithInstance error: %s",
			err,
		)
	}
	migrator, err := migrate.NewWithDatabaseInstance(
		"file://../../migrations",
		"postgres", driver,
	)
	if err != nil {
		log.Panicf(
			"repository_test.setup: migrate.NewWithDatabaseInstance error: %s",
			err,
		)
	}

	// run migrations
	err = migrator.Up()
	if err != nil {
		log.Panicf("repository_test.setup: migrator.Up error: %s", err)
	}

	// prepare teardown
	return func() {
		// drop migrations
		err = migrator.Drop()
		if err != nil {
			log.Panicf("repository_test.teardown: migrator.Drop error: %s", err)
		}
	}
}

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
			"repository_test.TestMain: could not connect to database %q: %s",
			dsn,
			err,
		)
	}
	err = db.Ping()
	if err != nil {
		log.Panicf(
			"repository_test.TestMain: could not ping database %q: %s",
			dsn,
			err,
		)
	}

	// Run tests
	code := m.Run()

	// close db
	err = db.Close()
	if err != nil {
		log.Panicf("repository_test.TestMain: db.Close error: %s", err)
	}

	os.Exit(code)
}

func TestGetLink(t *testing.T) {
	require := require.New(t)

	teardown := setup()
	t.Cleanup(teardown)

	r := repository.NewRepository(db)

	// create helper user
	user := &domain.User{Username: "username"}
	err := r.CreateUser(user)
	require.NoError(err)

	linkShortenedString := "LaLiLuLeLo"

	// first there's no link
	link, err := r.GetLink(linkShortenedString)
	require.Equal(err, domain_errors.ErrLinkNotFound)
	require.Nil(link)

	// create a new link
	err = r.CreateLink(
		&domain.Link{
			ShortenedString: linkShortenedString,
			URL:             "url",
			Username:        user.Username,
		},
	)
	require.NoError(err)

	// get link
	link, err = r.GetLink(linkShortenedString)
	require.NoError(err)
	require.Equal(
		&domain.Link{
			ShortenedString: linkShortenedString,
			URL:             "url",
			Username:        user.Username,
		},
		link,
	)
}

func TestCreateLink(t *testing.T) {
	require := require.New(t)

	teardown := setup()
	t.Cleanup(teardown)

	r := repository.NewRepository(db)

	// create helper user
	user := &domain.User{Username: "username"}
	err := r.CreateUser(user)
	require.NoError(err)

	linkShortenedString := "LaLiLuLeLo"

	// create link
	err = r.CreateLink(
		&domain.Link{
			ShortenedString: linkShortenedString,
			URL:             "url",
			Username:        user.Username,
		},
	)
	require.NoError(err)

	// assert link is created
	link, err := r.GetLink(linkShortenedString)
	require.NoError(err)
	require.Equal(
		&domain.Link{
			ShortenedString: linkShortenedString,
			URL:             "url",
			Username:        user.Username,
		},
		link,
	)
}

func TestGetUser(t *testing.T) {
	require := require.New(t)

	teardown := setup()
	t.Cleanup(teardown)

	r := repository.NewRepository(db)

	username := "snakePlissken"

	// first there's no user
	user, err := r.GetUser(username)
	require.Equal(err, domain_errors.ErrUserNotFound)
	require.Nil(user)

	// create a new user
	err = r.CreateUser(&domain.User{
		Username: username,
		Password: "password",
	})
	require.NoError(err)

	// get user
	user, err = r.GetUser(username)
	require.NoError(err)
	require.Equal(
		&domain.User{
			Username: username,
			Password: "password",
		},
		user,
	)
}

func TestCreateUser(t *testing.T) {
	require := require.New(t)

	teardown := setup()
	t.Cleanup(teardown)

	r := repository.NewRepository(db)

	username := "snakePlissken"

	// create user
	err := r.CreateUser(
		&domain.User{
			Username: username,
			Password: "password",
		},
	)
	require.NoError(err)

	// assert user is created
	user, err := r.GetUser(username)
	require.NoError(err)
	require.Equal(
		&domain.User{
			Username: username,
			Password: "password",
		},
		user,
	)
}
