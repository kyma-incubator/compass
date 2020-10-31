import { graphql } from 'react-apollo';
import { compose } from 'recompose';
import { fromRenderProps } from 'recompose';
import {
  GET_ASSIGNMENT_FOR_SCENARIO,
  DELETE_ASSIGNMENT_FOR_SCENARIO,
} from '../../gql';
import { SEND_NOTIFICATION } from '../../../../gql';

import ScenarioAssignment from './ScenarioAssignment.component';
import ScenarioNameContext from './../ScenarioNameContext';

export default compose(
  fromRenderProps(ScenarioNameContext.Consumer, scenarioName => ({
    scenarioName,
  })),
  graphql(DELETE_ASSIGNMENT_FOR_SCENARIO, {
    props: ({ mutate }) => ({
      deleteScenarioAssignment: async scenarioName =>
        await mutate({
          variables: { scenarioName },
        }),
    }),
  }),
  graphql(GET_ASSIGNMENT_FOR_SCENARIO, {
    name: 'getScenarioAssignment',
    options: ({ scenarioName }) => {
      return {
        variables: {
          errorPolicy: 'ignore',
          scenarioName: scenarioName,
        },
      };
    },
  }),
  graphql(SEND_NOTIFICATION, {
    name: 'sendNotification',
  }),
)(ScenarioAssignment);
