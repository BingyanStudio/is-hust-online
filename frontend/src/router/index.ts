import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      component: () => import('@/views/public/StatusPage.vue'),
    },
    {
      path: '/:id',
      component: () => import('@/views/public/SiteDetail.vue'),
    },
    {
      path: '/admin',
      component: () => import('@/views/admin/AdminLayout.vue'),
      children: [
        { path: '', redirect: '/admin/sites' },
        {
          path: 'sites',
          component: () => import('@/views/admin/Sites.vue'),
        },
        {
          path: 'clients',
          component: () => import('@/views/admin/Clients.vue'),
        },
        {
          path: 'check-configs',
          component: () => import('@/views/admin/CheckConfigs.vue'),
        },
      ],
    },
  ],
})

export default router
