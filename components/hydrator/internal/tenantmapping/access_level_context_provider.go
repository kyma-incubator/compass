package tenantmapping

import (
	"context"
	"fmt"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/model"

	directorErrors "github.com/kyma-incubator/compass/components/hydrator/internal/director"
	"github.com/kyma-incubator/compass/components/hydrator/pkg/tenantmapping"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/consumer"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/hydrator/pkg/oathkeeper"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const GlobalAccessLevel = "global"

type accessLevelContextProvider struct {
	directorClient DirectorClient
	tenantKeys     KeysExtra
	scopesGetter   ScopesGetter
}

// NewAccessLevelContextProvider implements the ObjectContextProvider interface by looking tenant header and access levels defined in the auth session extra.
func NewAccessLevelContextProvider(clientProvider DirectorClient, scopesGetter ScopesGetter) *accessLevelContextProvider {
	return &accessLevelContextProvider{
		directorClient: clientProvider,
		tenantKeys: KeysExtra{
			TenantKey:         tenantmapping.ConsumerTenantKey,
			ExternalTenantKey: tenantmapping.ExternalTenantKey,
		},
		scopesGetter: scopesGetter,
	}
}

// GetObjectContext is the accessLevelContextProvider implementation of the ObjectContextProvider interface.
// It receives the tenant information extracted from the tenant header in this case, and it verifies that the caller has access to this tenant.
func (p *accessLevelContextProvider) GetObjectContext(ctx context.Context, reqData oathkeeper.ReqData, authDetails oathkeeper.AuthDetails) (ObjectContext, error) {
	consumerType := reqData.ConsumerType()
	logger := log.C(ctx).WithFields(logrus.Fields{
		"consumer_type": consumerType,
	})
	ctx = log.ContextWithLogger(ctx, logger)

	scopes, err := p.directorScopes(consumerType)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to get scopes for consumer type %s: %v", consumerType, err)
		return ObjectContext{}, apperrors.NewInternalError(fmt.Sprintf("Failed to extract scopes for consumer with type %s", consumerType))
	}

	fmt.Println("ALEX NewAccessLevelContextProvider consumer", consumerType)

	externalTenantID, err := reqData.GetExternalTenantID()
	if err != nil {
		if apperrors.IsKeyDoesNotExist(err) {
			log.C(ctx).Infof("No tenant provided, will proceed with empty tenant context...")
			if err := p.verifyTenantAccessLevels(GlobalAccessLevel, authDetails, reqData); err != nil {
				log.C(ctx).WithError(err).Errorf("Failed to verify tenant access level: %v", err)
				return ObjectContext{}, err
			}

			return NewObjectContext(&graphql.Tenant{}, p.tenantKeys, scopes, mergeWithOtherScopes,
				"", "", authDetails.AuthID, authDetails.AuthFlow, consumer.Type(consumerType), tenantmapping.CertServiceObjectContextProvider, authDetails.Subject), nil
		}
		return ObjectContext{}, err
	}

	log.C(ctx).Infof("Getting the tenant with external ID: %s", externalTenantID)
	tenantMapping, region, err := getTenantWithRegion(ctx, p.directorClient, externalTenantID)
	if err != nil {
		if directorErrors.IsGQLNotFoundError(err) {
			// tenant not in DB yet, might be because we have not imported all subaccounts yet
			log.C(ctx).Warningf("Could not find tenant with external ID: %s, error: %s", externalTenantID, err.Error())
			log.C(ctx).Infof("Returning tenant context with empty internal tenant ID and external ID %s", externalTenantID)
			return NewObjectContext(&graphql.Tenant{ID: externalTenantID}, p.tenantKeys, scopes, mergeWithOtherScopes, "",
				"", authDetails.AuthID, authDetails.AuthFlow, consumer.Type(consumerType), tenantmapping.CertServiceObjectContextProvider, authDetails.Subject), nil
		}
		return ObjectContext{}, errors.Wrapf(err, "while getting external tenant mapping [ExternalTenantID=%s]", externalTenantID)
	}

	authDetails.Region = region

	if err := p.verifyTenantAccessLevels(tenantMapping.Type, authDetails, reqData); err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to verify tenant access level: %v", err)
		return ObjectContext{}, err
	}

	objCtx := NewObjectContext(tenantMapping, p.tenantKeys, scopes, mergeWithOtherScopes,
		authDetails.Region, "", authDetails.AuthID, authDetails.AuthFlow, consumer.Type(consumerType), tenantmapping.CertServiceObjectContextProvider, authDetails.Subject)
	log.C(ctx).Infof("Successfully got object context: %+v", RedactConsumerIDForLogging(objCtx))
	return objCtx, nil
}

// Match will only match requests coming from certificate consumers that are able to access only a subset of tenants.
func (p *accessLevelContextProvider) Match(_ context.Context, data oathkeeper.ReqData) (bool, *oathkeeper.AuthDetails, error) {
	if len(data.TenantAccessLevels()) == 0 {
		return false, nil, nil
	}

	idVal := data.Body.Header.Get(oathkeeper.ClientIDCertKey)
	certIssuer := data.Body.Header.Get(oathkeeper.ClientIDCertIssuer)
	subject1 := data.Body.Header.Get(oathkeeper.SubjectKey)
	subject2 := data.Body.Extra[oathkeeper.SubjectKey]
	fmt.Println("ALEX NewAccessLevelContextProvider", subject1, subject2)

	if idVal == "" || certIssuer != oathkeeper.ExternalIssuer {
		return false, nil, nil
	}

	if _, err := data.GetExternalTenantID(); err != nil {
		if !apperrors.IsKeyDoesNotExist(err) {
			return false, nil, err
		}
	}

	return true, &oathkeeper.AuthDetails{AuthID: idVal, AuthFlow: oathkeeper.CertificateFlow, CertIssuer: certIssuer, Subject: subject1}, nil
}

func (p *accessLevelContextProvider) verifyTenantAccessLevels(accessLevel string, authDetails oathkeeper.AuthDetails, reqData oathkeeper.ReqData) error {
	grantedAccessLevels := reqData.TenantAccessLevels()
	for _, al := range grantedAccessLevels {
		if accessLevel == al {
			return nil
		}
	}

	if accessLevel == GlobalAccessLevel {
		return apperrors.NewUnauthorizedError(fmt.Sprintf("Certificate with auth ID %s does not have global access", authDetails.AuthID))
	}

	externalTenantID, err := reqData.GetExternalTenantID()
	if err != nil {
		return err
	}

	return apperrors.NewUnauthorizedError(fmt.Sprintf("Certificate with auth ID %s has no access to %s tenant with ID %s", authDetails.AuthID, accessLevel, externalTenantID))
}

func (p *accessLevelContextProvider) directorScopes(consumerType model.SystemAuthReferenceObjectType) (string, error) {
	declaredScopes, err := p.scopesGetter.GetRequiredScopes(buildPath(consumerType))
	if err != nil {
		return "", errors.Wrap(err, "while fetching scopes")
	}
	return strings.Join(declaredScopes, " "), nil
}
