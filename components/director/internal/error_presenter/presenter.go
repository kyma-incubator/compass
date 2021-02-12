package error_presenter

import (
	"context"
	"errors"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/gqlerror"
)

type UUIDService interface {
	Generate() string
}

type presenter struct {
	uuidService UUIDService
}

func NewPresenter(service UUIDService) *presenter {
	return &presenter{uuidService: service}
}

func (p *presenter) Do(ctx context.Context, err error) *gqlerror.Error {
	customErr := apperrors.Error{}
	errID := p.uuidService.Generate()

	if found := errors.As(err, &customErr); !found {
		log.C(ctx).WithField("errorID", errID).WithError(err).Error("Unknown error")
		return newGraphqlErrorResponse(ctx, apperrors.InternalError, "Internal Server Error [errorID=%s]", errID)
	}

	if apperrors.ErrorCode(customErr) == apperrors.InternalError {
		log.C(ctx).WithField("errorID", errID).WithError(err).Error("Internal Server Error")
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
