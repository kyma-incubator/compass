package destinationcreator_test

import (
	"encoding/json"
	"fmt"

	destinationcreatorpkg "github.com/kyma-incubator/compass/components/director/pkg/destinationcreator"

	esmdestinationcreator "github.com/kyma-incubator/compass/components/external-services-mock/internal/destinationcreator"
)

var (
	regionParamValue       = "region"
	subaccountIDParamValue = "subaccountId"
	correlationIDsKey      = "correlationIds"
	destNameParamKey       = "destinationName"
	certNameParamKey       = "certificateName"
	nameParamKey           = "name"

	testRegion                       = "testRegion"
	testSubaccountID                 = "testSubaccountID"
	testServiceInstanceID            = "testServiceInstanceID"
	testDestinationCertName          = "test-destination-cert-name"
	testCertChain                    = esmdestinationcreator.CertChain
	testDestinationCertWithExtension = testDestinationCertName + destinationcreatorpkg.JavaKeyStoreFileExtension

	url = "https://target-url.com"

	noAuthDestName        = "test-no-auth-dest"
	basicAuthDestName     = "test-basic-dest"
	samlAssertionDestName = "test-saml-assertion-dest"

	destinationCreatorNoAuthDestReqBody        = fmt.Sprintf(`{"name":"%s","url":"http://localhost","type":"HTTP","proxyType":"Internet","authenticationType":"NoAuthentication","additionalProperties":{"customKey":"customValue"}}`, noAuthDestName)
	destinationCreatorBasicAuthDestReqBody     = fmt.Sprintf(`{"name":"%s","url":"http://localhost","type":"HTTP","proxyType":"Internet","authenticationType":"BasicAuthentication","user":"my-first-user","password":"secretPassword","additionalProperties":{"%s":"value"}}`, basicAuthDestName, correlationIDsKey)
	destinationCreatorSAMLAssertionDestReqBody = fmt.Sprintf(`{"name":"%s","url":"https://localhost","type":"HTTP","proxyType":"Internet","authenticationType":"SAMLAssertion","audience":"https://localhost","keyStoreLocation":"test.jks","additionalProperties":{"%s":"value"}}`, samlAssertionDestName, correlationIDsKey)

	destinationServiceNoAuthDestReqBody    = fmt.Sprintf(`{"name":"%s","url":"http://localhost","type":"HTTP","proxyType":"Internet","authentication":"NoAuthentication"}`, noAuthDestName)
	destinationServiceBasicAuthReqBody     = fmt.Sprintf(`{"name":"%s","url":"http://localhost","type":"HTTP","proxyType":"Internet","authentication":"BasicAuthentication","user":"my-first-user","password":"secretPassword"}`, basicAuthDestName)
	destinationServiceSAMLAssertionReqBody = fmt.Sprintf(`{"name":"%s","url":"https://localhost","type":"HTTP","proxyType":"Internet","authentication":"SAMLAssertion","audience":"https://localhost","keyStoreLocation":"test.jks"}`, samlAssertionDestName)

	destinationCreatorReqBodyWithoutAuthType = fmt.Sprintf(`{"name":"%s","url":"http://localhost","type":"HTTP","proxyType":"Internet","additionalProperties":{"customKey":"customValue"}}`, noAuthDestName)

	destinationCreatorCertReqBody      = fmt.Sprintf(`{"name":"%s"}`, testDestinationCertName)
	destinationServiceCertResponseBody = fmt.Sprintf(`{"Name":"%s","Content":"%s"}`, testDestinationCertWithExtension, testCertChain)
)

func fixDestinationMappings(destName string, bodyBytes []byte) map[string]json.RawMessage {
	return map[string]json.RawMessage{
		destName: bodyBytes,
	}
}

func fixCertMappings(certName string, bodyBytes []byte) map[string]json.RawMessage {
	return map[string]json.RawMessage{
		certName: bodyBytes,
	}
}
