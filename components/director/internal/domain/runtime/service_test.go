package runtime_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime"
	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime/automock"
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/tenant"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_Create(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "foo"
	desc := "Lorem ipsum"
	modelInput := model.RuntimeInput{
		Name:        "Foo",
		Description: &desc,
	}

	runtimeModel := mock.MatchedBy(func(rtm *model.Runtime) bool {
		return rtm.Name == modelInput.Name && rtm.Description == modelInput.Description &&
			rtm.AgentAuth != nil && rtm.AgentAuth.Credential.Basic != nil &&
			rtm.Status.Condition == model.RuntimeStatusConditionInitial
	})

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, "tenant")

	testCases := []struct {
		Name         string
		RepositoryFn func() *automock.RuntimeRepository
		UIDServiceFn func() *automock.UIDService
		Input        model.RuntimeInput
		ExpectedErr  error
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Create", runtimeModel).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id).Once()
				return svc
			},
			Input:       modelInput,
			ExpectedErr: nil,
		},
		{
			Name: "Returns error when runtime creation failed",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Create", runtimeModel).Return(testErr).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return("").Once()
				return svc
			},
			Input:       modelInput,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			idSvc := testCase.UIDServiceFn()
			svc := runtime.NewService(repo, idSvc)

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
	modelInput := model.RuntimeInput{
		Name: "Bar",
	}

	inputRuntimeModel := mock.MatchedBy(func(rtm *model.Runtime) bool {
		return rtm.Name == modelInput.Name
	})

	runtimeModel := &model.Runtime{
		ID:          "foo",
		Name:        "Foo",
		Description: &desc,
	}

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, "tenant")

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.RuntimeRepository
		Input              model.RuntimeInput
		InputID            string
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", "foo").Return(runtimeModel, nil).Once()
				repo.On("Update", inputRuntimeModel).Return(nil).Once()
				return repo
			},
			InputID:            "foo",
			Input:              modelInput,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when application update failed",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", "foo").Return(runtimeModel, nil).Once()
				repo.On("Update", inputRuntimeModel).Return(testErr).Once()
				return repo
			},
			InputID:            "foo",
			Input:              modelInput,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when runtime retrieval failed",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", "foo").Return(nil, testErr).Once()
				return repo
			},
			InputID:            "foo",
			Input:              modelInput,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := runtime.NewService(repo, nil)

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

	runtimeModel := &model.Runtime{
		ID:          "foo",
		Name:        "Foo",
		Description: &desc,
	}

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, "tenant")

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.RuntimeRepository
		Input              model.RuntimeInput
		InputID            string
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", id).Return(runtimeModel, nil).Once()
				repo.On("Delete", runtimeModel).Return(nil).Once()
				return repo
			},
			InputID:            id,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when runtime deletion failed",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", id).Return(runtimeModel, nil).Once()
				repo.On("Delete", runtimeModel).Return(testErr).Once()
				return repo
			},
			InputID:            id,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when runtime retrieval failed",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", id).Return(nil, testErr).Once()
				return repo
			},
			InputID:            id,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := runtime.NewService(repo, nil)

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

	runtimeModel := &model.Runtime{
		ID:          "foo",
		Name:        "Foo",
		Description: &desc,
	}

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, "tenant")

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.RuntimeRepository
		Input              model.RuntimeInput
		InputID            string
		ExpectedRuntime    *model.Runtime
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", id).Return(runtimeModel, nil).Once()
				return repo
			},
			InputID:            id,
			ExpectedRuntime:    runtimeModel,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when runtime retrieval failed",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", id).Return(nil, testErr).Once()
				return repo
			},
			InputID:            id,
			ExpectedRuntime:    runtimeModel,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := runtime.NewService(repo, nil)

			// when
			rtm, err := svc.Get(ctx, testCase.InputID)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedRuntime, rtm)
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

	modelRuntimes := []*model.Runtime{
		fixModelRuntime("foo", "Foo", "Lorem Ipsum"),
		fixModelRuntime("bar", "Bar", "Lorem Ipsum"),
	}
	runtimePage := &model.RuntimePage{
		Data:       modelRuntimes,
		TotalCount: len(modelRuntimes),
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
		RepositoryFn       func() *automock.RuntimeRepository
		InputLabelFilters  []*labelfilter.LabelFilter
		InputPageSize      *int
		InputCursor        *string
		ExpectedResult     *model.RuntimePage
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("List", filter, &first, &after).Return(runtimePage, nil).Once()
				return repo
			},
			InputLabelFilters:  filter,
			InputPageSize:      &first,
			InputCursor:        &after,
			ExpectedResult:     runtimePage,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when runtime listing failed",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
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

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := runtime.NewService(repo, nil)

			// when
			rtm, err := svc.List(ctx, testCase.InputLabelFilters, testCase.InputPageSize, testCase.InputCursor)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedResult, rtm)
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

	runtimeID := "foo"
	modifiedRuntimeModel := fixModelRuntimeWithAnnotations(runtimeID, "Foo", map[string]interface{}{
		"key": "value",
	})
	modifiedRuntimeModel.Description = &desc

	annotationKey := "key"
	annotationValue := "value"

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.RuntimeRepository
		InputRuntimeID     string
		InputKey           string
		InputValue         string
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", runtimeID).Return(fixModelRuntime(runtimeID, "Foo", desc), nil).Once()
				repo.On("Update", modifiedRuntimeModel).Return(nil).Once()

				return repo
			},
			InputRuntimeID:     runtimeID,
			InputKey:           annotationKey,
			InputValue:         annotationValue,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when runtime update failed",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", runtimeID).Return(fixModelRuntime(runtimeID, "Foo", desc), nil).Once()
				repo.On("Update", modifiedRuntimeModel).Return(testErr).Once()

				return repo
			},
			InputRuntimeID:     runtimeID,
			InputKey:           annotationKey,
			InputValue:         annotationValue,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when runtime retrieval failed",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", runtimeID).Return(nil, testErr).Once()

				return repo
			},
			InputRuntimeID:     runtimeID,
			InputKey:           annotationKey,
			InputValue:         annotationValue,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := runtime.NewService(repo, nil)

			// when
			err := svc.AddAnnotation(ctx, testCase.InputRuntimeID, testCase.InputKey, testCase.InputValue)

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

	runtimeID := "foo"
	modifiedRuntimeModel := fixModelRuntimeWithAnnotations(runtimeID, "Foo", map[string]interface{}{})

	annotationKey := "key"

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.RuntimeRepository
		InputRuntimeID     string
		InputKey           string
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", runtimeID).Return(
					fixModelRuntimeWithAnnotations(runtimeID, "Foo", map[string]interface{}{
						"key": "value",
					}), nil).Once()
				repo.On("Update", modifiedRuntimeModel).Return(nil).Once()

				return repo
			},
			InputRuntimeID:     runtimeID,
			InputKey:           annotationKey,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when runtime update failed",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", runtimeID).Return(
					fixModelRuntimeWithAnnotations(runtimeID, "Foo", map[string]interface{}{
						"key": "value",
					}), nil).Once()
				repo.On("Update", modifiedRuntimeModel).Return(testErr).Once()

				return repo
			},
			InputRuntimeID:     runtimeID,
			InputKey:           annotationKey,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when runtime retrieval failed",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", runtimeID).Return(nil, testErr).Once()

				return repo
			},
			InputRuntimeID:     runtimeID,
			InputKey:           annotationKey,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := runtime.NewService(repo, nil)

			// when
			err := svc.DeleteAnnotation(ctx, testCase.InputRuntimeID, testCase.InputKey)

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

	runtimeID := "foo"
	modifiedRuntimeModel := fixModelRuntimeWithLabels(runtimeID, "Foo", map[string][]string{
		"key": {"value1"},
	})
	modifiedRuntimeModel.Description = &desc

	labelKey := "key"
	labelValues := []string{"value1"}

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.RuntimeRepository
		InputRuntimeID     string
		InputKey           string
		InputValues        []string
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", runtimeID).Return(fixModelRuntime(runtimeID, "Foo", desc), nil).Once()
				repo.On("Update", modifiedRuntimeModel).Return(nil).Once()

				return repo
			},
			InputRuntimeID:     runtimeID,
			InputKey:           labelKey,
			InputValues:        labelValues,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when runtime update failed",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", runtimeID).Return(fixModelRuntime(runtimeID, "Foo", desc), nil).Once()
				repo.On("Update", modifiedRuntimeModel).Return(testErr).Once()

				return repo
			},
			InputRuntimeID:     runtimeID,
			InputKey:           labelKey,
			InputValues:        labelValues,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when runtime retrieval failed",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", runtimeID).Return(nil, testErr).Once()

				return repo
			},
			InputRuntimeID:     runtimeID,
			InputKey:           labelKey,
			InputValues:        labelValues,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := runtime.NewService(repo, nil)

			// when
			err := svc.AddLabel(ctx, testCase.InputRuntimeID, testCase.InputKey, testCase.InputValues)

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

	runtimeID := "foo"
	modifiedRuntimeModel := fixModelRuntimeWithLabels(runtimeID, "Foo", map[string][]string{})

	labelKey := "key"
	labelValues := []string{"value1", "value2"}

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.RuntimeRepository
		InputRuntimeID     string
		InputKey           string
		InputValues        []string
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", runtimeID).Return(
					fixModelRuntimeWithLabels(runtimeID, "Foo", map[string][]string{
						"key": {"value1", "value2"},
					}), nil).Once()
				repo.On("Update", modifiedRuntimeModel).Return(nil).Once()

				return repo
			},
			InputRuntimeID:     runtimeID,
			InputKey:           labelKey,
			InputValues:        labelValues,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when runtime update failed",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", runtimeID).Return(
					fixModelRuntimeWithLabels(runtimeID, "Foo", map[string][]string{
						"key": {"value1", "value2"},
					}), nil).Once()
				repo.On("Update", modifiedRuntimeModel).Return(testErr).Once()

				return repo
			},
			InputRuntimeID:     runtimeID,
			InputKey:           labelKey,
			InputValues:        labelValues,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when runtime retrieval failed",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", runtimeID).Return(nil, testErr).Once()

				return repo
			},
			InputRuntimeID:     runtimeID,
			InputKey:           labelKey,
			InputValues:        labelValues,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := runtime.NewService(repo, nil)

			// when
			err := svc.DeleteLabel(ctx, testCase.InputRuntimeID, testCase.InputKey, testCase.InputValues)

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
