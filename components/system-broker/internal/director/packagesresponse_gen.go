// GENERATED. DO NOT MODIFY!

package director

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)


func (p *PackagesResponse) PageInfo() *graphql.PageInfo {
	return &p.Result.Packages.Page
}

func (p *PackagesResponse) ListAll(ctx context.Context, pager *Pager) (PackagessOutput, error) {
	pageResult := PackagessOutput{}

	for pager.HasNext() {
		items := &PackagesResponse{}
		if err := pager.Next(ctx, items); err != nil {
			return nil, err
		}
		pageResult = append(pageResult, items.Result.Packages.Data...)
	}
	return pageResult, nil
}
