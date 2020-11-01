import LuigiClient from '@luigi-project/client';

export default async function unassignScenarioHandler(
  entityName,
  entityId,
  currentEntityLabels,
  unassignMutation,
  deleteScenarioMutation,
  scenarioAssignment,
  scenarioName,
  successCallback,
) {
  const showConfirmation = () =>
    LuigiClient.uxManager().showConfirmationModal({
      header: `Unassign ${entityName}`,
      body: `Are you sure you want to unassign "${entityName}"?`,
      buttonConfirm: 'Confirm',
      buttonDismiss: 'Cancel',
    });

  const showAlert = () =>
    LuigiClient.uxManager().showAlert({
      text: `Please remove the associated automatic scenario assignment.`,
      type: 'error',
      closeAfter: 5000,
    });

  const tryDeleteScenario = async () => {
    try {
      const scenarios = currentEntityLabels.scenarios.filter(
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

  let canDelete = true;
  if (
    scenarioAssignment &&
    currentEntityLabels[scenarioAssignment.selector.key]
  ) {
    let asaLabelKey = scenarioAssignment.selector.key;
    let asaLabelValue = scenarioAssignment.selector.value;

    canDelete =
      !currentEntityLabels[asaLabelKey] ||
      currentEntityLabels[asaLabelKey] !== asaLabelValue;
  }

  if (!canDelete) {
    showAlert().catch(() => {});
    return;
  }

  showConfirmation()
    .then(tryDeleteScenario)
    .catch(() => {});
}
