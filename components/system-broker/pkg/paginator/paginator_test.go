package paginator_test

import (
	"context"
	"errors"
	"testing"

	"github.com/kyma-incubator/compass/components/system-broker/pkg/paginator"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/paginator/paginatorfakes"
	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/assert"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

type PaginatorTestStruct struct {
	Field1 int
	Field2 []string
	Page   *graphql.PageInfo
}

func (p PaginatorTestStruct) PageInfo() *graphql.PageInfo {
	return p.Page
}

func TestPaginator_Next(t *testing.T) {
	query := "test-query-1"
	pageSize := 1
	errorMsg := "test-error"

	t.Run("for single page", func(t *testing.T) {
		expected := PaginatorTestStruct{
			Field1: 5,
			Field2: []string{"test1", "test2"},
			Page: &graphql.PageInfo{
				StartCursor: "",
				EndCursor:   "",
				HasNextPage: false,
			},
		}

		client := &paginatorfakes.FakeClient{}
		client.DoCalls(func(ctx context.Context, r *gcli.Request, i interface{}) error {
			vars := r.Vars()
			matched := vars["first"] == pageSize
			matched = matched && vars["after"] == ""
			matched = matched && r.Query() == query
			assert.True(t, matched)

			arg := i.(*paginator.PageItem)
			pagedArg := (*arg).(*PaginatorTestStruct)
			pagedArg.Field1 = expected.Field1
			pagedArg.Field2 = expected.Field2
			pagedArg.Page = expected.Page

			return nil
		})

		pager := paginator.NewPaginator(query, pageSize, client)

		got := PaginatorTestStruct{}

		hasNext, err := pager.Next(context.Background(), &got)
		assert.NoError(t, err)
		assert.False(t, hasNext)
		assert.EqualValues(t, expected, got)
		assert.Equal(t, 1, client.DoCallCount())
	})
	t.Run("for multiple pages", func(t *testing.T) {
		expected := PaginatorTestStruct{
			Field1: 5,
			Field2: []string{"test1", "test2"},
			Page: &graphql.PageInfo{
				StartCursor: "",
				EndCursor:   "test-cursor",
				HasNextPage: true,
			},
		}
		expected2 := PaginatorTestStruct{
			Field1: 5,
			Field2: []string{"test-test-1", "test-test-2"},
			Page: &graphql.PageInfo{
				StartCursor: "",
				EndCursor:   "",
				HasNextPage: false,
			},
		}

		client := &paginatorfakes.FakeClient{}
		client.DoCalls(func(c context.Context, r *gcli.Request, i interface{}) error {
			vars := r.Vars()
			matched := vars["first"] == pageSize
			matched = matched && vars["after"] == ""
			matched = matched && r.Query() == query
			assert.True(t, matched)

			arg := i.(*paginator.PageItem)
			pagedArg := (*arg).(*PaginatorTestStruct)
			pagedArg.Field1 = expected.Field1
			pagedArg.Field2 = expected.Field2
			pagedArg.Page = expected.Page

			return nil
		})

		pager := paginator.NewPaginator(query, pageSize, client)

		got := PaginatorTestStruct{}

		hasNext, err := pager.Next(context.Background(), &got)
		assert.NoError(t, err)
		assert.True(t, hasNext)
		assert.EqualValues(t, expected, got)

		client.DoCalls(func(c context.Context, r *gcli.Request, i interface{}) error {
			vars := r.Vars()
			matched := vars["first"] == pageSize
			matched = matched && vars["after"] == string(expected.Page.EndCursor)
			matched = matched && r.Query() == query
			assert.True(t, matched)

			arg := i.(*paginator.PageItem)
			pagedArg := (*arg).(*PaginatorTestStruct)
			pagedArg.Field1 = expected2.Field1
			pagedArg.Field2 = expected2.Field2
			pagedArg.Page = expected2.Page

			return nil
		})

		hasNext, err = pager.Next(context.Background(), &got)
		assert.NoError(t, err)
		assert.False(t, hasNext)
		assert.EqualValues(t, expected2, got)
		assert.Equal(t, 2, client.DoCallCount())
	})
	t.Run("when client returns an error", func(t *testing.T) {
		client := &paginatorfakes.FakeClient{}
		client.DoReturns(errors.New(errorMsg))

		pager := paginator.NewPaginator(query, pageSize, client)
		got := PaginatorTestStruct{}

		hasNext, err := pager.Next(context.Background(), &got)
		assert.Error(t, err)
		assert.False(t, hasNext)
		assert.Contains(t, err.Error(), errorMsg)
		assert.Equal(t, 1, client.DoCallCount())
	})
}
