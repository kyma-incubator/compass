package eventdef

import (
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/domain/version"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/pkg/errors"
)

//go:generate mockery -name=VersionConverter -output=automock -outpkg=automock -case=underscore
type VersionConverter interface {
	ToGraphQL(in *model.Version) *graphql.Version
	InputFromGraphQL(in *graphql.VersionInput) *model.VersionInput
	FromEntity(version version.Version) *model.Version
	ToEntity(version model.Version) version.Version
}

//go:generate mockery -name=SpecConverter -output=automock -outpkg=automock -case=underscore
type SpecConverter interface {
	ToGraphQLEventSpec(in *model.Spec) (*graphql.EventSpec, error)
	InputFromGraphQLEventSpec(in *graphql.EventSpecInput) (*model.SpecInput, error)
}

type converter struct {
	vc VersionConverter
	sc SpecConverter
}

func NewConverter(vc VersionConverter, sc SpecConverter) *converter {
	return &converter{vc: vc, sc: sc}
}

func (c *converter) ToGraphQL(in *model.EventDefinition, spec *model.Spec) (*graphql.EventDefinition, error) {
	if in == nil {
		return nil, nil
	}

	s, err := c.sc.ToGraphQLEventSpec(spec)
	if err != nil {
		return nil, err
	}

	var bundleID string
	if in.BundleID != nil {
		bundleID = *in.BundleID
	}

	return &graphql.EventDefinition{
		BundleID:    bundleID,
		Name:        in.Name,
		Description: in.Description,
		Group:       in.Group,
		Spec:        s,
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

func (c *converter) MultipleToGraphQL(in []*model.EventDefinition, specs []*model.Spec) ([]*graphql.EventDefinition, error) {
	if len(in) != len(specs) {
		return nil, errors.New("different events and specs count provided")
	}

	var events []*graphql.EventDefinition
	for i, e := range in {
		if e == nil {
			continue
		}

		event, err := c.ToGraphQL(e, specs[i])
		if err != nil {
			return nil, err
		}

		events = append(events, event)
	}

	return events, nil
}

func (c *converter) MultipleInputFromGraphQL(in []*graphql.EventDefinitionInput) ([]*model.EventDefinitionInput, []*model.SpecInput, error) {
	var eventDefs []*model.EventDefinitionInput
	var specs []*model.SpecInput

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

func (c *converter) InputFromGraphQL(in *graphql.EventDefinitionInput) (*model.EventDefinitionInput, *model.SpecInput, error) {
	if in == nil {
		return nil, nil, nil
	}

	spec, err := c.sc.InputFromGraphQLEventSpec(in.Spec)
	if err != nil {
		return nil, nil, err
	}

	return &model.EventDefinitionInput{
		Name:        in.Name,
		Description: in.Description,
		Group:       in.Group,
		Version:     c.vc.InputFromGraphQL(in.Version),
	}, spec, nil
}

func (c *converter) FromEntity(entity Entity) model.EventDefinition {
	return model.EventDefinition{
		BundleID:            repo.StringPtrFromNullableString(entity.BundleID),
		PackageID:           repo.StringPtrFromNullableString(entity.PackageID),
		Tenant:              entity.TenantID,
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
		Successor:           repo.StringPtrFromNullableString(entity.Successor),
		ChangeLogEntries:    repo.JSONRawMessageFromNullableString(entity.ChangeLogEntries),
		Labels:              repo.JSONRawMessageFromNullableString(entity.Labels),
		Visibility:          repo.StringPtrFromNullableString(entity.Visibility),
		Disabled:            repo.BoolPtrFromNullableBool(entity.Disabled),
		PartOfProducts:      repo.JSONRawMessageFromNullableString(entity.PartOfProducts),
		LineOfBusiness:      repo.JSONRawMessageFromNullableString(entity.LineOfBusiness),
		Industry:            repo.JSONRawMessageFromNullableString(entity.Industry),
		Version:             c.vc.FromEntity(entity.Version),
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

func (c *converter) ToEntity(eventModel model.EventDefinition) Entity {
	return Entity{
		TenantID:            eventModel.Tenant,
		BundleID:            repo.NewNullableString(eventModel.BundleID),
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
		Successor:           repo.NewNullableString(eventModel.Successor),
		ChangeLogEntries:    repo.NewNullableStringFromJSONRawMessage(eventModel.ChangeLogEntries),
		Labels:              repo.NewNullableStringFromJSONRawMessage(eventModel.Labels),
		Visibility:          repo.NewNullableString(eventModel.Visibility),
		Disabled:            repo.NewNullableBool(eventModel.Disabled),
		PartOfProducts:      repo.NewNullableStringFromJSONRawMessage(eventModel.PartOfProducts),
		LineOfBusiness:      repo.NewNullableStringFromJSONRawMessage(eventModel.LineOfBusiness),
		Industry:            repo.NewNullableStringFromJSONRawMessage(eventModel.Industry),
		Version:             c.convertVersionToEntity(eventModel.Version),
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
