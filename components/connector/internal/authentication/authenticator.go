package authentication

import (
	"context"

	"github.com/pkg/errors"
)

//go:generate mockery -name=Authenticator
type Authenticator interface {
	AuthenticateToken(context context.Context) (string, error)
	Authenticate(context context.Context) (string, error)
	AuthenticateCertificate(context context.Context) (string, string, error)
}

func NewAuthenticator() Authenticator {
	return &authenticator{}
}

type authenticator struct {
}

func (a *authenticator) Authenticate(context context.Context) (string, error) {
	clientId, tokenAuthErr := a.AuthenticateToken(context)
	if tokenAuthErr == nil {
		return clientId, nil
	}

	clientId, _, certAuthErr := a.AuthenticateCertificate(context)
	if certAuthErr != nil {
		return "", errors.Errorf("Failed to authenticate request. Token authentication error: %s. Certificate authentication error: %s",
			tokenAuthErr.Error(), certAuthErr.Error())
	}

	return clientId, nil
}

func (a *authenticator) AuthenticateToken(context context.Context) (string, error) {
	clientId, err := GetStringFromContext(context, ClientIdFromTokenKey)
	if err != nil {
		return "", errors.Wrap(err, "Failed to authenticate request, token not provided")
	}

	if clientId == "" {
		return "", errors.New("Failed to authenticate with one time token.")
	}

	return clientId, nil
}

func (a *authenticator) AuthenticateCertificate(context context.Context) (string, string, error) {
	clientId, err := GetStringFromContext(context, ClientIdFromCertificateKey)
	if err != nil {
		return "", "", errors.Wrap(err, "Failed to authenticate with Certificate. Invalid subject.")
	}

	if clientId == "" {
		return "", "", errors.New("Failed to authenticate with Certificate. Invalid subject.")
	}

	certificateHash, err := GetStringFromContext(context, ClientCertificateHashKey)
	if err != nil {
		return "", "", errors.Wrap(err, "Failed to authenticate with Certificate. Invalid certificate hash.")
	}

	return clientId, certificateHash, nil
}
