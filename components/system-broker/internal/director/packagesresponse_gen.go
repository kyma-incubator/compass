// GENERATED. DO NOT MODIFY!

package director

import (
	"context"
	
	"github.com/kyma-incubator/compass/components/system-broker/pkg/paginator"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)


func (p *PackagesResponse) PageInfo() *graphql.PageInfo {
	return &p.Result.Packages.Page
}

func (p *PackagesResponse) ListAll(ctx context.Context, paginator *paginator.Paginator) (PackagessOutput, error) {
	pageResult := PackagessOutput{}

	for {
		items := &PackagesResponse{}

		hasNext, err := paginator.Next(ctx, items)
		if err != nil {
			return nil, err
		}

		pageResult = append(pageResult, items.Result.Packages.Data...)
		if !hasNext {
			return pageResult, nil
		}
	}
}
