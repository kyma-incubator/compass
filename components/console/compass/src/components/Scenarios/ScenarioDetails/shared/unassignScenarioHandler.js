import LuigiClient from '@luigi-project/client';

export default async function unassignScenarioHandler(
  entityName,
  entityId,
  currentEntityScenarios,
  unassignMutation,
  deleteScenarioMutation,
  scenarioName,
  successCallback,
) {
  const showConfirmation = () =>
    LuigiClient.uxManager().showConfirmationModal({
      header: `Unassign ${entityName}`,
      body: `Are you sure you want to unassign ${entityName}?`,
      buttonConfirm: 'Confirm',
      buttonDismiss: 'Cancel',
    });

  const tryDeleteScenario = async () => {
    try {
      const scenarios = currentEntityScenarios.filter(
        scenario => scenario !== scenarioName,
      );

      if (scenarios.length) {
        await unassignMutation({ variables: { id: entityId, scenarios } });
      } else {
        await deleteScenarioMutation(entityId);
      }

      if (successCallback) {
        successCallback();
      }
    } catch (error) {
      console.warn(error);
      LuigiClient.uxManager().showAlert({
        text: error.message,
        type: 'error',
        closeAfter: 10000,
      });
    }
  };

  showConfirmation()
    .then(tryDeleteScenario)
    .catch(() => {});
}

export async function unassignScenarioAssignmentHandler(
  deleteScenarioAssignmentMutation,
  scenarioName,
  successCallback,
) {
  const showConfirmation = () =>
    LuigiClient.uxManager().showConfirmationModal({
      header: `Unassign scenario assignment`,
      body: `Are you sure you want to delete ${scenarioName}'s automatic scenario assignment?`,
      buttonConfirm: 'Confirm',
      buttonDismiss: 'Cancel',
    });

  const tryDeleteScenarioAssignment = async () => {
    try {
      await deleteScenarioAssignmentMutation(scenarioName);

      if (successCallback) {
        successCallback();
      }
    } catch (error) {
      console.warn(error);
      LuigiClient.uxManager().showAlert({
        text: error.message,
        type: 'error',
        closeAfter: 10000,
      });
    }
  };

  showConfirmation()
    .then(tryDeleteScenarioAssignment)
    .catch(() => {});
}
