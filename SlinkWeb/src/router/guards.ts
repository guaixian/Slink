import type { Router } from 'vue-router'
import { useUserStore } from '../stores/user'
import { http } from '../utils/request'

const authRoutes = ['/admin']

export function setupRouterGuards(router: Router) {
  router.beforeEach(async (to, from, next) => {
    if (to.path === '/init') {
      try {
        const response = await http.get('/api/init/status')
        if (response.data.code === 200 && response.data.data.is_initialized) {
          next('/')
          return
        }
      } catch {
        // 允许进入初始化页
      }
    }

    const userStore = useUserStore()

    if (!userStore.isLoggedIn) {
      const restored = userStore.initUserFromStorage()
      if (restored) {
        try {
          await userStore.getUserInfo()
        } catch {
          userStore.logout()
          next('/')
          return
        }
      }
    } else {
      try {
        await userStore.getUserInfo()
      } catch {
        userStore.logout()
        next('/')
        return
      }
    }

    if (to.path === '/' && userStore.isLoggedIn) {
      next('/admin')
      return
    }

    const requiresAuth = authRoutes.some((route) => to.path.startsWith(route))
    if (requiresAuth && !userStore.isLoggedIn) {
      next('/')
      return
    }

    next()
  })
}
