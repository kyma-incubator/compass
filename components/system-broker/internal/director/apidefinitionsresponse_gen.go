// GENERATED. DO NOT MODIFY!

package director

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)


func (p *ApiDefinitionsResponse) PageInfo() *graphql.PageInfo {
	return &p.Result.Package.ApiDefinitions.Page
}

func (p *ApiDefinitionsResponse) ListAll(ctx context.Context, pager *Pager) (ApiDefinitionsOutput, error) {
	pageResult := ApiDefinitionsOutput{}

	for pager.HasNext() {
		items := &ApiDefinitionsResponse{}
		if err := pager.Next(ctx, items); err != nil {
			return nil, err
		}
		pageResult = append(pageResult, items.Result.Package.ApiDefinitions.Data...)
	}
	return pageResult, nil
}
