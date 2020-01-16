package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/gqlcli"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/reqerror"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const serviceIDVarKey = "serviceId"

type Handler struct {
	cliProvider gqlcli.Provider
	logger      *log.Logger
}

func NewHandler(cliProvider gqlcli.Provider, logger *log.Logger) *Handler {
	return &Handler{
		cliProvider: cliProvider,
		logger:      logger,
	}
}

func (h *Handler) Create(rw http.ResponseWriter, rq *http.Request) {
	h.logger.Println("Create")
	rw.WriteHeader(http.StatusOK)
}

func (h *Handler) Get(rw http.ResponseWriter, rq *http.Request) {
	h.logger.Println("Get")
	rw.WriteHeader(http.StatusOK)
}

func (h *Handler) List(rw http.ResponseWriter, rq *http.Request) {
	h.logger.Println("List")
	rw.WriteHeader(http.StatusOK)
}

func (h *Handler) Update(rw http.ResponseWriter, rq *http.Request) {
	h.logger.Println("Update")
	rw.WriteHeader(http.StatusOK)
}

func (h *Handler) Delete(writer http.ResponseWriter, request *http.Request) {
	gqlCli := h.cliProvider.GQLClient(request)

	vars := mux.Vars(request)

	id := vars[serviceIDVarKey]
	gqlRequest := prepareUnregisterApplicationRequest(id)

	err := gqlCli.Run(context.Background(), gqlRequest, nil)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			message := fmt.Sprintf("entity with ID %s not found", id)
			reqerror.WriteErrorMessage(writer, message, apperrors.CodeNotFound)
			return
		}

		h.logger.Error(errors.Wrapf(err, "while deleting service with ID %s", id))
		reqerror.WriteError(writer, err, apperrors.CodeInternal)
		return
	}

	writer.WriteHeader(http.StatusNoContent)
}
