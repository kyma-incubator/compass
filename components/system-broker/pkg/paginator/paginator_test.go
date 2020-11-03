package paginator

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/system-broker/pkg/paginator/mocks"
	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

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

		client := &mocks.Client{}
		client.On("Do", mock.Anything, mock.MatchedBy(func(req *gcli.Request) bool {
			vars := req.Vars()
			fmt.Println(">>> 1", vars["after"])
			matched := vars["first"] == pageSize
			matched = matched && vars["after"] == ""
			matched = matched && req.Query() == query
			return matched
		}), mock.Anything).Return(nil).Run(func(args mock.Arguments) {
			fmt.Println(">>> RUN 1")
			arg := args.Get(2).(*PageItem)
			pagedArg := (*arg).(*PaginatorTestStruct)
			pagedArg.Field1 = expected.Field1
			pagedArg.Field2 = expected.Field2
			pagedArg.Page = expected.Page
		}).Times(1)

		client.On("Do", mock.Anything, mock.MatchedBy(func(req *gcli.Request) bool {
			vars := req.Vars()
			fmt.Println(">>> 2", vars["after"])
			matched := vars["first"] == pageSize
			matched = matched && vars["after"] == expected.Page.EndCursor
			matched = matched && req.Query() == query
			return matched
		}), mock.Anything).Return(nil).Run(func(args mock.Arguments) {
			arg := args.Get(2).(*PageItem)
			pagedArg := (*arg).(*PaginatorTestStruct)
			pagedArg.Field1 = expected2.Field1
			pagedArg.Field2 = expected2.Field2
			pagedArg.Page = expected2.Page
		}).Times(1)

		paginator := NewPaginator(query, pageSize, client)

		got := PaginatorTestStruct{}

		hasNext, err := paginator.Next(context.Background(), &got)
		assert.NoError(t, err)
		assert.True(t, hasNext)
		assert.EqualValues(t, expected, got)

		hasNext, err = paginator.Next(context.Background(), &got)
		assert.NoError(t, err)
		assert.False(t, hasNext)
		assert.EqualValues(t, expected2, got)

		mock.AssertExpectationsForObjects(t, client)
	})
}
