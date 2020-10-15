import React from 'react';
import PropTypes from 'prop-types';
import LuigiClient from '@luigi-project/client';

import { Panel, TabGroup, Tab, Button } from 'fundamental-react';
import EditApiHeader from './../EditApiHeader/EditApiHeader.container';
import ResourceNotFound from 'components/Shared/ResourceNotFound.component';
import ApiForm from '../Forms/ApiForm';
import { Dropdown } from 'components/Shared/Dropdown/Dropdown';
import './EditApi.scss';

import { getRefsValues, useMutationObserver } from 'react-shared';
import { createApiData, verifyApiInput } from './../ApiHelpers';
import ApiEditorForm from '../Forms/ApiEditorForm';

const commonPropTypes = {
  apiId: PropTypes.string.isRequired,
  applicationId: PropTypes.string.isRequired, // used in container file
  apiPackageId: PropTypes.string.isRequired, // used in container file
  updateApiDefinition: PropTypes.func.isRequired,
  sendNotification: PropTypes.func.isRequired,
};

EditApi.propTypes = {
  originalApi: PropTypes.object.isRequired,
  applicationName: PropTypes.string.isRequired,
  apiPackageName: PropTypes.string.isRequired,
  ...commonPropTypes,
};

function EditApi({
  originalApi,
  applicationName,
  apiPackageName,
  apiId,
  updateApiDefinition,
  sendNotification,
}) {
  const formRef = React.useRef(null);
  const [formValid, setFormValid] = React.useState(true);
  const [specProvided, setSpecProvided] = React.useState(!!originalApi.spec);
  const [format, setFormat] = React.useState(
    originalApi.spec ? originalApi.spec.format : 'YAML',
  );
  const [apiType, setApiType] = React.useState(
    originalApi.spec ? originalApi.spec.type : 'OPEN_API',
  );
  const [specText, setSpecText] = React.useState(
    originalApi.spec ? originalApi.spec.data : '',
  );

  const formValues = {
    name: React.useRef(null),
    description: React.useRef(null),
    group: React.useRef(null),
    targetURL: React.useRef(null),
  };

  const revalidateForm = () =>
    setFormValid(!!formRef.current && formRef.current.checkValidity());

  useMutationObserver(formRef, revalidateForm);

  const saveChanges = async () => {
    const basicData = getRefsValues(formValues);
    const specData = specProvided
      ? { data: specText, format, type: apiType }
      : null;
    const apiData = createApiData(basicData, specData);

    try {
      await updateApiDefinition(apiId, apiData);

      const name = basicData.name;
      sendNotification({
        variables: {
          content: `Updated API "${name}".`,
          title: name,
          color: '#359c46',
          icon: 'accept',
          instanceName: name,
        },
      });
    } catch (e) {
      console.warn(e);
      LuigiClient.uxManager().showAlert({
        text: `Cannot update API: ${e.message}`,
        type: 'error',
        closeAfter: 10000,
      });
    }
  };

  const updateSpecText = text => {
    setSpecText(text);
    revalidateForm();
  };

  return (
    <>
      <EditApiHeader
        api={originalApi}
        applicationName={applicationName}
        apiPackageName={apiPackageName}
        saveChanges={saveChanges}
        canSaveChanges={formValid}
      />
      <form ref={formRef} onChange={revalidateForm}>
        <TabGroup className="edit-api-tabs">
          <Tab
            key="general-information"
            id="general-information"
            title="General Information"
          >
            <Panel>
              <Panel.Header>
                <p className="fd-has-type-1">General Information</p>
              </Panel.Header>
              <Panel.Body>
                <ApiForm
                  formValues={formValues}
                  defaultValues={{ ...originalApi }}
                />
              </Panel.Body>
            </Panel>
          </Tab>
          <Tab
            key="api-documentation"
            id="api-documentation"
            title="API Documentation"
          >
            <Panel className="spec-editor-panel">
              <Panel.Header>
                <p className="fd-has-type-1">API Documentation</p>
                <Panel.Actions>
                  {specProvided && (
                    <>
                      <Dropdown
                        options={{ JSON: 'JSON', YAML: 'YAML', XML: 'XML' }}
                        selectedOption={format}
                        onSelect={setFormat}
                        width="90px"
                      />
                      <Dropdown
                        options={{ OPEN_API: 'Open API', ODATA: 'OData' }}
                        selectedOption={apiType}
                        onSelect={setApiType}
                        className="fd-has-margin-x-small"
                        width="120px"
                      />
                      <Button
                        type="negative"
                        onClick={() => setSpecProvided(false)}
                      >
                        No documentation
                      </Button>
                    </>
                  )}
                  {!specProvided && (
                    <Button onClick={() => setSpecProvided(true)}>
                      Provide documentation
                    </Button>
                  )}
                </Panel.Actions>
              </Panel.Header>
              <Panel.Body>
                {specProvided && (
                  <ApiEditorForm
                    specText={specText}
                    setSpecText={updateSpecText}
                    specProvided={specProvided}
                    apiType={apiType}
                    format={format}
                    verifyApi={verifyApiInput}
                    revalidateForm={revalidateForm}
                  />
                )}
              </Panel.Body>
            </Panel>
          </Tab>
        </TabGroup>
      </form>
    </>
  );
}

EditApiWrapper.propTypes = {
  apiDataQuery: PropTypes.object.isRequired,
  ...commonPropTypes,
};

export default function EditApiWrapper(props) {
  const dataQuery = props.apiDataQuery;

  if (dataQuery.loading) {
    return <p>Loading...</p>;
  }
  if (dataQuery.error) {
    return <p>`Error! ${dataQuery.error.message}`</p>;
  }

  const originalApi =
    dataQuery.application &&
    dataQuery.application.package &&
    dataQuery.application.package.apiDefinition;

  if (!originalApi) {
    return (
      <ResourceNotFound
        resource="Api Definition"
        breadcrumb="Application"
        navigationPath="/"
        navigationContext="application"
      />
    );
  }

  return (
    <EditApi
      {...props}
      originalApi={originalApi}
      applicationName={dataQuery.application.name}
      apiPackageName={dataQuery.application.package.name}
    />
  );
}
