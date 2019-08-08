package labeldef_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/labeldef"
	"github.com/kyma-incubator/compass/components/director/internal/domain/labeldef/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/tenant"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScenariosService_EnsureScenariosLabelDefinitionExists(t *testing.T) {
	testErr := errors.New("Test error")
	id := "foo"

	tnt := "tenant"
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt)

	var scenariosSchema interface{} = model.ScenariosSchema
	scenariosLD := model.LabelDefinition{
		ID:     id,
		Tenant: tnt,
		Key:    model.ScenariosKey,
		Schema: &scenariosSchema,
	}

	testCases := []struct {
		Name           string
		LabelDefRepoFn func() *automock.Repository
		UIDServiceFn   func() *automock.UIDService
		ExpectedErr    error
	}{
		{
			Name: "Success",
			LabelDefRepoFn: func() *automock.Repository {
				repo := &automock.Repository{}
				repo.On("Exists", contextThatHasTenant(tnt), tnt, model.ScenariosKey).Return(true, nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				return svc
			},
			ExpectedErr: nil,
		},
		{
			Name: "Success when scenarios label definition does not exist",
			LabelDefRepoFn: func() *automock.Repository {
				repo := &automock.Repository{}
				repo.On("Exists", contextThatHasTenant(tnt), tnt, model.ScenariosKey).Return(false, nil).Once()
				repo.On("Create", contextThatHasTenant(tnt), scenariosLD).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			ExpectedErr: nil,
		},
		{
			Name: "Returns error when checking if label definition exists failed",
			LabelDefRepoFn: func() *automock.Repository {
				repo := &automock.Repository{}
				repo.On("Exists", contextThatHasTenant(tnt), tnt, model.ScenariosKey).Return(false, testErr).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				return svc
			},
			ExpectedErr: testErr,
		},
		{
			Name: "Returns error when creating scenarios label definition failed",
			LabelDefRepoFn: func() *automock.Repository {
				repo := &automock.Repository{}
				repo.On("Exists", contextThatHasTenant(tnt), tnt, model.ScenariosKey).Return(false, nil).Once()
				repo.On("Create", contextThatHasTenant(tnt), scenariosLD).Return(testErr).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ldRepo := testCase.LabelDefRepoFn()
			uidSvc := testCase.UIDServiceFn()
			svc := labeldef.NewScenariosService(ldRepo, uidSvc)

			// when
			err := svc.EnsureScenariosLabelDefinitionExists(ctx, tnt)

			// then
			if testCase.ExpectedErr != nil {
				require.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				require.Nil(t, err)
			}

			ldRepo.AssertExpectations(t)
			uidSvc.AssertExpectations(t)
		})
	}
}
