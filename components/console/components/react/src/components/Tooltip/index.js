import React from 'react';
import PropTypes from 'prop-types';
import styled from 'styled-components';

import Icon from '../Icon';

import {
  TooltipWrapper,
  TooltipContainer,
  TooltipContent,
  TooltipHeader,
  TooltipFooter,
} from './components';

class Tooltip extends React.Component {
  static propTypes = {
    children: PropTypes.any.isRequired,
    content: PropTypes.oneOfType([
      PropTypes.string,
      PropTypes.array,
      PropTypes.element,
    ]),
    show: PropTypes.bool,
    title: PropTypes.string,
    type: PropTypes.string,
    minWidth: PropTypes.string,
    maxWidth: PropTypes.string,
    orientation: PropTypes.string,
    showTooltipTimeout: PropTypes.number,
    hideTooltipTimeout: PropTypes.number,
  };

  static defaultProps = {
    orientation: 'top',
  };

  constructor(props) {
    super(props);
    this.state = {
      visibleTooltip: this.props.show === undefined ? false : this.props.show,
      showTooltip: this.props.show === undefined ? false : this.props.show,
    };
  }

  setVisibility = visible => {
    this.setState({ visibleTooltip: visible });
  };

  handleShowTooltip = () => {
    const { showTooltipTimeout } = this.props;
    if (typeof this.setVisibility === 'function' && !this.state.showTooltip) {
      setTimeout(
        () => this.setVisibility(true),
        showTooltipTimeout ? showTooltipTimeout : 100,
      );
    }
  };

  handleHideTooltip = () => {
    const { hideTooltipTimeout } = this.props;
    if (typeof this.setVisibility === 'function' && !this.state.showTooltip) {
      setTimeout(
        () => this.setVisibility(false),
        hideTooltipTimeout ? hideTooltipTimeout : 100,
      );
    }
  };

  render() {
    const { visibleTooltip, showTooltip } = this.state;
    const {
      children,
      title,
      content,
      minWidth,
      maxWidth,
      icon,
      type,
      orientation,
      wrapperStyles,
    } = this.props;

    return (
      <TooltipWrapper
        onMouseEnter={this.handleShowTooltip}
        onMouseLeave={this.handleHideTooltip}
        type={type === undefined ? 'default' : type}
        wrapperStyles={wrapperStyles}
      >
        {children}
        {visibleTooltip && content && (
          <TooltipContainer
            minWidth={minWidth}
            maxWidth={maxWidth}
            type={type === undefined ? 'default' : type}
            show={showTooltip}
            orientation={orientation}
          >
            {title && (
              <TooltipHeader type={type === undefined ? 'default' : type}>
                {title}
              </TooltipHeader>
            )}
            <TooltipContent type={type === undefined ? 'default' : type}>
              {content}
            </TooltipContent>
          </TooltipContainer>
        )}
      </TooltipWrapper>
    );
  }
}

export default Tooltip;
