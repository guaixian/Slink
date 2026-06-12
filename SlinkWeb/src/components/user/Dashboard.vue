<template>
  <PageLayout 
    title="仪表盘"
    description="系统概览和快速访问"
    icon-class="fa fa-tachometer"
  >
    <!-- 刷新按钮 -->
    <div class="refresh-section">
      <button 
        @click="fetchDashboardData" 
        :disabled="loading"
        class="refresh-btn"
      >
        <i class="fa fa-refresh" :class="{ 'fa-spin': loading }"></i>
        {{ loading ? '加载中...' : '刷新数据' }}
      </button>
    </div>

    <!-- 统计卡片 -->
    <div class="stats-grid">
      <div class="stat-card">
        <div class="stat-icon">
          <i class="fa fa-image text-blue-500"></i>
        </div>
        <div class="stat-content">
          <div class="stat-value">
            <span v-if="loading" class="loading-indicator">...</span>
            <span v-else>{{ dashboardData.image_count }}</span>
          </div>
          <div class="stat-label">总图片数</div>
        </div>
      </div>
      
      <div class="stat-card">
        <div class="stat-icon">
          <i class="fa fa-cloud text-green-500"></i>
        </div>
        <div class="stat-content">
          <div class="stat-value">
            <span v-if="loading" class="loading-indicator">...</span>
            <span v-else>{{ formatStorageSize(dashboardData.used_size_mb) }}</span>
          </div>
          <div class="stat-label">已用存储</div>
        </div>
      </div>
      
      <div class="stat-card">
        <div class="stat-icon">
          <i class="fa fa-upload text-orange-500"></i>
        </div>
        <div class="stat-content">
          <div class="stat-value">
            <span v-if="loading" class="loading-indicator">...</span>
            <span v-else>{{ dashboardData.today_upload_count }}</span>
          </div>
          <div class="stat-label">今日上传</div>
        </div>
      </div>
    </div>

    <!-- 快速操作 -->
    <div class="quick-actions">
      <h2 class="section-title">快速操作</h2>
      <div class="actions-grid">
        <router-link to="/admin" class="action-card">
          <div class="action-icon">
            <i class="fa fa-cloud-upload text-blue-500"></i>
          </div>
          <div class="action-content">
            <h3>上传图片</h3>
            <p>快速上传新图片</p>
          </div>
        </router-link>
        
        <router-link to="/admin/images" class="action-card">
          <div class="action-icon">
            <i class="fa fa-images text-green-500"></i>
          </div>
          <div class="action-content">
            <h3>我的图片</h3>
            <p>查看和管理图片</p>
          </div>
        </router-link>
        
        <router-link to="/admin/basic-settings" class="action-card">
          <div class="action-icon">
            <i class="fa fa-cog text-purple-500"></i>
          </div>
          <div class="action-content">
            <h3>偏好设置</h3>
            <p>默认策略与上传习惯</p>
          </div>
        </router-link>
      </div>
    </div>

    <!-- 最近活动 -->
    <div class="recent-activity">
      <h2 class="section-title">最近活动</h2>
      <div class="activity-list">
        <div class="activity-item">
          <div class="activity-icon">
            <i class="fa fa-upload text-green-500"></i>
          </div>
          <div class="activity-content">
            <div class="activity-title">上传了新图片</div>
            <div class="activity-time">2小时前</div>
          </div>
        </div>
        
        <div class="activity-item">
          <div class="activity-icon">
            <i class="fa fa-image text-blue-500"></i>
          </div>
          <div class="activity-content">
            <div class="activity-title">图片管理</div>
            <div class="activity-time">1天前</div>
          </div>
        </div>
      </div>
    </div>
  </PageLayout>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import PageLayout from '../admin/PageLayout.vue'
import { userAPI } from '../../api'

// 响应式数据
const dashboardData = ref({
  image_count: 0,
  today_upload_count: 0,
  used_size_mb: 0
})

// 加载状态
const loading = ref(false)

// 格式化存储大小
const formatStorageSize = (sizeInMB: number): string => {
  if (sizeInMB < 1) {
    return `${(sizeInMB * 1024).toFixed(2)} KB`
  } else if (sizeInMB < 1024) {
    return `${sizeInMB.toFixed(2)} MB`
  } else {
    return `${(sizeInMB / 1024).toFixed(2)} GB`
  }
}

// 获取仪表盘数据
const fetchDashboardData = async () => {
  loading.value = true
  try {
    const response = await userAPI.getDashboardData()
    if (response.data.status === 'success') {
      dashboardData.value = response.data.data.dashboard
    } else {
      console.error('获取仪表盘数据失败:', response.data.message)
    }
  } catch (error) {
    console.error('获取仪表盘数据失败:', error)
  } finally {
    loading.value = false
  }
}

// 组件挂载时获取数据
onMounted(() => {
  fetchDashboardData()
})
</script>

<style scoped>
.dashboard {
  max-width: 1200px;
  margin: 0 auto;
}

.page-header {
  margin-bottom: 32px;
}

.section-title {
  font-size: 18px;
  font-weight: 600;
  color: #374151;
  margin-bottom: 16px;
}

/* 统计卡片 */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 16px;
  margin-bottom: 32px;
}

.stat-card {
  background-color: white;
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.05);
  display: flex;
  align-items: center;
  gap: 16px;
}

.stat-icon {
  font-size: 24px;
  width: 48px;
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  background-color: #f3f4f6;
}

.stat-content {
  flex: 1;
}

.stat-value {
  font-size: 24px;
  font-weight: 700;
  color: #374151;
  margin-bottom: 4px;
}

.stat-label {
  font-size: 14px;
  color: #6b7280;
}

/* 快速操作 */
.quick-actions {
  margin-bottom: 32px;
}

.actions-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 16px;
}

.action-card {
  background-color: white;
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.05);
  text-decoration: none;
  color: inherit;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  gap: 16px;
}

.action-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0,0,0,0.1);
}

.action-icon {
  font-size: 24px;
  width: 48px;
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  background-color: #f3f4f6;
}

.action-content h3 {
  font-size: 16px;
  font-weight: 600;
  color: #374151;
  margin-bottom: 4px;
}

.action-content p {
  font-size: 14px;
  color: #6b7280;
}

/* 最近活动 */
.recent-activity {
  margin-bottom: 32px;
}

.activity-list {
  background-color: white;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.05);
  overflow: hidden;
}

.activity-item {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px 20px;
  border-bottom: 1px solid #f3f4f6;
}

.activity-item:last-child {
  border-bottom: none;
}

.activity-icon {
  font-size: 16px;
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 6px;
  background-color: #f3f4f6;
}

.activity-content {
  flex: 1;
}

.activity-title {
  font-size: 14px;
  font-weight: 500;
  color: #374151;
  margin-bottom: 2px;
}

.activity-time {
  font-size: 12px;
  color: #6b7280;
}

/* 刷新按钮 */
.refresh-section {
  text-align: right;
  margin-bottom: 32px;
}

.refresh-btn {
  background-color: #4f46e5;
  color: white;
  padding: 8px 16px;
  border-radius: 8px;
  border: none;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  font-weight: 600;
  transition: background-color 0.2s;
}

.refresh-btn:hover:not(:disabled) {
  background-color: #4338ca;
}

.refresh-btn:disabled {
  background-color: #d1d5db;
  color: #6b7280;
  cursor: not-allowed;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .stats-grid {
    grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
    gap: 12px;
  }
  
  .actions-grid {
    grid-template-columns: 1fr;
  }
  
  .stat-card,
  .action-card {
    padding: 16px;
  }
  
  .stat-icon,
  .action-icon {
    font-size: 20px;
    width: 40px;
    height: 40px;
  }
  
  .stat-value {
    font-size: 20px;
  }
}

/* 加载指示器样式 */
.loading-indicator {
  color: #6b7280;
  font-weight: 400;
}
</style> 