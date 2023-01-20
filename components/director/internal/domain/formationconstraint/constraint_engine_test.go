package formationconstraint_test

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	formationconstraint2 "github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestConstraintEngine_EnforceConstraints(t *testing.T) {
	// GIVEN
	ctx := context.TODO()

	//testErr := errors.New("test error")

	testCases := []struct {
		Name                       string
		Location                   formationconstraint.JoinPointLocation
		Details                    formationconstraint2.JoinPointDetails
		FormationConstraintService func() *automock.FormationConstraintSvc
		ExpectedError              error
	}{
		{
			Name:     "Success",
			Location: location,
			Details:  &details,
			FormationConstraintService: func() *automock.FormationConstraintSvc {
				svc := &automock.FormationConstraintSvc{}
				svc.On("ListMatchingConstraints", ctx, formationTemplateID, location, details.GetMatchingDetails()).Return([]*model.FormationConstraint{formationConstraintModel}, nil).Once()
				return svc
			},
			ExpectedError: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			formationConstraintSvc := testCase.FormationConstraintService()

			engine := formationconstraint.NewConstraintEngine(formationConstraintSvc, nil, nil, nil, nil)
			engine.Set()
			// WHEN
			err := engine.EnforceConstraints(ctx, testCase.Location, testCase.Details, formationTemplateID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, formationConstraintSvc)
		})
	}
}
