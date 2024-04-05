package processor

import (
	"context"

	"github.com/kyma-incubator/compass/components/ias-adapter/internal/errors"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/logger"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/service/ucl"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/types"
	"github.com/rs/zerolog/log"
)

type AsyncProcessor struct {
	TenantMappingsService TenantMappingsService
	UCLService            UCLService
}

const (
	locationHeader = "Location"
)

//go:generate mockery --name=TenantMappingsService --output=automock --outpkg=automock --case=underscore --disable-version-string
type TenantMappingsService interface {
	CanSafelyRemoveTenantMapping(ctx context.Context, formationID string) (bool, error)
	ProcessTenantMapping(ctx context.Context, tenantMapping types.TenantMapping) error
	RemoveTenantMapping(ctx context.Context, tenantMapping types.TenantMapping) error
}

//go:generate mockery --name=UCLService --output=automock --outpkg=automock --case=underscore --disable-version-string
type UCLService interface {
	ReportStatus(ctx context.Context, url string, statusReport ucl.StatusReport) error
}

func (p AsyncProcessor) ProcessTMRequest(ctx context.Context, tenantMapping types.TenantMapping) {
	reverseAssignmentState := tenantMapping.AssignedTenant.ReverseAssignmentState
	if tenantMapping.Operation == types.OperationAssign {
		if reverseAssignmentState != types.StateInitial && reverseAssignmentState != types.StateReady {
			log.Warn().Msgf("skipping processing tenant mapping notification with $.assignedTenants[0].reverseAssignmentState '%s'",
				reverseAssignmentState)
			p.reportStatus(ctx, ucl.StatusReport{State: types.StateConfigPending})
			return
		}
	}

	operation := tenantMapping.Operation

	if err := p.TenantMappingsService.ProcessTenantMapping(ctx, tenantMapping); err != nil {
		err = errors.Newf("failed to process tenant mapping notification: %w", err)

		if operation == types.OperationAssign {
			if errors.Is(err, errors.IASApplicationNotFound) {
				p.reportStatus(ctx, ucl.StatusReport{State: errorState(operation), Error: err.Error()})
				return
			}

			if errors.Is(err, errors.S4CertificateNotFound) {
				logger.FromContext(ctx).Info().Msgf("S/4 certificate not provided. Responding with CONFIG_PENDING.")
				s4Config := &types.TenantMappingConfiguration{
					Credentials: types.Credentials{
						OutboundCommunicationCredentials: types.CommunicationCredentials{
							OAuth2mTLSAuthentication: types.OAuth2mTLSAuthentication{
								CorrelationIds: []string{types.S4SAPManagedCommunicationScenario},
							},
						},
					},
				}
				p.reportStatus(ctx, ucl.StatusReport{State: types.StateConfigPending, Configuration: s4Config})
				return
			}
		}

		p.reportStatus(ctx, ucl.StatusReport{State: errorState(operation), Error: err.Error()})
		return
	}

	p.reportStatus(ctx, ucl.StatusReport{State: readyState(operation)})
}

func (p AsyncProcessor) reportStatus(ctx context.Context, statusReport ucl.StatusReport) {
	log := logger.FromContext(ctx)

	statusReportURL := ctx.Value(locationHeader).(string)
	if err := p.UCLService.ReportStatus(ctx, statusReportURL, statusReport); err != nil {
		log.Error().Msgf("failed to report status to '%s': %s", statusReportURL, err)
	}
}

func readyState(operation types.Operation) types.State {
	if operation == types.OperationAssign {
		return types.StateCreateReady
	}
	return types.StateDeleteReady
}

func errorState(operation types.Operation) types.State {
	if operation == types.OperationAssign {
		return types.StateCreateError
	}
	return types.StateDeleteError
}
