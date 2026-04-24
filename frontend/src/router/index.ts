import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '../stores/user'

const routes = [
  {
    path: '/',
    component: () => import('../components/AppLayout.vue'),
    children: [
      { path: '', component: () => import('../pages/Home.vue'), meta: { public: true } },
      { path: 'problems', component: () => import('../pages/ProblemList.vue') },
      { path: 'problems/:id', component: () => import('../pages/ProblemDetail.vue'), meta: { wide: true } },
      { path: 'submissions', component: () => import('../pages/MySubmissions.vue') },
      { path: 'submissions/:id', component: () => import('../pages/SubmissionDetail.vue') },
      {
        path: 'submissions/:id/diff/:otherId?',
        component: () => import('../pages/SubmissionDiff.vue'),
      },
      { path: 'problemsets', component: () => import('../pages/ProblemsetList.vue') },
      { path: 'problemsets/:id', component: () => import('../pages/ProblemsetDetail.vue') },
      { path: 'ranking', component: () => import('../pages/Ranking.vue') },
      { path: 'me', component: () => import('../pages/Me.vue') },
    ],
  },
  {
    path: '/admin',
    component: () => import('../components/AdminLayout.vue'),
    meta: { adminOnly: true },
    children: [
      { path: '', component: () => import('../pages/admin/Dashboard.vue') },
      { path: 'users', component: () => import('../pages/admin/Users.vue') },
      { path: 'users/:id', component: () => import('../pages/admin/UserProfile.vue') },
      { path: 'problems', component: () => import('../pages/admin/Problems.vue') },
      { path: 'problems/:id', component: () => import('../pages/admin/ProblemEdit.vue') },
      { path: 'problemsets', component: () => import('../pages/admin/Problemsets.vue') },
      { path: 'tags', component: () => import('../pages/admin/Tags.vue') },
      { path: 'submissions', component: () => import('../pages/admin/Submissions.vue') },
      { path: 'ai', component: () => import('../pages/admin/AITasks.vue') },
    ],
  },
]

const router = createRouter({ history: createWebHistory(), routes })

router.beforeEach((to) => {
  const user = useUserStore()
  if (!user.token) user.hydrate()

  // Only the home page is public. Gating non-logged-in visitors redirects to
  // the home page with ?next=<original>, where Home.vue renders an inline
  // login card. 登录成功后 LoginCard 会用 next 跳回。
  const isPublic = to.meta.public === true
  if (isPublic) return true

  if (!user.isLoggedIn) {
    return { path: '/', query: { next: to.fullPath } }
  }
  if (to.matched.some((m) => m.meta.adminOnly) && !user.isAdmin) {
    return { path: '/problems' }
  }
  return true
})

export default router
