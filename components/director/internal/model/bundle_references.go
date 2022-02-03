package model

import "github.com/pkg/errors"

// BundleReference missing godoc
type BundleReference struct {
	ID                  string
	BundleID            *string
	ObjectType          BundleReferenceObjectType
	ObjectID            *string
	APIDefaultTargetURL *string
	IsDefaultBundle     bool
	Visibility          string
}

// BundleReferenceObjectType missing godoc
type BundleReferenceObjectType string

const (
	// BundleAPIReference missing godoc
	BundleAPIReference BundleReferenceObjectType = "API"
	// BundleEventReference missing godoc
	BundleEventReference BundleReferenceObjectType = "Event"

	publicVisibility string = "public"
)

// BundleReferenceInput missing godoc
type BundleReferenceInput struct {
	APIDefaultTargetURL *string
	IsDefaultBundle     bool
	Visibility          *string
}

// ToBundleReference missing godoc
func (b *BundleReferenceInput) ToBundleReference(id string, objectType BundleReferenceObjectType, bundleID, objectID *string) (*BundleReference, error) {
	if b == nil {
		return nil, nil
	}

	if objectType == BundleAPIReference && b.APIDefaultTargetURL == nil {
		return nil, errors.New("default targetURL for API cannot be empty")
	}

	var visibility string
	if b.Visibility == nil {
		visibility = publicVisibility
	} else {
		visibility = *b.Visibility
	}

	return &BundleReference{
		ID:                  id,
		BundleID:            bundleID,
		ObjectType:          objectType,
		ObjectID:            objectID,
		APIDefaultTargetURL: b.APIDefaultTargetURL,
		IsDefaultBundle:     b.IsDefaultBundle,
		Visibility:          visibility,
	}, nil
}
