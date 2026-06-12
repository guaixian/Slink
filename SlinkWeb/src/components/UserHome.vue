<template>
  <div class="font-inter bg-neutral-100 text-neutral-800 min-h-screen flex relative">
    <!-- 移动端遮罩层 -->
    <div
      v-if="isMobileMenuOpen"
      class="mobile-overlay"
      @click="closeMobileMenu"
    ></div>

    <!-- 左侧侧边栏 -->
    <aside class="sidebar" :class="{ 'sidebar-mobile-open': isMobileMenuOpen }">
        <div class="sidebar-header">
        <div class="flex items-center space-x-2">
          <i class="fa fa-cloud text-primary text-xl"></i>
          <span class="font-bold text-lg sidebar-title">图床</span>
        </div>
      </div>
      
      <ul class="sidebar-nav">
        <li class="nav-item" :class="{ active: $route.name === 'admin-upload' }">
          <router-link to="/admin" class="flex items-center">
            <i class="fa fa-cloud-upload mr-3"></i>
            <span>上传图片</span>
          </router-link>
        </li>
        <li class="nav-item" :class="{ active: $route.name === 'admin-watermark' }">
          <router-link to="/admin/watermark" class="flex items-center">
            <i class="fa fa-tint mr-3"></i>
            <span>图片水印</span>
          </router-link>
        </li>
        <li class="nav-item" :class="{ active: $route.name === 'admin-image-tools' }">
          <router-link to="/admin/image-tools" class="flex items-center">
            <i class="fa fa-magic mr-3"></i>
            <span>图片处理</span>
          </router-link>
        </li>
        <li class="nav-item" :class="{ active: $route.name === 'admin-ai-tools' }">
          <router-link to="/admin/ai-tools" class="flex items-center">
            <i class="fa fa-flask mr-3"></i>
            <span>AI 处理</span>
          </router-link>
        </li>
        <li class="nav-item" :class="{ active: $route.name === 'admin-dashboard' }">
          <router-link to="/admin/dashboard" class="flex items-center">
            <i class="fa fa-tachometer mr-3"></i>
            <span>仪表盘</span>
          </router-link>
        </li>
        <li class="nav-item" :class="{ active: $route.name === 'admin-images' }">
          <router-link to="/admin/images" class="flex items-center">
            <i class="fa fa-image mr-3"></i>
            <span>我的图片</span>
          </router-link>
        </li>
        <li class="nav-item" :class="{ active: $route.name === 'admin-basic-settings' }">
          <router-link to="/admin/basic-settings" class="flex items-center">
            <i class="fa fa-cog mr-3"></i>
            <span>偏好设置</span>
          </router-link>
        </li>
        <li class="nav-item" :class="{ active: $route.name === 'admin-api-reference' }">
          <router-link to="/admin/api-reference" class="flex items-center">
            <i class="fa fa-book mr-3"></i>
            <span>开放 API</span>
          </router-link>
        </li>

        <li class="nav-divider">
          <span class="divider-text">系统</span>
        </li>

        <li class="nav-item" :class="{ active: $route.name === 'admin-image-management' }">
          <router-link to="/admin/image-management" class="flex items-center">
            <i class="fa fa-cogs mr-3"></i>
            <span>图片管理</span>
          </router-link>
        </li>
        <li class="nav-item" :class="{ active: $route.name === 'admin-console' }">
          <router-link to="/admin/console" class="flex items-center">
            <i class="fa fa-terminal mr-3"></i>
            <span>控制台</span>
          </router-link>
        </li>
        <li class="nav-item" :class="{ active: $route.name === 'admin-upload-policy' }">
          <router-link to="/admin/upload-policy" class="flex items-center">
            <i class="fa fa-upload mr-3"></i>
            <span>上传策略</span>
          </router-link>
        </li>
        <li class="nav-item" :class="{ active: $route.name === 'admin-settings' }">
          <router-link to="/admin/settings" class="flex items-center">
            <i class="fa fa-cogs mr-3"></i>
            <span>系统设置</span>
          </router-link>
        </li>
        <li class="nav-item" :class="{ active: $route.name === 'admin-storage' || $route.name === 'admin-storage-create' || $route.name === 'admin-storage-edit' }">
          <router-link to="/admin/storage" class="flex items-center">
            <i class="fa fa-database mr-3"></i>
            <span>储存策略</span>
          </router-link>
        </li>
      </ul>
      
      <div class="sidebar-footer">
        <div class="text-sm font-medium mb-2">容量使用</div>
        <div class="progress">
          <div class="progress-bar" :style="{ width: storageUsage + '%' }"></div>
        </div>
        <div class="text-xs text-neutral-500 mt-1">
          {{ usedStorage }} / {{ totalStorage }}
        </div>
      </div>
    </aside>

    <!-- 主要内容区 -->
    <div class="main-content" :class="{ 'main-content-mobile': isMobile }">
      <!-- 顶部导航栏 -->
      <header class="header">
        <!-- 汉堡菜单按钮 - 只在移动端显示 -->
        <button
          v-if="isMobile"
          @click="toggleMobileMenu"
          class="mobile-menu-btn"
        >
          <i class="fa fa-bars"></i>
        </button>

        <div class="header-title">
          <i class="fa fa-cloud-upload mr-2"></i>
          {{ pageTitle }}
        </div>
        <div class="header-user">
          <span v-if="!isMobile" class="user-info">
            <i class="fa fa-server mr-1"></i>
            WebDav分区
          </span>
          <span class="user-role">
            <i class="fa fa-user-shield mr-1"></i>
            个人图床
          </span>
          <button @click="handleLogout" class="logout-btn">
            <i class="fa fa-sign-out mr-1"></i>
            <span class="logout-text">退出</span>
          </button>
        </div>
      </header>

      <!-- 主要内容 -->
      <main class="main">
        <router-view />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { userAPI } from '../api'
import { useUserStore } from '../stores/user'

const route = useRoute()
const userStore = useUserStore()

// 移动端菜单状态
const isMobileMenuOpen = ref(false)
const isMobile = ref(false)

const usedStorage = ref('—')
const totalStorage = ref('—')
const storageUsage = ref(0)

function formatUsedMB(mb: number): string {
  if (mb < 1) return `${(mb * 1024).toFixed(2)} KB`
  if (mb < 1024) return `${mb.toFixed(2)} MB`
  return `${(mb / 1024).toFixed(2)} GB`
}

async function loadSidebarUsage() {
  const quotaGb = userStore.defaultStorageGB || 5
  totalStorage.value = `${quotaGb} GB（配额）`
  try {
    if (!userStore.isSystemConfigLoaded) {
      await userStore.getSystemInfo()
    }
    const res = await userAPI.getDashboardData()
    if (res.data?.status === 'success' && res.data.data?.dashboard) {
      const mb = Number(res.data.data.dashboard.used_size_mb) || 0
      usedStorage.value = formatUsedMB(mb)
      const capMb = quotaGb * 1024
      storageUsage.value = capMb > 0 ? Math.min(100, Math.round((mb / capMb) * 100)) : 0
    }
  } catch {
    usedStorage.value = '—'
  }
}

// 窗口大小监听
const updateMobileStatus = () => {
  isMobile.value = window.innerWidth < 768
}

// 计算页面标题
const pageTitle = computed(() => {
  switch (route.name) {
    case 'admin-dashboard':
      return '仪表盘'
    case 'admin-images':
      return '我的图片'
    case 'admin-basic-settings':
      return '偏好设置'
    case 'admin-api-reference':
      return '开放 API 说明'
    case 'admin-upload':
      return '上传图片'
    case 'admin-watermark':
      return '图片水印'
    case 'admin-image-tools':
      return '图片处理'
    case 'admin-ai-tools':
      return 'AI 处理'
    case 'admin-image-management':
      return '图片管理'
    case 'admin-console':
      return '系统控制台'
    case 'admin-upload-policy':
      return '上传策略'
    case 'admin-settings':
      return '系统设置'
    case 'admin-storage':
    case 'admin-storage-create':
    case 'admin-storage-edit':
      return '储存策略'
    default:
      return '上传图片'
  }
})

// 退出登录处理
const handleLogout = () => {
  // 清除用户信息
  localStorage.removeItem('userToken')
  localStorage.removeItem('userInfo')
  localStorage.removeItem('token')

  // 跳转到首页
  window.location.href = '/'
}

// 移动端菜单方法
const toggleMobileMenu = () => {
  isMobileMenuOpen.value = !isMobileMenuOpen.value
}

const closeMobileMenu = () => {
  isMobileMenuOpen.value = false
}

// 生命周期
onMounted(() => {
  // 确保Font Awesome已加载
  if (!document.querySelector('link[href*="font-awesome"]')) {
    const link = document.createElement('link')
    link.href = 'https://cdn.jsdelivr.net/npm/font-awesome@4.7.0/css/font-awesome.min.css'
    link.rel = 'stylesheet'
    document.head.appendChild(link)
  }

  // 初始化移动端状态
  updateMobileStatus()

  // 监听窗口大小变化
  window.addEventListener('resize', updateMobileStatus)

  loadSidebarUsage()
})

onUnmounted(() => {
  window.removeEventListener('resize', updateMobileStatus)
})
</script>

<style scoped>
/* 移动端遮罩层 */
.mobile-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  z-index: 25;
}

/* 侧边栏样式 */
.sidebar {
  width: 240px;
  background-color: #fff;
  box-shadow: 0 0 8px rgba(0,0,0,0.05);
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  height: 100vh;
  position: fixed;
  left: 0;
  top: 0;
  z-index: 30;
  transform: translateX(-100%);
  transition: transform 0.3s ease;
}

.sidebar-mobile-open {
  transform: translateX(0);
}

.sidebar-header {
  padding: 20px;
  font-size: 18px;
  font-weight: bold;
  color: #333;
  border-bottom: 1px solid #eee;
}

.sidebar-nav {
  list-style: none;
  flex: 1;
  overflow-y: auto;
}

.nav-item {
  color: #555;
  transition: all 0.3s ease;
  font-size: 14px;
  border-left: 3px solid transparent;
}

.nav-item:hover {
  background-color: #f8f9fa;
  color: #3b82f6;
}

.nav-item.active {
  background-color: #eff6ff;
  color: #3b82f6;
  font-weight: 500;
  border-left-color: #3b82f6;
}

.nav-item i {
  width: 16px;
  text-align: center;
}

.nav-item a {
  color: inherit;
  text-decoration: none;
  display: flex;
  align-items: center;
  width: 100%;
  padding: 12px 20px;
  cursor: pointer;
}

/* 分隔线样式 */
.nav-divider {
  padding: 16px 20px 8px 20px;
  border-bottom: 1px solid #e5e7eb;
  margin-bottom: 8px;
}

.divider-text {
  font-size: 12px;
  font-weight: 600;
  color: #9ca3af;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.sidebar-footer {
  padding: 15px 20px;
  border-top: 1px solid #eee;
  background-color: #f8f9fa;
}

.progress {
  width: 100%;
  height: 6px;
  background-color: #e5e7eb;
  border-radius: 3px;
  overflow: hidden;
  margin: 8px 0;
}

.progress-bar {
  height: 100%;
  background-color: #3b82f6;
  border-radius: 3px;
  transition: width 0.3s ease;
}

/* 主内容区样式 */
.main-content {
  flex: 1;
  margin-left: 240px;
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  transition: margin-left 0.3s ease;
}

.main-content-mobile {
  margin-left: 0;
}

/* 顶部导航栏 */
.header {
  background-color: #fff;
  border-bottom: 1px solid #e5e7eb;
  padding: 0 24px;
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  box-shadow: 0 1px 3px rgba(0,0,0,0.05);
}

/* 汉堡菜单按钮 */
.mobile-menu-btn {
  background: none;
  border: none;
  font-size: 20px;
  color: #6b7280;
  cursor: pointer;
  padding: 8px;
  border-radius: 6px;
  transition: all 0.2s ease;
  margin-right: 12px;
}

.mobile-menu-btn:hover {
  background-color: #f3f4f6;
  color: #374151;
}

.header-title {
  font-size: 18px;
  font-weight: 600;
  color: #374151;
  display: flex;
  align-items: center;
}

.header-user {
  display: flex;
  align-items: center;
  gap: 16px;
}

.user-info, .user-role {
  font-size: 14px;
  color: #6b7280;
  display: flex;
  align-items: center;
}

.logout-btn {
  background-color: #ef4444;
  color: white;
  border: none;
  padding: 8px 16px;
  border-radius: 6px;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
}

.logout-text {
  display: inline;
}

.logout-btn:hover {
  background-color: #dc2626;
  transform: translateY(-1px);
}

.logout-btn:active {
  transform: translateY(0);
}

/* 主要内容 */
.main {
  flex: 1;
  padding: 32px;
  background-color: #f9fafb;
  min-height: calc(100vh - 64px);
  max-width: 1400px;
  margin: 0 auto;
  width: 100%;
}

/* 桌面端样式 */
@media (min-width: 769px) {
  .sidebar {
    transform: translateX(0); /* 桌面端始终显示 */
  }

  .mobile-menu-btn {
    display: none; /* 桌面端隐藏汉堡菜单 */
  }
}

/* 响应式设计 */
@media (max-width: 768px) {
  .sidebar {
    transform: translateX(-100%);
    width: 280px; /* 移动端抽屉宽度 */
  }

  .sidebar-mobile-open {
    transform: translateX(0);
  }

  .sidebar-title {
    display: block !important; /* 确保标题在移动端显示 */
  }

  .main-content {
    margin-left: 0;
  }

  .header {
    padding: 0 16px;
  }

  .header-title {
    font-size: 16px;
  }

  .user-info {
    display: none; /* 在移动端隐藏分区信息 */
  }

  .user-role {
    font-size: 12px;
  }

  .logout-text {
    display: none; /* 在移动端隐藏退出按钮文本 */
  }

  .logout-btn {
    padding: 8px;
    min-width: 36px;
  }

  .main {
    padding: 16px 12px;
  }
}

/* 滚动条样式 */
.sidebar-nav::-webkit-scrollbar {
  width: 4px;
}

.sidebar-nav::-webkit-scrollbar-track {
  background: transparent;
}

.sidebar-nav::-webkit-scrollbar-thumb {
  background: #d1d5db;
  border-radius: 2px;
}

.sidebar-nav::-webkit-scrollbar-thumb:hover {
  background: #9ca3af;
}
</style>