// GENERATED. DO NOT MODIFY!

package director

import (
	"context"
	
	"github.com/kyma-incubator/compass/components/system-broker/pkg/paginator"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)


func (p *ApplicationResponse) PageInfo() *graphql.PageInfo {
	return &p.Result.Page
}

func (p *ApplicationResponse) ListAll(ctx context.Context, paginator *paginator.Paginator) (ApplicationsOutput, error) {
	pageResult := ApplicationsOutput{}

	for {
		items := &ApplicationResponse{}

		hasNext, err := paginator.Next(ctx, items)
		if err != nil {
			return nil, err
		}

		pageResult = append(pageResult, items.Result.Data...)
		if !hasNext {
			return pageResult, nil
		}
	}
}
