<template>
  <div class="system-console">
    <!-- 页面标题 -->
    <div class="page-header">
      <h1 class="text-2xl font-bold text-neutral-800 mb-2">
        <i class="fa fa-terminal mr-3 text-primary"></i>
        系统控制台
      </h1>
      <p class="text-neutral-600 text-sm">
        系统概览和运行状态监控
      </p>
    </div>

    <!-- 概览统计 -->
    <div class="overview-section">
      <h2 class="section-title">概览</h2>
      <div v-if="loading" class="loading-state">
        <i class="fa fa-spinner fa-spin"></i>
        <span>加载中...</span>
      </div>
      <div v-else class="stats-grid">
        <div class="stat-card">
          <div class="stat-icon">
            <i class="fa fa-image text-red-500"></i>
          </div>
          <div class="stat-content">
            <div class="stat-value text-red-500">{{ stats.imageCount }}</div>
            <div class="stat-label">图片数量</div>
          </div>
        </div>

        <div class="stat-card">
          <div class="stat-icon">
            <i class="fa fa-folder text-blue-500"></i>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.albumCount }}</div>
            <div class="stat-label">相册数量</div>
          </div>
        </div>

        <div class="stat-card">
          <div class="stat-icon">
            <i class="fa fa-user text-green-500"></i>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.userCount }}</div>
            <div class="stat-label">账号（单用户）</div>
          </div>
        </div>

        <div class="stat-card">
          <div class="stat-icon">
            <i class="fa fa-cloud text-purple-500"></i>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.usedStorage }}</div>
            <div class="stat-label">占用储存</div>
          </div>
        </div>

        <div class="stat-card">
          <div class="stat-icon">
            <i class="fa fa-upload text-orange-500"></i>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.todayUpload }}</div>
            <div class="stat-label">今日上传</div>
          </div>
        </div>

        <div class="stat-card">
          <div class="stat-icon">
            <i class="fa fa-upload text-orange-500"></i>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.yesterdayUpload }}</div>
            <div class="stat-label">昨日上传</div>
          </div>
        </div>

        <div class="stat-card">
          <div class="stat-icon">
            <i class="fa fa-upload text-orange-500"></i>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.weekUpload }}</div>
            <div class="stat-label">本周上传</div>
          </div>
        </div>

        <div class="stat-card">
          <div class="stat-icon">
            <i class="fa fa-upload text-orange-500"></i>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.monthUpload }}</div>
            <div class="stat-label">本月上传</div>
          </div>
        </div>
      </div>
    </div>

    <!-- 趋势图表 -->
    <div class="trend-section">
      <h2 class="section-title">近30天内统计</h2>
      <div class="chart-container">
        <div v-if="loading" class="chart-placeholder">
          <i class="fa fa-spinner fa-spin text-4xl text-neutral-300 mb-4"></i>
          <p class="text-neutral-500">加载中...</p>
        </div>
        <div v-else-if="trendData.length > 0" class="chart-wrapper">
          <div class="chart-bars">
            <div
              v-for="(item, index) in trendData"
              :key="index"
              class="chart-bar-wrapper"
              :title="`${item.date}: ${item.count}张`"
            >
              <div class="chart-bar" :style="{ height: getBarHeight(item.count) + '%' }"></div>
              <div class="chart-label" v-if="index % 5 === 0">{{ formatChartDate(item.date) }}</div>
            </div>
          </div>
        </div>
        <div v-else class="chart-placeholder">
          <i class="fa fa-bar-chart text-4xl text-neutral-300 mb-4"></i>
          <p class="text-neutral-500">暂无数据</p>
        </div>
        <div class="chart-legend">
          <div class="legend-item">
            <span class="legend-color blue"></span>
            <span>每日上传</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 系统信息 -->
    <div class="system-info-section">
      <h2 class="section-title">系统情况</h2>
      <div class="info-grid">
        <div class="info-item">
          <span class="info-label">操作系统:</span>
          <span class="info-value">{{ systemInfo.os || '加载中...' }}</span>
        </div>
        <div class="info-item">
          <span class="info-label">运行环境:</span>
          <span class="info-value">{{ systemInfo.webServer || '加载中...' }}</span>
        </div>
        <div class="info-item">
          <span class="info-label">Go 版本:</span>
          <span class="info-value">{{ systemInfo.goVersion || '加载中...' }}</span>
        </div>
        <div class="info-item">
          <span class="info-label">文件上传限制:</span>
          <span class="info-value">{{ systemInfo.uploadLimit || '加载中...' }}</span>
        </div>
        <div class="info-item">
          <span class="info-label">POST 数据最大限制:</span>
          <span class="info-value">{{ systemInfo.postLimit || '加载中...' }}</span>
        </div>
      </div>
    </div>

    <!-- 软件信息 -->
    <div class="software-info-section">
      <h2 class="section-title">软件信息</h2>
      <div class="info-grid">
        <div class="info-item">
          <span class="info-label">软件版本:</span>
          <span class="info-value">{{ softwareInfo.version || '加载中...' }}</span>
        </div>
        <div class="info-item" v-if="softwareInfo.description">
          <span class="info-label">软件描述:</span>
          <span class="info-value">{{ softwareInfo.description }}</span>
        </div>
        <div class="info-item" v-if="softwareInfo.website">
          <span class="info-label">官方网站:</span>
          <a :href="softwareInfo.website" class="info-link" target="_blank">{{ softwareInfo.website }}</a>
        </div>
        <div class="info-item" v-if="softwareInfo.docs">
          <span class="info-label">使用手册:</span>
          <a :href="softwareInfo.docs" class="info-link" target="_blank">{{ softwareInfo.docs }}</a>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { adminAPI } from '../../api'

// 响应式数据
const loading = ref(false)
const stats = ref({
  imageCount: 0,
  albumCount: 0,
  userCount: 0,
  usedStorage: '0 MB',
  todayUpload: 0,
  yesterdayUpload: 0,
  weekUpload: 0,
  monthUpload: 0
})

const systemInfo = ref({
  os: '',
  webServer: '',
  goVersion: '',
  uploadLimit: '',
  postLimit: ''
})

const softwareInfo = ref({
  version: '',
  website: '',
  docs: '',
  description: ''
})

const trendData = ref<Array<{ date: string; count: number }>>([])
const maxCount = ref(0)

// 加载系统统计信息
const loadSystemStats = async () => {
  loading.value = true
  try {
    const response = await adminAPI.getSystemStats()
    console.log('系统统计API响应:', response)
    if (response.data && response.data.status) {
      const data = response.data.data
      console.log('系统统计数据:', data)

      // 更新统计数据
      stats.value = {
        imageCount: data.total_images || 0,
        albumCount: data.total_albums || 0,
        userCount: data.total_users || 0,
        usedStorage: formatSize(data.total_size_mb || 0),
        todayUpload: data.today_upload || 0,
        yesterdayUpload: data.yesterday_upload || 0,
        weekUpload: data.week_upload || 0,
        monthUpload: data.month_upload || 0
      }

      // 更新趋势数据
      if (data.trend_data && Array.isArray(data.trend_data)) {
        trendData.value = data.trend_data
        maxCount.value = Math.max(...data.trend_data.map((item: any) => item.count), 1)
      }

      // 更新系统信息
      if (data.system_info) {
        systemInfo.value = {
          os: data.system_info.os || 'Unknown',
          webServer: data.system_info.web_server || 'Unknown',
          goVersion: data.system_info.go_version || 'N/A',
          uploadLimit: data.system_info.upload_limit || 'Unknown',
          postLimit: data.system_info.post_limit || 'Unknown'
        }
      }

      // 更新软件信息
      if (data.software_info) {
        softwareInfo.value = {
          version: data.software_info.version || 'Unknown',
          website: data.software_info.website || '',
          docs: data.software_info.docs || '',
          description: data.software_info.description || ''
        }
      }
    } else {
      console.error('API返回失败:', response.data?.message)
    }
  } catch (error) {
    console.error('加载系统统计失败:', error)
  } finally {
    loading.value = false
  }
}

// 格式化文件大小
const formatSize = (sizeInMB: number) => {
  if (sizeInMB < 1) {
    return `${(sizeInMB * 1024).toFixed(2)} KB`
  } else if (sizeInMB < 1024) {
    return `${sizeInMB.toFixed(2)} MB`
  } else {
    return `${(sizeInMB / 1024).toFixed(2)} GB`
  }
}

// 计算柱状图高度
const getBarHeight = (count: number) => {
  if (maxCount.value === 0) return 0
  return (count / maxCount.value) * 100
}

// 格式化图表日期
const formatChartDate = (dateStr: string) => {
  const date = new Date(dateStr)
  return `${date.getMonth() + 1}/${date.getDate()}`
}

// 页面初始化
onMounted(() => {
  loadSystemStats()
})
</script>

<style scoped>
.system-console {
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

/* 概览统计 */
.overview-section {
  margin-bottom: 32px;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
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

/* 加载状态 */
.loading-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  color: #6b7280;
  font-size: 14px;
  background-color: white;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.05);
}

.loading-state i {
  font-size: 24px;
  margin-bottom: 12px;
}

/* 趋势图表 */
.trend-section {
  margin-bottom: 32px;
}

.chart-container {
  background-color: white;
  border-radius: 8px;
  padding: 24px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.05);
}

.chart-placeholder {
  height: 300px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background-color: #f9fafb;
  border-radius: 8px;
  border: 2px dashed #d1d5db;
}

.chart-wrapper {
  height: 300px;
  margin-bottom: 20px;
}

.chart-bars {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  height: 100%;
  gap: 2px;
  padding: 0 10px;
}

.chart-bar-wrapper {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  height: 100%;
  position: relative;
}

.chart-bar {
  width: 100%;
  background: linear-gradient(to top, #3b82f6, #60a5fa);
  border-radius: 4px 4px 0 0;
  transition: all 0.3s ease;
  cursor: pointer;
  min-height: 2px;
}

.chart-bar:hover {
  background: linear-gradient(to top, #2563eb, #3b82f6);
  transform: scaleY(1.05);
}

.chart-label {
  position: absolute;
  bottom: -20px;
  font-size: 10px;
  color: #6b7280;
  white-space: nowrap;
}

.chart-legend {
  display: flex;
  justify-content: center;
  gap: 24px;
  margin-top: 30px;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  color: #6b7280;
}

.legend-color {
  width: 12px;
  height: 12px;
  border-radius: 2px;
}

.legend-color.blue {
  background-color: #3b82f6;
}

.legend-color.green {
  background-color: #10b981;
}

.legend-color.orange {
  background-color: #f59e0b;
}

/* 系统信息 */
.system-info-section,
.software-info-section {
  margin-bottom: 32px;
}

.info-grid {
  background-color: white;
  border-radius: 8px;
  padding: 24px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.05);
}

.info-item {
  display: flex;
  padding: 12px 0;
  border-bottom: 1px solid #f3f4f6;
}

.info-item:last-child {
  border-bottom: none;
}

.info-label {
  font-weight: 600;
  color: #374151;
  min-width: 150px;
  margin-right: 16px;
}

.info-value {
  color: #6b7280;
  flex: 1;
}

.info-link {
  color: #3b82f6;
  text-decoration: none;
}

.info-link:hover {
  text-decoration: underline;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .stats-grid {
    grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
    gap: 12px;
  }
  
  .stat-card {
    padding: 16px;
  }
  
  .stat-icon {
    font-size: 20px;
    width: 40px;
    height: 40px;
  }
  
  .stat-value {
    font-size: 20px;
  }
  
  .info-item {
    flex-direction: column;
    gap: 4px;
  }
  
  .info-label {
    min-width: auto;
    margin-right: 0;
  }
  
  .chart-legend {
    flex-direction: column;
    gap: 12px;
  }
}
</style> 