package runtime_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime"
	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime/automock"
	"github.com/kyma-incubator/compass/components/director/internal/graphql"
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestResolver_CreateRuntime(t *testing.T) {
	// given
	modelRuntime := fixModelRuntime("foo", "Foo", "Lorem ipsum")
	gqlRuntime := fixGQLRuntime("foo", "Foo", "Lorem ipsum")
	testErr := errors.New("Test error")

	desc := "Lorem ipsum"
	gqlInput := graphql.RuntimeInput{
		Name:        "Foo",
		Description: &desc,
	}
	modelInput := model.RuntimeInput{
		Name:        "Foo",
		Description: &desc,
	}

	testCases := []struct {
		Name            string
		ServiceFn       func() *automock.RuntimeService
		ConverterFn     func() *automock.RuntimeConverter
		Input           graphql.RuntimeInput
		ExpectedRuntime *graphql.Runtime
		ExpectedErr     error
	}{
		{
			Name: "Success",
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", context.TODO(), "foo").Return(modelRuntime, nil).Once()
				svc.On("Create", context.TODO(), modelInput).Return("foo", nil).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				conv.On("InputFromGraphQL", gqlInput).Return(modelInput).Once()
				conv.On("ToGraphQL", modelRuntime).Return(gqlRuntime).Once()
				return conv
			},
			Input:           gqlInput,
			ExpectedRuntime: gqlRuntime,
			ExpectedErr:     nil,
		},
		{
			Name: "Create Error",
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Create", context.TODO(), modelInput).Return("", testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				conv.On("InputFromGraphQL", gqlInput).Return(modelInput).Once()
				return conv
			},
			Input:           gqlInput,
			ExpectedRuntime: nil,
			ExpectedErr:     testErr,
		},
		{
			Name: "Get Error",
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Create", context.TODO(), modelInput).Return("foo", nil).Once()
				svc.On("Get", context.TODO(), "foo").Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				conv.On("InputFromGraphQL", gqlInput).Return(modelInput).Once()
				return conv
			},
			Input:           gqlInput,
			ExpectedRuntime: nil,
			ExpectedErr:     testErr,
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := runtime.NewResolver(svc)
			resolver.SetConverter(converter)

			// when
			result, err := resolver.CreateRuntime(context.TODO(), testCase.Input)

			// then
			assert.Equal(t, testCase.ExpectedRuntime, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}

func TestResolver_UpdateRuntime(t *testing.T) {
	// given
	modelRuntime := fixModelRuntime("foo", "Foo", "Lorem ipsum")
	gqlRuntime := fixGQLRuntime("foo", "Foo", "Lorem ipsum")
	testErr := errors.New("Test error")

	desc := "Lorem ipsum"
	gqlInput := graphql.RuntimeInput{
		Name:        "Foo",
		Description: &desc,
	}
	modelInput := model.RuntimeInput{
		Name:        "Foo",
		Description: &desc,
	}
	runtimeID := "foo"

	testCases := []struct {
		Name            string
		ServiceFn       func() *automock.RuntimeService
		ConverterFn     func() *automock.RuntimeConverter
		RuntimeID       string
		Input           graphql.RuntimeInput
		ExpectedRuntime *graphql.Runtime
		ExpectedErr     error
	}{
		{
			Name: "Success",
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", context.TODO(), "foo").Return(modelRuntime, nil).Once()
				svc.On("Update", context.TODO(), runtimeID, modelInput).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				conv.On("InputFromGraphQL", gqlInput).Return(modelInput).Once()
				conv.On("ToGraphQL", modelRuntime).Return(gqlRuntime).Once()
				return conv
			},
			RuntimeID:       runtimeID,
			Input:           gqlInput,
			ExpectedRuntime: gqlRuntime,
			ExpectedErr:     nil,
		},
		{
			Name: "Update Error",
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Update", context.TODO(), runtimeID, modelInput).Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				conv.On("InputFromGraphQL", gqlInput).Return(modelInput).Once()
				return conv
			},
			RuntimeID:       runtimeID,
			Input:           gqlInput,
			ExpectedRuntime: nil,
			ExpectedErr:     testErr,
		},
		{
			Name: "Get Error",
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Update", context.TODO(), runtimeID, modelInput).Return(nil).Once()
				svc.On("Get", context.TODO(), "foo").Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				conv.On("InputFromGraphQL", gqlInput).Return(modelInput).Once()
				return conv
			},
			RuntimeID:       runtimeID,
			Input:           gqlInput,
			ExpectedRuntime: nil,
			ExpectedErr:     testErr,
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := runtime.NewResolver(svc)
			resolver.SetConverter(converter)

			// when
			result, err := resolver.UpdateRuntime(context.TODO(), testCase.RuntimeID, testCase.Input)

			// then
			assert.Equal(t, testCase.ExpectedRuntime, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}

func TestResolver_DeleteRuntime(t *testing.T) {
	// given
	modelRuntime := fixModelRuntime("foo", "Foo", "Bar")
	gqlRuntime := fixGQLRuntime("foo", "Foo", "Bar")
	testErr := errors.New("Test error")

	testCases := []struct {
		Name            string
		ServiceFn       func() *automock.RuntimeService
		ConverterFn     func() *automock.RuntimeConverter
		InputID         string
		ExpectedRuntime *graphql.Runtime
		ExpectedErr     error
	}{
		{
			Name: "Success",
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", context.TODO(), "foo").Return(modelRuntime, nil).Once()
				svc.On("Delete", context.TODO(), "foo").Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				conv.On("ToGraphQL", modelRuntime).Return(gqlRuntime).Once()
				return conv
			},
			InputID:         "foo",
			ExpectedRuntime: gqlRuntime,
			ExpectedErr:     nil,
		},
		{
			Name: "Delete Error",
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", context.TODO(), "foo").Return(modelRuntime, nil).Once()
				svc.On("Delete", context.TODO(), "foo").Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				conv.On("ToGraphQL", modelRuntime).Return(gqlRuntime).Once()
				return conv
			},
			InputID:         "foo",
			ExpectedRuntime: nil,
			ExpectedErr:     testErr,
		},
		{
			Name: "Get Error",
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", context.TODO(), "foo").Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				return conv
			},
			InputID:         "foo",
			ExpectedRuntime: nil,
			ExpectedErr:     testErr,
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := runtime.NewResolver(svc)
			resolver.SetConverter(converter)

			// when
			result, err := resolver.DeleteRuntime(context.TODO(), testCase.InputID)

			// then
			assert.Equal(t, testCase.ExpectedRuntime, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}

func TestResolver_Runtime(t *testing.T) {
	// given
	modelRuntime := fixModelRuntime("foo", "Foo", "Bar")
	gqlRuntime := fixGQLRuntime("foo", "Foo", "Bar")
	testErr := errors.New("Test error")

	testCases := []struct {
		Name            string
		ServiceFn       func() *automock.RuntimeService
		ConverterFn     func() *automock.RuntimeConverter
		InputID         string
		ExpectedRuntime *graphql.Runtime
		ExpectedErr     error
	}{
		{
			Name: "Success",
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", context.TODO(), "foo").Return(modelRuntime, nil).Once()

				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				conv.On("ToGraphQL", modelRuntime).Return(gqlRuntime).Once()
				return conv
			},
			InputID:         "foo",
			ExpectedRuntime: gqlRuntime,
			ExpectedErr:     nil,
		},
		{
			Name: "Not Found",
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", context.TODO(), "foo").Return(nil, nil).Once()

				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				var param *model.Runtime
				conv.On("ToGraphQL", param).Return(nil).Once()
				return conv
			},
			InputID:         "foo",
			ExpectedRuntime: nil,
			ExpectedErr:     nil,
		},
		{
			Name: "Error",
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", context.TODO(), "foo").Return(nil, testErr).Once()

				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				return conv
			},
			InputID:         "foo",
			ExpectedRuntime: nil,
			ExpectedErr:     testErr,
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := runtime.NewResolver(svc)
			resolver.SetConverter(converter)

			// when
			result, err := resolver.Runtime(context.TODO(), testCase.InputID)

			// then
			assert.Equal(t, testCase.ExpectedRuntime, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}

func TestResolver_Runtimes(t *testing.T) {
	// given
	modelRuntimes := []*model.Runtime{
		fixModelRuntime("foo", "Foo", "Lorem Ipsum"),
		fixModelRuntime("bar", "Bar", "Lorem Ipsum"),
	}

	gqlRuntimes := []*graphql.Runtime{
		fixGQLRuntime("foo", "Foo", "Lorem Ipsum"),
		fixGQLRuntime("bar", "Bar", "Lorem Ipsum"),
	}

	first := 2
	gqlAfter := graphql.PageCursor("test")
	after := "test"
	filter := []*labelfilter.LabelFilter{
		{Label: "", Values: []string{"foo", "bar"}, Operator: labelfilter.FilterOperatorAll},
	}
	gqlFilter := []*graphql.LabelFilter{
		{Label: "", Values: []string{"foo", "bar"}},
	}
	testErr := errors.New("Test error")

	testCases := []struct {
		Name              string
		ServiceFn         func() *automock.RuntimeService
		ConverterFn       func() *automock.RuntimeConverter
		InputLabelFilters []*graphql.LabelFilter
		InputFirst        *int
		InputAfter        *graphql.PageCursor
		ExpectedResult    *graphql.RuntimePage
		ExpectedErr       error
	}{
		{
			Name: "Success",
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("List", context.TODO(), filter, &first, &after).Return(fixRuntimePage(modelRuntimes), nil).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				conv.On("MultipleToGraphQL", modelRuntimes).Return(gqlRuntimes).Once()
				return conv
			},
			InputFirst:        &first,
			InputAfter:        &gqlAfter,
			InputLabelFilters: gqlFilter,
			ExpectedResult:    fixGQLRuntimePage(gqlRuntimes),
			ExpectedErr:       nil,
		},
		{
			Name: "service Error",
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("List", context.TODO(), filter, &first, &after).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				return conv
			},
			InputFirst:        &first,
			InputAfter:        &gqlAfter,
			InputLabelFilters: gqlFilter,
			ExpectedResult:    nil,
			ExpectedErr:       testErr,
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := runtime.NewResolver(svc)
			resolver.SetConverter(converter)

			// when
			result, err := resolver.Runtimes(context.TODO(), testCase.InputLabelFilters, testCase.InputFirst, testCase.InputAfter)

			// then
			assert.Equal(t, testCase.ExpectedResult, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}

func TestResolver_AddRuntimeLabel(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	runtimeID := "foo"
	gqlLabel := &graphql.Label{
		Key:    "key",
		Values: []string{"foo", "bar"},
	}

	testCases := []struct {
		Name           string
		ServiceFn      func() *automock.RuntimeService
		ConverterFn    func() *automock.RuntimeConverter
		InputRuntimeID string
		InputKey       string
		InputValues    []string
		ExpectedLabel  *graphql.Label
		ExpectedErr    error
	}{
		{
			Name: "Success",
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("AddLabel", context.TODO(), runtimeID, gqlLabel.Key, gqlLabel.Values).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				return conv
			},
			InputRuntimeID: runtimeID,
			InputKey:       gqlLabel.Key,
			InputValues:    gqlLabel.Values,
			ExpectedLabel:  gqlLabel,
			ExpectedErr:    nil,
		},
		{
			Name: "Error",
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("AddLabel", context.TODO(), runtimeID, gqlLabel.Key, gqlLabel.Values).Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				return conv
			},
			InputRuntimeID: runtimeID,
			InputKey:       gqlLabel.Key,
			InputValues:    gqlLabel.Values,
			ExpectedLabel:  nil,
			ExpectedErr:    testErr,
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := runtime.NewResolver(svc)
			resolver.SetConverter(converter)

			// when
			result, err := resolver.AddRuntimeLabel(context.TODO(), testCase.InputRuntimeID, testCase.InputKey, testCase.InputValues)

			// then
			assert.Equal(t, testCase.ExpectedLabel, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}

func TestResolver_DeleteRuntimeLabel(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	runtimeID := "foo"
	rtm := fixModelRuntimeWithLabels(runtimeID, "Foo", map[string][]string{"key": {"foo", "bar"}})

	gqlLabel := &graphql.Label{
		Key:    "key",
		Values: []string{"foo", "bar"},
	}

	testCases := []struct {
		Name           string
		ServiceFn      func() *automock.RuntimeService
		ConverterFn    func() *automock.RuntimeConverter
		InputRuntimeID string
		InputKey       string
		InputValues    []string
		ExpectedLabel  *graphql.Label
		ExpectedErr    error
	}{
		{
			Name: "Success",
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", context.TODO(), runtimeID).Return(rtm, nil).Once()
				svc.On("DeleteLabel", context.TODO(), runtimeID, gqlLabel.Key, gqlLabel.Values).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				return conv
			},
			InputRuntimeID: runtimeID,
			InputKey:       gqlLabel.Key,
			InputValues:    gqlLabel.Values,
			ExpectedLabel:  gqlLabel,
			ExpectedErr:    nil,
		},
		{
			Name: "Get Error",
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", context.TODO(), runtimeID).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				return conv
			},
			InputRuntimeID: runtimeID,
			InputKey:       gqlLabel.Key,
			InputValues:    gqlLabel.Values,
			ExpectedLabel:  nil,
			ExpectedErr:    testErr,
		},
		{
			Name: "Delete Error",
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", context.TODO(), runtimeID).Return(rtm, nil).Once()
				svc.On("DeleteLabel", context.TODO(), runtimeID, gqlLabel.Key, gqlLabel.Values).Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				return conv
			},
			InputRuntimeID: runtimeID,
			InputKey:       gqlLabel.Key,
			InputValues:    gqlLabel.Values,
			ExpectedLabel:  nil,
			ExpectedErr:    testErr,
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := runtime.NewResolver(svc)
			resolver.SetConverter(converter)

			// when
			result, err := resolver.DeleteRuntimeLabel(context.TODO(), testCase.InputRuntimeID, testCase.InputKey, testCase.InputValues)

			// then
			assert.Equal(t, testCase.ExpectedLabel, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}

func TestResolver_AddRuntimeAnnotation(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	runtimeID := "foo"
	gqlAnnotation := &graphql.Annotation{
		Key:   "key",
		Value: "value",
	}

	testCases := []struct {
		Name               string
		ServiceFn          func() *automock.RuntimeService
		ConverterFn        func() *automock.RuntimeConverter
		InputRuntimeID     string
		InputKey           string
		InputValue         interface{}
		ExpectedAnnotation *graphql.Annotation
		ExpectedErr        error
	}{
		{
			Name: "Success",
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("AddAnnotation", context.TODO(), runtimeID, gqlAnnotation.Key, gqlAnnotation.Value).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				return conv
			},
			InputRuntimeID:     runtimeID,
			InputKey:           gqlAnnotation.Key,
			InputValue:         gqlAnnotation.Value,
			ExpectedAnnotation: gqlAnnotation,
			ExpectedErr:        nil,
		},
		{
			Name: "Error",
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("AddAnnotation", context.TODO(), runtimeID, gqlAnnotation.Key, gqlAnnotation.Value).Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				return conv
			},
			InputRuntimeID:     runtimeID,
			InputKey:           gqlAnnotation.Key,
			InputValue:         gqlAnnotation.Value,
			ExpectedAnnotation: nil,
			ExpectedErr:        testErr,
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := runtime.NewResolver(svc)
			resolver.SetConverter(converter)

			// when
			result, err := resolver.AddRuntimeAnnotation(context.TODO(), testCase.InputRuntimeID, testCase.InputKey, testCase.InputValue)

			// then
			assert.Equal(t, testCase.ExpectedAnnotation, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}

func TestResolver_DeleteRuntimeAnnotation(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	runtimeID := "foo"
	rtm := fixModelRuntimeWithAnnotations(runtimeID, "Foo", map[string]interface{}{"key": "value"})

	gqlAnnotation := &graphql.Annotation{
		Key:   "key",
		Value: "value",
	}

	testCases := []struct {
		Name               string
		ServiceFn          func() *automock.RuntimeService
		ConverterFn        func() *automock.RuntimeConverter
		InputRuntimeID     string
		InputKey           string
		ExpectedAnnotation *graphql.Annotation
		ExpectedErr        error
	}{
		{
			Name: "Success",
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", context.TODO(), runtimeID).Return(rtm, nil).Once()
				svc.On("DeleteAnnotation", context.TODO(), runtimeID, gqlAnnotation.Key).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				return conv
			},
			InputRuntimeID:     runtimeID,
			InputKey:           gqlAnnotation.Key,
			ExpectedAnnotation: gqlAnnotation,
			ExpectedErr:        nil,
		},
		{
			Name: "Get Error",
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", context.TODO(), runtimeID).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				return conv
			},
			InputRuntimeID:     runtimeID,
			InputKey:           gqlAnnotation.Key,
			ExpectedAnnotation: nil,
			ExpectedErr:        testErr,
		},
		{
			Name: "Delete Error",
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", context.TODO(), runtimeID).Return(rtm, nil).Once()
				svc.On("DeleteAnnotation", context.TODO(), runtimeID, gqlAnnotation.Key).Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				return conv
			},
			InputRuntimeID:     runtimeID,
			InputKey:           gqlAnnotation.Key,
			ExpectedAnnotation: nil,
			ExpectedErr:        testErr,
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := runtime.NewResolver(svc)
			resolver.SetConverter(converter)

			// when
			result, err := resolver.DeleteRuntimeAnnotation(context.TODO(), testCase.InputRuntimeID, testCase.InputKey)

			// then
			assert.Equal(t, testCase.ExpectedAnnotation, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}
