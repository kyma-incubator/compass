package destination_test

import (
	"github.com/kyma-incubator/compass/components/director/internal/domain/destination"
	"github.com/kyma-incubator/compass/components/director/internal/domain/destination/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint/operators"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	destinationcreatorpkg "github.com/kyma-incubator/compass/components/director/pkg/destinationcreator"
)

const (
	// IDs constants
	destinationID                          = "126ac686-5773-4ad0-8eb1-2349e931f852"
	internalDestinationSubaccountID        = "553ac686-5773-4ad0-8eb1-2349e931f852"
	externalDestinationSubaccountID        = "452ac686-5773-4ad0-8eb1-2349e931f852"
	secondDestinationFormationAssignmentID = "098ac686-5773-4ad0-8eb1-2349e931f852"
	destinationFormationAssignmentID       = "654ac686-5773-4ad0-8eb1-2349e931f852"
	destinationBundleID                    = "765ac686-5773-4ad0-8eb1-2349e931f852"
	destinationInstanceID                  = "999ac686-5773-4ad0-8eb1-2349e931f852"

	// Destination constants
	destinationName                  = "test-destination-name"
	destinationType                  = destinationcreatorpkg.TypeHTTP
	destinationProxyType             = destinationcreatorpkg.ProxyTypeInternet
	destinationNoAuthn               = destinationcreatorpkg.AuthTypeNoAuth
	designTimeDestName               = "test-design-time-dest-name"
	basicDestName                    = "test-basic-dest-name"
	samlAssertionDestName            = "test-saml-assertion-dest-name"
	clientCertAuthDestName           = "test-client-cert-auth-dest-name"
	oauth2ClientCredsDestName        = "test-oauth2-client-creds-dest-name"
	destinationURL                   = "http://dest-test-url"
	destinationDescription           = "test-dest-description"
	oauth2ClientCredsTokenServiceURL = "http://oauth2-token-service-url"
	oauth2ClientCredsClientID        = "test-client-id"
	oauth2ClientCredsClientSecret    = "test-client-secret"

	// Destination Creds constants
	basicDestUser     = "basic-user"
	basicDestPassword = "basic-pwd"

	// Other
	destinationLatestRevision = "2"
)

var (
	faID              = destinationFormationAssignmentID
	instanceID        = destinationInstanceID
	destinationEntity = fixDestinationEntity(destinationName)
	destinationModel  = fixDestinationModel(destinationName)
	initialDepth      = uint8(0)
)

func fixDestinationModel(name string) *model.Destination {
	return &model.Destination{
		ID:                    destinationID,
		Name:                  name,
		Type:                  string(destinationType),
		URL:                   destinationURL,
		Authentication:        string(destinationNoAuthn),
		SubaccountID:          internalDestinationSubaccountID,
		InstanceID:            &instanceID,
		FormationAssignmentID: &faID,
	}
}

func fixDestinationModelWithAuthnAndFAID(name, authn, formationAssignmentID string) *model.Destination {
	return &model.Destination{
		ID:                    destinationID,
		Name:                  name,
		Type:                  string(destinationType),
		URL:                   destinationURL,
		Authentication:        authn,
		SubaccountID:          internalDestinationSubaccountID,
		InstanceID:            &instanceID,
		FormationAssignmentID: &formationAssignmentID,
	}
}

func fixDestinationEntity(name string) *destination.Entity {
	return &destination.Entity{
		ID:                    destinationID,
		Name:                  name,
		Type:                  string(destinationType),
		URL:                   destinationURL,
		Authentication:        string(destinationNoAuthn),
		TenantID:              internalDestinationSubaccountID,
		InstanceID:            repo.NewValidNullableString(instanceID),
		FormationAssignmentID: repo.NewValidNullableString(faID),
	}
}

func fixDestinationInput() model.DestinationInput {
	return model.DestinationInput{
		Name:           destinationName,
		Type:           string(destinationType),
		URL:            destinationURL,
		Authentication: string(destinationNoAuthn),
	}
}

func fixDestinationDetails(name, authentication, subaccountID string) operators.Destination {
	return operators.Destination{
		Name:           name,
		Type:           string(destinationType),
		Description:    destinationDescription,
		ProxyType:      string(destinationProxyType),
		Authentication: authentication,
		URL:            destinationURL,
		SubaccountID:   subaccountID,
		InstanceID:     destinationInstanceID,
	}
}

func fixDestinationInfo(authType destinationcreatorpkg.AuthType, destType destinationcreatorpkg.Type, url string) *destinationcreatorpkg.DestinationInfo {
	return &destinationcreatorpkg.DestinationInfo{
		AuthenticationType: authType,
		Type:               destType,
		URL:                url,
	}
}

func fixBasicAuthn() operators.BasicAuthentication {
	return operators.BasicAuthentication{
		URL:      destinationURL,
		UIURL:    destinationURL,
		Username: basicDestUser,
		Password: basicDestPassword,
	}
}

func fixBasicDestInfo() *destinationcreatorpkg.DestinationInfo {
	return fixDestinationInfo(destinationcreatorpkg.AuthTypeBasic, destinationType, destinationURL)
}

func fixSAMLDestInfo() *destinationcreatorpkg.DestinationInfo {
	return fixDestinationInfo(destinationcreatorpkg.AuthTypeSAMLAssertion, destinationType, destinationURL)
}

func fixClientCertDestInfo() *destinationcreatorpkg.DestinationInfo {
	return fixDestinationInfo(destinationcreatorpkg.AuthTypeClientCertificate, destinationType, destinationURL)
}

func fixOAuth2ClientCredsDestInfo() *destinationcreatorpkg.DestinationInfo {
	return fixDestinationInfo(destinationcreatorpkg.AuthTypeOAuth2ClientCredentials, destinationType, destinationURL)
}

func fixSAMLAssertionAuthentication() *operators.SAMLAssertionAuthentication {
	return &operators.SAMLAssertionAuthentication{URL: destinationURL}
}

func fixClientCertAuthTypeAuthentication() *operators.ClientCertAuthentication {
	return &operators.ClientCertAuthentication{URL: destinationURL}
}

func fixDesignTimeDestinationsDetails() []operators.Destination {
	return []operators.Destination{
		fixDestinationDetails(designTimeDestName, string(destinationcreatorpkg.AuthTypeNoAuth), externalDestinationSubaccountID),
	}
}

func fixBasicDestinationsDetails() []operators.Destination {
	return []operators.Destination{
		fixDestinationDetails(basicDestName, string(destinationcreatorpkg.AuthTypeBasic), externalDestinationSubaccountID),
	}
}

func fixSAMLAssertionDestinationsDetails() []operators.Destination {
	return []operators.Destination{
		fixDestinationDetails(samlAssertionDestName, string(destinationcreatorpkg.AuthTypeSAMLAssertion), externalDestinationSubaccountID),
	}
}

func fixClientCertAuthDestinationsDetails() []operators.Destination {
	return []operators.Destination{
		fixDestinationDetails(clientCertAuthDestName, string(destinationcreatorpkg.AuthTypeClientCertificate), externalDestinationSubaccountID),
	}
}

func fixOAuth2ClientCredsDestinationsDetails() []operators.Destination {
	return []operators.Destination{
		fixDestinationDetails(oauth2ClientCredsDestName, string(destinationcreatorpkg.AuthTypeOAuth2ClientCredentials), externalDestinationSubaccountID),
	}
}

func fixOAuth2ClientCredsAuthn() operators.OAuth2ClientCredentialsAuthentication {
	return operators.OAuth2ClientCredentialsAuthentication{
		URL:             destinationURL,
		TokenServiceURL: oauth2ClientCredsTokenServiceURL,
		ClientID:        oauth2ClientCredsClientID,
		ClientSecret:    oauth2ClientCredsClientSecret,
	}
}

func fixColumns() []string {
	return []string{"id", "name", "type", "url", "authentication", "tenant_id", "bundle_id", "revision", "instance_id", "formation_assignment_id"}
}

func fixUUID() string {
	return destinationID
}

func unusedDestinationCreatorService() *automock.DestinationCreatorService {
	return &automock.DestinationCreatorService{}
}

func unusedTenantRepository() *automock.TenantRepository {
	return &automock.TenantRepository{}
}

func unusedDestinationRepository() *automock.DestinationRepository {
	return &automock.DestinationRepository{}
}

func unusedUIDService() *automock.UIDService {
	return &automock.UIDService{}
}
