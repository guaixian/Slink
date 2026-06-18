import request, { http } from '../utils/request'

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

// 用户信息接口（登录名存于 account；历史上曾用 email 字段名）
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
    return http.upload<ApiResponse<ImageInfo>>('/api/image/upload', file, onProgress, 'image', strategyId)
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

  // 从 URL 上传图片（失败由用户手动重试）
  uploadFromUrl: (data: { url: string; strategy_id?: number }) => {
    return http.post<ApiResponse<ImageInfo>>('/api/image/upload-url-v2', data)
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

  /** 全局上传策略（含水印配置），需登录；与管理员「上传策略」同源数据 */
  getUploadGroupConfig: () => {
    return http.get<ApiResponse<Record<string, unknown>>>('/api/image/config')
  },

  // ===== 本地图片处理（纯Go，无外部依赖）=====

  /** 本地处理能力（支持的操作/格式/上限） */
  getProcessCapabilities: () => {
    return http.get<ApiResponse<ProcessCapabilities>>('/api/image/process/capabilities')
  },

  /**
   * 本地处理图片：压缩/格式转换/高质量缩放放大/缩略图。
   * 统一以 response=json 返回 base64 与元数据，便于前端预览与下载。
   */
  processImage: (file: File, options: ProcessOptions = {}) => {
    const fd = new FormData()
    fd.append('image', file)
    fd.append('response', 'json')
    if (options.operation) fd.append('operation', options.operation)
    if (options.format) fd.append('format', options.format)
    if (options.quality != null) fd.append('quality', String(options.quality))
    if (options.width != null) fd.append('width', String(options.width))
    if (options.height != null) fd.append('height', String(options.height))
    if (options.scale != null) fd.append('scale', String(options.scale))
    if (options.max_size != null) fd.append('max_size', String(options.max_size))
    return http.post<ApiResponse<ProcessedImage>>('/api/image/process', fd, {
      headers: { 'Content-Type': 'multipart/form-data' },
      timeout: 60000,
    })
  },

  // ===== AI 图片处理（外置扩展，依赖外部服务）=====

  /** 查询 AI 外置扩展是否已配置及可用能力 */
  getAIStatus: () => {
    return http.get<ApiResponse<AIStatus>>('/api/image/ai/status')
  },

  /** 智能去水印（外置扩展，未配置返回 501） */
  aiDewatermark: (file: File, options: AIOptions = {}) => {
    return aiImageRequest('/api/image/ai/dewatermark', file, options)
  },

  /** AI 超分辨率/高清放大（外置扩展，未配置返回 501） */
  aiUpscale: (file: File, options: AIOptions = {}) => {
    return aiImageRequest('/api/image/ai/upscale', file, options)
  },

  /** 智能识别/自动打标（外置扩展，未配置返回 501） */
  aiTag: (file: File, options: AIOptions = {}) => {
    const fd = buildAIFormData(file, options)
    return http.post<ApiResponse<{ tags: AITag[]; meta?: Record<string, unknown> }>>(
      '/api/image/ai/tag',
      fd,
      { headers: { 'Content-Type': 'multipart/form-data' }, timeout: 180000 },
    )
  },
}

// 本地处理操作类型
export type ProcessOperation = 'compress' | 'convert' | 'resize' | 'thumbnail'

export interface ProcessOptions {
  operation?: ProcessOperation
  format?: string
  quality?: number
  width?: number
  height?: number
  scale?: number
  max_size?: number
}

export interface ProcessedImage {
  format: string
  mimetype: string
  width: number
  height: number
  size_bytes: number
  base64: string
  data_uri: string
  origin_name: string
}

export interface ProcessCapabilities {
  operations: string[]
  input_formats: string[]
  output_formats: string[]
  max_dimension: number
  default_quality: number
  note: string
  webp_output_local: boolean
}

export interface AIStatus {
  configured: boolean
  endpoint: string
  capabilities: string[]
  note: string
}

export interface AIOptions {
  scale?: number
  prompt?: string
  model?: string
  regions?: unknown
}

export interface AIProcessedImage {
  mimetype: string
  size_bytes: number
  base64: string
  data_uri: string
  meta?: Record<string, unknown>
}

export interface AITag {
  name: string
  score: number
}

// 组装 AI 请求的 FormData（图像类能力共用）
function buildAIFormData(file: File, options: AIOptions): FormData {
  const fd = new FormData()
  fd.append('image', file)
  if (options.scale != null) fd.append('scale', String(options.scale))
  if (options.prompt) fd.append('prompt', options.prompt)
  if (options.model) fd.append('model', options.model)
  if (options.regions != null) fd.append('regions', JSON.stringify(options.regions))
  return fd
}

// 返回图像的 AI 能力统一以 response=json 取回 base64
function aiImageRequest(url: string, file: File, options: AIOptions) {
  const fd = buildAIFormData(file, options)
  fd.append('response', 'json')
  return http.post<ApiResponse<AIProcessedImage>>(url, fd, {
    headers: { 'Content-Type': 'multipart/form-data' },
    timeout: 180000,
  })
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

  // ===== AI 设置（图像扩展 / LLM / Embedding 平台与密钥）=====

  /** 读取 AI 设置（密钥脱敏，env 锁定字段带 from_env 标记） */
  getAISettings: () => {
    return http.get<ApiResponse<AISettingsData>>('/api/admin/ai-settings')
  },

  /** 保存 AI 设置（密钥留空/未改则保留原值；env 锁定字段忽略） */
  updateAISettings: (values: Record<string, string>) => {
    return http.put<ApiResponse<any>>('/api/admin/ai-settings', values)
  },

  /** 测试 AI 连接，target: image|llm|embedding */
  testAISettings: (target: 'image' | 'llm' | 'embedding') => {
    return http.post<ApiResponse<{ ok: boolean; status?: number; message: string }>>(
      '/api/admin/ai-settings/test',
      { target },
    )
  },

  updateUploadPolicy: (data: Record<string, unknown>) => {
    return http.put<ApiResponse<any>>('/api/admin/upload-policy', data)
  },

  /** 全量 zip：config 目录、SQLite 数据文件、static 本地图 */
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

/** 与 Gin `admin/backup/import` 返回的 JSON 一致 */
export interface BackupImportResult {
  status: boolean
  message?: string
  error?: string
}

// 单个 AI 配置字段（密钥已脱敏；from_env 表示由环境变量锁定，只读）
export interface AISettingField {
  key: string
  value: string
  has_value: boolean
  from_env: boolean
  is_secret: boolean
}

export interface AISettingsData {
  image: {
    endpoint: AISettingField
    token: AISettingField
    timeout: AISettingField
  }
  llm: {
    base_url: AISettingField
    api_key: AISettingField
    model: AISettingField
  }
  embedding: {
    base_url: AISettingField
    api_key: AISettingField
    model: AISettingField
    dimensions: AISettingField
  }
  note: string
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