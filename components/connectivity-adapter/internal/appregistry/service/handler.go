package service

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/gqlcli"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/reqerror"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"net/http"
)

const serviceIDVarKey = "serviceId"

type Handler struct {
	directorURL string
}

func NewHandler(directorURL string) *Handler {
	return &Handler{directorURL: directorURL}
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
	gqlCli := h.gqlClient(request)

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

func (h *Handler) gqlClient(rq *http.Request) *gcli.Client {
	return gqlcli.NewAuthorizedGraphQLClient(h.directorURL, rq)
}
