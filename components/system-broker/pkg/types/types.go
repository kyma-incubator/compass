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

package types

import (
	"context"

	"github.com/kyma-incubator/compass/components/system-broker/internal/director"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . applicationsLister
type ApplicationLister interface {
	FetchApplication(ctx context.Context, id string) (*director.ApplicationOutput, error)
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . applicationsLister
type ApplicationsLister interface {
	FetchApplications(ctx context.Context) (*director.ApplicationsOutput, error)
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . bundleCredentialsFetcher
type BundleCredentialsFetcher interface {
	FetchBundleInstanceAuth(ctx context.Context, in *director.BundleInstanceInput) (*director.BundleInstanceAuthOutput, error)
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . bundleCredentialsFetcherForInstance
type BundleCredentialsFetcherForInstance interface {
	FetchBundleInstanceCredentials(ctx context.Context, in *director.BundleInstanceInput) (*director.BundleInstanceCredentialsOutput, error)
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . bundleCredentialsCreateRequester
type BundleCredentialsCreateRequester interface {
	RequestBundleInstanceCredentialsCreation(ctx context.Context, in *director.BundleInstanceCredentialsInput) (*director.BundleInstanceAuthOutput, error)
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . bundleCredentialsDeleteRequester
type BundleCredentialsDeleteRequester interface {
	RequestBundleInstanceCredentialsDeletion(ctx context.Context, in *director.BundleInstanceAuthDeletionInput) (*director.BundleInstanceAuthDeletionOutput, error)
}
