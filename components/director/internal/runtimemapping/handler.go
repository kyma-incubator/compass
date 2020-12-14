package runtimemapping

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/form3tech-oss/jwt-go"
	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/oathkeeper"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
)

//go:generate mockery -name=TokenVerifier -output=automock -outpkg=automock -case=underscore
type TokenVerifier interface {
	Verify(ctx context.Context, token string) (*jwt.MapClaims, error)
}

//go:generate mockery -name=RuntimeService -output=automock -outpkg=automock -case=underscore
type RuntimeService interface {
	GetByTokenIssuer(ctx context.Context, issuer string) (*model.Runtime, error)
}

//go:generate mockery -name=TenantService -output=automock -outpkg=automock -case=underscore
type TenantService interface {
	GetExternalTenant(ctx context.Context, id string) (string, error)
}

//go:generate mockery -name=ReqDataParser -output=automock -outpkg=automock -case=underscore
type ReqDataParser interface {
	Parse(req *http.Request) (oathkeeper.ReqData, error)
}

type Handler struct {
	reqDataParser ReqDataParser
	transact      persistence.Transactioner
	tokenVerifier TokenVerifier
	runtimeSvc    RuntimeService
	tenantSvc     TenantService
}

func NewHandler(
	reqDataParser ReqDataParser,
	transact persistence.Transactioner,
	tokenVerifier TokenVerifier,
	runtimeSvc RuntimeService,
	tenantSvc TenantService) *Handler {
	return &Handler{
		reqDataParser: reqDataParser,
		transact:      transact,
		tokenVerifier: tokenVerifier,
		runtimeSvc:    runtimeSvc,
		tenantSvc:     tenantSvc,
	}
}

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

	tx, err := h.transact.Begin()
	if err != nil {
		h.logError(ctx, err, "An error has occurred while opening the db transaction.")
		h.respond(ctx, writer, reqData.Body)
		return
	}
	defer h.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(req.Context(), tx)

	err = h.processRequest(ctx, &reqData)
	if err != nil {
		h.logError(ctx, err, "An error has occurred while processing the request.")
		h.respond(ctx, writer, reqData.Body)
		return
	}

	if err = tx.Commit(); err != nil {
		h.logError(ctx, err, "An error has occurred while committing the transaction.")
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

	runtime, err := h.runtimeSvc.GetByTokenIssuer(ctx, issuer)
	if err != nil {
		return errors.Wrap(err, "when getting the runtime")
	}

	extTenantID, err := h.tenantSvc.GetExternalTenant(ctx, runtime.Tenant)
	if err != nil {
		return errors.Wrap(err, "unable to fetch external tenant based on runtime tenant")
	}

	reqData.SetExternalTenantID(extTenantID)
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
