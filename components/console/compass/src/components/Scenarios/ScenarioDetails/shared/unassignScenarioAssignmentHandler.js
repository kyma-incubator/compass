import LuigiClient from '@luigi-project/client';

export default async function unassignScenarioAssignmentHandler(
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
