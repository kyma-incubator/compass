import React from 'react';
import PropTypes from 'prop-types';

import { GenericList } from 'react-shared';
import { useMutation } from '@apollo/react-hooks';

import deleteScenarioAssignmentHandler from '../shared/deleteScenarioAssignmentHandler';
import { SEND_NOTIFICATION } from '../../../../gql';

ScenarioAssignment.propTypes = {
  scenarioName: PropTypes.string.isRequired,
  getRuntimesForScenario: PropTypes.object.isRequired,
  getScenarioAssignment: PropTypes.object.isRequired,
  deleteScenarioAssignment: PropTypes.func.isRequired,
};

export default function ScenarioAssignment({
  scenarioName,
  getRuntimesForScenario,
  getScenarioAssignment,
  deleteScenarioAssignment,
}) {
  const [sendNotification] = useMutation(SEND_NOTIFICATION);

  let hasScenarioAssignment = true;
  if (getScenarioAssignment.loading) {
    return <p>Loading...</p>;
  }
  if (getScenarioAssignment.error) {
    let errorType =
      getScenarioAssignment.error.graphQLErrors[0].extensions.error;
    if (!errorType.includes('NotFound')) {
      return `Error! ${getScenarioAssignment.error.message}`;
    }

    hasScenarioAssignment = false;
  }

  const showSuccessNotification = scenarioAssignmentName => {
    sendNotification({
      variables: {
        content: `Removed automatic scenario assignment from "${scenarioName}".`,
        title: `Successfully removed`,
        color: '#359c46',
        icon: 'accept',
        instanceName: scenarioName,
      },
    });
  };

  const actions = [
    {
      name: 'Delete',
      handler: async scenarioAssignment => {
        await deleteScenarioAssignmentHandler(
          deleteScenarioAssignment,
          scenarioName,
          async () => {
            showSuccessNotification(scenarioAssignment.selector.key);
            await getScenarioAssignment.refetch();
            await getRuntimesForScenario.refetch();
          },
        );
      },
    },
  ];

  let scenarioAssignments = [];
  if (hasScenarioAssignment) {
    scenarioAssignments[0] =
      getScenarioAssignment.automaticScenarioAssignmentForScenario;
  }

  return (
    <GenericList
      title="Automatic Scenario Assignment"
      notFoundMessage="No Automatic Scenario Assignment for this Scenario"
      entries={scenarioAssignments}
      headerRenderer={() => ['Key', 'Value']}
      actions={actions}
      rowRenderer={asa => [asa.selector.key, asa.selector.value]}
      showSearchField={false}
    />
  );
}
