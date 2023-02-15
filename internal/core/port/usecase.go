package port

import "github.com/aria3ppp/url-shortener/internal/core/domain"

type ServiceUseCases interface {
	// link usecases
	GetLink(shortenedString string) (*domain.Link, error)
	CreateLink(
		url string,
		shortenedString string,
		user *domain.User,
	) (*domain.Link, error)
	// user usecases
	GetLinkUser(shortenedString string) (*domain.User, error)
	CreateUser(user *domain.User) error
}
