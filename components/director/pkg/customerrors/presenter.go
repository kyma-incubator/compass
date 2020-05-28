package customerrors

import (
	"context"
	"errors"
	"fmt"

	log "github.com/sirupsen/logrus"

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

func (p *presenter) ErrorPresenter(ctx context.Context, err error) *gqlerror.Error {
	customErr := Error{}
	if found := errors.As(err, &customErr); found {
		if customErr.errorCode == InternalError {
			errID := p.uuidService.Generate()
			log.WithField("errorID", errID).Infof("Internal Server Error: %s", err.Error())
			return newInternalErrResponse(ctx, errID)
		}

		return &gqlerror.Error{
			Message:    customErr.Error(),
			Path:       graphql.GetResolverContext(ctx).Path(),
			Extensions: map[string]interface{}{"error_code": customErr.errorCode, "error": customErr.errorCode.String()},
		}
	}
	log.Infof("Unknown error: %s\n", err.Error())
	return &gqlerror.Error{
		Message: err.Error(),
		Path:    graphql.GetResolverContext(ctx).Path(),
	}
}

func newInternalErrResponse(ctx context.Context, uuid string) *gqlerror.Error {
	return &gqlerror.Error{
		Message:    fmt.Sprintf("Internal Server Error [errorID=%s]", uuid),
		Path:       graphql.GetResolverContext(ctx).Path(),
		Extensions: map[string]interface{}{"error_code": InternalError, "error": InternalError.String()},
	}
}
