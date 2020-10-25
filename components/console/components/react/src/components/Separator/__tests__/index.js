import React from 'react';
import Separator from '../index';

import { renderWithTheme } from 'test_setup/helpers';

describe('Separator Element', () => {
  it('renders correctly', () => {
    const tree = renderWithTheme(<Separator />).toJSON();
    expect(tree).toMatchSnapshot();
  });
});
