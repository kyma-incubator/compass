package formation_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formation"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formation/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	webhookclient "github.com/kyma-incubator/compass/components/director/pkg/webhook_client"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_GenerateFormationLifecycleNotifications(t *testing.T) {
	ctx := context.Background()
	formationInput := fixFormationModelWithoutError()

	testCases := []struct {
		name                              string
		notificationsBuilderFn            func() *automock.NotificationBuilder
		expectedErrMsg                    string
		expectedFormationNotificationReqs []*webhookclient.FormationNotificationRequest
	}{
		{
			name: "Successfully generate formation notifications",
			notificationsBuilderFn: func() *automock.NotificationBuilder {
				notificationBuilder := &automock.NotificationBuilder{}
				notificationBuilder.On("BuildFormationNotificationRequests", ctx, formationNotificationDetails, formationInput, formationLifecycleSyncWebhooks).Return(formationNotificationSyncCreateRequests, nil).Once()
				return notificationBuilder
			},
			expectedFormationNotificationReqs: formationNotificationSyncCreateRequests,
		},
		{
			name: "Success when generating formation lifecycle notifications results in not generated notification due to error",
			notificationsBuilderFn: func() *automock.NotificationBuilder {
				notificationBuilder := &automock.NotificationBuilder{}
				notificationBuilder.On("BuildFormationNotificationRequests", ctx, formationNotificationDetails, formationInput, formationLifecycleSyncWebhooks).Return(nil, testErr).Once()
				return notificationBuilder
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			notificationsBuilder := unusedNotificationsBuilder()
			if testCase.notificationsBuilderFn != nil {
				notificationsBuilder = testCase.notificationsBuilderFn()
			}

			defer mock.AssertExpectationsForObjects(t, notificationsBuilder)

			notificationGenerator := formation.NewNotificationsGenerator(notificationsBuilder)
			formationNotificationReqs, err := notificationGenerator.GenerateFormationLifecycleNotifications(ctx, formationLifecycleSyncWebhooks, TntInternalID, formationInput, testFormationTemplateName, FormationTemplateID, model.CreateFormation, customerTenantContext)

			if testCase.expectedErrMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.expectedErrMsg)
				require.Empty(t, formationNotificationReqs)
			} else {
				require.NoError(t, err)
				require.ElementsMatch(t, formationNotificationReqs, testCase.expectedFormationNotificationReqs)
			}
		})
	}
}
