import React from 'react';
import { ConfigProvider } from './ConfigContext';
import { NotificationProvider } from './NotificationContext';
import { MicrofrontendContextProvider } from './MicrofrontendContext';

const withProvider = Provider => Component => props => (
  <Provider {...props}>
    <Component {...props} />
  </Provider>
);

export const Microfrontend = [
  MicrofrontendContextProvider,
  ConfigProvider,
  NotificationProvider,
].reduce(
  (component, provider) => withProvider(provider)(component),
  ({ children }) => <>{children}</>,
);
