package domain

import validation "github.com/go-ozzo/ozzo-validation/v4"

type User struct {
	Username string `json:"username"` // unique
	Password string `json:"password,omitempty"`
	// password is saved in plain text to simplify implementation. but in general saving plain text passwords is a bad practice.
	// also to omit password from encoding set this field to an empty string
}

var _ validation.Validatable = User{}

func (r User) Validate() error {
	return validation.ValidateStruct(
		&r,
		validation.Field(
			&r.Username,
			validation.Required,
			validation.Length(8, 40),
		),
		validation.Field(
			&r.Password,
			validation.Required,
			validation.Length(8, 40),
		),
	)
}
