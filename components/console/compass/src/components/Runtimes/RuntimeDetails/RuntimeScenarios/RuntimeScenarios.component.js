import React from 'react';
import PropTypes from 'prop-types';
import LuigiClient from '@luigi-project/client';
import RuntimeScenarioModal from './RuntimeScenarioModal.container';
import { RuntimeQueryContext } from '../RuntimeDetails';
import { GET_SCENARIO_ASSIGNMENTS } from '../../../Scenarios/gql';

import { GenericList } from 'react-shared';
import { useQuery } from '@apollo/react-hooks';

RuntimeScenarios.propTypes = {
  runtimeId: PropTypes.string.isRequired,
  scenarios: PropTypes.arrayOf(PropTypes.string).isRequired,
  updateScenarios: PropTypes.func.isRequired,
  sendNotification: PropTypes.func.isRequired,
};

export default function RuntimeScenarios({
  runtimeId,
  scenarios,
  updateScenarios,
  sendNotification,
}) {
  const runtimeQuery = React.useContext(RuntimeQueryContext);

  let { data: fetchedScenarioAssignments, error, loading } = useQuery(
    GET_SCENARIO_ASSIGNMENTS,
  );

  if (loading) {
    return <p>Loading...</p>;
  }
  if (error) {
    return <p>{`Error! ${error.message}`}</p>;
  }

  const headerRenderer = () => ['Name'];
  const rowRenderer = label => [<b>{label.scenario}</b>];
  const actions = [
    {
      name: 'Deactivate',
      handler: deactivateScenario,
    },
  ];

  async function deactivateScenario(entry) {
    const scenarioAssignments =
      fetchedScenarioAssignments.automaticScenarioAssignments.data;
    const scenarioName = entry.scenario;

    let canDeactivate = true;
    scenarioAssignments.forEach(function(asa) {
      if (asa.scenarioName === scenarioName) {
        // there's no way to break foreach loop, but the unnecessary iterations should be minimal even in production cases
        canDeactivate = false;
      }
    });

    if (!canDeactivate) {
      await LuigiClient.uxManager().showAlert({
        text: `Please remove the associated automatic scenario assignment.`,
        type: 'error',
        closeAfter: 5000,
      });
      return;
    }

    LuigiClient.uxManager()
      .showConfirmationModal({
        header: 'Deactivate Scenario',
        body: `Are you sure you want to deactivate "${scenarioName}"?`,
        buttonConfirm: 'Confirm',
        buttonDismiss: 'Cancel',
      })
      .then(async () => {
        try {
          await updateScenarios(
            runtimeId,
            scenarios.filter(scenario => scenario !== scenarioName),
          );
          runtimeQuery.refetch();
          sendNotification({
            variables: {
              content: `"${scenarioName}" deactivated in the runtime.`,
              title: `${scenarioName}`,
              color: '#359c46',
              icon: 'accept',
              instanceName: scenarioName,
            },
          });
        } catch (error) {
          console.warn(error);
          LuigiClient.uxManager().showAlert({
            text: error.message,
            type: 'error',
            closeAfter: 10000,
          });
        }
      })
      .catch(() => {});
  }

  const extraHeaderContent = (
    <header>
      <RuntimeScenarioModal
        title="Activate scenario"
        entityId={runtimeId}
        scenarios={scenarios}
        notSelectedMessage="This Runtime doesn't have any active scenarios."
        entityQuery={runtimeQuery}
      />
    </header>
  );

  const entries = scenarios.map(scenario => {
    return { scenario };
  }); // list requires a list of objects

  return (
    <GenericList
      extraHeaderContent={extraHeaderContent}
      title="Active scenarios"
      notFoundMessage="This Runtime doesn't have any active scenarios."
      actions={actions}
      entries={entries}
      headerRenderer={headerRenderer}
      rowRenderer={rowRenderer}
    />
  );
}
