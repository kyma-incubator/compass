package error_presenter

import (
	"context"
	"errors"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	log "github.com/sirupsen/logrus"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/gqlerror"
)

type UUIDService interface {
	Generate() string
}

type presenter struct {
	uuidService UUIDService
	Logger      *log.Logger
}

func NewPresenter(logger *log.Logger, service UUIDService) *presenter {
	return &presenter{Logger: logger, uuidService: service}
}

func (p *presenter) Do(ctx context.Context, err error) *gqlerror.Error {
	customErr := apperrors.Error{}
	errID := p.uuidService.Generate()
	if found := errors.As(err, &customErr); !found {
		p.Logger.WithField("errorID", errID).Infof("Unknown error: %s\n", err.Error())
		return newGraphqlErrorResponse(ctx, apperrors.InternalError, "Internal Server Error [errorID=%s]", errID)
	}

	if apperrors.ErrorCode(customErr) == apperrors.InternalError {
		p.Logger.WithField("errorID", errID).Infof("Internal Server Error: %s", err.Error())
		return newGraphqlErrorResponse(ctx, apperrors.InternalError, "Internal Server Error [errorID=%s]", errID)
	}

	return newGraphqlErrorResponse(ctx, apperrors.ErrorCode(customErr), customErr.Error())
}

func newGraphqlErrorResponse(ctx context.Context, errCode apperrors.ErrorType, msg string, args ...interface{}) *gqlerror.Error {
	return &gqlerror.Error{
		Message:    fmt.Sprintf(msg, args...),
		Path:       graphql.GetResolverContext(ctx).Path(),
		Extensions: map[string]interface{}{"error_code": errCode, "error": errCode.String()},
	}
}
