package model

import (
	schema "github.com/kyma-incubator/compass/components/connector/pkg/graphql/externalschema"
)

func ToCertInfo(signingRequestInfo *schema.CertificateSigningRequestInfo) CertInfo {
	if signingRequestInfo == nil {
		return CertInfo{}
	}

	return CertInfo{
		Subject:      signingRequestInfo.Subject,
		KeyAlgorithm: signingRequestInfo.KeyAlgorithm,
	}
}

func ToCertResponse(certificationResult schema.CertificationResult) CertResponse {
	return CertResponse{
		CRTChain:  certificationResult.CertificateChain,
		CaCRT:     certificationResult.CaCertificate,
		ClientCRT: certificationResult.ClientCertificate,
	}
}
