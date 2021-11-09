package handler

import (
	"context"
	"encoding/json"

	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/nsadapter/httputil"
	"github.com/kyma-incubator/compass/components/director/internal/nsadapter/nsmodel"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"net/http"
)

type ApplicationService interface {
	Create(ctx context.Context, in model.ApplicationRegisterInput) (string, error)
	Upsert(ctx context.Context, in model.ApplicationRegisterInput) (string, error)
	Update(ctx context.Context, id string, in model.ApplicationUpdateInput) error
	GetSystem(ctx context.Context, subaccount, locationID, virtualHost string) (*model.Application, error)
	MarkAsUnreachable(ctx context.Context, id string) error
	ListBySCC(ctx context.Context, filter []*labelfilter.LabelFilter) ([]*model.Application, error)
	SetLabel(ctx context.Context, label *model.LabelInput) error
	GetLabel(ctx context.Context, applicationID string, key string) (*model.Label, error)
}

func NewHandler(service ApplicationService) *Handler {
	return &Handler{appSvc: service}
}

type Handler struct {
	appSvc ApplicationService
}

//Description	Upsert ExposedSystems is a bulk create-or-update operation on exposed on-premise systems. It takes a list of fully described exposed systems, creates the ones CMP isn't aware of and updates the metadata for the ones it is.
//URL	/api/v1/notifications
//Path Params
//Query Params	reportType=full,delta
//HTTP Method	PUT
//HTTP Headers
//Content-Type: application/json
//HTTP Codes
//204 No Content
//200 OK
//400 Bad Request
//401 Unauthorized
//408 Request Timeout
//500 Internal Server Error
//502 Bad Gateway
//Response Formats	json
//Authentication	TODO
func (a *Handler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	logger := log.C(ctx)

	defer func() {
		if err := req.Body.Close(); err != nil {
			logger.Error("Got error on closing request body", err)
		}
	}()

	decoder := json.NewDecoder(req.Body)
	var reportData nsmodel.Report
	err := decoder.Decode(&reportData)
	if err != nil {
		logger.Warnf("Got error on decoding Request Body: %v\n", err)
		httputil.RespondWithError(ctx, rw, http.StatusBadRequest, nsmodel.NewNSError("failed to parse request body"))
		return
	}

	if err := reportData.Validate(); err != nil {
		logger.Warnf("Got error while validating Request Body: %v\n", err)
		httputil.RespondWithError(ctx, rw, http.StatusBadRequest, nsmodel.NewNSError(err.Error()))
		return
	}

	reportType := req.URL.Query().Get("reportType")

	if reportType != "full" && reportType != "delta" {
		httputil.RespondWithError(ctx, rw, http.StatusBadRequest, nsmodel.NewNSError("missing or invalid required report type query parameter"))
		return
	}
	
	if reportType == "delta" {
		for _, scc := range reportData.Value {
			for _, system := range scc.ExposedSystems {
				if system.SystemNumber != "" {
					a.appSvc.Upsert(ctx,)
				} else {
					system := GetSystem(subacc, locId, virtual_svc); //fetch all systems for the current scc from the db
					if system exists {
						UpdateApp()
						GetLabel()
						SetLabel()
						if needed
					} else {
						CreateApplication()

					}
				}
				//If there are missing systems in SCC report -> mark the respective systems in CMP as unreachable
				allDBsystems = getAllForSCC(subacc, locationID)
				unreachable
				filterUnreachable(dbSystems, systems)
				for _, system := range unreachable {
					markAsUnreachable(system)
				}
			}
		}
		//Check for each SCC, if there is a difference in its systems in CMP and the reported system
		//If there are new systems added to the SCC report-> create applications in CMP for them

		//For all systems which are present in both places - compare data (type and protocol)
	    //[\{"protocol":"HTTP","host":"127.0.0.1:8080","description":"","type":"otherSAPsys","status":"disabled"}]
//label with protocol, virtual_host and location ID
// join tenants_apps on app_id applications on app_id labels where correct protocol, virtual_host and location ID
		//Systems should be labeled that the source is NS
		//Identify which is the landscape
	} else if reportType == "full" {

	}

	if err := json.NewEncoder(rw).Encode(nsmodel.NewNSError("response test")); err != nil {
		logger.Warnf("Got error on encoding response: %v\n", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
}


