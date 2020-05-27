package customerrors

import (
	"context"
	"errors"
	"fmt"

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
			fmt.Printf("Internal Error, errorID:%s, %s\n", errID, err.Error())
			return NewInternalErrResponse(ctx, errID)
		}

		return &gqlerror.Error{
			Message:    customErr.Error(),
			Path:       graphql.GetResolverContext(ctx).Path(),
			Extensions: map[string]interface{}{"error_code": customErr.errorCode, "error": customErr.errorCode.String()},
		}
	}
	errID := p.uuidService.Generate()
	fmt.Printf("Not handled error yet, errorID:%s, %v\n", errID, err)
	return &gqlerror.Error{
		Message:    fmt.Sprintf("%s, errorID:%s", err.Error(), errID),
		Path:       graphql.GetResolverContext(ctx).Path(),
		Extensions: map[string]interface{}{"error_code": UnhandledError, "error": UnhandledError.String()},
	}
}

func NewInternalErrResponse(ctx context.Context, uuid string) *gqlerror.Error {
	return &gqlerror.Error{
		Message:    fmt.Sprintf("Internal Error Server, errorID:%s", uuid),
		Path:       graphql.GetResolverContext(ctx).Path(),
		Extensions: map[string]interface{}{"error_code": InternalError, "error": InternalError.String()},
	}
}
