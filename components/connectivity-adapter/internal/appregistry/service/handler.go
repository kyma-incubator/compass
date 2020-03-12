package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/appregistry/model"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/reqerror"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

//go:generate mockery -name=DirectorClient -output=automock -outpkg=automock -case=underscore
type DirectorClient interface {
	CreatePackage(appID string, in graphql.PackageCreateInput) (string, error)
	GetPackage(appID string, packageID string) (graphql.PackageExt, error)
	ListPackages(appID string) ([]*graphql.PackageExt, error)
	DeletePackage(packageID string) error
	UpdatePackage(packageID string, in graphql.PackageUpdateInput) error

	CreateAPIDefinition(packageID string, apiDefinitionInput graphql.APIDefinitionInput) (string, error)
	CreateEventDefinition(packageID string, eventDefinitionInput graphql.EventDefinitionInput) (string, error)
	CreateDocument(packageID string, documentInput graphql.DocumentInput) (string, error)
	DeleteAPIDefinition(apiID string) error
	DeleteEventDefinition(eventID string) error
	DeleteDocument(documentID string) error
}

//go:generate mockery -name=Converter -output=automock -outpkg=automock -case=underscore
type Converter interface {
	DetailsToGraphQLCreateInput(deprecated model.ServiceDetails) (graphql.PackageCreateInput, error)
	GraphQLCreateInputToUpdateInput(in graphql.PackageCreateInput) graphql.PackageUpdateInput
	GraphQLToServiceDetails(converted graphql.PackageExt) (model.ServiceDetails, error)
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

const serviceIDVarKey = "serviceId"

type Handler struct {
	logger             *log.Logger
	validator          Validator
	converter          Converter
	reqContextProvider RequestContextProvider
}

func NewHandler(converter Converter, validator Validator, reqContextProvider RequestContextProvider, logger *log.Logger) *Handler {
	return &Handler{
		converter:          converter,
		validator:          validator,
		reqContextProvider: reqContextProvider,
		logger:             logger,
	}
}

func (h *Handler) Create(writer http.ResponseWriter, request *http.Request) {
	defer h.closeBody(request)

	serviceDetails, err := h.decodeAndValidateInput(request)
	if err != nil {
		h.logger.Error(err)
		reqerror.WriteAppError(writer, err)
		return
	}

	converted, err := h.converter.DetailsToGraphQLCreateInput(serviceDetails)
	if err != nil {
		wrappedErr := errors.Wrap(err, "while converting service input")
		h.logger.Error(wrappedErr)
		reqerror.WriteError(writer, wrappedErr, apperrors.CodeInternal)
		return
	}

	reqContext, err := h.loadRequestContext(request)
	if err != nil {
		h.writeErrorInternal(writer, err)
		return
	}

	appID := reqContext.AppID
	serviceID, err := reqContext.DirectorClient.CreatePackage(appID, converted)
	if err != nil {
		wrappedErr := errors.Wrap(err, "while creating Service")
		h.logger.Error(wrappedErr)
		reqerror.WriteError(writer, wrappedErr, apperrors.CodeInternal)
		return
	}

	successResponse := SuccessfulCreateResponse{
		ID: serviceID,
	}

	err = json.NewEncoder(writer).Encode(&successResponse)
	if err != nil {
		wrappedErr := errors.Wrap(err, "while encoding response")
		h.logger.Error(wrappedErr)
		reqerror.WriteError(writer, wrappedErr, apperrors.CodeInternal)
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
		reqerror.WriteError(writer, wrappedErr, apperrors.CodeInternal)
		return
	}

	h.getAndWriteServiceByID(writer, serviceID, reqContext)
}

func (h *Handler) List(writer http.ResponseWriter, request *http.Request) {
	defer h.closeBody(request)

	reqContext, err := h.reqContextProvider.ForRequest(request)
	if err != nil {
		wrappedErr := errors.Wrap(err, "while requesting Request Context")
		reqerror.WriteError(writer, wrappedErr, apperrors.CodeInternal)
		return
	}

	appID := reqContext.AppID
	packages, err := reqContext.DirectorClient.ListPackages(appID)
	if err != nil {
		wrappedErr := errors.Wrap(err, "while fetching Services")
		h.logger.Error(wrappedErr)
		reqerror.WriteError(writer, wrappedErr, apperrors.CodeInternal)
		return
	}

	services := make([]model.Service, 0)
	for _, pkg := range packages {
		if pkg == nil {
			continue
		}

		detailedService, err := h.converter.GraphQLToServiceDetails(*pkg)
		if err != nil {
			wrappedErr := errors.Wrap(err, "while converting graphql to detailed service")
			h.logger.Error(wrappedErr)
			reqerror.WriteError(writer, wrappedErr, apperrors.CodeInternal)
			return
		}

		service, err := h.converter.ServiceDetailsToService(detailedService, pkg.ID)
		if err != nil {
			wrappedErr := errors.Wrap(err, "while converting detailed service to service")
			h.logger.Error(wrappedErr)
			reqerror.WriteError(writer, wrappedErr, apperrors.CodeInternal)
			return
		}

		services = append(services, service)
	}
	err = json.NewEncoder(writer).Encode(&services)
	if err != nil {
		wrappedErr := errors.Wrap(err, "while encoding response")
		h.logger.Error(wrappedErr)
		reqerror.WriteError(writer, wrappedErr, apperrors.CodeInternal)
		return
	}
}

func (h *Handler) Update(writer http.ResponseWriter, request *http.Request) {
	defer h.closeBody(request)

	id := h.getServiceID(request)

	serviceDetails, err := h.decodeAndValidateInput(request)
	if err != nil {
		h.logger.Error(err)
		reqerror.WriteAppError(writer, err)
		return
	}

	createInput, err := h.converter.DetailsToGraphQLCreateInput(serviceDetails)
	if err != nil {
		wrappedErr := errors.Wrap(err, "while converting service input")
		h.logger.Error(wrappedErr)
		reqerror.WriteError(writer, wrappedErr, apperrors.CodeInternal)
		return
	}

	reqContext, err := h.loadRequestContext(request)
	if err != nil {
		h.logger.Error(err)
		h.writeErrorInternal(writer, err)
		return
	}
	dirCli := reqContext.DirectorClient

	previousPackage, err := dirCli.GetPackage(reqContext.AppID, id)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			h.writeErrorNotFound(writer, id)
			return
		}
		wrappedErr := errors.Wrap(err, "while fetching service")
		h.logger.Error(wrappedErr)
		reqerror.WriteError(writer, wrappedErr, apperrors.CodeInternal)
		return
	}

	pkgID := previousPackage.ID

	err = dirCli.UpdatePackage(id, h.converter.GraphQLCreateInputToUpdateInput(createInput))
	if err != nil {
		wrappedErr := errors.Wrap(err, "while updating Service")
		h.logger.WithField("ID", id).Error(wrappedErr)
		reqerror.WriteError(writer, wrappedErr, apperrors.CodeInternal)
		return
	}

	err = h.deleteRelatedObjectsForPackage(dirCli, previousPackage)
	if err != nil {
		wrappedErr := errors.Wrap(err, "while deleting related objects for Service")
		h.logger.WithField("ID", id).Error(wrappedErr)
		reqerror.WriteError(writer, wrappedErr, apperrors.CodeInternal)
		return
	}

	err = h.createRelatedObjectsForPackage(dirCli, pkgID, createInput)
	if err != nil {
		wrappedErr := errors.Wrap(err, "while creating related objects for Service")
		h.logger.WithField("ID", id).Error(wrappedErr)
		reqerror.WriteError(writer, wrappedErr, apperrors.CodeInternal)
		return
	}

	h.getAndWriteServiceByID(writer, pkgID, reqContext)
}

func (h *Handler) Delete(writer http.ResponseWriter, request *http.Request) {
	defer h.closeBody(request)

	id := h.getServiceID(request)

	reqContext, err := h.loadRequestContext(request)
	if err != nil {
		h.writeErrorInternal(writer, err)
		return
	}

	err = reqContext.DirectorClient.DeletePackage(id)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			h.writeErrorNotFound(writer, id)
			return
		}

		h.logger.WithField("ID", id).Error(errors.Wrap(err, "while deleting service"))
		reqerror.WriteError(writer, err, apperrors.CodeInternal)
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
	reqerror.WriteErrorMessage(writer, message, apperrors.CodeNotFound)
}

func (h *Handler) writeErrorInternal(writer http.ResponseWriter, err error) {
	reqerror.WriteError(writer, err, apperrors.CodeInternal)
}

func (h *Handler) getServiceID(request *http.Request) string {
	vars := mux.Vars(request)
	id := vars[serviceIDVarKey]
	return id
}

func (h *Handler) createRelatedObjectsForPackage(dirCli DirectorClient, pkgID string, pkg graphql.PackageCreateInput) error {
	for _, apiDef := range pkg.APIDefinitions {
		if apiDef == nil {
			continue
		}

		_, err := dirCli.CreateAPIDefinition(pkgID, *apiDef)
		if err != nil {
			return errors.Wrap(err, "while creating API Definition")
		}
	}

	for _, eventDef := range pkg.EventDefinitions {
		if eventDef == nil {
			continue
		}

		_, err := dirCli.CreateEventDefinition(pkgID, *eventDef)
		if err != nil {
			return errors.Wrap(err, "while creating Event Definition")
		}
	}

	for _, doc := range pkg.Documents {
		if doc == nil {
			continue
		}

		_, err := dirCli.CreateDocument(pkgID, *doc)
		if err != nil {
			return errors.Wrap(err, "while creating Document")
		}
	}

	return nil
}

func (h *Handler) deleteRelatedObjectsForPackage(dirCli DirectorClient, pkg graphql.PackageExt) error {
	for _, apiDef := range pkg.APIDefinitions.Data {
		if apiDef == nil {
			continue
		}

		err := dirCli.DeleteAPIDefinition(apiDef.ID)
		if err != nil {
			return errors.Wrapf(err, "while deleting API Definition with ID '%s'", apiDef.ID)
		}
	}

	for _, eventDef := range pkg.EventDefinitions.Data {
		if eventDef == nil {
			continue
		}

		err := dirCli.DeleteEventDefinition(eventDef.ID)
		if err != nil {
			return errors.Wrapf(err, "while deleting Event Definition with ID '%s'", eventDef.ID)
		}
	}

	for _, doc := range pkg.Documents.Data {
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
	output, err := reqContext.DirectorClient.GetPackage(reqContext.AppID, serviceID)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			h.writeErrorNotFound(writer, serviceID)
			return
		}
		wrappedErr := errors.Wrap(err, "while fetching service")
		h.logger.Error(wrappedErr)
		reqerror.WriteError(writer, wrappedErr, apperrors.CodeInternal)
		return
	}

	service, err := h.converter.GraphQLToServiceDetails(output)
	if err != nil {
		wrappedErr := errors.Wrap(err, "while converting service")
		h.logger.Error(wrappedErr)
		reqerror.WriteError(writer, wrappedErr, apperrors.CodeInternal)
		return
	}

	err = json.NewEncoder(writer).Encode(&service)
	if err != nil {
		wrappedErr := errors.Wrap(err, "while encoding response")
		h.logger.Error(wrappedErr)
		reqerror.WriteError(writer, wrappedErr, apperrors.CodeInternal)
		return
	}
}
