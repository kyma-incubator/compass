// GENERATED. DO NOT MODIFY!

package director

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)


func (p *EventDefinitionsResponse) PageInfo() *graphql.PageInfo {
	return &p.Result.Package.EventDefinitions.Page
}

func (p *EventDefinitionsResponse) ListAll(ctx context.Context, pager *Pager) (EventDefinitionsOutput, error) {
	pageResult := EventDefinitionsOutput{}

	for pager.HasNext() {
		items := &EventDefinitionsResponse{}
		if err := pager.Next(ctx, items); err != nil {
			return nil, err
		}
		pageResult = append(pageResult, items.Result.Package.EventDefinitions.Data...)
	}
	return pageResult, nil
}
