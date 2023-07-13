package destinationcreator

const (
	// TypeHTTP represents the HTTP destination type
	TypeHTTP Type = "HTTP"
	// TypeRFC represents the RFC destination type
	TypeRFC Type = "RFC"
	// TypeLDAP represents the LDAP destination type
	TypeLDAP Type = "LDAP"
	// TypeMAIL represents the MAIL destination type
	TypeMAIL Type = "MAIL"

	// AuthTypeNoAuth represents the NoAuth destination authentication
	AuthTypeNoAuth AuthType = "NoAuthentication"
	// AuthTypeBasic represents the Basic destination authentication
	AuthTypeBasic AuthType = "BasicAuthentication"
	// AuthTypeSAMLAssertion represents the SAMLAssertion destination authentication
	AuthTypeSAMLAssertion AuthType = "SAMLAssertion"
	// AuthTypeSAMLBearer represents the SAMLBearer destination authentication
	AuthTypeSAMLBearer AuthType = "OAuth2SAMLBearerAssertion"

	// ProxyTypeInternet represents the Internet proxy type
	ProxyTypeInternet ProxyType = "Internet"
	// ProxyTypeOnPremise represents the OnPremise proxy type
	ProxyTypeOnPremise ProxyType = "OnPremise"
	// ProxyTypePrivateLink represents the PrivateLink proxy type
	ProxyTypePrivateLink ProxyType = "PrivateLink"

	JavaKeyStoreFileExtension = ".jks"
)

// Type represents the destination type
type Type string

// AuthType represents the destination authentication type
type AuthType string

// ProxyType represents the destination proxy type
type ProxyType string

// NoAuthenticationDestination is structure representing a no authentication destination entity and its data from the remote destination service
type NoAuthenticationDestination struct {
	Name           string    `json:"name"`
	Type           Type      `json:"type"`
	URL            string    `json:"url"`
	Authentication AuthType  `json:"authentication"`
	ProxyType      ProxyType `json:"proxyType"`
}

// BasicDestination is structure representing a basic destination entity and its data from the remote destination service
type BasicDestination struct {
	NoAuthenticationDestination
	User     string `json:"user"`
	Password string `json:"password"`
}

// SAMLAssertionDestination is structure representing a SAML assertion destination entity and its data from the remote destination service
type SAMLAssertionDestination struct {
	NoAuthenticationDestination
	Audience         string `json:"audience"`
	KeyStoreLocation string `json:"keyStoreLocation"`
}

// DestinationSvcCertificateResponse contains the response data from destination service certificate request
type DestinationSvcCertificateResponse struct {
	Name    string `json:"Name"`
	Content string `json:"Content"`
}
