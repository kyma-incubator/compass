package runtime_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"

	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime"
	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/tenant"
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
		Name:        "foo.bar-not",
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
		Name                 string
		RuntimeRepositoryFn  func() *automock.RuntimeRepository
		LabelUpsertServiceFn func() *automock.LabelUpsertService
		UIDServiceFn         func() *automock.UIDService
		Input                model.RuntimeInput
		ExpectedErr          error
	}{
		{
			Name: "Success",
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Create", ctx, runtimeModel).Return(nil).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				repo := &automock.LabelUpsertService{}
				repo.On("UpsertMultipleLabels", ctx, "tenant", model.RuntimeLabelableObject, id, modelInput.Labels).Return(nil).Once()
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
			Name: "Returns error when name is empty",
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				repo := &automock.LabelUpsertService{}
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				return svc
			},
			Input:       model.RuntimeInput{Name: ""},
			ExpectedErr: errors.New("a DNS-1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character")},
		{
			Name: "Returns error when name contains upper case letter",
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				repo := &automock.LabelUpsertService{}
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				return svc
			},
			Input:       model.RuntimeInput{Name: "upperCase"},
			ExpectedErr: errors.New("a DNS-1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character"),
		},
		{
			Name: "Returns error when runtime creation failed",
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Create", ctx, runtimeModel).Return(testErr).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				repo := &automock.LabelUpsertService{}
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
			repo := testCase.RuntimeRepositoryFn()
			idSvc := testCase.UIDServiceFn()
			labelSvc := testCase.LabelUpsertServiceFn()
			svc := runtime.NewService(repo, nil, labelSvc, idSvc)

			// when
			result, err := svc.Create(ctx, testCase.Input)

			// then
			assert.IsType(t, "string", result)
			if err == nil {
				require.Nil(t, testCase.ExpectedErr)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			}

			repo.AssertExpectations(t)
			idSvc.AssertExpectations(t)
			labelSvc.AssertExpectations(t)
		})
	}
}

func TestService_Update(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	desc := "Lorem ipsum"

	modelInput := model.RuntimeInput{
		Name: "bar",
		Labels: map[string]interface{}{
			"label1": "val1",
		},
	}

	inputRuntimeModel := mock.MatchedBy(func(rtm *model.Runtime) bool {
		return rtm.Name == modelInput.Name
	})

	runtimeModel := &model.Runtime{
		ID:          "foo",
		Name:        "Foo",
		Description: &desc,
	}

	tnt := "tenant"
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt)

	testCases := []struct {
		Name                 string
		RepositoryFn         func() *automock.RuntimeRepository
		LabelRepositoryFn    func() *automock.LabelRepository
		LabelUpsertServiceFn func() *automock.LabelUpsertService
		Input                model.RuntimeInput
		InputID              string
		ExpectedErrMessage   string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, tnt, "foo").Return(runtimeModel, nil).Once()
				repo.On("Update", ctx, inputRuntimeModel).Return(nil).Once()
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("DeleteAll", ctx, tnt, model.RuntimeLabelableObject, runtimeModel.ID).Return(nil).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				repo := &automock.LabelUpsertService{}
				repo.On("UpsertMultipleLabels", ctx, tnt, model.RuntimeLabelableObject, runtimeModel.ID, modelInput.Labels).Return(nil).Once()
				return repo
			},
			InputID:            "foo",
			Input:              modelInput,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when name is empty",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				repo := &automock.LabelUpsertService{}
				return repo
			},
			Input:              model.RuntimeInput{Name: ""},
			ExpectedErrMessage: "a DNS-1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character",
		},
		{
			Name: "Returns error when application update failed",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, tnt, "foo").Return(runtimeModel, nil).Once()
				repo.On("Update", ctx, inputRuntimeModel).Return(testErr).Once()
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				repo := &automock.LabelUpsertService{}
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
				repo.On("GetByID", ctx, tnt, "foo").Return(nil, testErr).Once()
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				repo := &automock.LabelUpsertService{}
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
			labelRepo := testCase.LabelRepositoryFn()
			labelSvc := testCase.LabelUpsertServiceFn()
			svc := runtime.NewService(repo, labelRepo, labelSvc, nil)

			// when
			err := svc.Update(ctx, testCase.InputID, testCase.Input)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
			labelRepo.AssertExpectations(t)
			labelSvc.AssertExpectations(t)
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

	tnt := "tenant"

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt)

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
				repo.On("Exists", ctx, tnt, id).Return(true, nil).Once()
				repo.On("Delete", ctx, runtimeModel.ID).Return(nil).Once()
				return repo
			},
			InputID:            id,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when runtime deletion failed",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, id).Return(true, nil).Once()
				repo.On("Delete", ctx, runtimeModel.ID).Return(testErr).Once()
				return repo
			},
			InputID:            id,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when runtime doesn't exist",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, id).Return(false, testErr).Once()
				return repo
			},
			InputID:            id,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			svc := runtime.NewService(repo, nil, nil, nil)

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
	tnt := "tenant"

	runtimeModel := &model.Runtime{
		ID:          "foo",
		Name:        "Foo",
		Description: &desc,
	}

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt)

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
				repo.On("GetByID", ctx, tnt, id).Return(runtimeModel, nil).Once()
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
				repo.On("GetByID", ctx, tnt, id).Return(nil, testErr).Once()
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

			svc := runtime.NewService(repo, nil, nil, nil)

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
		fixModelRuntime("foo", "tenant-foo", "Foo", "Lorem Ipsum"),
		fixModelRuntime("bar", "tenant-bar", "Bar", "Lorem Ipsum"),
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
	filter := []*labelfilter.LabelFilter{{Key: ""}}

	tnt := "tenant"

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt)

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
				repo.On("List", ctx, tnt, filter, &first, &after).Return(runtimePage, nil).Once()
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
				repo.On("List", ctx, tnt, filter, &first, &after).Return(nil, testErr).Once()
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

			svc := runtime.NewService(repo, nil, nil, nil)

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

func TestService_GetLabel(t *testing.T) {
	// given
	tnt := "tenant"
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt)

	testErr := errors.New("Test error")

	runtimeID := "foo"
	labelKey := "key"
	labelValue := []string{"value1"}

	label := &model.LabelInput{
		Key:        labelKey,
		Value:      labelValue,
		ObjectID:   runtimeID,
		ObjectType: model.RuntimeLabelableObject,
	}

	modelLabel := &model.Label{
		ID:         "5d23d9d9-3d04-4fa9-95e6-d22e1ae62c11",
		Tenant:     tnt,
		Key:        labelKey,
		Value:      labelValue,
		ObjectID:   runtimeID,
		ObjectType: model.RuntimeLabelableObject,
	}

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.RuntimeRepository
		LabelRepositoryFn  func() *automock.LabelRepository
		InputApplicationID string
		InputLabel         *model.LabelInput
		ExpectedLabel      *model.Label
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(true, nil).Once()
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("GetByKey", ctx, tnt, model.RuntimeLabelableObject, runtimeID, labelKey).Return(modelLabel, nil).Once()
				return repo
			},
			InputApplicationID: runtimeID,
			InputLabel:         label,
			ExpectedLabel:      modelLabel,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when label receiving failed",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(true, nil).Once()

				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("GetByKey", ctx, tnt, model.RuntimeLabelableObject, runtimeID, labelKey).Return(nil, testErr).Once()
				return repo
			},
			InputApplicationID: runtimeID,
			InputLabel:         label,
			ExpectedLabel:      nil,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when application doesn't exist",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(false, testErr).Once()

				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				return repo
			},
			InputApplicationID: runtimeID,
			InputLabel:         label,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			labelRepo := testCase.LabelRepositoryFn()
			svc := runtime.NewService(repo, labelRepo, nil, nil)

			// when
			l, err := svc.GetLabel(ctx, testCase.InputApplicationID, testCase.InputLabel.Key)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, l, testCase.ExpectedLabel)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
			labelRepo.AssertExpectations(t)
		})
	}
}

func TestService_ListLabel(t *testing.T) {
	// given
	tnt := "tenant"
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt)

	testErr := errors.New("Test error")

	runtimeID := "foo"
	labelKey := "key"
	labelValue := []string{"value1"}

	label := &model.LabelInput{
		Key:        labelKey,
		Value:      labelValue,
		ObjectID:   runtimeID,
		ObjectType: model.RuntimeLabelableObject,
	}

	modelLabel := &model.Label{
		ID:         "5d23d9d9-3d04-4fa9-95e6-d22e1ae62c11",
		Tenant:     tnt,
		Key:        labelKey,
		Value:      labelValue,
		ObjectID:   runtimeID,
		ObjectType: model.RuntimeLabelableObject,
	}

	labels := map[string]*model.Label{"first": modelLabel, "second": modelLabel}
	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.RuntimeRepository
		LabelRepositoryFn  func() *automock.LabelRepository
		InputApplicationID string
		InputLabel         *model.LabelInput
		ExpectedOutput     map[string]*model.Label
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(true, nil).Once()
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("List", ctx, tnt, model.RuntimeLabelableObject, runtimeID).Return(labels, nil).Once()
				return repo
			},
			InputApplicationID: runtimeID,
			InputLabel:         label,
			ExpectedOutput:     labels,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when labels receiving failed",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(true, nil).Once()

				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("List", ctx, tnt, model.RuntimeLabelableObject, runtimeID).Return(nil, testErr).Once()
				return repo
			},
			InputApplicationID: runtimeID,
			InputLabel:         label,
			ExpectedOutput:     nil,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when application doesn't exist",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(false, testErr).Once()

				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				return repo
			},
			InputApplicationID: runtimeID,
			InputLabel:         label,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			labelRepo := testCase.LabelRepositoryFn()
			svc := runtime.NewService(repo, labelRepo, nil, nil)

			// when
			l, err := svc.ListLabels(ctx, testCase.InputApplicationID)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, l, testCase.ExpectedOutput)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
			labelRepo.AssertExpectations(t)
		})
	}
}

func TestService_SetLabel(t *testing.T) {
	// given
	tnt := "tenant"
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt)

	testErr := errors.New("Test error")

	runtimeID := "foo"

	modelLabel := model.LabelInput{
		Key:        "key",
		Value:      []string{"value1"},
		ObjectID:   runtimeID,
		ObjectType: model.RuntimeLabelableObject,
	}

	testCases := []struct {
		Name                 string
		RepositoryFn         func() *automock.RuntimeRepository
		LabelUpsertServiceFn func() *automock.LabelUpsertService
		InputRuntimeID       string
		InputLabel           *model.LabelInput
		ExpectedErrMessage   string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(true, nil).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				svc.On("UpsertLabel", ctx, tnt, &modelLabel).Return(nil).Once()
				return svc
			},
			InputRuntimeID:     runtimeID,
			InputLabel:         &modelLabel,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when runtime update failed",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(true, nil).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				svc.On("UpsertLabel", ctx, tnt, &modelLabel).Return(testErr).Once()
				return svc
			},
			InputRuntimeID:     runtimeID,
			InputLabel:         &modelLabel,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when runtime retrieval failed",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(false, testErr).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				return svc
			},
			InputRuntimeID:     runtimeID,
			InputLabel:         &modelLabel,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			labelSvc := testCase.LabelUpsertServiceFn()
			svc := runtime.NewService(repo, nil, labelSvc, nil)

			// when
			err := svc.SetLabel(ctx, testCase.InputLabel)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
			labelSvc.AssertExpectations(t)
		})
	}
}

func TestService_DeleteLabel(t *testing.T) {
	// given
	tnt := "tenant"
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt)

	testErr := errors.New("Test error")

	runtimeID := "foo"
	//modifiedRuntimeModel := fixModelRuntimeWithLabels(runtimeID, tnt, "Foo", map[string]interface{}{})

	labelKey := "key"

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.RuntimeRepository
		LabelRepositoryFn  func() *automock.LabelRepository
		InputRuntimeID     string
		InputKey           string
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(true, nil).Once()
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("Delete", ctx, tnt, model.RuntimeLabelableObject, runtimeID, labelKey).Return(nil).Once()
				return repo
			},
			InputRuntimeID:     runtimeID,
			InputKey:           labelKey,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when runtime label update failed",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(true, nil).Once()
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("Delete", ctx, tnt, model.RuntimeLabelableObject, runtimeID, labelKey).Return(testErr).Once()
				return repo
			},
			InputRuntimeID:     runtimeID,
			InputKey:           labelKey,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when runtime retrieval failed",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(false, testErr).Once()
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				return repo
			},
			InputRuntimeID:     runtimeID,
			InputKey:           labelKey,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			labelRepo := testCase.LabelRepositoryFn()
			svc := runtime.NewService(repo, labelRepo, nil, nil)

			// when
			err := svc.DeleteLabel(ctx, testCase.InputRuntimeID, testCase.InputKey)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}
}
