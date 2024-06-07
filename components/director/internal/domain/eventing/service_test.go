package eventing

import (
	"context"
	"fmt"
	"net/url"
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

		svc := NewService(nil, nil, labelRepo, nil)

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
		svc := NewService(nil, nil, nil, nil)

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

		svc := NewService(nil, nil, labelRepo, nil)

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
	formations := []*model.Formation{{Name: formationName}}
	formationNames := []string{formationName}

	t.Run("Success when assigning new default runtime, when there was no previous one", func(t *testing.T) {
		// GIVEN
		ctx := fixCtxWithTenant()
		app := fixApplicationModel("test-app")
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), []string{}, fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixEmptyRuntimePage(), nil).Once()
		runtimeRepo.On("GetByID", ctx, tenantID.String(), runtimeID.String()).Return(fixRuntimes()[0], nil).Once()
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("Upsert", ctx, tenantID.String(), mock.MatchedBy(fixMatcherDefaultEventingForAppLabel())).Return(nil).Once()
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.RuntimeLabelableObject,
			runtimeID.String(), RuntimeEventingURLLabel).Return(fixRuntimeEventingURLLabel(), nil).Once()
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.RuntimeLabelableObject,
			runtimeID.String(), isNormalizedLabel).Return(nil, apperrors.NewNotFoundError(resource.Runtime, runtimeID.String())).Once()
		formationSvc := &automock.FormationService{}
		formationSvc.On("ListFormationsForObject", ctx, app.ID).Return(formations, nil).Once()
		formationSvc.On("ListObjectIDsOfTypeForFormations", ctx, tenantID.String(), formationNames, model.FormationAssignmentTypeRuntime).Return([]string{runtimeID.String()}, nil).Once()

		svc := NewService(appNameNormalizer, runtimeRepo, labelRepo, formationSvc)

		// WHEN
		eventingCfg, err := svc.SetForApplication(ctx, runtimeID, app)

		// THEN
		require.NoError(t, err)
		require.NotNil(t, eventingCfg)
		require.Equal(t, appNormalizedEventURL, eventingCfg.DefaultURL)
		mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo, formationSvc)
	})

	t.Run("Success when assigning new default runtime, when there is already one assigned", func(t *testing.T) {
		// GIVEN
		ctx := fixCtxWithTenant()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), []string{}, fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixRuntimePageWithOne(), nil).Once()
		runtimeRepo.On("GetByID", ctx, tenantID.String(), runtimeID.String()).Return(fixRuntimes()[0], nil).Once()
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("Upsert", ctx, tenantID.String(), mock.MatchedBy(fixMatcherDefaultEventingForAppLabel())).Return(nil).Once()
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.RuntimeLabelableObject,
			runtimeID.String(), RuntimeEventingURLLabel).Return(fixRuntimeEventingURLLabel(), nil).Once()
		labelRepo.On("Delete", ctx, tenantID.String(), model.RuntimeLabelableObject, runtimeID.String(),
			getDefaultEventingForAppLabelKey(applicationID)).Return(nil).Once()
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.RuntimeLabelableObject,
			runtimeID.String(), isNormalizedLabel).Return(nil, apperrors.NewNotFoundError(resource.Runtime, runtimeID.String())).Once()
		formationSvc := &automock.FormationService{}
		formationSvc.On("ListFormationsForObject", ctx, app.ID).Return(formations, nil).Once()
		formationSvc.On("ListObjectIDsOfTypeForFormations", ctx, tenantID.String(), formationNames, model.FormationAssignmentTypeRuntime).Return([]string{runtimeID.String()}, nil).Once()
		svc := NewService(appNameNormalizer, runtimeRepo, labelRepo, formationSvc)

		// WHEN
		eventingCfg, err := svc.SetForApplication(ctx, runtimeID, app)

		// THEN
		require.NoError(t, err)
		require.NotNil(t, eventingCfg)
		require.Equal(t, appNormalizedEventURL, eventingCfg.DefaultURL)
	mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo, formationSvc)
	})

	t.Run("Success when there is runtime labeled for application eventing and is labeled for normalization", func(t *testing.T) {
		// GIVEN
		ctx := fixCtxWithTenant()
		app := fixApplicationModel("test-app")
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), []string{}, fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixEmptyRuntimePage(), nil)
		runtimeRepo.On("GetByID", ctx, tenantID.String(), runtimeID.String()).Return(fixRuntimes()[0], nil).Once()
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("Upsert", ctx, tenantID.String(), mock.MatchedBy(fixMatcherDefaultEventingForAppLabel())).Return(nil).Once()
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.RuntimeLabelableObject,
			runtimeID.String(), RuntimeEventingURLLabel).Return(fixRuntimeEventingURLLabel(), nil).Once()
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.RuntimeLabelableObject,
			runtimeID.String(), isNormalizedLabel).Return(&model.Label{Value: "true"}, nil).Once()
		formationSvc := &automock.FormationService{}
		formationSvc.On("ListFormationsForObject", ctx, app.ID).Return(formations, nil).Once()
		formationSvc.On("ListObjectIDsOfTypeForFormations", ctx, tenantID.String(), formationNames, model.FormationAssignmentTypeRuntime).Return([]string{runtimeID.String()}, nil).Once()
		svc := NewService(appNameNormalizer, runtimeRepo, labelRepo, formationSvc)

		// WHEN
		eventingCfg, err := svc.SetForApplication(ctx, runtimeID, app)

		// THEN
		require.NoError(t, err)
		require.NotNil(t, eventingCfg)
		require.Equal(t, appNormalizedEventURL, eventingCfg.DefaultURL)
	mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo, formationSvc)
	})

	t.Run("Success when there is runtime labeled for application eventing and is labeled not for normalization", func(t *testing.T) {
		// GIVEN
		ctx := fixCtxWithTenant()
		app := fixApplicationModel("test-app")
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), []string{}, fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixEmptyRuntimePage(), nil).Once()
		runtimeRepo.On("GetByID", ctx, tenantID.String(), runtimeID.String()).Return(fixRuntimes()[0], nil).Once()
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("Upsert", ctx, tenantID.String(), mock.MatchedBy(fixMatcherDefaultEventingForAppLabel())).Return(nil).Once()
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.RuntimeLabelableObject,
			runtimeID.String(), RuntimeEventingURLLabel).Return(fixRuntimeEventingURLLabel(), nil).Once()
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.RuntimeLabelableObject,
			runtimeID.String(), isNormalizedLabel).Return(&model.Label{Value: "false"}, nil).Once()
		formationSvc := &automock.FormationService{}
		formationSvc.On("ListFormationsForObject", ctx, app.ID).Return(formations, nil).Once()
		formationSvc.On("ListObjectIDsOfTypeForFormations", ctx, tenantID.String(), formationNames, model.FormationAssignmentTypeRuntime).Return([]string{runtimeID.String()}, nil).Once()
		svc := NewService(appNameNormalizer, runtimeRepo, labelRepo, formationSvc)

		// WHEN
		eventingCfg, err := svc.SetForApplication(ctx, runtimeID, app)

		// THEN
		require.NoError(t, err)
		require.NotNil(t, eventingCfg)
		require.Equal(t, appEventURL, eventingCfg.DefaultURL)
	mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo, formationSvc)
	})

	t.Run("Error when tenant not in context", func(t *testing.T) {
		// GIVEN
		svc := NewService(nil, nil, nil, nil)

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
		runtimeRepo.On("List", ctx, tenantID.String(), []string{}, fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(nil, errors.New("some-error"))

		svc := NewService(nil, runtimeRepo, nil, nil)

		// WHEN
		_, err := svc.SetForApplication(ctx, runtimeID, app)

		// THEN
		require.Error(t, err)
		require.EqualError(t, err, expectedError)
	mock.AssertExpectationsForObjects(t, runtimeRepo)
	})

	t.Run("Error when deleting existing default runtime, when getting current default runtime repository returns more than one runtime", func(t *testing.T) {
		// GIVEN
		expectedError := fmt.Sprintf(`while deleting default eventing for application: while getting default runtime for app eventing: got multpile runtimes labeled [key=%s] as default for eventing`, getDefaultEventingForAppLabelKey(applicationID))
		ctx := fixCtxWithTenant()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), []string{}, fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixRuntimePage(), nil)

		svc := NewService(nil, runtimeRepo, nil, nil)

		// WHEN
		_, err := svc.SetForApplication(ctx, runtimeID, app)

		// THEN
		require.Error(t, err)
		require.EqualError(t, err, expectedError)
		mock.AssertExpectationsForObjects(t, runtimeRepo)
	})

	t.Run("Error when deleting existing default runtime", func(t *testing.T) {
		// GIVEN
		expectedError := fmt.Sprintf(`while deleting default eventing for application: while deleting label: while deleting label [key=%s, runtimeID=%s]: some-error`, getDefaultEventingForAppLabelKey(applicationID), runtimeID)
		ctx := fixCtxWithTenant()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), []string{}, fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixRuntimePageWithOne(), nil)
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("Delete", ctx, tenantID.String(), model.RuntimeLabelableObject, runtimeID.String(),
			getDefaultEventingForAppLabelKey(applicationID)).Return(errors.New("some-error"))

		svc := NewService(nil, runtimeRepo, labelRepo, nil)

		// WHEN
		_, err := svc.SetForApplication(ctx, runtimeID, app)

		// THEN
		require.Error(t, err)
		require.EqualError(t, err, expectedError)
		mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo)
	})

	t.Run("Error when listing runtime IDs for application formation", func(t *testing.T) {
		// GIVEN
		expectedError := fmt.Sprintf(`while getting the runtime: while getting runtimes IDs in formation with application: %s: while getting Runtimes for Formations: [%s]: some error`, applicationID, formationName)
		ctx := fixCtxWithTenant()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), []string{}, fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixRuntimePageWithOne(), nil)
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("Delete", ctx, tenantID.String(), model.RuntimeLabelableObject, runtimeID.String(),
			getDefaultEventingForAppLabelKey(applicationID)).Return(nil)
		formationSvc := &automock.FormationService{}
		formationSvc.On("ListFormationsForObject", ctx, app.ID).Return(formations, nil).Once()
		formationSvc.On("ListObjectIDsOfTypeForFormations", ctx, tenantID.String(), formationNames, model.FormationAssignmentTypeRuntime).Return(nil, errors.New("some error")).Once()
		svc := NewService(appNameNormalizer, runtimeRepo, labelRepo, formationSvc)

		// WHEN
		_, err := svc.SetForApplication(ctx, runtimeID, app)

		// THEN
		require.Error(t, err)
		require.EqualError(t, err, expectedError)
	mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo, formationSvc)
	})

	t.Run("Error when listing formations for application", func(t *testing.T) {
		// GIVEN
		expectedError := fmt.Sprintf(`while getting the runtime: while getting runtimes IDs in formation with application: %s: while getting Formations for Application with ID: %s: some error`, applicationID, applicationID)
		ctx := fixCtxWithTenant()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), []string{}, fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixRuntimePageWithOne(), nil)
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("Delete", ctx, tenantID.String(), model.RuntimeLabelableObject, runtimeID.String(),
			getDefaultEventingForAppLabelKey(applicationID)).Return(nil)
		formationSvc := &automock.FormationService{}
		formationSvc.On("ListFormationsForObject", ctx, app.ID).Return(nil, errors.New("some error")).Once()
		svc := NewService(appNameNormalizer, runtimeRepo, labelRepo, formationSvc)

		// WHEN
		_, err := svc.SetForApplication(ctx, runtimeID, app)

		// THEN
		require.Error(t, err)
		require.EqualError(t, err, expectedError)
	mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo, formationSvc)
	})

	t.Run("Error when given runtime does not belong to the application scenarios", func(t *testing.T) {
		// GIVEN
		expectedError := fmt.Sprintf(`does not find the given runtime [ID=%s] assigned to the application scenarios`, runtimeID)
		ctx := fixCtxWithTenant()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), []string{}, fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixRuntimePageWithOne(), nil)
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("Delete", ctx, tenantID.String(), model.RuntimeLabelableObject, runtimeID.String(),
			getDefaultEventingForAppLabelKey(applicationID)).Return(nil)
		formationSvc := &automock.FormationService{}
		formationSvc.On("ListFormationsForObject", ctx, app.ID).Return(formations, nil).Once()
		formationSvc.On("ListObjectIDsOfTypeForFormations", ctx, tenantID.String(), formationNames, model.FormationAssignmentTypeRuntime).Return(nil, nil).Once()
		svc := NewService(appNameNormalizer, runtimeRepo, labelRepo, formationSvc)

		// WHEN
		_, err := svc.SetForApplication(ctx, runtimeID, app)

		// THEN
		require.Error(t, err)
		require.EqualError(t, err, expectedError)
	mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo, formationSvc)
	})

	t.Run("Error when getting new runtime to assign as default", func(t *testing.T) {
		// GIVEN
		expectedError := fmt.Sprintf(`while getting the runtime: while getting the runtime [ID=%s] with scenarios with filter: some-error`, runtimeID)
		ctx := fixCtxWithTenant()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), []string{}, fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixRuntimePageWithOne(), nil)
		runtimeRepo.On("GetByID", ctx, tenantID.String(), runtimeID.String()).Return(nil, errors.New("some-error"))
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("Delete", ctx, tenantID.String(), model.RuntimeLabelableObject, runtimeID.String(),
			getDefaultEventingForAppLabelKey(applicationID)).Return(nil)
		formationSvc := &automock.FormationService{}
		formationSvc.On("ListFormationsForObject", ctx, app.ID).Return(formations, nil).Once()
		formationSvc.On("ListObjectIDsOfTypeForFormations", ctx, tenantID.String(), formationNames, model.FormationAssignmentTypeRuntime).Return([]string{runtimeID.String()}, nil).Once()
		svc := NewService(appNameNormalizer, runtimeRepo, labelRepo, formationSvc)

		// WHEN
		_, err := svc.SetForApplication(ctx, runtimeID, app)

		// THEN
		require.Error(t, err)
		require.EqualError(t, err, expectedError)
	mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo, formationSvc)
	})

	t.Run("Error when assigning new default runtime", func(t *testing.T) {
		// GIVEN
		expectedError := fmt.Sprintf(`while setting the runtime as default for eventing for application: while labeling the runtime [ID=%s] as default for eventing for application [ID=%s]: some-error`, runtimeID, applicationID)
		ctx := fixCtxWithTenant()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), []string{}, fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixRuntimePageWithOne(), nil)
		runtimeRepo.On("GetByID", ctx, tenantID.String(), runtimeID.String()).Return(fixRuntimes()[0], nil)
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("Upsert", ctx, tenantID.String(), mock.MatchedBy(fixMatcherDefaultEventingForAppLabel())).Return(errors.New("some-error"))
		labelRepo.On("Delete", ctx, tenantID.String(), model.RuntimeLabelableObject, runtimeID.String(),
			getDefaultEventingForAppLabelKey(applicationID)).Return(nil)
		formationSvc := &automock.FormationService{}
		formationSvc.On("ListFormationsForObject", ctx, app.ID).Return(formations, nil).Once()
		formationSvc.On("ListObjectIDsOfTypeForFormations", ctx, tenantID.String(), formationNames, model.FormationAssignmentTypeRuntime).Return([]string{runtimeID.String()}, nil).Once()
		svc := NewService(appNameNormalizer, runtimeRepo, labelRepo, formationSvc)
		// WHEN
		_, err := svc.SetForApplication(ctx, runtimeID, app)

		// THEN
		require.Error(t, err)
		require.EqualError(t, err, expectedError)
	mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo, formationSvc)
	})

	t.Run("Error when getting eventing configuration for a given runtime", func(t *testing.T) {
		// GIVEN
		expectedError := fmt.Sprintf(`while fetching eventing configuration for runtime: while getting the label [key=%s] for runtime [ID=%s]: some-error`, RuntimeEventingURLLabel, runtimeID)
		ctx := fixCtxWithTenant()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), []string{}, fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixRuntimePageWithOne(), nil)
		runtimeRepo.On("GetByID", ctx, tenantID.String(), runtimeID.String()).Return(fixRuntimes()[0], nil)
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("Upsert", ctx, tenantID.String(), mock.MatchedBy(fixMatcherDefaultEventingForAppLabel())).Return(nil)
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.RuntimeLabelableObject,
			runtimeID.String(), RuntimeEventingURLLabel).Return(nil, errors.New("some-error"))
		labelRepo.On("Delete", ctx, tenantID.String(), model.RuntimeLabelableObject, runtimeID.String(),
			getDefaultEventingForAppLabelKey(applicationID)).Return(nil)
		formationSvc := &automock.FormationService{}
		formationSvc.On("ListFormationsForObject", ctx, app.ID).Return(formations, nil).Once()
		formationSvc.On("ListObjectIDsOfTypeForFormations", ctx, tenantID.String(), formationNames, model.FormationAssignmentTypeRuntime).Return([]string{runtimeID.String()}, nil).Once()
		svc := NewService(appNameNormalizer, runtimeRepo, labelRepo, formationSvc)

		// WHEN
		_, err := svc.SetForApplication(ctx, runtimeID, app)

		// THEN
		require.Error(t, err)
		require.EqualError(t, err, expectedError)
	mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo, formationSvc)
	})

	t.Run("Error when getting runtime normalization label", func(t *testing.T) {
		// GIVEN
		expectedError := fmt.Sprintf(`while determining whether application name should be normalized in runtime eventing URL: while getting the label [key=%s] for runtime [ID=%s]: some error`, isNormalizedLabel, runtimeID)
		ctx := fixCtxWithTenant()
		app := fixApplicationModel("test-app")
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), []string{}, fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixEmptyRuntimePage(), nil)
		runtimeRepo.On("GetByID", ctx, tenantID.String(), runtimeID.String()).Return(fixRuntimes()[0], nil)
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("Upsert", ctx, tenantID.String(), mock.MatchedBy(fixMatcherDefaultEventingForAppLabel())).Return(nil)
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.RuntimeLabelableObject,
			runtimeID.String(), RuntimeEventingURLLabel).Return(fixRuntimeEventingURLLabel(), nil)
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.RuntimeLabelableObject,
			runtimeID.String(), isNormalizedLabel).Return(nil, errors.New("some error"))
		formationSvc := &automock.FormationService{}
		formationSvc.On("ListFormationsForObject", ctx, app.ID).Return(formations, nil).Once()
		formationSvc.On("ListObjectIDsOfTypeForFormations", ctx, tenantID.String(), formationNames, model.FormationAssignmentTypeRuntime).Return([]string{runtimeID.String()}, nil).Once()
		svc := NewService(appNameNormalizer, runtimeRepo, labelRepo, formationSvc)

		// WHEN
		_, err := svc.SetForApplication(ctx, runtimeID, app)

		// THEN
		require.Error(t, err)
		require.EqualError(t, err, expectedError)
	mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo, formationSvc)
	})
}

func Test_UnsetForApplication(t *testing.T) {
	appEventURL := fixAppEventURL(t, app.Name)
	appNormalizedEventURL := fixAppEventURL(t, appNameNormalizer.Normalize(app.Name))

	t.Run("Success when there is no default runtime assigned for eventing", func(t *testing.T) {
		// GIVEN
		ctx := fixCtxWithTenant()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), []string{}, fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixEmptyRuntimePage(), nil)

		svc := NewService(nil, runtimeRepo, nil, nil)

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
		runtimeRepo.On("List", ctx, tenantID.String(), []string{}, fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixRuntimePageWithOne(), nil)
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.RuntimeLabelableObject,
			runtimeID.String(), RuntimeEventingURLLabel).Return(fixRuntimeEventingURLLabel(), nil)
		labelRepo.On("Delete", ctx, tenantID.String(), model.RuntimeLabelableObject, runtimeID.String(),
			getDefaultEventingForAppLabelKey(applicationID)).Return(nil)
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.RuntimeLabelableObject,
			runtimeID.String(), isNormalizedLabel).Return(nil, apperrors.NewNotFoundError(resource.Runtime, runtimeID.String()))

		svc := NewService(appNameNormalizer, runtimeRepo, labelRepo, nil)

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
		runtimeRepo.On("List", ctx, tenantID.String(), []string{}, fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixRuntimePageWithOne(), nil)
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.RuntimeLabelableObject,
			runtimeID.String(), RuntimeEventingURLLabel).Return(fixRuntimeEventingURLLabel(), nil)
		labelRepo.On("Delete", ctx, tenantID.String(), model.RuntimeLabelableObject, runtimeID.String(),
			getDefaultEventingForAppLabelKey(applicationID)).Return(nil)
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.RuntimeLabelableObject,
			runtimeID.String(), isNormalizedLabel).Return(&model.Label{Value: "true"}, nil)

		svc := NewService(appNameNormalizer, runtimeRepo, labelRepo, nil)

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
		runtimeRepo.On("List", ctx, tenantID.String(), []string{}, fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixRuntimePageWithOne(), nil)
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.RuntimeLabelableObject,
			runtimeID.String(), RuntimeEventingURLLabel).Return(fixRuntimeEventingURLLabel(), nil)
		labelRepo.On("Delete", ctx, tenantID.String(), model.RuntimeLabelableObject, runtimeID.String(),
			getDefaultEventingForAppLabelKey(applicationID)).Return(nil)
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.RuntimeLabelableObject,
			runtimeID.String(), isNormalizedLabel).Return(&model.Label{Value: "false"}, nil)

		svc := NewService(appNameNormalizer, runtimeRepo, labelRepo, nil)

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
		svc := NewService(nil, nil, nil, nil)

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
		runtimeRepo.On("List", ctx, tenantID.String(), []string{}, fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(nil, errors.New("some-error"))

		svc := NewService(nil, runtimeRepo, nil, nil)

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
		runtimeRepo.On("List", ctx, tenantID.String(), []string{}, fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixRuntimePage(), nil)

		svc := NewService(nil, runtimeRepo, nil, nil)

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
		runtimeRepo.On("List", ctx, tenantID.String(), []string{}, fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixRuntimePageWithOne(), nil)
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("Delete", ctx, tenantID.String(), model.RuntimeLabelableObject, runtimeID.String(),
			getDefaultEventingForAppLabelKey(applicationID)).Return(errors.New("some-error"))

		svc := NewService(nil, runtimeRepo, labelRepo, nil)

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
		runtimeRepo.On("List", ctx, tenantID.String(), []string{}, fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixRuntimePageWithOne(), nil)
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.RuntimeLabelableObject,
			runtimeID.String(), RuntimeEventingURLLabel).Return(nil, errors.New("some error"))
		labelRepo.On("Delete", ctx, tenantID.String(), model.RuntimeLabelableObject, runtimeID.String(),
			getDefaultEventingForAppLabelKey(applicationID)).Return(nil)

		svc := NewService(nil, runtimeRepo, labelRepo, nil)

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
		runtimeRepo.On("List", ctx, tenantID.String(), []string{}, fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixRuntimePageWithOne(), nil)
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.RuntimeLabelableObject,
			runtimeID.String(), RuntimeEventingURLLabel).Return(fixRuntimeEventingURLLabel(), nil)
		labelRepo.On("Delete", ctx, tenantID.String(), model.RuntimeLabelableObject, runtimeID.String(),
			getDefaultEventingForAppLabelKey(applicationID)).Return(nil)
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.RuntimeLabelableObject,
			runtimeID.String(), isNormalizedLabel).Return(nil, errors.New("some error"))

		svc := NewService(appNameNormalizer, runtimeRepo, labelRepo, nil)

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
	formations := []*model.Formation{{Name: formationName}}
	formationNames := []string{formationName}
	runtimeIDs := []string{runtimeID.String()}

	t.Run("Success when there is default runtime labeled for application eventing", func(t *testing.T) {
		// GIVEN
		ctx := fixCtxWithTenant()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), []string{}, fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixRuntimePageWithOne(), nil)
		runtimeRepo.On("GetByID", ctx, tenantID.String(), runtimeID.String()).Return(fixRuntimes()[0], nil)
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.RuntimeLabelableObject,
			runtimeID.String(), RuntimeEventingURLLabel).Return(fixRuntimeEventingURLLabel(), nil)
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.RuntimeLabelableObject,
			runtimeID.String(), isNormalizedLabel).Return(nil, apperrors.NewNotFoundError(resource.Runtime, runtimeID.String()))
		formationSvc := &automock.FormationService{}
		formationSvc.On("ListFormationsForObject", ctx, app.ID).Return(formations, nil).Once()
		formationSvc.On("ListObjectIDsOfTypeForFormations", ctx, tenantID.String(), formationNames, model.FormationAssignmentTypeRuntime).Return(runtimeIDs, nil).Once()
		svc := NewService(appNameNormalizer, runtimeRepo, labelRepo, formationSvc)

		// WHEN
		eventingCfg, err := svc.GetForApplication(ctx, app)

		// THEN
		require.NoError(t, err)
		require.NotNil(t, eventingCfg)
		require.Equal(t, appNormalizedEventURL, eventingCfg.DefaultURL)
	mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo, formationSvc)
	})

	t.Run("Success when labeling oldest runtime for application eventing", func(t *testing.T) {
		// GIVEN
		ctx := fixCtxWithTenant()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), []string{}, fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixEmptyRuntimePage(), nil)
		runtimeRepo.On("GetOldestFromIDs", ctx, tenantID.String(), runtimeIDs).
			Return(fixRuntimes()[0], nil)
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("Upsert", ctx, tenantID.String(), mock.MatchedBy(fixMatcherDefaultEventingForAppLabel())).Return(nil)
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.RuntimeLabelableObject,
			runtimeID.String(), RuntimeEventingURLLabel).Return(fixRuntimeEventingURLLabel(), nil)
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.RuntimeLabelableObject,
			runtimeID.String(), isNormalizedLabel).Return(nil, apperrors.NewNotFoundError(resource.Runtime, runtimeID.String()))
		formationSvc := &automock.FormationService{}
		formationSvc.On("ListFormationsForObject", ctx, app.ID).Return(formations, nil).Once()
		formationSvc.On("ListObjectIDsOfTypeForFormations", ctx, tenantID.String(), formationNames, model.FormationAssignmentTypeRuntime).Return(runtimeIDs, nil).Once()
		svc := NewService(appNameNormalizer, runtimeRepo, labelRepo, formationSvc)

		// WHEN
		eventingCfg, err := svc.GetForApplication(ctx, app)

		// THEN
		require.NoError(t, err)
		require.NotNil(t, eventingCfg)
		require.Equal(t, appNormalizedEventURL, eventingCfg.DefaultURL)
	mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo, formationSvc)
	})

	t.Run("Success when there is runtime labeled for application eventing and is labeled for normalization", func(t *testing.T) {
		// GIVEN
		ctx := fixCtxWithTenant()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), []string{}, fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixRuntimePageWithOne(), nil)
		runtimeRepo.On("GetByID", ctx, tenantID.String(), runtimeID.String()).Return(fixRuntimes()[0], nil)
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.RuntimeLabelableObject,
			runtimeID.String(), RuntimeEventingURLLabel).Return(fixRuntimeEventingURLLabel(), nil)
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.RuntimeLabelableObject,
			runtimeID.String(), isNormalizedLabel).Return(&model.Label{Value: "true"}, nil)
		formationSvc := &automock.FormationService{}
		formationSvc.On("ListFormationsForObject", ctx, app.ID).Return(formations, nil).Once()
		formationSvc.On("ListObjectIDsOfTypeForFormations", ctx, tenantID.String(), formationNames, model.FormationAssignmentTypeRuntime).Return(runtimeIDs, nil).Once()
		svc := NewService(appNameNormalizer, runtimeRepo, labelRepo, formationSvc)

		// WHEN
		eventingCfg, err := svc.GetForApplication(ctx, app)

		// THEN
		require.NoError(t, err)
		require.NotNil(t, eventingCfg)
		require.Equal(t, appNormalizedEventURL, eventingCfg.DefaultURL)
	mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo, formationSvc)
	})

	t.Run("Success when there is runtime labeled for application eventing and is labeled not for normalization", func(t *testing.T) {
		// GIVEN
		ctx := fixCtxWithTenant()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), []string{}, fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixRuntimePageWithOne(), nil)
		runtimeRepo.On("GetByID", ctx, tenantID.String(), runtimeID.String()).Return(fixRuntimes()[0], nil)
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.RuntimeLabelableObject,
			runtimeID.String(), RuntimeEventingURLLabel).Return(fixRuntimeEventingURLLabel(), nil)
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.RuntimeLabelableObject,
			runtimeID.String(), isNormalizedLabel).Return(&model.Label{Value: "false"}, nil)
		formationSvc := &automock.FormationService{}
		formationSvc.On("ListFormationsForObject", ctx, app.ID).Return(formations, nil).Once()
		formationSvc.On("ListObjectIDsOfTypeForFormations", ctx, tenantID.String(), formationNames, model.FormationAssignmentTypeRuntime).Return(runtimeIDs, nil).Once()
		svc := NewService(appNameNormalizer, runtimeRepo, labelRepo, formationSvc)

		// WHEN
		eventingCfg, err := svc.GetForApplication(ctx, app)

		// THEN
		require.NoError(t, err)
		require.NotNil(t, eventingCfg)
		require.Equal(t, appEventURL, eventingCfg.DefaultURL)
	mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo, formationSvc)
	})

	t.Run("Success when there is no oldest runtime for application eventing", func(t *testing.T) {
		// GIVEN
		ctx := fixCtxWithTenant()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), []string{}, fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixEmptyRuntimePage(), nil)
		runtimeRepo.On("GetOldestFromIDs", ctx, tenantID.String(), runtimeIDs).
			Return(nil, apperrors.NewNotFoundError(resource.Runtime, ""))
		labelRepo := &automock.LabelRepository{}
		formationSvc := &automock.FormationService{}
		formationSvc.On("ListFormationsForObject", ctx, app.ID).Return(formations, nil).Once()
		formationSvc.On("ListObjectIDsOfTypeForFormations", ctx, tenantID.String(), formationNames, model.FormationAssignmentTypeRuntime).Return(runtimeIDs, nil).Once()
		svc := NewService(appNameNormalizer, runtimeRepo, labelRepo, formationSvc)

		// WHEN
		eventingCfg, err := svc.GetForApplication(ctx, app)

		// THEN
		require.NoError(t, err)
		require.NotNil(t, eventingCfg)
		require.Equal(t, "", eventingCfg.DefaultURL.String())
	mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo, formationSvc)
	})

	t.Run("Success when there is no oldest runtime for application eventing (scenarios label does not exist)", func(t *testing.T) {
		// GIVEN
		ctx := fixCtxWithTenant()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), []string{}, fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixEmptyRuntimePage(), nil)
		labelRepo := &automock.LabelRepository{}
		formationSvc := &automock.FormationService{}
		formationSvc.On("ListFormationsForObject", ctx, app.ID).Return(nil, nil).Once()
		svc := NewService(appNameNormalizer, runtimeRepo, labelRepo, formationSvc)

		// WHEN
		eventingCfg, err := svc.GetForApplication(ctx, app)

		// THEN
		require.NoError(t, err)
		require.NotNil(t, eventingCfg)
		require.Equal(t, "", eventingCfg.DefaultURL.String())
		mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo, formationSvc)
	})

	t.Run("Empty configuration when there are no scenarios assigned to application", func(t *testing.T) {
		// GIVEN
		ctx := fixCtxWithTenant()
		emptyConfiguration := &model.ApplicationEventingConfiguration{
			EventingConfiguration: model.EventingConfiguration{
				DefaultURL: url.URL{},
			},
		}
		runtimeRepo := &automock.RuntimeRepository{}
		runtimePage := fixRuntimePageWithOne()
		runtimeRepo.On("List", ctx, tenantID.String(), []string{}, fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(runtimePage, nil)
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("Delete", ctx, tenantID.String(), model.RuntimeLabelableObject, runtimePage.Data[0].ID, getDefaultEventingForAppLabelKey(applicationID)).Return(nil)
		formationSvc := &automock.FormationService{}
		formationSvc.On("ListFormationsForObject", ctx, app.ID).Return(nil, nil).Twice()
		svc := NewService(appNameNormalizer, runtimeRepo, labelRepo, formationSvc)

		// WHEN
		conf, err := svc.GetForApplication(ctx, app)

		// THEN
		require.NoError(t, err)
		require.Equal(t, emptyConfiguration, conf)
	mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo, formationSvc)
	})

	t.Run("Success when current default runtime no longer belongs to the application scenarios", func(t *testing.T) {
		// GIVEN
		ctx := fixCtxWithTenant()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), []string{}, fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixRuntimePageWithOne(), nil)
		runtimeRepo.On("GetOldestFromIDs", ctx, tenantID.String(), runtimeIDs).
			Return(nil, apperrors.NewNotFoundError(resource.Runtime, ""))
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("Delete", ctx, tenantID.String(), model.RuntimeLabelableObject, runtimeID.String(),
			getDefaultEventingForAppLabelKey(applicationID)).Return(nil)
		formationSvc := &automock.FormationService{}
		formationSvc.On("ListFormationsForObject", ctx, app.ID).Return(formations, nil).Twice()
		formationSvc.On("ListObjectIDsOfTypeForFormations", ctx, tenantID.String(), formationNames, model.FormationAssignmentTypeRuntime).Return([]string{}, nil).Once()
		formationSvc.On("ListObjectIDsOfTypeForFormations", ctx, tenantID.String(), formationNames, model.FormationAssignmentTypeRuntime).Return(runtimeIDs, nil).Once()
		svc := NewService(appNameNormalizer, runtimeRepo, labelRepo, formationSvc)

		// WHEN
		eventingCfg, err := svc.GetForApplication(ctx, app)

		// THEN
		require.NoError(t, err)
		require.NotNil(t, eventingCfg)
		require.Equal(t, "", eventingCfg.DefaultURL.String())
	mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo, formationSvc)
	})

	t.Run("Success when new default runtime is elected", func(t *testing.T) {
		// GIVEN
		defaultURL := "https://default.URL"
		ctx := fixCtxWithTenant()
		runtimeRepo := &automock.RuntimeRepository{}
		newRuntime := fixRuntimes()[0]
		runtimePage := fixRuntimePageWithOne()
		runtimeRepo.On("List", ctx, tenantID.String(), []string{}, fixLabelFilterForRuntimeDefaultEventingForApp(), 1, mock.Anything).Return(runtimePage, nil)
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.RuntimeLabelableObject,
			newRuntime.ID, RuntimeEventingURLLabel).Return(&model.Label{Key: RuntimeEventingURLLabel, Value: defaultURL}, nil).Once()
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.RuntimeLabelableObject,
			newRuntime.ID, isNormalizedLabel).Return(&model.Label{Value: "false"}, nil).Once()
		labelRepo.On("Delete", ctx, tenantID.String(), model.RuntimeLabelableObject, runtimePage.Data[0].ID, getDefaultEventingForAppLabelKey(applicationID)).Return(nil)
		labelRepo.On("Upsert", ctx, tenantID.String(), mock.Anything).Return(nil)

		runtimeRepo.On("GetOldestFromIDs", ctx, tenantID.String(), runtimeIDs).Return(newRuntime, nil)
		formationSvc := &automock.FormationService{}
		formationSvc.On("ListFormationsForObject", ctx, app.ID).Return(formations, nil).Twice()
		formationSvc.On("ListObjectIDsOfTypeForFormations", ctx, tenantID.String(), formationNames, model.FormationAssignmentTypeRuntime).Return([]string{}, nil).Once()
		formationSvc.On("ListObjectIDsOfTypeForFormations", ctx, tenantID.String(), formationNames, model.FormationAssignmentTypeRuntime).Return(runtimeIDs, nil).Once()
		svc := NewService(appNameNormalizer, runtimeRepo, labelRepo, formationSvc)

		// WHEN
		conf, err := svc.GetForApplication(ctx, app)

		// THEN
		require.NoError(t, err)
		require.Equal(t, fmt.Sprintf("%s/%s/v1/events", defaultURL, app.Name), conf.DefaultURL.String())
	mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo, formationSvc)
	})

	t.Run("Error when tenant not in context", func(t *testing.T) {
		// GIVEN
		svc := NewService(nil, nil, nil, nil)

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
		runtimeRepo.On("List", ctx, tenantID.String(), []string{}, fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixEmptyRuntimePage(), nil)
		runtimeRepo.On("GetOldestFromIDs", ctx, tenantID.String(), runtimeIDs).
			Return(fixRuntimes()[0], nil)
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("Upsert", ctx, tenantID.String(),
			mock.MatchedBy(fixMatcherDefaultEventingForAppLabel())).
			Return(errors.New("some error"))
		formationSvc := &automock.FormationService{}
		formationSvc.On("ListFormationsForObject", ctx, app.ID).Return(formations, nil).Once()
		formationSvc.On("ListObjectIDsOfTypeForFormations", ctx, tenantID.String(), formationNames, model.FormationAssignmentTypeRuntime).Return(runtimeIDs, nil).Once()
		svc := NewService(appNameNormalizer, runtimeRepo, labelRepo, formationSvc)

		// WHEN
		_, err := svc.GetForApplication(ctx, app)

		// THEN
		require.Error(t, err)
		require.EqualError(t, err, expectedError)
	mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo, formationSvc)
	})

	t.Run("Error when getting the oldest runtime for application eventing returns error", func(t *testing.T) {
		// GIVEN
		expectedError := fmt.Sprintf(`while getting the oldest runtime for scenarios: while getting the oldest runtime for application [ID=%s] scenarios with filter: some error`, applicationID)
		ctx := fixCtxWithTenant()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), []string{}, fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixEmptyRuntimePage(), nil)
		runtimeRepo.On("GetOldestFromIDs", ctx, tenantID.String(), runtimeIDs).
			Return(nil, errors.New("some error"))
		labelRepo := &automock.LabelRepository{}
		formationSvc := &automock.FormationService{}
		formationSvc.On("ListFormationsForObject", ctx, app.ID).Return(formations, nil).Once()
		formationSvc.On("ListObjectIDsOfTypeForFormations", ctx, tenantID.String(), formationNames, model.FormationAssignmentTypeRuntime).Return(runtimeIDs, nil).Once()
		svc := NewService(appNameNormalizer, runtimeRepo, labelRepo, formationSvc)

		// WHEN
		_, err := svc.GetForApplication(ctx, app)

		// THEN
		require.Error(t, err)
		require.EqualError(t, err, expectedError)
	mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo, formationSvc)
	})

	t.Run("Error when getting the oldest runtime for application eventing returns error on converting scenarios label to slice of strings", func(t *testing.T) {
		// GIVEN
		expectedError := fmt.Sprintf(`while getting the oldest runtime for scenarios: while getting runtimes IDs in formation with application: %s: while getting Formations for Application with ID: %s: some error`, applicationID, applicationID)
		ctx := fixCtxWithTenant()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), []string{}, fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixEmptyRuntimePage(), nil)
		labelRepo := &automock.LabelRepository{}
		formationSvc := &automock.FormationService{}
		formationSvc.On("ListFormationsForObject", ctx, app.ID).Return(nil, errors.New("some error")).Once()
		svc := NewService(appNameNormalizer, runtimeRepo, labelRepo, formationSvc)

		// WHEN
		_, err := svc.GetForApplication(ctx, app)

		// THEN
		require.Error(t, err)
		require.EqualError(t, err, expectedError)
		mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo, formationSvc)
	})

	t.Run("Error when getting the oldest runtime for application eventing returns error when getting scenarios label", func(t *testing.T) {
		// GIVEN
		expectedError := fmt.Sprintf(`while getting the oldest runtime for scenarios: while getting runtimes IDs in formation with application: %s: while getting Formations for Application with ID: %s: some error`, applicationID, applicationID)
		ctx := fixCtxWithTenant()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), []string{}, fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixEmptyRuntimePage(), nil)
		formationSvc := &automock.FormationService{}
		formationSvc.On("ListFormationsForObject", ctx, app.ID).Return(nil, errors.New("some error")).Once()
		svc := NewService(appNameNormalizer, runtimeRepo, nil, formationSvc)

		// WHEN
		_, err := svc.GetForApplication(ctx, app)

		// THEN
		require.Error(t, err)
		require.EqualError(t, err, expectedError)
		mock.AssertExpectationsForObjects(t, runtimeRepo, formationSvc)
	})

	t.Run("Error when getting default runtime labeled for application eventing returns error", func(t *testing.T) {
		// GIVEN
		expectedError := fmt.Sprintf(`while getting default runtime for app eventing: while fetching runtimes with label [key=%s]: some error`, getDefaultEventingForAppLabelKey(applicationID))
		ctx := fixCtxWithTenant()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), []string{}, fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(nil, errors.New("some error"))
		svc := NewService(nil, runtimeRepo, nil, nil)

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
		runtimeRepo.On("List", ctx, tenantID.String(), []string{}, fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixRuntimePage(), nil)

		svc := NewService(nil, runtimeRepo, nil, nil)

		// WHEN
		_, err := svc.GetForApplication(ctx, app)

		// THEN
		require.Error(t, err)
		require.EqualError(t, err, expectedError)
		mock.AssertExpectationsForObjects(t, runtimeRepo)
	})

	t.Run("Error when ensuring the scenarios, getting scenarios returns error", func(t *testing.T) {
		// GIVEN
		expectedError := fmt.Sprintf(`while ensuring the scenarios assigned to the runtime and application: while verifing whether runtime [ID=%s] belongs to the application scenarios: while getting runtimes IDs in formation with application: %s: while getting Formations for Application with ID: %s: some error`, runtimeID, applicationID, applicationID)
		ctx := fixCtxWithTenant()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), []string{}, fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixRuntimePageWithOne(), nil)
		formationSvc := &automock.FormationService{}
		formationSvc.On("ListFormationsForObject", ctx, app.ID).Return(nil, errors.New("some error")).Once()
		svc := NewService(appNameNormalizer, runtimeRepo, nil, formationSvc)

		// WHEN
		_, err := svc.GetForApplication(ctx, app)

		// THEN
		require.Error(t, err)
		require.EqualError(t, err, expectedError)
	mock.AssertExpectationsForObjects(t, runtimeRepo, formationSvc)
	})

	t.Run("Error when verifing whether given runtime belongs to the application scenarios - repository returns error", func(t *testing.T) {
		// GIVEN
		expectedError := fmt.Sprintf(`while ensuring the scenarios assigned to the runtime and application: while verifing whether runtime [ID=%s] belongs to the application scenarios: while getting the runtime [ID=%s] with scenarios with filter: some-error`, runtimeID, runtimeID)
		ctx := fixCtxWithTenant()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), []string{}, fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixRuntimePageWithOne(), nil)
		runtimeRepo.On("GetByID", ctx, tenantID.String(), runtimeID.String()).Return(nil, errors.New("some-error"))

		formationSvc := &automock.FormationService{}
		formationSvc.On("ListFormationsForObject", ctx, app.ID).Return(formations, nil).Once()
		formationSvc.On("ListObjectIDsOfTypeForFormations", ctx, tenantID.String(), formationNames, model.FormationAssignmentTypeRuntime).Return(runtimeIDs, nil).Once()
		svc := NewService(appNameNormalizer, runtimeRepo, nil, formationSvc)

		// WHEN
		_, err := svc.GetForApplication(ctx, app)

		// THEN
		require.Error(t, err)
		require.EqualError(t, err, expectedError)
	mock.AssertExpectationsForObjects(t, runtimeRepo, formationSvc)
	})

	t.Run("Error when deleting label from the given runtime because it does not belong to application scenarios - repository returns error", func(t *testing.T) {
		// GIVEN
		expectedError := fmt.Sprintf(`while ensuring the scenarios assigned to the runtime and application: when deleting current default runtime for the application because of scenarios mismatch: while deleting label [key=%s, runtimeID=%s]: some-error`, getDefaultEventingForAppLabelKey(applicationID), runtimeID)
		ctx := fixCtxWithTenant()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), []string{}, fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixRuntimePageWithOne(), nil)
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("Delete", ctx, tenantID.String(), model.RuntimeLabelableObject, runtimeID.String(),
			getDefaultEventingForAppLabelKey(applicationID)).Return(errors.New("some-error"))
		formationSvc := &automock.FormationService{}
		formationSvc.On("ListFormationsForObject", ctx, app.ID).Return(formations, nil).Once()
		formationSvc.On("ListObjectIDsOfTypeForFormations", ctx, tenantID.String(), formationNames, model.FormationAssignmentTypeRuntime).Return([]string{}, nil).Once()
		svc := NewService(appNameNormalizer, runtimeRepo, labelRepo, formationSvc)

		// WHEN
		_, err := svc.GetForApplication(ctx, app)

		// THEN
		require.Error(t, err)
		require.EqualError(t, err, expectedError)
	mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo, formationSvc)
	})

	t.Run("Error when getting eventing configuration for a given runtime", func(t *testing.T) {
		// GIVEN
		expectedError := fmt.Sprintf(`while fetching eventing configuration for runtime: while getting the label [key=%s] for runtime [ID=%s]: some error`, RuntimeEventingURLLabel, runtimeID)
		ctx := fixCtxWithTenant()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), []string{}, fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixRuntimePageWithOne(), nil)
		runtimeRepo.On("GetByID", ctx, tenantID.String(), runtimeID.String()).Return(fixRuntimes()[0], nil)
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.RuntimeLabelableObject,
			runtimeID.String(), RuntimeEventingURLLabel).Return(nil, errors.New("some error"))
		formationSvc := &automock.FormationService{}
		formationSvc.On("ListFormationsForObject", ctx, app.ID).Return(formations, nil).Once()
		formationSvc.On("ListObjectIDsOfTypeForFormations", ctx, tenantID.String(), formationNames, model.FormationAssignmentTypeRuntime).Return(runtimeIDs, nil).Once()
		svc := NewService(appNameNormalizer, runtimeRepo, labelRepo, formationSvc)

		// WHEN
		_, err := svc.GetForApplication(ctx, app)

		// THEN
		require.Error(t, err)
		require.EqualError(t, err, expectedError)
	mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo, formationSvc)
	})

	t.Run("Error when getting runtime normalization label", func(t *testing.T) {
		// GIVEN
		expectedError := fmt.Sprintf(`while determining whether application name should be normalized in runtime eventing URL: while getting the label [key=%s] for runtime [ID=%s]: some error`, isNormalizedLabel, runtimeID)
		ctx := fixCtxWithTenant()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("List", ctx, tenantID.String(), []string{}, fixLabelFilterForRuntimeDefaultEventingForApp(),
			1, mock.Anything).Return(fixRuntimePageWithOne(), nil)
		runtimeRepo.On("GetByID", ctx, tenantID.String(), runtimeID.String()).Return(fixRuntimes()[0], nil)
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.RuntimeLabelableObject,
			runtimeID.String(), RuntimeEventingURLLabel).Return(fixRuntimeEventingURLLabel(), nil)
		labelRepo.On("GetByKey", ctx, tenantID.String(), model.RuntimeLabelableObject,
			runtimeID.String(), isNormalizedLabel).Return(nil, errors.New("some error"))
		formationSvc := &automock.FormationService{}
		formationSvc.On("ListFormationsForObject", ctx, app.ID).Return(formations, nil).Once()
		formationSvc.On("ListObjectIDsOfTypeForFormations", ctx, tenantID.String(), formationNames, model.FormationAssignmentTypeRuntime).Return(runtimeIDs, nil).Once()
		svc := NewService(appNameNormalizer, runtimeRepo, labelRepo, formationSvc)
		// WHEN
		_, err := svc.GetForApplication(ctx, app)

		// THEN
		require.Error(t, err)
		require.EqualError(t, err, expectedError)
	mock.AssertExpectationsForObjects(t, runtimeRepo, labelRepo, formationSvc)
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
		eventingSvc := NewService(nil, nil, labelRepo, nil)

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
		eventingSvc := NewService(nil, nil, labelRepo, nil)

		// WHEN
		eventingCfg, err := eventingSvc.GetForRuntime(ctx, runtimeID)

		// THEN
		require.NoError(t, err)
		require.Equal(t, expectedEventingCfg, eventingCfg)
		mock.AssertExpectationsForObjects(t, labelRepo)
	})

	t.Run("Error when tenant not in context", func(t *testing.T) {
		// GIVEN
		svc := NewService(nil, nil, nil, nil)

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
		eventingSvc := NewService(nil, nil, labelRepo, nil)

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
		eventingSvc := NewService(nil, nil, labelRepo, nil)

		// WHEN
		_, err := eventingSvc.GetForRuntime(ctx, runtimeID)

		// THEN
		require.Error(t, err)
		require.EqualError(t, err, `unable to cast label [key=`+RuntimeEventingURLLabel+`, runtimeID=`+runtimeID.String()+`] value as a string`)
		mock.AssertExpectationsForObjects(t, labelRepo)
	})
}
