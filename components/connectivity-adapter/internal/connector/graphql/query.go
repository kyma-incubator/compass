package graphql

import "fmt"

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
    }`, csr)
}
