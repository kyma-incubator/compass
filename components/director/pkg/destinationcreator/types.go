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
	// AuthTypeSAMLBearerAssertion represents the OAuth2SAMLBearerAssertion destination authentication
	AuthTypeSAMLBearerAssertion AuthType = "OAuth2SAMLBearerAssertion"
	// AuthTypeClientCertificate represents the ClientCertificate destination authentication
	AuthTypeClientCertificate AuthType = "ClientCertificateAuthentication"

	// ProxyTypeInternet represents the Internet proxy type
	ProxyTypeInternet ProxyType = "Internet"
	// ProxyTypeOnPremise represents the OnPremise proxy type
	ProxyTypeOnPremise ProxyType = "OnPremise"
	// ProxyTypePrivateLink represents the PrivateLink proxy type
	ProxyTypePrivateLink ProxyType = "PrivateLink"

	// MaxDestinationNameLength is the maximum length for destination name resources - certificate and destination names
	MaxDestinationNameLength = 64
	// JavaKeyStoreFileExtension is a java key store extension name
	JavaKeyStoreFileExtension = ".jks"
)

// Type represents the destination type
type Type string

// AuthType represents the destination authentication type
type AuthType string

// ProxyType represents the destination proxy type
type ProxyType string
