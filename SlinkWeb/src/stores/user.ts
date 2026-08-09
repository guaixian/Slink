import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import { http } from '../utils/request'

// 用户信息接口
interface UserInfo {
  user_id: number
  account: string
  name: string
  group_id: number
  is_admin: boolean
  token: string
}

// API响应接口
interface ApiResponse<T> {
  status: boolean
  message: string
  data: T
}

export const useUserStore = defineStore('user', () => {
  // 状态
  const userInfo = ref<UserInfo | null>(null)
  const isLoggedIn = ref(false)
  const isLoading = ref(false)
  const systemConfig = ref<any[]>([])
  const isSystemConfigLoaded = ref(false)
  // 用户信息上次成功拉取时间：路由守卫每次跳转都会调用 getUserInfo，
  // TTL 内直接复用，避免每次页面切换都打一次 /api/user/info
  const lastUserInfoAt = ref(0)
  const USER_INFO_TTL = 60_000

  // 计算属性
  const token = computed(() => userInfo.value?.token || '')
  const isAdmin = computed(() => userInfo.value?.is_admin || false)
  const userId = computed(() => userInfo.value?.user_id || 0)
  const userName = computed(() => userInfo.value?.name || '')
  const userAccount = computed(() => userInfo.value?.account || '')
  
  // 系统配置相关计算属性
  const isEmailVerificationEnabled = computed(() => {
    const emailVerifyConfig = systemConfig.value.find(config => config.ConfigKey === 'email_verify')
    return emailVerifyConfig?.Value === 'true'
  })
  
  const isRegisterEnabled = computed(() => {
    const registerConfig = systemConfig.value.find(config => config.ConfigKey === 'enable_register')
    return registerConfig?.Value === 'true'
  })
  
  const isGalleryEnabled = computed(() => {
    const galleryConfig = systemConfig.value.find(config => config.ConfigKey === 'enable_gallery')
    return galleryConfig?.Value === 'true'
  })
  
  const isApiEnabled = computed(() => {
    const apiConfig = systemConfig.value.find(config => config.ConfigKey === 'enable_api')
    return apiConfig?.Value === 'true'
  })
  
  const isGuestUploadEnabled = computed(() => {
    const guestUploadConfig = systemConfig.value.find(config => config.ConfigKey === 'guest_upload')
    return guestUploadConfig?.Value === 'true'
  })
  
  const defaultStorageGB = computed(() => {
    const storageConfig = systemConfig.value.find(config => config.ConfigKey === 'default_storage_gb')
    return storageConfig?.Value ? parseInt(storageConfig.Value) : 5
  })

  // 从本地存储恢复token
  const initUserFromStorage = () => {
    console.log('尝试从localStorage恢复token...')
    const token = localStorage.getItem('token')
    console.log('localStorage中的token:', token ? '存在' : '不存在')
    
    if (token) {
      // 从localStorage恢复token，但需要重新获取用户信息来确保权限状态正确
      console.log('设置临时用户信息（仅token）')
      userInfo.value = { token } as UserInfo
      isLoggedIn.value = true
      console.log('从 localStorage 恢复token完成')
      console.log('临时userInfo:', userInfo.value)
      console.log('临时isLoggedIn:', isLoggedIn.value)
      return true
    } else {
      console.log('localStorage中没有找到token')
      return false
    }
  }

  // 保存token到本地存储
  const saveTokenToStorage = (token: string) => {
    localStorage.setItem('token', token)
  }

  // 清除本地存储的token
  const clearTokenFromStorage = () => {
    localStorage.removeItem('token')
  }

  // 登录
  const login = async (username: string, password: string): Promise<ApiResponse<UserInfo>> => {
    isLoading.value = true
    try {
      const response = await http.post<ApiResponse<any>>('/api/login', {
        username,
        password
      })
      
      if (response.data.status) {
        const d = response.data.data
        userInfo.value = {
          user_id: d.user_id,
          account: d.account,
          name: d.name,
          group_id: 1,
          is_admin: d.is_admin,
          token: d.token
        } as UserInfo
        isLoggedIn.value = true
        saveTokenToStorage(d.token)
      }
      
      return response.data
    } catch (error: any) {
      console.error('Login error:', error)
      throw error
    } finally {
      isLoading.value = false
    }
  }

  // 注册
  const register = async (email: string, password: string, name: string, verificationCode: string): Promise<ApiResponse<UserInfo>> => {
    isLoading.value = true
    try {
      const response = await http.post<ApiResponse<UserInfo>>('/api/register', {
        email,
        password,
        name,
        verification_code: verificationCode
      })
      
      if (response.data.status) {
        userInfo.value = response.data.data
        isLoggedIn.value = true
        saveTokenToStorage(response.data.data.token)
      }
      
      return response.data
    } catch (error: any) {
      console.error('Register error:', error)
      throw error
    } finally {
      isLoading.value = false
    }
  }

  // 获取系统配置
  const getSystemInfo = async (): Promise<ApiResponse<any>> => {
    try {
      const response = await http.get<ApiResponse<any>>('/api/base_config')
      if (response.data && response.data.data) {
        systemConfig.value = response.data.data
        isSystemConfigLoaded.value = true
      }
      return response.data
    } catch (error: any) {
      console.error('Get system info error:', error)
      throw error
    }
  }



  // 发送验证码
  const sendVerificationCode = async (email: string): Promise<ApiResponse<any>> => {
    try {
      const response = await http.post<ApiResponse<any>>('/api/auth/send-verification-code', {
        email
      })
      return response.data
    } catch (error: any) {
      console.error('Send verification code error:', error)
      throw error
    }
  }

  // 直接设置登录数据（用于注册后自动登录）
  const setLoginData = (token: string, userId: number, account: string, name: string) => {
    userInfo.value = {
      user_id: userId,
      account,
      name,
      group_id: 1,
      is_admin: false,
      token
    } as UserInfo
    isLoggedIn.value = true
    saveTokenToStorage(token)
  }

  // 登出
  const logout = () => {
    userInfo.value = null
    isLoggedIn.value = false
    lastUserInfoAt.value = 0
    clearTokenFromStorage()
  }

  // 获取用户信息
  const getUserInfo = async (force: boolean = false): Promise<ApiResponse<UserInfo>> => {
    // TTL 内且已有完整用户信息时直接复用（路由切换高频调用此函数）
    if (!force && userInfo.value?.user_id && Date.now() - lastUserInfoAt.value < USER_INFO_TTL) {
      return { status: true, message: 'cached', data: userInfo.value }
    }
    try {
      console.log('开始获取用户信息...')
      const response = await http.get<any>('/api/user/info')
      console.log('API完整响应:', JSON.stringify(response.data, null, 2))

      if (response.data.status) {
        // 后端返回的数据结构是嵌套的，需要提取user对象
        const userData = response.data.data.user
        console.log('提取的用户数据:', JSON.stringify(userData, null, 2))

        // 构建userInfo对象，保留token
        const currentToken = userInfo.value?.token || localStorage.getItem('token') || ''

        const newUserInfo = {
          user_id: userData.id,
          account: userData.account,
          name: userData.name,
          group_id: userData.group_id || 1,
          is_admin: userData.is_admin,
          token: currentToken
        }

        console.log('构建的新userInfo:', JSON.stringify(newUserInfo, null, 2))
        userInfo.value = newUserInfo
        lastUserInfoAt.value = Date.now()

        isLoggedIn.value = true
        console.log('用户信息设置完成')
        console.log('当前userInfo:', JSON.stringify(userInfo.value, null, 2))
        console.log('当前isLoggedIn:', isLoggedIn.value)
        console.log('当前isAdmin计算属性:', isAdmin.value)
        console.log('userInfo.is_admin直接访问:', userInfo.value?.is_admin)
      } else {
        console.log('API返回失败状态:', response.data.message)
      }
      return response.data
    } catch (error: any) {
      console.error('获取用户信息失败:', error)
      throw error
    }
  }

  // 更新用户信息
  const updateUserInfo = async (userData: Partial<UserInfo>): Promise<ApiResponse<UserInfo>> => {
    try {
      const response = await http.put<ApiResponse<UserInfo>>('/api/user/info', userData)
      if (response.data.status) {
        // 后端 /api/user/info 不返回 token，需保留本地现有 token，避免被 undefined 覆盖
        const currentToken = userInfo.value?.token || localStorage.getItem('token') || ''
        userInfo.value = { ...response.data.data, token: currentToken }
      }
      return response.data
    } catch (error: any) {
      console.error('Update user info error:', error)
      throw error
    }
  }

  return {
    // 状态
    userInfo,
    isLoggedIn,
    isLoading,
    systemConfig,
    isSystemConfigLoaded,
    
    // 计算属性
    token,
    isAdmin,
    userId,
    userName,
    userAccount,
    
    // 系统配置相关计算属性
    isEmailVerificationEnabled,
    isRegisterEnabled,
    isGalleryEnabled,
    isApiEnabled,
    isGuestUploadEnabled,
    defaultStorageGB,
    
    // 方法
    login,
    register,
    sendVerificationCode,
    logout,
    setLoginData,
    getUserInfo,
    updateUserInfo,
    getSystemInfo,
    initUserFromStorage
  }
}) 