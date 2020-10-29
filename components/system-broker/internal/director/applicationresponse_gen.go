// GENERATED. DO NOT MODIFY!

package director

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)


func (p *ApplicationResponse) PageInfo() *graphql.PageInfo {
	return &p.Result.Page
}

func (p *ApplicationResponse) ListAll(ctx context.Context, pager *Pager) (ApplicationsOutput, error) {
	pageResult := ApplicationsOutput{}

	for pager.HasNext() {
		items := &ApplicationResponse{}
		if err := pager.Next(ctx, items); err != nil {
			return nil, err
		}
		pageResult = append(pageResult, items.Result.Data...)
	}
	return pageResult, nil
}
