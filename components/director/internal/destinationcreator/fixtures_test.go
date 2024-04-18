package destinationcreator_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	destinationcreatorpkg "github.com/kyma-incubator/compass/components/director/pkg/destinationcreator"

	"github.com/kyma-incubator/compass/components/director/internal/destinationcreator"
	"github.com/kyma-incubator/compass/components/director/internal/destinationcreator/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint/operators"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/tenant"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
)

const (
	// Destination constants
	designTimeDestName               = "test-design-time-dest-name"
	basicDestName                    = "test-basic-dest-name"
	samlAssertionDestName            = "test-saml-assertion-dest-name"
	samlAssertionDestURL             = "test-saml-assertion-dest-url"
	clientCertAuthDestName           = "test-client-cert-auth-dest-name"
	oauth2ClientCredsDestName        = "test-oauth2-client-creds-dest-name"
	oauth2mTLSDestName               = "test-oauth2-mTLS-dest-name"
	destinationDescription           = "test-dest-description"
	destinationTypeHTTP              = string(destinationcreatorpkg.TypeHTTP)
	destinationProxyTypeInternet     = string(destinationcreatorpkg.ProxyTypeInternet)
	destinationURL                   = "https://dest-test-url"
	invalidDestAuthType              = "invalidDestAuthTypeValue"
	destinationInternalSubaccountID  = "destination-internal-subaccount-id"
	destinationExternalSubaccountID  = "destination-external-subaccount-id"
	destinationExternalSubaccountID2 = "destination-external-subaccount-id-2"
	destinationTenantName            = "testDestinationTenantName"
	destinationInstanceID            = "destination-instance"
	basicDestURL                     = "basic-url"
	basicDestUser                    = "basic-user"
	basicDestPassword                = "basic-pwd"
	oauth2ClientCredsURL             = "oauth2-url"
	oauth2ClientCredsTokenURL        = "oauth2-token-url"
	oauth2ClientCredsClientID        = "oauth2-client-id"
	oauth2ClientCredsClientSecret    = "oauth2-client-secret"
	oauth2mTLSURL                    = "oauth2-mTLS-url"
	oauth2mTLSTokenURL               = "oauth2-mTLS-token-url"
	oauth2mTLSClientID               = "oauth2-mTLS-client-id"

	// Destination Certificate constants
	certificateName            = "testCertificateName"
	certificateFileNameKey     = "testCertFileNameKey"
	certificateCommonNameKey   = "testCertCommonNameKey"
	certificateChainKey        = "testCertChainKey"
	certificateFileNameValue   = "testCertFileNameValue"
	certificateCommonNameValue = "testCertCommonNameValue"
	certificateChainValue      = "testCertChainValue"

	// Formation Assignment constants
	testAssignmentID  = "TestAssignmentID"
	testFormationID   = "TestFormationID"
	testTenantID      = "TestTenantID"
	testSourceID      = "TestSourceID"
	testTargetID      = "TestTargetID"
	invalidTargetType = "invalidTargetType"

	// Application constants
	appID   = "testAppID"
	appName = "testAppName"

	// Tenant constants
	internalAccountTenantID = "internalAccountTenantID"

	// Runtime + Runtime Context constants
	runtimeID    = "testRuntimeID"
	runtimeCtxID = "testRuntimeCtxID"
)

var (
	emptyCtx = context.Background()
	testErr  = errors.New("Test Error")

	appTemplateID   = "testAppTemplateID"
	appBaseURL      = "http://app-test-base-url"
	appEmptyBaseURL = ""
	testURLPath     = "/test/path"

	TestEmptyErrorValueRawJSON = json.RawMessage(`\"\"`)
	TestConfigValueRawJSON     = json.RawMessage(`{"configKey":"configValue"}`)

	emptyLblMap = map[string]*model.Label{}

	subaccountnLbl = map[string]*model.Label{
		destinationcreator.GlobalSubaccountLabelKey: {
			Value: destinationExternalSubaccountID,
		},
	}

	subaccountnLblWithInvalidIDValue = map[string]*model.Label{
		destinationcreator.GlobalSubaccountLabelKey: {
			Value: "invalidID",
		},
	}

	invalidLblValue               = 0
	subaccountnLblWithInvalidType = map[string]*model.Label{
		destinationcreator.GlobalSubaccountLabelKey: {
			Value: invalidLblValue,
		},
	}

	lblWithEmptyValue = &model.Label{Value: ""}
	regionLbl         = &model.Label{Value: "testRegionValue"}

	regionLblWithInvalidType = &model.Label{
		Value: invalidLblValue,
	}

	subaccTenant = &model.BusinessTenantMapping{
		ID:             destinationInternalSubaccountID,
		Name:           destinationTenantName,
		ExternalTenant: destinationExternalSubaccountID,
		Type:           tenant.Subaccount,
	}
	accountTenant = &model.BusinessTenantMapping{
		ID:             internalAccountTenantID,
		Name:           "externalAccountTenantName",
		ExternalTenant: "externalAccountTenantID",
		Type:           tenant.Account,
	}

	testApp = &model.Application{
		BaseEntity: &model.BaseEntity{
			ID: appID,
		},
		Name:                  appName,
		BaseURL:               &appBaseURL,
		ApplicationTemplateID: &appTemplateID,
	}
	testAppWithEmptyBaseURL = &model.Application{
		BaseEntity: &model.BaseEntity{
			ID: appID,
		},
		Name:                  appName,
		BaseURL:               &appEmptyBaseURL,
		ApplicationTemplateID: &appTemplateID,
	}
	testAppWithoutTmplID = &model.Application{
		BaseEntity: &model.BaseEntity{
			ID: appID,
		},
		Name: appName,
	}

	runtimeCtx = &model.RuntimeContext{
		ID:        runtimeCtxID,
		RuntimeID: runtimeID,
	}
)

func fixDestinationConfig() *destinationcreator.Config {
	return &destinationcreator.Config{
		CorrelationIDsKey: "testCorrelationIDsKey",
		DestinationAPIConfig: &destinationcreator.DestinationAPIConfig{
			BaseURL:              "testDestinationBaseURL/",
			SubaccountLevelPath:  "testDestinationPath",
			RegionParam:          "testDestinationRegionParam",
			SubaccountIDParam:    "testDestinationSubaccountIDParam",
			DestinationNameParam: "testDestinationNameParam",
		},
		CertificateAPIConfig: &destinationcreator.CertificateAPIConfig{
			BaseURL:              "testCertificateBaseURL/",
			SubaccountLevelPath:  "testCertificatePath",
			RegionParam:          "testCertificateRegionParam",
			SubaccountIDParam:    "testCertificateSubaccountIDParam",
			CertificateNameParam: "testCertificateNameParam",
			FileNameKey:          certificateFileNameKey,
			CommonNameKey:        certificateCommonNameKey,
			CertificateChainKey:  certificateChainKey,
		},
	}
}

func fixDestinationInfo(authType, destType, url string) *destinationcreatorpkg.DestinationInfo {
	return &destinationcreatorpkg.DestinationInfo{
		AuthenticationType: destinationcreatorpkg.AuthType(authType),
		Type:               destinationcreatorpkg.Type(destType),
		URL:                url,
	}
}

func fixDesignTimeDestinationDetails() operators.DestinationRaw {
	return operators.DestinationRaw{
		Destination: json.RawMessage(fmt.Sprintf(`{"url":"%s","name":"%s","type":"%s","proxyType":"%s","authenticationType":"%s","subaccountId":"%s","description":"%s"}`, destinationURL, designTimeDestName, destinationTypeHTTP, destinationProxyTypeInternet, string(destinationcreatorpkg.AuthTypeNoAuth), destinationExternalSubaccountID, destinationDescription))}
}

func fixBasicDestinationDetails() operators.Destination {
	return fixDestinationDetails(basicDestName, string(destinationcreatorpkg.AuthTypeBasic), destinationExternalSubaccountID)
}

func fixSAMLAssertionDestinationDetails() operators.Destination {
	return fixDestinationDetails(samlAssertionDestName, string(destinationcreatorpkg.AuthTypeSAMLAssertion), destinationExternalSubaccountID)
}

func fixSAMLAssertionDestinationsDetails() []operators.Destination {
	return []operators.Destination{
		fixSAMLAssertionDestinationDetails(),
	}
}

func fixClientCertAuthDestinationDetails() operators.Destination {
	return fixDestinationDetails(clientCertAuthDestName, string(destinationcreatorpkg.AuthTypeClientCertificate), destinationExternalSubaccountID)
}

func fixOAuth2ClientCredsDestinationDetails() operators.Destination {
	dest := fixDestinationDetails(oauth2ClientCredsDestName, string(destinationcreatorpkg.AuthTypeOAuth2ClientCredentials), destinationExternalSubaccountID)
	dest.TokenServiceURLType = string(destinationcreatorpkg.DedicatedTokenServiceURLType)
	return dest
}

func fixOAuth2mTLSDestinationDetails() operators.Destination {
	dest := fixDestinationDetails(oauth2mTLSDestName, string(destinationcreatorpkg.AuthTypeOAuth2mTLS), destinationExternalSubaccountID)
	dest.TokenServiceURLType = string(destinationcreatorpkg.DedicatedTokenServiceURLType)
	return dest
}

func fixDestinationsDetailsWithoutSubaccountID() []operators.Destination {
	return []operators.Destination{
		fixDestinationDetails(samlAssertionDestName, string(destinationcreatorpkg.AuthTypeSAMLAssertion), ""),
	}
}

func fixDestinationsDetailsWitDifferentSubaccountIDs() []operators.Destination {
	return []operators.Destination{
		fixDestinationDetails(samlAssertionDestName, string(destinationcreatorpkg.AuthTypeSAMLAssertion), destinationExternalSubaccountID),
		fixDestinationDetails(samlAssertionDestName, string(destinationcreatorpkg.AuthTypeSAMLAssertion), destinationExternalSubaccountID2),
	}
}

func fixDestinationDetails(name, authentication, subaccountID string) operators.Destination {
	return operators.Destination{
		Name:           name,
		Type:           destinationTypeHTTP,
		Description:    destinationDescription,
		ProxyType:      destinationProxyTypeInternet,
		Authentication: authentication,
		URL:            destinationURL,
		SubaccountID:   subaccountID,
	}
}

func fixBasicAuthCreds(url, username, password string) operators.BasicAuthentication {
	return operators.BasicAuthentication{
		URL:      url,
		Username: username,
		Password: password,
	}
}

func fixSAMLAssertionAuthCreds(url string) *operators.SAMLAssertionAuthentication {
	return &operators.SAMLAssertionAuthentication{
		URL: url,
	}
}

func fixOAuth2ClientCreds(url, tokenServiceURL, clientID, clientSecret string) *operators.OAuth2ClientCredentialsAuthentication {
	return &operators.OAuth2ClientCredentialsAuthentication{
		URL:             url,
		TokenServiceURL: tokenServiceURL,
		ClientID:        clientID,
		ClientSecret:    clientSecret,
	}
}

func fixOAuth2mTLSCreds(url, tokenServiceURL, clientID string) *operators.OAuth2mTLSAuthentication {
	return &operators.OAuth2mTLSAuthentication{
		URL:             url,
		TokenServiceURL: tokenServiceURL,
		ClientID:        clientID,
	}
}

func fixClientCertAuthTypeCreds() *operators.ClientCertAuthentication {
	return &operators.ClientCertAuthentication{URL: destinationURL}
}

func fixBasicRequestBody(url string) *destinationcreator.BasicAuthDestinationRequestBody {
	return &destinationcreator.BasicAuthDestinationRequestBody{
		BaseDestinationRequestBody: destinationcreator.BaseDestinationRequestBody{
			Name:                 basicDestName,
			URL:                  url,
			Type:                 destinationcreatorpkg.TypeHTTP,
			ProxyType:            destinationcreatorpkg.ProxyTypeInternet,
			AuthenticationType:   destinationcreatorpkg.AuthTypeBasic,
			AdditionalProperties: json.RawMessage("{\"testCorrelationIDsKey\":\"correlation-id-1,correlation-id-2\"}"),
		},
		User:     basicDestUser,
		Password: basicDestPassword,
	}
}

func fixFormationAssignmentModelWithParameters(id, formationID, tenantID, sourceID, targetID string, sourceType, targetType model.FormationAssignmentType, state string, configValue, errorValue json.RawMessage) *model.FormationAssignment {
	return &model.FormationAssignment{
		ID:          id,
		FormationID: formationID,
		TenantID:    tenantID,
		Source:      sourceID,
		SourceType:  sourceType,
		Target:      targetID,
		TargetType:  targetType,
		State:       state,
		Value:       configValue,
		Error:       errorValue,
	}
}

func fixCertificateResponse(fileName, commonName, certChain string) *destinationcreator.CertificateResponse {
	return &destinationcreator.CertificateResponse{
		FileName:         fileName,
		CommonName:       commonName,
		CertificateChain: certChain,
	}
}

func fixHTTPResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func requestThatHasMethod(expectedMethod string) interface{} {
	return mock.MatchedBy(func(actualReq *http.Request) bool {
		return actualReq.Method == expectedMethod
	})
}

func fixUnusedHTTPClient() *automock.HttpClient {
	return &automock.HttpClient{}
}

func fixUnusedAppRepo() *automock.ApplicationRepository {
	return &automock.ApplicationRepository{}
}

func fixUnusedRuntimeRepo() *automock.RuntimeRepository {
	return &automock.RuntimeRepository{}
}

func fixUnusedRuntimeCtxRepo() *automock.RuntimeCtxRepository {
	return &automock.RuntimeCtxRepository{}
}

func fixUnusedLabelRepo() *automock.LabelRepository {
	return &automock.LabelRepository{}
}

func fixUnusedTenantRepo() *automock.TenantRepository {
	return &automock.TenantRepository{}
}
