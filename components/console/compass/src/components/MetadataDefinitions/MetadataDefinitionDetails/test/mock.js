import { GET_LABEL_DEFINITION, UPDATE_LABEL_DEFINITION } from '../../gql';
const invalidSchema = { thisSchema: "has no 'properties' key" };
const validSchema = { properties: { propertyOne: {} } };

export const mocks = [
  {
    request: {
      query: GET_LABEL_DEFINITION,
      variables: {
        key: 'noschemalabel',
      },
    },
    result: {
      data: {
        labelDefinition: {
          key: 'noschemalabel',
          schema: null,
        },
      },
    },
  },
  {
    request: {
      query: GET_LABEL_DEFINITION,
      variables: {
        key: 'labelWithInvalidSchema',
      },
    },
    result: {
      data: {
        labelDefinition: {
          key: 'labelWithInvalidSchema',
          schema: JSON.stringify(invalidSchema),
        },
      },
    },
  },

  {
    request: {
      query: GET_LABEL_DEFINITION,
      variables: {
        key: 'labelWithValidSchema',
      },
    },
    result: {
      data: {
        labelDefinition: {
          key: 'labelWithValidSchema',
          schema: JSON.stringify(validSchema),
        },
      },
    },
  },

  {
    request: {
      query: UPDATE_LABEL_DEFINITION,
      variables: {
        in: {
          key: 'noschemalabel',
          schema: null,
        },
      },
    },
    result: jest.fn().mockReturnValue({
      data: {
        updateLabelDefinition: {
          key: 'noschemalabel',
          schema: null,
        },
      },
    }),
  },

  {
    request: {
      query: UPDATE_LABEL_DEFINITION,
      variables: {
        in: {
          key: 'labelWithValidSchema',
          schema: JSON.stringify(validSchema),
        },
      },
    },
    result: jest.fn().mockReturnValue({
      data: {
        updateLabelDefinition: {
          key: 'labelWithValidSchema',
          schema: JSON.stringify(validSchema),
        },
      },
    }),
  },
];
