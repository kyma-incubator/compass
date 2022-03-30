package tenantfetchersvc

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"net/http"
)

type fetcher struct {
	gqlClient   DirectorGraphQLClient
	provisioner TenantProvisioner
}

// NewFetcher creates new fetcher
func NewFetcher(directorClient DirectorGraphQLClient, provisioner TenantProvisioner) *fetcher {
	return &fetcher{
		gqlClient:   directorClient,
		provisioner: provisioner,
	}
}

func (f *fetcher) FetchTenantOnDemand(ctx context.Context, request *http.Request) error {
	vars := mux.Vars(request)
	_, ok := vars["tenantId"]
	if !ok {
		log.C(ctx).Error("Subaccount path parameter is missing from request")
		return nil
	}
	//TODO extract and call SyncTenant from tenant fetcher deployment
	return nil
}
