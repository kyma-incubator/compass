package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/appregistry/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	gcli "github.com/machinebox/graphql"

	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/gqlcli"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/reqerror"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

//go:generate mockery -name=Converter -output=automock -outpkg=automock -case=underscore
type Converter interface {
	DetailsToGraphQLInput(in model.ServiceDetails) (graphql.ApplicationRegisterInput, error)
	GraphQLToDetailsModel(in graphql.ApplicationExt) (model.ServiceDetails, error)
	GraphQLToModel(in graphql.ApplicationExt) (model.Service, error)
}

//go:generate mockery -name=Validator -output=automock -outpkg=automock -case=underscore
type Validator interface {
	Validate(details model.ServiceDetails) apperrors.AppError
}

//go:generate mockery -name=GraphQLRequestBuilder -output=automock -outpkg=automock -case=underscore
type GraphQLRequestBuilder interface {
	RegisterApplicationRequest(input graphql.ApplicationRegisterInput) (*gcli.Request, error)
	UnregisterApplicationRequest(id string) *gcli.Request
	GetApplicationRequest(id string) *gcli.Request
}

const serviceIDVarKey = "serviceId"

type Handler struct {
	cliProvider       gqlcli.Provider
	logger            *log.Logger
	validator         Validator
	converter         Converter
	gqlRequestBuilder GraphQLRequestBuilder
}

func NewHandler(cliProvider gqlcli.Provider, converter Converter, validator Validator, gqlRequestBuilder GraphQLRequestBuilder, logger *log.Logger) *Handler {
	return &Handler{
		cliProvider:       cliProvider,
		converter:         converter,
		validator:         validator,
		gqlRequestBuilder: gqlRequestBuilder,
		logger:            logger,
	}
}

func (h *Handler) Create(writer http.ResponseWriter, request *http.Request) {
	defer h.closeBody(request)

	var details model.ServiceDetails
	err := json.NewDecoder(request.Body).Decode(&details)
	if err != nil {
		wrappedErr := errors.Wrap(err, "while unmarshalling service")
		h.logger.Error(wrappedErr)
		reqerror.WriteError(writer, wrappedErr, apperrors.CodeWrongInput)
		return
	}

	appErr := h.validator.Validate(details)
	if appErr != nil {
		wrappedErr := errors.Wrap(appErr, "while validating input")
		reqerror.WriteError(writer, wrappedErr, appErr.Code())
		return
	}

	input, err := h.converter.DetailsToGraphQLInput(details)
	if err != nil {
		wrappedErr := errors.Wrap(err, "while converting service input")
		h.logger.Error(wrappedErr)
		reqerror.WriteError(writer, wrappedErr, apperrors.CodeInternal)
		return
	}

	gqlRequest, err := h.gqlRequestBuilder.RegisterApplicationRequest(input)
	if err != nil {
		wrappedErr := errors.Wrap(err, "while building Application Register input")
		h.logger.Error(wrappedErr)
		reqerror.WriteError(writer, wrappedErr, apperrors.CodeInternal)
		return
	}

	var resp gqlCreateApplicationResponse
	gqlCli := h.cliProvider.GQLClient(request)
	err = gqlCli.Run(context.Background(), gqlRequest, &resp)
	if err != nil {
		wrappedErr := errors.Wrap(err, "while creating service")
		h.logger.Error(wrappedErr)
		reqerror.WriteError(writer, wrappedErr, apperrors.CodeInternal)
		return
	}

	successResponse := SuccessfulCreateResponse{
		ID: resp.Result.ID,
	}
	err = json.NewEncoder(writer).Encode(&successResponse)
	if err != nil {
		wrappedErr := errors.Wrap(err, "while encoding response")
		h.logger.Error(wrappedErr)
		reqerror.WriteError(writer, wrappedErr, apperrors.CodeInternal)
		return
	}

	writer.WriteHeader(http.StatusOK)
}

type SuccessfulCreateResponse struct {
	ID string `json:"id"`
}

func (h *Handler) Get(writer http.ResponseWriter, request *http.Request) {
	defer h.closeBody(request)
	gqlCli := h.cliProvider.GQLClient(request)

	id := h.getServiceID(request)
	gqlRequest := h.gqlRequestBuilder.GetApplicationRequest(id)

	var resp gqlGetApplicationResponse
	err := gqlCli.Run(context.Background(), gqlRequest, &resp)
	if err != nil {
		wrappedErr := errors.Wrap(err, "while creating service")
		h.logger.Error(wrappedErr)
		reqerror.WriteError(writer, wrappedErr, apperrors.CodeInternal)
		return
	}

	if resp.Result == nil {
		h.writeErrorNotFound(writer, id)
		return
	}

	serviceModel, err := h.converter.GraphQLToDetailsModel(*resp.Result)
	if err != nil {
		wrappedErr := errors.Wrap(err, "while converting model")
		h.logger.Error(wrappedErr)
		reqerror.WriteError(writer, wrappedErr, apperrors.CodeInternal)
		return
	}

	err = json.NewEncoder(writer).Encode(&serviceModel)
	if err != nil {
		wrappedErr := errors.Wrap(err, "while encoding response")
		h.logger.Error(wrappedErr)
		reqerror.WriteError(writer, wrappedErr, apperrors.CodeInternal)
		return
	}

	writer.WriteHeader(http.StatusOK)
}

func (h *Handler) List(writer http.ResponseWriter, request *http.Request) {
	h.logger.Println("List")
	// TODO: Implement it
	writer.WriteHeader(http.StatusNotImplemented)
}

func (h *Handler) Update(writer http.ResponseWriter, request *http.Request) {
	defer h.closeBody(request)

	// TODO: Implement it
	writer.WriteHeader(http.StatusNotImplemented)
}

func (h *Handler) Delete(writer http.ResponseWriter, request *http.Request) {
	defer h.closeBody(request)
	gqlCli := h.cliProvider.GQLClient(request)

	id := h.getServiceID(request)
	gqlRequest := h.gqlRequestBuilder.UnregisterApplicationRequest(id)

	err := gqlCli.Run(context.Background(), gqlRequest, nil)
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

func (h *Handler) writeErrorNotFound(writer http.ResponseWriter, id string) {
	message := fmt.Sprintf("entity with ID %s not found", id)
	reqerror.WriteErrorMessage(writer, message, apperrors.CodeNotFound)
}

func (h *Handler) getServiceID(request *http.Request) string {
	vars := mux.Vars(request)
	id := vars[serviceIDVarKey]
	return id
}
