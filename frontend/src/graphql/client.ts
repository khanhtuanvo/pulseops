import { ApolloClient, HttpLink, InMemoryCache, split } from '@apollo/client/core';
import { GraphQLWsLink } from '@apollo/client/link/subscriptions';
import { getMainDefinition } from '@apollo/client/utilities';
import { createClient } from 'graphql-ws';

function apiBaseUrl() {
  return import.meta.env.VITE_API_URL ?? '';
}

function httpGraphqlUrl() {
  return `${apiBaseUrl()}/query`;
}

function wsGraphqlUrl() {
  const base = apiBaseUrl() || globalThis.location?.origin || 'http://localhost:8080';
  const url = new URL('/query', base);
  url.protocol = url.protocol === 'https:' ? 'wss:' : 'ws:';
  return url.toString();
}

const httpLink = new HttpLink({
  uri: httpGraphqlUrl(),
  credentials: 'include',
});

const link =
  typeof window === 'undefined'
    ? httpLink
    : split(
        ({ query }) => {
          const definition = getMainDefinition(query);
          return definition.kind === 'OperationDefinition' && definition.operation === 'subscription';
        },
        new GraphQLWsLink(
          createClient({
            url: wsGraphqlUrl(),
            connectionParams: {},
          }),
        ),
        httpLink,
      );

export const apolloClient = new ApolloClient({
  cache: new InMemoryCache(),
  link,
});
