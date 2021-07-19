package internal_auth

import (
    "context"
    "encoding/json"
    "fmt"
    "github.com/form3tech-oss/jwt-go"
    "github.com/kyma-incubator/compass/components/director/internal/consumer"
    "github.com/kyma-incubator/compass/components/director/internal/model"
    "github.com/kyma-incubator/compass/components/director/internal/oathkeeper"
    "github.com/kyma-incubator/compass/components/director/pkg/apperrors"
    "github.com/kyma-incubator/compass/components/director/pkg/authenticator"
    "github.com/kyma-incubator/compass/components/director/pkg/log"
    "github.com/kyma-incubator/compass/components/director/pkg/persistence"
    "github.com/lestrrat-go/iter/arrayiter"
    "github.com/lestrrat-go/jwx/jwk"
    "github.com/pkg/errors"
    "net/http"
)

const claimsIssuerKey = "iss"

//go:generate mockery --name=TokenVerifier --output=automock --outpkg=automock --case=underscore
type TokenVerifier interface {
    Verify(ctx context.Context, token string) (*jwt.MapClaims, error)
}

//go:generate mockery --name=ReqDataParser --output=automock --outpkg=automock --case=underscore
type ReqDataParser interface {
    Parse(req *http.Request) (oathkeeper.ReqData, error)
}

//go:generate mockery --name=TenantRepository --output=automock --outpkg=automock --case=underscore
type TenantRepository interface {
    GetByExternalTenant(ctx context.Context, externalTenant string) (*model.BusinessTenantMapping, error)
}

type Handler struct {
    reqDataParser ReqDataParser
    tenantRepo    TenantRepository
    transact      persistence.Transactioner
}

func NewHandler(
    reqDataParser ReqDataParser,
    tenantRepo    TenantRepository,
    transact      persistence.Transactioner,
) *Handler {
    return &Handler{
        reqDataParser: reqDataParser,
        tenantRepo: tenantRepo,
        transact: transact,
    }
}

type TenantContext struct {
    ExternalTenantID string
    TenantID         string
}

func NewTenantContext(externalTenantID, tenantID string) TenantContext {
    return TenantContext{
        ExternalTenantID: externalTenantID,
        TenantID:         tenantID,
    }
}

type Claims struct {
    Tenant         string                `json:"tenant"`
    ExternalTenant string                `json:"externalTenant"`
    Scopes         string                `json:"scopes"`
    ConsumerID     string                `json:"consumerID"`
    ConsumerType   consumer.ConsumerType `json:"consumerType"`
    jwt.StandardClaims
}

func (h *Handler) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
    ctx := req.Context()
    log.C(ctx).Info("Enter into internal auth handler!")

    if req.Method != http.MethodPost {
        http.Error(writer, fmt.Sprintf("Bad request method. Got %s, expected POST", req.Method), http.StatusBadRequest)
        return
    }

    reqData, err := h.reqDataParser.Parse(req)
    if err != nil {
        logError(ctx, err, "An error has occurred while parsing the request.")
        respond(ctx, writer, oathkeeper.ReqBody{})
        return
    }

    log.C(ctx).Info("Parsed request data")
    log.C(ctx).Infof("reqData headers --> %s", reqData.Header)
    log.C(ctx).Infof("reqData body subject --> %s", reqData.Body.Subject)
    log.C(ctx).Infof("reqData body extra --> %v", reqData.Body.Extra)
    log.C(ctx).Infof("reqData body headers --> %v", reqData.Body.Header)

    //err = h.verifyToken(ctx, &reqData)
    //if err != nil {
    //   logError(ctx, err, "An error has occurred while verifying the token.")
    //   respond(ctx, writer, reqData.Body)
    //   return
    //}
    //log.C(ctx).Info("token is verified!")

    tc, err := h.getTenantContext(ctx, &reqData)
    if err != nil {
        logError(ctx, err, "An error has occurred while getting tenant context.")
        respond(ctx, writer, oathkeeper.ReqBody{})
        return
    }

    log.C(ctx).Infof("Tenant ID: %s", tc.TenantID)
    log.C(ctx).Infof("External Tenant ID: %s", tc.ExternalTenantID)

    reqData.Body.Extra["tenant"] = tc.TenantID
    reqData.Body.Extra["externalTenant"] = tc.ExternalTenantID

    respond(ctx, writer, reqData.Body)
}

func (h *Handler) getTenantContext(ctx context.Context, reqData *oathkeeper.ReqData) (TenantContext, error) {
    log.C(ctx).Info("Enter into getTenantContext function")

    tx, err := h.transact.Begin()
    if err != nil {
        log.C(ctx).WithError(err).Errorf("An error occurred while opening db transaction: %v", err)
        return TenantContext{}, err
    }
    defer h.transact.RollbackUnlessCommitted(ctx, tx)

    newCtx := persistence.SaveToContext(ctx, tx)

    externalTenantID, err := reqData.GetExternalTenantID()
    if err != nil {
        log.C(newCtx).Errorf("Could not get tenant external id, error: %s", err.Error())
        return TenantContext{}, errors.Wrap(err, "while getting external tenant ID [ExternalTenantId=%s]")
    }

    log.C(newCtx).Infof("Getting the tenant with external ID: %s", externalTenantID)
    tenantMapping, err := h.tenantRepo.GetByExternalTenant(newCtx, externalTenantID)
    if err != nil {
        log.C(newCtx).Errorf("Could not find tenant with external ID: %s, error: %s", externalTenantID, err.Error())
        return TenantContext{}, err
    }

    if err := tx.Commit(); err != nil {
        log.C(newCtx).WithError(err).Errorf("An error occurred while committing transaction: %v", err)
        return TenantContext{}, err
    }

        return NewTenantContext(externalTenantID, tenantMapping.ID), nil
}

func (h *Handler) verifyToken(ctx context.Context, reqData *oathkeeper.ReqData) error {
    bearerToken := reqData.Header.Get("X-Internal-Identity")
    if bearerToken == "" {
        return apperrors.NewUnauthorizedError("token cannot be empty")
    }

    log.C(ctx).Info("Parsing claims...")

    _, err := parseClaims(ctx, bearerToken)
    if err != nil {
        return apperrors.NewUnauthorizedError(err.Error())
    }

    //claims, err := h.tokenVerifier.Verify(ctx, reqData.Header.Get("X-Internal-Identity"))
    //if err != nil {
    //    return errors.Wrap(err, "while verifying the token")
    //}

   return nil
}

func logError(ctx context.Context, err error, message string) {
    log.C(ctx).WithError(err).Error(message)
}

func respond(ctx context.Context, writer http.ResponseWriter, body oathkeeper.ReqBody) {
    writer.Header().Set("Content-Type", "application/json")
    err := json.NewEncoder(writer).Encode(body)
    if err != nil {
        logError(ctx, err, "An error has occurred while encoding data.")
    }
}

func  parseClaims(ctx context.Context, bearerToken string) (Claims, error) {
    claims := Claims{}
    _, err := jwt.ParseWithClaims(bearerToken, &claims, getKeyFunc(ctx))
    if err != nil {
        log.C(ctx).Info("Error while parsing claims!")
        return Claims{}, err
    }

    log.C(ctx).Info("Successfully parsed claims")

    return claims, nil
}

func getKeyFunc(ctx context.Context) func(token *jwt.Token) (interface{}, error) {
    return func(token *jwt.Token) (interface{}, error) {
        if token.Method.Alg() != jwt.SigningMethodRS256.Name {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Method.Alg())
        }

        keyID, err := getKeyID(*token)
        if err != nil {
            log.C(ctx).WithError(err).Errorf("An error occurred while getting the token signing key ID: %v", err)
            return nil, errors.Wrap(err, "while getting the key ID")
        }

        keyIterator := &authenticator.JWTKeyIterator{
            AlgorithmCriteria: func(alg string) bool {
                return token.Method.Alg() == alg
            },
            IDCriteria: func(id string) bool {
              return true
            },
        }

        //jwks, err := authenticator.FetchJWK(ctx, "https://container.googleapis.com/v1beta1/projects/sap-cp-cmp-dev/locations/europe-west1/clusters/sap-cp-cmp-dev-validation/jwks")
       jwks, err := jwk.Parse([]byte(`{
  "keys": [
    {
      "kty": "RSA",
      "kid": "fl0184hdz3XsHedwti0xlYFcLC4hGBVwCy8SEVRXABw",
      "n": "1KTVTHWfnQwUh3645qPOb6DekEBqUcigJUisQxkEYWiwtDnylNvPK2LvUPkZOBxuPWeFJh-KvJkqs2gWHzrPY0S2vxG25JRUPGM4KEysEr1fhT4tr1Lal-Fhr4ZW11OsZml_JFe-Asv2QfCG_H-5o4zufNoSs-tgKVvODVcqsvQZpjuVZygz2y8KB8UBjlPCPfzZg-P1T4X6XHlM_m_cPbtVIh4AAuxhtO_2g9HEIAkHanAmjyrn076Wq-5RAfeB6NGhop6V3L0pLTMXCuQpy6DoliZPIFjnYtmQMRzoLNIk16WFUtxr_GeGJ8-_SEIo3UpVP-dy-9VAQY5_TdRepQ",
      "e": "AQAB",
      "alg": "RS256",
      "use": "sig"
    }
  ]
}`))
        if err := arrayiter.Walk(ctx, jwks, keyIterator); err != nil {
            log.C(ctx).WithError(err).Errorf("An error occurred while walking through the JWKS: %v", err)
            return nil, err
        }

        if keyIterator.ResultingKey == nil {
            log.C(ctx).Debugf("Signing key %s is not found", keyID)
            return nil, apperrors.NewKeyDoesNotExistError(keyID)
        }

        return keyIterator.ResultingKey, nil

    }
}

func getKeyID(token jwt.Token) (string, error) {
    keyID, ok := token.Header["kid"]
    if !ok {
        return "", apperrors.NewInternalError("unable to find the key ID in the token")
    }

    keyIDStr, ok := keyID.(string)
    if !ok {
        return "", apperrors.NewInternalError("unable to cast the key ID to a string")
    }

    return keyIDStr, nil
}

func (c Claims) Valid() error {
    err := c.StandardClaims.Valid()
    if err != nil {
        return err
    }

    return nil
}