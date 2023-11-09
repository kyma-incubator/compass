/*
 * Copyright 2020 The Compass Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package auth

import (
	"context"
	"crypto/rsa"
	"fmt"
	"github.com/form3tech-oss/jwt-go"
	directorjwt "github.com/kyma-incubator/compass/components/director/pkg/jwt"
	"time"

	"github.com/pkg/errors"
)

const (
	// CustomerIDClaimKey is the key for a self-signed JWT claim indicating the customer ID for which the token belongs to
	CustomerIDClaimKey = "customerId"
)

// selfSignedJwtTokenAuthorizationProvider presents a AuthorizationProvider implementation which crafts a self-signed JWT bearer token values for the Authorization header
type selfSignedJwtTokenAuthorizationProvider struct {
	config directorjwt.Config
}

// NewSelfSignedJWTTokenAuthorizationProvider constructs an selfSignedJwtTokenAuthorizationProvider object
func NewSelfSignedJWTTokenAuthorizationProvider(config directorjwt.Config) *selfSignedJwtTokenAuthorizationProvider {
	return &selfSignedJwtTokenAuthorizationProvider{
		config: config,
	}
}

// Name specifies the name of the AuthorizationProvider
func (p selfSignedJwtTokenAuthorizationProvider) Name() string {
	return "SelfSignedTokenAuthorizationProvider"
}

// Matches contains the logic for matching the AuthorizationProvider
func (p selfSignedJwtTokenAuthorizationProvider) Matches(ctx context.Context) bool {
	credentials, err := LoadFromContext(ctx)
	if err != nil {
		return false
	}

	return credentials.Type() == SelfSignedTokenCredentialType
}

// GetAuthorization crafts a self-signed JWT bearer token to inject as part of the executing request
func (p selfSignedJwtTokenAuthorizationProvider) GetAuthorization(ctx context.Context) (string, error) {
	credentials, err := LoadFromContext(ctx)
	if err != nil {
		return "", err
	}

	selfSignedJWTCredentials, ok := credentials.Get().(*SelfSignedTokenCredentials)
	if !ok {
		return "", errors.New("failed to cast credentials to self-signed token credentials type")
	}

	pk, err := p.readPrivateKey(selfSignedJWTCredentials)
	if err != nil {
		return "", err
	}

	token, err := p.buildJWTToken(pk, selfSignedJWTCredentials.Claims)
	if err != nil {
		return "", err
	}

	fmt.Println("ALEX token: ", token)

	return "Bearer " + token, nil
}

func (p selfSignedJwtTokenAuthorizationProvider) readPrivateKey(selfSignedJWTCredentials *SelfSignedTokenCredentials) (*rsa.PrivateKey, error) {
	certsCache := selfSignedJWTCredentials.KeysCache.Get()[selfSignedJWTCredentials.JwtSelfSignCertSecretName]

	pk, ok := certsCache.PrivateKey.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("cannot parse private key variable")
	}

	return pk, nil
}

func (p selfSignedJwtTokenAuthorizationProvider) buildJWTToken(privateKey *rsa.PrivateKey, jwtClaims map[string]interface{}) (string, error) {
	token := jwt.New(jwt.SigningMethodRS256)
	claims := token.Claims.(jwt.MapClaims)
	for claimKey, claimValue := range jwtClaims {
		claims[claimKey] = claimValue
		fmt.Println("Claim", claimKey, claimValue)
	}

	claims["iat"] = time.Now().Unix()
	claims["exp"] = time.Now().Add(p.config.ExpireAfter).Unix()

	fmt.Println("ALEX buildJWTToken")

	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		return "", errors.Wrap(err, "while signing a jwt token")
	}

	return tokenString, nil
}
