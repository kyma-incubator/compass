package eventing

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/normalizer"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/internal/domain/eventing/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	appNameNormalizer = &normalizer.DefaultNormalizator{}
	app               = fixApplicationModel("test-app")
)

func Test_CleanupAfterUnregisteringApplication(t *testing.T) {
	t.Run("Success when cleanup does not return errors", func(t *testing.T) {
		// GIVEN
		ctx := fixCtxWithTenant()
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("DeleteByKey", ctx, tenantID.String(), getDefaultEventingForAppLabelKey(applicationID)).Return(nil)

		svc := NewService(nil, nil, labelRepo)

		// WHEN
		eventingCfg, err := svc.CleanupAfterUnregisteringApplication(ctx, applicationID)

		// THEN

		require.NoError(t, err)
		require.NotNil(t, eventingCfg)
		require.Equal(t, "", eventingCfg.DefaultURL.String())
		mock.AssertExpectationsForObjects(t, labelRepo)
	})

	t.Run("Error when tenant not in context", func(t *testing.T) {
		// GIVEN
		svc := NewService(nil, nil, nil)

		// WHEN
		_, err := svc.CleanupAfterUnregisteringApplication(context.TODO(), uuid.Nil)

		// THEN
		require.Error(t, err)
		require.EqualError(t, err, "while loading tenant from context: cannot read tenant from context")
	})

	t.Run("Error when cleanup returns errors", func(t *testing.T) {
		// GIVEN
		expectedError := fmt.Sprintf(`while deleting Labels for Application with id %s: some-error`, applicationID)
		ctx := fixCtxWithTenant()
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("DeleteByKey", ctx, tenantID.String(), getDefaultEventingForAppLabelKey(applicationID)).Return(errors.New("some-error"))

		svc := NewService(nil, nil, labelRepo)

		// WHEN
		_, err := svc.CleanupAfterUnregisteringApplication(ctx, applicationID)

		// THEN

		require.Error(t, err)
		require.EqualError(t, err, expectedError)
		mock.AssertExpectationsForObjects(t, labelRepo)
	})
}

func Test_SetForApplication(t *testing.T) {
	appEventURL := fixAppEventURL(t, app.Name)
	appNormalizedEventURL := fixAppEventURL(t, appNameNormalizer.Normalize(app.Name))

	t.Run("Success when assigning new default runtime, when there was no previous one", func(t *testing.T) {
		// GIVEN
		ctx := fixCtxWithTenant()
		app := fixApplicationModel("test-app")
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixEmptyRuntimePage(), nil)
		runtimeRepo.On("GetByFiltersAndID", ctx, tenantID.String(), runtimeID.String(),
			fixLabelFilterForRuntimeScenarios()).Return(fixRuntimes()[0], nil)
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.ApplicationLabelableObject,
			applicationID.String(), model.ScenariosKey).Return(fixApplicationScenariosLabel(), nil)
		labelRepo.On("Upsert", ctx, mock.MatchedBy(fixMatcherDefaultEventingForAppLabel())).Return(nil)
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.RuntimeLabelableObject,
			runtimeID.String(), RuntimeEventingURLLabel).Return(fixRuntimeEventingURLLabel(), nil)
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.RuntimeLabelableObject,
			runtimeID.String(), isNormalizedLabel).Return(nil, apperrors.NewNotFoundError(resource.Runtime, runtimeID.String()))

		svc := NewService(appNameNormalizer, runtimeRepo, labelRepo)

		// WHEN
		eventingCfg, err := svc.SetForApplication(ctx, runtimeID, app)

		// THEN
		require.NoError(t, err)
		require.NotNil(t, eventingCfg)
		require.Equal(t, appNormalizedEventURL, eventingCfg.DefaultURL)
		mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo)
	})

	t.Run("Success when assigning new default runtime, when there is already one assigned", func(t *testing.T) {
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
		labelRepo.On("Upsert", ctx, mock.MatchedBy(fixMatcherDefaultEventingForAppLabel())).Return(nil)
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.RuntimeLabelableObject,
			runtimeID.String(), RuntimeEventingURLLabel).Return(fixRuntimeEventingURLLabel(), nil)
		labelRepo.On("Delete", ctx, tenantID.String(), model.RuntimeLabelableObject, runtimeID.String(),
			getDefaultEventingForAppLabelKey(applicationID)).Return(nil)
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.RuntimeLabelableObject,
			runtimeID.String(), isNormalizedLabel).Return(nil, apperrors.NewNotFoundError(resource.Runtime, runtimeID.String()))

		svc := NewService(appNameNormalizer, runtimeRepo, labelRepo)

		// WHEN
		eventingCfg, err := svc.SetForApplication(ctx, runtimeID, app)

		// THEN
		require.NoError(t, err)
		require.NotNil(t, eventingCfg)
		require.Equal(t, appNormalizedEventURL, eventingCfg.DefaultURL)
		mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo)
	})

	t.Run("Success when there is runtime labeled for application eventing and is labeled for normalization", func(t *testing.T) {
		// GIVEN
		ctx := fixCtxWithTenant()
		app := fixApplicationModel("test-app")
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixEmptyRuntimePage(), nil)
		runtimeRepo.On("GetByFiltersAndID", ctx, tenantID.String(), runtimeID.String(),
			fixLabelFilterForRuntimeScenarios()).Return(fixRuntimes()[0], nil)
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.ApplicationLabelableObject,
			applicationID.String(), model.ScenariosKey).Return(fixApplicationScenariosLabel(), nil)
		labelRepo.On("Upsert", ctx, mock.MatchedBy(fixMatcherDefaultEventingForAppLabel())).Return(nil)
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.RuntimeLabelableObject,
			runtimeID.String(), RuntimeEventingURLLabel).Return(fixRuntimeEventingURLLabel(), nil)
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.RuntimeLabelableObject,
			runtimeID.String(), isNormalizedLabel).Return(&model.Label{Value: "true"}, nil)

		svc := NewService(appNameNormalizer, runtimeRepo, labelRepo)

		// WHEN
		eventingCfg, err := svc.SetForApplication(ctx, runtimeID, app)

		// THEN
		require.NoError(t, err)
		require.NotNil(t, eventingCfg)
		require.Equal(t, appNormalizedEventURL, eventingCfg.DefaultURL)
		mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo)
	})

	t.Run("Success when there is runtime labeled for application eventing and is labeled not for normalization", func(t *testing.T) {
		// GIVEN
		ctx := fixCtxWithTenant()
		app := fixApplicationModel("test-app")
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixEmptyRuntimePage(), nil)
		runtimeRepo.On("GetByFiltersAndID", ctx, tenantID.String(), runtimeID.String(),
			fixLabelFilterForRuntimeScenarios()).Return(fixRuntimes()[0], nil)
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.ApplicationLabelableObject,
			applicationID.String(), model.ScenariosKey).Return(fixApplicationScenariosLabel(), nil)
		labelRepo.On("Upsert", ctx, mock.MatchedBy(fixMatcherDefaultEventingForAppLabel())).Return(nil)
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.RuntimeLabelableObject,
			runtimeID.String(), RuntimeEventingURLLabel).Return(fixRuntimeEventingURLLabel(), nil)
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.RuntimeLabelableObject,
			runtimeID.String(), isNormalizedLabel).Return(&model.Label{Value: "false"}, nil)

		svc := NewService(appNameNormalizer, runtimeRepo, labelRepo)

		// WHEN
		eventingCfg, err := svc.SetForApplication(ctx, runtimeID, app)

		// THEN
		require.NoError(t, err)
		require.NotNil(t, eventingCfg)
		require.Equal(t, appEventURL, eventingCfg.DefaultURL)
		mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo)
	})

	t.Run("Error when tenant not in context", func(t *testing.T) {
		// GIVEN
		svc := NewService(nil, nil, nil)

		// WHEN
		_, err := svc.SetForApplication(context.TODO(), uuid.Nil, model.Application{})

		// THEN
		require.Error(t, err)
		require.EqualError(t, err, "while loading tenant from context: cannot read tenant from context")
	})

	t.Run("Error when deleting existing default runtime, when getting current default runtime", func(t *testing.T) {
		// GIVEN
		expectedError := fmt.Sprintf(`while deleting default eventing for application: while getting default runtime for app eventing: while fetching runtimes with label [key=%s]: some-error`, getDefaultEventingForAppLabelKey(applicationID))
		ctx := fixCtxWithTenant()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(nil, errors.New("some-error"))
		labelRepo := &automock.LabelRepository{}

		svc := NewService(nil, runtimeRepo, labelRepo)

		// WHEN
		_, err := svc.SetForApplication(ctx, runtimeID, app)

		// THEN
		require.Error(t, err)
		require.EqualError(t, err, expectedError)
		mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo)
	})

	t.Run("Error when deleting existing default runtime, when getting current default runtime repository returns more than one runtime", func(t *testing.T) {
		// GIVEN
		expectedError := fmt.Sprintf(`while deleting default eventing for application: while getting default runtime for app eventing: got multpile runtimes labeled [key=%s] as default for eventing`, getDefaultEventingForAppLabelKey(applicationID))
		ctx := fixCtxWithTenant()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixRuntimePage(), nil)
		labelRepo := &automock.LabelRepository{}

		svc := NewService(nil, runtimeRepo, labelRepo)

		// WHEN
		_, err := svc.SetForApplication(ctx, runtimeID, app)

		// THEN
		require.Error(t, err)
		require.EqualError(t, err, expectedError)
		mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo)
	})

	t.Run("Error when deleting existing default runtime", func(t *testing.T) {
		// GIVEN
		expectedError := fmt.Sprintf(`while deleting default eventing for application: while deleting label: while deleting label [key=%s, runtimeID=%s]: some-error`, getDefaultEventingForAppLabelKey(applicationID), runtimeID)
		ctx := fixCtxWithTenant()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixRuntimePageWithOne(), nil)
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("Delete", ctx, tenantID.String(), model.RuntimeLabelableObject, runtimeID.String(),
			getDefaultEventingForAppLabelKey(applicationID)).Return(errors.New("some-error"))

		svc := NewService(nil, runtimeRepo, labelRepo)

		// WHEN
		_, err := svc.SetForApplication(ctx, runtimeID, app)

		// THEN
		require.Error(t, err)
		require.EqualError(t, err, expectedError)
		mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo)
	})

	t.Run("Error when getting scenarios returns error", func(t *testing.T) {
		// GIVEN
		expectedError := fmt.Sprintf(`while getting the runtime: while getting application scenarios: while getting the label [key=%s] for application [ID=%s]: some error`, model.ScenariosKey, applicationID)
		ctx := fixCtxWithTenant()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixRuntimePageWithOne(), nil)
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("Delete", ctx, tenantID.String(), model.RuntimeLabelableObject, runtimeID.String(),
			getDefaultEventingForAppLabelKey(applicationID)).Return(nil)
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.ApplicationLabelableObject,
			applicationID.String(), model.ScenariosKey).Return(nil, errors.New("some error"))

		svc := NewService(nil, runtimeRepo, labelRepo)

		// WHEN
		_, err := svc.SetForApplication(ctx, runtimeID, app)

		// THEN
		require.Error(t, err)
		require.EqualError(t, err, expectedError)
		mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo)
	})

	t.Run("Error when there are no scenarios assigned to application", func(t *testing.T) {
		// GIVEN
		expectedError := fmt.Sprintf(`while getting the runtime: Internal Server Error: application does not belong to scenarios`)
		ctx := fixCtxWithTenant()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixRuntimePageWithOne(), nil)
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.ApplicationLabelableObject,
			applicationID.String(), model.ScenariosKey).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
		labelRepo.On("Delete", ctx, tenantID.String(), model.RuntimeLabelableObject, runtimeID.String(),
			getDefaultEventingForAppLabelKey(applicationID)).Return(nil)

		svc := NewService(nil, runtimeRepo, labelRepo)

		// WHEN
		_, err := svc.SetForApplication(ctx, runtimeID, app)

		// THEN
		require.Error(t, err)
		require.EqualError(t, err, expectedError)
		mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo)
	})

	t.Run("Error when given runtime does not belong to the application scenarios", func(t *testing.T) {
		// GIVEN
		expectedError := fmt.Sprintf(`does not find the given runtime [ID=%s] assigned to the application scenarios`, runtimeID)
		ctx := fixCtxWithTenant()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixRuntimePageWithOne(), nil)
		runtimeRepo.On("GetByFiltersAndID", ctx, tenantID.String(), runtimeID.String(),
			fixLabelFilterForRuntimeScenarios()).Return(nil, apperrors.NewNotFoundError(resource.Runtime, ""))
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.ApplicationLabelableObject,
			applicationID.String(), model.ScenariosKey).Return(fixApplicationScenariosLabel(), nil)
		labelRepo.On("Delete", ctx, tenantID.String(), model.RuntimeLabelableObject, runtimeID.String(),
			getDefaultEventingForAppLabelKey(applicationID)).Return(nil)

		svc := NewService(nil, runtimeRepo, labelRepo)

		// WHEN
		_, err := svc.SetForApplication(ctx, runtimeID, app)

		// THEN
		require.Error(t, err)
		require.EqualError(t, err, expectedError)
		mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo)
	})

	t.Run("Error when getting new runtime to assign as default", func(t *testing.T) {
		// GIVEN
		expectedError := fmt.Sprintf(`while getting the runtime: while getting the runtime [ID=%s] with scenarios with filter: some-error`, runtimeID)
		ctx := fixCtxWithTenant()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixRuntimePageWithOne(), nil)
		runtimeRepo.On("GetByFiltersAndID", ctx, tenantID.String(), runtimeID.String(),
			fixLabelFilterForRuntimeScenarios()).Return(nil, errors.New("some-error"))
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.ApplicationLabelableObject,
			applicationID.String(), model.ScenariosKey).Return(fixApplicationScenariosLabel(), nil)
		labelRepo.On("Delete", ctx, tenantID.String(), model.RuntimeLabelableObject, runtimeID.String(),
			getDefaultEventingForAppLabelKey(applicationID)).Return(nil)

		svc := NewService(nil, runtimeRepo, labelRepo)

		// WHEN
		_, err := svc.SetForApplication(ctx, runtimeID, app)

		// THEN
		require.Error(t, err)
		require.EqualError(t, err, expectedError)
		mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo)
	})

	t.Run("Error when assigning new default runtime", func(t *testing.T) {
		// GIVEN
		expectedError := fmt.Sprintf(`while setting the runtime as default for eventing for application: while labeling the runtime [ID=%s] as default for eventing for application [ID=%s]: some-error`, runtimeID, applicationID)
		ctx := fixCtxWithTenant()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixRuntimePageWithOne(), nil)
		runtimeRepo.On("GetByFiltersAndID", ctx, tenantID.String(), runtimeID.String(),
			fixLabelFilterForRuntimeScenarios()).Return(fixRuntimes()[0], nil)
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.ApplicationLabelableObject,
			applicationID.String(), model.ScenariosKey).Return(fixApplicationScenariosLabel(), nil)
		labelRepo.On("Upsert", ctx, mock.MatchedBy(fixMatcherDefaultEventingForAppLabel())).Return(errors.New("some-error"))
		labelRepo.On("Delete", ctx, tenantID.String(), model.RuntimeLabelableObject, runtimeID.String(),
			getDefaultEventingForAppLabelKey(applicationID)).Return(nil)

		svc := NewService(nil, runtimeRepo, labelRepo)

		// WHEN
		_, err := svc.SetForApplication(ctx, runtimeID, app)

		// THEN
		require.Error(t, err)
		require.EqualError(t, err, expectedError)
		mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo)
	})

	t.Run("Error when getting eventing configuration for a given runtime", func(t *testing.T) {
		// GIVEN
		expectedError := fmt.Sprintf(`while fetching eventing configuration for runtime: while getting the label [key=%s] for runtime [ID=%s]: some-error`, RuntimeEventingURLLabel, runtimeID)
		ctx := fixCtxWithTenant()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixRuntimePageWithOne(), nil)
		runtimeRepo.On("GetByFiltersAndID", ctx, tenantID.String(), runtimeID.String(),
			fixLabelFilterForRuntimeScenarios()).Return(fixRuntimes()[0], nil)
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.ApplicationLabelableObject,
			applicationID.String(), model.ScenariosKey).Return(fixApplicationScenariosLabel(), nil)
		labelRepo.On("Upsert", ctx, mock.MatchedBy(fixMatcherDefaultEventingForAppLabel())).Return(nil)
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.RuntimeLabelableObject,
			runtimeID.String(), RuntimeEventingURLLabel).Return(nil, errors.New("some-error"))
		labelRepo.On("Delete", ctx, tenantID.String(), model.RuntimeLabelableObject, runtimeID.String(),
			getDefaultEventingForAppLabelKey(applicationID)).Return(nil)

		svc := NewService(nil, runtimeRepo, labelRepo)

		// WHEN
		_, err := svc.SetForApplication(ctx, runtimeID, app)

		// THEN
		require.Error(t, err)
		require.EqualError(t, err, expectedError)
		mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo)
	})

	t.Run("Error when getting runtime normalization label", func(t *testing.T) {
		// GIVEN
		expectedError := fmt.Sprintf(`while determining whether application name should be normalized in runtime eventing URL: while getting the label [key=%s] for runtime [ID=%s]: some error`, isNormalizedLabel, runtimeID)
		ctx := fixCtxWithTenant()
		app := fixApplicationModel("test-app")
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixEmptyRuntimePage(), nil)
		runtimeRepo.On("GetByFiltersAndID", ctx, tenantID.String(), runtimeID.String(),
			fixLabelFilterForRuntimeScenarios()).Return(fixRuntimes()[0], nil)
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.ApplicationLabelableObject,
			applicationID.String(), model.ScenariosKey).Return(fixApplicationScenariosLabel(), nil)
		labelRepo.On("Upsert", ctx, mock.MatchedBy(fixMatcherDefaultEventingForAppLabel())).Return(nil)
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.RuntimeLabelableObject,
			runtimeID.String(), RuntimeEventingURLLabel).Return(fixRuntimeEventingURLLabel(), nil)
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.RuntimeLabelableObject,
			runtimeID.String(), isNormalizedLabel).Return(nil, errors.New("some error"))

		svc := NewService(appNameNormalizer, runtimeRepo, labelRepo)

		// WHEN
		_, err := svc.SetForApplication(ctx, runtimeID, app)

		// THEN
		require.Error(t, err)
		require.EqualError(t, err, expectedError)
		mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo)
	})

}

func Test_UnsetForApplication(t *testing.T) {
	appEventURL := fixAppEventURL(t, app.Name)
	appNormalizedEventURL := fixAppEventURL(t, appNameNormalizer.Normalize(app.Name))

	t.Run("Success when there is no default runtime assigned for eventing", func(t *testing.T) {
		// GIVEN
		ctx := fixCtxWithTenant()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixEmptyRuntimePage(), nil)

		svc := NewService(nil, runtimeRepo, nil)

		// WHEN
		eventingCfg, err := svc.UnsetForApplication(ctx, app)

		// THEN
		require.NoError(t, err)
		require.NotNil(t, eventingCfg)
		require.Equal(t, EmptyEventingURL, eventingCfg.DefaultURL.String())
		mock.AssertExpectationsForObjects(t, runtimeRepo)
	})

	t.Run("Success when there is default runtime assigned for eventing", func(t *testing.T) {
		// GIVEN
		ctx := fixCtxWithTenant()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixRuntimePageWithOne(), nil)
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.RuntimeLabelableObject,
			runtimeID.String(), RuntimeEventingURLLabel).Return(fixRuntimeEventingURLLabel(), nil)
		labelRepo.On("Delete", ctx, tenantID.String(), model.RuntimeLabelableObject, runtimeID.String(),
			getDefaultEventingForAppLabelKey(applicationID)).Return(nil)
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.RuntimeLabelableObject,
			runtimeID.String(), isNormalizedLabel).Return(nil, apperrors.NewNotFoundError(resource.Runtime, runtimeID.String()))

		svc := NewService(appNameNormalizer, runtimeRepo, labelRepo)

		// WHEN
		eventingCfg, err := svc.UnsetForApplication(ctx, app)

		// THEN
		require.NoError(t, err)
		require.NotNil(t, eventingCfg)
		require.Equal(t, appNormalizedEventURL, eventingCfg.DefaultURL)
		mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo)
	})

	t.Run("Success when there is runtime labeled for application eventing and is labeled for normalization", func(t *testing.T) {
		// GIVEN
		ctx := fixCtxWithTenant()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixRuntimePageWithOne(), nil)
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.RuntimeLabelableObject,
			runtimeID.String(), RuntimeEventingURLLabel).Return(fixRuntimeEventingURLLabel(), nil)
		labelRepo.On("Delete", ctx, tenantID.String(), model.RuntimeLabelableObject, runtimeID.String(),
			getDefaultEventingForAppLabelKey(applicationID)).Return(nil)
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.RuntimeLabelableObject,
			runtimeID.String(), isNormalizedLabel).Return(&model.Label{Value: "true"}, nil)

		svc := NewService(appNameNormalizer, runtimeRepo, labelRepo)

		// WHEN
		eventingCfg, err := svc.UnsetForApplication(ctx, app)

		// THEN
		require.NoError(t, err)
		require.NotNil(t, eventingCfg)
		require.Equal(t, appNormalizedEventURL, eventingCfg.DefaultURL)
		mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo)
	})

	t.Run("Success when there is runtime labeled for application eventing and is labeled not for normalization", func(t *testing.T) {
		// GIVEN
		ctx := fixCtxWithTenant()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixRuntimePageWithOne(), nil)
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.RuntimeLabelableObject,
			runtimeID.String(), RuntimeEventingURLLabel).Return(fixRuntimeEventingURLLabel(), nil)
		labelRepo.On("Delete", ctx, tenantID.String(), model.RuntimeLabelableObject, runtimeID.String(),
			getDefaultEventingForAppLabelKey(applicationID)).Return(nil)
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.RuntimeLabelableObject,
			runtimeID.String(), isNormalizedLabel).Return(&model.Label{Value: "false"}, nil)

		svc := NewService(appNameNormalizer, runtimeRepo, labelRepo)

		// WHEN
		eventingCfg, err := svc.UnsetForApplication(ctx, app)

		// THEN
		require.NoError(t, err)
		require.NotNil(t, eventingCfg)
		require.Equal(t, appEventURL, eventingCfg.DefaultURL)
		mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo)
	})

	t.Run("Error when tenant not in context", func(t *testing.T) {
		// GIVEN
		svc := NewService(nil, nil, nil)

		// WHEN
		_, err := svc.UnsetForApplication(context.TODO(), app)

		// THEN
		require.Error(t, err)
		require.EqualError(t, err, "while loading tenant from context: cannot read tenant from context")
	})

	t.Run("Error when getting default runtime assigned for eventing returns error", func(t *testing.T) {
		// GIVEN
		expectedError := fmt.Sprintf(`while deleting default eventing for application: while getting default runtime for app eventing: while fetching runtimes with label [key=%s]: some-error`, getDefaultEventingForAppLabelKey(applicationID))
		ctx := fixCtxWithTenant()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(nil, errors.New("some-error"))

		svc := NewService(nil, runtimeRepo, nil)

		// WHEN
		_, err := svc.UnsetForApplication(ctx, app)

		// THEN
		require.Error(t, err)
		require.EqualError(t, err, expectedError)
		mock.AssertExpectationsForObjects(t, runtimeRepo)
	})

	t.Run("Error when getting default runtime assigned for eventing returns more than one element", func(t *testing.T) {
		// GIVEN
		expectedError := fmt.Sprintf(`while deleting default eventing for application: while getting default runtime for app eventing: got multpile runtimes labeled [key=%s] as default for eventing`, getDefaultEventingForAppLabelKey(applicationID))
		ctx := fixCtxWithTenant()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixRuntimePage(), nil)

		svc := NewService(nil, runtimeRepo, nil)

		// WHEN
		_, err := svc.UnsetForApplication(ctx, app)

		// THEN
		require.Error(t, err)
		require.EqualError(t, err, expectedError)
		mock.AssertExpectationsForObjects(t, runtimeRepo)
	})

	t.Run("Error when deleting default app eventing label", func(t *testing.T) {
		// GIVEN
		expectedError := fmt.Sprintf(`while deleting default eventing for application: while deleting label: while deleting label [key=%s, runtimeID=%s]: some-error`, getDefaultEventingForAppLabelKey(applicationID), runtimeID)
		ctx := fixCtxWithTenant()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixRuntimePageWithOne(), nil)
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("Delete", ctx, tenantID.String(), model.RuntimeLabelableObject, runtimeID.String(),
			getDefaultEventingForAppLabelKey(applicationID)).Return(errors.New("some-error"))

		svc := NewService(nil, runtimeRepo, labelRepo)

		// WHEN
		_, err := svc.UnsetForApplication(ctx, app)

		// THEN
		require.Error(t, err)
		require.EqualError(t, err, expectedError)
		mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo)
	})

	t.Run("Error when getting eventing configuration for a deleted default runtime", func(t *testing.T) {
		// GIVEN
		expectedError := fmt.Sprintf(`while fetching eventing configuration for runtime: while getting the label [key=%s] for runtime [ID=%s]: some error`, RuntimeEventingURLLabel, runtimeID)
		ctx := fixCtxWithTenant()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixRuntimePageWithOne(), nil)
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.RuntimeLabelableObject,
			runtimeID.String(), RuntimeEventingURLLabel).Return(nil, errors.New("some error"))
		labelRepo.On("Delete", ctx, tenantID.String(), model.RuntimeLabelableObject, runtimeID.String(),
			getDefaultEventingForAppLabelKey(applicationID)).Return(nil)

		svc := NewService(nil, runtimeRepo, labelRepo)

		// WHEN
		_, err := svc.UnsetForApplication(ctx, app)

		// THEN
		require.Error(t, err)
		require.EqualError(t, err, expectedError)
		mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo)
	})

	t.Run("Success when there is runtime labeled for application eventing and is labeled not for normalization", func(t *testing.T) {
		// GIVEN
		expectedError := fmt.Sprintf(`while determining whether application name should be normalized in runtime eventing URL: while getting the label [key=%s] for runtime [ID=%s]: some error`, isNormalizedLabel, runtimeID)
		ctx := fixCtxWithTenant()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixRuntimePageWithOne(), nil)
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.RuntimeLabelableObject,
			runtimeID.String(), RuntimeEventingURLLabel).Return(fixRuntimeEventingURLLabel(), nil)
		labelRepo.On("Delete", ctx, tenantID.String(), model.RuntimeLabelableObject, runtimeID.String(),
			getDefaultEventingForAppLabelKey(applicationID)).Return(nil)
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.RuntimeLabelableObject,
			runtimeID.String(), isNormalizedLabel).Return(nil, errors.New("some error"))

		svc := NewService(appNameNormalizer, runtimeRepo, labelRepo)

		// WHEN
		_, err := svc.UnsetForApplication(ctx, app)

		// THEN
		require.Error(t, err)
		require.EqualError(t, err, expectedError)
		mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo)
	})
}

func Test_GetForApplication(t *testing.T) {
	appEventURL := fixAppEventURL(t, app.Name)
	appNormalizedEventURL := fixAppEventURL(t, appNameNormalizer.Normalize(app.Name))

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
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.RuntimeLabelableObject,
			runtimeID.String(), isNormalizedLabel).Return(nil, apperrors.NewNotFoundError(resource.Runtime, runtimeID.String()))

		svc := NewService(appNameNormalizer, runtimeRepo, labelRepo)

		// WHEN
		eventingCfg, err := svc.GetForApplication(ctx, app)

		// THEN
		require.NoError(t, err)
		require.NotNil(t, eventingCfg)
		require.Equal(t, appNormalizedEventURL, eventingCfg.DefaultURL)
		mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo)
	})

	t.Run("Success when labeling oldest runtime for application eventing", func(t *testing.T) {
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
		labelRepo.On("Upsert", ctx, mock.MatchedBy(fixMatcherDefaultEventingForAppLabel())).Return(nil)
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.RuntimeLabelableObject,
			runtimeID.String(), RuntimeEventingURLLabel).Return(fixRuntimeEventingURLLabel(), nil)
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.RuntimeLabelableObject,
			runtimeID.String(), isNormalizedLabel).Return(nil, apperrors.NewNotFoundError(resource.Runtime, runtimeID.String()))
		svc := NewService(appNameNormalizer, runtimeRepo, labelRepo)

		// WHEN
		eventingCfg, err := svc.GetForApplication(ctx, app)

		// THEN
		require.NoError(t, err)
		require.NotNil(t, eventingCfg)
		require.Equal(t, appNormalizedEventURL, eventingCfg.DefaultURL)
		mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo)
	})

	t.Run("Success when there is runtime labeled for application eventing and is labeled for normalization", func(t *testing.T) {
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
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.RuntimeLabelableObject,
			runtimeID.String(), isNormalizedLabel).Return(&model.Label{Value: "true"}, nil)

		svc := NewService(appNameNormalizer, runtimeRepo, labelRepo)

		// WHEN
		eventingCfg, err := svc.GetForApplication(ctx, app)

		// THEN
		require.NoError(t, err)
		require.NotNil(t, eventingCfg)
		require.Equal(t, appNormalizedEventURL, eventingCfg.DefaultURL)
		mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo)
	})

	t.Run("Success when there is runtime labeled for application eventing and is labeled not for normalization", func(t *testing.T) {
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
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.RuntimeLabelableObject,
			runtimeID.String(), isNormalizedLabel).Return(&model.Label{Value: "false"}, nil)

		svc := NewService(appNameNormalizer, runtimeRepo, labelRepo)

		// WHEN
		eventingCfg, err := svc.GetForApplication(ctx, app)

		// THEN
		require.NoError(t, err)
		require.NotNil(t, eventingCfg)
		require.Equal(t, appEventURL, eventingCfg.DefaultURL)
		mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo)
	})

	t.Run("Success when there is no oldest runtime for application eventing", func(t *testing.T) {
		// GIVEN
		ctx := fixCtxWithTenant()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixEmptyRuntimePage(), nil)
		runtimeRepo.On("GetOldestForFilters", ctx, tenantID.String(), fixLabelFilterForRuntimeScenarios()).
			Return(nil, apperrors.NewNotFoundError(resource.Runtime, ""))
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.ApplicationLabelableObject,
			applicationID.String(), model.ScenariosKey).Return(fixApplicationScenariosLabel(), nil)

		svc := NewService(nil, runtimeRepo, labelRepo)

		// WHEN
		eventingCfg, err := svc.GetForApplication(ctx, app)

		// THEN
		require.NoError(t, err)
		require.NotNil(t, eventingCfg)
		require.Equal(t, "", eventingCfg.DefaultURL.String())
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
			applicationID.String(), model.ScenariosKey).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
		svc := NewService(nil, runtimeRepo, labelRepo)

		// WHEN
		eventingCfg, err := svc.GetForApplication(ctx, app)

		// THEN
		require.NoError(t, err)
		require.NotNil(t, eventingCfg)
		require.Equal(t, "", eventingCfg.DefaultURL.String())
		mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo)
	})

	t.Run("Success when current default runtime no longer belongs to the application scenarios", func(t *testing.T) {
		// GIVEN
		ctx := fixCtxWithTenant()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixRuntimePageWithOne(), nil)
		runtimeRepo.On("GetByFiltersAndID", ctx, tenantID.String(), runtimeID.String(),
			fixLabelFilterForRuntimeScenarios()).Return(nil, apperrors.NewNotFoundError(resource.Runtime, ""))
		runtimeRepo.On("GetOldestForFilters", ctx, tenantID.String(), fixLabelFilterForRuntimeScenarios()).
			Return(nil, apperrors.NewNotFoundError(resource.Runtime, ""))
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.ApplicationLabelableObject,
			applicationID.String(), model.ScenariosKey).Return(fixApplicationScenariosLabel(), nil)
		labelRepo.On("Delete", ctx, tenantID.String(), model.RuntimeLabelableObject, runtimeID.String(),
			getDefaultEventingForAppLabelKey(applicationID)).Return(nil)

		svc := NewService(nil, runtimeRepo, labelRepo)

		// WHEN
		eventingCfg, err := svc.GetForApplication(ctx, app)

		// THEN
		require.NoError(t, err)
		require.NotNil(t, eventingCfg)
		require.Equal(t, "", eventingCfg.DefaultURL.String())
		mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo)
	})

	t.Run("Error when tenant not in context", func(t *testing.T) {
		// GIVEN
		svc := NewService(nil, nil, nil)

		// WHEN
		_, err := svc.GetForApplication(context.TODO(), app)

		// THEN
		require.Error(t, err)
		require.EqualError(t, err, "while loading tenant from context: cannot read tenant from context")
	})

	t.Run("Error when labeling oldest runtime for application eventing returns error", func(t *testing.T) {
		// GIVEN
		expectedError := fmt.Sprintf(`while setting the runtime as default for eventing for application: while labeling the runtime [ID=%s] as default for eventing for application [ID=%s]: some error`, runtimeID, applicationID)
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
		svc := NewService(nil, runtimeRepo, labelRepo)

		// WHEN
		_, err := svc.GetForApplication(ctx, app)

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
		svc := NewService(nil, runtimeRepo, labelRepo)

		// WHEN
		_, err := svc.GetForApplication(ctx, app)

		// THEN
		require.Error(t, err)
		require.EqualError(t, err, expectedError)
		mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo)
	})

	t.Run("Error when getting the oldest runtime for application eventing returns error on converting scenarios label to slice of strings", func(t *testing.T) {
		// GIVEN
		expectedError := fmt.Sprintf(`while getting the oldest runtime for scenarios: while getting application scenarios: while converting label [key=%s] value to a slice of strings: Internal Server Error: cannot convert label value to slice of strings`, model.ScenariosKey)
		ctx := fixCtxWithTenant()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixEmptyRuntimePage(), nil)
		labelRepo := &automock.LabelRepository{}
		scenariosLabel := fixApplicationScenariosLabel()
		scenariosLabel.Value = "abc"
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.ApplicationLabelableObject,
			applicationID.String(), model.ScenariosKey).Return(scenariosLabel, nil)
		svc := NewService(nil, runtimeRepo, labelRepo)

		// WHEN
		_, err := svc.GetForApplication(ctx, app)

		// THEN
		require.Error(t, err)
		require.EqualError(t, err, expectedError)
		mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo)
	})

	t.Run("Error when getting the oldest runtime for application eventing returns error when getting scenarios label", func(t *testing.T) {
		// GIVEN
		expectedError := fmt.Sprintf(`while getting the oldest runtime for scenarios: while getting application scenarios: while getting the label [key=%s] for application [ID=%s]: some error`, model.ScenariosKey, applicationID)
		ctx := fixCtxWithTenant()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixEmptyRuntimePage(), nil)
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.ApplicationLabelableObject,
			applicationID.String(), model.ScenariosKey).Return(nil, errors.New("some error"))
		svc := NewService(nil, runtimeRepo, labelRepo)

		// WHEN
		_, err := svc.GetForApplication(ctx, app)

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
		svc := NewService(nil, runtimeRepo, nil)

		// WHEN
		_, err := svc.GetForApplication(ctx, app)

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

		svc := NewService(nil, runtimeRepo, labelRepo)

		// WHEN
		_, err := svc.GetForApplication(ctx, app)

		// THEN
		require.Error(t, err)
		require.EqualError(t, err, expectedError)
		mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo)
	})

	t.Run("Error when ensuring the scenarios, getting scenarios returns error", func(t *testing.T) {
		// GIVEN
		expectedError := fmt.Sprintf(`while ensuring the scenarios assigned to the runtime and application: while verifing whether runtime [ID=%s] belongs to the application scenarios: while getting application scenarios: while getting the label [key=%s] for application [ID=%s]: some error`, runtimeID, model.ScenariosKey, applicationID)
		ctx := fixCtxWithTenant()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixRuntimePageWithOne(), nil)
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.ApplicationLabelableObject,
			applicationID.String(), model.ScenariosKey).Return(nil, errors.New("some error"))

		svc := NewService(nil, runtimeRepo, labelRepo)

		// WHEN
		_, err := svc.GetForApplication(ctx, app)

		// THEN
		require.Error(t, err)
		require.EqualError(t, err, expectedError)
		mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo)
	})

	t.Run("Error when ensuring the scenarios, there are no scenarios assigned to application", func(t *testing.T) {
		// GIVEN
		expectedError := fmt.Sprintf(`while ensuring the scenarios assigned to the runtime and application: while verifing whether runtime [ID=%s] belongs to the application scenarios: Internal Server Error: application does not belong to scenarios`, runtimeID)
		ctx := fixCtxWithTenant()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixRuntimePageWithOne(), nil)
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.ApplicationLabelableObject,
			applicationID.String(), model.ScenariosKey).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))

		svc := NewService(nil, runtimeRepo, labelRepo)

		// WHEN
		_, err := svc.GetForApplication(ctx, app)

		// THEN
		require.Error(t, err)
		require.EqualError(t, err, expectedError)
		mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo)
	})

	t.Run("Error when verifing whether given runtime belongs to the application scenarios - repository returns error", func(t *testing.T) {
		// GIVEN
		expectedError := fmt.Sprintf(`while ensuring the scenarios assigned to the runtime and application: while verifing whether runtime [ID=%s] belongs to the application scenarios: while getting the runtime [ID=%s] with scenarios with filter: some-error`, runtimeID, runtimeID)
		ctx := fixCtxWithTenant()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixRuntimePageWithOne(), nil)
		runtimeRepo.On("GetByFiltersAndID", ctx, tenantID.String(), runtimeID.String(),
			fixLabelFilterForRuntimeScenarios()).Return(nil, errors.New("some-error"))
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.ApplicationLabelableObject,
			applicationID.String(), model.ScenariosKey).Return(fixApplicationScenariosLabel(), nil)

		svc := NewService(nil, runtimeRepo, labelRepo)

		// WHEN
		_, err := svc.GetForApplication(ctx, app)

		// THEN
		require.Error(t, err)
		require.EqualError(t, err, expectedError)
		mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo)
	})

	t.Run("Error when deleting label from the given runtime because it does not belong to application scenarios - repository returns error", func(t *testing.T) {
		// GIVEN
		expectedError := fmt.Sprintf(`while ensuring the scenarios assigned to the runtime and application: when deleting current default runtime for the application because of scenarios mismatch: while deleting label [key=%s, runtimeID=%s]: some-error`, getDefaultEventingForAppLabelKey(applicationID), runtimeID)
		ctx := fixCtxWithTenant()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixRuntimePageWithOne(), nil)
		runtimeRepo.On("GetByFiltersAndID", ctx, tenantID.String(), runtimeID.String(),
			fixLabelFilterForRuntimeScenarios()).Return(nil, apperrors.NewNotFoundError(resource.Runtime, ""))
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.ApplicationLabelableObject,
			applicationID.String(), model.ScenariosKey).Return(fixApplicationScenariosLabel(), nil)
		labelRepo.On("Delete", ctx, tenantID.String(), model.RuntimeLabelableObject, runtimeID.String(),
			getDefaultEventingForAppLabelKey(applicationID)).Return(errors.New("some-error"))

		svc := NewService(nil, runtimeRepo, labelRepo)

		// WHEN
		_, err := svc.GetForApplication(ctx, app)

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

		svc := NewService(nil, runtimeRepo, labelRepo)

		// WHEN
		_, err := svc.GetForApplication(ctx, app)

		// THEN
		require.Error(t, err)
		require.EqualError(t, err, expectedError)
		mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo)
	})

	t.Run("Error when getting runtime normalization label", func(t *testing.T) {
		// GIVEN
		expectedError := fmt.Sprintf(`while determining whether application name should be normalized in runtime eventing URL: while getting the label [key=%s] for runtime [ID=%s]: some error`, isNormalizedLabel, runtimeID)
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
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.RuntimeLabelableObject,
			runtimeID.String(), isNormalizedLabel).Return(nil, errors.New("some error"))

		svc := NewService(appNameNormalizer, runtimeRepo, labelRepo)

		// WHEN
		_, err := svc.GetForApplication(ctx, app)

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
			Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
		expectedEventingCfg := fixRuntimeEventngCfgWithEmptyURL(t)
		eventingSvc := NewService(nil, nil, labelRepo)

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
		expectedEventingCfg := fixRuntimeEventngCfgWithURL(t, runtimeEventURL)
		eventingSvc := NewService(nil, nil, labelRepo)

		// WHEN
		eventingCfg, err := eventingSvc.GetForRuntime(ctx, runtimeID)

		// THEN
		require.NoError(t, err)
		require.Equal(t, expectedEventingCfg, eventingCfg)
		mock.AssertExpectationsForObjects(t, labelRepo)
	})

	t.Run("Error when tenant not in context", func(t *testing.T) {
		// GIVEN
		svc := NewService(nil, nil, nil)

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
		eventingSvc := NewService(nil, nil, labelRepo)

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
		eventingSvc := NewService(nil, nil, labelRepo)

		// WHEN
		_, err := eventingSvc.GetForRuntime(ctx, runtimeID)

		// THEN
		require.Error(t, err)
		require.EqualError(t, err, `unable to cast label [key=`+RuntimeEventingURLLabel+`, runtimeID=`+runtimeID.String()+`] value as a string`)
		mock.AssertExpectationsForObjects(t, labelRepo)
	})
}
