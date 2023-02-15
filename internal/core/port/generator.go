package port

//go:generate mockgen -package mockups -destination mockups/mock_generator.go . RandomStringGenerator

type RandomStringGenerator interface {
	RandomString() string
}
