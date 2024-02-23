package destinationcreator

import "strings"

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
	// AuthTypeOAuth2ClientCredentials represents the OAuth2ClientCredentials destination authentication
	AuthTypeOAuth2ClientCredentials AuthType = "OAuth2ClientCredentials"
	// AuthTypeOAuth2mTLS represents the OAuth2 mTLS destination authentication
	AuthTypeOAuth2mTLS AuthType = "OAuth2mTLS"

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

	// SAMLAssertionDestPath represents the SAML Assertion destination type in the assignment config
	SAMLAssertionDestPath = "credentials.inboundCommunication.samlAssertion"
	// ClientCertAuthDestPath represents the client certificate authentication destination type in the assignment config
	ClientCertAuthDestPath = "credentials.inboundCommunication.clientCertificateAuthentication"
	// Oauth2mTLSAuthDestPath represents the oauth2mTLS authentication destination type in the assignment config
	Oauth2mTLSAuthDestPath = "credentials.inboundCommunication.oauth2mtls"

	// DedicatedTokenServiceURLType represents the 'Dedicated' token service URL type of OAuth2ClientCredentials destination
	DedicatedTokenServiceURLType TokenServiceURLType = "Dedicated"
	// CommonTokenServiceURLType represents the 'Common' token service URL type of OAuth2ClientCredentials destination
	CommonTokenServiceURLType TokenServiceURLType = "Common"

	correlationIDDelimiter = ","
)

// Type represents the destination type
type Type string

// AuthType represents the destination authentication type
type AuthType string

// ProxyType represents the destination proxy type
type ProxyType string

// TokenServiceURLType represents the token service URL type of OAuth2ClientCredentials destination
type TokenServiceURLType string

// DestinationInfo holds information about some destination fields
// these fields are calculated before calling DestinationCreator and then are passed to it
// however, after we call the DestinationCreator, we have to store the given destination in our db; we want to reuse the already calculated fields via this struct so that we can use them when storing the destination in the db
type DestinationInfo struct {
	AuthenticationType AuthType
	Type               Type
	URL                string
}

func ConstructCorrelationIDsString(correlationIDs []string) string {
	return strings.Join(correlationIDs, correlationIDDelimiter)
}

func DeconstructCorrelationIDs(correlationIDs string) []string {
	return strings.Split(correlationIDs, correlationIDDelimiter)
}
