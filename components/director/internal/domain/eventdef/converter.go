package eventdef

import (
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
	fr            FetchRequestConverter
	vc            VersionConverter
	sc SpecConverter
}

func NewConverter(fr FetchRequestConverter, vc VersionConverter, sc SpecConverter) *converter {
	return &converter{fr: fr, vc: vc, sc: sc}
}

func (c *converter) ToGraphQL(in *model.EventDefinition, spec *model.Spec) (*graphql.EventDefinition,error) {
	if in == nil {
		return nil, nil
	}

	s, err := c.sc.ToGraphQLEventSpec(spec)
	if err != nil {
		return nil, err
	}

	return &graphql.EventDefinition{
		ID:          in.ID,
		BundleID:    in.BundleID,
		Name:        in.Name,
		Description: in.Description,
		Group:       in.Group,
		Spec:        s,
		Version:     c.vc.ToGraphQL(in.Version),
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

func (c *converter) FromEntity(entity Entity) (model.EventDefinition, error) {
	return model.EventDefinition{
		ID:          entity.ID,
		Tenant:      entity.TenantID,
		BundleID:    entity.BndlID,
		Name:        entity.Name,
		Description: repo.StringPtrFromNullableString(entity.Description),
		Group:       repo.StringPtrFromNullableString(entity.GroupName),
		Version:     c.vc.FromEntity(entity.Version),
	}, nil
}

func (c *converter) ToEntity(eventModel model.EventDefinition) (Entity, error) {
	return Entity{
		ID:          eventModel.ID,
		TenantID:    eventModel.Tenant,
		BndlID:      eventModel.BundleID,
		Name:        eventModel.Name,
		Description: repo.NewNullableString(eventModel.Description),
		GroupName:   repo.NewNullableString(eventModel.Group),
		Version:     c.convertVersionToEntity(eventModel.Version),
	}, nil
}

func (c *converter) convertVersionToEntity(inVer *model.Version) version.Version {
	if inVer == nil {
		return version.Version{}
	}

	return c.vc.ToEntity(*inVer)
}
