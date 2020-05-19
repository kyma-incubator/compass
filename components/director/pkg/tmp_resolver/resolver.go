package tmp_resolver

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/customerrors"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/pkg/errors"
)

func Resolve(ctx context.Context, errID int) (*graphql.Tenant, error) {
	if errID == -1 {
		return &graphql.Tenant{
			ID:         "id",
			InternalID: "interal ID",
			Name:       str.Ptr("No error flow"),
		}, nil
	}
	var err error

	statusCode := customerrors.ErrorCode(errID)

	switch statusCode {
	case customerrors.InternalError:
		{
			err = customerrors.NewErrorBuilder(customerrors.InternalError).With("damian", "testuje").Build()
			err = errors.Wrap(err, "while doing 1st internal err")
			err = errors.Wrap(err, "while doing 2st internal err")
		}
	case customerrors.NotUnique:
		{
			err = customerrors.NewErrorBuilder(customerrors.NotUnique).With("damian", "testuje").Build()
			err = errors.Wrap(err, "while doing 1st not unique")
			err = errors.Wrap(err, "while doing 2st not unique")
		}
	case customerrors.NotFound:
		err = customerrors.NewErrorBuilder(customerrors.NotFound).With("object", "id").Build()
	case customerrors.UnhandledError:
		err = errors.New("error which is not handled")
		err = errors.Wrap(err, "while doing 1st thing")
	case customerrors.TenantNotFound:
		err = customerrors.NewErrorBuilder(customerrors.TenantNotFound).With("tenant", "not found").Build()
		err = errors.Wrap(err, "while doing 1st not found")
		err = errors.Wrap(err, "while doing 2st not found")
	case customerrors.InvalidData:
		err = customerrors.NewErrorBuilder(customerrors.InvalidData).With("invalid", "input").Build()
		err = errors.Wrap(err, "while doing 1st invalid data")
		err = errors.Wrap(err, "while doing 2st invalid data")
	}

	return &graphql.Tenant{
		ID:         "id",
		InternalID: "",
		Name:       str.Ptr(string(statusCode)),
	}, err
}
