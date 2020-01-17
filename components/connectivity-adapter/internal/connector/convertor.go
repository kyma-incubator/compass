package connector

import schema "github.com/kyma-incubator/compass/components/connector/pkg/graphql/externalschema"

func ToCertInfo(configuration schema.Configuration) CertInfo {
	return CertInfo{
		Subject:      configuration.CertificateSigningRequestInfo.Subject,
		KeyAlgorithm: configuration.CertificateSigningRequestInfo.KeyAlgorithm,
	}
}

func ToCertResponse(result schema.CertificationResult) CertResponse {
	return CertResponse{
		CRTChain:  result.CertificateChain,
		CaCRT:     result.CaCertificate,
		ClientCRT: result.ClientCertificate,
	}
}
