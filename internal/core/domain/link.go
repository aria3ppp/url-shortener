package domain

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type Link struct {
	ShortenedString string `json:"shortened_string"` // unique
	URL             string `json:"url"`
	Username        string `json:"username"`
}

var _ validation.Validatable = Link{}

func (r Link) Validate() error {
	return validation.ValidateStruct(
		&r,
		validation.Field(
			&r.URL,
			validation.Required,
			is.URL,
		),
		validation.Field(
			&r.ShortenedString,
			is.Alphanumeric,
		),
	)
}
