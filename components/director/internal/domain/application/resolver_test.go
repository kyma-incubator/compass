package application_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/application"
	"github.com/kyma-incubator/compass/components/director/internal/domain/application/automock"
	"github.com/kyma-incubator/compass/components/director/internal/graphql"
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestResolver_CreateApplication(t *testing.T) {
	// given
	modelApplication := fixModelApplication("foo", "Foo", "Lorem ipsum")
	gqlApplication := fixGQLApplication("foo", "Foo", "Lorem ipsum")
	testErr := errors.New("Test error")

	desc := "Lorem ipsum"
	gqlInput := graphql.ApplicationInput{
		Name:        "Foo",
		Description: &desc,
	}
	modelInput := model.ApplicationInput{
		Name:        "Foo",
		Description: &desc,
	}

	testCases := []struct {
		Name                string
		ServiceFn           func() *automock.ApplicationService
		ConverterFn         func() *automock.ApplicationConverter
		Input               graphql.ApplicationInput
		ExpectedApplication *graphql.Application
		ExpectedErr         error
	}{
		{
			Name: "Success",
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Get", context.TODO(), "foo").Return(modelApplication, nil).Once()
				svc.On("Create", context.TODO(), modelInput).Return("foo", nil).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				conv.On("InputFromGraphQL", gqlInput).Return(modelInput).Once()
				conv.On("ToGraphQL", modelApplication).Return(gqlApplication).Once()
				return conv
			},
			Input:               gqlInput,
			ExpectedApplication: gqlApplication,
			ExpectedErr:         nil,
		},
		{
			Name: "Returns error when application creation failed",
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Create", context.TODO(), modelInput).Return("", testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				conv.On("InputFromGraphQL", gqlInput).Return(modelInput).Once()
				return conv
			},
			Input:               gqlInput,
			ExpectedApplication: nil,
			ExpectedErr:         testErr,
		},
		{
			Name: "Returns error when application creation failed",
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Create", context.TODO(), modelInput).Return("foo", nil).Once()
				svc.On("Get", context.TODO(), "foo").Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				conv.On("InputFromGraphQL", gqlInput).Return(modelInput).Once()
				return conv
			},
			Input:               gqlInput,
			ExpectedApplication: nil,
			ExpectedErr:         testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := application.NewResolver(svc, nil, nil, nil, nil, nil, nil, nil, nil)
			resolver.SetConverter(converter)

			// when
			result, err := resolver.CreateApplication(context.TODO(), testCase.Input)

			// then
			assert.Equal(t, testCase.ExpectedApplication, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}

func TestResolver_UpdateApplication(t *testing.T) {
	// given
	modelApplication := fixModelApplication("foo", "Foo", "Lorem ipsum")
	gqlApplication := fixGQLApplication("foo", "Foo", "Lorem ipsum")
	testErr := errors.New("Test error")

	desc := "Lorem ipsum"
	gqlInput := graphql.ApplicationInput{
		Name:        "Foo",
		Description: &desc,
	}
	modelInput := model.ApplicationInput{
		Name:        "Foo",
		Description: &desc,
	}
	applicationID := "foo"

	testCases := []struct {
		Name                string
		ServiceFn           func() *automock.ApplicationService
		ConverterFn         func() *automock.ApplicationConverter
		ApplicationID       string
		Input               graphql.ApplicationInput
		ExpectedApplication *graphql.Application
		ExpectedErr         error
	}{
		{
			Name: "Success",
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Get", context.TODO(), "foo").Return(modelApplication, nil).Once()
				svc.On("Update", context.TODO(), applicationID, modelInput).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				conv.On("InputFromGraphQL", gqlInput).Return(modelInput).Once()
				conv.On("ToGraphQL", modelApplication).Return(gqlApplication).Once()
				return conv
			},
			ApplicationID:       applicationID,
			Input:               gqlInput,
			ExpectedApplication: gqlApplication,
			ExpectedErr:         nil,
		},
		{
			Name: "Returns error when application update failed",
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Update", context.TODO(), applicationID, modelInput).Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				conv.On("InputFromGraphQL", gqlInput).Return(modelInput).Once()
				return conv
			},
			ApplicationID:       applicationID,
			Input:               gqlInput,
			ExpectedApplication: nil,
			ExpectedErr:         testErr,
		},
		{
			Name: "Returns error when application retrieval failed",
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Update", context.TODO(), applicationID, modelInput).Return(nil).Once()
				svc.On("Get", context.TODO(), "foo").Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				conv.On("InputFromGraphQL", gqlInput).Return(modelInput).Once()
				return conv
			},
			ApplicationID:       applicationID,
			Input:               gqlInput,
			ExpectedApplication: nil,
			ExpectedErr:         testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := application.NewResolver(svc, nil, nil, nil, nil, nil, nil, nil, nil)
			resolver.SetConverter(converter)

			// when
			result, err := resolver.UpdateApplication(context.TODO(), testCase.ApplicationID, testCase.Input)

			// then
			assert.Equal(t, testCase.ExpectedApplication, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}

func TestResolver_DeleteApplication(t *testing.T) {
	// given
	modelApplication := fixModelApplication("foo", "Foo", "Bar")
	gqlApplication := fixGQLApplication("foo", "Foo", "Bar")
	testErr := errors.New("Test error")

	testCases := []struct {
		Name                string
		ServiceFn           func() *automock.ApplicationService
		ConverterFn         func() *automock.ApplicationConverter
		InputID             string
		ExpectedApplication *graphql.Application
		ExpectedErr         error
	}{
		{
			Name: "Success",
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Get", context.TODO(), "foo").Return(modelApplication, nil).Once()
				svc.On("Delete", context.TODO(), "foo").Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				conv.On("ToGraphQL", modelApplication).Return(gqlApplication).Once()
				return conv
			},
			InputID:             "foo",
			ExpectedApplication: gqlApplication,
			ExpectedErr:         nil,
		},
		{
			Name: "Returns error when application deletion failed",
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Get", context.TODO(), "foo").Return(modelApplication, nil).Once()
				svc.On("Delete", context.TODO(), "foo").Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				conv.On("ToGraphQL", modelApplication).Return(gqlApplication).Once()
				return conv
			},
			InputID:             "foo",
			ExpectedApplication: nil,
			ExpectedErr:         testErr,
		},
		{
			Name: "Returns error when application retrieval failed",
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Get", context.TODO(), "foo").Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				return conv
			},
			InputID:             "foo",
			ExpectedApplication: nil,
			ExpectedErr:         testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := application.NewResolver(svc, nil, nil, nil, nil, nil, nil, nil, nil)
			resolver.SetConverter(converter)

			// when
			result, err := resolver.DeleteApplication(context.TODO(), testCase.InputID)

			// then
			assert.Equal(t, testCase.ExpectedApplication, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}

func TestResolver_Application(t *testing.T) {
	// given
	modelApplication := fixModelApplication("foo", "Foo", "Bar")
	gqlApplication := fixGQLApplication("foo", "Foo", "Bar")
	testErr := errors.New("Test error")

	testCases := []struct {
		Name                string
		ServiceFn           func() *automock.ApplicationService
		ConverterFn         func() *automock.ApplicationConverter
		InputID             string
		ExpectedApplication *graphql.Application
		ExpectedErr         error
	}{
		{
			Name: "Success",
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Get", context.TODO(), "foo").Return(modelApplication, nil).Once()

				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				conv.On("ToGraphQL", modelApplication).Return(gqlApplication).Once()
				return conv
			},
			InputID:             "foo",
			ExpectedApplication: gqlApplication,
			ExpectedErr:         nil,
		},
		{
			Name: "Returns error when application retrieval failed",
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Get", context.TODO(), "foo").Return(nil, testErr).Once()

				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				return conv
			},
			InputID:             "foo",
			ExpectedApplication: nil,
			ExpectedErr:         testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := application.NewResolver(svc, nil, nil, nil, nil, nil, nil, nil, nil)
			resolver.SetConverter(converter)

			// when
			result, err := resolver.Application(context.TODO(), testCase.InputID)

			// then
			assert.Equal(t, testCase.ExpectedApplication, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}

func TestResolver_Applications(t *testing.T) {
	// given
	modelApplications := []*model.Application{
		fixModelApplication("foo", "Foo", "Lorem Ipsum"),
		fixModelApplication("bar", "Bar", "Lorem Ipsum"),
	}

	gqlApplications := []*graphql.Application{
		fixGQLApplication("foo", "Foo", "Lorem Ipsum"),
		fixGQLApplication("bar", "Bar", "Lorem Ipsum"),
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
		ServiceFn         func() *automock.ApplicationService
		ConverterFn       func() *automock.ApplicationConverter
		InputLabelFilters []*graphql.LabelFilter
		InputFirst        *int
		InputAfter        *graphql.PageCursor
		ExpectedResult    *graphql.ApplicationPage
		ExpectedErr       error
	}{
		{
			Name: "Success",
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("List", context.TODO(), filter, &first, &after).Return(fixApplicationPage(modelApplications), nil).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				conv.On("MultipleToGraphQL", modelApplications).Return(gqlApplications).Once()
				return conv
			},
			InputFirst:        &first,
			InputAfter:        &gqlAfter,
			InputLabelFilters: gqlFilter,
			ExpectedResult:    fixGQLApplicationPage(gqlApplications),
			ExpectedErr:       nil,
		},
		{
			Name: "Returns error when application listing failed",
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("List", context.TODO(), filter, &first, &after).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				return conv
			},
			InputFirst:        &first,
			InputAfter:        &gqlAfter,
			InputLabelFilters: gqlFilter,
			ExpectedResult:    nil,
			ExpectedErr:       testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := application.NewResolver(svc, nil, nil, nil, nil, nil, nil, nil, nil)
			resolver.SetConverter(converter)

			// when
			result, err := resolver.Applications(context.TODO(), testCase.InputLabelFilters, testCase.InputFirst, testCase.InputAfter)

			// then
			assert.Equal(t, testCase.ExpectedResult, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}

func TestResolver_AddApplicationLabel(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	applicationID := "foo"
	gqlLabel := &graphql.Label{
		Key:    "key",
		Values: []string{"foo", "bar"},
	}

	testCases := []struct {
		Name               string
		ServiceFn          func() *automock.ApplicationService
		ConverterFn        func() *automock.ApplicationConverter
		InputApplicationID string
		InputKey           string
		InputValues        []string
		ExpectedLabel      *graphql.Label
		ExpectedErr        error
	}{
		{
			Name: "Success",
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("AddLabel", context.TODO(), applicationID, gqlLabel.Key, gqlLabel.Values).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				return conv
			},
			InputApplicationID: applicationID,
			InputKey:           gqlLabel.Key,
			InputValues:        gqlLabel.Values,
			ExpectedLabel:      gqlLabel,
			ExpectedErr:        nil,
		},
		{
			Name: "Returns error when adding label to application failed",
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("AddLabel", context.TODO(), applicationID, gqlLabel.Key, gqlLabel.Values).Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				return conv
			},
			InputApplicationID: applicationID,
			InputKey:           gqlLabel.Key,
			InputValues:        gqlLabel.Values,
			ExpectedLabel:      nil,
			ExpectedErr:        testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := application.NewResolver(svc, nil, nil, nil, nil, nil, nil, nil, nil)
			resolver.SetConverter(converter)

			// when
			result, err := resolver.AddApplicationLabel(context.TODO(), testCase.InputApplicationID, testCase.InputKey, testCase.InputValues)

			// then
			assert.Equal(t, testCase.ExpectedLabel, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}

func TestResolver_DeleteApplicationLabel(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	applicationID := "foo"
	app := fixModelApplicationWithLabels(applicationID, "Foo", map[string][]string{"key": {"foo", "bar"}})

	gqlLabel := &graphql.Label{
		Key:    "key",
		Values: []string{"foo", "bar"},
	}

	testCases := []struct {
		Name               string
		ServiceFn          func() *automock.ApplicationService
		ConverterFn        func() *automock.ApplicationConverter
		InputApplicationID string
		InputKey           string
		InputValues        []string
		ExpectedLabel      *graphql.Label
		ExpectedErr        error
	}{
		{
			Name: "Success",
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Get", context.TODO(), applicationID).Return(app, nil).Once()
				svc.On("DeleteLabel", context.TODO(), applicationID, gqlLabel.Key, gqlLabel.Values).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				return conv
			},
			InputApplicationID: applicationID,
			InputKey:           gqlLabel.Key,
			InputValues:        gqlLabel.Values,
			ExpectedLabel:      gqlLabel,
			ExpectedErr:        nil,
		},
		{
			Name: "Returns error when application retrieval failed",
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("DeleteLabel", context.TODO(), applicationID, gqlLabel.Key, gqlLabel.Values).Return(nil).Once()
				svc.On("Get", context.TODO(), applicationID).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				return conv
			},
			InputApplicationID: applicationID,
			InputKey:           gqlLabel.Key,
			InputValues:        gqlLabel.Values,
			ExpectedLabel:      nil,
			ExpectedErr:        testErr,
		},
		{
			Name: "Returns error when deleting application's label failed",
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("DeleteLabel", context.TODO(), applicationID, gqlLabel.Key, gqlLabel.Values).Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				return conv
			},
			InputApplicationID: applicationID,
			InputKey:           gqlLabel.Key,
			InputValues:        gqlLabel.Values,
			ExpectedLabel:      nil,
			ExpectedErr:        testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := application.NewResolver(svc, nil, nil, nil, nil, nil, nil, nil, nil)
			resolver.SetConverter(converter)

			// when
			result, err := resolver.DeleteApplicationLabel(context.TODO(), testCase.InputApplicationID, testCase.InputKey, testCase.InputValues)

			// then
			assert.Equal(t, testCase.ExpectedLabel, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}

func TestResolver_AddApplicationAnnotation(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	applicationID := "foo"
	gqlAnnotation := &graphql.Annotation{
		Key:   "key",
		Value: "value",
	}

	testCases := []struct {
		Name               string
		ServiceFn          func() *automock.ApplicationService
		ConverterFn        func() *automock.ApplicationConverter
		InputApplicationID string
		InputKey           string
		InputValue         interface{}
		ExpectedAnnotation *graphql.Annotation
		ExpectedErr        error
	}{
		{
			Name: "Success",
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("AddAnnotation", context.TODO(), applicationID, gqlAnnotation.Key, gqlAnnotation.Value).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				return conv
			},
			InputApplicationID: applicationID,
			InputKey:           gqlAnnotation.Key,
			InputValue:         gqlAnnotation.Value,
			ExpectedAnnotation: gqlAnnotation,
			ExpectedErr:        nil,
		},
		{
			Name: "Returns error when adding annotation to application failed",
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("AddAnnotation", context.TODO(), applicationID, gqlAnnotation.Key, gqlAnnotation.Value).Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				return conv
			},
			InputApplicationID: applicationID,
			InputKey:           gqlAnnotation.Key,
			InputValue:         gqlAnnotation.Value,
			ExpectedAnnotation: nil,
			ExpectedErr:        testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := application.NewResolver(svc, nil, nil, nil, nil, nil, nil, nil, nil)
			resolver.SetConverter(converter)

			// when
			result, err := resolver.AddApplicationAnnotation(context.TODO(), testCase.InputApplicationID, testCase.InputKey, testCase.InputValue)

			// then
			assert.Equal(t, testCase.ExpectedAnnotation, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}

func TestResolver_DeleteApplicationAnnotation(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	applicationID := "foo"
	app := fixModelApplicationWithAnnotations(applicationID, "Foo", map[string]interface{}{"key": "value"})

	gqlAnnotation := &graphql.Annotation{
		Key:   "key",
		Value: "value",
	}

	testCases := []struct {
		Name               string
		ServiceFn          func() *automock.ApplicationService
		ConverterFn        func() *automock.ApplicationConverter
		InputApplicationID string
		InputKey           string
		ExpectedAnnotation *graphql.Annotation
		ExpectedErr        error
	}{
		{
			Name: "Success",
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Get", context.TODO(), applicationID).Return(app, nil).Once()
				svc.On("DeleteAnnotation", context.TODO(), applicationID, gqlAnnotation.Key).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				return conv
			},
			InputApplicationID: applicationID,
			InputKey:           gqlAnnotation.Key,
			ExpectedAnnotation: gqlAnnotation,
			ExpectedErr:        nil,
		},
		{
			Name: "Returns error when application retrieval failed",
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Get", context.TODO(), applicationID).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				return conv
			},
			InputApplicationID: applicationID,
			InputKey:           gqlAnnotation.Key,
			ExpectedAnnotation: nil,
			ExpectedErr:        testErr,
		},
		{
			Name: "Returns error when deleting application's annotation failed",
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Get", context.TODO(), applicationID).Return(app, nil).Once()
				svc.On("DeleteAnnotation", context.TODO(), applicationID, gqlAnnotation.Key).Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				return conv
			},
			InputApplicationID: applicationID,
			InputKey:           gqlAnnotation.Key,
			ExpectedAnnotation: nil,
			ExpectedErr:        testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := application.NewResolver(svc, nil, nil, nil, nil, nil, nil, nil, nil)
			resolver.SetConverter(converter)

			// when
			result, err := resolver.DeleteApplicationAnnotation(context.TODO(), testCase.InputApplicationID, testCase.InputKey)

			// then
			assert.Equal(t, testCase.ExpectedAnnotation, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}

func TestResolver_Documents(t *testing.T) {
	// given
	applicationID := "fooid"
	modelDocuments := []*model.Document{
		fixModelDocument(applicationID, "foo"),
		fixModelDocument(applicationID, "bar"),
	}
	gqlDocuments := []*graphql.Document{
		fixGQLDocument("foo"),
		fixGQLDocument("bar"),
	}
	app := fixGQLApplication(applicationID, "foo", "bar")

	first := 2
	gqlAfter := graphql.PageCursor("test")
	after := "test"
	testErr := errors.New("Test error")

	testCases := []struct {
		Name           string
		ServiceFn      func() *automock.DocumentService
		ConverterFn    func() *automock.DocumentConverter
		InputFirst     *int
		InputAfter     *graphql.PageCursor
		ExpectedResult *graphql.DocumentPage
		ExpectedErr    error
	}{
		{
			Name: "Success",
			ServiceFn: func() *automock.DocumentService {
				svc := &automock.DocumentService{}
				svc.On("List", context.TODO(), applicationID, &first, &after).Return(fixModelDocumentPage(modelDocuments), nil).Once()
				return svc
			},
			ConverterFn: func() *automock.DocumentConverter {
				conv := &automock.DocumentConverter{}
				conv.On("MultipleToGraphQL", modelDocuments).Return(gqlDocuments).Once()
				return conv
			},
			InputFirst:     &first,
			InputAfter:     &gqlAfter,
			ExpectedResult: fixGQLDocumentPage(gqlDocuments),
			ExpectedErr:    nil,
		},
		{
			Name: "Returns error when document listing failed",
			ServiceFn: func() *automock.DocumentService {
				svc := &automock.DocumentService{}
				svc.On("List", context.TODO(), applicationID, &first, &after).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.DocumentConverter {
				conv := &automock.DocumentConverter{}
				return conv
			},
			InputFirst:     &first,
			InputAfter:     &gqlAfter,
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := application.NewResolver(nil, nil, nil, svc, nil, nil, converter, nil, nil)

			// when
			result, err := resolver.Documents(context.TODO(), app, testCase.InputFirst, testCase.InputAfter)

			// then
			assert.Equal(t, testCase.ExpectedResult, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}

func TestResolver_Webhooks(t *testing.T) {
	// given
	applicationID := "fooid"
	modelWebhooks := []*model.ApplicationWebhook{
		fixModelWebhook(applicationID, "foo"),
		fixModelWebhook(applicationID, "bar"),
	}
	gqlWebhooks := []*graphql.ApplicationWebhook{
		fixGQLWebhook("foo"),
		fixGQLWebhook("bar"),
	}
	app := fixGQLApplication(applicationID, "foo", "bar")
	testErr := errors.New("Test error")

	testCases := []struct {
		Name           string
		ServiceFn      func() *automock.WebhookService
		ConverterFn    func() *automock.WebhookConverter
		ExpectedResult []*graphql.ApplicationWebhook
		ExpectedErr    error
	}{
		{
			Name: "Success",
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("List", context.TODO(), applicationID).Return(modelWebhooks, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("MultipleToGraphQL", modelWebhooks).Return(gqlWebhooks).Once()
				return conv
			},
			ExpectedResult: gqlWebhooks,
			ExpectedErr:    nil,
		},
		{
			Name: "Returns error when webhook listing failed",
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("List", context.TODO(), applicationID).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				return conv
			},
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := application.NewResolver(nil, nil, nil, nil, svc, nil, nil, converter, nil)

			// when
			result, err := resolver.Webhooks(context.TODO(), app)

			// then
			assert.Equal(t, testCase.ExpectedResult, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}

func TestResolver_AddApplicationWebhook(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	applicationID := "foo"
	id := "bar"
	gqlWebhookInput := fixGQLWebhookInput()
	modelWebhookInput := fixModelWebhookInput()
	gqlWebhook := fixGQLWebhook(id)
	modelWebhook := fixModelWebhook(applicationID, id)

	testCases := []struct {
		Name               string
		ServiceFn          func() *automock.WebhookService
		ConverterFn        func() *automock.WebhookConverter
		InputApplicationID string
		InputWebhook       graphql.ApplicationWebhookInput
		ExpectedWebhook    *graphql.ApplicationWebhook
		ExpectedErr        error
	}{
		{
			Name: "Success",
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("Create", context.TODO(), applicationID, *modelWebhookInput).Return(id, nil).Once()
				svc.On("Get", context.TODO(), id).Return(modelWebhook, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("InputFromGraphQL", gqlWebhookInput).Return(modelWebhookInput).Once()
				conv.On("ToGraphQL", modelWebhook).Return(gqlWebhook).Once()
				return conv
			},
			InputApplicationID: applicationID,
			InputWebhook:       *gqlWebhookInput,
			ExpectedWebhook:    gqlWebhook,
			ExpectedErr:        nil,
		},
		{
			Name: "Returns error when webhook creation failed",
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("Create", context.TODO(), applicationID, *modelWebhookInput).Return("", testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("InputFromGraphQL", gqlWebhookInput).Return(modelWebhookInput).Once()
				return conv
			},
			InputApplicationID: applicationID,
			InputWebhook:       *gqlWebhookInput,
			ExpectedWebhook:    nil,
			ExpectedErr:        testErr,
		},
		{
			Name: "Returns error when webhook retrieval failed",
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("Create", context.TODO(), applicationID, *modelWebhookInput).Return(id, nil).Once()
				svc.On("Get", context.TODO(), id).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("InputFromGraphQL", gqlWebhookInput).Return(modelWebhookInput).Once()
				return conv
			},
			InputApplicationID: applicationID,
			InputWebhook:       *gqlWebhookInput,
			ExpectedWebhook:    nil,
			ExpectedErr:        testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := application.NewResolver(nil, nil, nil, nil, svc, nil, nil, converter, nil)

			// when
			result, err := resolver.AddApplicationWebhook(context.TODO(), testCase.InputApplicationID, testCase.InputWebhook)

			// then
			assert.Equal(t, testCase.ExpectedWebhook, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}

func TestResolver_UpdateApplicationWebhook(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	applicationID := "foo"
	id := "bar"
	gqlWebhookInput := fixGQLWebhookInput()
	modelWebhookInput := fixModelWebhookInput()
	gqlWebhook := fixGQLWebhook(id)
	modelWebhook := fixModelWebhook(applicationID, id)

	testCases := []struct {
		Name            string
		ServiceFn       func() *automock.WebhookService
		ConverterFn     func() *automock.WebhookConverter
		InputWebhookID  string
		InputWebhook    graphql.ApplicationWebhookInput
		ExpectedWebhook *graphql.ApplicationWebhook
		ExpectedErr     error
	}{
		{
			Name: "Success",
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("Update", context.TODO(), id, *modelWebhookInput).Return(nil).Once()
				svc.On("Get", context.TODO(), id).Return(modelWebhook, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("InputFromGraphQL", gqlWebhookInput).Return(modelWebhookInput).Once()
				conv.On("ToGraphQL", modelWebhook).Return(gqlWebhook).Once()
				return conv
			},
			InputWebhookID:  id,
			InputWebhook:    *gqlWebhookInput,
			ExpectedWebhook: gqlWebhook,
			ExpectedErr:     nil,
		},
		{
			Name: "Returns error when webhook update failed",
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("Update", context.TODO(), id, *modelWebhookInput).Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("InputFromGraphQL", gqlWebhookInput).Return(modelWebhookInput).Once()
				return conv
			},
			InputWebhookID:  id,
			InputWebhook:    *gqlWebhookInput,
			ExpectedWebhook: nil,
			ExpectedErr:     testErr,
		},
		{
			Name: "Returns error when webhook retrieval failed",
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("Update", context.TODO(), id, *modelWebhookInput).Return(nil).Once()
				svc.On("Get", context.TODO(), id).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("InputFromGraphQL", gqlWebhookInput).Return(modelWebhookInput).Once()
				return conv
			},
			InputWebhookID:  id,
			InputWebhook:    *gqlWebhookInput,
			ExpectedWebhook: nil,
			ExpectedErr:     testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := application.NewResolver(nil, nil, nil, nil, svc, nil, nil, converter, nil)

			// when
			result, err := resolver.UpdateApplicationWebhook(context.TODO(), testCase.InputWebhookID, testCase.InputWebhook)

			// then
			assert.Equal(t, testCase.ExpectedWebhook, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}

func TestResolver_DeleteApplicationWebhook(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	applicationID := "foo"
	id := "bar"
	gqlWebhookInput := fixGQLWebhookInput()
	gqlWebhook := fixGQLWebhook(id)
	modelWebhook := fixModelWebhook(applicationID, id)

	testCases := []struct {
		Name            string
		ServiceFn       func() *automock.WebhookService
		ConverterFn     func() *automock.WebhookConverter
		InputWebhookID  string
		InputWebhook    graphql.ApplicationWebhookInput
		ExpectedWebhook *graphql.ApplicationWebhook
		ExpectedErr     error
	}{
		{
			Name: "Success",
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("Get", context.TODO(), id).Return(modelWebhook, nil).Once()
				svc.On("Delete", context.TODO(), id).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", modelWebhook).Return(gqlWebhook).Once()
				return conv
			},
			InputWebhookID:  id,
			InputWebhook:    *gqlWebhookInput,
			ExpectedWebhook: gqlWebhook,
			ExpectedErr:     nil,
		},
		{
			Name: "Returns error when webhook retrieval failed",
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("Get", context.TODO(), id).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				return conv
			},
			InputWebhookID:  id,
			InputWebhook:    *gqlWebhookInput,
			ExpectedWebhook: nil,
			ExpectedErr:     testErr,
		},
		{
			Name: "Returns error when webhook deletion failed",
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("Get", context.TODO(), id).Return(modelWebhook, nil).Once()
				svc.On("Delete", context.TODO(), id).Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", modelWebhook).Return(gqlWebhook).Once()
				return conv
			},
			InputWebhookID:  id,
			InputWebhook:    *gqlWebhookInput,
			ExpectedWebhook: nil,
			ExpectedErr:     testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := application.NewResolver(nil, nil, nil, nil, svc, nil, nil, converter, nil)

			// when
			result, err := resolver.DeleteApplicationWebhook(context.TODO(), testCase.InputWebhookID)

			// then
			assert.Equal(t, testCase.ExpectedWebhook, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}

// TODO: Test Resolvers for:
// 	- Apis
// 	- EventAPIs
