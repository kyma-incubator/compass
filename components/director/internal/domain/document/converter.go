package document

import (
	"database/sql"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

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
			String: *in.Data,
			Valid:  true,
		}
	}

	var fetchRequestID sql.NullString
	if in.FetchRequestID != nil {
		fetchRequestID = sql.NullString{
			String: *in.FetchRequestID,
			Valid:  true,
		}
	}

	out := Entity{
		ID:             in.ID,
		AppID:          in.ApplicationID,
		TenantID:       in.Tenant,
		Title:          in.Title,
		DisplayName:    in.DisplayName,
		Description:    in.Description,
		Format:         string(in.Format),
		Kind:           nullKind,
		Data:           nullData,
		FetchRequestID: fetchRequestID,
	}

	return out, nil
}

func (c *converter) FromEntity(in Entity) (model.Document, error) {
	var kindPtr *string
	var dataPtr *string
	if in.Kind.Valid {
		kindPtr = &in.Kind.String
	}
	if in.Data.Valid {
		dataPtr = &in.Data.String
	}

	var fetchRequestID *string
	if in.FetchRequestID.Valid {
		fetchRequestID = &in.FetchRequestID.String
	}

	out := model.Document{
		ID:             in.ID,
		ApplicationID:  in.AppID,
		Tenant:         in.TenantID,
		Title:          in.Title,
		DisplayName:    in.DisplayName,
		Description:    in.Description,
		Format:         model.DocumentFormat(in.Format),
		Kind:           kindPtr,
		Data:           dataPtr,
		FetchRequestID: fetchRequestID,
	}
	return out, nil
}
