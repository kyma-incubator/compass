package connector

import schema "github.com/kyma-incubator/compass/components/connector/pkg/graphql/externalschema"

func ToCertInfo(configuration schema.Configuration) CertInfo {
	return CertInfo{
		Subject:      configuration.CertificateSigningRequestInfo.Subject,
		KeyAlgorithm: configuration.CertificateSigningRequestInfo.KeyAlgorithm,
	}
}
