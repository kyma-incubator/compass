package tracking

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/form3tech-oss/jwt-go"
	coathkeeper "github.com/kyma-incubator/compass/components/connector/pkg/oathkeeper"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	doathkeeper "github.com/kyma-incubator/compass/components/director/internal/oathkeeper"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
)

const (
	emailExtraKey          = "email"
	authorizationHeaderKey = "Authorization"
	bearerTokenPrefix      = "Bearer "
	certificateDataHeader  = "Certificate-Data"
	subjectClaimKey        = "sub"
)

// GetClientIDFromRequest extracts client ID from a request.
func GetClientIDFromRequest(r *http.Request, reqData doathkeeper.ReqData) string {
	ctx := r.Context()
	unknownIdentity := "Unknown"

	// Valid Oathkeeper Subject flow
	if len(reqData.Body.Subject) > 0 {
		if email := reqData.Body.Extra[emailExtraKey]; email != nil {
			return email.(string)
		}
		return reqData.Body.Subject
	}

	// JWT flow
	authorizationHeader := reqData.Header.Get(authorizationHeaderKey)
	if authorizationHeader != "" && strings.HasPrefix(strings.ToLower(authorizationHeader), strings.ToLower(bearerTokenPrefix)) {
		token := strings.TrimSpace(authorizationHeader[len(bearerTokenPrefix):])

		parsedToken, _, err := (&jwt.Parser{}).ParseUnverified(token, jwt.MapClaims{})
		if err != nil {
			log.C(ctx).Error("Failed to parse JWT token: ", err)
			return unknownIdentity
		}

		if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok && len(claims[subjectClaimKey].(string)) > 0 {
			return claims[subjectClaimKey].(string)
		} else {
			log.C(ctx).Error("Failed to get subject claim from JWT token: ", err)
			return unknownIdentity
		}
	}

	// Certificates flow
	certHeaderParser := coathkeeper.NewHeaderParser(certificateDataHeader, "", nil, func(subject string) string { return subject })
	cn, _, ok := certHeaderParser.GetCertificateData(r)
	if ok {
		return cn
	}

	// One-time-token flow
	connectorToken := r.Header.Get(coathkeeper.ConnectorTokenHeader)
	if connectorToken == "" {
		connectorToken = r.URL.Query().Get(coathkeeper.ConnectorTokenQueryParam)
	}
	if connectorToken != "" {
		decodedToken, err := base64.URLEncoding.DecodeString(connectorToken)
		if err != nil {
			log.C(ctx).Error("Failed to decode one-time-token")
			return unknownIdentity
		}

		var tokenData model.TokenData
		if err := json.Unmarshal(decodedToken, &tokenData); err != nil {
			log.C(ctx).Error("Failed to unmarshal one-time-token")
			return unknownIdentity
		}

		return tokenData.SystemAuthID
	}

	return unknownIdentity
}
