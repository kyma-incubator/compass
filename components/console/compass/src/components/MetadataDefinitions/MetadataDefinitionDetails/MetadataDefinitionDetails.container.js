import { graphql } from 'react-apollo';
import { compose } from 'recompose';

import {
  DELETE_LABEL_DEFINITION,
  GET_LABEL_DEFINITION,
  UPDATE_LABEL_DEFINITION,
} from '../gql';
import { SEND_NOTIFICATION } from '../../../gql';

import MetadataDefinitionDetails from './MetadataDefinitionDetails.component';

export default compose(
  graphql(GET_LABEL_DEFINITION, {
    name: 'metadataDefinition',
    options: props => {
      return {
        fetchPolicy: 'cache-and-network',
        errorPolicy: 'all',
        variables: {
          key: props.definitionKey,
        },
      };
    },
  }),
  graphql(UPDATE_LABEL_DEFINITION, {
    props: ({ mutate, error }) => ({
      updateLabelDefinition: data => {
        const input = {
          ...data,
          schema: data.schema ? JSON.stringify(data.schema) : null,
        };
        mutate({
          variables: {
            in: input,
          },
        });
      },
    }),
  }),
  graphql(DELETE_LABEL_DEFINITION, {
    props: ({ mutate }) => ({
      deleteLabelDefinition: key =>
        mutate({
          variables: {
            key,
          },
        }),
    }),
  }),
  graphql(SEND_NOTIFICATION, {
    name: 'sendNotification',
  }),
)(MetadataDefinitionDetails);
