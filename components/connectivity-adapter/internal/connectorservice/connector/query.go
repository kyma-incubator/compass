package connector

import (
	"fmt"
	"strings"
	"unicode"
)

type queryProvider struct{}

func (qp queryProvider) configuration() string {
	return `query{
 		result: configuration()
        {
 			 token { token }
			 certificateSigningRequestInfo { subject keyAlgorithm }
			 managementPlaneInfo { directorURL certificateSecuredConnectorURL }
		}	
     }`
}

func (qp queryProvider) signCSR(csr string) string {
	return fmt.Sprintf(`mutation {
	result: signCertificateSigningRequest(csr: "%s")
  	{
	 	certificateChain
		caCertificate
		clientCertificate
	}
    }`, removeWhiteChars(csr))
}

func (qp queryProvider) token(application string) string {
	return fmt.Sprintf(`mutation {
    result: generateApplicationToken(appID: "%s")
  	{
    	token
  	}
	}`, application)
}

func (qp queryProvider) revoke() string {
	return `mutation {
    result: revokeCertificate 
	}`
}

func removeWhiteChars(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}

		return r
	}, s)
}
