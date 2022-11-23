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
	"os"

	"github.com/pkg/errors"
)

// DefaultServiceAccountTokenPath missing godoc
const DefaultServiceAccountTokenPath = "/var/run/secrets/kubernetes.io/serviceaccount/token"

// serviceAccountTokenAuthorizationProvider presents an AuthorizationProvider implementation which uses K8S Service Account tokens for the Authorization header
type serviceAccountTokenAuthorizationProvider struct {
	path string
}

// NewServiceAccountTokenAuthorizationProvider constructs an serviceAccountTokenAuthorizationProvider
func NewServiceAccountTokenAuthorizationProvider() *serviceAccountTokenAuthorizationProvider {
	return &serviceAccountTokenAuthorizationProvider{}
}

// NewServiceAccountTokenAuthorizationProviderWithPath constructs an serviceAccountTokenAuthorizationProvider with a given path
func NewServiceAccountTokenAuthorizationProviderWithPath(path string) *serviceAccountTokenAuthorizationProvider {
	return &serviceAccountTokenAuthorizationProvider{
		path: path,
	}
}

// Name specifies the name of the AuthorizationProvider
func (u serviceAccountTokenAuthorizationProvider) Name() string {
	return "ServiceAccountTokenAuthorizationProvider"
}

// Matches contains the logic for matching the AuthorizationProvider
func (u serviceAccountTokenAuthorizationProvider) Matches(ctx context.Context) bool {
	_, err := LoadFromContext(ctx)
	return err != nil
}

// GetAuthorization reads pod's service account token from the filesystem
func (u serviceAccountTokenAuthorizationProvider) GetAuthorization(_ context.Context) (string, error) {
	path := u.path
	if len(path) == 0 {
		path = DefaultServiceAccountTokenPath
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", errors.Wrapf(err, "Unable to read service account token file")
	}

	return "Bearer " + string(data), nil
}
