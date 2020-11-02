// GENERATED. DO NOT MODIFY!

package director

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)


func (p *ApiDefinitionsResponse) PageInfo() *graphql.PageInfo {
	return &p.Result.Package.ApiDefinitions.Page
}

func (p *ApiDefinitionsResponse) ListAll(ctx context.Context, pager *Paginator) (ApiDefinitionsOutput, error) {
	pageResult := ApiDefinitionsOutput{}

	for {
		items := &ApiDefinitionsResponse{}

		hasNext, err := pager.Next(ctx, items)
		if err != nil {
			return nil, err
		}

		pageResult = append(pageResult, items.Result.Package.ApiDefinitions.Data...)
		if !hasNext {
			return pageResult, nil
		}
	}
}
