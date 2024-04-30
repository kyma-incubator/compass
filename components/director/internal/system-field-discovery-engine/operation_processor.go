package systemfielddiscoveryengine

import (
	"context"
	"encoding/json"

	"github.com/kyma-incubator/compass/components/director/internal/system-field-discovery-engine/data"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
)

// SystemFieldDiscoveryService responsible for the service-layer operations of system field discovery
//
//go:generate mockery --name=ORDService --output=automock --outpkg=automock --case=underscore --disable-version-string
type SystemFieldDiscoveryService interface {
	ProcessSaasRegistryApplication(ctx context.Context, appID, tenantID string) error
}

// OperationsProcessor defines Open Resource Discovery operation processor
type OperationsProcessor struct {
	SystemFieldDiscoverySvc SystemFieldDiscoveryService
}

// Process processes the given operation
func (p *OperationsProcessor) Process(ctx context.Context, operation *model.Operation) error {
	var opData data.SystemFieldDiscoveryOperationData
	if err := json.Unmarshal(operation.Data, &opData); err != nil {
		return errors.Wrapf(err, "while unmarshalling operation with id %q", operation.ID)
	}

	log.C(ctx).Infof("Starting processing of operation with id %q", operation.ID)
	if operation.OpType == model.OperationTypeSaasRegistryDiscovery {
		if opData.ApplicationID != "" && opData.TenantID != "" {
			if err := p.SystemFieldDiscoverySvc.ProcessSaasRegistryApplication(ctx, opData.ApplicationID, opData.TenantID); err != nil {
				return err
			}
		} else {
			log.C(ctx).Infof("Operation with ID %q does not have an application ID or tenant ID defined in operation data", operation.ID)
		}
	}

	log.C(ctx).Infof("Processing of operation with id %q finished successfully", operation.ID)
	return nil
}
