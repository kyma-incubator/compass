package handler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/kyma-incubator/compass/components/kyma-adapter/internal/gqlclient"
	"github.com/kyma-incubator/compass/components/kyma-adapter/internal/types/credentials"
	"github.com/kyma-incubator/compass/components/kyma-adapter/internal/types/tenantmapping"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
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
	CreateBundleInstanceAuth(ctx context.Context, tenant string, input gqlclient.BundleInstanceAuthInput) error
	UpdateBundleInstanceAuth(ctx context.Context, tenant string, input gqlclient.BundleInstanceAuthInput) error
	DeleteBundleInstanceAuth(ctx context.Context, tenant string, input gqlclient.BundleInstanceAuthInput) error
}

// AdapterHandler is the Kyma Tenant Mapping Adapter handler which processes the received requests
type AdapterHandler struct {
	DirectorGqlClient Client
}

// NewHandler creates an AdapterHandler
func NewHandler(directorGqlClient Client) *AdapterHandler {
	return &AdapterHandler{DirectorGqlClient: directorGqlClient}
}

type modifyBundleInstanceAuthFunc func(ctx context.Context, tenant string, input gqlclient.BundleInstanceAuthInput) (string, error)

// HandlerFunc is the implementation of AdapterHandler
func (a AdapterHandler) HandlerFunc(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	log.C(ctx).Info("The Kyma Tenant Mapping Adapter was hit.")

	log.C(ctx).Info("Decoding the request body...")
	var reqBody tenantmapping.Body
	if err := decodeJSONBody(r, &reqBody); err != nil {
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

	operation := reqBody.Context.Operation
	log.C(ctx).Infof("The request operation is %q", operation)

	if operation == assignOperation && reqBody.GetApplicationConfiguration() == (tenantmapping.Configuration{}) { // config is missing
		respondWithSuccess(ctx, w, configPendingState, fmt.Sprintf("Configuration is missing. Responding with %q state...", configPendingState))
		return
	}

	appID := reqBody.GetApplicationID()
	rtmID := reqBody.GetRuntimeID()

	log.C(ctx).Infof("Getting application bundles for app with id %q and tenant %q", appID, tenantID)
	bundles, err := a.DirectorGqlClient.GetApplicationBundles(ctx, appID, tenantID)
	if err != nil {
		respondWithError(ctx, w, http.StatusBadRequest, createErrorState, errors.Wrapf(err, "while getting application bundles"))
		return
	}

	if len(bundles) == 0 {
		respondWithSuccess(ctx, w, readyState, fmt.Sprintf("There are no bundles for application with ID %q", appID))
		return
	}

	instanceAuthExist := false
	for _, bundle := range bundles {
		for _, instanceAuth := range bundle.InstanceAuths {
			if *instanceAuth.RuntimeID == rtmID {
				instanceAuthExist = true
				break
			}
		}
	}

	if !instanceAuthExist && operation == unassignOperation {
		respondWithSuccess(ctx, w, readyState, fmt.Sprintf("There are no bundle instance auths for deletion for runtime with ID %q and application with id %q", rtmID, appID))
		return
	}

	creds := credentials.NewCredentials(reqBody)
	modifyFunc := a.determineAuthModifyFunc(instanceAuthExist, operation)
	for _, bundle := range bundles {
		input := buildInstanceAuthInput(instanceAuthExist, operation, bundle, rtmID, creds)
		if state, err := modifyFunc(ctx, tenantID, input); err != nil {
			respondWithError(ctx, w, http.StatusBadRequest, state, err)
			return
		}
	}

	respondWithSuccess(ctx, w, readyState, "Successfully processed Kyma integration.")
}

func (a AdapterHandler) determineAuthModifyFunc(authExists bool, operation string) modifyBundleInstanceAuthFunc {
	if operation == assignOperation && authExists {
		// Update func
		return func(ctx context.Context, tenant string, input gqlclient.BundleInstanceAuthInput) (string, error) {
			if err := a.DirectorGqlClient.UpdateBundleInstanceAuth(ctx, tenant, input); err != nil {
				return createErrorState, err
			}
			return "", nil
		}
	} else if operation == assignOperation && !authExists {
		// Create func
		return func(ctx context.Context, tenant string, input gqlclient.BundleInstanceAuthInput) (string, error) {
			if err := a.DirectorGqlClient.CreateBundleInstanceAuth(ctx, tenant, input); err != nil {
				return createErrorState, err
			}
			return "", nil
		}
	} else {
		// Delete func
		return func(ctx context.Context, tenant string, input gqlclient.BundleInstanceAuthInput) (string, error) {
			if err := a.DirectorGqlClient.DeleteBundleInstanceAuth(ctx, tenant, input); err != nil {
				return deleteErrorState, err
			}
			return "", nil
		}
	}
}

func buildInstanceAuthInput(authExists bool, operation string, bundle *graphql.BundleExt, rtmID string, credentials credentials.Credentials) gqlclient.BundleInstanceAuthInput {
	if operation == assignOperation && authExists {
		// Update input
		return gqlclient.UpdateBundleInstanceAuthInput{
			Bundle:      bundle,
			RuntimeID:   rtmID,
			Credentials: credentials,
		}
	} else if operation == assignOperation && !authExists {
		// Create input
		return gqlclient.CreateBundleInstanceAuthInput{
			BundleID:    bundle.ID,
			RuntimeID:   rtmID,
			Credentials: credentials,
		}

	} else {
		// Delete input
		return gqlclient.DeleteBundleInstanceAuthInput{
			Bundle:    bundle,
			RuntimeID: rtmID,
		}
	}
}

type malformedRequest struct {
	status int
	msg    string
}

func (mr *malformedRequest) Error() string {
	return mr.msg
}
