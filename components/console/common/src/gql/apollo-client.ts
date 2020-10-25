import ApolloClient from 'apollo-client';
import { createHttpLink } from 'apollo-link-http';
import { ApolloLink, split } from 'apollo-link';
import { setContext } from 'apollo-link-context';
import { InMemoryCache } from 'apollo-cache-inmemory';
import { getMainDefinition } from 'apollo-utilities';
import { onError } from 'apollo-link-error';

import { WebSocketLink } from './apollo-client-ws';
import { appInitializer } from '../core';
import { getApiUrl } from '../utils';

interface Options {
  enableSubscriptions?: boolean;
}

export function createApolloClient({ enableSubscriptions = false }: Options) {
  const graphqlApiUrl = getApiUrl(
    process.env.REACT_APP_LOCAL_API ? 'graphqlApiUrlLocal' : 'graphqlApiUrl',
  );
  const subscriptionsApiUrl = getApiUrl(
    process.env.REACT_APP_LOCAL_API
      ? 'subscriptionsApiUrlLocal'
      : 'subscriptionsApiUrl',
  );

  const httpLink = createHttpLink({ uri: graphqlApiUrl });
  const authLink = setContext((_, { headers }) => ({
    headers: {
      ...headers,
      authorization: appInitializer.getBearerToken() || null,
    },
  }));
  const cache = new InMemoryCache();
  const authHttpLink = authLink.concat(httpLink);
  const errorLink = onError(({ graphQLErrors, networkError }) => {
    if (process.env.REACT_APP_ENV !== 'production') {
      if (graphQLErrors) {
        graphQLErrors.map(({ message, locations, path }) =>
          // tslint:disable-next-line
          console.error(
            `[GraphQL error]: Message: ${message}, Location: ${locations}, Path: ${path}`,
          ),
        );
      }

      // tslint:disable-next-line
      if (networkError) {
        console.error(`[Network error]: ${networkError}`);
      }
    }
  });

  let link: ApolloLink | null = null;
  if (enableSubscriptions) {
    const wsLink = new WebSocketLink({
      uri: subscriptionsApiUrl,
      options: {
        reconnect: true,
      },
    });

    link = split(
      ({ query }) => {
        const definition = getMainDefinition(query);
        return (
          definition.kind === 'OperationDefinition' &&
          definition.operation === 'subscription'
        );
      },
      wsLink,
      authHttpLink,
    );
  }

  const client = new ApolloClient({
    link: ApolloLink.from([errorLink, link ? link : authHttpLink]),
    cache,
    connectToDevTools: true,
  });

  return client;
}
