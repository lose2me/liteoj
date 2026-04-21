import axios from 'axios'
import router from '../router'
import { useUserStore } from '../stores/user'

export const http = axios.create({ baseURL: '/api' })

http.interceptors.request.use((cfg) => {
  const user = useUserStore()
  if (user.token) cfg.headers.Authorization = `Bearer ${user.token}`
  return cfg
})

http.interceptors.response.use(
  (r) => r,
  (err) => {
    if (err?.response?.status === 401) {
      const user = useUserStore()
      user.logout()
      // 记住当前路径，登录后跳回。避免把 / 自己也回跳造成一直停留。
      const here = location.pathname + location.search
      const next = here === '/' ? '' : here
      router.replace({ path: '/', query: next ? { next } : undefined })
    }
    return Promise.reject(err)
  },
)
