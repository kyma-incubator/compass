// GENERATED. DO NOT MODIFY!

package director

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)


func (p *DocumentsResponse) PageInfo() *graphql.PageInfo {
	return &p.Result.Package.Documents.Page
}

func (p *DocumentsResponse) ListAll(ctx context.Context, pager *Pager) (DocumentsOutput, error) {
	pageResult := DocumentsOutput{}

	for pager.HasNext() {
		items := &DocumentsResponse{}
		if err := pager.Next(ctx, items); err != nil {
			return nil, err
		}
		pageResult = append(pageResult, items.Result.Package.Documents.Data...)
	}
	return pageResult, nil
}
