package destinationcreator_test

import (
	"encoding/json"
	"fmt"

	destinationcreatorpkg "github.com/kyma-incubator/compass/components/director/pkg/destinationcreator"
	esmdestinationcreator "github.com/kyma-incubator/compass/components/external-services-mock/internal/destinationcreator"
	esmdestcreatorpkg "github.com/kyma-incubator/compass/components/external-services-mock/pkg/destinationcreator"
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

	testDestURL              = "http://localhost"
	testSecureDestURL        = "https://localhost"
	testDestType             = "HTTP"
	testDestProxyType        = "Internet"
	testDestKeyStoreLocation = "test.jks"

	destinationCreatorNoAuthDestReqBody        = fmt.Sprintf(`{"name":"%s","url":"http://localhost","type":"HTTP","proxyType":"Internet","authenticationType":"NoAuthentication","additionalProperties":{"customKey":"customValue"}}`, noAuthDestName)
	destinationCreatorBasicAuthDestReqBody     = fmt.Sprintf(`{"name":"%s","url":"http://localhost","type":"HTTP","proxyType":"Internet","authenticationType":"BasicAuthentication","user":"my-first-user","password":"secretPassword","additionalProperties":{"%s":"value"}}`, basicAuthDestName, correlationIDsKey)
	destinationCreatorSAMLAssertionDestReqBody = fmt.Sprintf(`{"name":"%s","url":"https://localhost","type":"HTTP","proxyType":"Internet","authenticationType":"SAMLAssertion","audience":"https://localhost","keyStoreLocation":"test.jks","additionalProperties":{"%s":"value"}}`, samlAssertionDestName, correlationIDsKey)

	destinationServiceBasicAuthReqBody = fmt.Sprintf(`{"name":"%s","url":"http://localhost","type":"HTTP","proxyType":"Internet","authentication":"BasicAuthentication","user":"my-first-user","password":"secretPassword"}`, basicAuthDestName)

	destinationCreatorReqBodyWithoutAuthType = fmt.Sprintf(`{"name":"%s","url":"http://localhost","type":"HTTP","proxyType":"Internet","additionalProperties":{"customKey":"customValue"}}`, noAuthDestName)

	destinationCreatorCertReqBody                             = fmt.Sprintf(`{"name":"%s"}`, testDestinationCertName)
	destinationServiceCertResponseBody                        = fmt.Sprintf(`{"Name":"%s","Content":"%s"}`, testDestinationCertWithExtension, testCertChain)
	destinationServiceSAMLDestCertResponseBody                = fmt.Sprintf(`{"Name":"%s","Content":"%s"}`, testDestKeyStoreLocation, testCertChain)
	destinationServiceFindAPIResponseBodyForSAMLAssertionDest = fmt.Sprintf(esmdestinationcreator.FindAPISAMLAssertionDestResponseTemplate, testSubaccountID, testServiceInstanceID, samlAssertionDestName, testDestType, testSecureDestURL, "SAMLAssertion", testDestProxyType, testSecureDestURL, testDestKeyStoreLocation, testDestKeyStoreLocation)
)

func fixNoAuthDestination(name string) esmdestcreatorpkg.Destination {
	return &esmdestcreatorpkg.NoAuthenticationDestination{
		Name:           name,
		URL:            testDestURL,
		Type:           destinationcreatorpkg.Type(testDestType),
		ProxyType:      destinationcreatorpkg.ProxyType(testDestProxyType),
		Authentication: "NoAuthentication",
	}
}

func fixBasicDestination(name string) esmdestcreatorpkg.Destination {
	return &esmdestcreatorpkg.BasicDestination{
		NoAuthenticationDestination: esmdestcreatorpkg.NoAuthenticationDestination{
			Name:           name,
			URL:            testDestURL,
			Type:           destinationcreatorpkg.Type(testDestType),
			ProxyType:      destinationcreatorpkg.ProxyType(testDestProxyType),
			Authentication: "BasicAuthentication",
		},
		User:     "my-first-user",
		Password: "secretPassword",
	}
}

func fixSAMLAssertionDestination(name string) esmdestcreatorpkg.Destination {
	return &esmdestcreatorpkg.SAMLAssertionDestination{
		NoAuthenticationDestination: esmdestcreatorpkg.NoAuthenticationDestination{
			Name:           name,
			URL:            testSecureDestURL,
			Type:           destinationcreatorpkg.Type(testDestType),
			ProxyType:      destinationcreatorpkg.ProxyType(testDestProxyType),
			Authentication: "SAMLAssertion",
		},
		Audience:         testSecureDestURL,
		KeyStoreLocation: testDestKeyStoreLocation,
	}
}

func fixDestinationMappings(destName string, destination esmdestcreatorpkg.Destination) map[string]esmdestcreatorpkg.Destination {
	return map[string]esmdestcreatorpkg.Destination{
		destName: destination,
	}
}

func fixCertMappings(certName string, bodyBytes []byte) map[string]json.RawMessage {
	return map[string]json.RawMessage{
		certName: bodyBytes,
	}
}
