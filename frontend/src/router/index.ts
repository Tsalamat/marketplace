import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(),
  scrollBehavior(to, _from, savedPosition) {
    if (savedPosition) return savedPosition
    if (to.hash) return { el: to.hash, behavior: 'smooth' }
    return { top: 0 }
  },
  routes: [
    // ── Public ────────────────────────────────────────────────
    {
      path: '/',
      name: 'landing',
      component: () => import('@/pages/LandingPage.vue'),
      meta: { guestOnly: true },
    },
    {
      path: '/login',
      name: 'login',
      component: () => import('@/pages/LoginPage.vue'),
      meta: { guestOnly: true },
    },
    {
      path: '/register',
      name: 'register',
      component: () => import('@/pages/RegisterPage.vue'),
      meta: { guestOnly: true },
    },
    {
      path: '/forgot-password',
      name: 'forgot-password',
      component: () => import('@/pages/ForgotPasswordPage.vue'),
    },
    {
      path: '/reset-password',
      name: 'reset-password',
      component: () => import('@/pages/ResetPasswordPage.vue'),
    },
    {
      path: '/verify-email',
      name: 'verify-email',
      component: () => import('@/pages/VerifyEmailPage.vue'),
    },
    {
      path: '/auth/callback',
      name: 'auth-callback',
      component: () => import('@/pages/AuthCallbackPage.vue'),
    },
    {
      path: '/friends',
      name: 'friends',
      component: () => import('@/pages/FriendsPage.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/map',
      name: 'map',
      component: () => import('@/pages/MapPage.vue'),
      meta: { requiresAuth: true },
    },

    // ── Marketplace ───────────────────────────────────────────
    {
      path: '/marketplace',
      name: 'marketplace',
      component: () => import('@/pages/MarketplacePage.vue'),
    },
    {
      path: '/gig/:slug',
      name: 'gig-detail',
      component: () => import('@/pages/GigDetailPage.vue'),
    },
    {
      path: '/profile/:username',
      name: 'user-profile',
      component: () => import('@/pages/UserProfilePage.vue'),
    },

    // ── Protected ─────────────────────────────────────────────
    {
      path: '/dashboard',
      name: 'dashboard',
      component: () => import('@/pages/DashboardPage.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/orders',
      name: 'orders',
      component: () => import('@/pages/OrdersPage.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/orders/:id',
      name: 'order-detail',
      component: () => import('@/pages/OrderDetailPage.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/chat',
      name: 'chat',
      component: () => import('@/pages/ChatPage.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/community',
      name: 'community',
      component: () => import('@/pages/CommunityPage.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/create-gig',
      name: 'create-gig',
      component: () => import('@/pages/CreateGigPage.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/settings',
      name: 'settings',
      component: () => import('@/pages/SettingsPage.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/notifications',
      name: 'notifications',
      component: () => import('@/pages/NotificationsPage.vue'),
      meta: { requiresAuth: true },
    },

    // ── Admin ─────────────────────────────────────────────────
    {
      path: '/admin',
      component: () => import('@/pages/admin/AdminLayout.vue'),
      meta: { requiresAuth: true, requiresAdmin: true },
      children: [
        { path: '', redirect: '/admin/dashboard' },
        { path: 'dashboard', component: () => import('@/pages/admin/AdminDashboard.vue') },
        { path: 'users', component: () => import('@/pages/admin/AdminUsers.vue') },
        { path: 'services', component: () => import('@/pages/admin/AdminServices.vue') },
        { path: 'orders', component: () => import('@/pages/admin/AdminOrders.vue') },
        { path: 'reports', component: () => import('@/pages/admin/AdminReports.vue') },
        { path: 'revenue', component: () => import('@/pages/admin/AdminRevenue.vue') },
      ],
    },

    // ── 404 ───────────────────────────────────────────────────
    { path: '/:pathMatch(.*)*', name: 'not-found', component: () => import('@/pages/NotFoundPage.vue') },
  ],
})

router.beforeEach(async (to) => {
  const auth = useAuthStore()

  if (!auth.initialized) {
    await auth.initAuth()
  }

  if (to.meta.requiresAuth && !auth.isAuthenticated) {
    return { name: 'login', query: { redirect: to.fullPath } }
  }

  if (to.meta.requiresAdmin && auth.user?.role !== 'admin') {
    return { name: 'dashboard' }
  }

  if (to.meta.guestOnly && auth.isAuthenticated) {
    return { name: 'marketplace' }
  }
})

export default router
