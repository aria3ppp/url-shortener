package domain_errors

import "errors"

var (
	ErrLinkNotFound        = errors.New("link not found")
	ErrUserNotFound        = errors.New("user not found")
	ErrUsernameTaken       = errors.New("username taken")
	ErrIncorrectPassword   = errors.New("incorrect password")
	ErrUsedShortenedString = errors.New("used shortened string")
)
