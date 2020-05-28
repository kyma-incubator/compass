package tmp_resolver

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/customerrors"
	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
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

	statusCode := customerrors.ErrorType(errID)

	switch statusCode {
	case customerrors.InternalError:
		{
			err = customerrors.NewInternalError("while producing error")
			err = errors.Wrap(err, "while doing 1st internal err")
			err = errors.Wrap(err, "while doing 2st internal err")
		}
	case customerrors.NotUnique:
		{
			err = customerrors.NewNotUniqueErr(customerrors.Package)
			err = errors.Wrap(err, "while doing 1st not unique")
			err = errors.Wrap(err, "while doing 2st not unique")
		}
	case customerrors.NotFound:
		err = customerrors.NewNotFoundError(customerrors.Application, "uuuid")
	case customerrors.UnknownError:
		err = errors.New("error which is not know to the library")
		err = errors.Wrap(err, "while doing 1st thing")
	case customerrors.TenantIsRequired:
		err = customerrors.NewTenantNotFound("tenant")
		err = errors.Wrap(err, "while doing 1st not found")
		err = errors.Wrap(err, "while doing 2st not found")
	case customerrors.InvalidData:
		err = customerrors.NewInvalidDataError("field name is not valid")
		err = errors.Wrap(err, "while doing 1st invalid data")
		err = errors.Wrap(err, "while doing 2st invalid data")
	default:
		panic(errors.New("Panic!"))
	}

	return &graphql.Tenant{
		ID:         "id",
		InternalID: "",
		Name:       str.Ptr(string(errID)),
	}, err
}
