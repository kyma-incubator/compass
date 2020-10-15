import React from 'react';
import ReactDOM from 'react-dom';
import './index.css';
import App from './App.container';
import { Microfrontend } from 'react-shared';
import { ApolloClientProvider } from './ApolloClientProvider';

(async () => {
  ReactDOM.render(
    <Microfrontend env={process.env}>
      <ApolloClientProvider>
        <App />
      </ApolloClientProvider>
    </Microfrontend>,
    document.getElementById('root'),
  );
})();
