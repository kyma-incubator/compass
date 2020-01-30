package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/appregistry/appdetails"

	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/appregistry/model"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/reqerror"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

//go:generate mockery -name=Converter -output=automock -outpkg=automock -case=underscore
type Converter interface {
	DetailsToGraphQLInput(id string, deprecated model.ServiceDetails) (model.GraphQLServiceDetailsInput, error)
	GraphQLToServiceDetails(converted model.GraphQLServiceDetails) (model.ServiceDetails, error)
}

////go:generate mockery -name=AppOperator -output=automock -outpkg=automock -case=underscore
//type AppOperator interface {
//	SaveServiceInLabels(appID string, details ConvertedServiceDetails) (graphql.Labels, error) // add or replace service
//	GetServiceFromLabels(serviceID string, labels graphql.Labels) (ConvertedServiceDetails, error)
//	ListServicesFromLabels(labels graphql.Labels) ([]ConvertedServiceDetails, error)
//}
//
////go:generate mockery -name=DirectorClient -output=automock -outpkg=automock -case=underscore
//type DirectorClient interface {
//	//SetApplicationLegacyServicesLabel(id string, legacyServices []ConvertedServiceDetails) error
//
//	CreateAPIDefinition(appID string, apiDefinitionInput graphql.APIDefinitionInput) (string, error)
//	CreateEventDefinition(appID string, eventDefinitionInput graphql.EventDefinitionInput) (string, error)
//
//	DeleteAPIDefinition(apiID string) (string, error)
//	DeleteEventDefinition(eventID string) (string, error)
//}

//go:generate mockery -name=ServiceManagerProvider -output=automock -outpkg=automock -case=underscore
type ServiceManagerProvider interface {
	ForRequest(r *http.Request) (ServiceManager, error)
}

//go:generate mockery -name=ServiceManager -output=automock -outpkg=automock -case=underscore
type ServiceManager interface {
	Create(serviceDetails model.GraphQLServiceDetailsInput) error
	GetFromApplicationDetails(serviceID string) (model.GraphQLServiceDetails, error)
	ListFromApplicationDetails() ([]model.GraphQLServiceDetails, error)
	Update(serviceDetails model.GraphQLServiceDetailsInput) error
	Delete(serviceID string) error
}

//go:generate mockery -name=UIDService -output=automock -outpkg=automock -case=underscore
type UIDService interface {
	Generate() string
}

//go:generate mockery -name=Validator -output=automock -outpkg=automock -case=underscore
type Validator interface {
	Validate(details model.ServiceDetails) apperrors.AppError
}

const serviceIDVarKey = "serviceId"

type Handler struct {
	uidService     UIDService
	logger         *log.Logger
	validator      Validator
	converter      Converter
	serviceManager ServiceManagerProvider
}

func NewHandler(converter Converter, validator Validator, serviceManager ServiceManagerProvider, uidService UIDService, logger *log.Logger) *Handler {
	return &Handler{
		converter:      converter,
		validator:      validator,
		serviceManager: serviceManager,
		uidService:     uidService,
		logger:         logger,
	}
}

func (h *Handler) Create(writer http.ResponseWriter, request *http.Request) {
	defer h.closeBody(request)

	serviceDetails, err := h.decodeAndValidateInput(request)
	if err != nil {
		h.logger.Error(err)
		reqerror.WriteAppError(writer, err)
	}

	serviceID := h.uidService.Generate()
	converted, err := h.converter.DetailsToGraphQLInput(serviceID, serviceDetails)
	if err != nil {
		wrappedErr := errors.Wrap(err, "while converting service input")
		h.logger.Error(wrappedErr)
		reqerror.WriteError(writer, wrappedErr, apperrors.CodeInternal)
		return
	}

	serviceManager, err := h.loadServiceManager(request)
	if err != nil {
		h.writeErrorInternal(writer, err)
		return
	}

	err = serviceManager.Create(converted)
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

	serviceManager, err := h.serviceManager.ForRequest(request)
	if err != nil {
		wrappedErr := errors.Wrap(err, "while requesting Service Manager")
		reqerror.WriteError(writer, wrappedErr, apperrors.CodeInternal)
		return
	}

	output, err := serviceManager.GetFromApplicationDetails(serviceID)
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

	writer.WriteHeader(http.StatusNotImplemented)
}

func (h *Handler) List(writer http.ResponseWriter, request *http.Request) {
	h.logger.Println("List")
	//TODO: Implement it, currently this endpoint purpose is for manually testing appdetails middleware

	ctx := request.Context()
	app, err := appdetails.LoadFromContext(ctx)
	if err != nil {
		wrappedErr := errors.Wrap(err, "while getting service from context")
		h.logger.Error(wrappedErr)
		reqerror.WriteError(writer, wrappedErr, apperrors.CodeInternal)
		return
	}
	log.Infof("Application from ctx: %+v", app)

	writer.WriteHeader(http.StatusNotImplemented)
}

func (h *Handler) Update(writer http.ResponseWriter, request *http.Request) {
	defer h.closeBody(request)

	id := h.getServiceID(request)

	serviceDetails, err := h.decodeAndValidateInput(request)
	if err != nil {
		h.logger.Error(err)
		reqerror.WriteAppError(writer, err)
	}

	converted, err := h.converter.DetailsToGraphQLInput(id, serviceDetails)
	if err != nil {
		wrappedErr := errors.Wrap(err, "while converting service input")
		h.logger.Error(wrappedErr)
		reqerror.WriteError(writer, wrappedErr, apperrors.CodeInternal)
		return
	}

	serviceManager, err := h.loadServiceManager(request)
	if err != nil {
		h.writeErrorInternal(writer, err)
		return
	}

	err = serviceManager.Update(converted)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			h.writeErrorNotFound(writer, id)
			return
		}

		wrappedErr := errors.Wrap(err, "while updating Service")
		h.logger.WithField("ID", id).Error(wrappedErr)
		reqerror.WriteError(writer, wrappedErr, apperrors.CodeInternal)
		return
	}

	writer.WriteHeader(http.StatusOK)
}

func (h *Handler) Delete(writer http.ResponseWriter, request *http.Request) {
	defer h.closeBody(request)

	id := h.getServiceID(request)

	serviceManager, err := h.loadServiceManager(request)
	if err != nil {
		h.writeErrorInternal(writer, err)
		return
	}

	err = serviceManager.Delete(id)
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

func (h *Handler) loadServiceManager(request *http.Request) (ServiceManager, error) {
	serviceManager, err := h.serviceManager.ForRequest(request)
	if err != nil {
		wrappedErr := errors.Wrap(err, "while requesting Service Manager")
		h.logger.Error(wrappedErr)
		return nil, wrappedErr
	}

	return serviceManager, nil
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
