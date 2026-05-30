import { createRouter, createWebHistory } from 'vue-router';

import { useAuthStore } from '@/stores/auth';
import AuthCallback from '@/views/AuthCallback.vue';
import DashboardView from '@/views/DashboardView.vue';
import IncidentDetailView from '@/views/IncidentDetailView.vue';
import LoginView from '@/views/LoginView.vue';
import TeamView from '@/views/TeamView.vue';

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', redirect: '/dashboard' },
    { path: '/login', name: 'login', component: LoginView },
    { path: '/auth/callback', name: 'auth-callback', component: AuthCallback },
    {
      path: '/dashboard',
      name: 'dashboard',
      component: DashboardView,
      meta: { requiresAuth: true },
    },
    {
      path: '/incidents/:id',
      name: 'incident-detail',
      component: IncidentDetailView,
      meta: { requiresAuth: true },
    },
    {
      path: '/team',
      name: 'team',
      component: TeamView,
      meta: { requiresAuth: true, roles: ['OWNER', 'RESPONDER'] },
    },
  ],
});

router.beforeEach(async (to) => {
  if (!to.meta.requiresAuth) {
    return true;
  }

  const authStore = useAuthStore();
  if (!authStore.checked) {
    await authStore.fetchMe();
  }

  if (!authStore.user) {
    return { path: '/login', query: { redirect: to.fullPath } };
  }

  const roles = to.meta.roles as string[] | undefined;
  if (roles && !roles.includes(authStore.user.role)) {
    return { path: '/dashboard', query: { error: 'unauthorized' } };
  }

  return true;
});

export default router;
