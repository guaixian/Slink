import axios from 'axios'
import type { AxiosInstance, AxiosRequestConfig, AxiosResponse, AxiosProgressEvent, InternalAxiosRequestConfig } from 'axios'
import { useUserStore } from '../stores/user'

// 与后端同域部署时留空；本地 Vite 开发见 vite.config 的 /api 代理。需跨域时设 VITE_API_BASE。
const baseURL = import.meta.env.VITE_API_BASE ?? ''
// 创建axios实例
const request: AxiosInstance = axios.create({
  baseURL,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// 请求拦截器
request.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const userStore = useUserStore()
    const token = userStore.token
    
    if (token) {
      config.headers.set('Authorization', `Bearer ${token}`)
    }
    
    return config
  },
  (error: any) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
request.interceptors.response.use(
  (response: AxiosResponse) => {
    return response
  },
  (error: any) => {
    if (error.response?.status === 401) {
      // token过期或无效，清除用户信息
      const userStore = useUserStore()
      userStore.logout()
    }
    return Promise.reject(error)
  }
)

// 封装请求方法
export const http = {
  // GET请求
  get<T = any>(url: string, config?: AxiosRequestConfig): Promise<AxiosResponse<T>> {
    return request.get(url, config)
  },

  getBlob(url: string, config?: AxiosRequestConfig): Promise<AxiosResponse<Blob>> {
    return request.get<Blob>(url, { ...config, responseType: 'blob', timeout: config?.timeout ?? 0 })
  },

  // POST请求
  post<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<AxiosResponse<T>> {
    return request.post(url, data, config)
  },

  // PUT请求
  put<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<AxiosResponse<T>> {
    return request.put(url, data, config)
  },

  // DELETE请求
  delete<T = any>(url: string, config?: AxiosRequestConfig): Promise<AxiosResponse<T>> {
    return request.delete(url, config)
  },

  // 文件上传（失败由用户手动重试）
  upload<T = any>(url: string, file: File, onProgress?: (progress: number) => void, fieldName: string = 'file', strategyId?: number): Promise<AxiosResponse<T>> {
    const formData = new FormData()
    formData.append(fieldName, file)
    if (strategyId !== undefined) {
      formData.append('strategy_id', strategyId.toString())
    }
    if (onProgress) {
      onProgress(0)
    }
    return request.post(url, formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
      onUploadProgress: (progressEvent: AxiosProgressEvent) => {
        if (onProgress && progressEvent.total) {
          const progress = Math.round((progressEvent.loaded * 100) / progressEvent.total)
          onProgress(progress)
        }
      },
    })
  },

  // 批量文件上传（失败由用户手动重试）
  uploadMultiple<T = any>(url: string, files: File[], onProgress?: (progress: number) => void, fieldName: string = 'files'): Promise<AxiosResponse<T>> {
    const formData = new FormData()
    files.forEach((file, index) => {
      formData.append(`${fieldName}[${index}]`, file)
    })
    if (onProgress) {
      onProgress(0)
    }
    return request.post(url, formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
      onUploadProgress: (progressEvent: AxiosProgressEvent) => {
        if (onProgress && progressEvent.total) {
          const progress = Math.round((progressEvent.loaded * 100) / progressEvent.total)
          onProgress(progress)
        }
      },
    })
  },
}

export default request 