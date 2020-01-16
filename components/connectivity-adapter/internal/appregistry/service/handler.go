package service

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/gqlcli"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/reqerror"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"net/http"
)

const serviceIDVarKey = "serviceId"

type Handler struct {
	cliProvider gqlcli.Provider
}

func NewHandler(cliProvider gqlcli.Provider) *Handler {
	return &Handler{
		cliProvider: cliProvider,
	}
}

func (h *Handler) Create(rw http.ResponseWriter, rq *http.Request) {
	log.Println("Create")
	rw.WriteHeader(http.StatusOK)
}

func (h *Handler) Get(rw http.ResponseWriter, rq *http.Request) {
	log.Println("Get")
	rw.WriteHeader(http.StatusOK)
}

func (h *Handler) List(rw http.ResponseWriter, rq *http.Request) {
	log.Println("List")
	rw.WriteHeader(http.StatusOK)
}

func (h *Handler) Update(rw http.ResponseWriter, rq *http.Request) {
	log.Println("Update")
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

		log.Error(errors.Wrapf(err, "while deleting service with ID %s", id))
		reqerror.WriteError(writer, err, apperrors.CodeInternal)
		return
	}

	writer.WriteHeader(http.StatusNoContent)
}
