package datainputbuilder_test

import (
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/kyma-incubator/compass/components/director/pkg/webhook"

	databuilder "github.com/kyma-incubator/compass/components/director/internal/domain/webhook/datainputbuilder"
	"github.com/kyma-incubator/compass/components/director/internal/domain/webhook/datainputbuilder/automock"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestWebhookTenantBuilder_GetTenantForObject(t *testing.T) {
	testCases := []struct {
		name                     string
		labelBuilder             func() *automock.LabelInputBuilder
		tenantRepo               func() *automock.TenantRepository
		objectID                 string
		objectType               resource.Type
		expectedTenantWithLabels *webhook.TenantWithLabels
		expectedErrMsg           string
	}{
		{
			name: "success",
			labelBuilder: func() *automock.LabelInputBuilder {
				builder := &automock.LabelInputBuilder{}
				builder.On("GetLabelsForObject", emptyCtx, ApplicationTenantID, ApplicationTenantID, model.TenantLabelableObject).Return(convertLabels(testTenantLabels), nil).Once()
				return builder
			},
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("GetLowestOwnerForResource", emptyCtx, resource.Application, ApplicationID).Return(ApplicationTenantID, nil).Once()
				repo.On("Get", emptyCtx, ApplicationTenantID).Return(testApplicationTenantOwner, nil).Once()
				return repo
			},
			objectID:                 ApplicationID,
			objectType:               resource.Application,
			expectedTenantWithLabels: testAppTenantWithLabels,
		},
		{
			name:         "error when getting lowest tenant owner fails",
			labelBuilder: unusedLabelBuilder,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("GetLowestOwnerForResource", emptyCtx, resource.Application, ApplicationID).Return("", testErr).Once()
				return repo
			},
			objectID:       ApplicationID,
			objectType:     resource.Application,
			expectedErrMsg: fmt.Sprintf("while getting tenant lowest owner for object with id %q", ApplicationID),
		},
		{
			name:         "error when getting tenant by id fails",
			labelBuilder: unusedLabelBuilder,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("GetLowestOwnerForResource", emptyCtx, resource.Application, ApplicationID).Return(ApplicationTenantID, nil).Once()
				repo.On("Get", emptyCtx, ApplicationTenantID).Return(nil, testErr).Once()
				return repo
			},
			objectID:       ApplicationID,
			objectType:     resource.Application,
			expectedErrMsg: fmt.Sprintf("while getting tenant for object with id %q", ApplicationTenantID),
		},
		{
			name: "error when building labels for object fails",
			labelBuilder: func() *automock.LabelInputBuilder {
				builder := &automock.LabelInputBuilder{}
				builder.On("GetLabelsForObject", emptyCtx, ApplicationTenantID, ApplicationTenantID, model.TenantLabelableObject).Return(nil, testErr).Once()
				return builder
			},
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("GetLowestOwnerForResource", emptyCtx, resource.Application, ApplicationID).Return(ApplicationTenantID, nil).Once()
				repo.On("Get", emptyCtx, ApplicationTenantID).Return(testApplicationTenantOwner, nil).Once()
				return repo
			},
			objectID:       ApplicationID,
			objectType:     resource.Application,
			expectedErrMsg: "while listing tenant labels",
		},
	}

	for _, tCase := range testCases {
		t.Run(tCase.name, func(t *testing.T) {
			// GIVEN
			labelBuilder := tCase.labelBuilder()
			tenantRepo := tCase.tenantRepo()
			defer mock.AssertExpectationsForObjects(t, labelBuilder, tenantRepo)

			webhookDataInputBuilder := databuilder.NewWebhookTenantBuilder(labelBuilder, tenantRepo)

			// WHEN
			tenantWithLabels, err := webhookDataInputBuilder.GetTenantForObject(emptyCtx, tCase.objectID, tCase.objectType)

			// THEN
			if tCase.expectedErrMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tCase.expectedErrMsg)
				require.Nil(t, tenantWithLabels)
			} else {
				require.NoError(t, err)
				require.Equal(t, tCase.expectedTenantWithLabels, tenantWithLabels)
			}
		})
	}
}

func TestWebhookTenantBuilder_GetTenantsForObjects(t *testing.T) {
	testCases := []struct {
		name                      string
		labelBuilder              func() *automock.LabelInputBuilder
		tenantRepo                func() *automock.TenantRepository
		objectIDs                 []string
		objectType                resource.Type
		expectedTenantsWithLabels map[string]*webhook.TenantWithLabels
		expectedErrMsg            string
	}{
		{
			name: "success",
			labelBuilder: func() *automock.LabelInputBuilder {
				builder := &automock.LabelInputBuilder{}
				builder.On("GetLabelsForObjects", emptyCtx, testTenantID, []string{ApplicationTenantID}, model.TenantLabelableObject).Return(map[string]map[string]string{
					ApplicationTenantID: convertLabels(testTenantLabels),
				}, nil).Once()
				return builder
			},
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("GetLowestOwnerForResource", emptyCtx, resource.Application, ApplicationID).Return(ApplicationTenantID, nil).Once()
				repo.On("GetLowestOwnerForResource", emptyCtx, resource.Application, Application2ID).Return(ApplicationTenantID, nil).Once()
				repo.On("Get", emptyCtx, ApplicationTenantID).Return(testApplicationTenantOwner, nil).Once()
				return repo
			},
			objectIDs:  []string{ApplicationID, Application2ID},
			objectType: resource.Application,
			expectedTenantsWithLabels: map[string]*webhook.TenantWithLabels{
				ApplicationID:  testAppTenantWithLabels,
				Application2ID: testAppTenantWithLabels,
			},
		},
		{
			name:         "error when getting lowest tenant resource owner for object fails",
			labelBuilder: unusedLabelBuilder,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("GetLowestOwnerForResource", emptyCtx, resource.Application, ApplicationID).Return("", testErr).Once()
				return repo
			},
			objectIDs:      []string{ApplicationID},
			objectType:     resource.Application,
			expectedErrMsg: fmt.Sprintf("while getting tenant for object with ID %q", ApplicationID),
		},
		{
			name:         "error when getting tenant by id fails",
			labelBuilder: unusedLabelBuilder,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("GetLowestOwnerForResource", emptyCtx, resource.Application, ApplicationID).Return(ApplicationTenantID, nil).Once()
				repo.On("Get", emptyCtx, ApplicationTenantID).Return(nil, testErr).Once()
				return repo
			},
			objectIDs:      []string{ApplicationID},
			objectType:     resource.Application,
			expectedErrMsg: fmt.Sprintf("while getting tenant with ID %q", ApplicationTenantID),
		},
		{
			name: "error when building labels for tenants fails",
			labelBuilder: func() *automock.LabelInputBuilder {
				builder := &automock.LabelInputBuilder{}
				builder.On("GetLabelsForObjects", emptyCtx, testTenantID, []string{ApplicationTenantID}, model.TenantLabelableObject).Return(nil, testErr).Once()
				return builder
			},
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("GetLowestOwnerForResource", emptyCtx, resource.Application, ApplicationID).Return(ApplicationTenantID, nil).Once()
				repo.On("Get", emptyCtx, ApplicationTenantID).Return(testApplicationTenantOwner, nil).Once()
				return repo
			},
			objectIDs:      []string{ApplicationID},
			objectType:     resource.Application,
			expectedErrMsg: "while building tenant labels",
		},
	}

	for _, tCase := range testCases {
		t.Run(tCase.name, func(t *testing.T) {
			// GIVEN
			labelBuilder := tCase.labelBuilder()
			tenantRepo := tCase.tenantRepo()
			defer mock.AssertExpectationsForObjects(t, labelBuilder, tenantRepo)

			webhookDataInputBuilder := databuilder.NewWebhookTenantBuilder(labelBuilder, tenantRepo)

			// WHEN
			tenantsWithLabels, err := webhookDataInputBuilder.GetTenantsForObjects(emptyCtx, testTenantID, tCase.objectIDs, tCase.objectType)

			// THEN
			if tCase.expectedErrMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tCase.expectedErrMsg)
				require.Nil(t, tenantsWithLabels)
			} else {
				require.NoError(t, err)
				require.Equal(t, tCase.expectedTenantsWithLabels, tenantsWithLabels)
			}
		})
	}
}

func TestWebhookTenantBuilder_GetTenantForApplicationTemplate(t *testing.T) {
	testCases := []struct {
		name                     string
		labelBuilder             func() *automock.LabelInputBuilder
		tenantRepo               func() *automock.TenantRepository
		objectLabels             map[string]string
		expectedTenantWithLabels *webhook.TenantWithLabels
		expectedErrMsg           string
	}{
		{
			name: "success when application templates has owner",
			labelBuilder: func() *automock.LabelInputBuilder {
				builder := &automock.LabelInputBuilder{}
				builder.On("GetLabelsForObject", emptyCtx, testTenantID, ApplicationTemplateTenantID, model.TenantLabelableObject).Return(convertLabels(testTenantLabels), nil).Once()
				return builder
			},
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("GetByExternalTenant", emptyCtx, ApplicationTemplateTenantID).Return(testApplicationTemplateTenantOwner, nil).Once()
				return repo
			},
			objectLabels:             fixLabelsMapForApplicationTemplateWithSubaccountLabels(),
			expectedTenantWithLabels: testAppTemplateTenantWithLabels,
		},
		{
			name:         "error when getting tenant fails",
			labelBuilder: unusedLabelBuilder,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("GetByExternalTenant", emptyCtx, ApplicationTemplateTenantID).Return(nil, testErr).Once()
				return repo
			},
			objectLabels:   fixLabelsMapForApplicationTemplateWithSubaccountLabels(),
			expectedErrMsg: fmt.Sprintf("while getting tenant by external ID %q", ApplicationTemplateTenantID),
		},
		{
			name: "error when building labels fails",
			labelBuilder: func() *automock.LabelInputBuilder {
				builder := &automock.LabelInputBuilder{}
				builder.On("GetLabelsForObject", emptyCtx, testTenantID, ApplicationTemplateTenantID, model.TenantLabelableObject).Return(nil, testErr).Once()
				return builder
			},
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("GetByExternalTenant", emptyCtx, ApplicationTemplateTenantID).Return(testApplicationTemplateTenantOwner, nil).Once()
				return repo
			},
			objectLabels:   fixLabelsMapForApplicationTemplateWithSubaccountLabels(),
			expectedErrMsg: "while listing tenant labels",
		},
		{
			name:                     "success when application templates has no owner",
			labelBuilder:             unusedLabelBuilder,
			tenantRepo:               unusedTenantRepo,
			objectLabels:             fixLabelsMapForApplicationTemplateWithLabels(),
			expectedTenantWithLabels: nil,
		},
	}

	for _, tCase := range testCases {
		t.Run(tCase.name, func(t *testing.T) {
			// GIVEN
			labelBuilder := tCase.labelBuilder()
			tenantRepo := tCase.tenantRepo()
			defer mock.AssertExpectationsForObjects(t, labelBuilder, tenantRepo)

			webhookDataInputBuilder := databuilder.NewWebhookTenantBuilder(labelBuilder, tenantRepo)

			// WHEN
			tenantWithLabels, err := webhookDataInputBuilder.GetTenantForApplicationTemplate(emptyCtx, testTenantID, tCase.objectLabels)

			// THEN
			if tCase.expectedErrMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tCase.expectedErrMsg)
				require.Nil(t, tenantWithLabels)
			} else {
				require.NoError(t, err)
				require.Equal(t, tCase.expectedTenantWithLabels, tenantWithLabels)
			}
		})
	}
}

func TestWebhookTenantBuilder_GetTenantsForApplicationTemplates(t *testing.T) {
	testCases := []struct {
		name                      string
		labelBuilder              func() *automock.LabelInputBuilder
		tenantRepo                func() *automock.TenantRepository
		objectLabels              map[string]map[string]string
		objectIDs                 []string
		expectedTenantsWithLabels map[string]*webhook.TenantWithLabels
		expectedErrMsg            string
	}{
		{
			name: "success when all application templates have owner",
			labelBuilder: func() *automock.LabelInputBuilder {
				builder := &automock.LabelInputBuilder{}
				builder.On("GetLabelsForObjects", emptyCtx, testTenantID, []string{ApplicationTemplateTenantID}, model.TenantLabelableObject).Return(map[string]map[string]string{
					ApplicationTemplateTenantID: convertLabels(testTenantLabels),
				}, nil).Once()
				return builder
			},
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("GetByExternalTenant", emptyCtx, ApplicationTemplateTenantID).Return(testApplicationTemplateTenantOwner, nil).Once()
				return repo
			},
			objectLabels: map[string]map[string]string{
				ApplicationTemplateID:  fixLabelsMapForApplicationTemplateWithSubaccountLabels(),
				ApplicationTemplate2ID: fixLabelsMapForApplicationTemplateWithSubaccountLabels(),
			},
			objectIDs: []string{ApplicationTemplateID, ApplicationTemplate2ID},
			expectedTenantsWithLabels: map[string]*webhook.TenantWithLabels{
				ApplicationTemplateID:  testAppTemplateTenantWithLabels,
				ApplicationTemplate2ID: testAppTemplateTenantWithLabels,
			},
		},
		{
			name: "success when not all application templates have owner",
			labelBuilder: func() *automock.LabelInputBuilder {
				builder := &automock.LabelInputBuilder{}
				builder.On("GetLabelsForObjects", emptyCtx, testTenantID, []string{ApplicationTemplateTenantID}, model.TenantLabelableObject).Return(map[string]map[string]string{
					ApplicationTemplateTenantID: convertLabels(testTenantLabels),
				}, nil).Once()
				return builder
			},
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("GetByExternalTenant", emptyCtx, ApplicationTemplateTenantID).Return(testApplicationTemplateTenantOwner, nil).Once()
				return repo
			},
			objectLabels: map[string]map[string]string{
				ApplicationTemplateID: fixLabelsMapForApplicationTemplateWithSubaccountLabels(),
			},
			objectIDs: []string{ApplicationTemplateID, ApplicationTemplate2ID},
			expectedTenantsWithLabels: map[string]*webhook.TenantWithLabels{
				ApplicationTemplateID: testAppTemplateTenantWithLabels,
			},
		},
		{
			name:         "error when getting tenant fails",
			labelBuilder: unusedLabelBuilder,
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("GetByExternalTenant", emptyCtx, ApplicationTemplateTenantID).Return(nil, testErr).Once()
				return repo
			},
			objectLabels: map[string]map[string]string{
				ApplicationTemplateID: fixLabelsMapForApplicationTemplateWithSubaccountLabels(),
			},
			objectIDs:      []string{ApplicationTemplateID},
			expectedErrMsg: fmt.Sprintf("while getting tenant by external ID %q", ApplicationTemplateTenantID),
		},
		{
			name: "error when building labels for tenants fails",
			labelBuilder: func() *automock.LabelInputBuilder {
				builder := &automock.LabelInputBuilder{}
				builder.On("GetLabelsForObjects", emptyCtx, testTenantID, []string{ApplicationTemplateTenantID}, model.TenantLabelableObject).Return(nil, testErr).Once()
				return builder
			},
			tenantRepo: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("GetByExternalTenant", emptyCtx, ApplicationTemplateTenantID).Return(testApplicationTemplateTenantOwner, nil).Once()
				return repo
			},
			objectLabels: map[string]map[string]string{
				ApplicationTemplateID: fixLabelsMapForApplicationTemplateWithSubaccountLabels(),
			},
			objectIDs:      []string{ApplicationTemplateID},
			expectedErrMsg: "while listing tenant labels",
		},
	}

	for _, tCase := range testCases {
		t.Run(tCase.name, func(t *testing.T) {
			// GIVEN
			labelBuilder := tCase.labelBuilder()
			tenantRepo := tCase.tenantRepo()
			defer mock.AssertExpectationsForObjects(t, labelBuilder, tenantRepo)

			webhookDataInputBuilder := databuilder.NewWebhookTenantBuilder(labelBuilder, tenantRepo)

			// WHEN
			tenantsWithLabels, err := webhookDataInputBuilder.GetTenantsForApplicationTemplates(emptyCtx, testTenantID, tCase.objectLabels, tCase.objectIDs)

			// THEN
			if tCase.expectedErrMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tCase.expectedErrMsg)
				require.Nil(t, tenantsWithLabels)
			} else {
				require.NoError(t, err)
				require.Equal(t, tCase.expectedTenantsWithLabels, tenantsWithLabels)
			}
		})
	}
}
