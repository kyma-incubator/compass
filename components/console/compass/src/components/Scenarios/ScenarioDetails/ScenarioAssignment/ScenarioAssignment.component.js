import React from 'react';
import PropTypes from 'prop-types';

import { GenericList } from 'react-shared';
import { useMutation } from '@apollo/react-hooks';

import unassignScenarioAssignmentHandler from '../shared/unassignScenarioAssignmentHandler';
import { SEND_NOTIFICATION } from '../../../../gql';

ScenarioAssignment.propTypes = {
  scenarioName: PropTypes.string.isRequired,
  getScenarioAssignment: PropTypes.object.isRequired,
  deleteScenarioAssignment: PropTypes.func.isRequired,
};

export default function ScenarioAssignment({
  scenarioName,
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
        content: `Unassigned "${scenarioAssignmentName}" from ${scenarioName}.`,
        title: scenarioAssignmentName,
        color: '#359c46',
        icon: 'accept',
        instanceName: scenarioName,
      },
    });
  };

  let scenarioAssignments = [];
  const actions = [
    {
      name: 'Unassign',
      handler: async scenarioAssignment => {
        await unassignScenarioAssignmentHandler(
          deleteScenarioAssignment,
          scenarioName,
          async () => {
            showSuccessNotification(scenarioAssignment.selector.key);
            await (function() {
              getScenarioAssignment.refetch();
            })();
          },
        );
      },
    },
  ];

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
