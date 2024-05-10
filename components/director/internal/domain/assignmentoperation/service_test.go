package assignmentoperation_test

import (
	"context"
	"errors"
	"fmt"
	assignmentOperation "github.com/kyma-incubator/compass/components/director/internal/domain/assignmentoperation"
	"github.com/kyma-incubator/compass/components/director/internal/domain/assignmentoperation/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestService_Create(t *testing.T) {
	ctx := context.TODO()

	testErr := errors.New("test error")

	assignmentOperationModel := mock.MatchedBy(func(op *model.AssignmentOperation) bool {
		return op.ID == operationID && op.Type == operationType && op.FormationAssignmentID == assignmentID && op.FormationID == formationID && op.TriggeredBy == operationTrigger && op.StartedAtTimestamp != nil && op.FinishedAtTimestamp == nil
	})

	testCases := []struct {
		Name                     string
		AssignmentOperationInput *model.AssignmentOperationInput
		AssignmentOperationRepo  func() *automock.AssignmentOperationRepository
		ExpectedOutput           string
		ExpectedErrorMsg         string
	}{
		{
			Name:                     "Success",
			AssignmentOperationInput: fixAssignmentOperationInput(),
			AssignmentOperationRepo: func() *automock.AssignmentOperationRepository {
				repo := &automock.AssignmentOperationRepository{}
				repo.On("Create", ctx, assignmentOperationModel).Return(nil).Once()
				return repo
			},
			ExpectedOutput: operationID,
		},
		{
			Name:                     "Error when creating assignment operation",
			AssignmentOperationInput: fixAssignmentOperationInput(),
			AssignmentOperationRepo: func() *automock.AssignmentOperationRepository {
				repo := &automock.AssignmentOperationRepository{}
				repo.On("Create", ctx, assignmentOperationModel).Return(testErr).Once()
				return repo
			},
			ExpectedOutput:   "",
			ExpectedErrorMsg: fmt.Sprintf("while creating assignment operation for formation assignment %s in formation %s with type %s, triggered by %s", assignmentID, formationID, operationType, operationTrigger),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			aoRepo := &automock.AssignmentOperationRepository{}
			if testCase.AssignmentOperationRepo != nil {
				aoRepo = testCase.AssignmentOperationRepo()
			}

			uuidSvc := fixUUIDService()

			svc := assignmentOperation.NewService(aoRepo, uuidSvc)

			// WHEN
			r, err := svc.Create(ctx, testCase.AssignmentOperationInput)

			if testCase.ExpectedErrorMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
			} else {
				require.NoError(t, err)
			}

			// THEN
			require.Equal(t, testCase.ExpectedOutput, r)

			mock.AssertExpectationsForObjects(t, aoRepo, uuidSvc)
		})
	}
}

func TestService_Finish(t *testing.T) {
	ctx := context.TODO()

	testErr := errors.New("test error")

	assignmentOperationModel := mock.MatchedBy(func(op *model.AssignmentOperation) bool {
		return op.ID == operationID && op.Type == operationType && op.FormationAssignmentID == assignmentID && op.FormationID == formationID && op.TriggeredBy == operationTrigger && op.StartedAtTimestamp != nil && op.FinishedAtTimestamp != nil
	})

	testCases := []struct {
		Name  string
		Input struct {
			assignmentID string
			formationID  string
		}
		AssignmentOperationRepo func() *automock.AssignmentOperationRepository
		ExpectedErrorMsg        string
	}{
		{
			Name: "Success",
			Input: struct {
				assignmentID string
				formationID  string
			}{
				assignmentID: assignmentID,
				formationID:  formationID,
			},
			AssignmentOperationRepo: func() *automock.AssignmentOperationRepository {
				repo := &automock.AssignmentOperationRepository{}
				repo.On("GetLatestOperation", ctx, assignmentID, formationID).Return(fixAssignmentOperationModel(), nil).Once()
				repo.On("Update", ctx, assignmentOperationModel).Return(nil).Once()
				return repo
			},
		},
		{
			Name: "Error when updating assignment operation",
			Input: struct {
				assignmentID string
				formationID  string
			}{
				assignmentID: assignmentID,
				formationID:  formationID,
			},
			AssignmentOperationRepo: func() *automock.AssignmentOperationRepository {
				repo := &automock.AssignmentOperationRepository{}
				repo.On("GetLatestOperation", ctx, assignmentID, formationID).Return(fixAssignmentOperationModel(), nil).Once()
				repo.On("Update", ctx, assignmentOperationModel).Return(testErr).Once()
				return repo
			},
			ExpectedErrorMsg: fmt.Sprintf("while updating the finished at timestamp for assignment operation with ID: %s", operationID),
		},
		{
			Name: "Error when getting assignment operation",
			Input: struct {
				assignmentID string
				formationID  string
			}{
				assignmentID: assignmentID,
				formationID:  formationID,
			},
			AssignmentOperationRepo: func() *automock.AssignmentOperationRepository {
				repo := &automock.AssignmentOperationRepository{}
				repo.On("GetLatestOperation", ctx, assignmentID, formationID).Return(nil, testErr).Once()
				return repo
			},
			ExpectedErrorMsg: fmt.Sprintf("while getting the latest operation for assignment with ID: %s, formation with ID: %s", assignmentID, formationID),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			aoRepo := &automock.AssignmentOperationRepository{}
			if testCase.AssignmentOperationRepo != nil {
				aoRepo = testCase.AssignmentOperationRepo()
			}

			svc := assignmentOperation.NewService(aoRepo, nil)

			// WHEN
			input := testCase.Input
			err := svc.Finish(ctx, input.assignmentID, input.formationID)

			if testCase.ExpectedErrorMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, aoRepo)
		})
	}
}

func TestService_Update(t *testing.T) {
	ctx := context.TODO()

	testErr := errors.New("test error")

	assignmentOperationModel := mock.MatchedBy(func(op *model.AssignmentOperation) bool {
		return op.ID == operationID && op.Type == operationType && op.FormationAssignmentID == assignmentID && op.FormationID == formationID && op.TriggeredBy == newOperationTrigger && op.StartedAtTimestamp != nil && op.FinishedAtTimestamp == nil
	})

	testCases := []struct {
		Name  string
		Input struct {
			assignmentID string
			formationID  string
			newTrigger   model.OperationTrigger
		}
		AssignmentOperationRepo func() *automock.AssignmentOperationRepository
		ExpectedErrorMsg        string
	}{
		{
			Name: "Success",
			Input: struct {
				assignmentID string
				formationID  string
				newTrigger   model.OperationTrigger
			}{
				assignmentID: assignmentID,
				formationID:  formationID,
				newTrigger:   newOperationTrigger,
			},
			AssignmentOperationRepo: func() *automock.AssignmentOperationRepository {
				repo := &automock.AssignmentOperationRepository{}
				repo.On("GetLatestOperation", ctx, assignmentID, formationID).Return(fixAssignmentOperationModelWithoutFinishedAt(), nil).Once()
				repo.On("Update", ctx, assignmentOperationModel).Return(nil).Once()
				return repo
			},
		},
		{
			Name: "Error when updating assignment operation",
			Input: struct {
				assignmentID string
				formationID  string
				newTrigger   model.OperationTrigger
			}{
				assignmentID: assignmentID,
				formationID:  formationID,
				newTrigger:   newOperationTrigger,
			},
			AssignmentOperationRepo: func() *automock.AssignmentOperationRepository {
				repo := &automock.AssignmentOperationRepository{}
				repo.On("GetLatestOperation", ctx, assignmentID, formationID).Return(fixAssignmentOperationModelWithoutFinishedAt(), nil).Once()
				repo.On("Update", ctx, assignmentOperationModel).Return(testErr).Once()
				return repo
			},
			ExpectedErrorMsg: fmt.Sprintf("while updating the finished at timestamp for assignment operation with ID: %s", operationID),
		},
		{
			Name: "Error when getting assignment operation",
			Input: struct {
				assignmentID string
				formationID  string
				newTrigger   model.OperationTrigger
			}{
				assignmentID: assignmentID,
				formationID:  formationID,
				newTrigger:   newOperationTrigger,
			},
			AssignmentOperationRepo: func() *automock.AssignmentOperationRepository {
				repo := &automock.AssignmentOperationRepository{}
				repo.On("GetLatestOperation", ctx, assignmentID, formationID).Return(nil, testErr).Once()
				return repo
			},
			ExpectedErrorMsg: fmt.Sprintf("while getting the latest operation for assignment with ID: %s, formation with ID: %s", assignmentID, formationID),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			aoRepo := &automock.AssignmentOperationRepository{}
			if testCase.AssignmentOperationRepo != nil {
				aoRepo = testCase.AssignmentOperationRepo()
			}

			svc := assignmentOperation.NewService(aoRepo, nil)

			// WHEN
			input := testCase.Input
			err := svc.Update(ctx, input.assignmentID, input.formationID, input.newTrigger)

			if testCase.ExpectedErrorMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, aoRepo)
		})
	}
}

func TestService_ListForFormationAssignmentIDs(t *testing.T) {
	ctx := context.TODO()

	testErr := errors.New("test error")
	pageSize := 200
	cursor := "after"

	assignmentOperationIDs := []string{assignmentID}

	testCases := []struct {
		Name                    string
		Input                   []string
		PageSize                int
		AssignmentOperationRepo func() *automock.AssignmentOperationRepository
		ExpectedOutput          []*model.AssignmentOperationPage
		ExpectedErrorMsg        string
	}{
		{
			Name:     "Success",
			Input:    assignmentOperationIDs,
			PageSize: pageSize,
			AssignmentOperationRepo: func() *automock.AssignmentOperationRepository {
				repo := &automock.AssignmentOperationRepository{}
				repo.On("ListForFormationAssignmentIDs", ctx, assignmentOperationIDs, pageSize, cursor).Return([]*model.AssignmentOperationPage{fixAssignmentOperationPage()}, nil).Once()
				return repo
			},
			ExpectedOutput: []*model.AssignmentOperationPage{fixAssignmentOperationPage()},
		},
		{
			Name:     "Error when updating assignment operation",
			Input:    assignmentOperationIDs,
			PageSize: pageSize,
			AssignmentOperationRepo: func() *automock.AssignmentOperationRepository {
				repo := &automock.AssignmentOperationRepository{}
				repo.On("ListForFormationAssignmentIDs", ctx, assignmentOperationIDs, pageSize, cursor).Return(nil, testErr).Once()
				return repo
			},
			ExpectedErrorMsg: testErr.Error(),
		},
		{
			Name:             "Error when page size is invalid",
			Input:            assignmentOperationIDs,
			PageSize:         900,
			ExpectedErrorMsg: "Invalid data [reason=page size must be between 1 and 200]",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			aoRepo := &automock.AssignmentOperationRepository{}
			if testCase.AssignmentOperationRepo != nil {
				aoRepo = testCase.AssignmentOperationRepo()
			}

			svc := assignmentOperation.NewService(aoRepo, nil)

			// WHEN
			output, err := svc.ListByFormationAssignmentIDs(ctx, testCase.Input, testCase.PageSize, "after")

			if testCase.ExpectedErrorMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, testCase.ExpectedOutput, output)

			mock.AssertExpectationsForObjects(t, aoRepo)
		})
	}
}

func TestService_DeleteByIDs(t *testing.T) {
	ctx := context.TODO()

	testErr := errors.New("test error")

	assignmentOperationIDs := []string{assignmentID}

	testCases := []struct {
		Name                    string
		Input                   []string
		AssignmentOperationRepo func() *automock.AssignmentOperationRepository
		ExpectedErrorMsg        string
	}{
		{
			Name:  "Success",
			Input: assignmentOperationIDs,
			AssignmentOperationRepo: func() *automock.AssignmentOperationRepository {
				repo := &automock.AssignmentOperationRepository{}
				repo.On("DeleteByIDs", ctx, assignmentOperationIDs).Return(nil).Once()
				return repo
			},
		},
		{
			Name:  "Error when updating assignment operation",
			Input: assignmentOperationIDs,
			AssignmentOperationRepo: func() *automock.AssignmentOperationRepository {
				repo := &automock.AssignmentOperationRepository{}
				repo.On("DeleteByIDs", ctx, assignmentOperationIDs).Return(testErr).Once()
				return repo
			},
			ExpectedErrorMsg: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			aoRepo := &automock.AssignmentOperationRepository{}
			if testCase.AssignmentOperationRepo != nil {
				aoRepo = testCase.AssignmentOperationRepo()
			}

			svc := assignmentOperation.NewService(aoRepo, nil)

			// WHEN
			err := svc.DeleteByIDs(ctx, testCase.Input)

			if testCase.ExpectedErrorMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, aoRepo)
		})
	}
}
