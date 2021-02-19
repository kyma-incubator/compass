package onetimetoken

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

type TokenGenerator interface {
	NewToken() (string, error)
}

type tokenGenerator struct {
	tokenLength int
}

func NewTokenGenerator(tokenLength int) TokenGenerator {
	return &tokenGenerator{tokenLength: tokenLength}
}

func (tg *tokenGenerator) NewToken() (string, error) {
	return generateRandomString(tg.tokenLength)
}

func generateRandomBytes(number int) ([]byte, error) {
	bytes := make([]byte, number)
	_, err := rand.Read(bytes)
	if err != nil {
		return nil, fmt.Errorf("Failed to generate random bytes: %s", err)
	}

	return bytes, nil
}

func generateRandomString(length int) (string, error) {
	bytes, err := generateRandomBytes(length)
	return base64.URLEncoding.EncodeToString(bytes), err
}
