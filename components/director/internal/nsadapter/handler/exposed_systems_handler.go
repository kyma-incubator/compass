package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"

	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/nsadapter/httputil"
	"github.com/kyma-incubator/compass/components/director/internal/nsadapter/nsmodel"
	"github.com/kyma-incubator/compass/components/director/pkg/httputils"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/pkg/errors"
)

const (
	deltaReportType      = "delta"
	fullReportType       = "full"
	reportTypeQueryParam = "reportType"
	notSubaccountMarker  = "not subaccount"
)

//go:generate mockery --exported --name=applicationService --output=automock --outpkg=automock --case=underscore --disable-version-string
type applicationService interface {
	CreateFromTemplate(ctx context.Context, in model.ApplicationRegisterInput, appTemplateID *string) (string, error)
	Upsert(ctx context.Context, in model.ApplicationRegisterInput) error
	Update(ctx context.Context, id string, in model.ApplicationUpdateInput) error
	GetSccSystem(ctx context.Context, sccSubaccount, locationID, virtualHost string) (*model.Application, error)
	ListBySCC(ctx context.Context, filter *labelfilter.LabelFilter) ([]*model.ApplicationWithLabel, error)
	SetLabel(ctx context.Context, label *model.LabelInput) error
	GetLabel(ctx context.Context, applicationID string, key string) (*model.Label, error)
	ListSCCs(ctx context.Context) ([]*model.SccMetadata, error)
}

//go:generate mockery --exported --name=applicationConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type applicationConverter interface {
	CreateInputJSONToModel(ctx context.Context, in string) (model.ApplicationRegisterInput, error)
}

//go:generate mockery --exported --name=applicationTemplateService --output=automock --outpkg=automock --case=underscore --disable-version-string
type applicationTemplateService interface {
	Get(ctx context.Context, id string) (*model.ApplicationTemplate, error)
	PrepareApplicationCreateInputJSON(appTemplate *model.ApplicationTemplate, values model.ApplicationFromTemplateInputValues) (string, error)
}

//go:generate mockery --exported --name=tenantService --output=automock --outpkg=automock --case=underscore --disable-version-string
type tenantService interface {
	ListsByExternalIDs(ctx context.Context, ids []string) ([]*model.BusinessTenantMapping, error)
}

// NewHandler returns new ns-adapter handler
func NewHandler(appSvc applicationService, appConverter applicationConverter, appTemplateSvc applicationTemplateService, tntSvc tenantService, transact persistence.Transactioner) *Handler {
	return &Handler{appSvc: appSvc, appConverter: appConverter, appTemplateSvc: appTemplateSvc, tntSvc: tntSvc, transact: transact}
}

// Handler implements handler interface
type Handler struct {
	appSvc         applicationService
	appConverter   applicationConverter
	appTemplateSvc applicationTemplateService
	tntSvc         tenantService
	transact       persistence.Transactioner
}

// Description - Bulk create-or-update operation on exposed on-premise systems. This handler supports two types of reports - full and delta.
// This handler takes a list of fully described SCCs together with the exposed systems.
// It creates new application for every exposed system for which CMP isn't aware of, and updates the metadata for the ones it is.
// - In case of full report: If there is SCC which was not reported, all exposed systems of this SCC are marked as unreachable.
// - In case of delta report: If there are missing exposed systems for a particular SCC, these systems are marked as unreachable.
// URL          - /api/v1/notifications
// Query Params - reportType=[full, delta]
// HTTP Method  - PUT
// Content-Type - application/json
// HTTP Codes:
// 204 No Content:
// - In case of delta report: if all systems are processed successfully
// - In case of full report: if the request was processed
// 200 OK:
// - In case of delta report: if update/create failed for some on-premise systems
// 400 Bad Request:
// - missing or invalid required report type query parameter
// - failed to parse request body
// - validating request body failed
// 500 Internal Server Error:
// - In case internal issue occurred. Example: db communication failed
func (a *Handler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	logger := log.C(ctx)

	defer func() {
		if err := req.Body.Close(); err != nil {
			logger.Error("Got error on closing request body", err)
		}
	}()

	reportHandlers := map[string]func(context.Context, []*nsmodel.SCC, []httputil.Detail, nsmodel.Report, http.ResponseWriter){
		fullReportType:  a.fullReportHandler,
		deltaReportType: a.deltaReportHandler,
	}

	reportType := req.URL.Query().Get(reportTypeQueryParam)
	reportHandler, found := reportHandlers[reportType]
	if !found {
		httputil.RespondWithError(ctx, rw, http.StatusBadRequest, httputil.Error{
			Code:    http.StatusBadRequest,
			Message: "the query parameter 'reportType' is missing or invalid",
		})
		return
	}

	var reportData nsmodel.Report
	err := json.NewDecoder(req.Body).Decode(&reportData)
	if err != nil {
		logger.Warnf("Got error on parsing Request Body: %v\n", err)
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

	sccs := make([]*nsmodel.SCC, 0, len(reportData.Value))
	externalIDs := make([]string, 0, len(reportData.Value))
	for _, scc := range reportData.Value {
		// New object with the same data is created and added to the sccs slice instead of adding &scc to the slice
		// because otherwise the slice is populated with copies of the last scc`s address
		s := &nsmodel.SCC{
			ExternalSubaccountID: scc.ExternalSubaccountID,
			InternalSubaccountID: scc.InternalSubaccountID,
			LocationID:           scc.LocationID,
			ExposedSystems:       scc.ExposedSystems,
		}
		externalIDs = append(externalIDs, scc.ExternalSubaccountID)
		sccs = append(sccs, s)
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
	mapExternalToInternal(ctx, tenants, sccs)
	details := make([]httputil.Detail, 0)
	filteredSccs := filterSccsByInternalID(ctx, sccs, &details)

	reportHandler(ctx, filteredSccs, details, reportData, rw)
}

func (a *Handler) deltaReportHandler(ctx context.Context, filteredSccs []*nsmodel.SCC, details []httputil.Detail, _ nsmodel.Report, rw http.ResponseWriter) {
	a.processDelta(ctx, filteredSccs, &details)
	if len(details) == 0 {
		httputils.RespondWithBody(ctx, rw, http.StatusNoContent, struct{}{})
		return
	}
	httputil.RespondWithError(ctx, rw, http.StatusOK, httputil.DetailedError{
		Code:    http.StatusOK,
		Message: "Update/create failed for some on-premise systems",
		Details: details,
	})
}

func (a *Handler) fullReportHandler(ctx context.Context, filteredSccs []*nsmodel.SCC, details []httputil.Detail, reportData nsmodel.Report, rw http.ResponseWriter) {
	a.processDelta(ctx, filteredSccs, &details)
	a.handleUnreachableScc(ctx, reportData)
	httputils.RespondWithBody(ctx, rw, http.StatusNoContent, struct{}{})
}

func (a *Handler) listTenantsByExternalIDs(ctx context.Context, ids []string) ([]*model.BusinessTenantMapping, error) {
	if len(ids) == 0 {
		return make([]*model.BusinessTenantMapping, 0), nil
	}

	tx, err := a.transact.Begin()
	if err != nil {
		log.C(ctx).Warn(errors.Wrapf(err, "while openning transaction"))
		return nil, err
	}
	defer a.transact.RollbackUnlessCommitted(ctx, tx)

	ctxWithTransaction := persistence.SaveToContext(ctx, tx)
	tenants, err := a.tntSvc.ListsByExternalIDs(ctxWithTransaction, ids)
	if err != nil {
		log.C(ctx).Warn(errors.Wrapf(err, "while listing tenants by external ids"))
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		log.C(ctx).Warn(errors.Wrapf(err, "while committing transaction"))
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
	defer a.transact.RollbackUnlessCommitted(ctx, tx)

	ctxWithTransaction := persistence.SaveToContext(ctx, tx)
	sccs, err := a.appSvc.ListSCCs(ctxWithTransaction)
	if err != nil {
		log.C(ctx).Warn(errors.Wrapf(err, "while listing all sccs"))
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		log.C(ctx).Warn(errors.Wrapf(err, "while committing transaction"))
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
			Subaccount: scc.ExternalSubaccountID,
			LocationID: scc.LocationID,
		})
	}

	sccsToMarkAsUnreachable := difference(sccs, sccsFromNs)

	externalSubaccounts := make([]string, 0, len(sccsToMarkAsUnreachable))
	for _, scc := range sccsToMarkAsUnreachable {
		externalSubaccounts = append(externalSubaccounts, scc.Subaccount)
	}

	internalSubaccounts, err := a.listTenantsByExternalIDs(ctx, externalSubaccounts)
	if err != nil {
		log.C(ctx).Warn(errors.Wrapf(err, "while listing subaccounts"))
		return
	}

	externalToInternalSub := make(map[string]string, len(internalSubaccounts))
	for _, subaccount := range internalSubaccounts {
		externalToInternalSub[subaccount.ExternalTenant] = subaccount.ID
	}

	for _, scc := range sccsToMarkAsUnreachable {
		scc.InternalSubaccountID = externalToInternalSub[scc.Subaccount]
	}

	for _, scc := range sccsToMarkAsUnreachable {
		ctxWithSubaccount := tenant.SaveToContext(ctx, scc.InternalSubaccountID, scc.Subaccount)
		appsWithLabels, ok := a.listAppsByScc(ctxWithSubaccount, scc.Subaccount, scc.LocationID)
		if ok {
			for _, appWithLabels := range appsWithLabels {
				a.markSystemAsUnreachable(ctxWithSubaccount, appWithLabels.App)
			}
		}
	}
}

func (a *Handler) processDelta(ctx context.Context, sccs []*nsmodel.SCC, details *[]httputil.Detail) {
	for _, scc := range sccs {
		ctxWithTenant := tenant.SaveToContext(ctx, scc.InternalSubaccountID, scc.ExternalSubaccountID)
		if ok := a.handleSccSystems(ctxWithTenant, *scc); !ok {
			addErrorDetailsMsg(details, scc, "Creation failed")
		}
	}
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
				txSucceeded = false
				log.C(ctx).Warn(errors.Wrapf(err, "while committing transaction"))
			}
		}

		a.transact.RollbackUnlessCommitted(ctx, tx)
		success = success && txSucceeded
	}
	return success
}

func (a *Handler) prepareAppInput(ctx context.Context, scc nsmodel.SCC, system nsmodel.System) (*model.ApplicationRegisterInput, error) {
	template, err := a.appTemplateSvc.Get(ctx, system.TemplateID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting application template with id: %s", system.TemplateID)
	}

	values := model.ApplicationFromTemplateInputValues{
		{
			Placeholder: "name",
			Value:       "on-premise-system" + uuid.New().String(),
		},
		{
			Placeholder: "description",
			Value:       system.Description,
		},
		{
			Placeholder: "subaccount",
			Value:       scc.ExternalSubaccountID,
		},
		{
			Placeholder: "location-id",
			Value:       scc.LocationID,
		},
		{
			Placeholder: "system-type",
			Value:       system.SystemType,
		},
		{
			Placeholder: "host",
			Value:       system.Host,
		},
		{
			Placeholder: "protocol",
			Value:       system.Protocol,
		},
		{
			Placeholder: "system-number",
			Value:       system.SystemNumber,
		},
		{
			Placeholder: "system-status",
			Value:       system.Status,
		},
	}

	appInputJSON, err := a.appTemplateSvc.PrepareApplicationCreateInputJSON(template, values)
	if err != nil {
		return nil, errors.Wrapf(err, "while preparing application create input from template with id:%s", system.TemplateID)
	}

	appInput, err := a.appConverter.CreateInputJSONToModel(ctx, appInputJSON)
	if err != nil {
		return nil, errors.Wrapf(err, "while preparing application create input from json")
	}

	if appInput.SystemNumber != nil && *appInput.SystemNumber == "" {
		appInput.SystemNumber = nil
	}

	return &appInput, nil
}

func (a *Handler) upsertWithSystemNumber(ctx context.Context, scc nsmodel.SCC, system nsmodel.System) bool {
	appInput, err := a.prepareAppInput(ctx, scc, system)
	if err != nil {
		log.C(ctx).Warn(errors.Wrapf(err, "while upserting Application"))
		return false
	}

	if err := a.appSvc.Upsert(ctx, *appInput); err != nil {
		log.C(ctx).Warn(errors.Wrapf(err, "while upserting Application"))
		return false
	}

	return true
}

func (a *Handler) upsert(ctx context.Context, scc nsmodel.SCC, system nsmodel.System) bool {
	app, err := a.appSvc.GetSccSystem(ctx, scc.ExternalSubaccountID, scc.LocationID, system.Host)

	if err != nil && isNotFoundError(err) {
		return a.createAppFromTemplate(ctx, scc, system)
	}

	if err != nil {
		log.C(ctx).Warn(errors.Wrapf(err, "while getting Application"))
		return false
	}

	return a.updateSystem(ctx, system, app)
}

func (a *Handler) createAppFromTemplate(ctx context.Context, scc nsmodel.SCC, system nsmodel.System) bool {
	appInput, err := a.prepareAppInput(ctx, scc, system)
	if err != nil {
		log.C(ctx).Warn(errors.Wrapf(err, "while creating Application"))
		return false
	}

	if _, err := a.appSvc.CreateFromTemplate(ctx, *appInput, str.Ptr(system.TemplateID)); err != nil {
		log.C(ctx).Warn(errors.Wrapf(err, "while creating Application"))
		return false
	}
	return true
}

func (a *Handler) updateSystem(ctx context.Context, system nsmodel.System, app *model.Application) bool {
	if err := a.appSvc.Update(ctx, app.ID, nsmodel.ToAppUpdateInput(system)); err != nil {
		log.C(ctx).Warn(errors.Wrapf(err, "while updating Application with id %s", app.ID))
		return false
	}

	if err := a.appSvc.SetLabel(ctx, &model.LabelInput{
		Key:        "systemType",
		Value:      system.SystemType,
		ObjectID:   app.ID,
		ObjectType: model.ApplicationLabelableObject,
	}); err != nil {
		log.C(ctx).Warn(errors.Wrapf(err, "while setting 'systemType' label for Application with id %s", app.ID))
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
	apps, ok := a.listAppsByScc(ctx, scc.ExternalSubaccountID, scc.LocationID)
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

	if err := a.appSvc.Update(ctxWithTransaction, system.ID, model.ApplicationUpdateInput{SystemStatus: str.Ptr("unreachable")}); err != nil {
		log.C(ctx).Warn(errors.Wrapf(err, "while marking application with id %s as unreachable", system.ID))
		return false
	}

	if err := tx.Commit(); err != nil {
		log.C(ctx).Warn(errors.Wrapf(err, "while committing transaction"))
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
	defer a.transact.RollbackUnlessCommitted(ctx, tx)

	ctxWithTransaction := persistence.SaveToContext(ctx, tx)
	apps, err := a.appSvc.ListBySCC(ctxWithTransaction, labelfilter.NewForKeyWithQuery("scc", fmt.Sprintf("{\"LocationID\":\"%s\", \"Subaccount\":\"%s\"}", locationID, subaccount)))
	if err != nil {
		log.C(ctx).Warn(errors.Wrapf(err, "while listing all applications for scc with subaccount %s and location id %s", subaccount, locationID))
		return nil, false
	}

	if err := tx.Commit(); err != nil {
		log.C(ctx).Warn(errors.Wrapf(err, "while committing transaction"))
		return nil, false
	}

	return apps, true
}

func filterUnreachable(apps []*model.ApplicationWithLabel, systems []nsmodel.System) []*model.Application {
	hostToSystem := make(map[string]interface{}, len(systems))

	for _, s := range systems {
		hostToSystem[s.Host] = struct{}{}
	}
	unreachable := make([]*model.Application, 0, len(apps))

	for _, a := range apps {
		result := a.SccLabel.Value.(map[string]interface{})["Host"]
		_, ok := hostToSystem[result.(string)]
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

func addErrorDetailsMsg(details *[]httputil.Detail, scc *nsmodel.SCC, message string) {
	*details = append(*details, httputil.Detail{
		Message:    message,
		Subaccount: scc.ExternalSubaccountID,
		LocationID: scc.LocationID,
	})
}

func mapExternalToInternal(ctx context.Context, tenants []*model.BusinessTenantMapping, sccs []*nsmodel.SCC) {
	externalToInternalTenants := make(map[string]*model.BusinessTenantMapping, len(tenants))
	for _, t := range tenants {
		externalToInternalTenants[t.ExternalTenant] = t
	}

	for _, scc := range sccs {
		t, exist := externalToInternalTenants[scc.ExternalSubaccountID]
		if !exist {
			continue
		}
		if t.Type == "subaccount" {
			scc.InternalSubaccountID = t.ID
		} else {
			log.C(ctx).Warnf("Got tenant with id: %s which is not a subaccount", t.ID)
			scc.InternalSubaccountID = notSubaccountMarker
		}
	}
}

func isNotFoundError(err error) bool {
	return strings.Contains(err.Error(), "Object not found")
}

func filterSccsByInternalID(ctx context.Context, sccs []*nsmodel.SCC, details *[]httputil.Detail) []*nsmodel.SCC {
	filteredSccs := make([]*nsmodel.SCC, 0, len(sccs))
	for _, scc := range sccs {
		if scc.InternalSubaccountID == "" {
			log.C(ctx).Warnf("Got SCC with external subaccount id: %s which has not associated internal tenant id", scc.ExternalSubaccountID)
			addErrorDetailsMsg(details, scc, "Subaccount not found")
		} else if scc.InternalSubaccountID == notSubaccountMarker {
			addErrorDetailsMsg(details, scc, "Provided id is not subaccount")
		} else {
			filteredSccs = append(filteredSccs, scc)
		}
	}
	return filteredSccs
}
