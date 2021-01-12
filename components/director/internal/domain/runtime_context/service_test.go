package runtime_context_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime_context"
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"

	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime_context/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_Create(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "foo"
	runtimeID := "runtime_id"
	key := "key"
	val := "val"
	labels := map[string]interface{}{
		model.ScenariosKey: "DEFAULT",
	}
	modelInput := model.RuntimeContextInput{
		Key:       key,
		Value:     val,
		RuntimeID: runtimeID,
		Labels:    labels,
	}

	modelInputWithoutLabels := model.RuntimeContextInput{
		Key:       key,
		Value:     val,
		RuntimeID: runtimeID,
	}

	var nilLabels map[string]interface{}

	runtimeCtxModel := mock.MatchedBy(func(rtmCtx *model.RuntimeContext) bool {
		return rtmCtx.Key == modelInput.Key && rtmCtx.Value == modelInput.Value && rtmCtx.RuntimeID == modelInput.RuntimeID
	})

	tnt := "tenant"
	externalTnt := "external-tnt"
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	testCases := []struct {
		Name                       string
		RuntimeContextRepositoryFn func() *automock.RuntimeContextRepository
		LabelUpsertServiceFn       func() *automock.LabelUpsertService
		UIDServiceFn               func() *automock.UIDService
		Input                      model.RuntimeContextInput
		ExpectedErr                error
	}{
		{
			Name: "Success",
			RuntimeContextRepositoryFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("Create", ctx, runtimeCtxModel).Return(nil).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				repo := &automock.LabelUpsertService{}
				repo.On("UpsertMultipleLabels", ctx, tnt, model.RuntimeContextLabelableObject, id, modelInput.Labels).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			Input:       modelInput,
			ExpectedErr: nil,
		},
		{
			Name: "Success when labels are empty",
			RuntimeContextRepositoryFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("Create", ctx, runtimeCtxModel).Return(nil).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				repo := &automock.LabelUpsertService{}
				repo.On("UpsertMultipleLabels", ctx, tnt, model.RuntimeContextLabelableObject, id, nilLabels).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			Input:       modelInputWithoutLabels,
			ExpectedErr: nil,
		},
		{
			Name: "Returns error when runtime context creation failed",
			RuntimeContextRepositoryFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("Create", ctx, runtimeCtxModel).Return(testErr).Once()
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
		{
			Name: "Returns error when label upserting failed",
			RuntimeContextRepositoryFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("Create", ctx, runtimeCtxModel).Return(nil).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				repo := &automock.LabelUpsertService{}
				repo.On("UpsertMultipleLabels", ctx, "tenant", model.RuntimeContextLabelableObject, id, modelInput.Labels).Return(testErr).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			Input:       modelInput,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RuntimeContextRepositoryFn()
			idSvc := testCase.UIDServiceFn()
			labelSvc := testCase.LabelUpsertServiceFn()
			svc := runtime_context.NewService(repo, nil, labelSvc, idSvc)

			// when
			result, err := svc.Create(ctx, testCase.Input)

			// then
			assert.IsType(t, "string", result)
			if err == nil {
				require.Nil(t, testCase.ExpectedErr)
			} else {
				require.NotNil(t, testCase.ExpectedErr)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			}

			repo.AssertExpectations(t)
			idSvc.AssertExpectations(t)
			labelSvc.AssertExpectations(t)
		})
	}

	t.Run("Returns error on loading tenant", func(t *testing.T) {
		// given
		svc := runtime_context.NewService(nil, nil, nil, nil)
		// when
		_, err := svc.Create(context.TODO(), model.RuntimeContextInput{})
		// then
		require.Error(t, err)
		assert.EqualError(t, err, "while loading tenant from context: cannot read tenant from context")
	})
}

func TestService_Update(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "foo"
	key := "key"
	val := "value"
	runtimeID := "runtime_id"

	labels := map[string]interface{}{
		"label1": "val1",
	}
	modelInput := model.RuntimeContextInput{
		Key:       key,
		Value:     val,
		RuntimeID: runtimeID,
		Labels:    labels,
	}

	inputRuntimeContextModel := mock.MatchedBy(func(rtmCtx *model.RuntimeContext) bool {
		return rtmCtx.Key == modelInput.Key && rtmCtx.Value == modelInput.Value && rtmCtx.RuntimeID == modelInput.RuntimeID
	})

	runtimeCtxModel := &model.RuntimeContext{
		ID:        id,
		Key:       key,
		Value:     val,
		RuntimeID: runtimeID,
	}

	tnt := "tenant"
	externalTnt := "external-tnt"
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	testCases := []struct {
		Name                 string
		RepositoryFn         func() *automock.RuntimeContextRepository
		LabelRepositoryFn    func() *automock.LabelRepository
		LabelUpsertServiceFn func() *automock.LabelUpsertService
		Input                model.RuntimeContextInput
		InputID              string
		ExpectedErrMessage   string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, tnt, "foo").Return(runtimeCtxModel, nil).Once()
				repo.On("Update", ctx, inputRuntimeContextModel).Return(nil).Once()
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("DeleteAll", ctx, tnt, model.RuntimeContextLabelableObject, runtimeCtxModel.ID).Return(nil).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				repo := &automock.LabelUpsertService{}
				repo.On("UpsertMultipleLabels", ctx, tnt, model.RuntimeContextLabelableObject, runtimeCtxModel.ID, modelInput.Labels).Return(nil).Once()
				return repo
			},
			InputID:            id,
			Input:              modelInput,
			ExpectedErrMessage: "",
		},
		{
			Name: "Success when labels are nil",
			RepositoryFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, tnt, "foo").Return(runtimeCtxModel, nil).Once()
				repo.On("Update", ctx, inputRuntimeContextModel).Return(nil).Once()
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("DeleteAll", ctx, tnt, model.RuntimeContextLabelableObject, runtimeCtxModel.ID).Return(nil).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				repo := &automock.LabelUpsertService{}
				return repo
			},
			InputID: "foo",
			Input: model.RuntimeContextInput{
				Key:       key,
				Value:     val,
				RuntimeID: runtimeID,
			},
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when runtime context update failed",
			RepositoryFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, tnt, "foo").Return(runtimeCtxModel, nil).Once()
				repo.On("Update", ctx, inputRuntimeContextModel).Return(testErr).Once()
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
			InputID:            id,
			Input:              modelInput,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when runtime context retrieval failed",
			RepositoryFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
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
			InputID:            id,
			Input:              modelInput,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when label deletion failed",
			RepositoryFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, tnt, "foo").Return(runtimeCtxModel, nil).Once()
				repo.On("Update", ctx, inputRuntimeContextModel).Return(nil).Once()
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("DeleteAll", ctx, tnt, model.RuntimeContextLabelableObject, runtimeCtxModel.ID).Return(testErr).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				repo := &automock.LabelUpsertService{}
				return repo
			},
			InputID:            id,
			Input:              modelInput,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when upserting labels failed",
			RepositoryFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, tnt, "foo").Return(runtimeCtxModel, nil).Once()
				repo.On("Update", ctx, inputRuntimeContextModel).Return(nil).Once()
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("DeleteAll", ctx, tnt, model.RuntimeContextLabelableObject, runtimeCtxModel.ID).Return(nil).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				repo := &automock.LabelUpsertService{}
				repo.On("UpsertMultipleLabels", ctx, tnt, model.RuntimeContextLabelableObject, runtimeCtxModel.ID, modelInput.Labels).Return(testErr).Once()
				return repo
			},
			InputID:            id,
			Input:              modelInput,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			labelRepo := testCase.LabelRepositoryFn()
			labelSvc := testCase.LabelUpsertServiceFn()
			svc := runtime_context.NewService(repo, labelRepo, labelSvc, nil)

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

	t.Run("Returns error on loading tenant", func(t *testing.T) {
		// given
		svc := runtime_context.NewService(nil, nil, nil, nil)
		// when
		err := svc.Update(context.TODO(), "id", model.RuntimeContextInput{})
		// then
		require.Error(t, err)
		assert.EqualError(t, err, "while loading tenant from context: cannot read tenant from context")
	})
}

func TestService_Delete(t *testing.T) {
	// given
	testErr := errors.New("Test error")
	id := "foo"
	key := "key"
	val := "value"
	runtimeID := "runtime_id"

	runtimeCtxModel := &model.RuntimeContext{
		ID:        id,
		Key:       key,
		Value:     val,
		RuntimeID: runtimeID,
	}

	tnt := "tenant"
	externalTnt := "external-tnt"
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.RuntimeContextRepository
		Input              model.RuntimeContextInput
		InputID            string
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("Delete", ctx, tnt, runtimeCtxModel.ID).Return(nil).Once()
				return repo
			},
			InputID:            id,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when runtime context deletion failed",
			RepositoryFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("Delete", ctx, tnt, runtimeCtxModel.ID).Return(testErr).Once()
				return repo
			},
			InputID:            id,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			svc := runtime_context.NewService(repo, nil, nil, nil)

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

	t.Run("Returns error on loading tenant", func(t *testing.T) {
		// given
		svc := runtime_context.NewService(nil, nil, nil, nil)
		// when
		err := svc.Delete(context.TODO(), "id")
		// then
		require.Error(t, err)
		assert.EqualError(t, err, "while loading tenant from context: cannot read tenant from context")
	})
}

func TestService_Get(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "foo"
	key := "key"
	val := "value"
	runtimeID := "runtime_id"
	tnt := "tenant"
	externalTnt := "external-tnt"

	runtimeCtxModel := &model.RuntimeContext{
		ID:        id,
		Key:       key,
		Value:     val,
		RuntimeID: runtimeID,
	}

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	testCases := []struct {
		Name                   string
		RepositoryFn           func() *automock.RuntimeContextRepository
		Input                  model.RuntimeContextInput
		InputID                string
		ExpectedRuntimeContext *model.RuntimeContext
		ExpectedErrMessage     string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, tnt, id).Return(runtimeCtxModel, nil).Once()
				return repo
			},
			InputID:                id,
			ExpectedRuntimeContext: runtimeCtxModel,
			ExpectedErrMessage:     "",
		},
		{
			Name: "Returns error when runtime context retrieval failed",
			RepositoryFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, tnt, id).Return(nil, testErr).Once()
				return repo
			},
			InputID:                id,
			ExpectedRuntimeContext: runtimeCtxModel,
			ExpectedErrMessage:     testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := runtime_context.NewService(repo, nil, nil, nil)

			// when
			rtmCtx, err := svc.Get(ctx, testCase.InputID)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedRuntimeContext, rtmCtx)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}

	t.Run("Returns error on loading tenant", func(t *testing.T) {
		// given
		svc := runtime_context.NewService(nil, nil, nil, nil)
		// when
		_, err := svc.Get(context.TODO(), "id")
		// then
		require.Error(t, err)
		assert.EqualError(t, err, "while loading tenant from context: cannot read tenant from context")
	})
}

func TestService_Exist(t *testing.T) {
	tnt := "tenant"
	externalTnt := "external-tnt"
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)
	testError := errors.New("Test error")

	rtmCtxID := "id"

	testCases := []struct {
		Name                  string
		RepositoryFn          func() *automock.RuntimeContextRepository
		InputRuntimeContextID string
		ExpectedValue         bool
		ExpectedError         error
	}{
		{
			Name: "RuntimeContext exists",
			RepositoryFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("Exists", ctx, tnt, rtmCtxID).Return(true, nil)
				return repo
			},
			InputRuntimeContextID: rtmCtxID,
			ExpectedValue:         true,
			ExpectedError:         nil,
		},
		{
			Name: "RuntimeContext not exits",
			RepositoryFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("Exists", ctx, tnt, rtmCtxID).Return(false, nil)
				return repo
			},
			InputRuntimeContextID: rtmCtxID,
			ExpectedValue:         false,
			ExpectedError:         nil,
		},
		{
			Name: "Returns error",
			RepositoryFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("Exists", ctx, tnt, rtmCtxID).Return(false, testError)
				return repo
			},
			InputRuntimeContextID: rtmCtxID,
			ExpectedValue:         false,
			ExpectedError:         testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			rtmCtxRepo := testCase.RepositoryFn()
			svc := runtime_context.NewService(rtmCtxRepo, nil, nil, nil)

			// WHEN
			value, err := svc.Exist(ctx, testCase.InputRuntimeContextID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				require.Nil(t, err)
			}

			assert.Equal(t, testCase.ExpectedValue, value)
			rtmCtxRepo.AssertExpectations(t)
		})
	}
	t.Run("Returns error on loading tenant", func(t *testing.T) {
		// given
		svc := runtime_context.NewService(nil, nil, nil, nil)
		// when
		_, err := svc.Exist(context.TODO(), "id")
		// then
		require.Error(t, err)
		assert.EqualError(t, err, "while loading tenant from context: cannot read tenant from context")
	})
}

func TestService_List(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	runtimeID := "runtime_id"

	id := "foo"
	key := "key"
	val := "value"

	id2 := "bar"
	key2 := "key2"
	val2 := "value2"

	modelRuntimeContexts := []*model.RuntimeContext{
		{
			ID:        id,
			Key:       key,
			Value:     val,
			RuntimeID: runtimeID,
		},
		{
			ID:        id2,
			Key:       key2,
			Value:     val2,
			RuntimeID: runtimeID,
		},
	}
	runtimePage := &model.RuntimeContextPage{
		Data:       modelRuntimeContexts,
		TotalCount: len(modelRuntimeContexts),
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
	externalTnt := "external-tnt"

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.RuntimeContextRepository
		InputLabelFilters  []*labelfilter.LabelFilter
		InputPageSize      int
		InputCursor        string
		ExpectedResult     *model.RuntimeContextPage
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("List", ctx, runtimeID, tnt, filter, first, after).Return(runtimePage, nil).Once()
				return repo
			},
			InputLabelFilters:  filter,
			InputPageSize:      first,
			InputCursor:        after,
			ExpectedResult:     runtimePage,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when runtime context listing failed",
			RepositoryFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("List", ctx, runtimeID, tnt, filter, first, after).Return(nil, testErr).Once()
				return repo
			},
			InputLabelFilters:  filter,
			InputPageSize:      first,
			InputCursor:        after,
			ExpectedResult:     nil,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when pageSize is less than 1",
			RepositoryFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				return repo
			},
			InputLabelFilters:  filter,
			InputPageSize:      0,
			InputCursor:        after,
			ExpectedResult:     nil,
			ExpectedErrMessage: "page size must be between 1 and 200",
		},
		{
			Name: "Returns error when pageSize is bigger than 100",
			RepositoryFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				return repo
			},
			InputLabelFilters:  filter,
			InputPageSize:      201,
			InputCursor:        after,
			ExpectedResult:     nil,
			ExpectedErrMessage: "page size must be between 1 and 200",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := runtime_context.NewService(repo, nil, nil, nil)

			// when
			rtmCtx, err := svc.List(ctx, runtimeID, testCase.InputLabelFilters, testCase.InputPageSize, testCase.InputCursor)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedResult, rtmCtx)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}

	t.Run("Returns error on loading tenant", func(t *testing.T) {
		// given
		svc := runtime_context.NewService(nil, nil, nil, nil)
		// when
		_, err := svc.List(context.TODO(), "", nil, 1, "")
		// then
		require.Error(t, err)
		assert.EqualError(t, err, "while loading tenant from context: cannot read tenant from context")
	})
}

func TestService_ListLabel(t *testing.T) {
	// given
	tnt := "tenant"
	externalTnt := "external-tnt"

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	testErr := errors.New("Test error")

	runtimeCtxID := "foo"
	labelKey := "key"
	labelValue := []string{"value1"}

	label := &model.LabelInput{
		Key:        labelKey,
		Value:      labelValue,
		ObjectID:   runtimeCtxID,
		ObjectType: model.RuntimeContextLabelableObject,
	}

	modelLabel := &model.Label{
		ID:         "5d23d9d9-3d04-4fa9-95e6-d22e1ae62c11",
		Tenant:     tnt,
		Key:        labelKey,
		Value:      labelValue,
		ObjectID:   runtimeCtxID,
		ObjectType: model.RuntimeContextLabelableObject,
	}

	labels := map[string]*model.Label{"first": modelLabel, "second": modelLabel}
	testCases := []struct {
		Name                  string
		RepositoryFn          func() *automock.RuntimeContextRepository
		LabelRepositoryFn     func() *automock.LabelRepository
		InputRuntimeContextID string
		InputLabel            *model.LabelInput
		ExpectedOutput        map[string]*model.Label
		ExpectedErrMessage    string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("Exists", ctx, tnt, runtimeCtxID).Return(true, nil).Once()
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, tnt, model.RuntimeContextLabelableObject, runtimeCtxID).Return(labels, nil).Once()
				return repo
			},
			InputRuntimeContextID: runtimeCtxID,
			InputLabel:            label,
			ExpectedOutput:        labels,
			ExpectedErrMessage:    "",
		},
		{
			Name: "Returns error when labels receiving failed",
			RepositoryFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("Exists", ctx, tnt, runtimeCtxID).Return(true, nil).Once()

				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, tnt, model.RuntimeContextLabelableObject, runtimeCtxID).Return(nil, testErr).Once()
				return repo
			},
			InputRuntimeContextID: runtimeCtxID,
			InputLabel:            label,
			ExpectedOutput:        nil,
			ExpectedErrMessage:    testErr.Error(),
		},
		{
			Name: "Returns error when runtime context exists function failed",
			RepositoryFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("Exists", ctx, tnt, runtimeCtxID).Return(false, testErr).Once()

				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				return repo
			},
			InputRuntimeContextID: runtimeCtxID,
			InputLabel:            label,
			ExpectedErrMessage:    testErr.Error(),
		},
		{
			Name: "Returns error when runtime context does not exists",
			RepositoryFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("Exists", ctx, tnt, runtimeCtxID).Return(false, nil).Once()

				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				return repo
			},
			InputRuntimeContextID: runtimeCtxID,
			InputLabel:            label,
			ExpectedErrMessage:    fmt.Sprintf("runtime Context with ID %s doesn't exist", runtimeCtxID),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			labelRepo := testCase.LabelRepositoryFn()
			svc := runtime_context.NewService(repo, labelRepo, nil, nil)

			// when
			l, err := svc.ListLabels(ctx, testCase.InputRuntimeContextID)

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

	t.Run("Returns error on loading tenant", func(t *testing.T) {
		// given
		svc := runtime_context.NewService(nil, nil, nil, nil)
		// when
		_, err := svc.ListLabels(context.TODO(), "id")
		// then
		require.Error(t, err)
		assert.EqualError(t, err, "while loading tenant from context: cannot read tenant from context")
	})
}
