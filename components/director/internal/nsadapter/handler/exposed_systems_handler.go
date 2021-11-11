package handler

import (
	"context"
	"encoding/json"
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/nsadapter/httputil"
	"github.com/kyma-incubator/compass/components/director/internal/nsadapter/nsmodel"
	"github.com/kyma-incubator/compass/components/director/pkg/httputils"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"net/http"
)

type ApplicationService interface {
	Create(ctx context.Context, in model.ApplicationRegisterInput) (string, error)
	Upsert(ctx context.Context, in model.ApplicationRegisterInput) (string, error)
	Update(ctx context.Context, id string, in model.ApplicationUpdateInput) error
	GetSystem(ctx context.Context, subaccount, locationID, virtualHost string) (*model.Application, error)
	MarkAsUnreachable(ctx context.Context, id string) error
	ListBySCC(ctx context.Context, filter []*labelfilter.LabelFilter) ([]*ApplicationWithLabel, error) //TODO specify location ID in  label filter query
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
		httputil.RespondWithError(ctx, rw, http.StatusBadRequest, httputil.Error{
			Code:    http.StatusBadRequest,
			Message: "failed to parse request body",
		})
		return
	}

	if err := reportData.Validate(); err != nil {
		logger.Warnf("Got error while validating Request Body: %v\n", err)
		httputil.RespondWithError(ctx, rw, http.StatusBadRequest, httputil.Error{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	reportType := req.URL.Query().Get("reportType")

	if reportType != "full" && reportType != "delta" {
		httputil.RespondWithError(ctx, rw, http.StatusBadRequest, nsmodel.NewNSError("missing or invalid required report type query parameter"))
		return
	}

	details := make([]httputil.Detail, 0, 0)
	if reportType == "delta" {
		for _, scc := range reportData.Value {
			if ok := a.handleSccSystems(ctx, scc); !ok {
				addErrorDetails(&details, &scc)
			}
		}
		if len(details) == 0 {
			httputils.RespondWithBody(ctx, rw, http.StatusNoContent, struct{}{})
			return
		}
		httputil.RespondWithError(ctx, rw, http.StatusOK, httputil.DetailedError{
			Code:    0, //TODO change me
			Message: "Update/create failed for some on-premise systems",
			Details: details,
		})
		return
	}

	//Full report
	//The same as above
	//Check all SCCs in CMP with all SCCs in full report
	//If in CMP there are SCCs missing from the full report - mark all systems in these SCCs as unreachable
	//At this step not possible to have something in full report which is not available in SCC as in the previous step we will have created applications for all new systems in full report
	if reportType == "full" {
		//TODO implement
	}

	if err := json.NewEncoder(rw).Encode(nsmodel.NewNSError("response test")); err != nil {
		logger.Warnf("Got error on encoding response: %v\n", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (a *Handler) handleSccSystems(ctx context.Context, scc nsmodel.SCC) bool {
	successfulUpsert := a.upsertSccSystems(ctx, scc)
	successfulMark := a.markAsUnreachable(ctx, scc)
	return successfulUpsert && successfulMark
}

//TODO Fix all errors
func (a *Handler) upsertSccSystems(ctx context.Context, scc nsmodel.SCC) bool {
	success := true
	for _, system := range scc.ExposedSystems {
		//TODO open transaction
		if system.SystemNumber != "" {
			if _, err := a.appSvc.Upsert(ctx, nsmodel.ToAppRegisterInput(system, scc.LocationID)); err != nil {
				success = false
				log.C(ctx).Warn(errors.Wrapf(err, "while upserting Application"))
			}
			continue
		}

		if ok := a.upsert(ctx, scc, system); !ok {
			success = false
		}

	}
	return success
}

func IsNotFoundError(err error) bool {
	_, ok := err.(*nsmodel.SystemNotFoundError)
	return ok
}

func (a *Handler) upsert(ctx context.Context, scc nsmodel.SCC, system nsmodel.System) bool {
	app, err := a.appSvc.GetSystem(ctx, scc.Subaccount, scc.LocationID, system.Host)

	if err != nil && IsNotFoundError(err) {
		if _, err := a.appSvc.Create(ctx, nsmodel.ToAppRegisterInput(system, scc.LocationID)); err != nil {
			log.C(ctx).Warn(errors.Wrapf(err, "while creating Application"))
			return false
		}
		return true
	}

	if err != nil {
		log.C(ctx).Warn(errors.Wrapf(err, "while getting Application"))
		return false
	}

	return a.updateSystem(ctx, system, app)
}

func (a *Handler) updateSystem(ctx context.Context, system nsmodel.System, app *model.Application) bool {
	if err := a.appSvc.Update(ctx, app.ID, nsmodel.ToAppUpdateInput(system)); err != nil {
		log.C(ctx).Warn(errors.Wrapf(err, "while updating Application with id %s", app.ID))
		return false
	}

	//TODO check if something additional is needed
	if err := a.appSvc.SetLabel(ctx, &model.LabelInput{
		Key:        "applicationType",
		Value:      system.SystemType,
		ObjectID:   app.ID,
		ObjectType: model.ApplicationLabelableObject,
	}); err != nil {
		log.C(ctx).Warn(errors.Wrapf(err, "while setting 'applicationType' label for Application with id %s", app.ID))
		return false
	}

	if err := a.appSvc.SetLabel(ctx, &model.LabelInput{
		Key:        "systemProtocol",
		Value:      system.Protocol,
		ObjectID:   app.ID,
		ObjectType: model.ApplicationLabelableObject,
	}); err != nil {
		log.C(ctx).Warn(errors.Wrapf(err, "while setting 'systemProtocol' label for Application with id %s", app.ID))
		return false
	}

	return true
}

func addErrorDetails(details *[]httputil.Detail, scc *nsmodel.SCC) {
	*details = append(*details, httputil.Detail{
		Code:       "0000", //TODO change me
		Message:    "Creation failed",
		Subaccount: scc.Subaccount,
		LocationId: scc.LocationID,
	})
}

func (a *Handler) markAsUnreachable(ctx context.Context, scc nsmodel.SCC) bool {
	success := true
	apps, err := a.appSvc.ListBySCC(ctx, nil) //TODO change nil to proper labelfilter
	if err != nil {
		success = false
		log.C(ctx).Warn(errors.Wrapf(err, "while listing all applications for scc with subaccount %s and location id %s", scc.Subaccount, scc.LocationID))
	}

	unreachable := filterUnreachable(apps, scc.ExposedSystems)
	for _, system := range unreachable {
		if err := a.appSvc.MarkAsUnreachable(ctx, system.ID); err != nil {
			success = false
			log.C(ctx).Warn(errors.Wrapf(err, "while marking application with id %s as unreachable", system.ID))
		}
	}
	return success
}

func filterUnreachable(apps []*ApplicationWithLabel, systems []nsmodel.System) []*model.Application {
	hostToSystem := make(map[string]interface{}, len(systems))

	for _, s := range systems {
		hostToSystem[s.Host] = struct{}{}
	}

	unreachable := make([]*model.Application, 0, 0)

	for _, a := range apps {
		result := gjson.Get(a.sccLabel.Value.(string), "Host")
		_, ok := hostToSystem[result.Value().(string)]
		if !ok {
			unreachable = append(unreachable, a.app)
		}
	}
	return unreachable
}

type ApplicationWithLabel struct {
	app      *model.Application
	sccLabel *model.Label
}
