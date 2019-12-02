package connector

import "fmt"

func configurationQuery() string {
	return fmt.Sprintf(`query {
		result: configuration() {
			%s
		}
	}`, configurationResult())
}

func configurationResult() string {
	return `token { 
		token 
	}
	certificateSigningRequestInfo { 
		subject 
		keyAlgorithm 
	}
	managementPlaneInfo { 
		directorURL
		certificateSecuredConnectorURL
	}`
}

func signCSRMutation(csr string) string {
	return fmt.Sprintf(`mutation {
		result: signCertificateSigningRequest(csr: "%s") {
			%s
		}
	}`, csr, certificationResult())
}

func certificationResult() string {
	return `certificateChain
	caCertificate
	clientCertificate`
}
