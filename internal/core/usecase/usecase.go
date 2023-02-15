package usecase

import (
	"errors"
	"fmt"

	"github.com/aria3ppp/url-shortener/internal/core/domain"
	domain_errors "github.com/aria3ppp/url-shortener/internal/core/errors"
	"github.com/aria3ppp/url-shortener/internal/core/port"
)

type serviceUseCases struct {
	repo      port.Repository
	generator port.RandomStringGenerator
}

func NewService(
	repo port.Repository,
	generator port.RandomStringGenerator,
) port.ServiceUseCases {
	return &serviceUseCases{repo: repo, generator: generator}
}

func (s *serviceUseCases) GetLink(
	shortenedString string,
) (*domain.Link, error) {
	link, err := s.repo.GetLink(shortenedString)
	if err != nil {
		if errors.Is(err, domain_errors.ErrLinkNotFound) {
			return nil, fmt.Errorf(
				"usecase.GetLink: link don't exists: %w", err)
		}
		return nil, fmt.Errorf(
			"usecase.GetLink: repository.GetLink unhandled error: %w", err)
	}
	return link, nil
}

func (s *serviceUseCases) CreateLink(
	url string,
	shortenedString string,
	user *domain.User,
) (*domain.Link, error) {
	// check user exists
	repoUser, err := s.repo.GetUser(user.Username)
	if err != nil {
		if errors.Is(err, domain_errors.ErrUserNotFound) {
			return nil, fmt.Errorf(
				"usecase.CreateLink: user don't exists: %w", err)
		}
		return nil, fmt.Errorf(
			"usecase.CreateLink: repository.GetUser unhandled error: %w", err)
	}

	// check the user password
	if repoUser.Password != user.Password {
		return nil, fmt.Errorf(
			"usecase.CreateLink: user password don't match: %w",
			domain_errors.ErrIncorrectPassword,
		)
	}

	if shortenedString != "" {
		// check user given shortened string is not used
		_, err := s.repo.GetLink(shortenedString)
		if err == nil {
			return nil, fmt.Errorf(
				"usecase.CreateLink: user given shortened string already used: %w",
				domain_errors.ErrUsedShortenedString,
			)
		} else if !errors.Is(err, domain_errors.ErrLinkNotFound) {
			return nil, fmt.Errorf(
				"usecase.CreateLink: repository.GetLink unhandled error: %w", err)
		}
	} else {
		// otherwise generate a random string
		shortenedString = s.generator.RandomString()
	}

	// create link
	link := &domain.Link{
		ShortenedString: shortenedString,
		URL:             url,
		Username:        repoUser.Username,
	}
	err = s.repo.CreateLink(link)
	if err != nil {
		return nil, fmt.Errorf(
			"usecase.CreateLink: repository.CreateLink unhandled error: %w",
			err,
		)
	}

	return link, nil
}

func (s *serviceUseCases) GetLinkUser(
	shortenedString string,
) (*domain.User, error) {
	// get link
	link, err := s.repo.GetLink(shortenedString)
	if err != nil {
		if errors.Is(err, domain_errors.ErrLinkNotFound) {
			return nil, fmt.Errorf(
				"usecase.GetLinkUser: link don't exists: %w", err)
		}
		return nil, fmt.Errorf(
			"usecase.GetLinkUser: repository.GetLink unhandled error: %w", err)
	}

	// get the user that created the link
	user, err := s.repo.GetUser(link.Username)
	if err != nil {
		return nil, fmt.Errorf(
			"usecase.GetLinkUser: repository.GetUser unhandled error: %w", err)
	}

	// omit password from serialization
	user.Password = ""
	return user, nil
}

func (s *serviceUseCases) CreateUser(user *domain.User) error {
	// check username is unique
	_, err := s.repo.GetUser(user.Username)
	if err == nil {
		return fmt.Errorf(
			"usecase.CreateUser: username already taken: %w",
			domain_errors.ErrUsernameTaken,
		)
	} else if !errors.Is(err, domain_errors.ErrUserNotFound) {
		return fmt.Errorf(
			"usecase.CreateUser: repository.GetUser unhandled error: %w", err)
	}

	// create the user
	err = s.repo.CreateUser(user)
	if err != nil {
		return fmt.Errorf(
			"usecase.CreateUser: repository.CreateUser unhandled error: %w",
			err,
		)
	}

	return nil
}
