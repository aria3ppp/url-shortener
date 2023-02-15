package port

import "github.com/aria3ppp/url-shortener/internal/core/domain"

//go:generate mockgen -package mockups -destination mockups/mock_repository.go . Repository

type Repository interface {
	// link
	GetLink(shortenedString string) (*domain.Link, error)
	CreateLink(link *domain.Link) error
	// user
	GetUser(username string) (*domain.User, error)
	CreateUser(user *domain.User) error
}
