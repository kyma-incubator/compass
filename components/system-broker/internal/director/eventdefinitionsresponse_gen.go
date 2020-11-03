// GENERATED. DO NOT MODIFY!

package director

import (
	"context"
	
	"github.com/kyma-incubator/compass/components/system-broker/pkg/paginator"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)


func (p *EventDefinitionsResponse) PageInfo() *graphql.PageInfo {
	return &p.Result.Package.EventDefinitions.Page
}

func (p *EventDefinitionsResponse) ListAll(ctx context.Context, paginator *paginator.Paginator) (EventDefinitionsOutput, error) {
	pageResult := EventDefinitionsOutput{}

	for {
		items := &EventDefinitionsResponse{}

		hasNext, err := paginator.Next(ctx, items)
		if err != nil {
			return nil, err
		}

		pageResult = append(pageResult, items.Result.Package.EventDefinitions.Data...)
		if !hasNext {
			return pageResult, nil
		}
	}
}
