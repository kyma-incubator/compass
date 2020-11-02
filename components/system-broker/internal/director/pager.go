package director

import (
	"context"
	"fmt"
	"reflect"

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

type Pager struct {
	QueryGenerator func(pageSize int, page string) string
	PageSize       int
	PageToken      string
	Client         Client
	hasNext        bool
}

func NewPager(queryGenerator func(pageSize int, page string) string, pageSize int, client Client) *Pager {
	return &Pager{
		QueryGenerator: queryGenerator,
		PageSize:       pageSize,
		Client:         client,
		hasNext:        true,
	}
}

func (p *Pager) Next(ctx context.Context, output PageItem) error {
	if !p.hasNext {
		return errors.New("no more pages")
	}

	query := p.QueryGenerator(p.PageSize, p.PageToken)
	req := gcli.NewRequest(query)
	err := p.Client.Do(ctx, req, &output)
	if err != nil {
		return errors.Wrap(err, "while getting page")
	}

	pageInfo := output.PageInfo()
	if !pageInfo.HasNextPage {
		p.hasNext = false
		return nil
	}

	p.PageToken = string(pageInfo.EndCursor)

	return nil
}

func (p *Pager) HasNext() bool {
	return p.hasNext
}

func (p *Pager) ListAll(ctx context.Context, output interface{}) error {
	itemsType := reflect.TypeOf(output)
	if itemsType.Kind() != reflect.Ptr || itemsType.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("items should be a pointer to a slice, but got %v", itemsType)
	}

	allItems := reflect.MakeSlice(itemsType.Elem(), 0, 0)

	for p.HasNext() {
		pageSlice := reflect.New(itemsType.Elem())
		err := p.Next(ctx, pageSlice.Interface().(PageItem))
		if err != nil {
			return err
		}

		allItems = reflect.AppendSlice(allItems, pageSlice.Elem())
	}

	reflect.ValueOf(output).Elem().Set(allItems)
	return nil
}

type PageItem interface {
	PageInfo() *graphql.PageInfo
}
