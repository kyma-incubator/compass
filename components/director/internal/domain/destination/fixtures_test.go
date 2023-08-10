package destination_test

import (
	"github.com/kyma-incubator/compass/components/director/internal/destinationcreator"
	"github.com/kyma-incubator/compass/components/director/internal/domain/destination"
	"github.com/kyma-incubator/compass/components/director/internal/domain/destination/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint/operators"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	destinationcreatorpkg "github.com/kyma-incubator/compass/components/director/pkg/destinationcreator"
)

const (
	destinationID                          = "126ac686-5773-4ad0-8eb1-2349e931f852"
	destinationName                        = "test-destination-name"
	destinationType                        = destinationcreatorpkg.TypeHTTP
	destinationProxyType                   = destinationcreatorpkg.ProxyTypeInternet
	destinationURL                         = "http://dest-test-url"
	destinationAuthn                       = destinationcreatorpkg.AuthTypeNoAuth
	destinationSubaccountID                = "553ac686-5773-4ad0-8eb1-2349e931f852"
	externalDestinationSubaccountID        = "452ac686-5773-4ad0-8eb1-2349e931f852"
	destinationFormationAssignmentID       = "654ac686-5773-4ad0-8eb1-2349e931f852"
	secondDestinationFormationAssignmentID = "098ac686-5773-4ad0-8eb1-2349e931f852"
	destinationBundleID                    = "765ac686-5773-4ad0-8eb1-2349e931f852"
	destinationLatestRevision              = "2"
	designTimeDestName                     = "test-design-time-dest-name"
	basicDestName                          = "test-basic-dest-name"
	samlAssertionDestName                  = "test-saml-assertion-dest-name"
	destinationDescription                 = "test-dest-description"
	basicDestUser                          = "basic-user"
	basicDestPassword                      = "basic-pwd"
)

var (
	faID              = destinationFormationAssignmentID
	destinationEntity = fixDestinationEntity(destinationName)
	destinationModel  = fixDestinationModel(destinationName)
)

func fixDestinationModel(name string) *model.Destination {
	return &model.Destination{
		ID:                    destinationID,
		Name:                  name,
		Type:                  string(destinationType),
		URL:                   destinationURL,
		Authentication:        string(destinationAuthn),
		SubaccountID:          destinationSubaccountID,
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
		SubaccountID:          destinationSubaccountID,
		FormationAssignmentID: &formationAssignmentID,
	}
}

func fixDestinationEntity(name string) *destination.Entity {
	return &destination.Entity{
		ID:                    destinationID,
		Name:                  name,
		Type:                  string(destinationType),
		URL:                   destinationURL,
		Authentication:        string(destinationAuthn),
		TenantID:              destinationSubaccountID,
		FormationAssignmentID: repo.NewValidNullableString(faID),
	}
}

func fixDestinationInput() model.DestinationInput {
	return model.DestinationInput{
		Name:           destinationName,
		Type:           string(destinationType),
		URL:            destinationURL,
		Authentication: string(destinationAuthn),
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

func fixBasicReqBody() *destinationcreator.BasicAuthDestinationRequestBody {
	return &destinationcreator.BasicAuthDestinationRequestBody{
		BaseDestinationRequestBody: destinationcreator.BaseDestinationRequestBody{
			Name:               basicDestName,
			URL:                destinationURL,
			Type:               destinationType,
			ProxyType:          destinationProxyType,
			AuthenticationType: destinationcreatorpkg.AuthTypeBasic,
		},
		User:     basicDestUser,
		Password: basicDestPassword,
	}
}

func fixSAMLAssertionAuthentication() *operators.SAMLAssertionAuthentication {
	return &operators.SAMLAssertionAuthentication{URL: destinationURL}
}

func fixDesignTimeDestinationDetails() operators.Destination {
	return fixDestinationDetails(designTimeDestName, string(destinationcreatorpkg.AuthTypeNoAuth), externalDestinationSubaccountID)
}

func fixBasicDestinationDetails() operators.Destination {
	return fixDestinationDetails(basicDestName, string(destinationcreatorpkg.AuthTypeBasic), externalDestinationSubaccountID)
}

func fixSAMLAssertionDestinationDetails() operators.Destination {
	return fixDestinationDetails(samlAssertionDestName, string(destinationcreatorpkg.AuthTypeSAMLAssertion), externalDestinationSubaccountID)
}

func fixColumns() []string {
	return []string{"id", "name", "type", "url", "authentication", "tenant_id", "bundle_id", "revision", "formation_assignment_id"}
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
