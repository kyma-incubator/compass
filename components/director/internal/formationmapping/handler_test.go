package formationmapping_test

// todo::: revert and fix unit test
//func TestHandler_UpdateFormationAssignmentStatus(t *testing.T) {
//	url := fmt.Sprintf("/v1/businessIntegrations/{%s}/assignments/{%s}/status", fm.FormationIDParam, fm.FormationAssignmentIDParam)
//	testValidConfig := `{"testK":"testV"}`
//	urlVars := map[string]string{
//		fm.FormationIDParam:           testFormationID,
//		fm.FormationAssignmentIDParam: testFormationAssignmentID,
//	}
//	configurationErr := errors.New("formation assignment configuration error")
//
//	// formation assignment fixtures with ASSIGN operation
//	faWithSourceAppAndTargetRuntime := fixFormationAssignmentModelWithStateAndConfig(testFormationAssignmentID, testFormationID, internalTntID, faSourceID, faTargetID, model.FormationAssignmentTypeApplication, model.FormationAssignmentTypeRuntime, model.ReadyAssignmentState, testValidConfig)
//	reverseFAWithSourceRuntimeAndTargetApp := fixFormationAssignmentModelWithStateAndConfig(testFormationAssignmentID, testFormationID, internalTntID, faTargetID, faSourceID, model.FormationAssignmentTypeRuntime, model.FormationAssignmentTypeApplication, model.ReadyAssignmentState, testValidConfig)
//
//	faWithSourceAppAndTargetRuntimeWithCreateErrorState := fixFormationAssignmentModelWithStateAndConfig(testFormationAssignmentID, testFormationID, internalTntID, faSourceID, faTargetID, model.FormationAssignmentTypeApplication, model.FormationAssignmentTypeRuntime, model.CreateErrorAssignmentState, "")
//
//	testFAReqMapping := formationassignment.FormationAssignmentRequestMapping{
//		Request:             fixEmptyNotificationRequest(),
//		FormationAssignment: faWithSourceAppAndTargetRuntime,
//	}
//
//	testReverseFAReqMapping := formationassignment.FormationAssignmentRequestMapping{
//		Request:             fixEmptyNotificationRequest(),
//		FormationAssignment: reverseFAWithSourceRuntimeAndTargetApp,
//	}
//
//	testAssignmentPair := &formationassignment.AssignmentMappingPairWithOperation{
//		AssignmentMappingPair: &formationassignment.AssignmentMappingPair{
//			Assignment:        &testReverseFAReqMapping,
//			ReverseAssignment: &testFAReqMapping,
//		},
//		Operation: model.AssignFormation,
//	}
//
//	// formation assignment fixtures with UNASSIGN operation
//	faWithSourceAppAndTargetRuntimeForUnassingOp := fixFormationAssignmentModelWithStateAndConfig(testFormationAssignmentID, testFormationID, internalTntID, faSourceID, faTargetID, model.FormationAssignmentTypeApplication, model.FormationAssignmentTypeRuntime, model.DeletingAssignmentState, testValidConfig)
//	faWithSourceAppAndTargetRuntimeForUnassingOpWithDeleteErrorState := fixFormationAssignmentModelWithStateAndConfig(testFormationAssignmentID, testFormationID, internalTntID, faSourceID, faTargetID, model.FormationAssignmentTypeApplication, model.FormationAssignmentTypeRuntime, model.DeletingAssignmentState, "")
//
//	faWithSameSourceAppAndTarget := fixFormationAssignmentModelWithStateAndConfig(testFormationAssignmentID, testFormationID, internalTntID, faSourceID, faSourceID, model.FormationAssignmentTypeApplication, model.FormationAssignmentTypeApplication, model.ReadyAssignmentState, "")
//
//	testFormationAssignmentsForObject := []*model.FormationAssignment{
//		{
//			ID: "ID1",
//		},
//		{
//			ID: "ID2",
//		},
//	}
//
//	emptyFormationAssignmentsForObject := []*model.FormationAssignment{}
//
//	testFormationWithReadyState := fixFormationWithState(model.ReadyFormationState)
//	testFormationWithInitialState := fixFormationWithState(model.InitialFormationState)
//
//	testCases := []struct {
//		name                string
//		transactFn          func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
//		faServiceFn         func() *automock.FormationAssignmentService
//		faNotificationSvcFn func() *automock.FormationAssignmentNotificationService
//		formationSvcFn      func() *automock.FormationService
//		faStatusSvcFn       func() *automock.FormationAssignmentStatusService
//		reqBody             fm.FormationAssignmentRequestBody
//		hasURLVars          bool
//		headers             map[string][]string
//		expectedStatusCode  int
//		expectedErrOutput   string
//	}{
//		// Request(+metadata) validation checks
//		{
//			name:               "Decode Error: Content-Type header is not application/json",
//			headers:            map[string][]string{httputils.HeaderContentTypeKey: {"invalidContentType"}},
//			expectedStatusCode: http.StatusUnsupportedMediaType,
//			expectedErrOutput:  "Content-Type header is not application/json",
//		},
//		{
//			name: "Error when one or more of the required path parameters are missing",
//			reqBody: fm.FormationAssignmentRequestBody{
//				State:         model.ReadyAssignmentState,
//				Configuration: json.RawMessage(testValidConfig),
//			},
//			expectedStatusCode: http.StatusBadRequest,
//			expectedErrOutput:  "Not all of the required parameters are provided",
//		},
//		// Request body validation checks
//		{
//			name: "Validate Error: error when we have ready state with config but also an error provided",
//			reqBody: fm.FormationAssignmentRequestBody{
//				State:         model.ReadyAssignmentState,
//				Configuration: json.RawMessage(testValidConfig),
//				Error:         configurationErr.Error(),
//			},
//			expectedStatusCode: http.StatusBadRequest,
//			expectedErrOutput:  "Request Body contains invalid input:",
//		},
//		{
//			name: "Validate Error: error when configuration is provided but the state is incorrect",
//			reqBody: fm.FormationAssignmentRequestBody{
//				State:         model.CreateErrorAssignmentState,
//				Configuration: json.RawMessage(testValidConfig),
//			},
//			expectedStatusCode: http.StatusBadRequest,
//			expectedErrOutput:  "Request Body contains invalid input:",
//		},
//		// Business logic unit tests for assign operation
//		{
//			name:       "Success when operation is assign",
//			transactFn: txGen.ThatSucceeds,
//			faServiceFn: func() *automock.FormationAssignmentService {
//				faSvc := &automock.FormationAssignmentService{}
//				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntime, nil).Once()
//				faSvc.On("GetReverseBySourceAndTarget", contextThatHasTenant(internalTntID), testFormationID, faSourceID, faTargetID).Return(reverseFAWithSourceRuntimeAndTargetApp, nil).Once()
//				faSvc.On("ProcessFormationAssignmentPair", contextThatHasTenant(internalTntID), testAssignmentPair).Return(false, nil).Once()
//				return faSvc
//			},
//			faNotificationSvcFn: func() *automock.FormationAssignmentNotificationService {
//				faNotificationSvc := &automock.FormationAssignmentNotificationService{}
//				faNotificationSvc.On("GenerateFormationAssignmentNotification", contextThatHasTenant(internalTntID), faWithSourceAppAndTargetRuntime).Return(fixEmptyNotificationRequest(), nil).Once()
//				faNotificationSvc.On("GenerateFormationAssignmentNotification", contextThatHasTenant(internalTntID), reverseFAWithSourceRuntimeAndTargetApp).Return(fixEmptyNotificationRequest(), nil).Once()
//				return faNotificationSvc
//			},
//			formationSvcFn: func() *automock.FormationService {
//				formationSvc := &automock.FormationService{}
//				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormationWithReadyState, nil).Once()
//				return formationSvc
//			},
//			faStatusSvcFn: func() *automock.FormationAssignmentStatusService {
//				updater := &automock.FormationAssignmentStatusService{}
//				updater.On("UpdateWithConstraints", txtest.CtxWithDBMatcher(), faWithSourceAppAndTargetRuntime, model.AssignFormation).Return(nil).Once()
//				return updater
//			},
//			reqBody: fm.FormationAssignmentRequestBody{
//				State:         model.ReadyAssignmentState,
//				Configuration: json.RawMessage(testValidConfig),
//			},
//			hasURLVars:         true,
//			expectedStatusCode: http.StatusOK,
//			expectedErrOutput:  "",
//		},
//		{
//			name:       "Success when state is not changed - only configuration is provided",
//			transactFn: txGen.ThatSucceeds,
//			faServiceFn: func() *automock.FormationAssignmentService {
//				faSvc := &automock.FormationAssignmentService{}
//				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntime, nil).Once()
//				faSvc.On("GetReverseBySourceAndTarget", contextThatHasTenant(internalTntID), testFormationID, faSourceID, faTargetID).Return(reverseFAWithSourceRuntimeAndTargetApp, nil).Once()
//				faSvc.On("ProcessFormationAssignmentPair", contextThatHasTenant(internalTntID), testAssignmentPair).Return(false, nil).Once()
//				return faSvc
//			},
//			faNotificationSvcFn: func() *automock.FormationAssignmentNotificationService {
//				faNotificationSvc := &automock.FormationAssignmentNotificationService{}
//				faNotificationSvc.On("GenerateFormationAssignmentNotification", contextThatHasTenant(internalTntID), faWithSourceAppAndTargetRuntime).Return(fixEmptyNotificationRequest(), nil).Once()
//				faNotificationSvc.On("GenerateFormationAssignmentNotification", contextThatHasTenant(internalTntID), reverseFAWithSourceRuntimeAndTargetApp).Return(fixEmptyNotificationRequest(), nil).Once()
//				return faNotificationSvc
//			},
//			formationSvcFn: func() *automock.FormationService {
//				formationSvc := &automock.FormationService{}
//				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormationWithInitialState, nil).Once()
//				return formationSvc
//			},
//			faStatusSvcFn: func() *automock.FormationAssignmentStatusService {
//				updater := &automock.FormationAssignmentStatusService{}
//				updater.On("UpdateWithConstraints", txtest.CtxWithDBMatcher(), faWithSourceAppAndTargetRuntime, model.AssignFormation).Return(nil).Once()
//				return updater
//			},
//			reqBody: fm.FormationAssignmentRequestBody{
//				Configuration: json.RawMessage(testValidConfig),
//			},
//			hasURLVars:         true,
//			expectedStatusCode: http.StatusOK,
//			expectedErrOutput:  "",
//		},
//		{
//			name:       "Error when transaction fails to begin",
//			transactFn: txGen.ThatFailsOnBegin,
//			reqBody: fm.FormationAssignmentRequestBody{
//				State:         model.ReadyAssignmentState,
//				Configuration: json.RawMessage(testValidConfig),
//			},
//			hasURLVars:         true,
//			expectedStatusCode: http.StatusInternalServerError,
//			expectedErrOutput:  "An unexpected error occurred while processing the request. X-Request-Id:",
//		},
//		{
//			name:       "Error when getting formation assignment globally",
//			transactFn: txGen.ThatDoesntExpectCommit,
//			faServiceFn: func() *automock.FormationAssignmentService {
//				faSvc := &automock.FormationAssignmentService{}
//				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(nil, testErr).Once()
//				return faSvc
//			},
//			reqBody: fm.FormationAssignmentRequestBody{
//				State:         model.ReadyAssignmentState,
//				Configuration: json.RawMessage(testValidConfig),
//			},
//			hasURLVars:         true,
//			expectedStatusCode: http.StatusInternalServerError,
//			expectedErrOutput:  "An unexpected error occurred while processing the request. X-Request-Id:",
//		},
//		{
//			name: "Error when getting formation by formation ID fail",
//			transactFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
//				return txGen.ThatDoesntExpectCommit()
//			},
//			faServiceFn: func() *automock.FormationAssignmentService {
//				faSvc := &automock.FormationAssignmentService{}
//				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntimeForUnassingOp, nil).Once()
//				return faSvc
//			},
//			formationSvcFn: func() *automock.FormationService {
//				formationSvc := &automock.FormationService{}
//				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(nil, testErr).Once()
//				return formationSvc
//			},
//			reqBody: fm.FormationAssignmentRequestBody{
//				State:         model.ReadyAssignmentState,
//				Configuration: json.RawMessage(testValidConfig),
//			},
//			hasURLVars:         true,
//			expectedStatusCode: http.StatusInternalServerError,
//			expectedErrOutput:  "An unexpected error occurred while processing the request. X-Request-Id:",
//		},
//		{
//			name: "Error when the retrieved formation is not in READY state",
//			transactFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
//				return txGen.ThatDoesntExpectCommit()
//			},
//			faServiceFn: func() *automock.FormationAssignmentService {
//				faSvc := &automock.FormationAssignmentService{}
//				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntimeForUnassingOp, nil).Once()
//				return faSvc
//			},
//			formationSvcFn: func() *automock.FormationService {
//				formationSvc := &automock.FormationService{}
//				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormationWithInitialState, nil).Once()
//				return formationSvc
//			},
//			reqBody: fm.FormationAssignmentRequestBody{
//				State:         model.ReadyAssignmentState,
//				Configuration: json.RawMessage(testValidConfig),
//			},
//			hasURLVars:         true,
//			expectedStatusCode: http.StatusBadRequest,
//			expectedErrOutput:  "An unexpected error occurred while processing the request. X-Request-Id:",
//		},
//		{
//			name:       "Error when request body state is not correct",
//			transactFn: txGen.ThatDoesntExpectCommit,
//			faServiceFn: func() *automock.FormationAssignmentService {
//				faSvc := &automock.FormationAssignmentService{}
//				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntime, nil).Once()
//				return faSvc
//			},
//			formationSvcFn: func() *automock.FormationService {
//				formationSvc := &automock.FormationService{}
//				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormationWithReadyState, nil).Once()
//				return formationSvc
//			},
//			reqBody: fm.FormationAssignmentRequestBody{
//				State: model.DeleteErrorAssignmentState,
//				Error: configurationErr.Error(),
//			},
//			hasURLVars:         true,
//			expectedStatusCode: http.StatusBadRequest,
//			expectedErrOutput:  fmt.Sprintf("An invalid state: %s is provided for %s operation", model.DeleteErrorAssignmentState, model.AssignFormation),
//		},
//		{
//			name:       "Error when fail to set the formation assignment to error state",
//			transactFn: txGen.ThatDoesntExpectCommit,
//			faServiceFn: func() *automock.FormationAssignmentService {
//				faSvc := &automock.FormationAssignmentService{}
//				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntime, nil).Once()
//				return faSvc
//			},
//			formationSvcFn: func() *automock.FormationService {
//				formationSvc := &automock.FormationService{}
//				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormationWithReadyState, nil).Once()
//				return formationSvc
//			},
//			faStatusSvcFn: func() *automock.FormationAssignmentStatusService {
//				updater := &automock.FormationAssignmentStatusService{}
//				updater.On("SetAssignmentToErrorStateWithConstraints", txtest.CtxWithDBMatcher(), faWithSourceAppAndTargetRuntime, configurationErr.Error(), formationassignment.AssignmentErrorCode(2), model.CreateErrorAssignmentState, model.CreateFormation).Return(testErr).Once()
//				return updater
//			},
//			reqBody: fm.FormationAssignmentRequestBody{
//				State: model.CreateErrorAssignmentState,
//				Error: configurationErr.Error(),
//			},
//			hasURLVars:         true,
//			expectedStatusCode: http.StatusInternalServerError,
//			expectedErrOutput:  "An unexpected error occurred while processing the request. X-Request-Id:",
//		},
//		{
//			name:       "Successfully update formation assignment when input state is create error",
//			transactFn: txGen.ThatSucceeds,
//			faServiceFn: func() *automock.FormationAssignmentService {
//				faSvc := &automock.FormationAssignmentService{}
//				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntimeWithCreateErrorState, nil).Once()
//				return faSvc
//			},
//			formationSvcFn: func() *automock.FormationService {
//				formationSvc := &automock.FormationService{}
//				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormationWithReadyState, nil).Once()
//				return formationSvc
//			},
//			faStatusSvcFn: func() *automock.FormationAssignmentStatusService {
//				updater := &automock.FormationAssignmentStatusService{}
//				updater.On("SetAssignmentToErrorStateWithConstraints", contextThatHasTenant(internalTntID), faWithSourceAppAndTargetRuntimeWithCreateErrorState, configurationErr.Error(), formationassignment.AssignmentErrorCode(formationassignment.ClientError), model.CreateErrorAssignmentState, model.CreateFormation).Return(nil).Once()
//				updater.On("UpdateWithConstraints", contextThatHasTenant(internalTntID), faWithSourceAppAndTargetRuntimeWithCreateErrorState, model.AssignFormation).Return(nil).Once()
//				return updater
//			},
//			reqBody: fm.FormationAssignmentRequestBody{
//				State: model.CreateErrorAssignmentState,
//				Error: configurationErr.Error(),
//			},
//			hasURLVars:         true,
//			expectedStatusCode: http.StatusOK,
//			expectedErrOutput:  "",
//		},
//		{
//			name:       "Error when transaction fail to commit after successful formation assignment update",
//			transactFn: txGen.ThatFailsOnCommit,
//			faServiceFn: func() *automock.FormationAssignmentService {
//				faSvc := &automock.FormationAssignmentService{}
//				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntimeWithCreateErrorState, nil).Once()
//				return faSvc
//			},
//			formationSvcFn: func() *automock.FormationService {
//				formationSvc := &automock.FormationService{}
//				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormationWithReadyState, nil).Once()
//				return formationSvc
//			},
//			faStatusSvcFn: func() *automock.FormationAssignmentStatusService {
//				updater := &automock.FormationAssignmentStatusService{}
//				updater.On("SetAssignmentToErrorStateWithConstraints", contextThatHasTenant(internalTntID), faWithSourceAppAndTargetRuntimeWithCreateErrorState, configurationErr.Error(), formationassignment.AssignmentErrorCode(formationassignment.ClientError), model.CreateErrorAssignmentState, model.CreateFormation).Return(nil).Once()
//				updater.On("UpdateWithConstraints", contextThatHasTenant(internalTntID), faWithSourceAppAndTargetRuntimeWithCreateErrorState, model.AssignFormation).Return(nil).Once()
//				return updater
//			},
//			reqBody: fm.FormationAssignmentRequestBody{
//				State: model.CreateErrorAssignmentState,
//				Error: configurationErr.Error(),
//			},
//			hasURLVars:         true,
//			expectedStatusCode: http.StatusInternalServerError,
//			expectedErrOutput:  "An unexpected error occurred while processing the request. X-Request-Id:",
//		},
//		{
//			name:       "Error when setting formation assignment state to error fail",
//			transactFn: txGen.ThatDoesntExpectCommit,
//			faServiceFn: func() *automock.FormationAssignmentService {
//				faSvc := &automock.FormationAssignmentService{}
//				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntimeWithCreateErrorState, nil).Once()
//				return faSvc
//			},
//			formationSvcFn: func() *automock.FormationService {
//				formationSvc := &automock.FormationService{}
//				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormationWithReadyState, nil).Once()
//				return formationSvc
//			},
//			faStatusSvcFn: func() *automock.FormationAssignmentStatusService {
//				updater := &automock.FormationAssignmentStatusService{}
//				updater.On("SetAssignmentToErrorStateWithConstraints", contextThatHasTenant(internalTntID), faWithSourceAppAndTargetRuntimeWithCreateErrorState, configurationErr.Error(), formationassignment.AssignmentErrorCode(formationassignment.ClientError), model.CreateErrorAssignmentState, model.CreateFormation).Return(nil).Once()
//				updater.On("UpdateWithConstraints", contextThatHasTenant(internalTntID), faWithSourceAppAndTargetRuntimeWithCreateErrorState, model.AssignFormation).Return(testErr).Once()
//				return updater
//			},
//			reqBody: fm.FormationAssignmentRequestBody{
//				State: model.CreateErrorAssignmentState,
//				Error: configurationErr.Error(),
//			},
//			hasURLVars:         true,
//			expectedStatusCode: http.StatusInternalServerError,
//			expectedErrOutput:  "An unexpected error occurred while processing the request. X-Request-Id:",
//		},
//		{
//			name:       "Error when generating notifications for assignment fail",
//			transactFn: txGen.ThatDoesntExpectCommit,
//			faServiceFn: func() *automock.FormationAssignmentService {
//				faSvc := &automock.FormationAssignmentService{}
//				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntime, nil).Once()
//				return faSvc
//			},
//			formationSvcFn: func() *automock.FormationService {
//				formationSvc := &automock.FormationService{}
//				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormationWithReadyState, nil).Once()
//				return formationSvc
//			},
//			faNotificationSvcFn: func() *automock.FormationAssignmentNotificationService {
//				faNotificationSvc := &automock.FormationAssignmentNotificationService{}
//				faNotificationSvc.On("GenerateFormationAssignmentNotification", contextThatHasTenant(internalTntID), faWithSourceAppAndTargetRuntime).Return(nil, testErr).Once()
//				return faNotificationSvc
//			},
//			faStatusSvcFn: func() *automock.FormationAssignmentStatusService {
//				updater := &automock.FormationAssignmentStatusService{}
//				updater.On("UpdateWithConstraints", contextThatHasTenant(internalTntID), faWithSourceAppAndTargetRuntime, model.AssignFormation).Return(nil).Once()
//				return updater
//			},
//			reqBody: fm.FormationAssignmentRequestBody{
//				State:         model.ReadyAssignmentState,
//				Configuration: json.RawMessage(testValidConfig),
//			},
//			hasURLVars:         true,
//			expectedStatusCode: http.StatusInternalServerError,
//			expectedErrOutput:  "An unexpected error occurred while processing the request. X-Request-Id:",
//		},
//		{
//			name:       "Error when getting reverse formation assignment fail",
//			transactFn: txGen.ThatDoesntExpectCommit,
//			faServiceFn: func() *automock.FormationAssignmentService {
//				faSvc := &automock.FormationAssignmentService{}
//				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntime, nil).Once()
//				faSvc.On("GetReverseBySourceAndTarget", contextThatHasTenant(internalTntID), testFormationID, faSourceID, faTargetID).Return(nil, testErr).Once()
//				return faSvc
//			},
//			formationSvcFn: func() *automock.FormationService {
//				formationSvc := &automock.FormationService{}
//				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormationWithReadyState, nil).Once()
//				return formationSvc
//			},
//			faNotificationSvcFn: func() *automock.FormationAssignmentNotificationService {
//				faNotificationSvc := &automock.FormationAssignmentNotificationService{}
//				faNotificationSvc.On("GenerateFormationAssignmentNotification", contextThatHasTenant(internalTntID), faWithSourceAppAndTargetRuntime).Return(fixEmptyNotificationRequest(), nil).Once()
//				return faNotificationSvc
//			},
//			faStatusSvcFn: func() *automock.FormationAssignmentStatusService {
//				updater := &automock.FormationAssignmentStatusService{}
//				updater.On("UpdateWithConstraints", contextThatHasTenant(internalTntID), faWithSourceAppAndTargetRuntime, model.AssignFormation).Return(nil).Once()
//				return updater
//			},
//			reqBody: fm.FormationAssignmentRequestBody{
//				State:         model.ReadyAssignmentState,
//				Configuration: json.RawMessage(testValidConfig),
//			},
//			hasURLVars:         true,
//			expectedStatusCode: http.StatusInternalServerError,
//			expectedErrOutput:  "An unexpected error occurred while processing the request. X-Request-Id:",
//		},
//		{
//			name:       "Error when generating reverse notifications for assignment fail",
//			transactFn: txGen.ThatDoesntExpectCommit,
//			faServiceFn: func() *automock.FormationAssignmentService {
//				faSvc := &automock.FormationAssignmentService{}
//				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntime, nil).Once()
//				faSvc.On("GetReverseBySourceAndTarget", contextThatHasTenant(internalTntID), testFormationID, faSourceID, faTargetID).Return(reverseFAWithSourceRuntimeAndTargetApp, nil).Once()
//				return faSvc
//			},
//			formationSvcFn: func() *automock.FormationService {
//				formationSvc := &automock.FormationService{}
//				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormationWithReadyState, nil).Once()
//				return formationSvc
//			},
//			faNotificationSvcFn: func() *automock.FormationAssignmentNotificationService {
//				faNotificationSvc := &automock.FormationAssignmentNotificationService{}
//				faNotificationSvc.On("GenerateFormationAssignmentNotification", contextThatHasTenant(internalTntID), faWithSourceAppAndTargetRuntime).Return(fixEmptyNotificationRequest(), nil).Once()
//				faNotificationSvc.On("GenerateFormationAssignmentNotification", contextThatHasTenant(internalTntID), reverseFAWithSourceRuntimeAndTargetApp).Return(nil, testErr).Once()
//				return faNotificationSvc
//			},
//			faStatusSvcFn: func() *automock.FormationAssignmentStatusService {
//				updater := &automock.FormationAssignmentStatusService{}
//				updater.On("UpdateWithConstraints", contextThatHasTenant(internalTntID), faWithSourceAppAndTargetRuntime, model.AssignFormation).Return(nil).Once()
//				return updater
//			},
//			reqBody: fm.FormationAssignmentRequestBody{
//				State:         model.ReadyAssignmentState,
//				Configuration: json.RawMessage(testValidConfig),
//			},
//			hasURLVars:         true,
//			expectedStatusCode: http.StatusInternalServerError,
//			expectedErrOutput:  "An unexpected error occurred while processing the request. X-Request-Id:",
//		},
//		{
//			name:       "Error when processing formation assignment pairs",
//			transactFn: txGen.ThatDoesntExpectCommit,
//			faServiceFn: func() *automock.FormationAssignmentService {
//				faSvc := &automock.FormationAssignmentService{}
//				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntime, nil).Once()
//				faSvc.On("GetReverseBySourceAndTarget", contextThatHasTenant(internalTntID), testFormationID, faSourceID, faTargetID).Return(reverseFAWithSourceRuntimeAndTargetApp, nil).Once()
//				faSvc.On("ProcessFormationAssignmentPair", contextThatHasTenant(internalTntID), testAssignmentPair).Return(false, testErr).Once()
//				return faSvc
//			},
//			formationSvcFn: func() *automock.FormationService {
//				formationSvc := &automock.FormationService{}
//				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormationWithReadyState, nil).Once()
//				return formationSvc
//			},
//			faNotificationSvcFn: func() *automock.FormationAssignmentNotificationService {
//				faNotificationSvc := &automock.FormationAssignmentNotificationService{}
//				faNotificationSvc.On("GenerateFormationAssignmentNotification", contextThatHasTenant(internalTntID), faWithSourceAppAndTargetRuntime).Return(fixEmptyNotificationRequest(), nil).Once()
//				faNotificationSvc.On("GenerateFormationAssignmentNotification", contextThatHasTenant(internalTntID), reverseFAWithSourceRuntimeAndTargetApp).Return(fixEmptyNotificationRequest(), nil).Once()
//				return faNotificationSvc
//			},
//			faStatusSvcFn: func() *automock.FormationAssignmentStatusService {
//				updater := &automock.FormationAssignmentStatusService{}
//				updater.On("UpdateWithConstraints", contextThatHasTenant(internalTntID), faWithSourceAppAndTargetRuntime, model.AssignFormation).Return(nil).Once()
//				return updater
//			},
//			reqBody: fm.FormationAssignmentRequestBody{
//				State:         model.ReadyAssignmentState,
//				Configuration: json.RawMessage(testValidConfig),
//			},
//			hasURLVars:         true,
//			expectedStatusCode: http.StatusInternalServerError,
//			expectedErrOutput:  "An unexpected error occurred while processing the request. X-Request-Id:",
//		},
//		{
//			name:       "Error when committing transaction fail",
//			transactFn: txGen.ThatFailsOnCommit,
//			faServiceFn: func() *automock.FormationAssignmentService {
//				faSvc := &automock.FormationAssignmentService{}
//				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntime, nil).Once()
//				faSvc.On("GetReverseBySourceAndTarget", contextThatHasTenant(internalTntID), testFormationID, faSourceID, faTargetID).Return(reverseFAWithSourceRuntimeAndTargetApp, nil).Once()
//				faSvc.On("ProcessFormationAssignmentPair", contextThatHasTenant(internalTntID), testAssignmentPair).Return(false, nil).Once()
//				return faSvc
//			},
//			formationSvcFn: func() *automock.FormationService {
//				formationSvc := &automock.FormationService{}
//				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormationWithReadyState, nil).Once()
//				return formationSvc
//			},
//			faNotificationSvcFn: func() *automock.FormationAssignmentNotificationService {
//				faNotificationSvc := &automock.FormationAssignmentNotificationService{}
//				faNotificationSvc.On("GenerateFormationAssignmentNotification", contextThatHasTenant(internalTntID), faWithSourceAppAndTargetRuntime).Return(fixEmptyNotificationRequest(), nil).Once()
//				faNotificationSvc.On("GenerateFormationAssignmentNotification", contextThatHasTenant(internalTntID), reverseFAWithSourceRuntimeAndTargetApp).Return(fixEmptyNotificationRequest(), nil).Once()
//				return faNotificationSvc
//			},
//			faStatusSvcFn: func() *automock.FormationAssignmentStatusService {
//				updater := &automock.FormationAssignmentStatusService{}
//				updater.On("UpdateWithConstraints", contextThatHasTenant(internalTntID), faWithSourceAppAndTargetRuntime, model.AssignFormation).Return(nil).Once()
//				return updater
//			},
//			reqBody: fm.FormationAssignmentRequestBody{
//				State:         model.ReadyAssignmentState,
//				Configuration: json.RawMessage(testValidConfig),
//			},
//			hasURLVars:         true,
//			expectedStatusCode: http.StatusInternalServerError,
//			expectedErrOutput:  "An unexpected error occurred while processing the request. X-Request-Id:",
//		},
//		// Business logic unit tests for unassign operation
//		{
//			name:       "Success when operation is unassign",
//			transactFn: txGen.ThatSucceedsTwice,
//			faServiceFn: func() *automock.FormationAssignmentService {
//				faSvc := &automock.FormationAssignmentService{}
//				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntimeForUnassingOp, nil).Once()
//				faSvc.On("ListFormationAssignmentsForObjectID", contextThatHasTenant(internalTntID), testFormationID, faSourceID).Return(testFormationAssignmentsForObject, nil).Once()
//				faSvc.On("ListFormationAssignmentsForObjectID", contextThatHasTenant(internalTntID), testFormationID, faTargetID).Return(testFormationAssignmentsForObject, nil).Once()
//				return faSvc
//			},
//			formationSvcFn: func() *automock.FormationService {
//				formationSvc := &automock.FormationService{}
//				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormationWithReadyState, nil).Once()
//				return formationSvc
//			},
//			faStatusSvcFn: func() *automock.FormationAssignmentStatusService {
//				faStatusSvc := &automock.FormationAssignmentStatusService{}
//				faStatusSvc.On("DeleteWithConstraints", contextThatHasTenant(internalTntID), testFormationAssignmentID).Return(nil).Once()
//				return faStatusSvc
//			},
//			reqBody: fm.FormationAssignmentRequestBody{
//				State:         model.ReadyAssignmentState,
//				Configuration: json.RawMessage(testValidConfig),
//			},
//			hasURLVars:         true,
//			expectedStatusCode: http.StatusOK,
//			expectedErrOutput:  "",
//		},
//		{
//			name:       "Error when request body state is not correct",
//			transactFn: txGen.ThatDoesntExpectCommit,
//			faServiceFn: func() *automock.FormationAssignmentService {
//				faSvc := &automock.FormationAssignmentService{}
//				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntimeForUnassingOp, nil).Once()
//				return faSvc
//			},
//			reqBody: fm.FormationAssignmentRequestBody{
//				State:         model.ConfigPendingAssignmentState,
//				Configuration: json.RawMessage(testValidConfig),
//			},
//			formationSvcFn: func() *automock.FormationService {
//				formationSvc := &automock.FormationService{}
//				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormationWithReadyState, nil).Once()
//				return formationSvc
//			},
//			hasURLVars:         true,
//			expectedStatusCode: http.StatusInternalServerError,
//			expectedErrOutput:  "An unexpected error occurred while processing the request. X-Request-Id:",
//		},
//		{
//			name:       "Successfully set the formation assignment to error state",
//			transactFn: txGen.ThatSucceeds,
//			faServiceFn: func() *automock.FormationAssignmentService {
//				faSvc := &automock.FormationAssignmentService{}
//				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntimeForUnassingOpWithDeleteErrorState, nil).Once()
//				return faSvc
//			},
//			reqBody: fm.FormationAssignmentRequestBody{
//				State: model.DeleteErrorAssignmentState,
//				Error: configurationErr.Error(),
//			},
//			formationSvcFn: func() *automock.FormationService {
//				formationSvc := &automock.FormationService{}
//				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormationWithReadyState, nil).Once()
//				return formationSvc
//			},
//			faStatusSvcFn: func() *automock.FormationAssignmentStatusService {
//				updater := &automock.FormationAssignmentStatusService{}
//				updater.On("SetAssignmentToErrorStateWithConstraints", contextThatHasTenant(internalTntID), faWithSourceAppAndTargetRuntimeForUnassingOpWithDeleteErrorState, configurationErr.Error(), formationassignment.AssignmentErrorCode(formationassignment.ClientError), model.DeleteErrorAssignmentState, model.UnassignFormation).Return(nil).Once()
//				return updater
//			},
//			hasURLVars:         true,
//			expectedStatusCode: http.StatusOK,
//			expectedErrOutput:  "",
//		},
//		{
//			name:       "Error when setting formation assignment state to error fail",
//			transactFn: txGen.ThatDoesntExpectCommit,
//			faServiceFn: func() *automock.FormationAssignmentService {
//				faSvc := &automock.FormationAssignmentService{}
//				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntimeForUnassingOpWithDeleteErrorState, nil).Once()
//				return faSvc
//			},
//			formationSvcFn: func() *automock.FormationService {
//				formationSvc := &automock.FormationService{}
//				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormationWithReadyState, nil).Once()
//				return formationSvc
//			},
//			faStatusSvcFn: func() *automock.FormationAssignmentStatusService {
//				updater := &automock.FormationAssignmentStatusService{}
//				updater.On("SetAssignmentToErrorStateWithConstraints", contextThatHasTenant(internalTntID), faWithSourceAppAndTargetRuntimeForUnassingOpWithDeleteErrorState, configurationErr.Error(), formationassignment.AssignmentErrorCode(formationassignment.ClientError), model.DeleteErrorAssignmentState, model.UnassignFormation).Return(testErr).Once()
//				return updater
//			},
//			reqBody: fm.FormationAssignmentRequestBody{
//				State: model.DeleteErrorAssignmentState,
//				Error: configurationErr.Error(),
//			},
//			hasURLVars:         true,
//			expectedStatusCode: http.StatusInternalServerError,
//			expectedErrOutput:  "An unexpected error occurred while processing the request. X-Request-Id:",
//		},
//		{
//			name:       "Error when deleting formation assignment fail",
//			transactFn: txGen.ThatDoesntExpectCommit,
//			faServiceFn: func() *automock.FormationAssignmentService {
//				faSvc := &automock.FormationAssignmentService{}
//				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntimeForUnassingOp, nil).Once()
//				return faSvc
//			},
//			reqBody: fm.FormationAssignmentRequestBody{
//				State:         model.ReadyAssignmentState,
//				Configuration: json.RawMessage(testValidConfig),
//			},
//			formationSvcFn: func() *automock.FormationService {
//				formationSvc := &automock.FormationService{}
//				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormationWithReadyState, nil).Once()
//				return formationSvc
//			},
//			faStatusSvcFn: func() *automock.FormationAssignmentStatusService {
//				faStatusSvc := &automock.FormationAssignmentStatusService{}
//				faStatusSvc.On("DeleteWithConstraints", contextThatHasTenant(internalTntID), testFormationAssignmentID).Return(testErr).Once()
//				return faStatusSvc
//			},
//			hasURLVars:         true,
//			expectedStatusCode: http.StatusInternalServerError,
//			expectedErrOutput:  "An unexpected error occurred while processing the request. X-Request-Id:",
//		},
//		{
//			name: "Error when listing formation assignments for object fail",
//			transactFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
//				return txGen.ThatSucceedsMultipleTimesAndCommitsMultipleTimes(2, 1)
//			},
//			faServiceFn: func() *automock.FormationAssignmentService {
//				faSvc := &automock.FormationAssignmentService{}
//				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntimeForUnassingOp, nil).Once()
//				faSvc.On("ListFormationAssignmentsForObjectID", contextThatHasTenant(internalTntID), testFormationID, faSourceID).Return(nil, testErr).Once()
//				return faSvc
//			},
//			formationSvcFn: func() *automock.FormationService {
//				formationSvc := &automock.FormationService{}
//				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormationWithReadyState, nil).Once()
//				return formationSvc
//			},
//			faStatusSvcFn: func() *automock.FormationAssignmentStatusService {
//				faStatusSvc := &automock.FormationAssignmentStatusService{}
//				faStatusSvc.On("DeleteWithConstraints", contextThatHasTenant(internalTntID), testFormationAssignmentID).Return(nil).Once()
//				return faStatusSvc
//			},
//			reqBody: fm.FormationAssignmentRequestBody{
//				State:         model.ReadyAssignmentState,
//				Configuration: json.RawMessage(testValidConfig),
//			},
//			hasURLVars:         true,
//			expectedStatusCode: http.StatusInternalServerError,
//			expectedErrOutput:  "An unexpected error occurred while processing the request. X-Request-Id:",
//		},
//		{
//			name:       "Successfully unassign formation when there are no formation assignment left",
//			transactFn: txGen.ThatSucceedsTwice,
//			faServiceFn: func() *automock.FormationAssignmentService {
//				faSvc := &automock.FormationAssignmentService{}
//				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntimeForUnassingOp, nil).Once()
//				faSvc.On("ListFormationAssignmentsForObjectID", contextThatHasTenant(internalTntID), testFormationID, faSourceID).Return(emptyFormationAssignmentsForObject, nil).Once()
//				faSvc.On("ListFormationAssignmentsForObjectID", contextThatHasTenant(internalTntID), testFormationID, faTargetID).Return(emptyFormationAssignmentsForObject, nil).Once()
//				return faSvc
//			},
//			formationSvcFn: func() *automock.FormationService {
//				formationSvc := &automock.FormationService{}
//				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormationWithReadyState, nil).Once()
//				formationSvc.On("UnassignFormation", contextThatHasTenant(internalTntID), internalTntID, faSourceID, graphql.FormationObjectType(faWithSourceAppAndTargetRuntimeForUnassingOp.SourceType), *testFormationWithReadyState).Return(testFormationWithReadyState, nil).Once()
//				formationSvc.On("UnassignFormation", contextThatHasTenant(internalTntID), internalTntID, faTargetID, graphql.FormationObjectType(faWithSourceAppAndTargetRuntimeForUnassingOp.TargetType), *testFormationWithReadyState).Return(testFormationWithReadyState, nil).Once()
//				return formationSvc
//			},
//			faStatusSvcFn: func() *automock.FormationAssignmentStatusService {
//				faStatusSvc := &automock.FormationAssignmentStatusService{}
//				faStatusSvc.On("DeleteWithConstraints", contextThatHasTenant(internalTntID), testFormationAssignmentID).Return(nil).Once()
//				return faStatusSvc
//			},
//			reqBody: fm.FormationAssignmentRequestBody{
//				State:         model.ReadyAssignmentState,
//				Configuration: json.RawMessage(testValidConfig),
//			},
//			hasURLVars:         true,
//			expectedStatusCode: http.StatusOK,
//			expectedErrOutput:  "",
//		},
//		{
//			name: "Error when unassigning source from formation fail when there are no formation assignment left",
//			transactFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
//				return txGen.ThatSucceedsMultipleTimesAndCommitsMultipleTimes(2, 1)
//			},
//			faServiceFn: func() *automock.FormationAssignmentService {
//				faSvc := &automock.FormationAssignmentService{}
//				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntimeForUnassingOp, nil).Once()
//				faSvc.On("ListFormationAssignmentsForObjectID", contextThatHasTenant(internalTntID), testFormationID, faSourceID).Return(emptyFormationAssignmentsForObject, nil).Once()
//				return faSvc
//			},
//			formationSvcFn: func() *automock.FormationService {
//				formationSvc := &automock.FormationService{}
//				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormationWithReadyState, nil).Once()
//				formationSvc.On("UnassignFormation", contextThatHasTenant(internalTntID), internalTntID, faSourceID, graphql.FormationObjectType(graphql.FormationAssignmentTypeApplication), *testFormationWithReadyState).Return(nil, testErr).Once()
//				return formationSvc
//			},
//			faStatusSvcFn: func() *automock.FormationAssignmentStatusService {
//				faStatusSvc := &automock.FormationAssignmentStatusService{}
//				faStatusSvc.On("DeleteWithConstraints", contextThatHasTenant(internalTntID), testFormationAssignmentID).Return(nil).Once()
//				return faStatusSvc
//			},
//			reqBody: fm.FormationAssignmentRequestBody{
//				State:         model.ReadyAssignmentState,
//				Configuration: json.RawMessage(testValidConfig),
//			},
//			hasURLVars:         true,
//			expectedStatusCode: http.StatusInternalServerError,
//			expectedErrOutput:  "An unexpected error occurred while processing the request. X-Request-Id:",
//		},
//		{
//			name:       "Error when trying to update formation assignment with same source and target",
//			transactFn: txGen.ThatDoesntExpectCommit,
//			faServiceFn: func() *automock.FormationAssignmentService {
//				faSvc := &automock.FormationAssignmentService{}
//				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSameSourceAppAndTarget, nil).Once()
//				return faSvc
//			},
//			formationSvcFn: func() *automock.FormationService {
//				formationSvc := &automock.FormationService{}
//				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormationWithReadyState, nil).Once()
//				return formationSvc
//			},
//			reqBody: fm.FormationAssignmentRequestBody{
//				State:         model.ReadyAssignmentState,
//				Configuration: json.RawMessage(testValidConfig),
//			},
//			hasURLVars:         true,
//			expectedStatusCode: http.StatusBadRequest,
//			expectedErrOutput:  "Cannot update formation assignment with source ",
//		},
//		{
//			name: "Error when unassigning target from formation fail when there are no formation assignment left",
//			transactFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
//				return txGen.ThatSucceedsMultipleTimesAndCommitsMultipleTimes(2, 1)
//			},
//			faServiceFn: func() *automock.FormationAssignmentService {
//				faSvc := &automock.FormationAssignmentService{}
//				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntimeForUnassingOp, nil).Once()
//				faSvc.On("ListFormationAssignmentsForObjectID", contextThatHasTenant(internalTntID), testFormationID, faSourceID).Return(emptyFormationAssignmentsForObject, nil).Once()
//				faSvc.On("ListFormationAssignmentsForObjectID", contextThatHasTenant(internalTntID), testFormationID, faTargetID).Return(emptyFormationAssignmentsForObject, nil).Once()
//				return faSvc
//			},
//			formationSvcFn: func() *automock.FormationService {
//				formationSvc := &automock.FormationService{}
//				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormationWithReadyState, nil).Once()
//				formationSvc.On("UnassignFormation", contextThatHasTenant(internalTntID), internalTntID, faSourceID, graphql.FormationObjectType(graphql.FormationAssignmentTypeApplication), *testFormationWithReadyState).Return(testFormationWithReadyState, nil).Once()
//				formationSvc.On("UnassignFormation", contextThatHasTenant(internalTntID), internalTntID, faTargetID, graphql.FormationObjectType(graphql.FormationAssignmentTypeRuntime), *testFormationWithReadyState).Return(nil, testErr).Once()
//				return formationSvc
//			},
//			faStatusSvcFn: func() *automock.FormationAssignmentStatusService {
//				faStatusSvc := &automock.FormationAssignmentStatusService{}
//				faStatusSvc.On("DeleteWithConstraints", contextThatHasTenant(internalTntID), testFormationAssignmentID).Return(nil).Once()
//				return faStatusSvc
//			},
//			reqBody: fm.FormationAssignmentRequestBody{
//				State:         model.ReadyAssignmentState,
//				Configuration: json.RawMessage(testValidConfig),
//			},
//			hasURLVars:         true,
//			expectedStatusCode: http.StatusInternalServerError,
//			expectedErrOutput:  "An unexpected error occurred while processing the request. X-Request-Id:",
//		},
//		{
//			name:       "Error when transaction fail to commit",
//			transactFn: txGen.ThatFailsOnCommit,
//			faServiceFn: func() *automock.FormationAssignmentService {
//				faSvc := &automock.FormationAssignmentService{}
//				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntimeForUnassingOp, nil).Once()
//				return faSvc
//			},
//			formationSvcFn: func() *automock.FormationService {
//				formationSvc := &automock.FormationService{}
//				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormationWithReadyState, nil).Once()
//				return formationSvc
//			},
//			faStatusSvcFn: func() *automock.FormationAssignmentStatusService {
//				faStatusSvc := &automock.FormationAssignmentStatusService{}
//				faStatusSvc.On("DeleteWithConstraints", contextThatHasTenant(internalTntID), testFormationAssignmentID).Return(nil).Once()
//				return faStatusSvc
//			},
//			reqBody: fm.FormationAssignmentRequestBody{
//				State:         model.ReadyAssignmentState,
//				Configuration: json.RawMessage(testValidConfig),
//			},
//			hasURLVars:         true,
//			expectedStatusCode: http.StatusInternalServerError,
//			expectedErrOutput:  "An unexpected error occurred while processing the request. X-Request-Id:",
//		},
//		{
//			name: "Error when second transaction fail on begin",
//			transactFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
//				persistTx := &persistenceautomock.PersistenceTx{}
//				persistTx.On("Commit").Return(nil).Once()
//
//				transact := &persistenceautomock.Transactioner{}
//				transact.On("Begin").Return(persistTx, nil).Once()
//				transact.On("Begin").Return(nil, testErr).Once()
//				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Once()
//
//				return persistTx, transact
//			},
//			faServiceFn: func() *automock.FormationAssignmentService {
//				faSvc := &automock.FormationAssignmentService{}
//				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntimeForUnassingOp, nil).Once()
//				return faSvc
//			},
//			formationSvcFn: func() *automock.FormationService {
//				formationSvc := &automock.FormationService{}
//				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormationWithReadyState, nil).Once()
//				return formationSvc
//			},
//			faStatusSvcFn: func() *automock.FormationAssignmentStatusService {
//				faStatusSvc := &automock.FormationAssignmentStatusService{}
//				faStatusSvc.On("DeleteWithConstraints", contextThatHasTenant(internalTntID), testFormationAssignmentID).Return(nil).Once()
//				return faStatusSvc
//			},
//			reqBody: fm.FormationAssignmentRequestBody{
//				State:         model.ReadyAssignmentState,
//				Configuration: json.RawMessage(testValidConfig),
//			},
//			hasURLVars:         true,
//			expectedStatusCode: http.StatusInternalServerError,
//			expectedErrOutput:  "An unexpected error occurred while processing the request. X-Request-Id:",
//		},
//		{
//			name: "Error when second transaction fail to commit",
//			transactFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
//				persistTx := &persistenceautomock.PersistenceTx{}
//				persistTx.On("Commit").Return(nil).Once()
//				persistTx.On("Commit").Return(testErr).Once()
//
//				transact := &persistenceautomock.Transactioner{}
//				transact.On("Begin").Return(persistTx, nil).Twice()
//				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Twice()
//
//				return persistTx, transact
//			},
//			faServiceFn: func() *automock.FormationAssignmentService {
//				faSvc := &automock.FormationAssignmentService{}
//				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntimeForUnassingOp, nil).Once()
//				faSvc.On("ListFormationAssignmentsForObjectID", contextThatHasTenant(internalTntID), testFormationID, faSourceID).Return(testFormationAssignmentsForObject, nil).Once()
//				faSvc.On("ListFormationAssignmentsForObjectID", contextThatHasTenant(internalTntID), testFormationID, faTargetID).Return(testFormationAssignmentsForObject, nil).Once()
//				return faSvc
//			},
//			formationSvcFn: func() *automock.FormationService {
//				formationSvc := &automock.FormationService{}
//				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormationWithReadyState, nil).Once()
//				return formationSvc
//			},
//			faStatusSvcFn: func() *automock.FormationAssignmentStatusService {
//				faStatusSvc := &automock.FormationAssignmentStatusService{}
//				faStatusSvc.On("DeleteWithConstraints", contextThatHasTenant(internalTntID), testFormationAssignmentID).Return(nil).Once()
//				return faStatusSvc
//			},
//			reqBody: fm.FormationAssignmentRequestBody{
//				State:         model.ReadyAssignmentState,
//				Configuration: json.RawMessage(testValidConfig),
//			},
//			hasURLVars:         true,
//			expectedStatusCode: http.StatusInternalServerError,
//			expectedErrOutput:  "An unexpected error occurred while processing the request. X-Request-Id:",
//		},
//	}
//
//	for _, tCase := range testCases {
//		t.Run(tCase.name, func(t *testing.T) {
//			// GIVEN
//			marshalBody, err := json.Marshal(tCase.reqBody)
//			require.NoError(t, err)
//
//			httpReq := httptest.NewRequest(http.MethodPatch, url, bytes.NewBuffer(marshalBody))
//			if tCase.hasURLVars {
//				httpReq = mux.SetURLVars(httpReq, urlVars)
//			}
//
//			if tCase.headers != nil {
//				httpReq.Header = tCase.headers
//			}
//			w := httptest.NewRecorder()
//
//			persist, transact := fixUnusedTransactioner()
//			if tCase.transactFn != nil {
//				persist, transact = tCase.transactFn()
//			}
//
//			faSvc := fixUnusedFormationAssignmentSvc()
//			if tCase.faServiceFn != nil {
//				faSvc = tCase.faServiceFn()
//			}
//
//			faNotificationSvc := fixUnusedFormationAssignmentNotificationSvc()
//			if tCase.faNotificationSvcFn != nil {
//				faNotificationSvc = tCase.faNotificationSvcFn()
//			}
//
//			formationSvc := fixUnusedFormationSvc()
//			if tCase.formationSvcFn != nil {
//				formationSvc = tCase.formationSvcFn()
//			}
//
//			faStatusSvcFn := &automock.FormationAssignmentStatusService{}
//			if tCase.faStatusSvcFn != nil {
//				faStatusSvcFn = tCase.faStatusSvcFn()
//			}
//
//			defer mock.AssertExpectationsForObjects(t, persist, transact, faSvc, faNotificationSvc, formationSvc, faStatusSvcFn)
//
//			handler := fm.NewFormationMappingHandler(transact, faSvc, faStatusSvcFn, faNotificationSvc, formationSvc, nil)
//
//			// WHEN
//			handler.UpdateFormationAssignmentStatus(w, httpReq)
//
//			// THEN
//			resp := w.Result()
//			body, err := io.ReadAll(resp.Body)
//			require.NoError(t, err)
//
//			if tCase.expectedErrOutput == "" {
//				require.NoError(t, err)
//			} else {
//				require.Contains(t, string(body), tCase.expectedErrOutput)
//			}
//			require.Equal(t, tCase.expectedStatusCode, resp.StatusCode)
//		})
//	}
//}

// todo::: revert and fix unit test
//func TestHandler_UpdateFormationStatus(t *testing.T) {
//	url := fmt.Sprintf("/v1/businessIntegrations/{%s}/status", fm.FormationIDParam)
//	urlVars := map[string]string{
//		fm.FormationIDParam: testFormationID,
//	}
//
//	formationWithInitialState := fixFormationWithState(model.InitialFormationState)
//	formationWithReadyState := fixFormationWithState(model.ReadyFormationState)
//	formationWithDeletingState := fixFormationWithState(model.DeletingFormationState)
//
//	testCases := []struct {
//		name                 string
//		transactFn           func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
//		formationSvcFn       func() *automock.FormationService
//		faStatusSvcFn        func() *automock.FormationAssignmentStatusService
//		formationStatusSvcFn func() *automock.FormationStatusService
//		reqBody              fm.FormationRequestBody
//		hasURLVars           bool
//		headers              map[string][]string
//		expectedStatusCode   int
//		expectedErrOutput    string
//	}{
//		// Request(+metadata) validation checks
//		{
//			name:               "Decode Error: Content-Type header is not application/json",
//			headers:            map[string][]string{httputils.HeaderContentTypeKey: {"invalidContentType"}},
//			expectedStatusCode: http.StatusUnsupportedMediaType,
//			expectedErrOutput:  "Content-Type header is not application/json",
//		},
//		{
//			name: "Error when the required path parameter is missing",
//			reqBody: fm.FormationRequestBody{
//				State: model.ReadyFormationState,
//			},
//			expectedStatusCode: http.StatusBadRequest,
//			expectedErrOutput:  "Not all of the required parameters are provided",
//		},
//		// Request body validation checks
//		{
//			name: "Validate Error: error when the state has unsupported value",
//			reqBody: fm.FormationRequestBody{
//				State: "unsupported",
//			},
//			expectedStatusCode: http.StatusBadRequest,
//			expectedErrOutput:  "Request Body contains invalid input:",
//		},
//		{
//			name: "Validate Error: error when we have an error with incorrect state",
//			reqBody: fm.FormationRequestBody{
//				State: model.ReadyFormationState,
//				Error: testErr.Error(),
//			},
//			expectedStatusCode: http.StatusBadRequest,
//			expectedErrOutput:  "Request Body contains invalid input:",
//		},
//		// Business logic unit tests
//		{
//			name:       "Error when transaction fails to begin",
//			transactFn: txGen.ThatFailsOnBegin,
//			reqBody: fm.FormationRequestBody{
//				State: model.ReadyFormationState,
//			},
//			hasURLVars:         true,
//			expectedStatusCode: http.StatusInternalServerError,
//			expectedErrOutput:  "An unexpected error occurred while processing the request. X-Request-Id:",
//		},
//		{
//			name:       "Error when getting formation globally fails",
//			transactFn: txGen.ThatDoesntExpectCommit,
//			formationSvcFn: func() *automock.FormationService {
//				formationSvc := &automock.FormationService{}
//				formationSvc.On("GetGlobalByID", txtest.CtxWithDBMatcher(), testFormationID).Return(nil, testErr).Once()
//				return formationSvc
//			},
//			reqBody: fm.FormationRequestBody{
//				State: model.ReadyFormationState,
//			},
//			hasURLVars:         true,
//			expectedStatusCode: http.StatusInternalServerError,
//			expectedErrOutput:  "An unexpected error occurred while processing the request. X-Request-Id:",
//		},
//		// Business logic unit tests for delete formation operation
//		{
//			name:       "Successfully update formation status when operation is delete formation",
//			transactFn: txGen.ThatSucceeds,
//			formationSvcFn: func() *automock.FormationService {
//				formationSvc := &automock.FormationService{}
//				formationSvc.On("GetGlobalByID", txtest.CtxWithDBMatcher(), testFormationID).Return(formationWithDeletingState, nil).Once()
//				return formationSvc
//			},
//			formationStatusSvcFn: func() *automock.FormationStatusService {
//				formationStatusSvc := &automock.FormationStatusService{}
//				formationStatusSvc.On("DeleteFormationEntityAndScenariosWithConstraints", contextThatHasTenant(internalTntID), internalTntID, formationWithDeletingState).Return(nil).Once()
//				return formationStatusSvc
//			},
//			reqBody: fm.FormationRequestBody{
//				State: model.ReadyFormationState,
//			},
//			hasURLVars:         true,
//			expectedStatusCode: http.StatusOK,
//		},
//		{
//			name:       "Successfully update formation status when operation is delete formation and state is DELETE_ERROR",
//			transactFn: txGen.ThatSucceeds,
//			formationSvcFn: func() *automock.FormationService {
//				formationSvc := &automock.FormationService{}
//				formationSvc.On("GetGlobalByID", txtest.CtxWithDBMatcher(), testFormationID).Return(formationWithDeletingState, nil).Once()
//				return formationSvc
//			},
//			formationStatusSvcFn: func() *automock.FormationStatusService {
//				formationStatusSvc := &automock.FormationStatusService{}
//				formationStatusSvc.On("SetFormationToErrorStateWithConstraints", contextThatHasTenant(internalTntID), formationWithDeletingState, testErr.Error(), formationassignment.AssignmentErrorCode(formationassignment.ClientError), model.DeleteErrorFormationState, model.DeleteFormation).Return(nil).Once()
//				return formationStatusSvc
//			},
//			reqBody: fm.FormationRequestBody{
//				State: model.DeleteErrorFormationState,
//				Error: testErr.Error(),
//			},
//			hasURLVars:         true,
//			expectedStatusCode: http.StatusOK,
//		},
//		{
//			name:       "Error when request body state is not correct for delete formation operation",
//			transactFn: txGen.ThatDoesntExpectCommit,
//			formationSvcFn: func() *automock.FormationService {
//				formationSvc := &automock.FormationService{}
//				formationSvc.On("GetGlobalByID", txtest.CtxWithDBMatcher(), testFormationID).Return(formationWithDeletingState, nil).Once()
//				return formationSvc
//			},
//			reqBody: fm.FormationRequestBody{
//				State: model.CreateErrorFormationState,
//			},
//			hasURLVars:         true,
//			expectedStatusCode: http.StatusInternalServerError,
//			expectedErrOutput:  "An unexpected error occurred while processing the request. X-Request-Id:",
//		},
//		{
//			name:       "Error when updating the formation to delete error state fails",
//			transactFn: txGen.ThatDoesntExpectCommit,
//			formationSvcFn: func() *automock.FormationService {
//				formationSvc := &automock.FormationService{}
//				formationSvc.On("GetGlobalByID", txtest.CtxWithDBMatcher(), testFormationID).Return(formationWithDeletingState, nil).Once()
//				return formationSvc
//			},
//			formationStatusSvcFn: func() *automock.FormationStatusService {
//				formationStatusSvc := &automock.FormationStatusService{}
//				formationStatusSvc.On("SetFormationToErrorStateWithConstraints", contextThatHasTenant(internalTntID), formationWithDeletingState, testErr.Error(), formationassignment.AssignmentErrorCode(formationassignment.ClientError), model.DeleteErrorFormationState, model.DeleteFormation).Return(testErr).Once()
//				return formationStatusSvc
//			},
//			reqBody: fm.FormationRequestBody{
//				State: model.DeleteErrorFormationState,
//				Error: testErr.Error(),
//			},
//			hasURLVars:         true,
//			expectedStatusCode: http.StatusInternalServerError,
//			expectedErrOutput:  "An unexpected error occurred while processing the request. X-Request-Id:",
//		},
//		{
//			name:       "Error when deleting formation fails",
//			transactFn: txGen.ThatDoesntExpectCommit,
//			formationSvcFn: func() *automock.FormationService {
//				formationSvc := &automock.FormationService{}
//				formationSvc.On("GetGlobalByID", txtest.CtxWithDBMatcher(), testFormationID).Return(formationWithDeletingState, nil).Once()
//				return formationSvc
//			},
//			formationStatusSvcFn: func() *automock.FormationStatusService {
//				formationStatusSvc := &automock.FormationStatusService{}
//				formationStatusSvc.On("DeleteFormationEntityAndScenariosWithConstraints", contextThatHasTenant(internalTntID), internalTntID, formationWithDeletingState).Return(testErr).Once()
//				return formationStatusSvc
//			},
//			reqBody: fm.FormationRequestBody{
//				State: model.ReadyFormationState,
//			},
//			hasURLVars:         true,
//			expectedStatusCode: http.StatusInternalServerError,
//			expectedErrOutput:  "An unexpected error occurred while processing the request. X-Request-Id:",
//		},
//		{
//			name:       "Error when transaction fails to commit after successful formation status update for delete operation",
//			transactFn: txGen.ThatFailsOnCommit,
//			formationSvcFn: func() *automock.FormationService {
//				formationSvc := &automock.FormationService{}
//				formationSvc.On("GetGlobalByID", txtest.CtxWithDBMatcher(), testFormationID).Return(formationWithDeletingState, nil).Once()
//				return formationSvc
//			},
//			formationStatusSvcFn: func() *automock.FormationStatusService {
//				formationStatusSvc := &automock.FormationStatusService{}
//				formationStatusSvc.On("DeleteFormationEntityAndScenariosWithConstraints", contextThatHasTenant(internalTntID), internalTntID, formationWithDeletingState).Return(nil).Once()
//				return formationStatusSvc
//			},
//			reqBody: fm.FormationRequestBody{
//				State: model.ReadyFormationState,
//			},
//			hasURLVars:         true,
//			expectedStatusCode: http.StatusInternalServerError,
//			expectedErrOutput:  "An unexpected error occurred while processing the request. X-Request-Id:",
//		},
//		// Business logic unit tests for create formation operation
//		{
//			name:       "Successfully update formation status when operation is create formation",
//			transactFn: txGen.ThatSucceeds,
//			formationSvcFn: func() *automock.FormationService {
//				formationSvc := &automock.FormationService{}
//				formationSvc.On("GetGlobalByID", txtest.CtxWithDBMatcher(), testFormationID).Return(formationWithInitialState, nil).Once()
//				formationSvc.On("ResynchronizeFormationNotifications", contextThatHasTenant(internalTntID), testFormationID, false).Return(nil, nil).Once()
//				return formationSvc
//			},
//			formationStatusSvcFn: func() *automock.FormationStatusService {
//				formationStatusSvc := &automock.FormationStatusService{}
//				formationStatusSvc.On("UpdateWithConstraints", contextThatHasTenant(internalTntID), formationWithReadyState, model.CreateFormation).Return(nil).Once()
//				return formationStatusSvc
//			},
//			reqBody: fm.FormationRequestBody{
//				State: model.ReadyFormationState,
//			},
//			hasURLVars:         true,
//			expectedStatusCode: http.StatusOK,
//		},
//		{
//			name:       "Successfully update formation status when operation is create formation and state is CREATE_ERROR",
//			transactFn: txGen.ThatSucceeds,
//			formationSvcFn: func() *automock.FormationService {
//				formationSvc := &automock.FormationService{}
//				formationSvc.On("GetGlobalByID", txtest.CtxWithDBMatcher(), testFormationID).Return(formationWithInitialState, nil).Once()
//				return formationSvc
//			},
//			formationStatusSvcFn: func() *automock.FormationStatusService {
//				formationStatusSvc := &automock.FormationStatusService{}
//				formationStatusSvc.On("SetFormationToErrorStateWithConstraints", contextThatHasTenant(internalTntID), formationWithInitialState, testErr.Error(), formationassignment.AssignmentErrorCode(formationassignment.ClientError), model.CreateErrorFormationState, model.CreateFormation).Return(nil).Once()
//				return formationStatusSvc
//			},
//			reqBody: fm.FormationRequestBody{
//				State: model.CreateErrorFormationState,
//				Error: testErr.Error(),
//			},
//			hasURLVars:         true,
//			expectedStatusCode: http.StatusOK,
//		},
//		{
//			name:       "Error when request body state is not correct for create formation operation",
//			transactFn: txGen.ThatDoesntExpectCommit,
//			formationSvcFn: func() *automock.FormationService {
//				formationSvc := &automock.FormationService{}
//				formationSvc.On("GetGlobalByID", txtest.CtxWithDBMatcher(), testFormationID).Return(formationWithInitialState, nil).Once()
//				return formationSvc
//			},
//			reqBody: fm.FormationRequestBody{
//				State: model.DeleteErrorFormationState,
//			},
//			hasURLVars:         true,
//			expectedStatusCode: http.StatusInternalServerError,
//			expectedErrOutput:  "An unexpected error occurred while processing the request. X-Request-Id:",
//		},
//		{
//			name:       "Error when updating the formation to create error state fails",
//			transactFn: txGen.ThatDoesntExpectCommit,
//			formationSvcFn: func() *automock.FormationService {
//				formationSvc := &automock.FormationService{}
//				formationSvc.On("GetGlobalByID", txtest.CtxWithDBMatcher(), testFormationID).Return(formationWithInitialState, nil).Once()
//				return formationSvc
//			},
//			formationStatusSvcFn: func() *automock.FormationStatusService {
//				formationStatusSvc := &automock.FormationStatusService{}
//				formationStatusSvc.On("SetFormationToErrorStateWithConstraints", contextThatHasTenant(internalTntID), formationWithInitialState, testErr.Error(), formationassignment.AssignmentErrorCode(formationassignment.ClientError), model.CreateErrorFormationState, model.CreateFormation).Return(testErr).Once()
//				return formationStatusSvc
//			},
//			reqBody: fm.FormationRequestBody{
//				State: model.CreateErrorFormationState,
//				Error: testErr.Error(),
//			},
//			hasURLVars:         true,
//			expectedStatusCode: http.StatusInternalServerError,
//			expectedErrOutput:  "An unexpected error occurred while processing the request. X-Request-Id:",
//		},
//		{
//			name:       "Error when updating the formation to ready state fails",
//			transactFn: txGen.ThatDoesntExpectCommit,
//			formationSvcFn: func() *automock.FormationService {
//				formationSvc := &automock.FormationService{}
//				formationSvc.On("GetGlobalByID", txtest.CtxWithDBMatcher(), testFormationID).Return(formationWithInitialState, nil).Once()
//				return formationSvc
//			},
//			formationStatusSvcFn: func() *automock.FormationStatusService {
//				formationStatusSvc := &automock.FormationStatusService{}
//				formationStatusSvc.On("UpdateWithConstraints", contextThatHasTenant(internalTntID), formationWithReadyState, model.CreateFormation).Return(testErr).Once()
//				return formationStatusSvc
//			},
//			reqBody: fm.FormationRequestBody{
//				State: model.ReadyFormationState,
//			},
//			hasURLVars:         true,
//			expectedStatusCode: http.StatusInternalServerError,
//			expectedErrOutput:  "An unexpected error occurred while processing the request. X-Request-Id:",
//		},
//		{
//			name:       "Error when resynchronize formation notifications fails",
//			transactFn: txGen.ThatDoesntExpectCommit,
//			formationSvcFn: func() *automock.FormationService {
//				formationSvc := &automock.FormationService{}
//				formationSvc.On("GetGlobalByID", txtest.CtxWithDBMatcher(), testFormationID).Return(formationWithInitialState, nil).Once()
//				formationSvc.On("ResynchronizeFormationNotifications", contextThatHasTenant(internalTntID), testFormationID, false).Return(nil, testErr).Once()
//				return formationSvc
//			},
//			formationStatusSvcFn: func() *automock.FormationStatusService {
//				formationStatusSvc := &automock.FormationStatusService{}
//				formationStatusSvc.On("UpdateWithConstraints", contextThatHasTenant(internalTntID), formationWithReadyState, model.CreateFormation).Return(nil).Once()
//				return formationStatusSvc
//			},
//			reqBody: fm.FormationRequestBody{
//				State: model.ReadyFormationState,
//			},
//			hasURLVars:         true,
//			expectedStatusCode: http.StatusInternalServerError,
//			expectedErrOutput:  "An unexpected error occurred while processing the request. X-Request-Id:",
//		},
//		{
//			name:       "Error when transaction fails to commit after successful formation status update for create operation",
//			transactFn: txGen.ThatFailsOnCommit,
//			formationSvcFn: func() *automock.FormationService {
//				formationSvc := &automock.FormationService{}
//				formationSvc.On("GetGlobalByID", txtest.CtxWithDBMatcher(), testFormationID).Return(formationWithInitialState, nil).Once()
//				formationSvc.On("ResynchronizeFormationNotifications", contextThatHasTenant(internalTntID), testFormationID, false).Return(nil, nil).Once()
//				return formationSvc
//			},
//			formationStatusSvcFn: func() *automock.FormationStatusService {
//				formationStatusSvc := &automock.FormationStatusService{}
//				formationStatusSvc.On("UpdateWithConstraints", contextThatHasTenant(internalTntID), formationWithReadyState, model.CreateFormation).Return(nil).Once()
//				return formationStatusSvc
//			},
//			reqBody: fm.FormationRequestBody{
//				State: model.ReadyFormationState,
//			},
//			hasURLVars:         true,
//			expectedStatusCode: http.StatusInternalServerError,
//			expectedErrOutput:  "An unexpected error occurred while processing the request. X-Request-Id:",
//		},
//	}
//
//	for _, tCase := range testCases {
//		t.Run(tCase.name, func(t *testing.T) {
//			// GIVEN
//			marshalBody, err := json.Marshal(tCase.reqBody)
//			require.NoError(t, err)
//
//			httpReq := httptest.NewRequest(http.MethodPatch, url, bytes.NewBuffer(marshalBody))
//			if tCase.hasURLVars {
//				httpReq = mux.SetURLVars(httpReq, urlVars)
//			}
//
//			if tCase.headers != nil {
//				httpReq.Header = tCase.headers
//			}
//			w := httptest.NewRecorder()
//
//			persist, transact := fixUnusedTransactioner()
//			if tCase.transactFn != nil {
//				persist, transact = tCase.transactFn()
//			}
//
//			formationSvc := fixUnusedFormationSvc()
//			if tCase.formationSvcFn != nil {
//				formationSvc = tCase.formationSvcFn()
//			}
//
//			faUpdater := &automock.FormationAssignmentStatusService{}
//			if tCase.faStatusSvcFn != nil {
//				faUpdater = tCase.faStatusSvcFn()
//			}
//
//			formationStatusSvc := &automock.FormationStatusService{}
//			if tCase.formationStatusSvcFn != nil {
//				formationStatusSvc = tCase.formationStatusSvcFn()
//			}
//
//			defer mock.AssertExpectationsForObjects(t, persist, transact, formationSvc, faUpdater, formationStatusSvc)
//
//			handler := fm.NewFormationMappingHandler(transact, nil, faUpdater, nil, formationSvc, formationStatusSvc)
//
//			// WHEN
//			handler.UpdateFormationStatus(w, httpReq)
//
//			// THEN
//			resp := w.Result()
//			body, err := io.ReadAll(resp.Body)
//			require.NoError(t, err)
//
//			if tCase.expectedErrOutput == "" {
//				require.NoError(t, err)
//			} else {
//				require.Contains(t, string(body), tCase.expectedErrOutput)
//			}
//			require.Equal(t, tCase.expectedStatusCode, resp.StatusCode)
//		})
//	}
//}
