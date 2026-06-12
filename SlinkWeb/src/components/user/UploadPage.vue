<template>
  <div class="upload-page">
    <div class="upload-card">
      <div class="card-header">
        <div class="header-left">
          <h2 class="text-xl font-bold text-neutral-800 mb-2">
            <i class="fa fa-cloud-upload mr-2 text-primary"></i>
            Image Upload
          </h2>
          <p class="text-neutral-600 text-sm">
            最大可上传 {{ maxFileSize }} 的图片，上传队列最多 {{ maxQueueSize }} 张，本站已托管 {{ totalImages }} 张图片。
          </p>
        </div>
        <div class="header-right">
          <div class="strategy-selector">
            <label class="strategy-label">
              <i class="fa fa-database mr-2"></i>
              存储策略
            </label>
            <select v-model="selectedStrategy" class="strategy-select" @change="handleStrategyChange">
              <option
                v-for="strategy in availableStrategies"
                :key="strategy.id"
                :value="strategy.id"
              >
                {{ strategy.name }}
              </option>
            </select>
          </div>
        </div>
      </div>
      
      <div 
        ref="uploadArea"
        class="upload-area"
        :class="{ 'upload-area-active': isDragOver }"
        @dragenter.prevent="handleDragEnter"
        @dragover.prevent="handleDragOver"
        @dragleave.prevent="handleDragLeave"
        @drop.prevent="handleDrop"
        @click="triggerFileInput"
      >
        <div class="upload-icon">
          <i class="fa fa-cloud-upload-alt text-5xl text-neutral-300"></i>
        </div>
        <h3 class="text-lg font-medium text-neutral-700 mb-2">
          拖拽文件到这里，支持多文件同时上传
        </h3>
        <p class="text-neutral-500 text-sm">
          点击上面的图标上传全部已选择文件
        </p>
        <input 
          ref="fileInput"
          type="file" 
          multiple 
          class="hidden" 
          @change="handleFileSelect"
          accept="image/*"
        />
      </div>

      <!-- URL 上传区域 -->
      <div class="url-upload-section">
        <div class="url-upload-header">
          <h3 class="text-md font-medium text-neutral-700 mb-2">
            <i class="fa fa-link mr-2 text-primary"></i>
            从 URL 上传图片
          </h3>
        </div>
        <div class="url-upload-form">
          <input
            v-model="urlInput"
            type="text"
            placeholder="输入图片 URL，例如：https://example.com/image.jpg"
            class="url-input"
            @keyup.enter="handleUrlUpload"
          />
          <button
            class="url-upload-btn"
            @click="handleUrlUpload"
            :disabled="!urlInput || isUrlUploading"
          >
            <i class="fa" :class="isUrlUploading ? 'fa-spinner fa-spin' : 'fa-download'"></i>
            {{ isUrlUploading ? '上传中...' : '从 URL 上传' }}
          </button>
        </div>
        <p class="text-neutral-500 text-xs mt-2">
          支持直接从网络 URL 下载并上传图片
        </p>
      </div>

             <!-- 上传进度状态栏 -->
               <div v-if="uploadQueue.length > 0" class="upload-queue">
          <div class="queue-list">
                                    <div 
               v-for="item in uploadQueue" 
               :key="item.id" 
               class="queue-item"
               :class="{ 'uploading': item.status === 'uploading' }"
               :style="item.status === 'uploading' ? { '--progress-width': item.progress + '%' } : {}"
             >
              <div class="item-content">
                <div class="item-preview">
                  <img
                    v-if="item.preview || item.result"
                    :src="item.preview || fixImageUrl(item.result?.links?.url)"
                    :alt="item.file.name"
                    class="preview-image"
                  >
                  <div v-else class="preview-placeholder">
                    <i class="fa fa-image"></i>
                  </div>
                </div>
                <div class="item-details">
                  <div class="item-name">{{ item.result ? item.result.origin_name : item.file.name }}</div>
                  <div class="item-actions">
                    <button
                        v-if="item.status === 'waiting'"
                        class="action-btn cancel-btn"
                        @click="removeFromQueue(item)"
                    >
                      <i class="fa fa-times"></i>
                      <span class="btn-text">取消</span>
                    </button>
                    <button
                        v-if="item.status === 'waiting'"
                        class="action-btn upload-btn"
                        @click="uploadFile(item)"
                    >
                      <i class="fa fa-upload"></i>
                      <span class="btn-text">上传</span>
                    </button>
                    <button
                        v-if="item.status === 'uploading'"
                        class="action-btn remove-btn"
                        @click="removeFromQueue(item)"
                    >
                      <i class="fa fa-times"></i>
                    </button>
                  </div>
                  <div class="item-row">
                    <span class="status-text" :class="getStatusClass(item.status)">
                      {{ getStatusText(item.status) }}
                    </span>
                    <span class="item-size">
                      {{ item.result ? formatMBSize(item.result.size) : formatFileSize(item.file.size) }}
                    </span>

                  </div>
                </div>
                             </div>
             </div>
         </div>
       </div>

             <!-- 链接生成区域 -->
               <div v-if="uploadedImages.length > 0" class="link-generator">
          <div class="tab-container">
           <div class="tab-header">
             <button 
               v-for="tab in linkTabs" 
               :key="tab.key"
               class="tab-btn"
               :class="{ 'tab-active': activeTab === tab.key }"
               @click="activeTab = tab.key"
             >
               {{ tab.label }}
             </button>
           </div>
           <div class="tab-content">
             <div class="images-links-list">
               <div 
                 v-for="(image, index) in uploadedImages" 
                 :key="image.id" 
                 class="image-link-item"
               >
                 <div class="link-display">
                   <input 
                     :value="getImageLink(image, index)" 
                     readonly 
                     class="link-input"
                     :ref="`linkInput${index}`"
                   >
                   <button class="copy-btn" @click="copyImageLink(index)">
                     <i class="fa fa-copy"></i>
                   </button>
                 </div>
               </div>
             </div>
           </div>
         </div>
       </div>
      
      <!-- 上传须知 -->
      <div class="upload-notice">
        <h4 class="font-medium text-sm mb-3 flex items-center">
          <i class="fa fa-exclamation-triangle mr-2 text-orange-500"></i>
          上传须知
        </h4>
        <ul class="space-y-2 text-neutral-600 text-xs">
          <li class="flex items-start">
            <i class="fa fa-check-circle text-green-500 mt-0.5 mr-2"></i>
            <span>支持 JPG、PNG、GIF、WebP、SVG 等常见图片格式</span>
          </li>
          <li class="flex items-start">
            <i class="fa fa-clock-o text-orange-500 mt-0.5 mr-2"></i>
            <span>单文件最大支持 {{ maxFileSize }}，总上传限制 {{ maxQueueSize }} 张</span>
          </li>
          <li class="flex items-start">
            <i class="fa fa-ban text-red-500 mt-0.5 mr-2"></i>
            <span>请勿上传违反法律法规的内容，我们保留追究责任的权利</span>
          </li>
        </ul>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { userAPI, imageAPI } from '../../api'
import { useMessage } from '../../composables/useMessage'
import { usePendingUploadStore } from '../../stores/pending-upload'

const { toast, confirm } = useMessage()
const pendingUploadStore = usePendingUploadStore()

// 响应式数据
const isDragOver = ref(false)
const uploadArea = ref<HTMLElement>()
const fileInput = ref<HTMLInputElement>()
const linkInput = ref<HTMLInputElement>()

// URL 上传
const urlInput = ref('')
const isUrlUploading = ref(false)

// 用户配置
const userConfig = ref({
  default_permission: 0,
  default_strategy: 1,
  is_auto_clear_preview: 0,
  pasted_action: 1
})

// 存储策略
const availableStrategies = ref<Array<{id: number, name: string, introduction: string}>>([])
const selectedStrategy = ref<number>(1)

// 上传队列
const uploadQueue = ref<Array<{
  id: string
  file: File
  preview: string | null
  status: 'waiting' | 'uploading' | 'success' | 'error'
  progress: number
  result?: any
  error?: string
}>>([])

// 已上传图片
const uploadedImages = ref<any[]>([])

// 链接生成
const activeTab = ref('url')
const linkTabs = [
  { key: 'url', label: 'URL' },
  { key: 'html', label: 'HTML' },
  { key: 'bbcode', label: 'BBCode' },
  { key: 'markdown', label: 'Markdown' },
  { key: 'markdown-link', label: 'Markdown with link' },
  { key: 'thumbnail', label: 'Thumbnail url' }
]

// 配置数据
const maxFileSize = '14.77 MB'
const maxQueueSize = 10
const totalImages = ref(0)

// 页面初始化
onMounted(async () => {
  await loadUserConfig()
  await loadStrategies()
  await loadTotalImages()

  const fromHome = pendingUploadStore.takeFiles()
  if (fromHome.length) {
    const space = Math.max(0, maxQueueSize - uploadQueue.value.length)
    addFilesToQueue(fromHome.slice(0, space))
    if (fromHome.length > space) {
      toast.warning(`上传队列最多 ${maxQueueSize} 张，已从首页带入前 ${space} 张`)
    }
  }

  // 添加粘贴事件监听器
  document.addEventListener('paste', handlePaste)
})

// 组件卸载时移除事件监听器
onUnmounted(() => {
  document.removeEventListener('paste', handlePaste)
})

// 加载用户配置
const loadUserConfig = async () => {
  try {
    const response = await userAPI.getUserInfo()
    console.log('用户信息API响应:', response)
    if (response.data && response.data.status) {
      const data = response.data.data
      if (data.config) {
        userConfig.value = data.config
        // 设置默认选中的策略
        selectedStrategy.value = data.config.default_strategy || 1
        console.log('用户配置:', userConfig.value)
        console.log('默认策略ID:', selectedStrategy.value)
      }
    }
  } catch (error) {
    console.error('加载用户配置失败:', error)
  }
}

// 加载存储策略列表
const loadStrategies = async () => {
  try {
    const response = await userAPI.getUserInfo()
    console.log('加载策略API响应:', response)
    if (response.data && response.data.status) {
      const data = response.data.data
      if (data.strategies && Array.isArray(data.strategies)) {
        availableStrategies.value = data.strategies.map((s: any) => ({
          id: s.id,
          name: s.name,
          introduction: s.introduction || ''
        }))
        console.log('可用策略列表:', availableStrategies.value)
      }
    }
  } catch (error) {
    console.error('加载存储策略失败:', error)
  }
}

// 处理策略变化
const handleStrategyChange = () => {
  console.log('切换存储策略:', selectedStrategy.value)
  const strategy = availableStrategies.value.find(s => s.id === selectedStrategy.value)
  if (strategy) {
    console.log('当前选中策略:', strategy.name)
  }
}

// 加载总图片数
const loadTotalImages = async () => {
  try {
    const response = await userAPI.getDashboardData()
    if (response.data?.status === 'success' && response.data.data?.dashboard) {
      totalImages.value = Number(response.data.data.dashboard.image_count) || 0
    }
  } catch (e) {
    console.error('加载图片统计失败:', e)
    totalImages.value = 0
  }
}

// 拖拽事件处理
const handleDragEnter = (e: DragEvent) => {
  e.preventDefault()
  isDragOver.value = true
}

const handleDragOver = (e: DragEvent) => {
  e.preventDefault()
  isDragOver.value = true
}

const handleDragLeave = (e: DragEvent) => {
  e.preventDefault()
  isDragOver.value = false
}

const handleDrop = (e: DragEvent) => {
  e.preventDefault()
  isDragOver.value = false
  
  const files = e.dataTransfer?.files
  if (files && files.length > 0) {
    addFilesToQueue(Array.from(files))
  }
}

// 文件选择处理
const triggerFileInput = () => {
  fileInput.value?.click()
}

const handleFileSelect = (event: Event) => {
  const target = event.target as HTMLInputElement
  if (target.files && target.files.length > 0) {
    addFilesToQueue(Array.from(target.files))
  }
}

// 处理粘贴事件
const handlePaste = (event: ClipboardEvent) => {
  const items = event.clipboardData?.items
  if (!items) return
  
  const files: File[] = []
  
  for (let i = 0; i < items.length; i++) {
    const item = items[i]
    if (item.type.startsWith('image/')) {
      const file = item.getAsFile()
      if (file) {
        files.push(file)
      }
    }
  }
  
  if (files.length > 0) {
    addFilesToQueue(files)
    
    // 根据pasted_action配置决定是否自动上传
    if (userConfig.value.pasted_action === 1) {
      // 立即上传粘贴的图片
      uploadQueue.value.forEach(item => {
        if (item.status === 'waiting' && files.some(f => f.name === item.file.name)) {
          uploadFile(item)
        }
      })
    }
  }
}

// 处理 URL 上传
const handleUrlUpload = async () => {
  if (!urlInput.value || isUrlUploading.value) return

  // 验证 URL 格式
  try {
    new URL(urlInput.value)
  } catch (error) {
    toast.warning('请输入有效的 URL 地址')
    return
  }

  isUrlUploading.value = true

  try {
    console.log('从URL上传图片，使用策略ID:', selectedStrategy.value)

    const response = await imageAPI.uploadFromUrl({
      url: urlInput.value,
      strategy_id: selectedStrategy.value
    })

    if (response.data.status) {
      // 上传成功，添加到已上传列表
      uploadedImages.value.unshift(response.data.data)

      // 清空输入框
      urlInput.value = ''

      toast.success('从 URL 上传成功！')
    } else {
      toast.error(`上传失败：${response.data.message || '未知错误'}`)
    }
  } catch (error: any) {
    console.error('URL 上传失败:', error)
    toast.error(`上传失败：${error.response?.data?.error || error.message || '网络错误'}`)
  } finally {
    isUrlUploading.value = false
  }
}

// 添加文件到队列
const addFilesToQueue = (files: File[]) => {
  files.forEach(file => {
    if (file.type.startsWith('image/')) {
      const id = Date.now() + Math.random().toString(36).substr(2, 9)
      const preview = userConfig.value.is_auto_clear_preview === 0 ? URL.createObjectURL(file) : null
      
      uploadQueue.value.push({
        id,
        file,
        preview,
        status: 'waiting',
        progress: 0
      })
    }
  })
  
  // 不自动上传，让用户手动选择
}

// 上传文件
const uploadFile = async (item: any) => {
  if (item.status !== 'waiting') return

  item.status = 'uploading'
  item.progress = 0

  try {
    // 创建FormData并添加策略ID
    const formData = new FormData()
    formData.append('image', item.file)
    formData.append('strategy_id', selectedStrategy.value.toString())

    console.log('上传图片，使用策略ID:', selectedStrategy.value)

    const response = await imageAPI.uploadImage(item.file, (progress) => {
      item.progress = progress
    }, selectedStrategy.value)

    if (response.data.status) {
      item.status = 'success'
      item.result = response.data.data
      uploadedImages.value.push(response.data.data)

      // 显示成功提示
      showSuccessMessage(`图片 "${response.data.data.origin_name}" 上传成功！`)

      // 清除预览
      if (item.preview) {
        URL.revokeObjectURL(item.preview)
        item.preview = null
      }
    } else {
      item.status = 'error'
      item.error = response.data.message
      showErrorMessage(response.data.message || '上传失败')
    }
  } catch (error) {
    item.status = 'error'
    item.error = '上传失败'
    showErrorMessage('上传失败，请重试')
    console.error('上传失败:', error)
  }
}

// 从队列移除
const removeFromQueue = (item: any) => {
  const index = uploadQueue.value.findIndex(i => i.id === item.id)
  if (index > -1) {
    if (item.preview) {
      URL.revokeObjectURL(item.preview)
    }
    uploadQueue.value.splice(index, 1)
  }
}

// 格式化文件大小
const formatFileSize = (bytes: number) => {
  if (bytes === 0) return '0 Bytes'
  const k = 1024
  const sizes = ['Bytes', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

// 格式化MB大小
const formatMBSize = (mb: number) => {
  if (mb < 1) {
    return (mb * 1024).toFixed(2) + ' KB'
  }
  return mb.toFixed(2) + ' MB'
}

// 修复URL中的反斜杠问题
const fixImageUrl = (url: string): string => {
  if (!url) return ''
  // 将反斜杠替换为正斜杠，确保URL格式正确
  return url.replace(/\\/g, '/')
}

// 获取状态样式
const getStatusClass = (status: string) => {
  switch (status) {
    case 'success': return 'status-success'
    case 'error': return 'status-error'
    case 'uploading': return 'status-uploading'
    default: return 'status-waiting'
  }
}

// 获取状态文本
const getStatusText = (status: string) => {
  switch (status) {
    case 'success': return '上传成功'
    case 'error': return '上传失败'
    case 'uploading': return '上传中...'
    default: return '等待上传'
  }
}



// 获取图片链接
const getImageLink = (image: any, index: number) => {
  // 使用API返回的links对象
  if (image.links) {
    switch (activeTab.value) {
      case 'html':
        return fixImageUrl(image.links.html || '')
      case 'bbcode':
        return fixImageUrl(image.links.bbcode || '')
      case 'markdown':
        return fixImageUrl(image.links.markdown || '')
      case 'markdown-link':
        return fixImageUrl(image.links.markdown_with_link || '')
      case 'thumbnail':
        return fixImageUrl(image.links.thumbnail_url || '')
      default:
        return fixImageUrl(image.links.url || '')
    }
  }
  
  // 降级处理，如果没有links对象
  const url = fixImageUrl(image.links?.url || image.url || '')
  const originalName = image.origin_name || image.original_name || 'image'
  
  switch (activeTab.value) {
    case 'html':
      return `<img src="${url}" alt="${originalName}" title="${originalName}" />`
    case 'bbcode':
      return `[img]${url}[/img]`
    case 'markdown':
      return `![${originalName}](${url})`
    case 'markdown-link':
      return `[![${originalName}](${url})](${url})`
    case 'thumbnail':
      return fixImageUrl(image.links?.thumbnail_url || url)
    default:
      return url
  }
}

// 获取当前链接
const getCurrentLink = () => {
  if (uploadedImages.value.length === 0) return ''
  
  const image = uploadedImages.value[uploadedImages.value.length - 1]
  
  // 使用API返回的links对象
  if (image.links) {
    switch (activeTab.value) {
      case 'html':
        return fixImageUrl(image.links.html || '')
      case 'bbcode':
        return fixImageUrl(image.links.bbcode || '')
      case 'markdown':
        return fixImageUrl(image.links.markdown || '')
      case 'markdown-link':
        return fixImageUrl(image.links.markdown_with_link || '')
      case 'thumbnail':
        return fixImageUrl(image.links.thumbnail_url || '')
      default:
        return fixImageUrl(image.links.url || '')
    }
  }
  
  // 降级处理，如果没有links对象
  const url = fixImageUrl(image.links?.url || image.url || '')
  const originalName = image.origin_name || image.original_name || 'image'
  
  switch (activeTab.value) {
    case 'html':
      return `<img src="${url}" alt="${originalName}" title="${originalName}" />`
    case 'bbcode':
      return `[img]${url}[/img]`
    case 'markdown':
      return `![${originalName}](${url})`
    case 'markdown-link':
      return `[![${originalName}](${url})](${url})`
    case 'thumbnail':
      return fixImageUrl(image.links?.thumbnail_url || url)
    default:
      return url
  }
}

// 显示成功消息
const showSuccessMessage = (message: string) => {
  toast.success(message)
}

// 显示错误消息
const showErrorMessage = (message: string) => {
  toast.error(message)
}



// 复制图片链接
const copyImageLink = async (index: number) => {
  const image = uploadedImages.value[index]
  if (!image) return
  
  const link = getImageLink(image, index)
  if (link) {
    try {
      await navigator.clipboard.writeText(link)
      showSuccessMessage(`图片 "${image.origin_name}" 的链接已复制到剪贴板！`)
    } catch (error) {
      // 降级方案
      const linkInput = document.querySelector(`[ref="linkInput${index}"]`) as HTMLInputElement
      if (linkInput) {
        linkInput.select()
        document.execCommand('copy')
        showSuccessMessage(`图片 "${image.origin_name}" 的链接已复制到剪贴板！`)
      }
    }
  }
}

// 复制链接
const copyLink = async () => {
  const link = getCurrentLink()
  if (link) {
    try {
      await navigator.clipboard.writeText(link)
      showSuccessMessage('链接已复制到剪贴板！')
    } catch (error) {
      // 降级方案
      linkInput.value?.select()
      document.execCommand('copy')
      showSuccessMessage('链接已复制到剪贴板！')
    }
  }
}
</script>

<style scoped>
.upload-page {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 20px;
}

.upload-card {
  background-color: #fff;
  border-radius: 12px;
  padding: 32px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.05);
}

.card-header {
  margin-bottom: 24px;
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 24px;
}

.header-left {
  flex: 1;
}

.header-right {
  flex-shrink: 0;
}

.strategy-selector {
  display: flex;
  flex-direction: column;
  gap: 8px;
  min-width: 200px;
}

.strategy-label {
  font-size: 14px;
  font-weight: 500;
  color: #374151;
  display: flex;
  align-items: center;
}

.strategy-select {
  padding: 8px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 14px;
  background-color: white;
  color: #374151;
  cursor: pointer;
  transition: all 0.2s ease;
}

.strategy-select:hover {
  border-color: #3b82f6;
}

.strategy-select:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.upload-area {
  border: 2px dashed #d1d5db;
  border-radius: 8px;
  padding: 80px 32px;
  text-align: center;
  cursor: pointer;
  transition: all 0.3s ease;
  background-color: #fafafa;
}

.upload-area:hover {
  border-color: #3b82f6;
  background-color: #f0f9ff;
}

.upload-area-active {
  border-color: #3b82f6;
  background-color: #eff6ff;
}

.upload-icon {
  margin-bottom: 16px;
}

.upload-icon i {
  transition: transform 0.3s ease;
}

.upload-area:hover .upload-icon i {
  transform: scale(1.1);
  color: #3b82f6;
}

/* URL 上传区域样式 */
.url-upload-section {
  margin-top: 24px;
  padding: 20px;
  background-color: #f8f9fa;
  border-radius: 8px;
  border: 1px solid #e5e7eb;
}

.url-upload-header {
  margin-bottom: 12px;
}

.url-upload-form {
  display: flex;
  gap: 12px;
}

.url-input {
  flex: 1;
  padding: 10px 16px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 14px;
  transition: all 0.2s ease;
}

.url-input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.url-upload-btn {
  padding: 10px 24px;
  background-color: #3b82f6;
  color: white;
  border: none;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  white-space: nowrap;
}

.url-upload-btn:hover:not(:disabled) {
  background-color: #2563eb;
  transform: translateY(-1px);
  box-shadow: 0 4px 6px rgba(59, 130, 246, 0.2);
}

.url-upload-btn:disabled {
  background-color: #9ca3af;
  cursor: not-allowed;
  opacity: 0.6;
}

.url-upload-btn i {
  margin-right: 6px;
}

/* 上传队列样式 */
.upload-queue {
  margin-top: 20px;
  background-color: #f8f9fa;
  border-radius: 10px;
  padding: 20px;
  border: 1px solid #e5e7eb;
}



.queue-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.queue-item {
  position: relative;
  padding: 12px 16px;
  background-color: white;
  border-radius: 8px;
  border: 1px solid #e5e7eb;
  overflow: hidden;
  transition: all 0.2s ease;
}

.queue-item:hover {
  box-shadow: 0 2px 4px rgba(0,0,0,0.05);
  border-color: #d1d5db;
}

.queue-item::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.1);
  opacity: 0;
  transition: opacity 0.3s ease;
  pointer-events: none;
}

.queue-item.uploading::before {
  opacity: 1;
}

.queue-item.uploading::after {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  width: var(--progress-width, 0%);
  height: 100%;
  background: linear-gradient(90deg, rgba(59, 130, 246, 0.2) 0%, rgba(59, 130, 246, 0.1) 100%);
  transition: width 0.3s ease;
  pointer-events: none;
  z-index: 1;
}

.item-content {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  position: relative;
  z-index: 2;
}

.item-preview {
  width: 48px;
  height: 48px;
  border-radius: 6px;
  overflow: hidden;
  flex-shrink: 0;
  box-shadow: 0 1px 2px rgba(0,0,0,0.1);
}

.preview-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.preview-placeholder {
  width: 100%;
  height: 100%;
  background-color: #f3f4f6;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #9ca3af;
}

.item-details {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
}


.item-row {
  display: flex;
  align-items: center;
  gap: 6px;
  position: relative;
  padding-right: 120px; /* 为悬浮按钮留出空间 */
}

.status-text {
  font-size: 12px;
  font-weight: 500;
  flex-shrink: 0;
}

.status-success {
  color: #059669;
}

.status-error {
  color: #dc2626;
}

.status-uploading {
  color: #2563eb;
}

.status-waiting {
  color: #6b7280;
}

.item-size {
  font-size: 12px;
  color: #6b7280;
  flex-shrink: 0;
}

.item-preview {
  width: 52px;
  height: 52px;
  border-radius: 6px;
  overflow: hidden;
  flex-shrink: 0;
  box-shadow: 0 1px 2px rgba(0,0,0,0.1);
}

.preview-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.preview-placeholder {
  width: 100%;
  height: 100%;
  background-color: #f3f4f6;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #9ca3af;
}

.item-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.item-name {
  font-size: 14px;
  font-weight: 500;
  color: #374151;
  word-break: break-word;
  line-height: 1.4;
  max-height: 2.8em; /* 限制为2行 */
  overflow: hidden;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.item-meta {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 12px;
  color: #6b7280;
}

.item-status {
  flex-shrink: 0;
}

.item-size {
  flex-shrink: 0;
}



.status-text {
  font-size: 12px;
  font-weight: 500;
}

.status-success {
  color: #059669;
}

.status-error {
  color: #dc2626;
}

.status-uploading {
  color: #2563eb;
}

.status-waiting {
  color: #6b7280;
}



.item-actions {
  display: flex;
  gap: 4px;
  position: absolute;
  right: 0;
  top: 0;
}

.action-btn {
  border-radius: 6px;
  border: none;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
  font-size: 11px;
  font-weight: 500;
  padding: 6px 10px;
  gap: 4px;
  white-space: nowrap;
  flex-shrink: 0;
  box-shadow: 0 1px 2px rgba(0,0,0,0.05);
}

.btn-text {
  display: inline;
}

.upload-btn {
  background-color: #10b981;
  color: white;
  min-width: 52px;
}

.upload-btn:hover {
  background-color: #059669;
  transform: translateY(-1px);
  box-shadow: 0 2px 4px rgba(16, 185, 129, 0.3);
}

.cancel-btn {
  background-color: #ef4444;
  color: white;
  min-width: 52px;
}

.cancel-btn:hover {
  background-color: #dc2626;
  transform: translateY(-1px);
  box-shadow: 0 2px 4px rgba(239, 68, 68, 0.3);
}

.remove-btn {
  background-color: #6b7280;
  color: white;
  width: 32px;
  height: 32px;
  padding: 0;
  border-radius: 6px;
  opacity: 0.8;
}

.remove-btn:hover {
  background-color: #4b5563;
  opacity: 1;
  transform: translateY(-1px);
}

.action-btn {
  border-radius: 6px;
  border: none;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
  font-size: 12px;
  font-weight: 500;
  padding: 6px 12px;
  gap: 4px;
  white-space: nowrap;
  flex-shrink: 0;
  box-shadow: 0 1px 2px rgba(0,0,0,0.05);
}

.btn-text {
  display: inline;
}

.upload-btn {
  background-color: #3b82f6;
  color: white;
  min-width: 56px;
}

.upload-btn:hover {
  background-color: #2563eb;
  transform: translateY(-1px);
  box-shadow: 0 4px 6px rgba(59, 130, 246, 0.2);
}

.cancel-btn {
  background-color: #f3f4f6;
  color: #6b7280;
  min-width: 56px;
  border: 1px solid #e5e7eb;
}

.cancel-btn:hover {
  background-color: #e5e7eb;
  color: #374151;
  border-color: #d1d5db;
  transform: translateY(-1px);
}

.remove-btn {
  background-color: #f3f4f6;
  color: #6b7280;
  width: 32px;
  height: 32px;
  padding: 0;
  border-radius: 6px;
  border: 1px solid #e5e7eb;
  box-shadow: 0 1px 2px rgba(0,0,0,0.05);
}

.upload-btn {
  background-color: #3b82f6;
  color: white;
  min-width: 50px;
}

.upload-btn:hover {
  background-color: #2563eb;
  transform: translateY(-1px);
  box-shadow: 0 2px 4px rgba(59, 130, 246, 0.2);
}

.cancel-btn {
  background-color: #f3f4f6;
  color: #6b7280;
  min-width: 50px;
}

.cancel-btn:hover {
  background-color: #e5e7eb;
  color: #374151;
  transform: translateY(-1px);
}

.remove-btn {
  background-color: #f3f4f6;
  color: #6b7280;
  width: 32px;
  height: 32px;
  padding: 0;
  border-radius: 4px;
}

.remove-btn:hover {
  background-color: #ef4444;
  color: white;
  transform: translateY(-1px);
  box-shadow: 0 2px 4px rgba(239, 68, 68, 0.2);
}

/* 链接生成样式 */
.link-generator {
  margin-top: 16px;
  background-color: #f8f9fa;
  border-radius: 8px;
  padding: 16px;
}



.images-links-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.image-link-item {
  padding: 12px;
  background-color: #f9fafb;
  border-radius: 6px;
  border: 1px solid #e5e7eb;
}



.tab-container {
  background-color: white;
  border-radius: 6px;
  overflow: hidden;
}

.tab-header {
  display: flex;
  border-bottom: 1px solid #e5e7eb;
}

.tab-btn {
  flex: 1;
  padding: 10px 16px;
  background: none;
  border: none;
  cursor: pointer;
  font-size: 14px;
  color: #6b7280;
  transition: all 0.2s;
  border-bottom: 2px solid transparent;
}

.tab-btn:hover {
  color: #374151;
  background-color: #f9fafb;
}

.tab-active {
  color: #8b5cf6;
  border-bottom-color: #8b5cf6;
}

.tab-content {
  padding: 12px;
}

.link-display {
  display: flex;
  gap: 8px;
}

.link-input {
  flex: 1;
  padding: 8px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 14px;
  background-color: #f9fafb;
  color: #374151;
}

.copy-btn {
  padding: 8px 12px;
  background-color: #3b82f6;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  transition: background-color 0.2s;
}

.copy-btn:hover {
  background-color: #2563eb;
}

.upload-notice {
  margin-top: 24px;
  padding: 16px;
  background-color: #f8f9fa;
  border-radius: 8px;
  border-left: 4px solid #f59e0b;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .upload-page {
    padding: 0 12px;
  }

  .upload-card {
    padding: 16px;
    border-radius: 8px;
  }

  /* 卡片头部移动端优化 */
  .card-header {
    flex-direction: column;
    gap: 16px;
    margin-bottom: 20px;
  }

  .header-left {
    text-align: center;
  }

  .header-right {
    width: 100%;
  }

  .strategy-selector {
    min-width: auto;
    width: 100%;
  }

  .strategy-select {
    width: 100%;
  }

  /* 上传区域移动端优化 */
  .upload-area {
    padding: 40px 16px;
    border-radius: 6px;
  }

  .upload-icon i {
    font-size: 3rem; /* text-4xl 替代 text-5xl */
  }

  .upload-area h3 {
    font-size: 16px;
    margin-bottom: 8px;
  }

  .upload-area p {
    font-size: 12px;
  }

  /* URL上传区域移动端优化 */
  .url-upload-section {
    margin-top: 20px;
    padding: 16px;
  }

  .url-upload-form {
    flex-direction: column;
    gap: 12px;
  }

  .url-input {
    width: 100%;
  }

  .url-upload-btn {
    width: 100%;
    justify-content: center;
  }

  /* 上传队列移动端优化 */
  .upload-queue {
    padding: 16px;
    margin-top: 16px;
    border-radius: 8px;
  }

  .link-generator {
    padding: 16px;
    margin-top: 16px;
  }

  .queue-item {
    padding: 12px;
    gap: 10px;
  }

  .item-content {
    gap: 10px;
  }

  .item-preview {
    width: 44px;
    height: 44px;
  }

  .item-details {
    gap: 6px;
  }

  .item-name {
    display: -webkit-box;
    -webkit-box-orient: vertical;
    -webkit-line-clamp: 1;
    overflow: hidden;

  }

  .item-row {
    gap: 4px;
    position: relative;
    padding-right: 100px; /* 移动端为悬浮按钮留出空间 */
  }

  .status-text {
    font-size: 11px;
  }

  .item-size {
    font-size: 11px;
  }

  .item-actions {
    gap: 3px;
    position: absolute;
    right: 0;
    top: 0;
  }

  .item-row {
    gap: 6px;
  }

  .status-text {
    font-size: 11px;
  }

  .item-size {
    font-size: 11px;
  }



  .action-btn {
    padding: 6px 8px;
    font-size: 11px;
    min-width: 44px;
  }

  .btn-text {
    display: inline;
  }

  .remove-btn {
    width: 32px;
    height: 32px;
  }

  .item-content {
    display: flex;
    align-items: flex-start;
    gap: 8px;
    width: 100%;
  }

  .item-preview {
    width: 40px;
    height: 40px;
    flex-shrink: 0;
  }

  .item-info {
    flex: 1;
    min-width: 0;
    max-width: calc(100% - 48px); /* 预览图宽度40px + gap 8px */
    display: flex;
    flex-direction: column;
    gap: 3px;
  }



  .item-meta {
    display: flex;
    flex-direction: column;
    gap: 2px;
    font-size: 11px;
    color: #6b7280;
  }

  .item-status,
  .item-size {
    font-size: 10px;
  }

  .status-text {
    font-size: 10px;
    font-weight: 600;
    flex-shrink: 0;
  }

  /* 优化移动端状态显示 */
  .item-size {
    flex-direction: row;
    align-items: center;
    gap: 4px;
    flex-wrap: wrap;
  }

  .item-actions {
    display: flex;
    justify-content: flex-end;
    gap: 6px;
  }

  .action-btn {
    padding: 5px 10px;
    font-size: 11px;
    border-radius: 5px;
    white-space: nowrap;
    min-width: auto;
    box-shadow: 0 1px 2px rgba(0,0,0,0.05);
  }

  .upload-btn,
  .cancel-btn {
    flex: 1;
    max-width: 52px;
    border: 1px solid transparent;
  }

  .upload-btn {
    border-color: #3b82f6;
  }

  .cancel-btn {
    border-color: #e5e7eb;
  }

  .upload-btn,
  .cancel-btn {
    flex: 1;
    max-width: 48px;
  }

  .remove-btn {
    width: 32px;
    padding: 6px;
  }

  .action-btn i {
    margin-right: 3px;
  }

  /* 在移动端保留按钮文本 */
  .btn-text {
    display: inline;
  }

  .upload-btn {
    flex: 1;
    max-width: 70px;
    background-color: #10b981;
    color: white;
  }

  .upload-btn:hover {
    background-color: #059669;
  }

  .cancel-btn {
    flex: 1;
    max-width: 70px;
    background-color: #f87171;
    color: white;
  }

  .cancel-btn:hover {
    background-color: #dc2626;
  }

  .remove-btn {
    width: 32px;
    padding: 6px;
    background-color: #9ca3af;
    color: white;
  }

  .remove-btn:hover {
    background-color: #6b7280;
  }

  /* 标签页移动端优化 */
  .tab-header {
    flex-wrap: wrap;
    gap: 4px;
  }

  .tab-btn {
    flex: 1;
    min-width: 70px;
    padding: 8px 6px;
    font-size: 12px;
  }

  /* 链接显示移动端优化 */
  .link-display {
    flex-direction: column;
    gap: 8px;
  }

  .link-input {
    font-size: 12px;
  }

  .copy-btn {
    width: 100%;
    justify-content: center;
  }

  /* 上传须知移动端优化 */
  .upload-notice {
    margin-top: 16px;
    padding: 12px;
  }

  .upload-notice ul {
    gap: 8px;
  }

  .upload-notice li {
    font-size: 12px;
  }

  /* 移动端额外优化 */
  .header-left h2 {
    font-size: 18px;
    margin-bottom: 4px;
  }

  .header-left p {
    font-size: 13px;
  }

  .strategy-label {
    font-size: 13px;
  }

  .url-upload-header h3 {
    font-size: 14px;
  }

  .url-upload-btn {
    font-size: 13px;
    padding: 8px 16px;
  }

  /* 确保在移动端不出现水平滚动 */
  .upload-page,
  .upload-card {
    overflow-x: hidden;
  }
}

/* 动画效果 */
@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.upload-card {
  animation: fadeIn 0.5s ease-out;
}
</style> 
