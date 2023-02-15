package usecase_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aria3ppp/url-shortener/internal/core/domain"
	domain_errors "github.com/aria3ppp/url-shortener/internal/core/errors"
	"github.com/aria3ppp/url-shortener/internal/core/port/mockups"
	"github.com/aria3ppp/url-shortener/internal/core/usecase"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

type mocks struct {
	repository *mockups.MockRepository
	generator  *mockups.MockRandomStringGenerator
}

func TestGetLink(t *testing.T) {
	type args struct {
		shortenedString string
	}
	type want struct {
		link *domain.Link
		err  error
	}

	tests := []struct {
		name string
		args args
		want want
		mock func(m mocks)
	}{
		{
			name: "link not found",
			args: args{shortenedString: "shortened_string"},
			want: want{
				link: nil,
				err: fmt.Errorf(
					"usecase.GetLink: link don't exists: %w",
					domain_errors.ErrLinkNotFound,
				),
			},
			mock: func(m mocks) {
				m.repository.EXPECT().
					GetLink("shortened_string").
					Return(nil, domain_errors.ErrLinkNotFound)
			},
		},
		{
			name: "GetLink unhandled error",
			args: args{shortenedString: "shortened_string"},
			want: want{
				link: nil,
				err: fmt.Errorf(
					"usecase.GetLink: repository.GetLink unhandled error: %w",
					errors.New("GetLink_unhandled_error"),
				),
			},
			mock: func(m mocks) {
				m.repository.EXPECT().
					GetLink("shortened_string").
					Return(nil, errors.New("GetLink_unhandled_error"))
			},
		},
		{
			name: "ok",
			args: args{shortenedString: "shortened_string"},
			want: want{
				link: &domain.Link{
					ShortenedString: "shortened_string",
					URL:             "url",
					Username:        "username",
				},
				err: nil,
			},
			mock: func(m mocks) {
				m.repository.EXPECT().
					GetLink("shortened_string").
					Return(
						&domain.Link{
							ShortenedString: "shortened_string",
							URL:             "url",
							Username:        "username",
						},
						nil,
					)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)

			controller := gomock.NewController(t)
			m := mocks{
				repository: mockups.NewMockRepository(controller),
				generator:  mockups.NewMockRandomStringGenerator(controller),
			}

			tt.mock(m)
			service := usecase.NewService(m.repository, m.generator)

			link, err := service.GetLink(tt.args.shortenedString)

			require.Equal(tt.want.err, err)
			require.Equal(tt.want.link, link)
		})
	}
}

func TestCreateLink(t *testing.T) {
	type args struct {
		url             string
		shortenedString string
		user            *domain.User
	}
	type want struct {
		link *domain.Link
		err  error
	}

	tests := []struct {
		name string
		args args
		want want
		mock func(m mocks)
	}{
		{
			name: "user not found",
			args: args{
				url: "url",
				user: &domain.User{
					Username: "username",
					Password: "password",
				},
			},
			want: want{
				link: nil,
				err: fmt.Errorf(
					"usecase.CreateLink: user don't exists: %w",
					domain_errors.ErrUserNotFound,
				),
			},
			mock: func(m mocks) {
				m.repository.EXPECT().
					GetUser("username").
					Return(nil, domain_errors.ErrUserNotFound)
			},
		},
		{
			name: "GetUser unhandled error",
			args: args{
				url: "url",
				user: &domain.User{
					Username: "username",
					Password: "password",
				},
			},
			want: want{
				link: nil,
				err: fmt.Errorf(
					"usecase.CreateLink: repository.GetUser unhandled error: %w",
					errors.New("GetUser_unhandled_error"),
				),
			},
			mock: func(m mocks) {
				m.repository.EXPECT().
					GetUser("username").
					Return(nil, errors.New("GetUser_unhandled_error"))
			},
		},
		{
			name: "incorrect password",
			args: args{
				url: "url",
				user: &domain.User{
					Username: "username",
					Password: "password",
				},
			},
			want: want{
				link: nil,
				err: fmt.Errorf(
					"usecase.CreateLink: user password don't match: %w",
					domain_errors.ErrIncorrectPassword,
				),
			},
			mock: func(m mocks) {
				m.repository.EXPECT().
					GetUser("username").
					Return(
						&domain.User{
							Username: "username",
							Password: "unmatched_password",
						},
						nil,
					)
			},
		},
		{
			name: "used shortened string",
			args: args{
				url:             "url",
				shortenedString: "used_shortened_string",
				user: &domain.User{
					Username: "username",
					Password: "password",
				},
			},
			want: want{
				link: nil,
				err: fmt.Errorf(
					"usecase.CreateLink: user given shortened string already used: %w",
					domain_errors.ErrUsedShortenedString,
				),
			},
			mock: func(m mocks) {
				getUserCall := m.repository.EXPECT().
					GetUser("username").
					Return(
						&domain.User{
							Username: "username",
							Password: "password",
						},
						nil,
					)

				m.repository.EXPECT().
					GetLink("used_shortened_string").
					Return(
						&domain.Link{
							ShortenedString: "used_shortened_string",
							URL:             "url",
							Username:        "perhaps_or_perhaps_not_another_user",
						},
						nil,
					).
					After(getUserCall)
			},
		},
		{
			name: "GetLink unhandled error",
			args: args{
				url:             "url",
				shortenedString: "shortened_string",
				user: &domain.User{
					Username: "username",
					Password: "password",
				},
			},
			want: want{
				link: nil,
				err: fmt.Errorf(
					"usecase.CreateLink: repository.GetLink unhandled error: %w",
					errors.New("GetLink_unhandled_error"),
				),
			},
			mock: func(m mocks) {
				getUserCall := m.repository.EXPECT().
					GetUser("username").
					Return(
						&domain.User{
							Username: "username",
							Password: "password",
						},
						nil,
					)

				m.repository.EXPECT().
					GetLink("shortened_string").
					Return(nil, errors.New("GetLink_unhandled_error")).
					After(getUserCall)
			},
		},
		{
			name: "CreateLink unhandled error",
			args: args{
				url: "url",
				user: &domain.User{
					Username: "username",
					Password: "password",
				},
			},
			want: want{
				link: nil,
				err: fmt.Errorf(
					"usecase.CreateLink: repository.CreateLink unhandled error: %w",
					errors.New("CreateLink_unhandled_error"),
				),
			},
			mock: func(m mocks) {
				getUserCall := m.repository.EXPECT().
					GetUser("username").
					Return(
						&domain.User{
							Username: "username",
							Password: "password",
						},
						nil,
					)

				generateRandomString := m.generator.EXPECT().
					RandomString().
					Return("random_shortened_string").
					After(getUserCall)

				m.repository.EXPECT().
					CreateLink(&domain.Link{
						ShortenedString: "random_shortened_string",
						URL:             "url",
						Username:        "username",
					}).
					Return(errors.New("CreateLink_unhandled_error")).
					After(generateRandomString)
			},
		},
		{
			name: "ok",
			args: args{
				url: "url",
				user: &domain.User{
					Username: "username",
					Password: "password",
				},
			},
			want: want{
				link: &domain.Link{
					ShortenedString: "random_shortened_string",
					URL:             "url",
					Username:        "username",
				},
				err: nil,
			},
			mock: func(m mocks) {
				getUserCall := m.repository.EXPECT().
					GetUser("username").
					Return(
						&domain.User{
							Username: "username",
							Password: "password",
						},
						nil,
					)

				generateRandomString := m.generator.EXPECT().
					RandomString().
					Return("random_shortened_string").
					After(getUserCall)

				m.repository.EXPECT().
					CreateLink(&domain.Link{
						ShortenedString: "random_shortened_string",
						URL:             "url",
						Username:        "username",
					}).
					Return(nil).
					After(generateRandomString)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)

			controller := gomock.NewController(t)
			m := mocks{
				repository: mockups.NewMockRepository(controller),
				generator:  mockups.NewMockRandomStringGenerator(controller),
			}

			tt.mock(m)
			service := usecase.NewService(m.repository, m.generator)

			link, err := service.CreateLink(
				tt.args.url,
				tt.args.shortenedString,
				tt.args.user,
			)

			require.Equal(tt.want.err, err)
			require.Equal(tt.want.link, link)
		})
	}
}

func TestGetLinkUser(t *testing.T) {
	type args struct {
		shortenedString string
	}
	type want struct {
		user *domain.User
		err  error
	}

	tests := []struct {
		name string
		args args
		want want
		mock func(m mocks)
	}{
		{
			name: "link not found",
			args: args{shortenedString: "shortened_string"},
			want: want{
				user: nil,
				err: fmt.Errorf(
					"usecase.GetLinkUser: link don't exists: %w",
					domain_errors.ErrLinkNotFound,
				),
			},
			mock: func(m mocks) {
				m.repository.EXPECT().
					GetLink("shortened_string").
					Return(nil, domain_errors.ErrLinkNotFound)
			},
		},
		{
			name: "GetLink unhandled error",
			args: args{shortenedString: "shortened_string"},
			want: want{
				user: nil,
				err: fmt.Errorf(
					"usecase.GetLinkUser: repository.GetLink unhandled error: %w",
					errors.New("GetLink_unhandled_error"),
				),
			},
			mock: func(m mocks) {
				m.repository.EXPECT().
					GetLink("shortened_string").
					Return(nil, errors.New("GetLink_unhandled_error"))
			},
		},
		{
			name: "GetUser unhandled error",
			args: args{shortenedString: "shortened_string"},
			want: want{
				user: nil,
				err: fmt.Errorf(
					"usecase.GetLinkUser: repository.GetUser unhandled error: %w",
					errors.New("GetUser_unhandled_error"),
				),
			},
			mock: func(m mocks) {
				getLinkCall := m.repository.EXPECT().
					GetLink("shortened_string").
					Return(
						&domain.Link{
							ShortenedString: "shortened_string",
							URL:             "url",
							Username:        "username",
						},
						nil,
					)

				m.repository.EXPECT().
					GetUser("username").
					Return(nil, errors.New("GetUser_unhandled_error")).
					After(getLinkCall)
			},
		},
		{
			name: "ok",
			args: args{shortenedString: "shortened_string"},
			want: want{
				user: &domain.User{
					Username: "username",
					Password: "", // omitted password from serialization
				},
				err: nil,
			},
			mock: func(m mocks) {
				getLinkCall := m.repository.EXPECT().
					GetLink("shortened_string").
					Return(
						&domain.Link{
							ShortenedString: "shortened_string",
							URL:             "url",
							Username:        "username",
						},
						nil,
					)

				m.repository.EXPECT().
					GetUser("username").
					Return(
						&domain.User{
							Username: "username",
							Password: "password",
						},
						nil,
					).
					After(getLinkCall)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)

			controller := gomock.NewController(t)
			m := mocks{
				repository: mockups.NewMockRepository(controller),
				generator:  mockups.NewMockRandomStringGenerator(controller),
			}

			tt.mock(m)
			service := usecase.NewService(m.repository, m.generator)

			link, err := service.GetLinkUser(tt.args.shortenedString)

			require.Equal(tt.want.err, err)
			require.Equal(tt.want.user, link)
		})
	}
}

func TestCreateUser(t *testing.T) {
	type args struct {
		user *domain.User
	}
	type want struct {
		err error
	}

	user := &domain.User{
		Username: "username",
		Password: "password",
	}

	tests := []struct {
		name string
		args args
		want want
		mock func(m mocks)
	}{
		{
			name: "username taken",
			args: args{user: user},
			want: want{
				err: fmt.Errorf(
					"usecase.CreateUser: username already taken: %w",
					domain_errors.ErrUsernameTaken,
				),
			},
			mock: func(m mocks) {
				m.repository.EXPECT().
					GetUser(user.Username).
					Return(
						&domain.User{
							Username: "username",
							Password: "perhaps_or_perhaps_not_another_password",
						},
						nil,
					)
			},
		},
		{
			name: "GetUser unhandled error",
			args: args{user: user},
			want: want{
				err: fmt.Errorf(
					"usecase.CreateUser: repository.GetUser unhandled error: %w",
					errors.New("GetUser_unhandled_error"),
				),
			},
			mock: func(m mocks) {
				m.repository.EXPECT().
					GetUser(user.Username).
					Return(nil, errors.New("GetUser_unhandled_error"))
			},
		},
		{
			name: "CreateUser unhandled error",
			args: args{user: user},
			want: want{
				err: fmt.Errorf(
					"usecase.CreateUser: repository.CreateUser unhandled error: %w",
					errors.New("CreateUser_unhandled_error"),
				),
			},
			mock: func(m mocks) {
				getUserCall := m.repository.EXPECT().
					GetUser(user.Username).
					Return(nil, domain_errors.ErrUserNotFound)

				m.repository.EXPECT().
					CreateUser(user).
					Return(errors.New("CreateUser_unhandled_error")).
					After(getUserCall)
			},
		},
		{
			name: "ok",
			args: args{user: user},
			want: want{
				err: nil,
			},
			mock: func(m mocks) {
				getUserCall := m.repository.EXPECT().
					GetUser(user.Username).
					Return(nil, domain_errors.ErrUserNotFound)

				m.repository.EXPECT().
					CreateUser(user).
					Return(nil).
					After(getUserCall)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)

			controller := gomock.NewController(t)
			m := mocks{
				repository: mockups.NewMockRepository(controller),
				generator:  mockups.NewMockRandomStringGenerator(controller),
			}

			tt.mock(m)
			service := usecase.NewService(m.repository, m.generator)

			err := service.CreateUser(tt.args.user)

			require.Equal(tt.want.err, err)
		})
	}
}
