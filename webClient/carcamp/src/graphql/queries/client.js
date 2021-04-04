import { ApolloClient } from 'apollo-client'
import { createHttpLink } from 'apollo-link-http'
import { InMemoryCache } from 'apollo-cache-inmemory'
import { setContext } from 'apollo-link-context'

import { ApolloClient, createHttpLink, InMemoryCache } from '@apollo/client/core'
import Auth from "@aws-amplify/auth";

const httpLink = createHttpLink({
  uri: process.env.VUE_APP_APOLLO_HTTP
})

const authLink = setContext(async (_, { headers }) => {
  var token = null;
  try {
    const session = await Auth.currentSession();
    token = session.idToken.jwtToken;
  } catch {
    console.log("failed to get session");
  }
  
  var newHeaders = {
    ...headers,
    'x-api-key': process.env.VUE_APP_API_KEY,
  }
  if(token) {
    newHeaders["authorization"] = `Bearer ${token}`
  }

  return {
      headers: newHeaders
  }
})


const cache = new InMemoryCache()
const apolloClient = new ApolloClient({
  link: authLink.concat(httpLink),
  cache
})

export default apolloClient
