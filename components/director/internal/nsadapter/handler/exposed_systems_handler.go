package handler

import (
	"context"
	"encoding/json"
	"fmt"
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
	ListBySCC(ctx context.Context, filter []*labelfilter.LabelFilter) ([]*model.ApplicationWithLabel, error) //TODO specify location ID in  label filter query
	SetLabel(ctx context.Context, label *model.LabelInput) error
	GetLabel(ctx context.Context, applicationID string, key string) (*model.Label, error)
	ListSCCs(ctx context.Context, key string) ([]*model.SccMetadata, error) //TODO check what tenant will be used to execute this query
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

//TODO handle transfer-encoding chunked
//TODO check if all subaccounts are really subaccounts

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

	if reportType == "delta" {
		details := a.processDelta(ctx, reportData)
		if len(details) == 0 {
			httputils.RespondWithBody(ctx, rw, http.StatusNoContent, struct{}{})
			return
		}
		httputil.RespondWithError(ctx, rw, http.StatusOK, httputil.DetailedError{
			Code:    http.StatusOK,
			Message: "Update/create failed for some on-premise systems",
			Details: details,
		})
		return
	}

	if reportType == "full" {
		a.processDelta(ctx, reportData)
		a.handleUnreachableScc(ctx, reportData)
		httputils.RespondWithBody(ctx, rw, http.StatusNoContent, struct{}{})
		return
	}
}

func (a *Handler) handleUnreachableScc(ctx context.Context, reportData nsmodel.Report) {
	sccFromDb, err := a.appSvc.ListSCCs(ctx, "scc")
	if err != nil {
		log.C(ctx).Warnf("Got error while listing SCCs: %v\n", err)
	}

	if len(sccFromDb) == len(reportData.Value) {
		return
	}

	sccFromNs := make([]*model.SccMetadata, 0, len(reportData.Value))
	for _, scc := range reportData.Value {
		sccFromNs = append(sccFromNs, &model.SccMetadata{
			Subaccount: scc.Subaccount,
			LocationId: scc.LocationID,
		})
	}

	sccsToMarkAsUnreachable := difference(sccFromDb, sccFromNs)
	for _, scc := range sccsToMarkAsUnreachable {
		//TODO add correct tenant in ctx
		apps, err := a.appSvc.ListBySCC(ctx, []*labelfilter.LabelFilter{labelfilter.NewForKeyWithQuery("scc",fmt.Sprintf("{\"locationId\":%s}", scc.LocationId))})
		if err != nil {
			log.C(ctx).Warn(errors.Wrapf(err, "while listing all applications for scc with subaccount %s and location id %s", scc.Subaccount, scc.LocationId))
			continue
		}
		for _, system := range apps {
			if err := a.appSvc.MarkAsUnreachable(ctx, system.App.ID); err != nil {
				log.C(ctx).Warn(errors.Wrapf(err, "while marking application with id %s as unreachable", system.App.ID))
			}
		}
	}
}

func (a *Handler) processDelta(ctx context.Context, reportData nsmodel.Report) []httputil.Detail {
	details := make([]httputil.Detail, 0, 0)
	for _, scc := range reportData.Value {
		if ok := a.handleSccSystems(ctx, scc); !ok {
			addErrorDetails(&details, &scc)
		}
	}
	return details
}

func (a *Handler) handleSccSystems(ctx context.Context, scc nsmodel.SCC) bool {
	successfulUpsert := a.upsertSccSystems(ctx, scc)
	successfulMark := a.markAsUnreachable(ctx, scc)
	return successfulUpsert && successfulMark
}

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

func (a *Handler) upsert(ctx context.Context, scc nsmodel.SCC, system nsmodel.System) bool {
	app, err := a.appSvc.GetSystem(ctx, scc.Subaccount, scc.LocationID, system.Host)

	if err != nil && nsmodel.IsNotFoundError(err) {
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

func (a *Handler) markAsUnreachable(ctx context.Context, scc nsmodel.SCC) bool {
	success := true
	apps, err := a.appSvc.ListBySCC(ctx, []*labelfilter.LabelFilter{labelfilter.NewForKeyWithQuery("scc",fmt.Sprintf("{\"locationId\":%s}", scc.LocationID))})
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

func filterUnreachable(apps []*model.ApplicationWithLabel, systems []nsmodel.System) []*model.Application {
	hostToSystem := make(map[string]interface{}, len(systems))

	for _, s := range systems {
		hostToSystem[s.Host] = struct{}{}
	}

	unreachable := make([]*model.Application, 0, 0)

	for _, a := range apps {
		result := gjson.Get(a.SccLabel.Value.(string), "Host")
		_, ok := hostToSystem[result.Value().(string)]
		if !ok {
			unreachable = append(unreachable, a.App)
		}
	}
	return unreachable
}

func difference(a, b []*model.SccMetadata) (diff []*model.SccMetadata) {
	m := make(map[model.SccMetadata]bool)

	for _, item := range b {
		m[*item] = true
	}

	for _, item := range a {
		if _, ok := m[*item]; !ok {
			diff = append(diff, item)
		}
	}
	return
}

func addErrorDetails(details *[]httputil.Detail, scc *nsmodel.SCC) {
	*details = append(*details, httputil.Detail{
		Message:    "Creation failed",
		Subaccount: scc.Subaccount,
		LocationId: scc.LocationID,
	})
}