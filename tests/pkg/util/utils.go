package util

type ApplicationType string

const (
	AuthorizationHeader = "Authorization"
	UserTokenHeader     = "X-user-token"
	ContentTypeHeader   = "Content-Type"

	ContentTypeApplicationJSON       = "application/json"
	ContentTypeApplicationURLEncoded = "application/x-www-form-urlencoded"

	ApplicationTypeC4C             ApplicationType = "SAP Cloud for Customer"
	ApplicationTypeS4HANAOnPremise ApplicationType = "SAP S/4HANA On-Premise"
)
