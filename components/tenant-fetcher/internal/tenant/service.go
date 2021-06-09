package tenant

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

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
		logger.WithError(err).Errorf("while extracting request body: %v", err)
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	tenantId := gjson.GetBytes(body, s.config.TenantProviderTenantIdProperty)
	if !tenantId.Exists() || tenantId.Type != gjson.String || len(tenantId.String()) == 0 {
		logger.Errorf("Property %q not found in body or it is not of String type", s.config.TenantProviderTenantIdProperty)
		http.Error(writer, fmt.Sprintf("Property %q not found in body or it is not of String type", s.config.TenantProviderTenantIdProperty), http.StatusInternalServerError)
		return
	}
	customerId := gjson.GetBytes(body, s.config.TenantProviderCustomerIdProperty)
	subdomain := gjson.GetBytes(body, s.config.TenantProviderSubdomainProperty)

	tenant := model.TenantModel{
		ID:             s.uidService.Generate(),
		TenantId:       tenantId.String(),
		CustomerId:     customerId.String(),
		Subdomain:      subdomain.String(),
		TenantProvider: s.config.TenantProvider,
		Status:         tenantEntity.Active,
	}

	tx, err := s.transact.Begin()
	if err != nil {
		logger.WithError(err).Errorf("while beginning db transaction: %v", err)
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	defer s.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	if err := s.repository.Create(ctx, tenant); err != nil {
		if !apperrors.IsNotUniqueError(err) {
			logger.WithError(err).Errorf("while creating tenant: %v", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		if err := tx.Commit(); err != nil {
			logger.WithError(err).Errorf("while committing transaction : %v", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	writer.Header().Set("Content-Type", "text/plain")
	writer.WriteHeader(http.StatusOK)
	if _, err := writer.Write([]byte(compassURL)); err != nil {
		logger.WithError(err).Errorf("while writing response body: %v", err)
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *service) DeleteByExternalID(writer http.ResponseWriter, request *http.Request) {
	ctx := request.Context()
	logger := log.C(ctx)

	body, err := extractBody(request, writer)
	if err != nil {
		logger.WithError(err).Errorf("while extracting request body: %v", err)
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	tenantId := gjson.GetBytes(body, s.config.TenantProviderTenantIdProperty)
	if !tenantId.Exists() || tenantId.Type != gjson.String || len(tenantId.String()) == 0 {
		logger.Errorf("Property %q not found in body or it is not of String type", s.config.TenantProviderTenantIdProperty)
		http.Error(writer, fmt.Sprintf("Property %q not found in body or it is not of String type", s.config.TenantProviderTenantIdProperty), http.StatusInternalServerError)
		return
	}

	tx, err := s.transact.Begin()
	if err != nil {
		logger.WithError(err).Errorf("while beginning db transaction: %v", err)
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	defer s.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	if err := s.repository.DeleteByExternalID(ctx, tenantId.String()); err != nil {
		if !apperrors.IsNotFoundError(err) {
			logger.WithError(err).Errorf("while deleting tenant: %v", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		if err := tx.Commit(); err != nil {
			logger.WithError(err).Errorf("while committing transaction: %v", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	writer.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(writer).Encode(map[string]interface{}{})
	if err != nil {
		logger.WithError(err).Errorf("while writing to response body: %v", err)
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
}

func extractBody(r *http.Request, w http.ResponseWriter) ([]byte, error) {
	logger := log.C(r.Context())

	buf, bodyErr := ioutil.ReadAll(r.Body)
	if bodyErr != nil {
		logger.Error(errors.Wrap(bodyErr, "while reading request body"))
		http.Error(w, bodyErr.Error(), http.StatusInternalServerError)
		return nil, bodyErr
	}

	defer func() {
		err := r.Body.Close()
		if err != nil {
			logger.Warnf("Unable to close request body. Cause: %v", err)
		}
	}()

	return buf, nil
}
