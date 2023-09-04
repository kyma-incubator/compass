package handler

import (
	"context"
	"net/http"

	"github.com/kyma-incubator/compass/components/instance-creator/internal/client/resources"
)

// Client is used to call SM
//go:generate mockery --name=Client --output=automock --outpkg=automock --case=underscore --disable-version-string
type Client interface {
	RetrieveResource(ctx context.Context, region, subaccountID string, resources resources.Resources, resourceArgs resources.ResourceArguments) (string, error)
	RetrieveResourceByID(ctx context.Context, region, subaccountID string, resource resources.Resource, resourceArgs resources.ResourceArguments) (resources.Resource, error)
	CreateResource(ctx context.Context, region, subaccountID string, resourceReqBody resources.ResourceRequestBody, resource resources.Resource) (string, error)
	DeleteResource(ctx context.Context, region, subaccountID string, resource resources.Resource, resourceArgs resources.ResourceArguments) error
	DeleteMultipleResources(ctx context.Context, region, subaccountID string, resources resources.Resources, resourceArgs resources.ResourceArguments) error
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
