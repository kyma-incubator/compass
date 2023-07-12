package handler

import (
	"context"
	"fmt"
	"github.com/kyma-incubator/compass/components/kyma-adapter/internal/types/credentials"
	"github.com/kyma-incubator/compass/components/kyma-adapter/internal/types/tenantmapping"
	"net/http"

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
	CreateBundleInstanceAuth(ctx context.Context, tenant, bndlID, rtmID string, credentials credentials.Credentials) error
	UpdateBundleInstanceAuth(ctx context.Context, tenant, authID, bndlID string, credentials credentials.Credentials) error
	DeleteBundleInstanceAuth(ctx context.Context, tenant, authID string) error
}

// AdapterHandler is the Kyma Tenant Mapping Adapter handler which processes the received requests
type AdapterHandler struct {
	DirectorGqlClient Client
}

// NewHandler creates an AdapterHandler
func NewHandler(directorGqlClient Client) *AdapterHandler {
	return &AdapterHandler{DirectorGqlClient: directorGqlClient}
}

// HandlerFunc is the implementation of AdapterHandler
func (a AdapterHandler) HandlerFunc(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	log.C(ctx).Info("The Kyma Tenant Mapping Adapter was hit.")

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

	if reqBody.Context.Operation == assignOperation {
		log.C(ctx).Infof("The request operation is %q", assignOperation)
		a.processAssignOperation(ctx, w, reqBody, tenantID)
	} else {
		log.C(ctx).Infof("The request operation is %q", unassignOperation)
		a.processUnassignOperation(ctx, w, reqBody, tenantID)
	}
}

func (a AdapterHandler) processAssignOperation(ctx context.Context, w http.ResponseWriter, reqBody tenantmapping.Body, tenant string) {
	if reqBody.GetApplicationConfiguration() == (tenantmapping.Configuration{}) { // config is missing
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

func (a AdapterHandler) processUnassignOperation(ctx context.Context, w http.ResponseWriter, reqBody tenantmapping.Body, tenant string) {
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

func (a AdapterHandler) processAuthCreation(ctx context.Context, w http.ResponseWriter, bundles []*graphql.BundleExt, reqBody tenantmapping.Body, tenant string) {
	rtmID := reqBody.GetRuntimeID()

	for _, bundle := range bundles {
		creds := credentials.NewCredentials(reqBody)

		if err := a.DirectorGqlClient.CreateBundleInstanceAuth(ctx, tenant, bundle.ID, rtmID, creds); err != nil {
			respondWithError(ctx, w, http.StatusBadRequest, createErrorState, errors.Wrapf(err, fmt.Sprintf("while creating bundle instance auth for bundle with id: %q", bundle.ID)))
			return
		}
	}

	respondWithSuccess(ctx, w, readyState, fmt.Sprintf("Successfully created the bundle instance auths for runtime with ID %q", rtmID))
}

func (a AdapterHandler) processAuthRotation(ctx context.Context, w http.ResponseWriter, bundles []*graphql.BundleExt, reqBody tenantmapping.Body, tenant string) {
	rtmID := reqBody.GetRuntimeID()

	for _, bundle := range bundles {
		for _, instanceAuth := range bundle.InstanceAuths {
			if *instanceAuth.RuntimeID == rtmID {
				creds := credentials.NewCredentials(reqBody)

				if err := a.DirectorGqlClient.UpdateBundleInstanceAuth(ctx, tenant, instanceAuth.ID, bundle.ID, creds); err != nil {
					respondWithError(ctx, w, http.StatusBadRequest, createErrorState, errors.Wrapf(err, fmt.Sprintf("while updating bundle instance auth with id: %q", instanceAuth.ID)))
					return
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
