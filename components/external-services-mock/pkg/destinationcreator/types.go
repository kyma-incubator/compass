package destinationcreator

import destinationcreatorpkg "github.com/kyma-incubator/compass/components/director/pkg/destinationcreator"

const (
	DesignTimeDestinationType    = "design time"
	BasicAuthDestinationType     = "basic"
	SAMLAssertionDestinationType = "SAML assertion"
	ClientCertDestinationType    = "client certificate authentication"
)

type Destination interface {
	GetType() string
}

// NoAuthenticationDestination is a structure representing a no authentication destination entity and its data from the remote destination service
type NoAuthenticationDestination struct {
	Name           string                          `json:"name"`
	URL            string                          `json:"url"`
	Type           destinationcreatorpkg.Type      `json:"type"`
	ProxyType      destinationcreatorpkg.ProxyType `json:"proxyType"`
	Authentication destinationcreatorpkg.AuthType  `json:"authentication"`
}

func (n *NoAuthenticationDestination) GetType() string {
	return DesignTimeDestinationType
}

// BasicDestination is a structure representing a basic destination entity and its data from the remote destination service
type BasicDestination struct {
	NoAuthenticationDestination
	User     string `json:"user"`
	Password string `json:"password"`
}

func (b *BasicDestination) GetType() string {
	return BasicAuthDestinationType
}

// SAMLAssertionDestination is a structure representing a SAML assertion destination entity and its data from the remote destination service
type SAMLAssertionDestination struct {
	NoAuthenticationDestination
	Audience         string `json:"audience"`
	KeyStoreLocation string `json:"keyStoreLocation"`
}

func (s *SAMLAssertionDestination) GetType() string {
	return SAMLAssertionDestinationType
}

// ClientCertificateAuthenticationDestination is a structure representing a client certificate authentication destination entity and its data from the remote destination service
type ClientCertificateAuthenticationDestination struct {
	NoAuthenticationDestination
	KeyStoreLocation string `json:"keyStoreLocation"`
}

func (c *ClientCertificateAuthenticationDestination) GetType() string {
	return ClientCertDestinationType
}

// DestinationSvcCertificateResponse contains the response data from destination service certificate request
type DestinationSvcCertificateResponse struct {
	Name    string `json:"Name"`
	Content string `json:"Content"`
}

// OwnerDetails contains data about the destination owner and is used in the response data from destination service 'find API'
type OwnerDetails struct {
	SubaccountID string `json:"SubaccountId"`
	InstanceID   string `json:"InstanceId"`
}

// AuthTokensDetails contains data about the destination auth tokens and is used in the response data from destination service 'find API'
type AuthTokensDetails struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// CertificateDetails contains data about the destination certificate and is used in the response data from destination service 'find API'
type CertificateDetails struct {
	Name    string `json:"Name"`
	Content string `json:"Content"`
}

// DestinationSvcNoAuthenticationDestResponse contains the response data from destination service 'find API' request for destination of type 'NoAuthentication'
type DestinationSvcNoAuthenticationDestResponse struct {
	Owner                    OwnerDetails                `json:"owner"`
	DestinationConfiguration NoAuthenticationDestination `json:"destinationConfiguration"`
}

// DestinationSvcBasicDestResponse contains the response data from destination service 'find API' request for destination of type 'BasicAuthentication'
type DestinationSvcBasicDestResponse struct {
	Owner                    OwnerDetails        `json:"owner"`
	DestinationConfiguration BasicDestination    `json:"destinationConfiguration"`
	AuthTokens               []AuthTokensDetails `json:"authTokens"`
}

// DestinationSvcSAMLAssertionDestResponse contains the response data from destination service 'find API' request for destination of type 'SAMLAssertion'
type DestinationSvcSAMLAssertionDestResponse struct {
	Owner                    OwnerDetails             `json:"owner"`
	DestinationConfiguration SAMLAssertionDestination `json:"destinationConfiguration"`
	CertificateDetails       []CertificateDetails     `json:"certificates"`
	AuthTokens               []AuthTokensDetails      `json:"authTokens"`
}

// DestinationSvcClientCertDestResponse contains the response data from destination service 'find API' request for destination of type 'ClientCertificate'
type DestinationSvcClientCertDestResponse struct {
	Owner                    OwnerDetails                               `json:"owner"`
	DestinationConfiguration ClientCertificateAuthenticationDestination `json:"destinationConfiguration"`
	CertificateDetails       []CertificateDetails                       `json:"certificates"`
}
