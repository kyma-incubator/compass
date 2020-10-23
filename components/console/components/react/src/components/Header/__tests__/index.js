import React from 'react';
import Header from '../index';
import 'jest-styled-components';

import { renderWithTheme } from 'test_setup/helpers';

describe('Header Element', () => {
  it('renders correctly', () => {
    const tree = renderWithTheme(<Header>A Header tag</Header>).toJSON();
    expect(tree).toMatchSnapshot();
  });
});
