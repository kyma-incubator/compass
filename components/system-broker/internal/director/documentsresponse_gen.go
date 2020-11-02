// GENERATED. DO NOT MODIFY!

package director

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)


func (p *DocumentsResponse) PageInfo() *graphql.PageInfo {
	return &p.Result.Package.Documents.Page
}

func (p *DocumentsResponse) ListAll(ctx context.Context, pager *Paginator) (DocumentsOutput, error) {
	pageResult := DocumentsOutput{}

	for {
		items := &DocumentsResponse{}

		hasNext, err := pager.Next(ctx, items)
		if err != nil {
			return nil, err
		}

		pageResult = append(pageResult, items.Result.Package.Documents.Data...)
		if !hasNext {
			return pageResult, nil
		}
	}
}
