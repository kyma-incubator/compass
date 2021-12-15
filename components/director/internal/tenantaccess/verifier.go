package tenantaccess

import (
	"context"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/oathkeeper"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/pkg/errors"
)

func NewVerifier(consumerExistsFuncs map[model.SystemAuthReferenceObjectType]func(context.Context, string) (bool, error)) *verifier {
	return &verifier{
		consumerExistsFuncs: consumerExistsFuncs,
	}
}

type verifier struct {
	consumerExistsFuncs map[model.SystemAuthReferenceObjectType]func(context.Context, string) (bool, error)
}

func (v *verifier) VerifyTenantAccess(ctx context.Context, tenant *model.BusinessTenantMapping, authDetails oathkeeper.AuthDetails, reqData oathkeeper.ReqData) error {
	if !reqData.IsIntegrationSystemFlow() {
		return nil
	}

	data := reqData.GetExtraDataWithDefaults()
	var accessLevelExists bool
	for _, al := range data.AccessLevels {
		if tenant.Type == al {
			accessLevelExists = true
			break
		}
	}

	if !accessLevelExists {
		return apperrors.NewUnauthorizedError(fmt.Sprintf("Certificate with auth ID %s has no access to %s tenant with ID %s", authDetails.AuthID, tenant.Type, tenant.ExternalTenant))
	}

	return v.verifyConsumerExists(ctx, data)
}

func (v *verifier) verifyConsumerExists(ctx context.Context, data oathkeeper.ExtraData) error {
	if data.InternalConsumerID == "" {
		return nil
	}

	found, err := v.consumerExistsFuncs[data.ConsumerType](ctx, data.InternalConsumerID)
	if err != nil {
		return errors.Wrapf(err, "while getting %s with ID %s", data.ConsumerType, data.InternalConsumerID)
	}
	if !found {
		return apperrors.NewUnauthorizedError(fmt.Sprintf("%s with ID %s does not exist", data.ConsumerType, data.InternalConsumerID))
	}
	return nil
}
