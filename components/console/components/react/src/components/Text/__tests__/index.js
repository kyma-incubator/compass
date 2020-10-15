import React from 'react';
import Text from '../index';
import 'jest-styled-components';

import { renderWithTheme } from 'test_setup/helpers';

describe('Text Element', () => {
  it('renders correctly', () => {
    const tree = renderWithTheme(<Text>A text tag</Text>).toJSON();
    expect(tree).toMatchSnapshot();
  });
});
