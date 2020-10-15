import React from 'react';
import Button from '../index';

import { renderWithTheme } from 'test_setup/helpers';

describe('Button Element', () => {
  it('renders correctly', () => {
    const tree = renderWithTheme(<Button>Button</Button>).toJSON();
    expect(tree).toMatchSnapshot();
  });
});
