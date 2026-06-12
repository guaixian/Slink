<template>
  <div class="font-inter bg-neutral-100 text-neutral-800 min-h-screen flex flex-col relative">
    <!-- 顶部导航栏 -->
    <header class="bg-white shadow-sm sticky top-0 z-40">
      <div class="container mx-auto px-4">
        <nav class="flex items-center justify-between h-12">
          <!-- Logo和品牌 -->
          <div class="flex items-center space-x-2">
            <img :src="logoImage" alt="Slink Logo" class="h-6 w-auto">
            <span class="font-bold text-base">个人图床</span>
          </div>
          
          <!-- 用户区域 -->
          <div class="flex items-center space-x-3">
            <template v-if="!isLoggedIn">
              <button 
                @click="openModal" 
                class="px-3 py-1.5 bg-primary text-white rounded-lg hover:bg-primary/90 transition-colors flex items-center text-sm"
              >
                <i class="fa fa-sign-in mr-1.5"></i>
                <span>登录</span>
              </button>
            </template>
            <template v-else>
              <div class="flex items-center space-x-2">
                <span class="text-sm text-neutral-600">
                  <i class="fa fa-user mr-1"></i>
                  {{ userName }}
                </span>
                <button 
                  @click="handleLogout" 
                  class="px-3 py-1.5 bg-neutral-200 text-neutral-700 rounded-lg hover:bg-neutral-300 transition-colors flex items-center text-sm"
                >
                  <i class="fa fa-sign-out mr-1.5"></i>
                  <span>退出</span>
                </button>
              </div>
            </template>
            <button class="md:hidden text-neutral-600">
              <i class="fa fa-bars"></i>
            </button>
          </div>
        </nav>
      </div>
    </header>

    <!-- 主要内容区 -->
    <main class="flex-grow container mx-auto px-4 py-6 pb-20">
      <div class="max-w-3xl mx-auto">
        <!-- 欢迎信息 -->
        <div class="text-center mb-6">
          <h1 class="text-[clamp(1.5rem,3vw,2rem)] font-bold text-neutral-800 mb-2">
            自托管个人图床
          </h1>
          <p class="text-neutral-600 max-w-xl mx-auto text-sm">
            私有部署、单账号管理。登录后可上传与管理图片，支持拖拽与多种格式（游客上传取决于系统设置）。
          </p>
        </div>
        
        <!-- 游客提示 -->
        <div v-if="!isLoggedIn" class="bg-primary/10 border border-primary/20 rounded-lg p-3 mb-6 text-center">
          <p class="text-neutral-700 text-sm">
            <i class="fa fa-info-circle mr-2 text-primary"></i>
            您当前以游客身份使用
            <span v-if="isSystemConfigLoaded && !isGuestUploadEnabled">
              （游客上传已关闭，请登录站长账号）
            </span>
            <span v-else>
              ，登录后可使用完整图床功能
              <button 
                @click="openModal" 
                class="text-primary font-medium hover:underline"
              >
                登录
              </button>
            </span>
          </p>
        </div>
        
        <!-- 上传区域 -->
        <div class="bg-white rounded-xl shadow-sm p-4 md:p-6">
          <div 
            ref="uploadArea"
            class="border-2 border-dashed border-neutral-300 rounded-lg p-6 md:p-8 text-center transition-all duration-300 hover:border-primary"
            :class="{ 'upload-area-active': isDragOver }"
            @dragenter.prevent="handleDragEnter"
            @dragover.prevent="handleDragOver"
            @dragleave.prevent="handleDragLeave"
            @drop.prevent="handleDrop"
          >
            <div class="mb-3 md:mb-4 text-4xl md:text-5xl text-neutral-300">
              <i class="fa fa-cloud-upload"></i>
            </div>
            <h3 class="text-lg md:text-xl font-medium mb-2">拖拽文件到此处上传</h3>
            <p class="text-neutral-500 mb-4 md:mb-6">或者</p>
            <button 
              @click="triggerFileInput"
              class="px-5 md:px-6 py-2 bg-primary text-white rounded-lg shadow-sm hover:bg-primary/90 transition-colors inline-flex items-center text-sm md:text-base"
            >
              <i class="fa fa-folder-open mr-2"></i> 选择文件
            </button>
            <input 
              ref="fileInput"
              type="file" 
              multiple 
              class="hidden" 
              @change="handleFileSelect"
              accept="image/*"
            />
            <p class="text-neutral-500 mt-4 text-xs">
              <i class="fa fa-info-circle mr-1"></i>
              最大可上传 14.77 MB 的图片；登录站长账号后按系统策略上传与管理
              <span v-if="isSystemConfigLoaded && defaultStorageGB > 0">
                ，默认配额约 {{ defaultStorageGB }} GB（可在后台调整）
              </span>
            </p>
          </div>
          
          <!-- 上传须知 -->
          <div class="mt-4 md:mt-6 bg-neutral-50 rounded-lg p-3 md:p-4">
            <h4 class="font-medium text-sm md:text-base mb-2 flex items-center">
              <i class="fa fa-exclamation-triangle mr-2 text-orange-500"></i>
              上传须知
            </h4>
            <ul class="space-y-1 text-neutral-600 text-xs">
              <li class="flex items-start">
                <i class="fa fa-check-circle text-green-500 mt-0.5 mr-2"></i>
                <span>支持 JPG、PNG、GIF、WebP、SVG 等常见图片格式</span>
              </li>
              <li class="flex items-start">
                <i class="fa fa-clock-o text-orange-500 mt-0.5 mr-2"></i>
                <span>若开启游客上传，请自行在后台制定保留策略；站长资源建议登录后管理</span>
              </li>
              <li class="flex items-start">
                <i class="fa fa-ban text-red-500 mt-0.5 mr-2"></i>
                <span>请勿上传违反法律法规的内容，我们保留追究责任的权利</span>
              </li>
            </ul>
          </div>
        </div>
        
        <!-- 功能介绍 -->
        <div class="mt-8 grid grid-cols-1 md:grid-cols-3 gap-4">
          <div class="bg-white rounded-xl shadow-sm p-4 text-center card-hover">
            <div class="w-12 h-12 bg-primary/10 rounded-full flex items-center justify-center mx-auto mb-3">
              <i class="fa fa-rocket text-primary text-lg"></i>
            </div>
            <h3 class="font-medium text-base mb-2">高速上传</h3>
            <p class="text-neutral-600 text-xs">
              采用优化的上传协议，支持断点续传，确保大文件快速稳定上传
            </p>
          </div>
          
          <div class="bg-white rounded-xl shadow-sm p-4 text-center card-hover">
            <div class="w-12 h-12 bg-primary/10 rounded-full flex items-center justify-center mx-auto mb-3">
              <i class="fa fa-lock text-primary text-lg"></i>
            </div>
            <h3 class="font-medium text-base mb-2">安全可靠</h3>
            <p class="text-neutral-600 text-xs">
              所有图片加密存储，多重备份，确保您的图片安全不丢失
            </p>
          </div>
          
          <div class="bg-white rounded-xl shadow-sm p-4 text-center card-hover">
            <div class="w-12 h-12 bg-primary/10 rounded-full flex items-center justify-center mx-auto mb-3">
              <i class="fa fa-share-alt text-primary text-lg"></i>
            </div>
            <h3 class="font-medium text-base mb-2">便捷分享</h3>
            <p class="text-neutral-600 text-xs">
              一键复制多种格式链接，支持Markdown、HTML等，方便在各平台分享
            </p>
          </div>
        </div>

        <!-- 统计信息 -->
        <div class="mt-8 bg-white rounded-xl shadow-sm p-4">
          <h3 class="text-base font-medium mb-4 text-center flex items-center justify-center">
            <i class="fa fa-bar-chart mr-2 text-primary"></i>
            使用提示
          </h3>
          <div class="grid grid-cols-1 md:grid-cols-3 gap-3 text-sm text-neutral-600 text-center">
            <div>数据保存在您自己的服务器或对象存储</div>
            <div>仅一个站长账号，无多租户</div>
            <div>存储路径、外链格式可在后台配置</div>
          </div>
        </div>
      </div>
    </main>

    <!-- 页脚 -->
    <footer class="bg-white border-t border-neutral-200 py-4 fixed bottom-0 left-0 right-0 z-30">
      <div class="container mx-auto px-4">
        <div class="flex flex-col md:flex-row justify-between items-center">
          <div class="mb-3 md:mb-0">
            <p class="text-neutral-500 text-xs flex items-center">
              <i class="fa fa-copyright mr-1"></i>
              Copyright © 2018 - present Slink. All rights reserved.
            </p>
            <p class="text-neutral-500 text-xs mt-1 flex items-center">
              <i class="fa fa-exclamation-triangle mr-1"></i>
              请勿上传违反中国大陆和香港法律的图片，违者后果自负。
            </p>
          </div>
          
          <div class="flex space-x-4">
            <a href="#" class="text-neutral-500 hover:text-primary transition-colors" title="GitHub">
              <i class="fa fa-github text-lg"></i>
            </a>
            <a href="#" class="text-neutral-500 hover:text-primary transition-colors" title="Twitter">
              <i class="fa fa-twitter text-lg"></i>
            </a>
            <a href="#" class="text-neutral-500 hover:text-primary transition-colors" title="微博">
              <i class="fa fa-weibo text-lg"></i>
            </a>
            <a href="#" class="text-neutral-500 hover:text-primary transition-colors" title="微信">
              <i class="fa fa-wechat text-lg"></i>
            </a>
          </div>
        </div>
      </div>
    </footer>

    <!-- 登录/注册模态框 -->
    <div 
      v-show="showModal" 
      class="modal-backdrop"
      @click="closeModal"
    >
      <div 
        class="bg-white rounded-xl shadow-xl w-full max-w-md p-6 relative transform transition-all"
        @click.stop
      >
        <button 
          @click="closeModal"
          class="absolute top-4 right-4 text-neutral-400 hover:text-neutral-600"
        >
          <i class="fa fa-times text-xl"></i>
        </button>
        
        <div class="mb-4 text-center text-neutral-700 font-medium">
          <i class="fa fa-sign-in mr-2 text-primary"></i>站长登录
        </div>

        <!-- 登录表单 -->
        <div>
          <!-- 错误提示 -->
          <div v-if="loginError" class="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg">
            <p class="text-red-600 text-sm flex items-center">
              <i class="fa fa-exclamation-circle mr-2"></i>
              {{ loginError }}
            </p>
          </div>
          
          <div class="mb-4">
            <label class="block text-neutral-700 mb-2 flex items-center">
              <i class="fa fa-user mr-2"></i>登录账号
            </label>
            <input 
              v-model="loginForm.username"
              type="text" 
              class="form-input" 
              placeholder="首次初始化时设置的账号"
              autocomplete="username"
              :disabled="isSubmitting"
              @keyup.enter="handleLogin"
            >
          </div>
          
          <div class="mb-6">
            <label class="block text-neutral-700 mb-2 flex items-center">
              <i class="fa fa-lock mr-2"></i>密码
            </label>
            <input 
              v-model="loginForm.password"
              type="password" 
              class="form-input" 
              placeholder="请输入您的密码"
              :disabled="isSubmitting"
              @keyup.enter="handleLogin"
            >
          </div>
          
          <div class="flex items-center justify-between mb-6">
            <label class="flex items-center">
              <input 
                v-model="loginForm.remember"
                type="checkbox" 
                class="mr-2"
                :disabled="isSubmitting"
              >
              <span class="text-sm text-neutral-600">记住我</span>
            </label>
            <a href="#" class="text-sm text-primary hover:underline flex items-center">
              <i class="fa fa-question-circle mr-1"></i>忘记密码？
            </a>
          </div>
          
          <button 
            @click="handleLogin"
            :disabled="isSubmitting"
            class="w-full py-3 bg-primary text-white rounded-lg hover:bg-primary/90 transition-colors font-medium flex items-center justify-center disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <i v-if="isSubmitting" class="fa fa-spinner fa-spin mr-2"></i>
            <i v-else class="fa fa-sign-in mr-2"></i>
            {{ isSubmitting ? '登录中...' : '登录' }}
          </button>
          
          <p class="mt-4 text-center text-xs text-neutral-500">
            账号仅在首次初始化时创建，不提供公开注册。
          </p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '../stores/user'
import logoImage from '../assets/logo.png'
import { http } from '../utils/request'

const router = useRouter()
const userStore = useUserStore()

// 响应式数据
const showModal = ref(false)
const isDragOver = ref(false)
const uploadArea = ref<HTMLElement>()
const fileInput = ref<HTMLInputElement>()

// 表单数据
const loginForm = reactive({
  username: '',
  password: '',
  remember: false
})

// 表单验证状态
const loginError = ref('')
const isSubmitting = ref(false)

// 计算属性
const isLoggedIn = computed(() => userStore.isLoggedIn)
const userName = computed(() => userStore.userName)

const isGuestUploadEnabled = computed(() => userStore.isGuestUploadEnabled)
const defaultStorageGB = computed(() => userStore.defaultStorageGB)
const isSystemConfigLoaded = computed(() => userStore.isSystemConfigLoaded)

// 模态框控制
const openModal = () => {
  if (isLoggedIn.value) {
    router.push('/admin')
    return
  }
  showModal.value = true
  document.body.style.overflow = 'hidden'
}

const closeModal = () => {
  showModal.value = false
  document.body.style.overflow = ''
  loginError.value = ''
  loginForm.username = ''
  loginForm.password = ''
  loginForm.remember = false
}

// 文件上传相关
const triggerFileInput = () => {
  if (!isLoggedIn.value) {
    openModal()
    return
  }
  fileInput.value?.click()
}

const handleFileSelect = (event: Event) => {
  const target = event.target as HTMLInputElement
  if (target.files && target.files.length > 0) {
    if (!isLoggedIn.value) {
      openModal()
      return
    }
    router.push('/admin')
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
    if (!isLoggedIn.value) {
      openModal()
      return
    }
    router.push('/admin')
  }
}

// 表单处理
const handleLogin = async () => {
  // 表单验证
  if (!loginForm.username.trim()) {
    loginError.value = '请输入登录账号'
    return
  }
  
  if (!loginForm.password.trim()) {
    loginError.value = '请输入密码'
    return
  }

  isSubmitting.value = true
  loginError.value = ''

  try {
    const response = await userStore.login(loginForm.username.trim(), loginForm.password)
    
    if (response.status) {
      closeModal()
      router.push('/admin')
    } else {
      loginError.value = response.message || '登录失败'
    }
  } catch (error: any) {
    console.error('Login error:', error)
    if (error.response?.data?.message) {
      loginError.value = error.response.data.message
    } else if (error.response?.status === 401) {
      loginError.value = '账号或密码错误'
    } else if (error.response?.status === 0 || error.code === 'NETWORK_ERROR') {
      loginError.value = '网络连接失败，请检查网络设置'
    } else {
      loginError.value = '登录失败，请稍后重试'
    }
  } finally {
    isSubmitting.value = false
  }
}

// 登出
const handleLogout = () => {
  userStore.logout()
  router.push('/')
}

// 生命周期
onMounted(async () => {
  // 添加Font Awesome CDN
  if (!document.querySelector('link[href*="font-awesome"]')) {
    const link = document.createElement('link')
    link.href = 'https://cdn.jsdelivr.net/npm/font-awesome@4.7.0/css/font-awesome.min.css'
    link.rel = 'stylesheet'
    document.head.appendChild(link)
  }
  
  // 检查初始化状态
  try {
    const response = await http.get('/api/init/status')
    if (response.data.code === 200 && !response.data.data.is_initialized) {
      // 系统未初始化，跳转到初始化页面
      router.push('/init')
      return
    }
  } catch (error) {
    console.error('检查初始化状态失败:', error)
  }
  
  // 获取系统配置
  try {
    await userStore.getSystemInfo()
    console.log('系统配置已加载:', userStore.systemConfig)
  } catch (error) {
    console.error('获取系统配置失败:', error)
  }
  
})
</script>

<style scoped>
/* 卡片悬停效果 */
.card-hover {
  transition: all 0.3s ease;
}

.card-hover:hover {
  transform: translateY(-4px);
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.1);
}

/* 上传区域激活状态 */
.upload-area-active {
  border-color: #3b82f6;
  background-color: #eff6ff;
}

/* 模态框背景 */
.modal-backdrop {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 50;
  backdrop-filter: blur(4px);
}

/* 表单输入框样式 */
.form-input {
  width: 100%;
  padding: 0.75rem;
  border: 1px solid #d1d5db;
  border-radius: 0.5rem;
  font-size: 0.875rem;
  transition: all 0.2s ease;
}

.form-input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

/* 按钮悬停效果增强 */
button:hover {
  transform: translateY(-1px);
}

/* 统计卡片动画 */
@keyframes countUp {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* 图标动画 */
.fa {
  transition: transform 0.2s ease;
}

button:hover .fa,
a:hover .fa {
  transform: scale(1.1);
}

/* 响应式优化 */
@media (max-width: 768px) {
  .modal-backdrop > div {
    margin: 1rem;
    max-height: calc(100vh - 2rem);
    overflow-y: auto;
  }
}

/* 加载动画 */
@keyframes pulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}

.loading {
  animation: pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
}

/* 成功状态 */
.success {
  color: #10b981;
}

/* 警告状态 */
.warning {
  color: #f59e0b;
}

/* 错误状态 */
.error {
  color: #ef4444;
}

/* 信息状态 */
.info {
  color: #3b82f6;
}

/* 表单输入框焦点效果 */
.form-input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
  transform: translateY(-1px);
}

/* 按钮禁用状态 */
button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
  transform: none !important;
}

/* 成功提示动画 */
@keyframes slideIn {
  from {
    opacity: 0;
    transform: translateY(-10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.bg-green-50 {
  animation: slideIn 0.3s ease-out;
}

/* 错误提示动画 */
@keyframes shake {
  0%, 100% {
    transform: translateX(0);
  }
  25% {
    transform: translateX(-5px);
  }
  75% {
    transform: translateX(5px);
  }
}

.bg-red-50 {
  animation: shake 0.5s ease-in-out;
}
</style> 