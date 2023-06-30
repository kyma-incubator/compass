package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/kyma-incubator/compass/components/kyma-adapter/internal/tenant_mapping_request"

	"github.com/go-openapi/runtime/middleware/header"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/httputils"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	directorGqlClient "github.com/kyma-incubator/compass/components/kyma-adapter/internal/director_gql_client"
	"github.com/pkg/errors"
)

const (
	assignOperation   = "assign"
	unassignOperation = "unassign"

	readyState         string = "READY"
	configPendingState string = "CONFIG_PENDING"
	createErrorState   string = "CREATE_ERROR"
	deleteErrorState   string = "DELETE_ERROR"
)

// AdapterHandler processes received requests
type AdapterHandler struct {
	DirectorGqlClient directorGqlClient.Client
}

// HandlerFunc is the implementation of AdapterHandler
func (a AdapterHandler) HandlerFunc(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	log.C(ctx).Info("The Kyma Tenant Mapping Adapter was hit.")

	var reqBody tenant_mapping_request.Body
	if err := decodeJSONBody(w, r, &reqBody); err != nil {
		var mr *malformedRequest
		if errors.As(err, &mr) {
			respondWithError(ctx, w, mr.status, "", mr)
		} else {
			respondWithError(ctx, w, http.StatusInternalServerError, "", errors.Wrap(err, "while decoding json request body"))
		}
		return
	}

	log.C(ctx).Info("Validating tenant mapping request body...")
	if err := reqBody.Validate(); err != nil {
		respondWithError(ctx, w, http.StatusBadRequest, "", errors.Wrapf(err, "while validating the request body"))
		return
	}
	tenantId := reqBody.ReceiverTenant.OwnerTenant
	log.C(ctx).Infof("The request has tenant with id %q", tenantId)

	if reqBody.Context.Operation == assignOperation {
		log.C(ctx).Infof("The request operation is %q", assignOperation)
		a.processAssignOperation(ctx, w, reqBody, tenantId)
	} else {
		log.C(ctx).Infof("The request operation is %q", unassignOperation)
		a.processUnassignOperation(ctx, w, reqBody, tenantId)
	}

	log.C(ctx).Info("The Kyma integration was successfully processed")
	return
}

func (a AdapterHandler) processAssignOperation(ctx context.Context, w http.ResponseWriter, reqBody tenant_mapping_request.Body, tenant string) {
	if reqBody.GetApplicationConfiguration() == (tenant_mapping_request.Configuration{}) { // config is missing
		respondWithSuccess(ctx, w, http.StatusOK, configPendingState, fmt.Sprintf("Configuration is missing. Responding with %q state...", configPendingState))
		return
	}

	appId := reqBody.GetApplicationId()
	rtmId := reqBody.GetRuntimeId()

	log.C(ctx).Infof("Getting application bundles for app with id %q and tenant %q", appId, tenant)
	bundles, err := a.DirectorGqlClient.GetApplicationBundles(ctx, appId, tenant)
	if err != nil {
		respondWithError(ctx, w, http.StatusBadRequest, createErrorState, errors.Wrapf(err, "while getting application bundles")) // check which status code to be
		return
	}

	if len(bundles) == 0 {
		respondWithSuccess(ctx, w, http.StatusOK, readyState, fmt.Sprintf("There are no bundles for application with ID %q", appId))
		return
	}

	// Decide if it is Assign or Resync operation
	instanceAuthExist := false
	for _, bundle := range bundles {
		for _, instanceAuth := range bundle.InstanceAuths {
			if *instanceAuth.RuntimeID == rtmId {
				instanceAuthExist = true
				break
			}
		}
		break
	}

	if instanceAuthExist { // resync case
		a.processAuthRotation(ctx, w, bundles, reqBody, tenant)
	} else { // assign case
		a.processAuthCreation(ctx, w, bundles, reqBody, tenant)
	}
}

func (a AdapterHandler) processUnassignOperation(ctx context.Context, w http.ResponseWriter, reqBody tenant_mapping_request.Body, tenant string) {
	appId := reqBody.GetApplicationId()
	rtmId := reqBody.GetRuntimeId()

	bundles, err := a.DirectorGqlClient.GetApplicationBundles(ctx, appId, tenant)
	if err != nil {
		respondWithError(ctx, w, http.StatusBadRequest, deleteErrorState, errors.Wrapf(err, "while getting application bundles")) // check which status code to be
		return
	}

	if len(bundles) == 0 {
		respondWithSuccess(ctx, w, http.StatusOK, readyState, fmt.Sprintf("There are no bundles for application with ID %q", reqBody.GetApplicationId()))
		return
	}

	instanceAuthExist := false
	for _, bundle := range bundles {
		for _, instanceAuth := range bundle.InstanceAuths {
			if *instanceAuth.RuntimeID == rtmId {
				instanceAuthExist = true

				if err = a.DirectorGqlClient.DeleteBundleInstanceAuth(ctx, tenant, instanceAuth.ID); err != nil {
					respondWithError(ctx, w, http.StatusBadRequest, deleteErrorState, errors.Wrapf(err, fmt.Sprintf("while deleting bundle instance auth with id: %q", instanceAuth.ID))) // check which status code to be
					return
				}
			}
		}
	}

	if !instanceAuthExist {
		respondWithSuccess(ctx, w, http.StatusOK, readyState, fmt.Sprintf("There are no bundle instance auths for deletion for runtime with ID %q and application with id %q", rtmId, appId))
		return
	}

	respondWithSuccess(ctx, w, http.StatusOK, readyState, fmt.Sprintf("Successfully deleted the bundle instance auths for runtime with ID %q and application with id %q", rtmId, appId))
}

func (a AdapterHandler) processAuthCreation(ctx context.Context, w http.ResponseWriter, bundles []*graphql.BundleExt, reqBody tenant_mapping_request.Body, tenant string) {
	rtmId := reqBody.GetRuntimeId()
	basicCreds := reqBody.GetBasicCredentials()
	oauthCreds := reqBody.GetOauthCredentials()

	for _, bundle := range bundles {
		if basicCreds.Username != "" { // basic creds
			if err := a.DirectorGqlClient.CreateBasicBundleInstanceAuth(ctx, tenant, bundle.ID, rtmId, basicCreds.Username, basicCreds.Password); err != nil {
				respondWithError(ctx, w, http.StatusBadRequest, createErrorState, errors.Wrapf(err, fmt.Sprintf("while creating bundle instance auth for bundle with id: %q", bundle.ID))) // check which status code to be
				return
			}
		} else { // oauth creds
			if err := a.DirectorGqlClient.CreateOauthBundleInstanceAuth(ctx, tenant, bundle.ID, rtmId, oauthCreds.TokenServiceUrl, oauthCreds.ClientId, oauthCreds.ClientSecret); err != nil {
				respondWithError(ctx, w, http.StatusBadRequest, createErrorState, errors.Wrapf(err, fmt.Sprintf("while creating bundle instance auth for bundle with id: %q", bundle.ID))) // check which status code to be
				return
			}
		}
	}

	respondWithSuccess(ctx, w, http.StatusOK, readyState, fmt.Sprintf("Successfully created the bundle instance auths for runtime with ID %q", rtmId))
}

func (a AdapterHandler) processAuthRotation(ctx context.Context, w http.ResponseWriter, bundles []*graphql.BundleExt, reqBody tenant_mapping_request.Body, tenant string) {
	rtmId := reqBody.GetRuntimeId()
	basicCreds := reqBody.GetBasicCredentials()
	oauthCreds := reqBody.GetOauthCredentials()

	for _, bundle := range bundles {
		for _, instanceAuth := range bundle.InstanceAuths {
			if *instanceAuth.RuntimeID == rtmId {
				if basicCreds.Username != "" { // basic creds
					if err := a.DirectorGqlClient.UpdateBasicBundleInstanceAuth(ctx, tenant, instanceAuth.ID, bundle.ID, basicCreds.Username, basicCreds.Password); err != nil {
						respondWithError(ctx, w, http.StatusBadRequest, createErrorState, errors.Wrapf(err, fmt.Sprintf("while updating bundle instance auth with id: %q", instanceAuth.ID))) // check which status code to be
						return
					}
				} else { // oauth creds
					if err := a.DirectorGqlClient.UpdateOauthBundleInstanceAuth(ctx, tenant, instanceAuth.ID, bundle.ID, oauthCreds.TokenServiceUrl, oauthCreds.ClientId, oauthCreds.ClientSecret); err != nil {
						respondWithError(ctx, w, http.StatusBadRequest, createErrorState, errors.Wrapf(err, fmt.Sprintf("while updating bundle instance auth with id: %q", instanceAuth.ID))) // check which status code to be
						return
					}
				}
			}
		}
	}

	respondWithSuccess(ctx, w, http.StatusOK, readyState, fmt.Sprintf("Successfully updated the bundle instance auths for runtime with ID %q", rtmId))
}

type malformedRequest struct {
	status int
	msg    string
}

func (mr *malformedRequest) Error() string {
	return mr.msg
}

type SuccessResponse struct {
	State string `json:"state,omitempty"`
}

// ErrorResponse structure used for the JSON encoded response
type ErrorResponse struct {
	State   string `json:"state,omitempty"`
	Message string `json:"error"`
}

func decodeJSONBody(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	if r.Header.Get(httputils.HeaderContentTypeKey) != "" {
		if value, _ := header.ParseValueAndParams(r.Header, httputils.HeaderContentTypeKey); value != httputils.ContentTypeApplicationJSON {
			return &malformedRequest{status: http.StatusUnsupportedMediaType, msg: "Content-Type header is not application/json"}
		}
	}

	// Use http.MaxBytesReader to enforce a maximum read of 1MB from the
	// response body. A request body larger than that will now result in
	// Decode() returning a "http: request body too large" error.
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	dec := json.NewDecoder(r.Body)

	if err := dec.Decode(&dst); err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError

		switch {
		case errors.As(err, &syntaxError):
			msg := fmt.Sprintf("Request body contains badly-formed JSON (at position %d)", syntaxError.Offset)
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case errors.Is(err, io.ErrUnexpectedEOF):
			return &malformedRequest{status: http.StatusBadRequest, msg: "Request body contains badly-formed JSON"}

		case errors.As(err, &unmarshalTypeError):
			msg := fmt.Sprintf("Request body contains an invalid value for the %q field (at position %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset)
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			msg := fmt.Sprintf("Request body contains unknown field %s", fieldName)
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case errors.Is(err, io.EOF):
			return &malformedRequest{status: http.StatusBadRequest, msg: "Request body must not be empty"}

		case err.Error() == "http: request body too large":
			return &malformedRequest{status: http.StatusRequestEntityTooLarge, msg: "Request body must not be larger than 1MB"}

		default:
			return err
		}
	}

	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return &malformedRequest{status: http.StatusBadRequest, msg: "Request body must only contain a single JSON object"}
	}

	return nil
}

// respondWithError writes a http response using with the JSON error wrapped in an ErrorResponse struct
func respondWithError(ctx context.Context, w http.ResponseWriter, status int, state string, err error) {
	log.C(ctx).Error(err.Error())
	w.Header().Add(httputils.HeaderContentTypeKey, httputils.ContentTypeApplicationJSON)
	w.WriteHeader(status)
	errorResponse := ErrorResponse{State: state, Message: err.Error()}
	httputils.RespondWithBody(ctx, w, status, errorResponse)
}

// respondWithSuccess writes a http response using with the JSON success wrapped in an SuccessResponse struct
func respondWithSuccess(ctx context.Context, w http.ResponseWriter, status int, state, msg string) {
	log.C(ctx).Info(msg)
	w.Header().Add(httputils.HeaderContentTypeKey, httputils.ContentTypeApplicationJSON)
	w.WriteHeader(status)
	successResponse := SuccessResponse{State: state}
	httputils.RespondWithBody(ctx, w, status, successResponse)
}
