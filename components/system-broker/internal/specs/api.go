package specs

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/system-broker/internal/director"
	. "github.com/kyma-incubator/compass/components/system-broker/pkg/log"
	"github.com/pivotal-cf/brokerapi/v7/domain/apiresponses"
	"github.com/pkg/errors"
)

const (
	SpecsAPI              = "/specifications"
	AppIDParameter        = "app_id"
	PackageIDParameter    = "package_id"
	DefinitionIDParameter = "definition_id"
)

type SpecsFetcher interface {
	FindSpecification(ctx context.Context, in *director.PackageSpecificationInput) (*director.PackageSpecificationOutput, error)
}

func API(rootAPI string, specsFetcher SpecsFetcher) func(router *mux.Router) {
	handler := &SpecsHandler{
		specsFetcher: specsFetcher,
		rootAPI:      rootAPI,
	}
	return handler.Routes
}

type SpecsHandler struct {
	specsFetcher SpecsFetcher
	rootAPI      string
}

func (h *SpecsHandler) Routes(router *mux.Router) {
	router.HandleFunc(h.rootAPI+SpecsAPI, h.FetchSpec).Methods(http.MethodGet)
}

func (h *SpecsHandler) FetchSpec(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	appID := req.FormValue(AppIDParameter)
	packageID := req.FormValue(PackageIDParameter)
	definitionID := req.FormValue(DefinitionIDParameter)

	logger := C(ctx).WithFields(map[string]interface{}{
		"appID":        appID,
		"packageID":    packageID,
		"definitionID": definitionID,
	})

	logger.Info("Fetching package specifications")

	if appID == "" {
		err := errors.New("app id cannot be empty")
		logger.WithError(err).Error("input validation error")
		h.respond(ctx, w, http.StatusBadRequest, apiresponses.ErrorResponse{
			Description: err.Error(),
		})
		return
	}

	if packageID == "" {
		err := errors.New("package id cannot be empty")
		logger.WithError(err).Error("input validation error")
		h.respond(ctx, w, http.StatusBadRequest, apiresponses.ErrorResponse{
			Description: err.Error(),
		})
		return
	}

	if definitionID == "" {
		err := errors.New("definition id cannot be empty")
		logger.WithError(err).Error("input validation error")
		h.respond(ctx, w, http.StatusBadRequest, apiresponses.ErrorResponse{
			Description: err.Error(),
		})
		return
	}

	specification, err := h.specsFetcher.FindSpecification(ctx, &director.PackageSpecificationInput{
		ApplicationID: appID,
		PackageID:     packageID,
		DefinitionID:  definitionID,
	})
	if err != nil {
		logger.WithError(err).Error("failed to fetch specification from director")
		h.respond(ctx, w, http.StatusInternalServerError, apiresponses.ErrorResponse{
			Description: err.Error(),
		})
		return
	}

	logger.Info("Successfully fetched package specifications")
	content, err := SpecForamtToContentTypeHeader(specification.Format)
	if err != nil {
		logger.WithError(err).Error("failed to prepare content type header")
		h.respond(ctx, w, http.StatusInternalServerError, apiresponses.ErrorResponse{
			Description: err.Error(),
		})
		return
	}

	if specification.Data == nil {
		h.respondWithContent(ctx, w, http.StatusNotFound, content, []byte(`{}`))
		return
	}

	body := []byte(string(*specification.Data))

	h.respondWithContent(ctx, w, http.StatusOK, content, body)
}

func (h *SpecsHandler) respond(ctx context.Context, w http.ResponseWriter, status int, response interface{}) {
	data, err := json.Marshal(response)
	if err != nil {
		C(ctx).WithError(err).Error("failed to marshal response")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	h.respondWithContent(ctx, w, status, "application/json", data)

}

func (h *SpecsHandler) respondWithContent(ctx context.Context, w http.ResponseWriter, status int, content string, response []byte) {
	w.Header().Set("Content-Type", content)

	w.WriteHeader(status)
	if _, err := w.Write(response); err != nil {
		C(ctx).WithError(err).Error("failed to write response")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

}

func SpecForamtToContentTypeHeader(format graphql.SpecFormat) (string, error) {
	switch format {
	case graphql.SpecFormatJSON:
		return "application/json", nil
	case graphql.SpecFormatXML:
		return "application/xml", nil
	case graphql.SpecFormatYaml:
		return "text/yaml", nil
	}

	return "", errors.Errorf("unknown spec format %s", format)
}
