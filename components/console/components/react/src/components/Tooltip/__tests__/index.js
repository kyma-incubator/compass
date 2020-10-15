import React from 'react';
import Tooltip from '../index';

import { renderWithTheme } from 'test_setup/helpers';

describe('Tooltip Element', () => {
  it('renders correctly', () => {
    const tree = renderWithTheme(
      <Tooltip content="Content inside">Content for tooltip</Tooltip>,
    ).toJSON();
    expect(tree).toMatchSnapshot();
  });
});
