package model

import (
	"errors"
	"fmt"

	schema "github.com/kyma-incubator/compass/components/connector/pkg/graphql/externalschema"
)

type InfoProviderFunc func(applicationName string, eventServiceBaseURL, tenant string, configuration schema.Configuration) (interface{}, error)

func NewCSRInfoResponseProvider(connectivityAdapterBaseURL, connectivityAdapterMTLSBaseURL string) InfoProviderFunc {

	return func(applicationName, eventServiceBaseURL, tenant string, configuration schema.Configuration) (interface{}, error) {
		if configuration.Token == nil {
			return nil, errors.New("empty token returned from Connector")
		}

		csrURL := connectivityAdapterBaseURL + CertsEndpoint

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
		}, nil
	}
}

func NewManagementInfoResponseProvider(connectivityAdapterMTLSBaseURL string) InfoProviderFunc {

	return func(applicationName, eventServiceBaseURL, tenant string, configuration schema.Configuration) (interface{}, error) {

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
		}, nil
	}
}

func makeRuntimeURLs(application, connectivityAdapterMTLSBaseURL string, eventServiceBaseURL string) *RuntimeURLs {
	return &RuntimeURLs{
		MetadataURL:   connectivityAdapterMTLSBaseURL + fmt.Sprintf(ApplicationRegistryEndpointFormat, application),
		EventsURL:     eventServiceBaseURL,
		EventsInfoURL: eventServiceBaseURL + EventsInfoEndpoint,
	}
}
