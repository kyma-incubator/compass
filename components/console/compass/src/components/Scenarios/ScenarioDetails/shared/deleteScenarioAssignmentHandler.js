import LuigiClient from '@luigi-project/client';

export default async function deleteScenarioAssignmentHandler(
  deleteScenarioAssignmentMutation,
  scenarioName,
  successCallback,
) {
  const showConfirmation = () =>
    LuigiClient.uxManager().showConfirmationModal({
      header: `Delete scenario assignment`,
      body: `Are you sure you want to delete the automatic scenario assignment for "${scenarioName}"?`,
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
