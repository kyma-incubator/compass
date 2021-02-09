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

// TODO: Delete after bundles are adopted
package packagetobundles

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"

	"github.com/pkg/errors"

	gqlgen "github.com/99designs/gqlgen/graphql"
	"github.com/kyma-incubator/compass/components/director/internal/consumer"
	"github.com/kyma-incubator/compass/components/director/internal/domain/label"
	"github.com/kyma-incubator/compass/components/director/internal/domain/labeldef"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/uid"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/vektah/gqlparser/gqlerror"
)

const useBundlesParam = "useBundles"

//go:generate mockery -name=LabelUpsertService -output=automock -outpkg=automock -case=underscore
type LabelUpsertService interface {
	UpsertLabel(ctx context.Context, tenant string, labelInput *model.LabelInput) error
}

type Handler struct {
	transact           persistence.Transactioner
	labelUpsertService LabelUpsertService
}

func NewHandler(transact persistence.Transactioner) *Handler {
	labelRepo := label.NewRepository(label.NewConverter())
	labelDefRepo := labeldef.NewRepository(labeldef.NewConverter())

	uidSvc := uid.NewService()
	labelUpsertSvc := label.NewLabelUpsertService(labelRepo, labelDefRepo, uidSvc)

	return &Handler{
		transact:           transact,
		labelUpsertService: labelUpsertSvc,
	}
}

func (h *Handler) Handler() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			useBundles := r.URL.Query().Get(useBundlesParam)
			if useBundles == "true" {
				consumerInfo, err := consumer.LoadFromContext(ctx)
				if err != nil {
					log.C(ctx).WithError(err).Error("Error determining request consumer")
					appErr := apperrors.InternalErrorFrom(err, "while determining request consumer")
					writeAppError(ctx, w, appErr, http.StatusInternalServerError)
					return
				}

				log.C(ctx).Infof("Will proceed without rewriting the request body. Bundles are adopted for consumer with ID %q and type %q", consumerInfo.ConsumerID, consumerInfo.ConsumerType)

				next.ServeHTTP(w, r)

				if consumerInfo.ConsumerType == consumer.Runtime {
					if err := h.labelRuntimeWithBundlesParam(ctx, consumerInfo); err != nil {
						log.C(ctx).WithError(err).Errorf("Error labelling runtime with %q", useBundlesParam)
					}
				}

				return
			}

			log.C(ctx).Info("Will rewrite the request body. Bundles are still not adopted")

			recorder := httptest.NewRecorder()

			reqBody, err := ioutil.ReadAll(r.Body)
			if err != nil {
				log.C(ctx).WithError(err).Error("Error reading request body")
				appErr := apperrors.InternalErrorFrom(err, "while reading request body")
				writeAppError(ctx, w, appErr, http.StatusInternalServerError)
				return
			}

			body := string(reqBody)
			body = strings.ReplaceAll(body, "\\n", "") // removes unnecessary complexity from the next regexes

			// rewrite Query/Mutation names
			body = strings.ReplaceAll(body, "packageByInstanceAuth", "bundleByInstanceAuth")
			body = strings.ReplaceAll(body, "packageInstanceAuth", "bundleInstanceAuth")
			body = strings.ReplaceAll(body, "addPackage", "addBundle")
			body = strings.ReplaceAll(body, "updatePackage", "updateBundle")
			body = strings.ReplaceAll(body, "deletePackage", "deleteBundle")
			body = strings.ReplaceAll(body, "addAPIDefinitionToPackage", "addAPIDefinitionToBundle")
			body = strings.ReplaceAll(body, "addEventDefinitionToPackage", "addEventDefinitionToBundle")
			body = strings.ReplaceAll(body, "addDocumentToPackage", "addDocumentToBundle")
			body = strings.ReplaceAll(body, "setPackageInstanceAuth", "setBundleInstanceAuth")
			body = strings.ReplaceAll(body, "deletePackageInstanceAuth", "deleteBundleInstanceAuth")
			body = strings.ReplaceAll(body, "requestPackageInstanceAuthCreation", "requestBundleInstanceAuthCreation")
			body = strings.ReplaceAll(body, "requestPackageInstanceAuthDeletion", "requestBundleInstanceAuthDeletion")

			// rewrite Query/Mutation arguments
			body = strings.ReplaceAll(body, "packageID", "bundleID")
			body = strings.ReplaceAll(body, "PackageCreateInput", "BundleCreateInput")
			body = strings.ReplaceAll(body, "PackageInstanceAuthRequestInput", "BundleInstanceAuthRequestInput")
			body = strings.ReplaceAll(body, "PackageInstanceAuthSetInput", "BundleInstanceAuthSetInput")
			body = strings.ReplaceAll(body, "PackageInstanceAuthStatusInput", "BundleInstanceAuthStatusInput")
			body = strings.ReplaceAll(body, "PackageUpdateInput", "BundleUpdateInput")

			// rewrite JSON input
			reqPackagesJSONPattern := regexp.MustCompile(`(\s*)packages(\s*:\s*\[)`) // matches ` packages:  [`
			body = reqPackagesJSONPattern.ReplaceAllString(body, "${1}bundles${2}")

			// rewrite GQL output
			reqPackagesGraphQLPattern := regexp.MustCompile(`(\s*)packages(\s*\{)`) // matches ` packages {`
			body = reqPackagesGraphQLPattern.ReplaceAllString(body, "${1}bundles${2}")

			reqPackageGraphQLPattern := regexp.MustCompile(`(\s*)package(\s*\(\s*id\s*:\s*)`) // matches ` package ( id : `
			body = reqPackageGraphQLPattern.ReplaceAllString(body, "${1}bundle${2}")

			reqPackageModeGraphQLPattern := regexp.MustCompile(`(\s*)mode(\s*):(\s*)PACKAGE(\s*)`) // matches ` mode: PACKAGE `
			body = reqPackageModeGraphQLPattern.ReplaceAllString(body, "${1}mode${2}:${3}BUNDLE${4}")

			r.Body = ioutil.NopCloser(strings.NewReader(body))
			r.ContentLength = int64(len(body))

			next.ServeHTTP(recorder, r)

			for key, values := range recorder.Header() {
				for _, v := range values {
					w.Header().Add(key, v)
				}
			}

			respBody, err := ioutil.ReadAll(recorder.Body)
			if err != nil {
				log.C(ctx).WithError(err).Error("Error reading response body")
				appErr := apperrors.InternalErrorFrom(err, "while reading response body")
				writeAppError(ctx, w, appErr, http.StatusInternalServerError)
				return
			}

			body = string(respBody)

			// rewrite Query/Mutation names
			body = strings.ReplaceAll(body, "bundleByInstanceAuth", "packageByInstanceAuth")
			body = strings.ReplaceAll(body, "bundleInstanceAuth", "packageInstanceAuth")
			body = strings.ReplaceAll(body, "addBundle", "addPackage")
			body = strings.ReplaceAll(body, "updateBundle", "updatePackage")
			body = strings.ReplaceAll(body, "deleteBundle", "deletePackage")
			body = strings.ReplaceAll(body, "addAPIDefinitionToBundle", "addAPIDefinitionToPackage")
			body = strings.ReplaceAll(body, "addEventDefinitionToBundle", "addEventDefinitionToPackage")
			body = strings.ReplaceAll(body, "addDocumentToBundle", "addDocumentToPackage")
			body = strings.ReplaceAll(body, "setBundleInstanceAuth", "setPackageInstanceAuth")
			body = strings.ReplaceAll(body, "deleteBundleInstanceAuth", "deletePackageInstanceAuth")
			body = strings.ReplaceAll(body, "requestBundleInstanceAuthCreation", "requestPackageInstanceAuthCreation")
			body = strings.ReplaceAll(body, "requestBundleInstanceAuthDeletion", "requestPackageInstanceAuthDeletion")

			respPackagesJSONPattern := regexp.MustCompile(`(\s*\")bundles(\"\s*:\s*\{)`) // matches ` "bundles":  {`
			body = respPackagesJSONPattern.ReplaceAllString(body, "${1}packages${2}")

			respPackageJSONPattern := regexp.MustCompile(`(\s*\")bundle(\"\s*:\s*\{)`) // matches ` "bundle":  {`
			body = respPackageJSONPattern.ReplaceAllString(body, "${1}package${2}")

			respPackageModeGraphQLPattern := regexp.MustCompile(`(\s*\")mode(\"\s*):(\s*\")BUNDLE(\"\s*)`) // matches ` "mode": "BUNDLE" `
			body = respPackageModeGraphQLPattern.ReplaceAllString(body, "${1}mode${2}:${3}PACKAGE${4}")

			w.WriteHeader(recorder.Code)
			if _, err := w.Write([]byte(body)); err != nil {
				log.C(ctx).WithError(err).Error("Error writing response body")
				appErr := apperrors.InternalErrorFrom(err, "while writing response body")
				writeAppError(ctx, w, appErr, http.StatusInternalServerError)
				return
			}
		})
	}
}

func (h *Handler) labelRuntimeWithBundlesParam(ctx context.Context, consumerInfo consumer.Consumer) error {
	tenantID, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrap(err, "while determining request tenant")
	}

	tx, err := h.transact.Begin()
	if err != nil {
		return errors.Wrap(err, "while opening db transaction")
	}
	defer h.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	log.C(ctx).Infof("Proceeding with labeling runtime with ID %q with label %q", consumerInfo.ConsumerID, useBundlesParam)
	if err := h.labelUpsertService.UpsertLabel(ctx, tenantID, &model.LabelInput{
		Key:        useBundlesParam,
		Value:      "true",
		ObjectID:   consumerInfo.ConsumerID,
		ObjectType: model.LabelableObject(consumerInfo.ConsumerType),
	}); err != nil {
		return errors.Wrapf(err, "while upserting %q label", useBundlesParam)
	}

	if err = tx.Commit(); err != nil {
		return errors.Wrap(err, "while committing database transaction")
	}

	return nil
}

func writeAppError(ctx context.Context, w http.ResponseWriter, appErr error, statusCode int) {
	errCode := apperrors.ErrorCode(appErr)
	if errCode == apperrors.UnknownError || errCode == apperrors.InternalError {
		errCode = apperrors.InternalError
	}

	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")
	resp := gqlgen.Response{Errors: []*gqlerror.Error{{
		Message:    appErr.Error(),
		Extensions: map[string]interface{}{"error_code": errCode, "error": errCode.String()}}}}
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.C(ctx).WithError(err).Error("An error occurred while encoding data. ")
	}
}
