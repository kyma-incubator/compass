import { graphql } from 'react-apollo';
import { compose } from 'recompose';
import { fromRenderProps } from 'recompose';
import {
  GET_ASSIGNMENT_FOR_SCENARIO,
  DELETE_ASSIGNMENT_FOR_SCENARIO,
  GET_RUNTIMES_FOR_SCENARIO,
  createEqualityQuery,
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
        errorPolicy: 'all',
        variables: {
          scenarioName: scenarioName,
        },
      };
    },
  }),
  graphql(GET_RUNTIMES_FOR_SCENARIO, {
    name: 'getRuntimesForScenario',
    options: ({ scenarioName }) => {
      const filter = {
        key: 'scenarios',
        query: createEqualityQuery(scenarioName),
      };
      return {
        variables: {
          filter: [filter],
        },
      };
    },
  }),
  graphql(SEND_NOTIFICATION, {
    name: 'sendNotification',
  }),
)(ScenarioAssignment);
