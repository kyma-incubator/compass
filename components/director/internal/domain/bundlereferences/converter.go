package bundlereferences

import (
	"database/sql"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/pkg/errors"
)

type converter struct{}

func NewConverter() *converter {
	return &converter{}
}

func (c *converter) ToEntity(in model.BundleReference) Entity {
	var apiDefID sql.NullString
	var eventDefID sql.NullString
	var apiDefaultTargetURL sql.NullString

	switch in.ObjectType {
	case model.BundleAPIReference:
		apiDefID = repo.NewNullableString(in.ObjectID)
		apiDefaultTargetURL = repo.NewNullableString(in.APIDefaultTargetURL)
	case model.BundleEventReference:
		eventDefID = repo.NewNullableString(in.ObjectID)
	}

	return Entity{
		TenantID:            in.Tenant,
		BundleID:            repo.NewNullableString(in.BundleID),
		APIDefID:            apiDefID,
		EventDefID:          eventDefID,
		APIDefaultTargetURL: apiDefaultTargetURL,
	}
}

func (c *converter) FromEntity(in Entity) (model.BundleReference, error) {
	objectID, objectType, err := c.objectReferenceFromEntity(in)
	if err != nil {
		return model.BundleReference{}, errors.Wrap(err, "while determining object reference")
	}

	return model.BundleReference{
		Tenant:              in.TenantID,
		BundleID:            repo.StringPtrFromNullableString(in.BundleID),
		ObjectType:          objectType,
		ObjectID:            repo.StringPtrFromNullableString(objectID),
		APIDefaultTargetURL: repo.StringPtrFromNullableString(in.APIDefaultTargetURL),
	}, nil
}

func (c *converter) objectReferenceFromEntity(in Entity) (sql.NullString, model.BundleReferenceObjectType, error) {
	if in.APIDefID.Valid {
		return in.APIDefID, model.BundleAPIReference, nil
	}

	if in.EventDefID.Valid {
		return in.EventDefID, model.BundleEventReference, nil
	}

	return sql.NullString{}, "", fmt.Errorf("incorrect Object Reference ID and its type for Reference Entity with bundle ID %q", *repo.StringPtrFromNullableString(in.BundleID))
}
