package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/appregistry/model"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/res"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

//go:generate mockery -name=DirectorClient -output=automock -outpkg=automock -case=underscore
type DirectorClient interface {
	CreateBundle(appID string, in graphql.BundleCreateInput) (string, error)
	GetBundle(appID string, bundleID string) (graphql.BundleExt, error)
	ListBundles(appID string) ([]*graphql.BundleExt, error)
	DeleteBundle(bundleID string) error
	UpdateBundle(bundleID string, in graphql.BundleUpdateInput) error

	CreateAPIDefinition(bundleID string, apiDefinitionInput graphql.APIDefinitionInput) (string, error)
	CreateEventDefinition(bundleID string, eventDefinitionInput graphql.EventDefinitionInput) (string, error)
	CreateDocument(bundleID string, documentInput graphql.DocumentInput) (string, error)
	DeleteAPIDefinition(apiID string) error
	DeleteEventDefinition(eventID string) error
	DeleteDocument(documentID string) error

	SetApplicationLabel(appID string, label graphql.LabelInput) error
}

//go:generate mockery -name=Converter -output=automock -outpkg=automock -case=underscore
type Converter interface {
	DetailsToGraphQLCreateInput(deprecated model.ServiceDetails) (graphql.BundleCreateInput, error)
	GraphQLCreateInputToUpdateInput(in graphql.BundleCreateInput) graphql.BundleUpdateInput
	GraphQLToServiceDetails(converted graphql.BundleExt, legacyServiceReference LegacyServiceReference) (model.ServiceDetails, error)
	ServiceDetailsToService(in model.ServiceDetails, serviceID string) (model.Service, error)
}

//go:generate mockery -name=RequestContextProvider -output=automock -outpkg=automock -case=underscore
type RequestContextProvider interface {
	ForRequest(r *http.Request) (RequestContext, error)
}

//go:generate mockery -name=Validator -output=automock -outpkg=automock -case=underscore
type Validator interface {
	Validate(details model.ServiceDetails) apperrors.AppError
}

//go:generate mockery -name=AppLabeler -output=automock -outpkg=automock -case=underscore
type AppLabeler interface {
	WriteServiceReference(appLabels graphql.Labels, serviceReference LegacyServiceReference) (graphql.LabelInput, error)
	DeleteServiceReference(appLabels graphql.Labels, serviceID string) (graphql.LabelInput, error)
	ReadServiceReference(appLabels graphql.Labels, serviceID string) (LegacyServiceReference, error)
	ListServiceReferences(appLabels graphql.Labels) ([]LegacyServiceReference, error)
}

const serviceIDVarKey = "serviceId"

type Handler struct {
	logger             *log.Logger
	validator          Validator
	converter          Converter
	reqContextProvider RequestContextProvider
	appLabeler         AppLabeler
}

func NewHandler(converter Converter, validator Validator, reqContextProvider RequestContextProvider, logger *log.Logger, appLabeler AppLabeler) *Handler {
	return &Handler{
		converter:          converter,
		validator:          validator,
		reqContextProvider: reqContextProvider,
		logger:             logger,
		appLabeler:         appLabeler,
	}
}

func (h *Handler) Create(writer http.ResponseWriter, request *http.Request) {
	defer h.closeBody(request)

	serviceDetails, err := h.decodeAndValidateInput(request)
	if err != nil {
		h.logger.Error(err)
		res.WriteAppError(writer, err)
		return
	}

	converted, err := h.converter.DetailsToGraphQLCreateInput(serviceDetails)
	if err != nil {
		wrappedErr := errors.Wrap(err, "while converting service input")
		h.logger.Error(wrappedErr)
		res.WriteError(writer, wrappedErr, apperrors.CodeInternal)
		return
	}

	reqContext, err := h.loadRequestContext(request)
	if err != nil {
		h.writeErrorInternal(writer, err)
		return
	}

	if err := h.ensureUniqueIdentifier(serviceDetails.Identifier, reqContext); err != nil {
		h.logger.Error(errors.Wrap(err, "while ensuring legacy service identifier is unique"))
		res.WriteAppError(writer, err)
		return
	}

	h.logger.Infoln("doing GraphQL request...")

	appID := reqContext.AppID
	serviceID, err := reqContext.DirectorClient.CreateBundle(appID, converted)
	if err != nil {
		wrappedErr := errors.Wrap(err, "while creating Service")
		h.logger.Error(wrappedErr)
		res.WriteError(writer, wrappedErr, apperrors.CodeInternal)
		return
	}

	legacyServiceRef := LegacyServiceReference{
		ID:         serviceID,
		Identifier: serviceDetails.Identifier,
	}

	err = h.setAppLabelWithServiceRef(legacyServiceRef, reqContext)
	if err != nil {
		wrappedErr := errors.Wrap(err, "while setting Application label with legacy service metadata")
		h.logger.Error(wrappedErr)
		res.WriteError(writer, wrappedErr, apperrors.CodeInternal)
		return
	}

	successResponse := SuccessfulCreateResponse{
		ID: serviceID,
	}

	err = res.WriteJSONResponse(writer, &successResponse)
	if err != nil {
		wrappedErr := errors.Wrap(err, "while encoding response")
		h.logger.Error(wrappedErr)
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

	h.getAndWriteServiceByID(writer, serviceID, reqContext)
}

func (h *Handler) List(writer http.ResponseWriter, request *http.Request) {
	defer h.closeBody(request)

	reqContext, err := h.reqContextProvider.ForRequest(request)
	if err != nil {
		wrappedErr := errors.Wrap(err, "while requesting Request Context")
		res.WriteError(writer, wrappedErr, apperrors.CodeInternal)
		return
	}

	appID := reqContext.AppID
	bundles, err := reqContext.DirectorClient.ListBundles(appID)
	if err != nil {
		wrappedErr := errors.Wrap(err, "while fetching Services")
		h.logger.Error(wrappedErr)
		res.WriteError(writer, wrappedErr, apperrors.CodeInternal)
		return
	}

	services := make([]model.Service, 0)
	for _, bundle := range bundles {
		if bundle == nil {
			continue
		}

		legacyServiceReference, err := h.appLabeler.ReadServiceReference(reqContext.AppLabels, bundle.ID)
		if err != nil {
			wrappedErr := errors.Wrapf(err, "while reading legacy service reference for Bundle with ID '%s'", bundle.ID)
			h.logger.Error(wrappedErr)
			res.WriteError(writer, wrappedErr, apperrors.CodeInternal)
			return
		}

		detailedService, err := h.converter.GraphQLToServiceDetails(*bundle, legacyServiceReference)
		if err != nil {
			wrappedErr := errors.Wrap(err, "while converting graphql to detailed service")
			h.logger.Error(wrappedErr)
			res.WriteError(writer, wrappedErr, apperrors.CodeInternal)
			return
		}

		service, err := h.converter.ServiceDetailsToService(detailedService, bundle.ID)
		if err != nil {
			wrappedErr := errors.Wrap(err, "while converting detailed service to service")
			h.logger.Error(wrappedErr)
			res.WriteError(writer, wrappedErr, apperrors.CodeInternal)
			return
		}

		services = append(services, service)
	}
	err = res.WriteJSONResponse(writer, &services)
	if err != nil {
		wrappedErr := errors.Wrap(err, "while encoding response")
		h.logger.Error(wrappedErr)
		res.WriteError(writer, wrappedErr, apperrors.CodeInternal)
		return
	}
}

func (h *Handler) Update(writer http.ResponseWriter, request *http.Request) {
	defer h.closeBody(request)

	id := h.getServiceID(request)

	serviceDetails, err := h.decodeAndValidateInput(request)
	if err != nil {
		h.logger.Error(err)
		res.WriteAppError(writer, err)
		return
	}

	createInput, err := h.converter.DetailsToGraphQLCreateInput(serviceDetails)
	if err != nil {
		wrappedErr := errors.Wrap(err, "while converting service input")
		h.logger.Error(wrappedErr)
		res.WriteError(writer, wrappedErr, apperrors.CodeInternal)
		return
	}

	reqContext, err := h.loadRequestContext(request)
	if err != nil {
		h.logger.Error(err)
		h.writeErrorInternal(writer, err)
		return
	}
	dirCli := reqContext.DirectorClient

	previousBundle, err := dirCli.GetBundle(reqContext.AppID, id)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			h.writeErrorNotFound(writer, id)
			return
		}
		wrappedErr := errors.Wrap(err, "while fetching service")
		h.logger.Error(wrappedErr)
		res.WriteError(writer, wrappedErr, apperrors.CodeInternal)
		return
	}

	bundleID := previousBundle.ID

	err = dirCli.UpdateBundle(id, h.converter.GraphQLCreateInputToUpdateInput(createInput))
	if err != nil {
		wrappedErr := errors.Wrap(err, "while updating Service")
		h.logger.WithField("ID", id).Error(wrappedErr)
		res.WriteError(writer, wrappedErr, apperrors.CodeInternal)
		return
	}

	err = h.deleteRelatedObjectsForBundle(dirCli, previousBundle)
	if err != nil {
		wrappedErr := errors.Wrap(err, "while deleting related objects for Service")
		h.logger.WithField("ID", id).Error(wrappedErr)
		res.WriteError(writer, wrappedErr, apperrors.CodeInternal)
		return
	}

	err = h.createRelatedObjectsForBundle(dirCli, bundleID, createInput)
	if err != nil {
		wrappedErr := errors.Wrap(err, "while creating related objects for Service")
		h.logger.WithField("ID", id).Error(wrappedErr)
		res.WriteError(writer, wrappedErr, apperrors.CodeInternal)
		return
	}

	// Legacy service reference should be updated, but right now it contains only identifier field
	// which has to preserved during update (to match old metadata service behaviour), so there's nothing to update.

	h.getAndWriteServiceByID(writer, bundleID, reqContext)
}

func (h *Handler) Delete(writer http.ResponseWriter, request *http.Request) {
	defer h.closeBody(request)

	id := h.getServiceID(request)

	reqContext, err := h.loadRequestContext(request)
	if err != nil {
		h.writeErrorInternal(writer, err)
		return
	}

	err = reqContext.DirectorClient.DeleteBundle(id)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			h.writeErrorNotFound(writer, id)
			return
		}

		h.logger.WithField("ID", id).Error(errors.Wrap(err, "while deleting service"))
		res.WriteError(writer, err, apperrors.CodeInternal)
		return
	}

	label, err := h.appLabeler.DeleteServiceReference(reqContext.AppLabels, id)
	if err != nil {
		wrappedError := errors.Wrap(err, "while writing Application label")
		h.logger.Error(wrappedError)
		res.WriteError(writer, wrappedError, apperrors.CodeInternal)
		return
	}

	err = reqContext.DirectorClient.SetApplicationLabel(reqContext.AppID, label)
	if err != nil {
		wrappedError := errors.Wrap(err, "while setting Application label")
		h.logger.Error(wrappedError)
		res.WriteError(writer, wrappedError, apperrors.CodeInternal)
		return
	}

	writer.WriteHeader(http.StatusNoContent)
}

func (h *Handler) closeBody(rq *http.Request) {
	err := rq.Body.Close()
	if err != nil {
		h.logger.Error(errors.Wrap(err, "while closing body"))
	}
}

func (h *Handler) decodeAndValidateInput(request *http.Request) (model.ServiceDetails, error) {
	var details model.ServiceDetails

	err := json.NewDecoder(request.Body).Decode(&details)
	if err != nil {
		wrappedErr := errors.Wrap(err, "while unmarshalling service")
		h.logger.Error(wrappedErr)

		appErr := apperrors.WrongInput(wrappedErr.Error())
		return model.ServiceDetails{}, appErr
	}

	h.logger.Infoln("body decoded. validating...")

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
		h.logger.Error(wrappedErr)
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

func (h *Handler) createRelatedObjectsForBundle(dirCli DirectorClient, bundleID string, bundle graphql.BundleCreateInput) error {
	for _, apiDef := range bundle.APIDefinitions {
		if apiDef == nil {
			continue
		}

		_, err := dirCli.CreateAPIDefinition(bundleID, *apiDef)
		if err != nil {
			return errors.Wrap(err, "while creating API Definition")
		}
	}

	for _, eventDef := range bundle.EventDefinitions {
		if eventDef == nil {
			continue
		}

		_, err := dirCli.CreateEventDefinition(bundleID, *eventDef)
		if err != nil {
			return errors.Wrap(err, "while creating Event Definition")
		}
	}

	for _, doc := range bundle.Documents {
		if doc == nil {
			continue
		}

		_, err := dirCli.CreateDocument(bundleID, *doc)
		if err != nil {
			return errors.Wrap(err, "while creating Document")
		}
	}

	return nil
}

func (h *Handler) deleteRelatedObjectsForBundle(dirCli DirectorClient, bundle graphql.BundleExt) error {
	for _, apiDef := range bundle.APIDefinitions.Data {
		if apiDef == nil {
			continue
		}

		err := dirCli.DeleteAPIDefinition(apiDef.ID)
		if err != nil {
			return errors.Wrapf(err, "while deleting API Definition with ID '%s'", apiDef.ID)
		}
	}

	for _, eventDef := range bundle.EventDefinitions.Data {
		if eventDef == nil {
			continue
		}

		err := dirCli.DeleteEventDefinition(eventDef.ID)
		if err != nil {
			return errors.Wrapf(err, "while deleting Event Definition with ID '%s'", eventDef.ID)
		}
	}

	for _, doc := range bundle.Documents.Data {
		if doc == nil {
			continue
		}

		err := dirCli.DeleteDocument(doc.ID)
		if err != nil {
			return errors.Wrapf(err, "while deleting Document with ID '%s'", doc.ID)
		}
	}

	return nil
}

func (h *Handler) getAndWriteServiceByID(writer http.ResponseWriter, serviceID string, reqContext RequestContext) {
	output, err := reqContext.DirectorClient.GetBundle(reqContext.AppID, serviceID)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			h.writeErrorNotFound(writer, serviceID)
			return
		}
		wrappedErr := errors.Wrap(err, "while fetching service")
		h.logger.Error(wrappedErr)
		res.WriteError(writer, wrappedErr, apperrors.CodeInternal)
		return
	}

	legacyServiceReference, err := h.appLabeler.ReadServiceReference(reqContext.AppLabels, serviceID)
	if err != nil {
		wrappedErr := errors.Wrapf(err, "while reading legacy service reference for Bundle with ID '%s'", serviceID)
		h.logger.Error(wrappedErr)
		res.WriteError(writer, wrappedErr, apperrors.CodeInternal)
		return
	}

	service, err := h.converter.GraphQLToServiceDetails(output, legacyServiceReference)
	if err != nil {
		wrappedErr := errors.Wrap(err, "while converting service")
		h.logger.Error(wrappedErr)
		res.WriteError(writer, wrappedErr, apperrors.CodeInternal)
		return
	}

	err = res.WriteJSONResponse(writer, &service)
	if err != nil {
		wrappedErr := errors.Wrap(err, "while encoding response")
		h.logger.Error(wrappedErr)
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

func (h *Handler) setAppLabelWithServiceRef(serviceRef LegacyServiceReference, reqContext RequestContext) error {
	label, err := h.appLabeler.WriteServiceReference(reqContext.AppLabels, serviceRef)
	if err != nil {
		return errors.Wrap(err, "while writing Application label")
	}

	err = reqContext.DirectorClient.SetApplicationLabel(reqContext.AppID, label)
	if err != nil {
		return errors.Wrap(err, "while setting Application label")
	}

	return nil
}
