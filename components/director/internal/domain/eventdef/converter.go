package eventdef

import (
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/internal/domain/version"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/pkg/errors"
)

// VersionConverter converts Version between model.Version, graphql.Version and repo-layer version.Version
//go:generate mockery --name=VersionConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type VersionConverter interface {
	ToGraphQL(in *model.Version) *graphql.Version
	InputFromGraphQL(in *graphql.VersionInput) *model.VersionInput
	FromEntity(version version.Version) *model.Version
	ToEntity(version model.Version) version.Version
}

// SpecConverter converts Specifications between the model.Spec service-layer representation and the graphql-layer representation graphql.EventSpec.
//go:generate mockery --name=SpecConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type SpecConverter interface {
	ToGraphQLEventSpec(in *model.Spec) (*graphql.EventSpec, error)
	InputFromGraphQLEventSpec(in *graphql.EventSpecInput) (*model.SpecInput, error)
}

type converter struct {
	vc VersionConverter
	sc SpecConverter
}

// NewConverter returns a new Converter that can later be used to make the conversions between the GraphQL, service, and repository layer representations of a Compass EventDefinition.
func NewConverter(vc VersionConverter, sc SpecConverter) *converter {
	return &converter{vc: vc, sc: sc}
}

// ToGraphQL converts the provided service-layer representation of an EventDefinition to the graphql-layer one.
func (c *converter) ToGraphQL(in *model.EventDefinition, spec *model.Spec, bundleRef *model.BundleReference) (*graphql.EventDefinition, error) {
	if in == nil {
		return nil, nil
	}

	var bundleID string
	if bundleRef.BundleID != nil {
		bundleID = *bundleRef.BundleID
	}

	return &graphql.EventDefinition{
		BundleID:    bundleID,
		Name:        in.Name,
		Description: in.Description,
		Group:       in.Group,
		Version:     c.vc.ToGraphQL(in.Version),
		BaseEntity: &graphql.BaseEntity{
			ID:        in.ID,
			Ready:     in.Ready,
			CreatedAt: timePtrToTimestampPtr(in.CreatedAt),
			UpdatedAt: timePtrToTimestampPtr(in.UpdatedAt),
			DeletedAt: timePtrToTimestampPtr(in.DeletedAt),
			Error:     in.Error,
		},
	}, nil
}

// MultipleToGraphQL converts the provided service-layer representations of an EventDefinition to the graphql-layer ones.
func (c *converter) MultipleToGraphQL(in []*model.EventDefinition, specs []*model.Spec, bundleRefs []*model.BundleReference) ([]*graphql.EventDefinition, error) {
	if len(in) != len(bundleRefs) {
		return nil, errors.New("different events, specs and bundleRefs count provided")
	}

	events := make([]*graphql.EventDefinition, 0, len(in))
	for i, e := range in {
		if e == nil {
			continue
		}

		event, err := c.ToGraphQL(e, nil, bundleRefs[i])
		if err != nil {
			return nil, err
		}

		events = append(events, event)
	}

	return events, nil
}

// MultipleInputFromGraphQL converts the provided graphql-layer representations of an EventDefinition to the service-layer ones.
func (c *converter) MultipleInputFromGraphQL(in []*graphql.EventDefinitionInput) ([]*model.EventDefinitionInput, []*model.SpecInput, error) {
	eventDefs := make([]*model.EventDefinitionInput, 0, len(in))
	specs := make([]*model.SpecInput, 0, len(in))

	for _, item := range in {
		event, spec, err := c.InputFromGraphQL(item)
		if err != nil {
			return nil, nil, err
		}

		eventDefs = append(eventDefs, event)
		specs = append(specs, spec)
	}

	return eventDefs, specs, nil
}

// InputFromGraphQL converts the provided graphql-layer representation of an EventDefinition to the service-layer one.
func (c *converter) InputFromGraphQL(in *graphql.EventDefinitionInput) (*model.EventDefinitionInput, *model.SpecInput, error) {
	if in == nil {
		return nil, nil, nil
	}

	spec, err := c.sc.InputFromGraphQLEventSpec(in.Spec)
	if err != nil {
		return nil, nil, err
	}

	return &model.EventDefinitionInput{
		Name:         in.Name,
		Description:  in.Description,
		Group:        in.Group,
		VersionInput: c.vc.InputFromGraphQL(in.Version),
	}, spec, nil
}

// FromEntity converts the provided Entity repo-layer representation of an EventDefinition to the service-layer representation model.EventDefinition.
func (c *converter) FromEntity(entity *Entity) *model.EventDefinition {
	return &model.EventDefinition{
		ApplicationID:       entity.ApplicationID,
		PackageID:           repo.StringPtrFromNullableString(entity.PackageID),
		Name:                entity.Name,
		Description:         repo.StringPtrFromNullableString(entity.Description),
		Group:               repo.StringPtrFromNullableString(entity.GroupName),
		OrdID:               repo.StringPtrFromNullableString(entity.OrdID),
		ShortDescription:    repo.StringPtrFromNullableString(entity.ShortDescription),
		SystemInstanceAware: repo.BoolPtrFromNullableBool(entity.SystemInstanceAware),
		Tags:                repo.JSONRawMessageFromNullableString(entity.Tags),
		Countries:           repo.JSONRawMessageFromNullableString(entity.Countries),
		Links:               repo.JSONRawMessageFromNullableString(entity.Links),
		ReleaseStatus:       repo.StringPtrFromNullableString(entity.ReleaseStatus),
		SunsetDate:          repo.StringPtrFromNullableString(entity.SunsetDate),
		Successors:          repo.JSONRawMessageFromNullableString(entity.Successors),
		ChangeLogEntries:    repo.JSONRawMessageFromNullableString(entity.ChangeLogEntries),
		Labels:              repo.JSONRawMessageFromNullableString(entity.Labels),
		Visibility:          &entity.Visibility,
		Disabled:            repo.BoolPtrFromNullableBool(entity.Disabled),
		PartOfProducts:      repo.JSONRawMessageFromNullableString(entity.PartOfProducts),
		LineOfBusiness:      repo.JSONRawMessageFromNullableString(entity.LineOfBusiness),
		Industry:            repo.JSONRawMessageFromNullableString(entity.Industry),
		Version:             c.vc.FromEntity(entity.Version),
		Extensible:          repo.JSONRawMessageFromNullableString(entity.Extensible),
		ResourceHash:        repo.StringPtrFromNullableString(entity.ResourceHash),
		DocumentationLabels: repo.JSONRawMessageFromNullableString(entity.DocumentationLabels),
		BaseEntity: &model.BaseEntity{
			ID:        entity.ID,
			Ready:     entity.Ready,
			CreatedAt: entity.CreatedAt,
			UpdatedAt: entity.UpdatedAt,
			DeletedAt: entity.DeletedAt,
			Error:     repo.StringPtrFromNullableString(entity.Error),
		},
	}
}

// ToEntity converts the provided service-layer representation of an EventDefinition to the repository-layer one.
func (c *converter) ToEntity(eventModel *model.EventDefinition) *Entity {
	visibility := eventModel.Visibility
	if visibility == nil {
		visibility = str.Ptr("public")
	}

	return &Entity{
		ApplicationID:       eventModel.ApplicationID,
		PackageID:           repo.NewNullableString(eventModel.PackageID),
		Name:                eventModel.Name,
		Description:         repo.NewNullableString(eventModel.Description),
		GroupName:           repo.NewNullableString(eventModel.Group),
		OrdID:               repo.NewNullableString(eventModel.OrdID),
		ShortDescription:    repo.NewNullableString(eventModel.ShortDescription),
		SystemInstanceAware: repo.NewNullableBool(eventModel.SystemInstanceAware),
		Tags:                repo.NewNullableStringFromJSONRawMessage(eventModel.Tags),
		Countries:           repo.NewNullableStringFromJSONRawMessage(eventModel.Countries),
		Links:               repo.NewNullableStringFromJSONRawMessage(eventModel.Links),
		ReleaseStatus:       repo.NewNullableString(eventModel.ReleaseStatus),
		SunsetDate:          repo.NewNullableString(eventModel.SunsetDate),
		Successors:          repo.NewNullableStringFromJSONRawMessage(eventModel.Successors),
		ChangeLogEntries:    repo.NewNullableStringFromJSONRawMessage(eventModel.ChangeLogEntries),
		Labels:              repo.NewNullableStringFromJSONRawMessage(eventModel.Labels),
		Visibility:          *visibility,
		Disabled:            repo.NewNullableBool(eventModel.Disabled),
		PartOfProducts:      repo.NewNullableStringFromJSONRawMessage(eventModel.PartOfProducts),
		LineOfBusiness:      repo.NewNullableStringFromJSONRawMessage(eventModel.LineOfBusiness),
		Industry:            repo.NewNullableStringFromJSONRawMessage(eventModel.Industry),
		Version:             c.convertVersionToEntity(eventModel.Version),
		Extensible:          repo.NewNullableStringFromJSONRawMessage(eventModel.Extensible),
		ResourceHash:        repo.NewNullableString(eventModel.ResourceHash),
		DocumentationLabels: repo.NewNullableStringFromJSONRawMessage(eventModel.DocumentationLabels),
		BaseEntity: &repo.BaseEntity{
			ID:        eventModel.ID,
			Ready:     eventModel.Ready,
			CreatedAt: eventModel.CreatedAt,
			UpdatedAt: eventModel.UpdatedAt,
			DeletedAt: eventModel.DeletedAt,
			Error:     repo.NewNullableString(eventModel.Error),
		},
	}
}

func (c *converter) convertVersionToEntity(inVer *model.Version) version.Version {
	if inVer == nil {
		return version.Version{}
	}

	return c.vc.ToEntity(*inVer)
}

func timePtrToTimestampPtr(time *time.Time) *graphql.Timestamp {
	if time == nil {
		return nil
	}

	t := graphql.Timestamp(*time)
	return &t
}
