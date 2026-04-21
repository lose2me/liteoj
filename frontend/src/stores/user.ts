import { defineStore } from 'pinia'

export interface CurrentUser {
  id: number
  username: string
  name: string
  role: 'admin' | 'student'
}

export const useUserStore = defineStore('user', {
  state: () => ({
    token: '' as string,
    user: null as CurrentUser | null,
  }),
  getters: {
    isAdmin: (s) => s.user?.role === 'admin',
    isLoggedIn: (s) => !!s.token,
  },
  actions: {
    hydrate() {
      this.token = localStorage.getItem('liteoj.token') || ''
      const raw = localStorage.getItem('liteoj.user')
      this.user = raw ? JSON.parse(raw) : null
    },
    setAuth(token: string, user: CurrentUser) {
      this.token = token
      this.user = user
      localStorage.setItem('liteoj.token', token)
      localStorage.setItem('liteoj.user', JSON.stringify(user))
    },
    logout() {
      this.token = ''
      this.user = null
      localStorage.removeItem('liteoj.token')
      localStorage.removeItem('liteoj.user')
    },
  },
})
