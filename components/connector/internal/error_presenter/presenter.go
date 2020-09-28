package error_presenter

import (
	"context"
	"errors"
	"fmt"

	"github.com/kyma-incubator/compass/components/connector/internal/apperrors"

	log "github.com/sirupsen/logrus"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/gqlerror"
)

type UUIDService interface {
	Generate() string
}

type Presenter struct {
	uuidService UUIDService
	Logger      *log.Logger
}

func NewPresenter(logger *log.Logger, service UUIDService) *Presenter {
	return &Presenter{Logger: logger, uuidService: service}
}

func (p *Presenter) Do(ctx context.Context, err error) *gqlerror.Error {
	customErr := apperrors.Error{}
	errID := p.uuidService.Generate()
	if found := errors.As(err, &customErr); !found {
		p.Logger.WithField("errorID", errID).Infof("Unknown error: %s", err.Error())
		return newGraphqlErrorResponse(ctx, errID, "Internal Server Error")
	}

	if apperrors.ErrorCode(customErr) == apperrors.CodeInternal {
		p.Logger.WithField("errorID", errID).Infof("Internal Server Error: %s", err.Error())
		return newGraphqlErrorResponse(ctx, errID, "Internal Server Error")
	}

	p.Logger.WithField("errorID", errID).Info(err.Error())
	return newGraphqlErrorResponse(ctx, errID, err.Error())
}

func newGraphqlErrorResponse(ctx context.Context, errID string, msg string, args ...interface{}) *gqlerror.Error {
	return &gqlerror.Error{
		Message:    fmt.Sprintf(msg, args...),
		Path:       graphql.GetResolverContext(ctx).Path(),
		Extensions: map[string]interface{}{"error_id": errID},
	}
}
