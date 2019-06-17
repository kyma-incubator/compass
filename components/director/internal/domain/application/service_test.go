package application_test

import (
	"context"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/internal/domain/application"
	"github.com/kyma-incubator/compass/components/director/internal/domain/application/automock"
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/tenant"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestService_Create(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	desc := "Lorem ipsum"
	modelInput := model.ApplicationInput{
		Name:        "Foo",
		Description: &desc,
	}

	applicationModel := mock.MatchedBy(func(app *model.Application) bool {
		return app.Name == modelInput.Name && app.Description == modelInput.Description
	})

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, "tenant")

	testCases := []struct {
		Name         string
		RepositoryFn func() *automock.ApplicationRepository
		Input        model.ApplicationInput
		ExpectedErr  error
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("Create", applicationModel).Return(nil).Once()
				return repo
			},
			Input:       modelInput,
			ExpectedErr: nil,
		},
		{
			Name: "Error",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("Create", applicationModel).Return(testErr).Once()
				return repo
			},
			Input:       modelInput,
			ExpectedErr: testErr,
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := application.NewService(repo)

			// when
			result, err := svc.Create(ctx, testCase.Input)

			// then
			assert.IsType(t, "string", result)
			assert.Equal(t, testCase.ExpectedErr, err)

			repo.AssertExpectations(t)
		})
	}
}

func TestService_Update(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	desc := "Lorem ipsum"
	modelInput := model.ApplicationInput{
		Name: "Bar",
	}

	inputApplicationModel := mock.MatchedBy(func(app *model.Application) bool {
		return app.Name == modelInput.Name
	})

	applicationModel := &model.Application{
		ID:          "foo",
		Name:        "Foo",
		Description: &desc,
	}

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, "tenant")

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.ApplicationRepository
		Input              model.ApplicationInput
		InputID            string
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", "foo").Return(applicationModel, nil).Once()
				repo.On("Update", inputApplicationModel).Return(nil).Once()
				return repo
			},
			InputID:            "foo",
			Input:              modelInput,
			ExpectedErrMessage: "",
		},
		{
			Name: "Update Error",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", "foo").Return(applicationModel, nil).Once()
				repo.On("Update", inputApplicationModel).Return(testErr).Once()
				return repo
			},
			InputID:            "foo",
			Input:              modelInput,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Get Error",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", "foo").Return(nil, testErr).Once()
				return repo
			},
			InputID:            "foo",
			Input:              modelInput,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := application.NewService(repo)

			// when
			err := svc.Update(ctx, testCase.InputID, testCase.Input)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestService_Delete(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "foo"

	desc := "Lorem ipsum"

	applicationModel := &model.Application{
		ID:          "foo",
		Name:        "Foo",
		Description: &desc,
	}

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, "tenant")

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.ApplicationRepository
		Input              model.ApplicationInput
		InputID            string
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", id).Return(applicationModel, nil).Once()
				repo.On("Delete", applicationModel).Return(nil).Once()
				return repo
			},
			InputID:            id,
			ExpectedErrMessage: "",
		},
		{
			Name: "Delete Error",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", id).Return(applicationModel, nil).Once()
				repo.On("Delete", applicationModel).Return(testErr).Once()
				return repo
			},
			InputID:            id,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Get Error",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", id).Return(nil, testErr).Once()
				return repo
			},
			InputID:            id,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := application.NewService(repo)

			// when
			err := svc.Delete(ctx, testCase.InputID)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestService_Get(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "foo"

	desc := "Lorem ipsum"

	applicationModel := &model.Application{
		ID:          "foo",
		Name:        "Foo",
		Description: &desc,
	}

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, "tenant")

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.ApplicationRepository
		Input              model.ApplicationInput
		InputID            string
		ExpectedApplication    *model.Application
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", id).Return(applicationModel, nil).Once()
				return repo
			},
			InputID:            id,
			ExpectedApplication:    applicationModel,
			ExpectedErrMessage: "",
		},
		{
			Name: "Get Error",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", id).Return(nil, testErr).Once()
				return repo
			},
			InputID:            id,
			ExpectedApplication:    applicationModel,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := application.NewService(repo)

			// when
			app, err := svc.Get(ctx, testCase.InputID)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedApplication, app)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestService_List(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	modelApplications := []*model.Application{
		fixModelApplication("foo", "Foo", "Lorem Ipsum"),
		fixModelApplication("bar", "Bar", "Lorem Ipsum"),
	}
	applicationPage := &model.ApplicationPage{
		Data:       modelApplications,
		TotalCount: len(modelApplications),
		PageInfo: &pagination.Page{
			HasNextPage: false,
			EndCursor:   "end",
			StartCursor: "start",
		},
	}

	first := 2
	after := "test"
	filter := []*labelfilter.LabelFilter{
		{Label: "", Values: []string{"foo", "bar"}, Operator: labelfilter.FilterOperatorAll},
	}

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, "tenant")

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.ApplicationRepository
		InputLabelFilters  []*labelfilter.LabelFilter
		InputPageSize      *int
		InputCursor        *string
		ExpectedResult     *model.ApplicationPage
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("List", filter, &first, &after).Return(applicationPage, nil).Once()
				return repo
			},
			InputLabelFilters:  filter,
			InputPageSize:      &first,
			InputCursor:        &after,
			ExpectedResult:     applicationPage,
			ExpectedErrMessage: "",
		},
		{
			Name: "Get Error",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("List", filter, &first, &after).Return(nil, testErr).Once()
				return repo
			},
			InputLabelFilters:  filter,
			InputPageSize:      &first,
			InputCursor:        &after,
			ExpectedResult:     nil,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := application.NewService(repo)

			// when
			app, err := svc.List(ctx, testCase.InputLabelFilters, testCase.InputPageSize, testCase.InputCursor)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedResult, app)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestService_AddAnnotation(t *testing.T) {
	// given
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, "tenant")

	testErr := errors.New("Test error")

	desc := "Lorem ipsum"

	applicationID := "foo"
	modifiedApplicationModel := fixModelApplicationWithAnnotations(applicationID, "Foo", map[string]interface{}{
		"key": "value",
	})
	modifiedApplicationModel.Description = &desc

	annotationKey := "key"
	annotationValue := "value"

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.ApplicationRepository
		InputApplicationID     string
		InputKey           string
		InputValue         string
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", applicationID).Return(fixModelApplication(applicationID, "Foo", desc), nil).Once()
				repo.On("Update", modifiedApplicationModel).Return(nil).Once()

				return repo
			},
			InputApplicationID:     applicationID,
			InputKey:           annotationKey,
			InputValue:         annotationValue,
			ExpectedErrMessage: "",
		},
		{
			Name: "Update Error",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", applicationID).Return(fixModelApplication(applicationID, "Foo", desc), nil).Once()
				repo.On("Update", modifiedApplicationModel).Return(testErr).Once()

				return repo
			},
			InputApplicationID:     applicationID,
			InputKey:           annotationKey,
			InputValue:         annotationValue,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Get Error",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", applicationID).Return(nil, testErr).Once()

				return repo
			},
			InputApplicationID:     applicationID,
			InputKey:           annotationKey,
			InputValue:         annotationValue,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := application.NewService(repo)

			// when
			err := svc.AddAnnotation(ctx, testCase.InputApplicationID, testCase.InputKey, testCase.InputValue)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestService_DeleteAnnotation(t *testing.T) {
	// given
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, "tenant")

	testErr := errors.New("Test error")

	applicationID := "foo"
	modifiedApplicationModel := fixModelApplicationWithAnnotations(applicationID, "Foo", map[string]interface{}{})

	annotationKey := "key"

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.ApplicationRepository
		InputApplicationID     string
		InputKey           string
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", applicationID).Return(
					fixModelApplicationWithAnnotations(applicationID, "Foo", map[string]interface{}{
						"key": "value",
					}), nil).Once()
				repo.On("Update", modifiedApplicationModel).Return(nil).Once()

				return repo
			},
			InputApplicationID:     applicationID,
			InputKey:           annotationKey,
			ExpectedErrMessage: "",
		},
		{
			Name: "Update Error",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", applicationID).Return(
					fixModelApplicationWithAnnotations(applicationID, "Foo", map[string]interface{}{
						"key": "value",
					}), nil).Once()
				repo.On("Update", modifiedApplicationModel).Return(testErr).Once()

				return repo
			},
			InputApplicationID:     applicationID,
			InputKey:           annotationKey,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Get Error",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", applicationID).Return(nil, testErr).Once()

				return repo
			},
			InputApplicationID:     applicationID,
			InputKey:           annotationKey,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := application.NewService(repo)

			// when
			err := svc.DeleteAnnotation(ctx, testCase.InputApplicationID, testCase.InputKey)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestService_AddLabel(t *testing.T) {
	// given
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, "tenant")

	testErr := errors.New("Test error")

	desc := "Lorem ipsum"

	applicationID := "foo"
	modifiedApplicationModel := fixModelApplicationWithLabels(applicationID, "Foo", map[string][]string{
		"key": {"value1"},
	})
	modifiedApplicationModel.Description = &desc

	labelKey := "key"
	labelValues := []string{"value1"}

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.ApplicationRepository
		InputApplicationID     string
		InputKey           string
		InputValues        []string
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", applicationID).Return(fixModelApplication(applicationID, "Foo", desc), nil).Once()
				repo.On("Update", modifiedApplicationModel).Return(nil).Once()

				return repo
			},
			InputApplicationID:     applicationID,
			InputKey:           labelKey,
			InputValues:        labelValues,
			ExpectedErrMessage: "",
		},
		{
			Name: "Update Error",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", applicationID).Return(fixModelApplication(applicationID, "Foo", desc), nil).Once()
				repo.On("Update", modifiedApplicationModel).Return(testErr).Once()

				return repo
			},
			InputApplicationID:     applicationID,
			InputKey:           labelKey,
			InputValues:        labelValues,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Get Error",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", applicationID).Return(nil, testErr).Once()

				return repo
			},
			InputApplicationID:     applicationID,
			InputKey:           labelKey,
			InputValues:        labelValues,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := application.NewService(repo)

			// when
			err := svc.AddLabel(ctx, testCase.InputApplicationID, testCase.InputKey, testCase.InputValues)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestService_DeleteLabel(t *testing.T) {
	// given
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, "tenant")

	testErr := errors.New("Test error")

	applicationID := "foo"
	modifiedApplicationModel := fixModelApplicationWithLabels(applicationID, "Foo", map[string][]string{})

	labelKey := "key"
	labelValues := []string{"value1", "value2"}

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.ApplicationRepository
		InputApplicationID     string
		InputKey           string
		InputValues        []string
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", applicationID).Return(
					fixModelApplicationWithLabels(applicationID, "Foo", map[string][]string{
						"key": {"value1", "value2"},
					}), nil).Once()
				repo.On("Update", modifiedApplicationModel).Return(nil).Once()

				return repo
			},
			InputApplicationID:     applicationID,
			InputKey:           labelKey,
			InputValues:        labelValues,
			ExpectedErrMessage: "",
		},
		{
			Name: "Update Error",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", applicationID).Return(
					fixModelApplicationWithLabels(applicationID, "Foo", map[string][]string{
						"key": {"value1", "value2"},
					}), nil).Once()
				repo.On("Update", modifiedApplicationModel).Return(testErr).Once()

				return repo
			},
			InputApplicationID:     applicationID,
			InputKey:           labelKey,
			InputValues:        labelValues,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Get Error",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", applicationID).Return(nil, testErr).Once()

				return repo
			},
			InputApplicationID:     applicationID,
			InputKey:           labelKey,
			InputValues:        labelValues,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := application.NewService(repo)

			// when
			err := svc.DeleteLabel(ctx, testCase.InputApplicationID, testCase.InputKey, testCase.InputValues)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}
}

