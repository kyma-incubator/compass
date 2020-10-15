import { graphql } from 'react-apollo';
import { compose } from 'recompose';

import {
  UPDATE_SCENARIOS,
  GET_SCENARIOS_LABEL_SCHEMA,
  CREATE_SCENARIOS_LABEL,
  SET_APPLICATION_SCENARIOS,
  SET_RUNTIME_SCENARIOS,
} from '../../gql';
import { SEND_NOTIFICATION } from '../../../../gql';

import CreateScenarioModal from './CreateScenarioModal.component';

//create input for both create and update
function createLabelDefinitionInput(scenarios) {
  const schema = {
    type: 'array',
    minItems: 1,
    uniqueItems: true,
    items: { type: 'string', enum: scenarios },
  };
  return {
    key: 'scenarios',
    schema: JSON.stringify(schema),
  };
}

export default compose(
  graphql(UPDATE_SCENARIOS, {
    props: props => ({
      addScenario: async (currentScenarios, newScenario) => {
        const input = createLabelDefinitionInput([
          ...currentScenarios,
          newScenario,
        ]);
        await props.mutate({ variables: { in: input } });
      },
    }),
  }),
  graphql(CREATE_SCENARIOS_LABEL, {
    props: props => ({
      createScenarios: async scenarios => {
        // add reqquired scenario
        if (!scenarios.includes('DEFAULT')) {
          scenarios.push('DEFAULT');
        }
        const input = createLabelDefinitionInput(scenarios);
        await props.mutate({ variables: { in: input } });
      },
    }),
  }),
  graphql(SET_APPLICATION_SCENARIOS, {
    props: props => ({
      setApplicationScenarios: async (applicationId, scenarios) => {
        await props.mutate({
          variables: {
            id: applicationId,
            scenarios: scenarios,
          },
        });
      },
    }),
  }),
  graphql(SET_RUNTIME_SCENARIOS, {
    props: props => ({
      setRuntimeScenarios: async (runtimeId, scenarios) => {
        await props.mutate({
          variables: {
            id: runtimeId,
            scenarios: scenarios,
          },
        });
      },
    }),
  }),
  graphql(GET_SCENARIOS_LABEL_SCHEMA, {
    name: 'scenariosQuery',
  }),
  graphql(SEND_NOTIFICATION, {
    name: 'sendNotification',
  }),
)(CreateScenarioModal);
