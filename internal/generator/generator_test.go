package generator_test

import (
	"regexp"
	"testing"

	"github.com/aria3ppp/url-shortener/internal/generator"
	"github.com/stretchr/testify/require"
)

var alphanumericRegexp = regexp.MustCompile("^[a-zA-Z0-9]+$")

func TestRandomString(t *testing.T) {
	type fields struct {
		length int
	}
	tests := []struct {
		name   string
		fields fields
		panics bool
	}{
		{
			name: "tc1",
			fields: fields{
				length: 6,
			},
		},
		{
			name: "tc2",
			fields: fields{
				length: 32,
			},
		},
		{
			name: "tc3",
			fields: fields{
				length: 5,
			},
			panics: true,
		},
		{
			name: "tc4",
			fields: fields{
				length: 33,
			},
			panics: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)

			if tt.panics {
				require.PanicsWithValue(
					"generator: length should not be less than 6 or greater than 32",
					func() {
						g := generator.NewRandomStringGenerator(
							tt.fields.length,
						)
						g.RandomString()
					},
				)
			} else {
				g := generator.NewRandomStringGenerator(tt.fields.length)
				randomString := g.RandomString()

				require.Len(randomString, tt.fields.length)
				require.Regexp(alphanumericRegexp, randomString)
			}
		})
	}
}
