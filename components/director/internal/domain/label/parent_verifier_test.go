package label_test

import (
	"context"
	"errors"
	"github.com/kyma-incubator/compass/components/director/internal/domain/label"
	"github.com/kyma-incubator/compass/components/director/internal/domain/label/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParentAccessVerifier_Verify(t *testing.T) {
	tnt := tenantID
	externalTnt := "external-tenant"
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)
	globalSubaccountIDLabelKey := "global_subaccount_id"
	labelID := "label-id"
	testError := errors.New("test error")
	labelModel := &model.Label{
		ID:    labelID,
		Key:   globalSubaccountIDLabelKey,
		Value: externalTnt,
	}

	testCases := []struct {
		Name               string
		LabelRepoFn        func() *automock.LabelRepository
		InputResourceType  resource.Type
		InputObjectID      string
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("GetByKeyGlobal", ctx, model.AppTemplateLabelableObject, "app-template-id", globalSubaccountIDLabelKey).Return(labelModel, nil).Once()
				return repo
			},
			InputResourceType:  resource.ApplicationTemplate,
			InputObjectID:      "app-template-id",
			ExpectedErrMessage: "",
		},
		{
			Name: "Cannot map resourceType to labelable object",
			LabelRepoFn: func() *automock.LabelRepository {
				return &automock.LabelRepository{}
			},
			InputResourceType:  resource.Application,
			InputObjectID:      "app-template-id",
			ExpectedErrMessage: "unknown labelable object for resource",
		},
		{
			Name: "Fail while getting label by key",
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("GetByKeyGlobal", ctx, model.AppTemplateLabelableObject, "app-template-id", globalSubaccountIDLabelKey).Return(nil, testError).Once()
				return repo
			},
			InputResourceType:  resource.ApplicationTemplate,
			InputObjectID:      "app-template-id",
			ExpectedErrMessage: "cannot retrieve \"global_subaccount_id\" label for parent of type applicationTemplate with id app-template-id",
		},
		{
			Name: "Fail when label does not exist",
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("GetByKeyGlobal", ctx, model.AppTemplateLabelableObject, "app-template-id", globalSubaccountIDLabelKey).Return(nil, apperrors.NewNotFoundError(resource.Label, labelID)).Once()
				return repo
			},
			InputResourceType:  resource.ApplicationTemplate,
			InputObjectID:      "app-template-id",
			ExpectedErrMessage: "the parent of type applicationTemplate with id app-template-id does not have \"global_subaccount_id\" label",
		},
		{
			Name: "Fail when the provided tenant and the parent tenant do not match",
			LabelRepoFn: func() *automock.LabelRepository {
				labelModelWithWrongTenant := &model.Label{
					ID:    labelID,
					Key:   globalSubaccountIDLabelKey,
					Value: "wrong-tenant",
				}

				repo := &automock.LabelRepository{}
				repo.On("GetByKeyGlobal", ctx, model.AppTemplateLabelableObject, "app-template-id", globalSubaccountIDLabelKey).Return(labelModelWithWrongTenant, nil).Once()
				return repo
			},
			InputResourceType:  resource.ApplicationTemplate,
			InputObjectID:      "app-template-id",
			ExpectedErrMessage: "the provided tenant external-tenant and the parent tenant wrong-tenant do not match",
		},
		{
			Name: "Fail when the label value has unexpected type",
			LabelRepoFn: func() *automock.LabelRepository {
				labelModelWithWrongTenant := &model.Label{
					ID:    labelID,
					Key:   globalSubaccountIDLabelKey,
					Value: []string{},
				}

				repo := &automock.LabelRepository{}
				repo.On("GetByKeyGlobal", ctx, model.AppTemplateLabelableObject, "app-template-id", globalSubaccountIDLabelKey).Return(labelModelWithWrongTenant, nil).Once()
				return repo
			},
			InputResourceType:  resource.ApplicationTemplate,
			InputObjectID:      "app-template-id",
			ExpectedErrMessage: "unexpected type of \"global_subaccount_id\" label",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			labelRepo := testCase.LabelRepoFn()
			defer labelRepo.AssertExpectations(t)

			verifier := label.NewParentAccessVerifier(labelRepo)

			// WHEN
			err := verifier.Verify(ctx, testCase.InputResourceType, testCase.InputObjectID)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

		})
	}
}
