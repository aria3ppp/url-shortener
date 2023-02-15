package generator

import "math/rand"

type generator struct {
	length int
}

func NewRandomStringGenerator(length int) generator {
	if length < 6 || length > 32 {
		panic("generator: length should not be less than 6 or greater than 32")
	}
	return generator{length: length}
}

func (g generator) RandomString() string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, g.length)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
