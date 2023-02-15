package repository

import (
	"database/sql"
	"errors"

	"github.com/aria3ppp/url-shortener/internal/core/domain"
	domain_errors "github.com/aria3ppp/url-shortener/internal/core/errors"
	"github.com/aria3ppp/url-shortener/internal/core/port"
)

type postgresRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) port.Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) GetLink(
	shortenedString string,
) (*domain.Link, error) {
	link := new(domain.Link)

	err := r.db.QueryRow(
		"SELECT shortened_string, url, username FROM links WHERE shortened_string = $1",
		shortenedString,
	).Scan(&link.ShortenedString, &link.URL, &link.Username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain_errors.ErrLinkNotFound
		}
		return nil, err
	}

	return link, nil
}

func (r *postgresRepository) CreateLink(link *domain.Link) error {
	_, err := r.db.Exec(
		"INSERT INTO links (shortened_string, url, username) VALUES ($1, $2, $3)",
		link.ShortenedString,
		link.URL,
		link.Username,
	)
	return err
}

func (r *postgresRepository) GetUser(username string) (*domain.User, error) {
	user := new(domain.User)

	err := r.db.QueryRow(
		"SELECT username, password FROM users WHERE username = $1",
		username,
	).Scan(&user.Username, &user.Password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain_errors.ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

func (r *postgresRepository) CreateUser(user *domain.User) error {
	_, err := r.db.Exec(
		"INSERT INTO users (username, password) VALUES ($1, $2)",
		user.Username,
		user.Password,
	)
	return err
}
