package document

import (
	"database/sql"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

//go:generate mockery -name=FetchRequestConverter -output=automock -outpkg=automock -case=underscore
type FetchRequestConverter interface {
	ToGraphQL(in *model.FetchRequest) *graphql.FetchRequest
	InputFromGraphQL(in *graphql.FetchRequestInput) *model.FetchRequestInput
}

type converter struct {
	frConverter FetchRequestConverter
}

func NewConverter(frConverter FetchRequestConverter) *converter {
	return &converter{frConverter: frConverter}
}

func (c *converter) ToGraphQL(in *model.Document) *graphql.Document {
	if in == nil {
		return nil
	}

	var clob *graphql.CLOB
	if in.Data != nil {
		tmp := graphql.CLOB([]byte(*in.Data))
		clob = &tmp
	}

	return &graphql.Document{
		ID:            in.ID,
		ApplicationID: in.ApplicationID,
		Title:         in.Title,
		DisplayName:   in.DisplayName,
		Description:   in.Description,
		Format:        graphql.DocumentFormat(in.Format),
		Kind:          in.Kind,
		Data:          clob,
		FetchRequest:  c.frConverter.ToGraphQL(in.FetchRequest),
	}
}

func (c *converter) MultipleToGraphQL(in []*model.Document) []*graphql.Document {
	var documents []*graphql.Document
	for _, r := range in {
		if r == nil {
			continue
		}

		documents = append(documents, c.ToGraphQL(r))
	}

	return documents
}

func (c *converter) InputFromGraphQL(in *graphql.DocumentInput) *model.DocumentInput {
	if in == nil {
		return nil
	}

	var data *string
	if in.Data != nil {
		tmp := string(*in.Data)
		data = &tmp
	}

	return &model.DocumentInput{
		Title:        in.Title,
		DisplayName:  in.DisplayName,
		Description:  in.Description,
		Format:       model.DocumentFormat(in.Format),
		Kind:         in.Kind,
		Data:         data,
		FetchRequest: c.frConverter.InputFromGraphQL(in.FetchRequest),
	}
}

func (c *converter) MultipleInputFromGraphQL(in []*graphql.DocumentInput) []*model.DocumentInput {
	var inputs []*model.DocumentInput
	for _, r := range in {
		if r == nil {
			continue
		}

		inputs = append(inputs, c.InputFromGraphQL(r))
	}

	return inputs
}

func (c *converter) ToEntity(in model.Document) (Entity, error) {
	var nullKind sql.NullString
	if in.Kind != nil && len(*in.Kind) > 0 {
		nullKind = sql.NullString{
			String: *in.Kind,
			Valid:  true,
		}
	}

	var nullData sql.NullString
	if in.Data != nil && len(*in.Data) > 0 {
		nullData = sql.NullString{
			String: *in.Kind,
			Valid:  true,
		}
	}

	out := Entity{
		ID:           in.ID,
		AppID:        in.ApplicationID,
		TenantID:     in.Tenant,
		Title:        in.Title,
		DisplayName:  in.DisplayName,
		Description:  in.Description,
		Format:       string(in.Format),
		Kind:         nullKind,
		Data:         nullData,
		FetchRequest: sql.NullString{},
	}

	return out, nil
}

func (c *converter) FromEntity(in Entity) (model.Document, error) {
	out := model.Document{
		ID:            in.ID,
		ApplicationID: in.AppID,
		Tenant:        in.TenantID,
		Title:         in.Title,
		DisplayName:   in.DisplayName,
		Description:   in.Description,
		Format:        model.DocumentFormat(in.Format),
		Kind:          &in.Kind.String,
		Data:          &in.Data.String,
		FetchRequest:  nil,
	}
	return out, nil
}
