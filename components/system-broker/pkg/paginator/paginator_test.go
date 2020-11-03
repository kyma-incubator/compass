package paginator_test

import (
	"context"
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

//func TestPaginator_Next(t *testing.T) {
//	var tests = []struct {
//		Msg              string
//		Expected         PaginatorTestStruct
//		ClientCongurator func(client mocks.Client)
//	}{
//		{
//			Expected: PaginatorTestStruct{
//				Field1: 5,
//				Field2: []string{"test1", "test2"},
//				Page: &graphql.PageInfo{
//					StartCursor: "",
//					EndCursor:   "",
//					HasNextPage: false,
//				},
//			},
//			Msg: "returns correctly output only one page is returned from gql",
//			ClientCongurator: func(client mocks.Client, ) {
//				client.On("Do", mock.Anything, mock.MatchedBy(func(req *gcli.Request) bool {
//					return req.Query() == "test-query"
//				}), mock.Anything).Return(nil).Twice().Run(func(args mock.Arguments) {
//					arg := args.Get(2).(*PageItem)
//					pagedArg := (*arg).(*PaginatorTestStruct)
//					pagedArg.Field1 = expected.Field1
//					pagedArg.Field2 = expected.Field2
//					pagedArg.Page = expected.Page
//				})
//
//			},
//		},
//	}
//
//	client := &mocks.Client{}
//
//	paginator := NewPaginator("test-query", 1, client)
//
//	got := PaginatorTestStruct{}
//
//	paginator.Next(context.Background(), &got)
//	assert.EqualValues(t, expected, got)
//}

func TestPaginator_Next(t *testing.T) {
	//t.Run("for single page", func(t *testing.T) {
	//	expected := PaginatorTestStruct{
	//		Field1: 5,
	//		Field2: []string{"test1", "test2"},
	//		Page: &graphql.PageInfo{
	//			StartCursor: "",
	//			EndCursor:   "",
	//			HasNextPage: false,
	//		},
	//	}
	//	client := &mocks.Client{}
	//	client.On("Do", mock.Anything, mock.MatchedBy(func(req *gcli.Request) bool {
	//		return req.Query() == "test-query"
	//	}), mock.Anything).Return(nil).Run(func(args mock.Arguments) {
	//		arg := args.Get(2).(*PageItem)
	//		pagedArg := (*arg).(*PaginatorTestStruct)
	//		pagedArg.Field1 = expected.Field1
	//		pagedArg.Field2 = expected.Field2
	//		pagedArg.Page = expected.Page
	//	})
	//
	//	paginator := NewPaginator("test-query", 1, client)
	//
	//	got := PaginatorTestStruct{}
	//
	//	paginator.Next(context.Background(), &got)
	//	assert.EqualValues(t, expected, got)
	//
	//	mock.AssertExpectationsForObjects(t, client)
	//})
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

		query := "test-query-1"
		pageSize := 1

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

	})
}
