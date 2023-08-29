package handler

import (
	"context"
	"net/http"

	"github.com/kyma-incubator/compass/components/instance-creator/internal/client/types"
)

// Client is used to call SM
//go:generate mockery --name=Client --output=automock --outpkg=automock --case=underscore --disable-version-string
type Client interface {
	RetrieveServiceOffering(ctx context.Context, region, catalogName, subaccountID string) (string, error)
	RetrieveServicePlan(ctx context.Context, region, planName, offeringID, subaccountID string) (string, error)
	RetrieveServiceKeyByID(ctx context.Context, region, serviceKeyID, subaccountID string) (*types.ServiceKey, error)
	RetrieveServiceInstanceIDByName(ctx context.Context, region, serviceInstanceName, subaccountID string) (string, error)
	CreateServiceInstance(ctx context.Context, region, serviceInstanceName, planID, subaccountID string, parameters []byte) (string, error)
	CreateServiceKey(ctx context.Context, region, serviceKeyName, serviceInstanceID, subaccountID string, parameters []byte) (string, error)
	DeleteServiceInstance(ctx context.Context, region, serviceInstanceID, serviceInstanceName, subaccountID string) error
	DeleteServiceKeys(ctx context.Context, region, serviceInstanceID, serviceInstanceName, subaccountID string) error
}

// InstanceCreatorHandler processes received requests
type InstanceCreatorHandler struct {
	SMClient Client
}

// NewHandler creates an InstanceCreatorHandler
func NewHandler(smClient Client) *InstanceCreatorHandler {
	return &InstanceCreatorHandler{
		SMClient: smClient,
	}
}

// HandlerFunc is the implementation of InstanceCreatorHandler
func (i InstanceCreatorHandler) HandlerFunc(w http.ResponseWriter, r *http.Request) {

}
