package identification

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

// GetFromRequest retrieves identification for a request, i.e. like 'client_id'.
func GetFromRequest(r *http.Request) string {
	ctx := r.Context()
	unknownIdentity := "Unknown"

	// Valid Oathkeeper Subject flow
	reqDataParser := doathkeeper.NewReqDataParser()
	reqData, err := reqDataParser.Parse(r)
	if err != nil {
		log.C(ctx).Errorf("An error occurred while parsing request to extract identification: %s", err)
		return unknownIdentity
	} else if len(reqData.Body.Subject) > 0 {
		if email := reqData.Body.Extra["email"]; email != nil {
			return email.(string)
		}
		return reqData.Body.Subject
	}

	// JWT flow
	authorizationHeader := reqData.Header.Get("Authorization")
	if authorizationHeader != "" && strings.HasPrefix(strings.ToLower(authorizationHeader), "bearer ") {
		token := strings.TrimSpace(authorizationHeader[len("Bearer "):])

		parsedToken, _, err := (&jwt.Parser{}).ParseUnverified(token, jwt.MapClaims{})
		if err != nil {
			log.C(ctx).Error("Failed to parse JWT token: ", err)
			return unknownIdentity
		}

		if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok && len(claims["sub"].(string)) > 0 {
			return claims["sub"].(string)
		} else {
			log.C(ctx).Error("Failed to get subject claim from JWT token: ", err)
			return unknownIdentity
		}
	}

	// Certificates flow
	certHeaderParser := coathkeeper.NewHeaderParser("Certificate-Data", "", nil, func(subject string) string { return subject })
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
