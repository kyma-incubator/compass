import React from 'react';
import { shallow } from 'enzyme';
import toJson from 'enzyme-to-json';

import ScenarioAssignment from '../ScenarioAssignment.component';
import { runtimeResponseMock } from './mock';
import { assignmentResponseMock } from './mock';
jest.mock('react-shared', () => ({
  GenericList: function GenericListMocked(props) {
    return 'generic-list-mocked-content';
  },
}));

describe('ScenarioAssignment', () => {
  it('Renders with minimal props', () => {
    const component = shallow(
      <ScenarioAssignment
        scenarioName="test-scenario"
        getRuntimesForScenario={runtimeResponseMock}
        getScenarioAssignment={assignmentResponseMock}
        deleteScenarioAssignment={() => {}}
        sendNotification={() => {}}
      />,
    );

    expect(toJson(component)).toMatchSnapshot();
  });
});
