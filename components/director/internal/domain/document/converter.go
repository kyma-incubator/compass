package document

import (
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/pkg/errors"

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
		BundleID:    in.BundleID,
		Title:       in.Title,
		DisplayName: in.DisplayName,
		Description: in.Description,
		Format:      graphql.DocumentFormat(in.Format),
		Kind:        in.Kind,
		Data:        clob,
		BaseEntity: &graphql.BaseEntity{
			ID:        in.ID,
			Ready:     in.Ready,
			CreatedAt: timePtrToTimestampPtr(in.CreatedAt),
			UpdatedAt: timePtrToTimestampPtr(in.UpdatedAt),
			DeletedAt: timePtrToTimestampPtr(in.DeletedAt),
			Error:     in.Error,
		},
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

func (c *converter) InputFromGraphQL(in *graphql.DocumentInput) (*model.DocumentInput, error) {
	if in == nil {
		return nil, nil
	}

	var data *string
	if in.Data != nil {
		tmp := string(*in.Data)
		data = &tmp
	}

	fetchReq, err := c.frConverter.InputFromGraphQL(in.FetchRequest)
	if err != nil {
		return nil, errors.Wrap(err, "while converting FetchRequestInput input")
	}

	return &model.DocumentInput{
		Title:        in.Title,
		DisplayName:  in.DisplayName,
		Description:  in.Description,
		Format:       model.DocumentFormat(in.Format),
		Kind:         in.Kind,
		Data:         data,
		FetchRequest: fetchReq,
	}, nil
}

func (c *converter) MultipleInputFromGraphQL(in []*graphql.DocumentInput) ([]*model.DocumentInput, error) {
	var inputs []*model.DocumentInput
	for _, r := range in {
		if r == nil {
			continue
		}

		docInput, err := c.InputFromGraphQL(r)
		if err != nil {
			return nil, err
		}

		inputs = append(inputs, docInput)
	}

	return inputs, nil
}

func (c *converter) ToEntity(in model.Document) (*Entity, error) {
	kind := repo.NewNullableString(in.Kind)
	data := repo.NewNullableString(in.Data)

	out := &Entity{
		BndlID:      in.BundleID,
		TenantID:    in.Tenant,
		Title:       in.Title,
		DisplayName: in.DisplayName,
		Description: in.Description,
		Format:      string(in.Format),
		Kind:        kind,
		Data:        data,
		BaseEntity: &repo.BaseEntity{
			ID:        in.ID,
			Ready:     in.Ready,
			CreatedAt: in.CreatedAt,
			UpdatedAt: in.UpdatedAt,
			DeletedAt: in.DeletedAt,
			Error:     repo.NewNullableString(in.Error),
		},
	}

	return out, nil
}

func (c *converter) FromEntity(in Entity) (model.Document, error) {
	kind := repo.StringPtrFromNullableString(in.Kind)
	data := repo.StringPtrFromNullableString(in.Data)

	out := model.Document{
		BundleID:    in.BndlID,
		Tenant:      in.TenantID,
		Title:       in.Title,
		DisplayName: in.DisplayName,
		Description: in.Description,
		Format:      model.DocumentFormat(in.Format),
		Kind:        kind,
		Data:        data,
		BaseEntity: &model.BaseEntity{
			ID:        in.ID,
			Ready:     in.Ready,
			CreatedAt: in.CreatedAt,
			UpdatedAt: in.UpdatedAt,
			DeletedAt: in.DeletedAt,
			Error:     repo.StringPtrFromNullableString(in.Error),
		},
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
