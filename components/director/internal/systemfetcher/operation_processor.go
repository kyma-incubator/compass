package systemfetcher

import (
	"context"
	"encoding/json"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
)

// SystemFetcherService responsible for the service-layer operations of system fetcher
//
//go:generate mockery --name=SystemFetcherService --output=automock --outpkg=automock --case=underscore --disable-version-string
type SystemFetcherService interface {
	ProcessTenant(ctx context.Context, tenantID string) error
	SetTemplateRenderer(templateRenderer TemplateRenderer)
}

// OperationsProcessor defines Open Resource Discovery operation processor
type OperationsProcessor struct {
	SystemFetcherSvc SystemFetcherService
}

// Process processes the given operation
func (p *OperationsProcessor) Process(ctx context.Context, operation *model.Operation) error {
	var opData SystemFetcherOperationData
	if err := json.Unmarshal(operation.Data, &opData); err != nil {
		return errors.Wrapf(err, "while unmarshalling operation with id %q", operation.ID)
	}

	log.C(ctx).Infof("Starting processing of operation with id %q", operation.ID)
	if opData.TenantID != "" {
		if err := p.SystemFetcherSvc.ProcessTenant(ctx, opData.TenantID); err != nil {
			return err
		}
	} else {
		log.C(ctx).Infof("Operation with ID %q does not have a tenant ID defined in operation data", operation.ID)
	}

	log.C(ctx).Infof("Processing of operation with id %q finished successfully", operation.ID)
	return nil
}
