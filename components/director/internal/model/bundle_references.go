package model

import "github.com/pkg/errors"

type BundleReference struct {
	Tenant              string
	BundleID            *string
	ObjectType          BundleReferenceObjectType
	ObjectID            *string
	APIDefaultTargetURL *string
}

type BundleReferenceObjectType string

const (
	BundleAPIReference   BundleReferenceObjectType = "API"
	BundleEventReference BundleReferenceObjectType = "Event"
)

type BundleReferenceInput struct {
	APIDefaultTargetURL *string
}

func (b *BundleReferenceInput) ToBundleReference(tenant string, objectType BundleReferenceObjectType, bundleID, objectID *string) (*BundleReference, error) {
	if b == nil {
		return nil, nil
	}

	if objectType == BundleAPIReference && b.APIDefaultTargetURL == nil {
		return nil, errors.New("default targetURL for API cannot be empty")
	}

	return &BundleReference{
		Tenant:              tenant,
		BundleID:            bundleID,
		ObjectType:          objectType,
		ObjectID:            objectID,
		APIDefaultTargetURL: b.APIDefaultTargetURL,
	}, nil
}
