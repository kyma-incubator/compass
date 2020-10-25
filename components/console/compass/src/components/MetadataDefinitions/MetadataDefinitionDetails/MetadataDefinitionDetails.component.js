import React, { useState } from 'react';
import { ActionBar, Toggle } from 'fundamental-react';
import {
  Button,
  Breadcrumb,
  Panel,
  PanelHead,
  PanelHeader,
  PanelBody,
  PanelActions,
  Input,
} from '@kyma-project/react-components';
import LuigiClient from '@luigi-project/client';

import '../../../shared/styles/header.scss';
import ResourceNotFound from '../../Shared/ResourceNotFound.component';
import JSONEditorComponent from '../../Shared/JSONEditor';
import { handleDelete } from 'react-shared';

const Ajv = require('ajv');
const ajv = new Ajv();

const MetadataDefinitionDetails = ({
  metadataDefinition: metadataDefinitionQuery,
  updateLabelDefinition,
  sendNotification,
  deleteLabelDefinition,
}) => {
  const defaultSchema = { properties: {}, required: [] };

  const [isSchemaValid, setSchemaValid] = useState(true);
  const [schemaError, setSchemaError] = useState(null);
  const [editedSchema, setEditedSchema] = useState(defaultSchema);
  const [isEditorShown, setIsEditorShown] = useState(false);
  const [metadataDefinition, setMetadataDefinition] = useState(null);

  if (!metadataDefinition && !metadataDefinitionQuery.loading) {
    // INITIALIZATION
    const definition = metadataDefinitionQuery.labelDefinition;
    if (definition) {
      definition.schema = JSON.parse(definition.schema);

      setMetadataDefinition(definition);
      setIsEditorShown(!!definition.schema);
      setEditedSchema(definition.schema || defaultSchema);

      LuigiClient.uxManager().setDirtyStatus(false);
    }
  }

  const handleSchemaChange = currentSchema => {
    LuigiClient.uxManager().setDirtyStatus(
      currentSchema !== metadataDefinition.schema,
    );
    try {
      const parsedSchema =
        typeof currentSchema === 'string'
          ? JSON.parse(currentSchema)
          : currentSchema;

      if (!ajv.validateSchema(parsedSchema))
        throw new Error('Provided JSON is not a valid schema');

      setEditedSchema(parsedSchema);
      setSchemaError(null);
      setSchemaValid(true);
    } catch (e) {
      setSchemaError(e.message);
      setSchemaValid(false);
    }
  };

  const handleSaveChanges = async () => {
    try {
      await updateLabelDefinition({
        key: metadataDefinition.key,
        schema: isEditorShown && editedSchema ? editedSchema : null,
      });

      setMetadataDefinition({ ...metadataDefinition, schema: editedSchema }); // to format the JSON
      LuigiClient.uxManager().setDirtyStatus(false);
      await sendNotification({
        variables: {
          content: 'Metadata definition has been saved succesfully',
          title: 'Success',
          color: '#107E3E',
          icon: 'accept',
        },
      });
    } catch (e) {
      LuigiClient.uxManager().showAlert({
        text: `There was a problem with saving Metadata definition. ${e.message}`,
        type: 'error',
      });
    }
  };

  const handleSchemaToggle = () => {
    setIsEditorShown(!isEditorShown);
  };

  const navigateToList = () => {
    LuigiClient.linkManager()
      .fromClosestContext()
      .navigate(`/metadata-definitions`);
  };

  const { loading, error } = metadataDefinitionQuery;

  if (!metadataDefinitionQuery) {
    if (loading) return 'Loading...';
    if (error)
      return (
        <ResourceNotFound
          resource="Metadata definition"
          breadcrumb="MetadataDefinitions"
        />
      );
    return null;
  }
  if (error) {
    return `Error! ${error.message}`;
  }

  return (
    <>
      <header className="fd-has-background-color-background-2">
        <section className="fd-has-padding-regular fd-has-padding-bottom-none action-bar-wrapper">
          <section>
            <Breadcrumb>
              <Breadcrumb.Item
                name="Metadata definitions"
                url="#"
                onClick={() =>
                  LuigiClient.linkManager()
                    .fromClosestContext()
                    .navigate(`/metadata-definitions`)
                }
              />
              <Breadcrumb.Item />
            </Breadcrumb>
            <ActionBar.Header
              title={
                (metadataDefinition && metadataDefinition.key) ||
                'Loading name...'
              }
            />
          </section>
          <ActionBar.Actions>
            <Button
              onClick={handleSaveChanges}
              disabled={!isSchemaValid}
              option="emphasized"
              data-test-id="save"
            >
              Save
            </Button>
            <Button
              onClick={() => {
                handleDelete(
                  'Metadata definition',
                  metadataDefinition.key,
                  metadataDefinition.key,
                  deleteLabelDefinition,
                  navigateToList,
                );
              }}
              option="light"
              type="negative"
            >
              Delete
            </Button>
          </ActionBar.Actions>
        </section>
      </header>
      {metadataDefinition && (
        <section className="fd-section">
          <Panel>
            <PanelHeader>
              <PanelHead title="Schema" />
              <PanelActions>
                {isEditorShown && (
                  <Toggle checked onChange={handleSchemaToggle} />
                )}
                {!isEditorShown && <Toggle onChange={handleSchemaToggle} />}
              </PanelActions>
            </PanelHeader>

            {isEditorShown && (
              <PanelBody>
                <JSONEditorComponent
                  aria-label="schema-editor"
                  onChangeText={handleSchemaChange}
                  text={JSON.stringify(
                    metadataDefinition.schema || defaultSchema,
                    null,
                    2,
                  )}
                />
                <Input
                  type="hidden"
                  isError={schemaError}
                  message={schemaError}
                ></Input>
              </PanelBody>
            )}
          </Panel>
        </section>
      )}
    </>
  );
};

export default MetadataDefinitionDetails;
