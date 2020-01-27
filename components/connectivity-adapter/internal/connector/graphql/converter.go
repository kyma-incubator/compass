package graphql

import (
	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connector/model"
	schema "github.com/kyma-incubator/compass/components/connector/pkg/graphql/externalschema"
)

func ToCertInfo(configuration schema.Configuration) model.CertInfo {
	return model.CertInfo{
		Subject:      configuration.CertificateSigningRequestInfo.Subject,
		KeyAlgorithm: configuration.CertificateSigningRequestInfo.KeyAlgorithm,
	}
}

func ToCertResponse(result schema.CertificationResult) model.CertResponse {
	return model.CertResponse{
		CRTChain:  result.CertificateChain,
		CaCRT:     result.CaCertificate,
		ClientCRT: result.ClientCertificate,
	}
}
