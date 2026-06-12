import request, { http, withUploadRetry } from '../utils/request'

// API响应接口
interface ApiResponse<T> {
  status: string
  message: string
  data: T
}

// 用户配置接口
interface UserConfig {
  default_permission: number
  default_strategy: number
  is_auto_clear_preview: number
  pasted_action: number
}

// 存储策略接口
interface StorageStrategy {
  id: number
  name: string
  introduction: string
  key: string
}

// 用户信息接口
interface UserInfo {
  id: number
  user_id: number
  account: string
  name: string
  created_at: string
  image_nums: number
  is_admin: boolean
  registered_ip: string
  token: string
}

// 完整的用户信息响应接口
interface UserInfoResponse {
  config: UserConfig
  strategies: StorageStrategy[]
  user: UserInfo
}

// 图片链接接口
interface ImageLinks {
  bbcode: string
  html: string
  markdown: string
  markdown_with_link: string
  thumbnail_url: string
  url: string
}

// 图片信息接口
interface ImageInfo {
  id: number
  pathname: string
  origin_name: string
  size: number
  mimetype: string
  md5: string
  sha1: string
  links: ImageLinks
}

// 上传进度回调
type UploadProgressCallback = (progress: number) => void

// 认证相关API
export const authAPI = {
  // 登录
  login: (username: string, password: string) => {
    return http.post<ApiResponse<UserInfo>>('/api/login', { username, password })
  },

  // 注册
  register: (email: string, password: string, name: string, verificationCode: string) => {
    return http.post<ApiResponse<UserInfo>>('/api/register', {
      email,
      password,
      name,
      verification_code: verificationCode
    })
  },

  // 发送验证码
  sendVerificationCode: (email: string) => {
    return http.post<ApiResponse<any>>('/api/auth/send-verification-code', { email })
  },

  // 登出
  logout: () => {
    return http.post<ApiResponse<any>>('/api/auth/logout')
  }
}

// 用户相关API
export const userAPI = {
  // 获取用户信息
  getUserInfo: () => {
    return http.get<ApiResponse<UserInfoResponse>>('/api/user/info')
  },

  // 获取用户配置
  getUserConfig: () => {
    return http.get<ApiResponse<UserConfig>>('/api/user/config')
  },

  // 获取用户仪表盘数据
  getDashboardData: () => {
    return http.get<ApiResponse<{
      dashboard: {
        image_count: number
        today_upload_count: number
        used_size_mb: number
      }
    }>>('/api/user/dashboard')
  },

  // 更新用户信息
  updateUserInfo: (userData: Partial<UserInfo>) => {
    return http.put<ApiResponse<UserInfo>>('/api/user/info', userData)
  },

  // 修改密码
  changePassword: (oldPassword: string, newPassword: string) => {
    return http.put<ApiResponse<any>>('/api/user/change-password', {
      old_password: oldPassword,
      new_password: newPassword
    })
  }
}

// 图片相关API
export const imageAPI = {
  // 上传单张图片
  uploadImage: (file: File, onProgress?: UploadProgressCallback, strategyId?: number) => {
    return http.upload<ApiResponse<ImageInfo>>('/api/image/upload-v2', file, onProgress, 'image', strategyId)
  },

  // 批量上传图片
  uploadMultipleImages: (files: File[], onProgress?: UploadProgressCallback) => {
    return http.uploadMultiple<ApiResponse<ImageInfo[]>>('/api/image/upload-multiple', files, onProgress, 'image')
  },

  // 获取图片列表
  getImageList: (page: number = 1, limit: number = 20) => {
    return http.get<ApiResponse<ImageInfo[]>>(`/api/image/list?page=${page}&limit=${limit}`)
  },

  // 删除图片
  deleteImage: (imageId: number) => {
    return http.delete<ApiResponse<any>>(`/api/image/${imageId}`)
  },

  // 从 URL 上传图片（失败自动重试）
  uploadFromUrl: (data: { url: string; strategy_id?: number }) => {
    return withUploadRetry(() =>
      http.post<ApiResponse<ImageInfo>>('/api/image/upload-url-v2', data)
    )
  },

  // 批量删除图片
  batchDeleteImages: (imageIds: number[]) => {
    return http.post<ApiResponse<any>>('/api/image/batch-delete', { image_ids: imageIds })
  },

  // 重命名图片
  renameImage: (imageId: number, newName: string) => {
    return http.put<ApiResponse<any>>(`/api/image/${imageId}/rename`, { new_name: newName })
  },

  // 生成二维码
  generateQRCode: (imageId: number) => {
    return http.get<Blob>(`/api/image/${imageId}/qrcode`, { responseType: 'blob' })
  },

  // 生成二维码 Base64
  generateQRCodeBase64: (imageId: number) => {
    return http.get<ApiResponse<{ qrcode: string; url: string }>>(`/api/image/${imageId}/qrcode-base64`)
  },

  /** 全局上传策略（含水印配置），需登录 */
  getUploadGroupConfig: () => {
    return http.get<ApiResponse<Record<string, unknown>>>('/api/image/config')
  }
}

// 管理员相关API
export const adminAPI = {
  // 获取所有用户列表
  getAllUsers: () => {
    return http.get<ApiResponse<UserInfo[]>>('/api/admin/users')
  },

  // 获取指定用户
  getUser: (userId: number) => {
    return http.get<ApiResponse<UserInfo>>(`/api/admin/users/${userId}`)
  },

  // 更新用户
  updateUser: (userId: number, userData: any) => {
    return http.put<ApiResponse<any>>(`/api/admin/users/${userId}`, userData)
  },

  // 删除用户
  deleteUser: (userId: number) => {
    return http.delete<ApiResponse<any>>(`/api/admin/users/${userId}`)
  },

  // 更新用户权限
  updateUserRole: (userId: number, isAdmin: boolean) => {
    return http.put<ApiResponse<any>>(`/api/admin/users/${userId}/role`, {
      is_admin: isAdmin
    })
  },

  // 获取系统统计信息
  getSystemStats: () => {
    return http.get<ApiResponse<any>>('/api/admin/stats')
  },

  // 获取所有图片列表（管理员）
  getAllImages: (page: number = 1, limit: number = 20, userId?: number) => {
    const params: any = { page, limit }
    if (userId) {
      params.user_id = userId
    }
    return http.get<ApiResponse<{ images: ImageInfo[], total: number }>>('/api/admin/images', {
      params
    })
  },

  // 获取指定用户的图片列表
  getUserImages: (userId: number, page: number = 1, limit: number = 20) => {
    return http.get<ApiResponse<ImageInfo[]>>(`/api/admin/users/${userId}/images`, {
      params: { page, limit }
    })
  },

  // 管理员删除图片
  adminDeleteImage: (imageId: number) => {
    return http.delete<ApiResponse<any>>(`/api/admin/images/${imageId}`)
  },

  // 获取系统配置
  getSystemConfigs: () => {
    return http.get<ApiResponse<any>>('/api/admin/configs')
  },

  // 更新系统配置
  updateSystemConfigs: (configs: any) => {
    return http.put<ApiResponse<any>>('/api/admin/configs', configs)
  },

  getUploadPolicy: () => {
    return http.get<ApiResponse<any>>('/api/admin/upload-policy')
  },

  updateUploadPolicy: (data: Record<string, unknown>) => {
    return http.put<ApiResponse<any>>('/api/admin/upload-policy', data)
  },

  exportFullBackup: () => {
    return http.getBlob('/api/admin/backup/export')
  },
  importFullBackup: (file: File) => {
    const formData = new FormData()
    formData.append('file', file)
    formData.append('confirm', 'RESTORE')
    return request.post<BackupImportResult>('/api/admin/backup/import', formData, {
      timeout: 0,
      maxContentLength: Infinity,
      maxBodyLength: Infinity
    })
  }
}

export interface BackupImportResult {
  status: boolean
  message?: string
  error?: string
}

// 存储策略接口
interface StorageStrategy {
  id: number
  name: string
  introduction: string
  key: string
  configs: string
  image_count: number
  total_size: number
  used_size_mb: number
}

// 存储策略配置接口
interface StorageConfig {
  url: string
  root: string
  queries: string | null
  auth_type: string
}

// 系统设置相关API
export const systemAPI = {
  // 获取系统设置
  getSystemSettings: () => {
    return http.get<ApiResponse<any>>('/api/system/settings')
  },

  // 更新系统设置
  updateSystemSettings: (settings: any) => {
    return http.put<ApiResponse<any>>('/api/system/settings', settings)
  },

  // 获取存储策略列表
  getStorageStrategies: () => {
    return http.get<ApiResponse<StorageStrategy[]>>('/api/storages/strategies')
  },

  // 创建存储策略
  createStorageStrategy: (strategy: {
    name: string
    introduction: string
    key: string
    configs: StorageConfig
  }) => {
    return http.post<ApiResponse<StorageStrategy>>('/api/storages/strategies', strategy)
  },

  // 更新存储策略
  updateStorageStrategy: (id: number, strategy: {
    name: string
    introduction: string
    key: string
    configs: StorageConfig
  }) => {
    return http.put<ApiResponse<StorageStrategy>>(`/api/storages/strategies/${id}`, strategy)
  },

  // 删除存储策略
  deleteStorageStrategy: (id: number) => {
    return http.delete<ApiResponse<any>>(`/api/storages/strategies/${id}`)
  },

  // 获取存储策略详情
  getStorageStrategyDetail: (id: number) => {
    return http.get<ApiResponse<StorageStrategy>>(`/api/storages/strategies/${id}`)
  }
}

// 初始化相关API
export const initAPI = {
  // 获取初始化状态
  getInitStatus: () => {
    return http.get<ApiResponse<{ is_initialized: boolean; admin_account?: string; admin_email?: string }>>('/api/init/status')
  },

  // 执行初始化设置
  performSetup: (data: {
    admin_account: string
    admin_password: string
    database_config: any
    cache_type: string
    cache_path?: string
  }) => {
    return http.post<ApiResponse<any>>('/api/init/setup', data)
  },

  // 测试数据库连接
  testDatabaseConnection: (config: any) => {
    return http.post<ApiResponse<{ success: boolean }>>('/api/init/test-db', config)
  }
}

export default {
  auth: authAPI,
  user: userAPI,
  image: imageAPI,
  admin: adminAPI,
  system: systemAPI,
  init: initAPI
} 