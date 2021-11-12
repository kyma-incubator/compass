package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/nsadapter/httputil"
	"github.com/kyma-incubator/compass/components/director/internal/nsadapter/nsmodel"
	"github.com/kyma-incubator/compass/components/director/pkg/httputils"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/tenant"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"io"
	"net/http"
)

type ApplicationService interface {
	Create(ctx context.Context, in model.ApplicationRegisterInput) (string, error)
	Upsert(ctx context.Context, in model.ApplicationRegisterInput) (string, error)
	Update(ctx context.Context, id string, in model.ApplicationUpdateInput) error
	GetSystem(ctx context.Context, subaccount, locationID, virtualHost string) (*model.Application, error)
	MarkAsUnreachable(ctx context.Context, id string) error
	ListBySCC(ctx context.Context, filter []*labelfilter.LabelFilter) ([]*model.ApplicationWithLabel, error)
	SetLabel(ctx context.Context, label *model.LabelInput) error
	GetLabel(ctx context.Context, applicationID string, key string) (*model.Label, error)
	ListSCCs(ctx context.Context, key string) ([]*model.SccMetadata, error) //TODO check what tenant will be used to execute this query
}

type TenantService interface {
	ListsByExternalIDs(ctx context.Context, ids []string) ([]*model.BusinessTenantMapping, error)
}

func NewHandler(appSvc ApplicationService, tntSvc TenantService, transact persistence.Transactioner) *Handler {
	return &Handler{appSvc: appSvc, tntSvc: tntSvc, transact: transact}
}

type Handler struct {
	appSvc   ApplicationService
	tntSvc   TenantService
	transact persistence.Transactioner
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

	buf := &bytes.Buffer{}
	_, err := io.Copy(buf, req.Body)
	if err != nil {
		logger.Warnf("Failed to retrieve request body: %v\n", err)
		httputil.RespondWithError(ctx, rw, http.StatusBadRequest, httputil.Error{
			Code:    http.StatusBadRequest,
			Message: "failed to retrieve request body",
		})
		return
	}
	
	var reportData nsmodel.Report
	err = json.Unmarshal(buf.Bytes(), &reportData)
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

	externalIDToTenant := make(map[string]*model.BusinessTenantMapping, len(reportData.Value))
	externalIDs := make([]string, 0, len(reportData.Value))
	for _, scc := range reportData.Value {
		if _, ok := externalIDToTenant[scc.Subaccount]; !ok {
			externalIDs = append(externalIDs, scc.Subaccount)
			externalIDToTenant[scc.Subaccount] = nil
		}
	}

	tenants, err := a.listTenantsByExternalIDs(ctx, externalIDs)
	if err != nil {
		logger.Warnf("Got error while listing subaccounts: %v\n", err)
		httputil.RespondWithError(ctx, rw, http.StatusInternalServerError, httputil.Error{
			Code:    http.StatusInternalServerError,
			Message: "Update failed",
		})
		return
	}

	//TODO what if there are missing tenants? len(tenants) != len(externalIDs)

	for _, t := range tenants {
		if t.Type != tenant.Subaccount {
			logger.Warnf("Got tenant which is not asubaccount")
			httputil.RespondWithError(ctx, rw, http.StatusInternalServerError, httputil.Error{
				Code:    http.StatusInternalServerError,
				Message: "Update failed",
			})
			return
		}
		externalIDToTenant[t.ExternalTenant] = t
	}

	reportType := req.URL.Query().Get("reportType")

	if reportType != "full" && reportType != "delta" {
		httputil.RespondWithError(ctx, rw, http.StatusBadRequest, httputil.Error{
			Code:    http.StatusBadRequest,
			Message: "missing or invalid required report type query parameter",
		})
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

func (a *Handler) listTenantsByExternalIDs(ctx context.Context, ids []string) ([]*model.BusinessTenantMapping, error) {
	tx, err := a.transact.Begin()
	if err != nil {
		log.C(ctx).Warn(errors.Wrapf(err, "while openning transaction"))
		return nil, err
	}
	ctxWithTransaction := persistence.SaveToContext(ctx, tx)

	tenants, err := a.tntSvc.ListsByExternalIDs(ctxWithTransaction, ids)

	if err := tx.Commit(); err != nil {
		log.C(ctx).Warn(errors.Wrapf(err, "while commiting transaction"))
		return nil, err
	}

	return tenants, nil

}

func (a *Handler) listSCCs(ctx context.Context) ([]*model.SccMetadata, error) {
	tx, err := a.transact.Begin()
	if err != nil {
		log.C(ctx).Warn(errors.Wrapf(err, "while openning transaction"))
		return nil, err
	}
	ctxWithTransaction := persistence.SaveToContext(ctx, tx)

	sccs, err := a.appSvc.ListSCCs(ctxWithTransaction, "scc")
	if err != nil {
		if err := tx.Rollback(); err != nil {
			log.C(ctx).Warn(errors.Wrapf(err, "while rolling back transaction"))
		}
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		log.C(ctx).Warn(errors.Wrapf(err, "while commiting transaction"))
		return nil, err
	}

	return sccs, nil

}

func (a *Handler) handleUnreachableScc(ctx context.Context, reportData nsmodel.Report) {
	sccs, err := a.listSCCs(ctx)
	if err != nil {
		log.C(ctx).Warn(errors.Wrapf(err, "while listing sccs"))
		return
	}

	if len(sccs) == len(reportData.Value) {
		return
	}

	sccsFromNs := make([]*model.SccMetadata, 0, len(reportData.Value))
	for _, scc := range reportData.Value {
		sccsFromNs = append(sccsFromNs, &model.SccMetadata{
			Subaccount: scc.Subaccount,
			LocationId: scc.LocationID,
		})
	}

	sccsToMarkAsUnreachable := difference(sccs, sccsFromNs)
	for _, scc := range sccsToMarkAsUnreachable {
		//TODO add correct tenant in ctx
		appsWithLabels, ok := a.listAppsByScc(ctx, scc.Subaccount, scc.LocationId)
		if ok {
			for _, appWithLabels := range appsWithLabels {
				a.markSystemAsUnreachable(ctx, appWithLabels.App)
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
		tx, err := a.transact.Begin()
		if err != nil {
			log.C(ctx).Warn(errors.Wrapf(err, "while openning transaction"))
			return false
		}
		ctxWithTransaction := persistence.SaveToContext(ctx, tx)

		var txSucceeded bool
		if system.SystemNumber != "" {
			txSucceeded = a.upsertWithSystemNumber(ctxWithTransaction, scc, system)
		} else {
			txSucceeded = a.upsert(ctxWithTransaction, scc, system)
		}

		if txSucceeded {
			if err := tx.Commit(); err != nil {
				log.C(ctx).Warn(errors.Wrapf(err, "while commiting transaction"))
			}
		}

		a.transact.RollbackUnlessCommitted(ctx, tx)
		success = success && txSucceeded
	}
	return success
}

func (a *Handler) upsertWithSystemNumber(ctx context.Context, scc nsmodel.SCC, system nsmodel.System) bool {
	if _, err := a.appSvc.Upsert(ctx, nsmodel.ToAppRegisterInput(system, scc.LocationID)); err != nil {
		log.C(ctx).Warn(errors.Wrapf(err, "while upserting Application"))
		return false
	}

	return true
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
	apps, ok := a.listAppsByScc(ctx, scc.Subaccount, scc.LocationID)
	if !ok {
		return false
	}

	success := true
	unreachable := filterUnreachable(apps, scc.ExposedSystems)
	for _, system := range unreachable {
		success = a.markSystemAsUnreachable(ctx, system) && success
	}
	return success
}

func (a *Handler) markSystemAsUnreachable(ctx context.Context, system *model.Application) bool {
	tx, err := a.transact.Begin()
	if err != nil {
		log.C(ctx).Warn(errors.Wrapf(err, "while openning transaction"))
		return false
	}
	defer a.transact.RollbackUnlessCommitted(ctx, tx)

	ctxWithTransaction := persistence.SaveToContext(ctx, tx)

	if err := a.appSvc.MarkAsUnreachable(ctxWithTransaction, system.ID); err != nil {
		log.C(ctx).Warn(errors.Wrapf(err, "while marking application with id %s as unreachable", system.ID))
		return false
	}

	if err := tx.Commit(); err != nil {
		log.C(ctx).Warn(errors.Wrapf(err, "while commiting transaction"))
		return false
	}

	return true
}

func (a *Handler) listAppsByScc(ctx context.Context, subaccount, locationID string) ([]*model.ApplicationWithLabel, bool) {
	tx, err := a.transact.Begin()
	if err != nil {
		log.C(ctx).Warn(errors.Wrapf(err, "while openning transaction"))
		return nil, false
	}
	ctxWithTransaction := persistence.SaveToContext(ctx, tx)

	apps, err := a.appSvc.ListBySCC(ctxWithTransaction, []*labelfilter.LabelFilter{labelfilter.NewForKeyWithQuery("scc", fmt.Sprintf("{\"locationId\":%s}", locationID))})
	if err != nil {
		log.C(ctx).Warn(errors.Wrapf(err, "while listing all applications for scc with subaccount %s and location id %s", subaccount, locationID))
		if err := tx.Rollback(); err != nil {
			log.C(ctx).Warn(errors.Wrapf(err, "while rolling back transaction"))
		}
		return nil, false
	}

	if err := tx.Commit(); err != nil {
		log.C(ctx).Warn(errors.Wrapf(err, "while commiting transaction"))
		return nil, false
	}

	return apps, true
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
