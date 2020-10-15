import React from 'react';
import Paragraph from '../index';
import 'jest-styled-components';

import { renderWithTheme } from 'test_setup/helpers';

describe('Paragraph Element', () => {
  it('renders correctly', () => {
    const tree = renderWithTheme(
      <Paragraph>A paragraph tag</Paragraph>,
    ).toJSON();
    expect(tree).toMatchSnapshot();
  });

  it('adds the italic modifier', () => {
    const tree = renderWithTheme(
      <Paragraph modifiers={['italic']}>A paragraph</Paragraph>,
    ).toJSON();
    expect(tree).toMatchSnapshot();
  });

  it('adds the extraSmall modifier', () => {
    const tree = renderWithTheme(
      <Paragraph modifiers={['extraSmall']}>A paragraph</Paragraph>,
    ).toJSON();
    expect(tree).toMatchSnapshot();
  });

  it('adds the small modifier', () => {
    const tree = renderWithTheme(
      <Paragraph modifiers={['small']}>A paragraph</Paragraph>,
    ).toJSON();
    expect(tree).toMatchSnapshot();
  });

  it('adds the large modifier', () => {
    const tree = renderWithTheme(
      <Paragraph modifiers={['large']}>A paragraph</Paragraph>,
    ).toJSON();
    expect(tree).toMatchSnapshot();
  });

  it('adds the fontWeightBold modifier', () => {
    const tree = renderWithTheme(
      <Paragraph modifiers={['fontWeightBold']}>A paragraph</Paragraph>,
    ).toJSON();
    expect(tree).toMatchSnapshot();
  });
});
