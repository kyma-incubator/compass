package destinationcreator

import destinationcreatorpkg "github.com/kyma-incubator/compass/components/director/pkg/destinationcreator"

// NoAuthenticationDestination is a structure representing a no authentication destination entity and its data from the remote destination service
type NoAuthenticationDestination struct {
	Name           string                          `json:"name"`
	URL            string                          `json:"url"`
	SubaccountID   string                          `json:"subaccountId"`
	InstanceID     string                          `json:"instanceId"`
	Type           destinationcreatorpkg.Type      `json:"type"`
	ProxyType      destinationcreatorpkg.ProxyType `json:"proxyType"`
	Authentication destinationcreatorpkg.AuthType  `json:"authentication"`
}

// BasicDestination is a structure representing a basic destination entity and its data from the remote destination service
type BasicDestination struct {
	NoAuthenticationDestination
	User     string `json:"user"`
	Password string `json:"password"`
}

// SAMLAssertionDestination is a structure representing a SAML assertion destination entity and its data from the remote destination service
type SAMLAssertionDestination struct {
	NoAuthenticationDestination
	Audience         string `json:"audience"`
	KeyStoreLocation string `json:"keyStoreLocation"`
}

// ClientCertificateAuthenticationDestination is a structure representing a client certificate authentication destination entity and its data from the remote destination service
type ClientCertificateAuthenticationDestination struct {
	NoAuthenticationDestination
	KeyStoreLocation string `json:"keyStoreLocation"`
}

// DestinationSvcCertificateResponse contains the response data from destination service certificate request
type DestinationSvcCertificateResponse struct {
	Name    string `json:"Name"`
	Content string `json:"Content"`
}
