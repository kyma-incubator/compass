package formationconstraint_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	formationconstraint2 "github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestConstraintEngine_EnforceConstraints(t *testing.T) {
	// GIVEN
	ctx := context.TODO()

	testErr := errors.New("test error")

	testCases := []struct {
		Name                         string
		Location                     formationconstraint2.JoinPointLocation
		Details                      formationconstraint2.JoinPointDetails
		OperatorFunc                 func(ctx context.Context, input formationconstraint.OperatorInput) (bool, error)
		SetEmptyOperatorInputBuilder bool
		FormationConstraintService   func() *automock.FormationConstraintSvc
		ExpectedErrorMsg             string
	}{
		{
			Name:     "Success",
			Location: location,
			Details:  &details,
			OperatorFunc: func(ctx context.Context, input formationconstraint.OperatorInput) (bool, error) {
				return true, nil
			},
			FormationConstraintService: func() *automock.FormationConstraintSvc {
				svc := &automock.FormationConstraintSvc{}
				svc.On("ListMatchingConstraints", ctx, formationTemplateID, location, details.GetMatchingDetails()).Return([]*model.FormationConstraint{formationConstraintModel}, nil).Once()
				return svc
			},
			ExpectedErrorMsg: "",
		},
		{
			Name:     "Error while listing matching constraints",
			Location: location,
			Details:  &details,
			FormationConstraintService: func() *automock.FormationConstraintSvc {
				svc := &automock.FormationConstraintSvc{}
				svc.On("ListMatchingConstraints", ctx, formationTemplateID, location, details.GetMatchingDetails()).Return(nil, testErr).Once()
				return svc
			},
			ExpectedErrorMsg: "While listing matching constraints for target operation",
		},
		{
			Name:     "Error when operator not found",
			Location: location,
			Details:  &details,
			FormationConstraintService: func() *automock.FormationConstraintSvc {
				svc := &automock.FormationConstraintSvc{}
				svc.On("ListMatchingConstraints", ctx, formationTemplateID, location, details.GetMatchingDetails()).Return([]*model.FormationConstraint{formationConstraintUnsupportedOperatorModel}, nil).Once()
				return svc
			},
			ExpectedErrorMsg: "Operator \"unsupported\" not found",
		},
		{
			Name:     "Error when operator input builder not found",
			Location: location,
			Details:  &details,
			FormationConstraintService: func() *automock.FormationConstraintSvc {
				svc := &automock.FormationConstraintSvc{}
				svc.On("ListMatchingConstraints", ctx, formationTemplateID, location, details.GetMatchingDetails()).Return([]*model.FormationConstraint{formationConstraintModel}, nil).Once()
				return svc
			},
			OperatorFunc: func(ctx context.Context, input formationconstraint.OperatorInput) (bool, error) {
				return true, nil
			},
			SetEmptyOperatorInputBuilder: true,
			ExpectedErrorMsg:             "Operator input constructor for operator \"IsNotAssignedToAnyFormationOfType\" not found",
		},
		{
			Name:     "Error when executing operator",
			Location: location,
			Details:  &details,
			FormationConstraintService: func() *automock.FormationConstraintSvc {
				svc := &automock.FormationConstraintSvc{}
				svc.On("ListMatchingConstraints", ctx, formationTemplateID, location, details.GetMatchingDetails()).Return([]*model.FormationConstraint{formationConstraintModel}, nil).Once()
				return svc
			},
			OperatorFunc: func(ctx context.Context, input formationconstraint.OperatorInput) (bool, error) {
				return false, testErr
			},
			ExpectedErrorMsg: "An error occurred while executing operator",
		},
		{
			Name:     "Error when operator is not satisfied",
			Location: location,
			Details:  &details,
			FormationConstraintService: func() *automock.FormationConstraintSvc {
				svc := &automock.FormationConstraintSvc{}
				svc.On("ListMatchingConstraints", ctx, formationTemplateID, location, details.GetMatchingDetails()).Return([]*model.FormationConstraint{formationConstraintModel}, nil).Once()
				return svc
			},
			OperatorFunc: func(ctx context.Context, input formationconstraint.OperatorInput) (bool, error) {
				return false, nil
			},
			ExpectedErrorMsg: "Operator \"IsNotAssignedToAnyFormationOfType\" is not satisfied",
		},
		{
			Name:     "Error while parsing input template",
			Location: location,
			Details:  &details,
			FormationConstraintService: func() *automock.FormationConstraintSvc {
				svc := &automock.FormationConstraintSvc{}
				svc.On("ListMatchingConstraints", ctx, formationTemplateID, location, details.GetMatchingDetails()).Return([]*model.FormationConstraint{{Operator: operatorName, InputTemplate: "{invalid template"}}, nil).Once()
				return svc
			},
			OperatorFunc: func(ctx context.Context, input formationconstraint.OperatorInput) (bool, error) {
				return false, nil
			},
			ExpectedErrorMsg: "Failed to parse operator input template for operator",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			formationConstraintSvc := testCase.FormationConstraintService()

			engine := formationconstraint.NewConstraintEngine(formationConstraintSvc, nil, nil, nil, nil)
			if testCase.OperatorFunc != nil {
				engine.SetOperator(testCase.OperatorFunc)
			} else {
				engine.SetEmptyOperatorMap()
			}

			if testCase.SetEmptyOperatorInputBuilder {
				engine.SetEmptyOperatorInputBuilderMap()
			}

			// WHEN
			err := engine.EnforceConstraints(ctx, testCase.Location, testCase.Details, formationTemplateID)

			// THEN
			if testCase.ExpectedErrorMsg != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
			} else {
				assert.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, formationConstraintSvc)
		})
	}
}
