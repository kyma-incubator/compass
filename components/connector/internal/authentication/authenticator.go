package authentication

import (
	"context"

	"github.com/kyma-incubator/compass/components/connector/internal/tokens"
	"github.com/pkg/errors"
)

//go:generate mockery -name=Authenticator
type Authenticator interface {
	AuthenticateToken(context context.Context) (tokens.TokenData, error)
	AuthenticateCertificate(context context.Context) (CertificateData, error)
	AuthenticateTokenOrCertificate(context context.Context) (string, error)
}

func NewAuthenticator(tokenService tokens.Service) Authenticator {
	return &authenticator{
		tokenService: tokenService,
	}
}

type authenticator struct {
	tokenService tokens.Service
}

func (a *authenticator) AuthenticateTokenOrCertificate(context context.Context) (string, error) {
	tokenData, tokenAuthErr := a.AuthenticateToken(context)
	if tokenAuthErr == nil {
		return tokenData.ClientId, nil
	}

	certData, certAuthErr := a.AuthenticateCertificate(context)
	if certAuthErr != nil {
		return "", errors.Errorf("Failed to authenticate request. Token authentication error: %s. Certificate authentication error: %s",
			tokenAuthErr.Error(), certAuthErr.Error())
	}

	return certData.CommonName, nil
}

func (a *authenticator) AuthenticateToken(context context.Context) (tokens.TokenData, error) {
	token, err := GetStringFromContext(context, ConnectorTokenKey)
	if err != nil {
		return tokens.TokenData{}, errors.Wrap(err, "Failed to authenticate request, token not provided")
	}

	tokenData, err := a.tokenService.Resolve(token)
	if err != nil {
		return tokens.TokenData{}, errors.Wrap(err, "Failed to authenticate request, token is invalid")
	}

	return tokenData, nil
}

func (a *authenticator) AuthenticateCertificate(context context.Context) (CertificateData, error) {
	commonName, err := GetStringFromContext(context, CertificateCommonNameKey)
	if err != nil {
		return CertificateData{}, errors.Wrap(err, "Failed to authenticate request, no valid Common Name found")
	}

	hash, err := GetStringFromContext(context, CertificateHashKey)
	if err != nil {
		return CertificateData{}, errors.Wrap(err, "Failed to authenticate request, no certificate hash found")
	}

	// TODO: here we should check if certificate is revoked

	return CertificateData{
		Hash:       hash,
		CommonName: commonName,
	}, nil
}
