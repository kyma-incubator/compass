package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/model"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/res"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

//go:generate mockery --name=DirectorClient --output=automock --outpkg=automock --case=underscore --disable-version-string
type DirectorClient interface {
	CreateBundle(ctx context.Context, appID string, in graphql.BundleCreateInput) (string, error)
	GetBundle(ctx context.Context, appID string, bundleID string) (graphql.BundleExt, error)
	ListBundles(ctx context.Context, appID string) ([]*graphql.BundleExt, error)
	DeleteBundle(ctx context.Context, bundleID string) error
	UpdateBundle(ctx context.Context, bundleID string, in graphql.BundleUpdateInput) error

	CreateAPIDefinition(ctx context.Context, bundleID string, apiDefinitionInput graphql.APIDefinitionInput) (string, error)
	CreateEventDefinition(ctx context.Context, bundleID string, eventDefinitionInput graphql.EventDefinitionInput) (string, error)
	CreateDocument(ctx context.Context, bundleID string, documentInput graphql.DocumentInput) (string, error)
	DeleteAPIDefinition(ctx context.Context, apiID string) error
	DeleteEventDefinition(ctx context.Context, eventID string) error
	DeleteDocument(ctx context.Context, documentID string) error

	SetApplicationLabel(ctx context.Context, appID string, label graphql.LabelInput) error
}

//go:generate mockery --name=Converter --output=automock --outpkg=automock --case=underscore --disable-version-string
type Converter interface {
	DetailsToGraphQLCreateInput(deprecated model.ServiceDetails) (graphql.BundleCreateInput, error)
	GraphQLCreateInputToUpdateInput(in graphql.BundleCreateInput) graphql.BundleUpdateInput
	GraphQLToServiceDetails(converted graphql.BundleExt, legacyServiceReference LegacyServiceReference) (model.ServiceDetails, error)
	ServiceDetailsToService(in model.ServiceDetails, serviceID string) (model.Service, error)
}

//go:generate mockery --name=RequestContextProvider --output=automock --outpkg=automock --case=underscore --disable-version-string
type RequestContextProvider interface {
	ForRequest(r *http.Request) (RequestContext, error)
}

//go:generate mockery --name=Validator --output=automock --outpkg=automock --case=underscore --disable-version-string
type Validator interface {
	Validate(details model.ServiceDetails) apperrors.AppError
}

//go:generate mockery --name=AppLabeler --output=automock --outpkg=automock --case=underscore --disable-version-string
type AppLabeler interface {
	WriteServiceReference(appLabels graphql.Labels, serviceReference LegacyServiceReference) (graphql.LabelInput, error)
	DeleteServiceReference(appLabels graphql.Labels, serviceID string) (graphql.LabelInput, error)
	ReadServiceReference(appLabels graphql.Labels, serviceID string) (LegacyServiceReference, error)
	ListServiceReferences(appLabels graphql.Labels) ([]LegacyServiceReference, error)
}

const serviceIDVarKey = "serviceId"

type Handler struct {
	validator          Validator
	converter          Converter
	reqContextProvider RequestContextProvider
	appLabeler         AppLabeler
}

func NewHandler(converter Converter, validator Validator, reqContextProvider RequestContextProvider, appLabeler AppLabeler) *Handler {
	return &Handler{
		converter:          converter,
		validator:          validator,
		reqContextProvider: reqContextProvider,
		appLabeler:         appLabeler,
	}
}

func (h *Handler) Create(writer http.ResponseWriter, request *http.Request) {
	defer h.closeBody(request)

	logger := log.C(request.Context())
	serviceDetails, err := h.decodeAndValidateInput(request)
	if err != nil {
		logger.Error(err)
		res.WriteAppError(writer, err)
		return
	}

	converted, err := h.converter.DetailsToGraphQLCreateInput(serviceDetails)
	if err != nil {
		wrappedErr := errors.Wrap(err, "while converting service input")
		logger.Error(wrappedErr)
		res.WriteError(writer, wrappedErr, apperrors.CodeInternal)
		return
	}

	reqContext, err := h.loadRequestContext(request)
	if err != nil {
		h.writeErrorInternal(writer, err)
		return
	}

	if err := h.ensureUniqueIdentifier(serviceDetails.Identifier, reqContext); err != nil {
		logger.Error(errors.Wrap(err, "while ensuring legacy service identifier is unique"))
		res.WriteAppError(writer, err)
		return
	}

	logger.Infoln("doing GraphQL request...")

	appID := reqContext.AppID
	serviceID, err := reqContext.DirectorClient.CreateBundle(request.Context(), appID, converted)
	if err != nil {
		wrappedErr := errors.Wrap(err, "while creating Service")
		logger.Error(wrappedErr)
		res.WriteError(writer, wrappedErr, apperrors.CodeInternal)
		return
	}

	legacyServiceRef := LegacyServiceReference{
		ID:         serviceID,
		Identifier: serviceDetails.Identifier,
	}

	err = h.setAppLabelWithServiceRef(request.Context(), legacyServiceRef, reqContext)
	if err != nil {
		wrappedErr := errors.Wrap(err, "while setting Application label with legacy service metadata")
		logger.Error(wrappedErr)
		res.WriteError(writer, wrappedErr, apperrors.CodeInternal)
		return
	}

	successResponse := SuccessfulCreateResponse{
		ID: serviceID,
	}

	err = res.WriteJSONResponse(writer, request.Context(), &successResponse)
	if err != nil {
		wrappedErr := errors.Wrap(err, "while encoding response")
		logger.Error(wrappedErr)
		res.WriteError(writer, wrappedErr, apperrors.CodeInternal)
		return
	}
}

type SuccessfulCreateResponse struct {
	ID string `json:"id"`
}

func (h *Handler) Get(writer http.ResponseWriter, request *http.Request) {
	defer h.closeBody(request)
	serviceID := h.getServiceID(request)

	reqContext, err := h.reqContextProvider.ForRequest(request)
	if err != nil {
		wrappedErr := errors.Wrap(err, "while requesting Request Context")
		res.WriteError(writer, wrappedErr, apperrors.CodeInternal)
		return
	}

	h.getAndWriteServiceByID(request.Context(), writer, serviceID, reqContext)
}

func (h *Handler) List(writer http.ResponseWriter, request *http.Request) {
	defer h.closeBody(request)

	logger := log.C(request.Context())
	reqContext, err := h.reqContextProvider.ForRequest(request)
	if err != nil {
		wrappedErr := errors.Wrap(err, "while requesting Request Context")
		res.WriteError(writer, wrappedErr, apperrors.CodeInternal)
		return
	}

	appID := reqContext.AppID
	bundles, err := reqContext.DirectorClient.ListBundles(request.Context(), appID)
	if err != nil {
		wrappedErr := errors.Wrap(err, "while fetching Services")
		logger.Error(wrappedErr)
		res.WriteError(writer, wrappedErr, apperrors.CodeInternal)
		return
	}

	services := make([]model.Service, 0)
	for _, bndl := range bundles {
		if bndl == nil {
			continue
		}

		legacyServiceReference, err := h.appLabeler.ReadServiceReference(reqContext.AppLabels, bndl.ID)
		if err != nil {
			wrappedErr := errors.Wrapf(err, "while reading legacy service reference for Bundle with ID '%s'", bndl.ID)
			logger.Error(wrappedErr)
			res.WriteError(writer, wrappedErr, apperrors.CodeInternal)
			return
		}

		detailedService, err := h.converter.GraphQLToServiceDetails(*bndl, legacyServiceReference)
		if err != nil {
			wrappedErr := errors.Wrap(err, "while converting graphql to detailed service")
			logger.Error(wrappedErr)
			res.WriteError(writer, wrappedErr, apperrors.CodeInternal)
			return
		}

		service, err := h.converter.ServiceDetailsToService(detailedService, bndl.ID)
		if err != nil {
			wrappedErr := errors.Wrap(err, "while converting detailed service to service")
			logger.Error(wrappedErr)
			res.WriteError(writer, wrappedErr, apperrors.CodeInternal)
			return
		}

		services = append(services, service)
	}
	err = res.WriteJSONResponse(writer, request.Context(), &services)
	if err != nil {
		wrappedErr := errors.Wrap(err, "while encoding response")
		logger.Error(wrappedErr)
		res.WriteError(writer, wrappedErr, apperrors.CodeInternal)
		return
	}
}

func (h *Handler) Update(writer http.ResponseWriter, request *http.Request) {
	defer h.closeBody(request)

	logger := log.C(request.Context())
	id := h.getServiceID(request)

	serviceDetails, err := h.decodeAndValidateInput(request)
	if err != nil {
		logger.Error(err)
		res.WriteAppError(writer, err)
		return
	}

	createInput, err := h.converter.DetailsToGraphQLCreateInput(serviceDetails)
	if err != nil {
		wrappedErr := errors.Wrap(err, "while converting service input")
		logger.Error(wrappedErr)
		res.WriteError(writer, wrappedErr, apperrors.CodeInternal)
		return
	}

	reqContext, err := h.loadRequestContext(request)
	if err != nil {
		logger.Error(err)
		h.writeErrorInternal(writer, err)
		return
	}
	dirCli := reqContext.DirectorClient

	previousBundle, err := dirCli.GetBundle(request.Context(), reqContext.AppID, id)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			h.writeErrorNotFound(writer, id)
			return
		}
		wrappedErr := errors.Wrap(err, "while fetching service")
		logger.Error(wrappedErr)
		res.WriteError(writer, wrappedErr, apperrors.CodeInternal)
		return
	}

	bndlID := previousBundle.BaseEntity.ID

	err = dirCli.UpdateBundle(request.Context(), id, h.converter.GraphQLCreateInputToUpdateInput(createInput))
	if err != nil {
		wrappedErr := errors.Wrap(err, "while updating Service")
		logger.WithField("ID", id).Error(wrappedErr)
		res.WriteError(writer, wrappedErr, apperrors.CodeInternal)
		return
	}

	err = h.deleteRelatedObjectsForBundle(request.Context(), dirCli, previousBundle)
	if err != nil {
		wrappedErr := errors.Wrap(err, "while deleting related objects for Service")
		logger.WithField("ID", id).Error(wrappedErr)
		res.WriteError(writer, wrappedErr, apperrors.CodeInternal)
		return
	}

	err = h.createRelatedObjectsForBundle(request.Context(), dirCli, bndlID, createInput)
	if err != nil {
		wrappedErr := errors.Wrap(err, "while creating related objects for Service")
		logger.WithField("ID", id).Error(wrappedErr)
		res.WriteError(writer, wrappedErr, apperrors.CodeInternal)
		return
	}

	// Legacy service reference should be updated, but right now it contains only identifier field
	// which has to preserved during update (to match old metadata service behaviour), so there's nothing to update.

	h.getAndWriteServiceByID(request.Context(), writer, bndlID, reqContext)
}

func (h *Handler) Delete(writer http.ResponseWriter, request *http.Request) {
	defer h.closeBody(request)

	logger := log.C(request.Context())
	id := h.getServiceID(request)

	reqContext, err := h.loadRequestContext(request)
	if err != nil {
		h.writeErrorInternal(writer, err)
		return
	}

	err = reqContext.DirectorClient.DeleteBundle(request.Context(), id)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			h.writeErrorNotFound(writer, id)
			return
		}

		logger.WithField("ID", id).Error(errors.Wrap(err, "while deleting service"))
		res.WriteError(writer, err, apperrors.CodeInternal)
		return
	}

	label, err := h.appLabeler.DeleteServiceReference(reqContext.AppLabels, id)
	if err != nil {
		wrappedError := errors.Wrap(err, "while writing Application label")
		logger.Error(wrappedError)
		res.WriteError(writer, wrappedError, apperrors.CodeInternal)
		return
	}

	err = reqContext.DirectorClient.SetApplicationLabel(request.Context(), reqContext.AppID, label)
	if err != nil {
		wrappedError := errors.Wrap(err, "while setting Application label")
		logger.Error(wrappedError)
		res.WriteError(writer, wrappedError, apperrors.CodeInternal)
		return
	}

	writer.WriteHeader(http.StatusNoContent)
}

func (h *Handler) closeBody(rq *http.Request) {
	err := rq.Body.Close()
	if err != nil {
		log.C(rq.Context()).Error(errors.Wrap(err, "while closing body"))
	}
}

func (h *Handler) decodeAndValidateInput(request *http.Request) (model.ServiceDetails, error) {
	var details model.ServiceDetails
	logger := log.C(request.Context())

	err := json.NewDecoder(request.Body).Decode(&details)
	if err != nil {
		wrappedErr := errors.Wrap(err, "while unmarshalling service")
		logger.Error(wrappedErr)

		appErr := apperrors.WrongInput(wrappedErr.Error())
		return model.ServiceDetails{}, appErr
	}

	logger.Infoln("body decoded. validating...")

	appErr := h.validator.Validate(details)
	if appErr != nil {
		wrappedAppErr := appErr.Append("while validating input")
		return model.ServiceDetails{}, wrappedAppErr
	}

	return details, nil
}

func (h *Handler) loadRequestContext(request *http.Request) (RequestContext, error) {
	reqContext, err := h.reqContextProvider.ForRequest(request)
	if err != nil {
		wrappedErr := errors.Wrap(err, "while requesting Request Context")
		log.C(request.Context()).Error(wrappedErr)
		return RequestContext{}, wrappedErr
	}

	return reqContext, nil
}

func (h *Handler) writeErrorNotFound(writer http.ResponseWriter, id string) {
	message := fmt.Sprintf("entity with ID %s not found", id)
	res.WriteErrorMessage(writer, message, apperrors.CodeNotFound)
}

func (h *Handler) writeErrorInternal(writer http.ResponseWriter, err error) {
	res.WriteError(writer, err, apperrors.CodeInternal)
}

func (h *Handler) getServiceID(request *http.Request) string {
	vars := mux.Vars(request)
	id := vars[serviceIDVarKey]
	return id
}

func (h *Handler) createRelatedObjectsForBundle(ctx context.Context, dirCli DirectorClient, bndlID string, bndl graphql.BundleCreateInput) error {
	for _, apiDef := range bndl.APIDefinitions {
		if apiDef == nil {
			continue
		}

		_, err := dirCli.CreateAPIDefinition(ctx, bndlID, *apiDef)
		if err != nil {
			return errors.Wrap(err, "while creating API Definition")
		}
	}

	for _, eventDef := range bndl.EventDefinitions {
		if eventDef == nil {
			continue
		}

		_, err := dirCli.CreateEventDefinition(ctx, bndlID, *eventDef)
		if err != nil {
			return errors.Wrap(err, "while creating Event Definition")
		}
	}

	for _, doc := range bndl.Documents {
		if doc == nil {
			continue
		}

		_, err := dirCli.CreateDocument(ctx, bndlID, *doc)
		if err != nil {
			return errors.Wrap(err, "while creating Document")
		}
	}

	return nil
}

func (h *Handler) deleteRelatedObjectsForBundle(ctx context.Context, dirCli DirectorClient, bndl graphql.BundleExt) error {
	for _, apiDef := range bndl.APIDefinitions.Data {
		if apiDef == nil {
			continue
		}

		err := dirCli.DeleteAPIDefinition(ctx, apiDef.ID)
		if err != nil {
			return errors.Wrapf(err, "while deleting API Definition with ID '%s'", apiDef.ID)
		}
	}

	for _, eventDef := range bndl.EventDefinitions.Data {
		if eventDef == nil {
			continue
		}

		err := dirCli.DeleteEventDefinition(ctx, eventDef.ID)
		if err != nil {
			return errors.Wrapf(err, "while deleting Event Definition with ID '%s'", eventDef.ID)
		}
	}

	for _, doc := range bndl.Documents.Data {
		if doc == nil {
			continue
		}

		err := dirCli.DeleteDocument(ctx, doc.ID)
		if err != nil {
			return errors.Wrapf(err, "while deleting Document with ID '%s'", doc.ID)
		}
	}

	return nil
}

func (h *Handler) getAndWriteServiceByID(ctx context.Context, writer http.ResponseWriter, serviceID string, reqContext RequestContext) {
	logger := log.C(ctx)
	output, err := reqContext.DirectorClient.GetBundle(ctx, reqContext.AppID, serviceID)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			h.writeErrorNotFound(writer, serviceID)
			return
		}
		wrappedErr := errors.Wrap(err, "while fetching service")
		logger.Error(wrappedErr)
		res.WriteError(writer, wrappedErr, apperrors.CodeInternal)
		return
	}

	legacyServiceReference, err := h.appLabeler.ReadServiceReference(reqContext.AppLabels, serviceID)
	if err != nil {
		wrappedErr := errors.Wrapf(err, "while reading legacy service reference for Bundle with ID '%s'", serviceID)
		logger.Error(wrappedErr)
		res.WriteError(writer, wrappedErr, apperrors.CodeInternal)
		return
	}

	service, err := h.converter.GraphQLToServiceDetails(output, legacyServiceReference)
	if err != nil {
		wrappedErr := errors.Wrap(err, "while converting service")
		logger.Error(wrappedErr)
		res.WriteError(writer, wrappedErr, apperrors.CodeInternal)
		return
	}

	err = res.WriteJSONResponse(writer, ctx, &service)
	if err != nil {
		wrappedErr := errors.Wrap(err, "while encoding response")
		logger.Error(wrappedErr)
		res.WriteError(writer, wrappedErr, apperrors.CodeInternal)
		return
	}
}

func (h *Handler) ensureUniqueIdentifier(identifier string, reqContext RequestContext) apperrors.AppError {
	if identifier == "" {
		return nil
	}

	services, err := h.appLabeler.ListServiceReferences(reqContext.AppLabels)
	if err != nil {
		wrappedError := errors.Wrapf(err, "while listing legacy services for Application with ID '%s'", reqContext.AppID)
		return apperrors.Internal(wrappedError.Error())
	}

	for _, svc := range services {
		if svc.Identifier == identifier {
			return apperrors.AlreadyExists("Service with Identifier %s already exists", identifier)
		}
	}

	return nil
}

func (h *Handler) setAppLabelWithServiceRef(ctx context.Context, serviceRef LegacyServiceReference, reqContext RequestContext) error {
	label, err := h.appLabeler.WriteServiceReference(reqContext.AppLabels, serviceRef)
	if err != nil {
		return errors.Wrap(err, "while writing Application label")
	}

	err = reqContext.DirectorClient.SetApplicationLabel(ctx, reqContext.AppID, label)
	if err != nil {
		return errors.Wrap(err, "while setting Application label")
	}

	return nil
}
