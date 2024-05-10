package systemfielddiscoveryengine

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/internal/system-field-discovery-engine/data"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
)

// SystemFieldDiscoveryService responsible for the service-layer operations of system field discovery
//
//go:generate mockery --name=SystemFieldDiscoveryService --output=automock --outpkg=automock --case=underscore --disable-version-string
type SystemFieldDiscoveryService interface {
	ProcessSaasRegistryApplication(ctx context.Context, appID, tenantID string) error
}

type ProcessingError struct {
	Message string `json:"message"`
}

func (p *ProcessingError) Error() string {
	return p.toJSON()
}

func (p *ProcessingError) toJSON() string {
	bytes, err := json.Marshal(p)
	if err != nil {
		return fmt.Sprintf(`{"error": "failed to marshal error: %s"}`, err)
	}
	return string(bytes)
}

// OperationsProcessor defines Open Resource Discovery operation processor
type OperationsProcessor struct {
	systemFieldDiscoverySvc SystemFieldDiscoveryService
}

func NewOperationProcessor(systemFieldDiscoverySvc SystemFieldDiscoveryService) *OperationsProcessor {
	return &OperationsProcessor{
		systemFieldDiscoverySvc: systemFieldDiscoverySvc,
	}
}

// Process processes the given operation
func (p *OperationsProcessor) Process(ctx context.Context, operation *model.Operation) error {
	log.C(ctx).Infof("Starting processing of operation with id %q", operation.ID)
	if operation.OpType != model.OperationTypeSaasRegistryDiscovery {
		log.C(ctx).Infof("Unsupported operation type %v. Skipping processing.", operation.OpType)
		return errors.Errorf("unsupported operation type %v", operation.OpType)
	}

	var opData data.SystemFieldDiscoveryOperationData
	if err := json.Unmarshal(operation.Data, &opData); err != nil {
		return errors.Wrapf(err, "while unmarshalling operation with id %q", operation.ID)
	}

	if opData.ApplicationID != "" && opData.TenantID != "" {
		if err := p.systemFieldDiscoverySvc.ProcessSaasRegistryApplication(ctx, opData.ApplicationID, opData.TenantID); err != nil {
			return errors.Wrap(err, "while processing saas registry application")
		}
	} else {
		log.C(ctx).Infof("Operation with ID %q does not have an application ID or tenant ID defined in operation data", operation.ID)
	}

	log.C(ctx).Infof("Processing of operation with id %q finished successfully", operation.ID)
	return nil
}
