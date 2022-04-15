package runtimemapping

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	schema "github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/kyma-incubator/compass/components/hydrator/pkg/oathkeeper"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/form3tech-oss/jwt-go"
	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/pkg/errors"
)

// TokenVerifier missing godoc
//go:generate mockery --name=TokenVerifier --output=automock --outpkg=automock --case=underscore
type TokenVerifier interface {
	Verify(ctx context.Context, token string) (*jwt.MapClaims, error)
}

// ReqDataParser missing godoc
//go:generate mockery --name=ReqDataParser --output=automock --outpkg=automock --case=underscore
type ReqDataParser interface {
	Parse(req *http.Request) (oathkeeper.ReqData, error)
}

// DirectorClient missing godoc
//go:generate mockery --name=DirectorClient --output=automock --outpkg=automock --case=underscore
type DirectorClient interface {
	GetTenantByInternalID(ctx context.Context, tenantID string) (*schema.Tenant, error)
	GetTenantByLowestOwnerForResource(ctx context.Context, resourceID, resourceType string) (string, error)
	GetRuntimeByTokenIssuer(ctx context.Context, issuer string) (*schema.Runtime, error)
}

// Handler missing godoc
type Handler struct {
	reqDataParser  ReqDataParser
	directorClient DirectorClient
	tokenVerifier  TokenVerifier
}

// NewHandler missing godoc
func NewHandler(
	reqDataParser ReqDataParser,
	directorClient DirectorClient,
	tokenVerifier TokenVerifier) *Handler {
	return &Handler{
		reqDataParser:  reqDataParser,
		directorClient: directorClient,
		tokenVerifier:  tokenVerifier,
	}
}

// ServeHTTP missing godoc
func (h *Handler) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(writer, fmt.Sprintf("Bad request method. Got %s, expected POST", req.Method), http.StatusBadRequest)
		return
	}

	ctx := req.Context()

	reqData, err := h.reqDataParser.Parse(req)
	if err != nil {
		h.logError(ctx, err, "An error has occurred while parsing the request.")
		h.respond(ctx, writer, oathkeeper.ReqBody{})
		return
	}

	err = h.processRequest(ctx, &reqData)
	if err != nil {
		h.logError(ctx, err, "An error has occurred while processing the request.")
		h.respond(ctx, writer, reqData.Body)
		return
	}

	h.respond(ctx, writer, reqData.Body)
}

func (h *Handler) processRequest(ctx context.Context, reqData *oathkeeper.ReqData) error {
	claims, err := h.tokenVerifier.Verify(ctx, reqData.Header.Get("Authorization"))
	if err != nil {
		return errors.Wrap(err, "while verifying the token")
	}

	issuer, err := getTokenIssuer(*claims)
	if err != nil {
		return errors.Wrap(err, "unable to get the issuer")
	}

	runtime, err := h.directorClient.GetRuntimeByTokenIssuer(ctx, issuer)
	if err != nil {
		return errors.Wrap(err, "when getting the runtime")
	}

	internalTenantID, err := h.directorClient.GetTenantByLowestOwnerForResource(ctx, runtime.ID, string(resource.Runtime))
	if err != nil {
		return errors.Wrapf(err, "while getting lowest tenant for runtime %s", runtime.ID)
	}

	extTenant, err := h.directorClient.GetTenantByInternalID(ctx, internalTenantID)
	if err != nil {
		return errors.Wrap(err, "unable to fetch external tenant based on runtime tenant")
	}

	reqData.SetExternalTenantID(extTenant.ID)
	reqData.SetExtraFromClaims(*claims)
	return nil
}

func (h *Handler) logError(ctx context.Context, err error, message string) {
	log.C(ctx).WithError(err).Error(message)
}

func (h *Handler) respond(ctx context.Context, writer http.ResponseWriter, body oathkeeper.ReqBody) {
	writer.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(writer).Encode(body)
	if err != nil {
		h.logError(ctx, err, "An error has occurred while encoding data.")
	}
}
