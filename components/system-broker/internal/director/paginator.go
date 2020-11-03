package director

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

type GenericPage struct {
	Data       interface{}       `json:"data"`
	PageInfo   *graphql.PageInfo `json:"pageInfo"`
	TotalCount int               `json:"totalCount"`
}

type GenericOutput struct {
	Result *GenericPage `json:"result"`
}

type Paginator struct {
	query     string
	pageSize  int
	pageToken string
	client    Client
}

func NewPaginator(query string, pageSize int, client Client) *Paginator {
	return &Paginator{
		query:    query,
		pageSize: pageSize,
		client:   client,
	}
}

func (p *Paginator) Next(ctx context.Context, output PageItem) (bool, error) {
	req := gcli.NewRequest(p.query)
	req.Var("first", p.pageSize)
	req.Var("after", p.pageToken)

	err := p.client.Do(ctx, req, &output)
	if err != nil {
		return false, errors.Wrap(err, "while getting page")
	}

	pageInfo := output.PageInfo()
	if !pageInfo.HasNextPage {
		return false, nil
	}

	p.pageToken = string(pageInfo.EndCursor)

	return true, nil
}

type PageItem interface {
	PageInfo() *graphql.PageInfo
}
