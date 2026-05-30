import { createPinia } from 'pinia';
import { createApp } from 'vue';
import { DefaultApolloClient } from '@vue/apollo-composable';

import App from './App.vue';
import { apolloClient } from './graphql/client';
import router from './router';
import './style.css';

createApp(App).provide(DefaultApolloClient, apolloClient).use(createPinia()).use(router).mount('#app');
