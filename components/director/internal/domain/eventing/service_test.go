package eventing

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/domain/eventing/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func Test_DeleteDefaultForApplication(t *testing.T) {
	t.Run("Success when deletion does not return errors", func(t *testing.T) {
		// GIVEN
		ctx := fixCtxWithTenant()
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("DeleteByKey", ctx, tenantID.String(), getDefaultEventingForAppLabelKey(applicationID)).Return(nil)

		svc := NewService(nil, labelRepo)

		// WHEN
		eventingCfg, err := svc.DeleteDefaultForApplication(ctx, applicationID)

		// THEN

		require.NoError(t, err)
		require.NotNil(t, eventingCfg)
		require.Equal(t, EmptyEventingURL, eventingCfg.DefaultURL)
		mock.AssertExpectationsForObjects(t, labelRepo)
	})

	t.Run("Error when tenant not in context", func(t *testing.T) {
		// GIVEN
		svc := NewService(nil, nil)

		// WHEN
		_, err := svc.GetForApplication(context.TODO(), uuid.Nil)

		// THEN
		require.Error(t, err)
		require.EqualError(t, err, "while loading tenant from context: cannot read tenant from context")
	})

	t.Run("Error when deletion returns errors", func(t *testing.T) {
		// GIVEN
		expectedError := fmt.Sprintf(`while deleting labels [key=%s]: some-error`, getDefaultEventingForAppLabelKey(applicationID))
		ctx := fixCtxWithTenant()
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("DeleteByKey", ctx, tenantID.String(), getDefaultEventingForAppLabelKey(applicationID)).Return(errors.New("some-error"))

		svc := NewService(nil, labelRepo)

		// WHEN
		_, err := svc.DeleteDefaultForApplication(ctx, applicationID)

		// THEN

		require.Error(t, err)
		require.EqualError(t, err, expectedError)
		mock.AssertExpectationsForObjects(t, labelRepo)
	})
}

func Test_GetForApplication(t *testing.T) {
	t.Run("Success when there is default runtime labeled for application eventing", func(t *testing.T) {
		// GIVEN
		ctx := fixCtxWithTenant()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixRuntimePageWithOne(), nil)
		runtimeRepo.On("GetByFiltersAndID", ctx, tenantID.String(), runtimeID.String(),
			fixLabelFilterForRuntimeScenarios()).Return(fixRuntimes()[0], nil)
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.ApplicationLabelableObject,
			applicationID.String(), model.ScenariosKey).Return(fixApplicationScenariosLabel(), nil)
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.RuntimeLabelableObject,
			runtimeID.String(), RuntimeEventingURLLabel).Return(fixRuntimeEventingURLLabel(), nil)

		svc := NewService(runtimeRepo, labelRepo)

		// WHEN
		eventingCfg, err := svc.GetForApplication(ctx, applicationID)

		// THEN
		require.NoError(t, err)
		require.NotNil(t, eventingCfg)
		require.Equal(t, dummyEventingURL, eventingCfg.DefaultURL)
		mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo)
	})

	t.Run("Success when labelling oldest runtime for application eventing", func(t *testing.T) {
		// GIVEN
		ctx := fixCtxWithTenant()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixEmptyRuntimePage(), nil)
		runtimeRepo.On("GetOldestForFilters", ctx, tenantID.String(), fixLabelFilterForRuntimeScenarios()).
			Return(fixRuntimes()[0], nil)
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.ApplicationLabelableObject,
			applicationID.String(), model.ScenariosKey).Return(fixApplicationScenariosLabel(), nil)
		labelRepo.On("Upsert", ctx,
			mock.MatchedBy(fixMatcherDefaultEventingForAppLabel())).
			Return(nil)
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.RuntimeLabelableObject,
			runtimeID.String(), RuntimeEventingURLLabel).Return(fixRuntimeEventingURLLabel(), nil)
		svc := NewService(runtimeRepo, labelRepo)

		// WHEN
		eventingCfg, err := svc.GetForApplication(ctx, applicationID)

		// THEN
		require.NoError(t, err)
		require.NotNil(t, eventingCfg)
		require.Equal(t, dummyEventingURL, eventingCfg.DefaultURL)
		mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo)
	})

	t.Run("Success when there is no oldest runtime for application eventing", func(t *testing.T) {
		// GIVEN
		ctx := fixCtxWithTenant()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixEmptyRuntimePage(), nil)
		runtimeRepo.On("GetOldestForFilters", ctx, tenantID.String(), fixLabelFilterForRuntimeScenarios()).
			Return(nil, apperrors.NewNotFoundError(""))
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.ApplicationLabelableObject,
			applicationID.String(), model.ScenariosKey).Return(fixApplicationScenariosLabel(), nil)

		svc := NewService(runtimeRepo, labelRepo)

		// WHEN
		eventingCfg, err := svc.GetForApplication(ctx, applicationID)

		// THEN
		require.NoError(t, err)
		require.NotNil(t, eventingCfg)
		require.Equal(t, EmptyEventingURL, eventingCfg.DefaultURL)
		mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo)
	})

	t.Run("Success when there is no oldest runtime for application eventing (scenarios label does not exist)", func(t *testing.T) {
		// GIVEN
		ctx := fixCtxWithTenant()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixEmptyRuntimePage(), nil)
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.ApplicationLabelableObject,
			applicationID.String(), model.ScenariosKey).Return(nil, apperrors.NewNotFoundError(""))
		svc := NewService(runtimeRepo, labelRepo)

		// WHEN
		eventingCfg, err := svc.GetForApplication(ctx, applicationID)

		// THEN
		require.NoError(t, err)
		require.NotNil(t, eventingCfg)
		require.Equal(t, EmptyEventingURL, eventingCfg.DefaultURL)
		mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo)
	})

	t.Run("Success when current default runtime no longer belongs to the application scenarios", func(t *testing.T) {
		// GIVEN
		ctx := fixCtxWithTenant()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixRuntimePageWithOne(), nil)
		runtimeRepo.On("GetByFiltersAndID", ctx, tenantID.String(), runtimeID.String(),
			fixLabelFilterForRuntimeScenarios()).Return(nil, apperrors.NewNotFoundError(""))
		runtimeRepo.On("GetOldestForFilters", ctx, tenantID.String(), fixLabelFilterForRuntimeScenarios()).
			Return(nil, apperrors.NewNotFoundError(""))
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.ApplicationLabelableObject,
			applicationID.String(), model.ScenariosKey).Return(fixApplicationScenariosLabel(), nil)
		labelRepo.On("Delete", ctx, tenantID.String(), model.RuntimeLabelableObject, runtimeID.String(),
			getDefaultEventingForAppLabelKey(applicationID)).Return(nil)

		svc := NewService(runtimeRepo, labelRepo)

		// WHEN
		eventingCfg, err := svc.GetForApplication(ctx, applicationID)

		// THEN
		require.NoError(t, err)
		require.NotNil(t, eventingCfg)
		require.Equal(t, EmptyEventingURL, eventingCfg.DefaultURL)
		mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo)
	})

	t.Run("Error when tenant not in context", func(t *testing.T) {
		// GIVEN
		svc := NewService(nil, nil)

		// WHEN
		_, err := svc.GetForApplication(context.TODO(), uuid.Nil)

		// THEN
		require.Error(t, err)
		require.EqualError(t, err, "while loading tenant from context: cannot read tenant from context")
	})

	t.Run("Error when labeling oldest runtime for application eventing returns error", func(t *testing.T) {
		// GIVEN
		expectedError := fmt.Sprintf(`while setting the runtime as default for eveting for application: while labeling the runtime [ID=%s] as default for eventing for application [ID=%s]: some error`, runtimeID, applicationID)
		ctx := fixCtxWithTenant()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixEmptyRuntimePage(), nil)
		runtimeRepo.On("GetOldestForFilters", ctx, tenantID.String(), fixLabelFilterForRuntimeScenarios()).
			Return(fixRuntimes()[0], nil)
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.ApplicationLabelableObject,
			applicationID.String(), model.ScenariosKey).Return(fixApplicationScenariosLabel(), nil)
		labelRepo.On("Upsert", ctx,
			mock.MatchedBy(fixMatcherDefaultEventingForAppLabel())).
			Return(errors.New("some error"))
		svc := NewService(runtimeRepo, labelRepo)

		// WHEN
		_, err := svc.GetForApplication(ctx, applicationID)

		// THEN
		require.Error(t, err)
		require.EqualError(t, err, expectedError)
		mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo)
	})

	t.Run("Error when getting the oldest runtime for application eventing returns error", func(t *testing.T) {
		// GIVEN
		expectedError := fmt.Sprintf(`while getting the oldest runtime for scenarios: while getting the oldest runtime for application [ID=%s] scenarios with filter: some error`, applicationID)
		ctx := fixCtxWithTenant()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixEmptyRuntimePage(), nil)
		runtimeRepo.On("GetOldestForFilters", ctx, tenantID.String(), fixLabelFilterForRuntimeScenarios()).
			Return(nil, errors.New("some error"))
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.ApplicationLabelableObject,
			applicationID.String(), model.ScenariosKey).Return(fixApplicationScenariosLabel(), nil)
		svc := NewService(runtimeRepo, labelRepo)

		// WHEN
		_, err := svc.GetForApplication(ctx, applicationID)

		// THEN
		require.Error(t, err)
		require.EqualError(t, err, expectedError)
		mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo)
	})

	t.Run("Error when getting the oldest runtime for application eventing returns error on converting scenarios label to slice of strings", func(t *testing.T) {
		// GIVEN
		expectedError := fmt.Sprintf(`while getting the oldest runtime for scenarios: while converting label [key=%s] value to a slice of strings: cannot convert label value to slice of strings`, model.ScenariosKey)
		ctx := fixCtxWithTenant()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixEmptyRuntimePage(), nil)
		labelRepo := &automock.LabelRepository{}
		scenariosLabel := fixApplicationScenariosLabel()
		scenariosLabel.Value = "abc"
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.ApplicationLabelableObject,
			applicationID.String(), model.ScenariosKey).Return(scenariosLabel, nil)
		svc := NewService(runtimeRepo, labelRepo)

		// WHEN
		_, err := svc.GetForApplication(ctx, applicationID)

		// THEN
		require.Error(t, err)
		require.EqualError(t, err, expectedError)
		mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo)
	})

	t.Run("Error when getting the oldest runtime for application eventing returns error when getting scenarios label", func(t *testing.T) {
		// GIVEN
		expectedError := fmt.Sprintf(`while getting the oldest runtime for scenarios: while getting the label [key=%s] for application [ID=%s]: some error`, model.ScenariosKey, applicationID)
		ctx := fixCtxWithTenant()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixEmptyRuntimePage(), nil)
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.ApplicationLabelableObject,
			applicationID.String(), model.ScenariosKey).Return(nil, errors.New("some error"))
		svc := NewService(runtimeRepo, labelRepo)

		// WHEN
		_, err := svc.GetForApplication(ctx, applicationID)

		// THEN
		require.Error(t, err)
		require.EqualError(t, err, expectedError)
		mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo)
	})

	t.Run("Error when getting default runtime labeled for application eventing returns error", func(t *testing.T) {
		// GIVEN
		expectedError := fmt.Sprintf(`while getting default runtime for app eventing: while fetching runtimes with label [key=%s]: some error`, getDefaultEventingForAppLabelKey(applicationID))
		ctx := fixCtxWithTenant()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(nil, errors.New("some error"))
		svc := NewService(runtimeRepo, nil)

		// WHEN
		_, err := svc.GetForApplication(ctx, applicationID)

		// THEN
		require.Error(t, err)
		require.EqualError(t, err, expectedError)
		mock.AssertExpectationsForObjects(t, runtimeRepo)
	})

	t.Run("Error when getting default runtime labeled for application eventing returns error with multiple runtimes labeled", func(t *testing.T) {
		// GIVEN
		expectedError := fmt.Sprintf(`while getting default runtime for app eventing: got multpile runtimes labeled [key=%s] as default for eventing`, getDefaultEventingForAppLabelKey(applicationID))
		ctx := fixCtxWithTenant()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixRuntimePage(), nil)
		labelRepo := &automock.LabelRepository{}

		svc := NewService(runtimeRepo, labelRepo)

		// WHEN
		_, err := svc.GetForApplication(ctx, applicationID)

		// THEN
		require.Error(t, err)
		require.EqualError(t, err, expectedError)
		mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo)
	})

	t.Run("Error when getting scenarios returns error", func(t *testing.T) {
		// GIVEN
		expectedError := fmt.Sprintf(`while getting default runtime for app eventing: while verifing whether runtime [ID=%s] belongs to the application scenarios: while getting application scenarios: while getting the label [key=%s] for application [ID=%s]: some error`, runtimeID, model.ScenariosKey, applicationID)
		ctx := fixCtxWithTenant()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixRuntimePageWithOne(), nil)
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.ApplicationLabelableObject,
			applicationID.String(), model.ScenariosKey).Return(nil, errors.New("some error"))

		svc := NewService(runtimeRepo, labelRepo)

		// WHEN
		_, err := svc.GetForApplication(ctx, applicationID)

		// THEN
		require.Error(t, err)
		require.EqualError(t, err, expectedError)
		mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo)
	})

	t.Run("Error when there are no scenarios assigned to application", func(t *testing.T) {
		// GIVEN
		expectedError := fmt.Sprintf(`while getting default runtime for app eventing: while verifing whether runtime [ID=%s] belongs to the application scenarios: application does not belong to scenarios`, runtimeID)
		ctx := fixCtxWithTenant()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixRuntimePageWithOne(), nil)
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.ApplicationLabelableObject,
			applicationID.String(), model.ScenariosKey).Return(nil, apperrors.NewNotFoundError(""))

		svc := NewService(runtimeRepo, labelRepo)

		// WHEN
		_, err := svc.GetForApplication(ctx, applicationID)

		// THEN
		require.Error(t, err)
		require.EqualError(t, err, expectedError)
		mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo)
	})

	t.Run("Error when verifing whether given runtime belongs to application scenarios - repository returns error", func(t *testing.T) {
		// GIVEN
		expectedError := fmt.Sprintf(`while getting default runtime for app eventing: while verifing whether runtime [ID=%s] belongs to the application scenarios: while getting the runtime [ID=%s] with scenarios with filter: some-error`, runtimeID, runtimeID)
		ctx := fixCtxWithTenant()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixRuntimePageWithOne(), nil)
		runtimeRepo.On("GetByFiltersAndID", ctx, tenantID.String(), runtimeID.String(),
			fixLabelFilterForRuntimeScenarios()).Return(nil, errors.New("some-error"))
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.ApplicationLabelableObject,
			applicationID.String(), model.ScenariosKey).Return(fixApplicationScenariosLabel(), nil)

		svc := NewService(runtimeRepo, labelRepo)

		// WHEN
		_, err := svc.GetForApplication(ctx, applicationID)

		// THEN
		require.Error(t, err)
		require.EqualError(t, err, expectedError)
		mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo)
	})

	t.Run("Error when deleting label from the given runtime because it does not belong to application scenarios - repository returns error", func(t *testing.T) {
		// GIVEN
		expectedError := fmt.Sprintf(`while getting default runtime for app eventing: when deleting current default runtime for the application because of scenarios mismatch: while deleting label [key=%s, runtimeID=%s]: some-error`, getDefaultEventingForAppLabelKey(applicationID), runtimeID)
		ctx := fixCtxWithTenant()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixRuntimePageWithOne(), nil)
		runtimeRepo.On("GetByFiltersAndID", ctx, tenantID.String(), runtimeID.String(),
			fixLabelFilterForRuntimeScenarios()).Return(nil, apperrors.NewNotFoundError(""))
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.ApplicationLabelableObject,
			applicationID.String(), model.ScenariosKey).Return(fixApplicationScenariosLabel(), nil)
		labelRepo.On("Delete", ctx, tenantID.String(), model.RuntimeLabelableObject, runtimeID.String(),
			getDefaultEventingForAppLabelKey(applicationID)).Return(errors.New("some-error"))

		svc := NewService(runtimeRepo, labelRepo)

		// WHEN
		_, err := svc.GetForApplication(ctx, applicationID)

		// THEN
		require.Error(t, err)
		require.EqualError(t, err, expectedError)
		mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo)
	})

	t.Run("Error when getting eventing configuration for a given runtime", func(t *testing.T) {
		// GIVEN
		expectedError := fmt.Sprintf(`while fetching eventing configuration for runtime: while getting the label [key=%s] for runtime [ID=%s]: some error`, RuntimeEventingURLLabel, runtimeID)
		ctx := fixCtxWithTenant()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixRuntimePageWithOne(), nil)
		runtimeRepo.On("GetByFiltersAndID", ctx, tenantID.String(), runtimeID.String(),
			fixLabelFilterForRuntimeScenarios()).Return(fixRuntimes()[0], nil)
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.ApplicationLabelableObject,
			applicationID.String(), model.ScenariosKey).Return(fixApplicationScenariosLabel(), nil)
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.RuntimeLabelableObject,
			runtimeID.String(), RuntimeEventingURLLabel).Return(nil, errors.New("some error"))

		svc := NewService(runtimeRepo, labelRepo)

		// WHEN
		_, err := svc.GetForApplication(ctx, applicationID)

		// THEN
		require.Error(t, err)
		require.EqualError(t, err, expectedError)
		mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo)
	})
}

func Test_GetForRuntime(t *testing.T) {
	t.Run("Success when label repository returns NotFoundError", func(t *testing.T) {
		// GIVEN
		ctx := fixCtxWithTenant()
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.RuntimeLabelableObject, runtimeID.String(), RuntimeEventingURLLabel).
			Return(nil, apperrors.NewNotFoundError(""))
		expectedEventingCfg := fixRuntimeEventngCfgWithEmptyURL()
		eventingSvc := NewService(nil, labelRepo)

		// WHEN
		eventingCfg, err := eventingSvc.GetForRuntime(ctx, runtimeID)

		// THEN
		require.NoError(t, err)
		require.Equal(t, expectedEventingCfg, eventingCfg)
		mock.AssertExpectationsForObjects(t, labelRepo)
	})

	t.Run("Success when label repository returns lable instance with URL string value", func(t *testing.T) {
		// GIVEN
		ctx := fixCtxWithTenant()
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.RuntimeLabelableObject, runtimeID.String(), RuntimeEventingURLLabel).
			Return(fixRuntimeEventingURLLabel(), nil)
		expectedEventingCfg := fixRuntimeEventngCfgWithURL(dummyEventingURL)
		eventingSvc := NewService(nil, labelRepo)

		// WHEN
		eventingCfg, err := eventingSvc.GetForRuntime(ctx, runtimeID)

		// THEN
		require.NoError(t, err)
		require.Equal(t, expectedEventingCfg, eventingCfg)
		mock.AssertExpectationsForObjects(t, labelRepo)
	})

	t.Run("Error when tenant not in context", func(t *testing.T) {
		// GIVEN
		svc := NewService(nil, nil)

		// WHEN
		_, err := svc.GetForRuntime(context.TODO(), uuid.Nil)

		// THEN
		require.Error(t, err)
		require.EqualError(t, err, "while loading tenant from context: cannot read tenant from context")
	})

	t.Run("Error when label repository returns error", func(t *testing.T) {
		// GIVEN
		ctx := fixCtxWithTenant()
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.RuntimeLabelableObject, runtimeID.String(), RuntimeEventingURLLabel).
			Return(nil, errors.New("some error"))
		eventingSvc := NewService(nil, labelRepo)

		// WHEN
		_, err := eventingSvc.GetForRuntime(ctx, runtimeID)

		// THEN
		require.Error(t, err)
		require.EqualError(t, err, `while getting the label [key=`+RuntimeEventingURLLabel+`] for runtime [ID=`+runtimeID.String()+`]: some error`)
		mock.AssertExpectationsForObjects(t, labelRepo)
	})

	t.Run("Error when label value cannot be converted to a string", func(t *testing.T) {
		// GIVEN
		ctx := fixCtxWithTenant()

		label := fixRuntimeEventingURLLabel()
		label.Value = byte(1)

		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.RuntimeLabelableObject, runtimeID.String(), RuntimeEventingURLLabel).
			Return(label, nil)
		eventingSvc := NewService(nil, labelRepo)

		// WHEN
		_, err := eventingSvc.GetForRuntime(ctx, runtimeID)

		// THEN
		require.Error(t, err)
		require.EqualError(t, err, `unable to cast label [key=`+RuntimeEventingURLLabel+`, runtimeID=`+runtimeID.String()+`] value as a string`)
		mock.AssertExpectationsForObjects(t, labelRepo)
	})
}
