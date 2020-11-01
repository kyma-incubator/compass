export const runtimeResponseMock = {
  error: null,
  loading: false,
  runtimes: {
    data: [
      {
        name: 'Runtime 1',
        id: 'id 1',
        labels: {},
      },
    ],
    totalCount: 1,
  },
};

export const assignmentResponseMock = {
  error: null,
  loading: false,
  automaticScenarioAssignmentForScenario: {
    scenarioName: 'test-scenario',
    selector: {
      key: 'test-key',
      value: 'test-value',
    },
  },
};
