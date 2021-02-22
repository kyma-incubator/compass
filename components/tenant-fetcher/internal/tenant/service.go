package tenant

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	tenantEntity "github.com/kyma-incubator/compass/components/director/pkg/tenant"
	"github.com/tidwall/gjson"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/tenant-fetcher/internal/model"
	"github.com/pkg/errors"
)

//go:generate mockery --name=TenantService --output=automock --outpkg=automock --case=underscore
type TenantService interface {
	Create(ctx context.Context, item model.TenantModel) error
	DeleteByExternalID(ctx context.Context, tenantId string) error
}

//go:generate mockery --name=UIDService --output=automock --outpkg=automock --case=underscore
type UIDService interface {
	Generate() string
}

type service struct {
	repository TenantRepository
	transact   persistence.Transactioner
	uidService UIDService
	config     Config
}

func NewService(tenant TenantRepository, transact persistence.Transactioner, uidService UIDService, config Config) *service {
	return &service{
		repository: tenant,
		transact:   transact,
		uidService: uidService,
		config:     config,
	}
}

func (s *service) Create(writer http.ResponseWriter, request *http.Request) {
	ctx := request.Context()
	logger := log.C(ctx)

	body, err := extractBody(request, writer)
	if err != nil {
		logger.Error(errors.Wrapf(err, "while extracting request body"))
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	tenantId := gjson.GetBytes(body, s.config.TenantProviderTenantIdProperty).String()
	tenant := model.TenantModel{
		ID:             s.uidService.Generate(),
		TenantId:       tenantId,
		TenantProvider: s.config.TenantProvider,
		Status:         tenantEntity.Active,
	}

	tx, err := s.transact.Begin()
	if err != nil {
		logger.Error(errors.Wrapf(err, "while beginning db transaction"))
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	defer s.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	if err := s.repository.Create(ctx, tenant); err != nil {
		logger.Error(errors.Wrapf(err, "while creating tenant"))
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		logger.Error(errors.Wrapf(err, "while committing transaction"))
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "text/plain")
	writer.WriteHeader(http.StatusOK)
	if _, err := writer.Write([]byte(compassURL)); err != nil {
		logger.Error(errors.Wrapf(err, "while writing response body"))
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *service) DeleteByExternalID(writer http.ResponseWriter, request *http.Request) {
	ctx := request.Context()
	logger := log.C(ctx)

	body, err := extractBody(request, writer)
	if err != nil {
		logger.Error(errors.Wrapf(err, "while extracting request body"))
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	tenantId := gjson.GetBytes(body, s.config.TenantProviderTenantIdProperty).String()

	tx, err := s.transact.Begin()
	if err != nil {
		logger.Error(errors.Wrapf(err, "while beginning db transaction"))
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	defer s.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	if err := s.repository.DeleteByExternalID(ctx, tenantId); err != nil {
		logger.Error(errors.Wrapf(err, "while deleting tenant"))
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		logger.Error(errors.Wrapf(err, "while committing transaction"))
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(writer).Encode(map[string]interface{}{})
	if err != nil {
		logger.Error(errors.Wrapf(err, "while writing to response body"))
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
}
