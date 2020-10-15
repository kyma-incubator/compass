import styled from 'styled-components';
import { tabsStyling } from '../renderers/styled';

export const ContentUIWrapper = styled.div`
  &&& {
    width: 100%;
  }
`;

export const CatalogUIWrapper = styled.div`
  &&& {
    width: 100%;

    ${tabsStyling}

    .cms__tabs {
      border-bottom: none;
      border-left: none;
      border-right: none;
      border-radius: 0;
    }

    .cms__tabs-content {
      margin: 32px;
      border-radius: 4px;

      .custom-markdown-styling {
        background: #fff;
        border-radius: 4px;
      }
    }

    .cms__tab {
      &:first-child {
        margin-left: 32px;
      }
    }

    .grid-unit-content > div {
      border: none;
      margin: 0;
    }

    .headers-navigation-wrapper {
      margin-top: 16px;
      margin-right: 16px;
    }
  }
`;

export const InstancesUIWrapper = styled.div`
  &&&&&&&& {
    width: 100%;

    ${tabsStyling}

    .cms__tabs {
      border: none;
    }

    .cms__tabs-header {
      border-top-right-radius: 4px;
      border-top-left-radius: 4px;
    }

    .cms__tabs-content {
      margin: 0;
      background: #fff;
      border-bottom-right-radius: 4px;
      border-bottom-left-radius: 4px;
    }

    .cms__tab {
      &:first-child {
        margin-left: 16px;
      }
    }

    .custom-open-api-styling {
      border-top-right-radius: 0;
      border-top-left-radius: 0;
    }

    .custom-odata-styling {
      padding: 0 16px 16px 16px;

      button {
        margin: 16px 0;
      }

      > section {
        padding: 16px;
      }
    }

    .grid-unit-content > div {
      border: none;
      margin: 0;
    }

    .headers-navigation-wrapper {
      margin-top: 16px;
      margin-right: 16px;
    }

    .asyncapi {
      padding: 16px;
    }
  }
`;

export const CompassUIWrapper = styled.div`
  &&& {
    width: 100%;
    margin: 32px;
    background: #fff;
    border-radius: 4px;
  }
`;
