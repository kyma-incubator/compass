package oauth

const (
	GrantTypeFieldName   = "grant_type"
	CredentialsGrantType = "client_credentials"

	ScopeFieldName = "scope"
	Scopes         = "runtime:read runtime:write label_definition:read label_definition:write"

	ContentTypeHeader      = "Content-Type"
	FormEncodedContentType = "application/x-www-form-urlencoded"
)
