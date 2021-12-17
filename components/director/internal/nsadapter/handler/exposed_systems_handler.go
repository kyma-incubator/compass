package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"io"
	"net/http"
	"strings"

	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/nsadapter/httputil"
	"github.com/kyma-incubator/compass/components/director/internal/nsadapter/nsmodel"
	"github.com/kyma-incubator/compass/components/director/pkg/httputils"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

//go:generate mockery --name=ApplicationService --output=automock --outpkg=automock --case=underscore
type ApplicationService interface {
	CreateFromTemplate(ctx context.Context, in model.ApplicationRegisterInput, appTemplateID *string) (string, error)
	Upsert(ctx context.Context, in model.ApplicationRegisterInput) error
	Update(ctx context.Context, id string, in model.ApplicationUpdateInput) error
	GetSystem(ctx context.Context, sccSubaccount, locationID, virtualHost string) (*model.Application, error)
	ListBySCC(ctx context.Context, filter *labelfilter.LabelFilter) ([]*model.ApplicationWithLabel, error)
	SetLabel(ctx context.Context, label *model.LabelInput) error
	GetLabel(ctx context.Context, applicationID string, key string) (*model.Label, error)
	ListSCCs(ctx context.Context, key string) ([]*model.SccMetadata, error) //TODO check what tenant will be used to execute this query
}

//go:generate mockery --name=ApplicationConverter --output=automock --outpkg=automock --case=underscore
type ApplicationConverter interface {
	CreateInputJSONToModel(ctx context.Context, in string) (model.ApplicationRegisterInput, error)
}

//go:generate mockery --name=ApplicationTemplateService --output=automock --outpkg=automock --case=underscore
type ApplicationTemplateService interface {
	Get(ctx context.Context, id string) (*model.ApplicationTemplate, error)
	PrepareApplicationCreateInputJSON(appTemplate *model.ApplicationTemplate, values model.ApplicationFromTemplateInputValues) (string, error)
}

//go:generate mockery --name=TenantService --output=automock --outpkg=automock --case=underscore
type TenantService interface {
	ListsByExternalIDs(ctx context.Context, ids []string) ([]*model.BusinessTenantMapping, error)
}

func NewHandler(appSvc ApplicationService, appConverter ApplicationConverter, appTemplateSvc ApplicationTemplateService, tntSvc TenantService, transact persistence.Transactioner) *Handler {
	return &Handler{appSvc: appSvc, appConverter: appConverter, appTemplateSvc: appTemplateSvc, tntSvc: tntSvc, transact: transact}
}

type Handler struct {
	appSvc         ApplicationService
	appConverter   ApplicationConverter
	appTemplateSvc ApplicationTemplateService
	tntSvc         TenantService
	transact       persistence.Transactioner
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
	fmt.Println(">>>>>>>>>>>>>>>>>Start")
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

	reportType := req.URL.Query().Get("reportType")

	if reportType != "full" && reportType != "delta" {
		httputil.RespondWithError(ctx, rw, http.StatusBadRequest, httputil.Error{
			Code:    http.StatusBadRequest,
			Message: "missing or invalid required report type query parameter",
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

	externalSubaccountIDToSCC := make(map[string]*nsmodel.SCC, len(reportData.Value))
	externalIDs := make([]string, 0, len(reportData.Value))
	for _, scc := range reportData.Value {
		if _, ok := externalSubaccountIDToSCC[scc.ExternalSubaccountID]; !ok {
			externalIDs = append(externalIDs, scc.ExternalSubaccountID)
			externalSubaccountIDToSCC[scc.ExternalSubaccountID] = &scc
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
	fmt.Println("-----------------")

	mapExternalToInternal(ctx, tenants, externalSubaccountIDToSCC)
	fmt.Println("-----------------")

	details := make([]httputil.Detail, 0, 0)
	//TODO extract into separate method
	sccs := make([]*nsmodel.SCC, 0, len(reportData.Value))
	for _, scc := range reportData.Value {
		sccWithInternalID, ok := externalSubaccountIDToSCC[scc.ExternalSubaccountID]
		if !ok || sccWithInternalID.InternalSubaccountID == "" {
			log.C(ctx).Warnf("Got SCC with subaccount id: %s which is not found", scc.ExternalSubaccountID)
			addErrorDetailsMsg(&details, &scc, "Subaccount not found")
		} else if sccWithInternalID.InternalSubaccountID == "not subaccount" {
			addErrorDetailsMsg(&details, &scc, "Provided id is not subaccount")
		} else {
			sccs = append(sccs, sccWithInternalID)
		}
	}

	for _, scc := range sccs {
		fmt.Printf(">>>>>>>>>>>>>>>>>> systems: %+v", scc.ExposedSystems)
	}
	if reportType == "delta" {
		a.processDelta(ctx, sccs, &details)
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
		a.processDelta(ctx, sccs, &details)
		a.handleUnreachableScc(ctx, reportData)
		httputils.RespondWithBody(ctx, rw, http.StatusNoContent, struct{}{})
		return
	}
}

func (a *Handler) listTenantsByExternalIDs(ctx context.Context, ids []string) ([]*model.BusinessTenantMapping, error) {
	if len(ids) == 0 {
		return make([]*model.BusinessTenantMapping, 0, 0), nil
	}

	tx, err := a.transact.Begin()
	if err != nil {
		log.C(ctx).Warn(errors.Wrapf(err, "while openning transaction"))
		return nil, err
	}
	ctxWithTransaction := persistence.SaveToContext(ctx, tx)

	tenants, err := a.tntSvc.ListsByExternalIDs(ctxWithTransaction, ids)
	if err != nil {
		log.C(ctx).Warn(errors.Wrapf(err, "while listing tenants by external ids"))
		return nil, err
	}

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
			Subaccount: scc.ExternalSubaccountID,
			LocationId: scc.LocationID,
		})
	}

	sccsToMarkAsUnreachable := difference(sccs, sccsFromNs)

	externalSubaccountToSCC := make(map[string]*model.SccMetadata)
	externalSubaccounts := make([]string, 0, len(sccsToMarkAsUnreachable))
	for _, scc := range sccsToMarkAsUnreachable {
		externalSubaccounts = append(externalSubaccounts, scc.Subaccount)
		externalSubaccountToSCC[scc.Subaccount] = scc
	}

	internalSubaccounts, err := a.listTenantsByExternalIDs(ctx, externalSubaccounts)
	if err != nil {
		log.C(ctx).Warn(errors.Wrapf(err, "while listing subaccounts"))
		return
	}

	for _, subaccount := range internalSubaccounts {
		externalSubaccountToSCC[subaccount.ExternalTenant].InternalSubaccountID = subaccount.ID
	}

	fmt.Println("+++++++++++++++")

	for _, scc := range sccsToMarkAsUnreachable {
		//TODO use internal ID
		ctxWithSubaccount := tenant.SaveToContext(ctx, scc.InternalSubaccountID, "")
		appsWithLabels, ok := a.listAppsByScc(ctxWithSubaccount, scc.Subaccount, scc.LocationId)
		if ok {
			for _, appWithLabels := range appsWithLabels {
				a.markSystemAsUnreachable(ctxWithSubaccount, appWithLabels.App)
			}
		}
	}
}

func (a *Handler) processDelta(ctx context.Context, sccs []*nsmodel.SCC, details *[]httputil.Detail) {
	//TODO refactor details
	for _, scc := range sccs {
		fmt.Println(">>>>>>>> tenant: ", scc.InternalSubaccountID)
		ctxWithTenant := tenant.SaveToContext(ctx, scc.InternalSubaccountID, scc.ExternalSubaccountID)
		if ok := a.handleSccSystems(ctxWithTenant, *scc); !ok {
			addErrorDetails(details, scc)
		}
	}
}

func (a *Handler) handleSccSystems(ctx context.Context, scc nsmodel.SCC) bool {
	successfulUpsert := a.upsertSccSystems(ctx, scc)
	successfulMark := a.markAsUnreachable(ctx, scc)
	return successfulUpsert && successfulMark
}

func (a *Handler) upsertSccSystems(ctx context.Context, scc nsmodel.SCC) bool {
	fmt.Println("<<<<<<<<<<< upsertSystems")
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
				log.C(ctx).Warn(errors.Wrapf(err, "while commiting transaction"))
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
			Placeholder: "description",
			Value:       system.Description,
		},
		{
			Placeholder: "subaccount",
			Value:       scc.ExternalSubaccountID, //TODO use internal ID
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

	appInputJson, err := a.appTemplateSvc.PrepareApplicationCreateInputJSON(template, values)
	if err != nil {
		return nil, errors.Wrapf(err, "while preparing application create input from template with id:%s", system.TemplateID)
	}

	fmt.Println("AppInputJson: ", appInputJson)
	appInput, err := a.appConverter.CreateInputJSONToModel(ctx, appInputJson)
	if err != nil {
		return nil, errors.Wrapf(err, "while preparing application create input from json")
	}

	fmt.Println("AppInputLabels: ", appInput.Labels)
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
	fmt.Println("<<<<<<<<<<< upsert")
	app, err := a.appSvc.GetSystem(ctx, scc.ExternalSubaccountID, scc.LocationID, system.Host)

	if err != nil && isNotFoundError(err) {
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
	//TODO use internal ID
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

	apps, err := a.appSvc.ListBySCC(ctxWithTransaction, labelfilter.NewForKeyWithQuery("scc", fmt.Sprintf("{\"locationId\":\"%s\", \"subaccount\":\"%s\"}", locationID, subaccount)))
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
	addErrorDetailsMsg(details, scc, "Creation failed")
}

func addErrorDetailsMsg(details *[]httputil.Detail, scc *nsmodel.SCC, message string) {
	*details = append(*details, httputil.Detail{
		Message:    message,
		Subaccount: scc.ExternalSubaccountID,
		LocationId: scc.LocationID,
	})
}

func mapExternalToInternal(ctx context.Context, tenants []*model.BusinessTenantMapping, externalSubbaccountToSCC map[string]*nsmodel.SCC) {
	for _, t := range tenants {
		if t.Type == "subaccount" {
			externalSubbaccountToSCC[t.ExternalTenant].InternalSubaccountID = t.ID
		} else {
			log.C(ctx).Warnf("Got tenant with id: %s which is not asubaccount", t.ID)
			externalSubbaccountToSCC[t.ExternalTenant].InternalSubaccountID = "not subaccount"
		}
	}
}

func isNotFoundError(err error) bool {
	return strings.Contains(err.Error(), "Object not found")
}
