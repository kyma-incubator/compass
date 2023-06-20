package datainputbuilder_test

import (
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"testing"

	databuilder "github.com/kyma-incubator/compass/components/director/internal/domain/webhook/datainputbuilder"
	"github.com/kyma-incubator/compass/components/director/internal/domain/webhook/datainputbuilder/automock"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestWebhookLabelBuilder_GetLabelsForObject(t *testing.T) {
	testCases := []struct {
		name           string
		labelRepo      func() *automock.LabelRepository
		objectID       string
		objectType     model.LabelableObject
		expectedLabels map[string]string
		expectedErrMsg string
	}{
		{
			name: "success",
			labelRepo: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", emptyCtx, testTenantID, model.ApplicationLabelableObject, ApplicationID).Return(testLabels, nil).Once()
				return repo
			},
			objectType:     model.ApplicationLabelableObject,
			objectID:       ApplicationID,
			expectedLabels: convertLabels(testLabels),
		},
		{
			name: "success when fails to unquote label",
			labelRepo: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", emptyCtx, testTenantID, model.ApplicationLabelableObject, ApplicationID).Return(testLabelsComposite, nil).Once()
				return repo
			},
			objectType:     model.ApplicationLabelableObject,
			objectID:       ApplicationID,
			expectedLabels: convertLabels(testLabelsComposite),
		},
		{
			name: "error when fails during listing labels",
			labelRepo: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", emptyCtx, testTenantID, model.ApplicationLabelableObject, ApplicationID).Return(nil, testErr).Once()
				return repo
			},
			objectType:     model.ApplicationLabelableObject,
			objectID:       ApplicationID,
			expectedErrMsg: testErr.Error(),
		},
	}

	for _, tCase := range testCases {
		t.Run(tCase.name, func(t *testing.T) {
			// GIVEN
			labelRepo := tCase.labelRepo()
			defer mock.AssertExpectationsForObjects(t, labelRepo)

			webhookDataInputBuilder := databuilder.NewWebhookLabelBuilder(labelRepo)

			// WHEN
			resultLabels, err := webhookDataInputBuilder.GetLabelsForObject(emptyCtx, testTenantID, tCase.objectID, tCase.objectType)

			// THEN
			if tCase.expectedErrMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tCase.expectedErrMsg)
				require.Nil(t, resultLabels)
			} else {
				require.NoError(t, err)
				require.Equal(t, tCase.expectedLabels, resultLabels)
			}
		})
	}
}

func TestWebhookLabelBuilder_GetLabelsForObjects(t *testing.T) {
	testCases := []struct {
		name           string
		labelRepo      func() *automock.LabelRepository
		objectIDs      []string
		objectType     model.LabelableObject
		expectedLabels map[string]map[string]string
		expectedErrMsg string
	}{
		{
			name: "success",
			labelRepo: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObjectIDs", emptyCtx, testTenantID, model.ApplicationLabelableObject, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{ApplicationID, Application2ID}) })).Return(map[string]map[string]interface{}{
					ApplicationID:  fixApplicationLabelsMap(),
					Application2ID: fixApplicationLabelsMap(),
				}, nil).Once()
				return repo
			},
			objectType: model.ApplicationLabelableObject,
			objectIDs:  []string{ApplicationID, Application2ID},
			expectedLabels: map[string]map[string]string{
				ApplicationID:  fixLabelsMapForApplicationWithLabels(),
				Application2ID: fixLabelsMapForApplicationWithLabels(),
			},
		},
		{
			name: "success when fails to unquote label",
			labelRepo: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObjectIDs", emptyCtx, testTenantID, model.ApplicationLabelableObject, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{ApplicationID, Application2ID}) })).Return(map[string]map[string]interface{}{
					ApplicationID:  fixApplicationLabelsMapWithUnquotableLabels(),
					Application2ID: fixApplicationLabelsMapWithUnquotableLabels(),
				}, nil).Once()
				return repo
			},
			objectType: model.ApplicationLabelableObject,
			objectIDs:  []string{ApplicationID, Application2ID},
			expectedLabels: map[string]map[string]string{
				ApplicationID:  fixLabelsMapForApplicationWithCompositeLabels(),
				Application2ID: fixLabelsMapForApplicationWithCompositeLabels(),
			},
		},
		{
			name: "error when fails during listing labels",
			labelRepo: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObjectIDs", emptyCtx, testTenantID, model.ApplicationLabelableObject, []string{ApplicationID}).Return(nil, testErr).Once()
				return repo
			},
			objectType:     model.ApplicationLabelableObject,
			objectIDs:      []string{ApplicationID},
			expectedErrMsg: testErr.Error(),
		},
	}

	for _, tCase := range testCases {
		t.Run(tCase.name, func(t *testing.T) {
			// GIVEN
			labelRepo := tCase.labelRepo()
			defer mock.AssertExpectationsForObjects(t, labelRepo)

			webhookDataInputBuilder := databuilder.NewWebhookLabelBuilder(labelRepo)

			// WHEN
			resultLabels, err := webhookDataInputBuilder.GetLabelsForObjects(emptyCtx, testTenantID, tCase.objectIDs, tCase.objectType)

			// THEN
			if tCase.expectedErrMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tCase.expectedErrMsg)
				require.Nil(t, resultLabels)
			} else {
				require.NoError(t, err)
				require.Equal(t, tCase.expectedLabels, resultLabels)
			}
		})
	}
}
