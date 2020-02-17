package model

import (
	"fmt"

	schema "github.com/kyma-incubator/compass/components/connector/pkg/graphql/externalschema"
)

type InfoProviderFunc func(applicationName string, eventServiceBaseURL, tenant string, configuration schema.Configuration) interface{}

func NewCSRInfoResponseProvider(connectivityAdapterBaseURL, connectivityAdapterMTLSBaseURL string) InfoProviderFunc {

	return func(applicationName, eventServiceBaseURL, tenant string, configuration schema.Configuration) interface{} {
		csrURL := connectivityAdapterBaseURL + CertsEndpoint
		//TODO: handle case when configuration.Token is nil
		tokenParam := fmt.Sprintf(TokenFormat, configuration.Token.Token)

		api := Api{
			CertificatesURL: connectivityAdapterBaseURL + CertsEndpoint,
			InfoURL:         connectivityAdapterMTLSBaseURL + ManagementInfoEndpoint,
			RuntimeURLs:     makeRuntimeURLs(applicationName, connectivityAdapterMTLSBaseURL, eventServiceBaseURL),
		}

		return CSRInfoResponse{
			CsrURL:          csrURL + tokenParam,
			API:             api,
			CertificateInfo: ToCertInfo(configuration.CertificateSigningRequestInfo),
		}
	}
}

func NewManagementInfoResponseProvider(connectivityAdapterMTLSBaseURL string) InfoProviderFunc {

	return func(applicationName, eventServiceBaseURL, tenant string, configuration schema.Configuration) interface{} {

		clientIdentity := ClientIdentity{
			Application: applicationName,
			Tenant:      tenant,
			Group:       "",
		}

		managementURLs := MgmtURLs{
			RuntimeURLs:   makeRuntimeURLs(applicationName, connectivityAdapterMTLSBaseURL, eventServiceBaseURL),
			RenewCertURL:  fmt.Sprintf(RenewCertURLFormat, connectivityAdapterMTLSBaseURL),
			RevokeCertURL: fmt.Sprintf(RevocationCertURLFormat, connectivityAdapterMTLSBaseURL),
		}

		return MgmtInfoReponse{
			ClientIdentity:  clientIdentity,
			URLs:            managementURLs,
			CertificateInfo: ToCertInfo(configuration.CertificateSigningRequestInfo),
		}
	}
}

func makeRuntimeURLs(application, connectivityAdapterMTLSBaseURL string, eventServiceBaseURL string) *RuntimeURLs {
	return &RuntimeURLs{
		MetadataURL:   connectivityAdapterMTLSBaseURL + fmt.Sprintf(ApplicationRegistryEndpointFormat, application),
		EventsURL:     eventServiceBaseURL,
		EventsInfoURL: eventServiceBaseURL + EventsInfoEndpoint,
	}
}
