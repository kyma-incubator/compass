import React from 'react';
import Spinner from '../index';

import { renderWithTheme } from 'test_setup/helpers';

describe('Spinner Element', () => {
  it('renders correctly', () => {
    const tree = renderWithTheme(<Spinner />).toJSON();
    expect(tree).toMatchSnapshot();
  });
});
