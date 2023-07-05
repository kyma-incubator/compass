package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/kyma-incubator/compass/components/kyma-adapter/internal/types"

	"github.com/go-openapi/runtime/middleware/header"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/httputils"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
)

const (
	assignOperation   = "assign"
	unassignOperation = "unassign"

	readyState         string = "READY"
	configPendingState string = "CONFIG_PENDING"
	createErrorState   string = "CREATE_ERROR"
	deleteErrorState   string = "DELETE_ERROR"

	successStatusCode int = http.StatusOK
)

// Client is responsible for making internal graphql calls to the director
//go:generate mockery --name=Client --output=automock --outpkg=automock --case=underscore --disable-version-string
type Client interface {
	GetApplicationBundles(ctx context.Context, appID, tenant string) ([]*graphql.BundleExt, error)
	CreateBasicBundleInstanceAuth(ctx context.Context, tenant, bndlID, rtmID, username, password string) error
	CreateOauthBundleInstanceAuth(ctx context.Context, tenant, bndlID, rtmID, tokenServiceURL, clientID, clientSecret string) error
	UpdateBasicBundleInstanceAuth(ctx context.Context, tenant, authID, bndlID, username, password string) error
	UpdateOauthBundleInstanceAuth(ctx context.Context, tenant, authID, bndlID, tokenServiceURL, clientID, clientSecret string) error
	DeleteBundleInstanceAuth(ctx context.Context, tenant, authID string) error
}

// AdapterHandler is the Kyma Tenant Mapping Adapter handler which processes the received requests
type AdapterHandler struct {
	DirectorGqlClient Client
}

func NewHandler(directorGqlClient Client) *AdapterHandler {
	return &AdapterHandler{DirectorGqlClient: directorGqlClient}
}

// HandlerFunc is the implementation of AdapterHandler
func (a AdapterHandler) HandlerFunc(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	log.C(ctx).Info("The Kyma Tenant Mapping Adapter was hit.")

	var reqBody types.Body
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
	tenantID := reqBody.ReceiverTenant.OwnerTenant
	log.C(ctx).Infof("The request has tenant with id %q", tenantID)

	if reqBody.Context.Operation == assignOperation {
		log.C(ctx).Infof("The request operation is %q", assignOperation)
		a.processAssignOperation(ctx, w, reqBody, tenantID)
	} else {
		log.C(ctx).Infof("The request operation is %q", unassignOperation)
		a.processUnassignOperation(ctx, w, reqBody, tenantID)
	}

	log.C(ctx).Info("The Kyma integration was successfully processed")
}

func (a AdapterHandler) processAssignOperation(ctx context.Context, w http.ResponseWriter, reqBody types.Body, tenant string) {
	if reqBody.GetApplicationConfiguration() == (types.Configuration{}) { // config is missing
		respondWithSuccess(ctx, w, configPendingState, fmt.Sprintf("Configuration is missing. Responding with %q state...", configPendingState))
		return
	}

	appID := reqBody.GetApplicationID()
	rtmID := reqBody.GetRuntimeID()

	log.C(ctx).Infof("Getting application bundles for app with id %q and tenant %q", appID, tenant)
	bundles, err := a.DirectorGqlClient.GetApplicationBundles(ctx, appID, tenant)
	if err != nil {
		respondWithError(ctx, w, http.StatusBadRequest, createErrorState, errors.Wrapf(err, "while getting application bundles"))
		return
	}

	if len(bundles) == 0 {
		respondWithSuccess(ctx, w, readyState, fmt.Sprintf("There are no bundles for application with ID %q", appID))
		return
	}

	// Decide if it is Assign or Resync operation
	instanceAuthExist := false
	for _, instanceAuth := range bundles[0].InstanceAuths {
		if *instanceAuth.RuntimeID == rtmID {
			instanceAuthExist = true
			break
		}
	}

	if instanceAuthExist { // resync case
		a.processAuthRotation(ctx, w, bundles, reqBody, tenant)
	} else { // assign case
		a.processAuthCreation(ctx, w, bundles, reqBody, tenant)
	}
}

func (a AdapterHandler) processUnassignOperation(ctx context.Context, w http.ResponseWriter, reqBody types.Body, tenant string) {
	appID := reqBody.GetApplicationID()
	rtmID := reqBody.GetRuntimeID()

	bundles, err := a.DirectorGqlClient.GetApplicationBundles(ctx, appID, tenant)
	if err != nil {
		respondWithError(ctx, w, http.StatusBadRequest, deleteErrorState, errors.Wrapf(err, "while getting application bundles"))
		return
	}

	if len(bundles) == 0 {
		respondWithSuccess(ctx, w, readyState, fmt.Sprintf("There are no bundles for application with ID %q", reqBody.GetApplicationID()))
		return
	}

	instanceAuthExist := false
	for _, bundle := range bundles {
		for _, instanceAuth := range bundle.InstanceAuths {
			if *instanceAuth.RuntimeID == rtmID {
				instanceAuthExist = true

				if err = a.DirectorGqlClient.DeleteBundleInstanceAuth(ctx, tenant, instanceAuth.ID); err != nil {
					respondWithError(ctx, w, http.StatusBadRequest, deleteErrorState, errors.Wrapf(err, fmt.Sprintf("while deleting bundle instance auth with id: %q", instanceAuth.ID)))
					return
				}
			}
		}
	}

	if !instanceAuthExist {
		respondWithSuccess(ctx, w, readyState, fmt.Sprintf("There are no bundle instance auths for deletion for runtime with ID %q and application with id %q", rtmID, appID))
		return
	}

	respondWithSuccess(ctx, w, readyState, fmt.Sprintf("Successfully deleted the bundle instance auths for runtime with ID %q and application with id %q", rtmID, appID))
}

func (a AdapterHandler) processAuthCreation(ctx context.Context, w http.ResponseWriter, bundles []*graphql.BundleExt, reqBody types.Body, tenant string) {
	rtmID := reqBody.GetRuntimeID()
	basicCreds := reqBody.GetBasicCredentials()
	oauthCreds := reqBody.GetOauthCredentials()

	for _, bundle := range bundles {
		if basicCreds.Username != "" { // basic creds
			if err := a.DirectorGqlClient.CreateBasicBundleInstanceAuth(ctx, tenant, bundle.ID, rtmID, basicCreds.Username, basicCreds.Password); err != nil {
				respondWithError(ctx, w, http.StatusBadRequest, createErrorState, errors.Wrapf(err, fmt.Sprintf("while creating bundle instance auth for bundle with id: %q", bundle.ID)))
				return
			}
		} else { // oauth creds
			if err := a.DirectorGqlClient.CreateOauthBundleInstanceAuth(ctx, tenant, bundle.ID, rtmID, oauthCreds.TokenServiceURL, oauthCreds.ClientID, oauthCreds.ClientSecret); err != nil {
				respondWithError(ctx, w, http.StatusBadRequest, createErrorState, errors.Wrapf(err, fmt.Sprintf("while creating bundle instance auth for bundle with id: %q", bundle.ID)))
				return
			}
		}
	}

	respondWithSuccess(ctx, w, readyState, fmt.Sprintf("Successfully created the bundle instance auths for runtime with ID %q", rtmID))
}

func (a AdapterHandler) processAuthRotation(ctx context.Context, w http.ResponseWriter, bundles []*graphql.BundleExt, reqBody types.Body, tenant string) {
	rtmID := reqBody.GetRuntimeID()
	basicCreds := reqBody.GetBasicCredentials()
	oauthCreds := reqBody.GetOauthCredentials()

	for _, bundle := range bundles {
		for _, instanceAuth := range bundle.InstanceAuths {
			if *instanceAuth.RuntimeID == rtmID {
				if basicCreds.Username != "" { // basic creds
					if err := a.DirectorGqlClient.UpdateBasicBundleInstanceAuth(ctx, tenant, instanceAuth.ID, bundle.ID, basicCreds.Username, basicCreds.Password); err != nil {
						respondWithError(ctx, w, http.StatusBadRequest, createErrorState, errors.Wrapf(err, fmt.Sprintf("while updating bundle instance auth with id: %q", instanceAuth.ID)))
						return
					}
				} else { // oauth creds
					if err := a.DirectorGqlClient.UpdateOauthBundleInstanceAuth(ctx, tenant, instanceAuth.ID, bundle.ID, oauthCreds.TokenServiceURL, oauthCreds.ClientID, oauthCreds.ClientSecret); err != nil {
						respondWithError(ctx, w, http.StatusBadRequest, createErrorState, errors.Wrapf(err, fmt.Sprintf("while updating bundle instance auth with id: %q", instanceAuth.ID)))
						return
					}
				}
			}
		}
	}

	respondWithSuccess(ctx, w, readyState, fmt.Sprintf("Successfully updated the bundle instance auths for runtime with ID %q", rtmID))
}

type malformedRequest struct {
	status int
	msg    string
}

func (mr *malformedRequest) Error() string {
	return mr.msg
}

// SuccessResponse structure used for JSON encoded success response
type SuccessResponse struct {
	State string `json:"state,omitempty"`
}

// ErrorResponse structure used for JSON encoded error response
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
func respondWithSuccess(ctx context.Context, w http.ResponseWriter, state, msg string) {
	log.C(ctx).Info(msg)
	w.Header().Add(httputils.HeaderContentTypeKey, httputils.ContentTypeApplicationJSON)
	w.WriteHeader(successStatusCode)
	successResponse := SuccessResponse{State: state}
	httputils.RespondWithBody(ctx, w, successStatusCode, successResponse)
}
