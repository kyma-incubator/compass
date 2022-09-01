package bundle

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/pkg/errors"
)

// AuthConverter missing godoc
//go:generate mockery --name=AuthConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type AuthConverter interface {
	ToGraphQL(in *model.Auth) (*graphql.Auth, error)
	InputFromGraphQL(in *graphql.AuthInput) (*model.AuthInput, error)
}

type converter struct {
	auth     AuthConverter
	api      APIConverter
	event    EventConverter
	document DocumentConverter
}

// NewConverter missing godoc
func NewConverter(auth AuthConverter, api APIConverter, event EventConverter, document DocumentConverter) *converter {
	return &converter{
		auth:     auth,
		api:      api,
		event:    event,
		document: document,
	}
}

// ToEntity missing godoc
func (c *converter) ToEntity(in *model.Bundle) (*Entity, error) {
	if in == nil {
		return nil, nil
	}

	defaultInstanceAuth, err := c.marshalDefaultInstanceAuth(in.DefaultInstanceAuth)
	if err != nil {
		return nil, err
	}

	output := &Entity{
		ApplicationID:                 in.ApplicationID,
		Name:                          in.Name,
		Description:                   repo.NewNullableString(in.Description),
		InstanceAuthRequestJSONSchema: repo.NewNullableString(in.InstanceAuthRequestInputSchema),
		DefaultInstanceAuth:           repo.NewNullableString(defaultInstanceAuth),
		OrdID:                         repo.NewNullableString(in.OrdID),
		ShortDescription:              repo.NewNullableString(in.ShortDescription),
		Links:                         repo.NewNullableStringFromJSONRawMessage(in.Links),
		Labels:                        repo.NewNullableStringFromJSONRawMessage(in.Labels),
		CredentialExchangeStrategies:  repo.NewNullableStringFromJSONRawMessage(in.CredentialExchangeStrategies),
		CorrelationIDs:                repo.NewNullableStringFromJSONRawMessage(in.CorrelationIDs),
		DocumentationLabels:           repo.NewNullableStringFromJSONRawMessage(in.DocumentationLabels),
		BaseEntity: &repo.BaseEntity{
			ID:        in.ID,
			Ready:     in.Ready,
			CreatedAt: in.CreatedAt,
			UpdatedAt: in.UpdatedAt,
			DeletedAt: in.DeletedAt,
			Error:     repo.NewNullableString(in.Error),
		},
	}

	return output, nil
}

// FromEntity missing godoc
func (c *converter) FromEntity(entity *Entity) (*model.Bundle, error) {
	if entity == nil {
		return nil, apperrors.NewInternalError("the Bundle entity is nil")
	}

	defaultInstanceAuth, err := c.unmarshalDefaultInstanceAuth(entity.DefaultInstanceAuth)
	if err != nil {
		return nil, err
	}

	output := &model.Bundle{
		ApplicationID:                  entity.ApplicationID,
		Name:                           entity.Name,
		Description:                    repo.StringPtrFromNullableString(entity.Description),
		InstanceAuthRequestInputSchema: repo.StringPtrFromNullableString(entity.InstanceAuthRequestJSONSchema),
		DefaultInstanceAuth:            defaultInstanceAuth,
		OrdID:                          repo.StringPtrFromNullableString(entity.OrdID),
		ShortDescription:               repo.StringPtrFromNullableString(entity.ShortDescription),
		Links:                          repo.JSONRawMessageFromNullableString(entity.Links),
		Labels:                         repo.JSONRawMessageFromNullableString(entity.Labels),
		CredentialExchangeStrategies:   repo.JSONRawMessageFromNullableString(entity.CredentialExchangeStrategies),
		CorrelationIDs:                 repo.JSONRawMessageFromNullableString(entity.CorrelationIDs),
		DocumentationLabels:            repo.JSONRawMessageFromNullableString(entity.DocumentationLabels),
		BaseEntity: &model.BaseEntity{
			ID:        entity.ID,
			Ready:     entity.Ready,
			CreatedAt: entity.CreatedAt,
			UpdatedAt: entity.UpdatedAt,
			DeletedAt: entity.DeletedAt,
			Error:     repo.StringPtrFromNullableString(entity.Error),
		},
	}

	return output, nil
}

// ToGraphQL missing godoc
func (c *converter) ToGraphQL(in *model.Bundle) (*graphql.Bundle, error) {
	if in == nil {
		return nil, apperrors.NewInternalError("the model Bundle is nil")
	}

	auth, err := c.auth.ToGraphQL(in.DefaultInstanceAuth)
	if err != nil {
		return nil, errors.Wrap(err, "while converting DefaultInstanceAuth to GraphQL")
	}

	correlationIDs, err := c.correlationIDsToStringSlice(in.CorrelationIDs)
	if err != nil {
		return nil, errors.Wrap(err, "while converting CorrelationIDs to GraphQL")
	}
	return &graphql.Bundle{
		Name:                           in.Name,
		Description:                    in.Description,
		InstanceAuthRequestInputSchema: c.strPtrToJSONSchemaPtr(in.InstanceAuthRequestInputSchema),
		DefaultInstanceAuth:            auth,
		CorrelationIDs:                 correlationIDs,
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

// MultipleToGraphQL missing godoc
func (c *converter) MultipleToGraphQL(in []*model.Bundle) ([]*graphql.Bundle, error) {
	bundles := make([]*graphql.Bundle, 0, len(in))
	for _, r := range in {
		if r == nil {
			continue
		}
		bndl, err := c.ToGraphQL(r)
		if err != nil {
			return nil, errors.Wrap(err, "while converting Bundle to GraphQL")
		}
		bundles = append(bundles, bndl)
	}

	return bundles, nil
}

// CreateInputFromGraphQL missing godoc
func (c *converter) CreateInputFromGraphQL(in graphql.BundleCreateInput) (model.BundleCreateInput, error) {
	auth, err := c.auth.InputFromGraphQL(in.DefaultInstanceAuth)
	if err != nil {
		return model.BundleCreateInput{}, errors.Wrap(err, "while converting DefaultInstanceAuth input")
	}

	apiDefs, apiSpecs, err := c.api.MultipleInputFromGraphQL(in.APIDefinitions)
	if err != nil {
		return model.BundleCreateInput{}, errors.Wrap(err, "while converting APIDefinitions input")
	}

	documents, err := c.document.MultipleInputFromGraphQL(in.Documents)
	if err != nil {
		return model.BundleCreateInput{}, errors.Wrap(err, "while converting Documents input")
	}

	eventDefs, eventSpecs, err := c.event.MultipleInputFromGraphQL(in.EventDefinitions)
	if err != nil {
		return model.BundleCreateInput{}, errors.Wrap(err, "while converting EventDefinitions input")
	}

	correlationIDs, err := c.correlationIDsToJSONRaw(in.CorrelationIDs)
	if err != nil {
		return model.BundleCreateInput{}, errors.Wrap(err, "while converting CorrelationIDs input")
	}
	return model.BundleCreateInput{
		Name:                           in.Name,
		Description:                    in.Description,
		InstanceAuthRequestInputSchema: c.jsonSchemaPtrToStrPtr(in.InstanceAuthRequestInputSchema),
		DefaultInstanceAuth:            auth,
		APIDefinitions:                 apiDefs,
		APISpecs:                       apiSpecs,
		EventDefinitions:               eventDefs,
		EventSpecs:                     eventSpecs,
		Documents:                      documents,
		CorrelationIDs:                 correlationIDs,
	}, nil
}

// MultipleCreateInputFromGraphQL missing godoc
func (c *converter) MultipleCreateInputFromGraphQL(in []*graphql.BundleCreateInput) ([]*model.BundleCreateInput, error) {
	bundles := make([]*model.BundleCreateInput, 0, len(in))
	for _, item := range in {
		if item == nil {
			continue
		}
		bndl, err := c.CreateInputFromGraphQL(*item)
		if err != nil {
			return nil, err
		}
		bundles = append(bundles, &bndl)
	}

	return bundles, nil
}

// UpdateInputFromGraphQL missing godoc
func (c *converter) UpdateInputFromGraphQL(in graphql.BundleUpdateInput) (*model.BundleUpdateInput, error) {
	auth, err := c.auth.InputFromGraphQL(in.DefaultInstanceAuth)
	if err != nil {
		return nil, errors.Wrap(err, "while converting DefaultInstanceAuth from GraphQL")
	}

	return &model.BundleUpdateInput{
		Name:                           in.Name,
		Description:                    in.Description,
		InstanceAuthRequestInputSchema: c.jsonSchemaPtrToStrPtr(in.InstanceAuthRequestInputSchema),
		DefaultInstanceAuth:            auth,
	}, nil
}

func (c *converter) marshalDefaultInstanceAuth(defaultInstanceAuth *model.Auth) (*string, error) {
	if defaultInstanceAuth == nil {
		return nil, nil
	}

	output, err := json.Marshal(defaultInstanceAuth)
	if err != nil {
		return nil, errors.Wrap(err, "while marshaling default auth")
	}
	return str.Ptr(string(output)), nil
}

func (c *converter) unmarshalDefaultInstanceAuth(defaultInstanceAuthSQL sql.NullString) (*model.Auth, error) {
	var defaultInstanceAuth *model.Auth
	if defaultInstanceAuthSQL.Valid && defaultInstanceAuthSQL.String != "" {
		defaultInstanceAuth = &model.Auth{}
		err := json.Unmarshal([]byte(defaultInstanceAuthSQL.String), defaultInstanceAuth)
		if err != nil {
			return nil, errors.Wrap(err, "while unmarshalling default instance auth")
		}
	}

	return defaultInstanceAuth, nil
}

func (c *converter) strPtrToJSONSchemaPtr(in *string) *graphql.JSONSchema {
	if in == nil {
		return nil
	}
	out := graphql.JSONSchema(*in)
	return &out
}

func (c *converter) jsonSchemaPtrToStrPtr(in *graphql.JSONSchema) *string {
	if in == nil {
		return nil
	}
	out := string(*in)
	return &out
}

func (c *converter) correlationIDsToStringSlice(in json.RawMessage) ([]string, error) {
	if in == nil {
		return nil, nil
	}
	var out []string
	err := json.Unmarshal(in, &out)
	if err != nil {
		return nil, errors.Wrap(err, "while unmarshal correlation IDs")
	}
	return out, nil
}

func (c *converter) correlationIDsToJSONRaw(correlationIDs []string) (json.RawMessage, error) {
	if correlationIDs == nil {
		return nil, nil
	}
	out, err := json.Marshal(correlationIDs)
	if err != nil {
		return nil, errors.Wrap(err, "while marshalling correlation IDs")
	}
	return out, nil
}

func timePtrToTimestampPtr(time *time.Time) *graphql.Timestamp {
	if time == nil {
		return nil
	}

	t := graphql.Timestamp(*time)
	return &t
}
