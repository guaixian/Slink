<template>
  <div class="storage-strategy">
    <!-- 页面标题 -->
    <div class="page-header">
      <h1 class="text-2xl font-bold text-neutral-800 mb-2">
        <i class="fa fa-database mr-3 text-primary"></i>
        储存策略
      </h1>
      <p class="text-neutral-600 text-sm">
        管理系统储存策略和分区配置
      </p>
    </div>

    <!-- 操作栏 -->
    <div class="action-bar">
      <button class="create-btn" @click="handleCreate">
        <i class="fa fa-plus mr-2"></i>
        创建储存策略
      </button>
      <div class="search-box">
        <input 
          type="text" 
          placeholder="输入名称回车搜索..."
          class="search-input"
          v-model="searchKeyword"
          @keyup.enter="handleSearch"
        >
        <i class="fa fa-search search-icon"></i>
      </div>
    </div>

    <!-- 数据表格 -->
    <div class="table-container">
      <div v-if="loading" class="loading-state">
        <i class="fa fa-spinner fa-spin"></i>
        <span>加载中...</span>
      </div>
      <table v-else class="data-table">
        <thead>
          <tr>
            <th>ID</th>
            <th>名称</th>
            <th>驱动</th>
            <th>图片数量</th>
            <th>可用范围</th>
            <th>已使用储存</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="strategy in filteredStrategies" :key="strategy.id">
            <td>{{ strategy.id }}</td>
            <td>{{ strategy.name }}</td>
            <td>
              <span class="driver-badge">{{ strategy.key }}</span>
            </td>
            <td>{{ strategy.image_count }}</td>
            <td>
              <span class="text-neutral-600 text-sm">全局</span>
            </td>
            <td>{{ formatSize(strategy.used_size_mb) }}</td>
            <td>
              <div class="action-buttons">
                <button class="edit-btn" @click="handleEdit(strategy)">
                  编辑
                </button>
                <button 
                  class="delete-btn" 
                  @click="handleDelete(strategy)"
                  :disabled="deletingId === strategy.id"
                >
                  <i v-if="deletingId === strategy.id" class="fa fa-spinner fa-spin mr-1"></i>
                  {{ deletingId === strategy.id ? '删除中...' : '删除' }}
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
      <div v-if="!loading && strategies.length === 0" class="empty-state">
        <i class="fa fa-database"></i>
        <span>暂无存储策略</span>
      </div>
    </div>


  </div>
</template>

<script setup lang="ts">
import { ref, computed, reactive, onMounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { systemAPI } from '../../api'
import { useMessage } from '../../composables/useMessage'

const router = useRouter()
const route = useRoute()

// 响应式数据
const searchKeyword = ref('')
const loading = ref(false)
const deletingId = ref<number | null>(null)

const { toast, confirm } = useMessage()

// 储存策略数据接口
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

// 储存策略数据
const strategies = ref<StorageStrategy[]>([])

// 加载策略数据
const loadStrategies = async () => {
  loading.value = true
  try {
    const response = await systemAPI.getStorageStrategies()
    console.log('API响应:', response)
    if (response.data && response.data.status) {
      strategies.value = response.data.data || []
      console.log('加载的策略列表:', strategies.value)
    } else {
      console.error('API返回失败:', response.data?.message)
      toast.error(`加载策略失败: ${response.data?.message || '未知错误'}`)
    }
  } catch (error) {
    console.error('加载策略数据失败:', error)
    toast.error('加载策略失败，请检查网络连接')
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

// 页面初始化
onMounted(() => {
  loadStrategies()
})

// 监听路由变化，当从编辑页面返回时刷新数据
watch(() => route.path, (newPath, oldPath) => {
  if (newPath === '/admin/storage' && oldPath && oldPath.includes('/edit')) {
    loadStrategies()
  }
})

// 过滤后的策略
const filteredStrategies = computed(() => {
  if (!searchKeyword.value) return strategies.value
  return strategies.value.filter(strategy => 
    strategy.name.toLowerCase().includes(searchKeyword.value.toLowerCase())
  )
})

// 搜索处理
const handleSearch = () => {
  console.log('搜索关键词:', searchKeyword.value)
}

// 创建处理
const handleCreate = () => {
  router.push('/admin/storage/create')
}

// 编辑处理
const handleEdit = (strategy: any) => {
  router.push(`/admin/storage/edit/${strategy.id}`)
}

// 删除处理
const handleDelete = async (strategy: any) => {
  const confirmed = await confirm.danger(`确定要删除存储策略"${strategy.name}"吗？\n\n删除后：\n• 该策略下的所有图片将无法访问\n• 此操作不可恢复`, '删除存储策略')
  if (!confirmed) {
    return
  }

  deletingId.value = strategy.id
  try {
    const response = await systemAPI.deleteStorageStrategy(strategy.id)
    console.log('删除响应:', response)
    if (response.data && response.data.status) {
      // 重新加载数据
      await loadStrategies()
      toast.success('删除成功！')
    } else {
      toast.error(`删除失败: ${response.data?.message || '未知错误'}`)
    }
  } catch (error: any) {
    console.error('删除策略失败:', error)
    const errorMsg = error.response?.data?.error || error.message || '未知错误'
    toast.error(`删除失败: ${errorMsg}`)
  } finally {
    deletingId.value = null
  }
}


</script>

<style scoped>
.storage-strategy {
  max-width: 1200px;
  margin: 0 auto;
}

.page-header {
  margin-bottom: 24px;
}

.action-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.create-btn {
  background-color: #6b7280;
  color: white;
  padding: 10px 20px;
  border-radius: 6px;
  border: none;
  font-size: 14px;
  cursor: pointer;
  transition: background-color 0.2s;
}

.create-btn:hover {
  background-color: #4b5563;
}

.search-box {
  position: relative;
  width: 300px;
}

.search-input {
  width: 100%;
  padding: 10px 40px 10px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 14px;
  background-color: #f9fafb;
}

.search-input:focus {
  outline: none;
  border-color: #3b82f6;
  background-color: white;
}

.search-icon {
  position: absolute;
  right: 12px;
  top: 50%;
  transform: translateY(-50%);
  color: #9ca3af;
}

.table-container {
  background-color: white;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.05);
  overflow: hidden;
}

.data-table {
  width: 100%;
  border-collapse: collapse;
}

.data-table th {
  background-color: #f9fafb;
  padding: 12px 16px;
  text-align: left;
  font-weight: 600;
  color: #374151;
  border-bottom: 1px solid #e5e7eb;
}

.data-table td {
  padding: 12px 16px;
  border-bottom: 1px solid #f3f4f6;
  color: #374151;
}

.data-table tr:hover {
  background-color: #f9fafb;
}

.driver-badge {
  background-color: #3b82f6;
  color: white;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
}

.role-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.role-tag {
  background-color: #e5e7eb;
  color: #111827;
  padding: 2px 8px;
  border-radius: 9999px;
  font-size: 12px;
  line-height: 1.4;
}

.action-buttons {
  display: flex;
  gap: 8px;
}

.edit-btn, .delete-btn {
  background: none;
  border: none;
  font-size: 14px;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 4px;
  transition: background-color 0.2s;
}

.edit-btn {
  color: #8b5cf6;
}

.edit-btn:hover {
  background-color: #f3f4f6;
}

.delete-btn {
  color: #ef4444;
}

.delete-btn:hover {
  background-color: #fef2f2;
}

.delete-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.delete-btn:disabled:hover {
  background-color: transparent;
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
}

.loading-state i {
  font-size: 24px;
  margin-bottom: 12px;
}

/* 空状态 */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  color: #9ca3af;
  font-size: 14px;
}

.empty-state i {
  font-size: 48px;
  margin-bottom: 16px;
  opacity: 0.5;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .action-bar {
    flex-direction: column;
    gap: 16px;
    align-items: stretch;
  }
  
  .search-box {
    width: 100%;
  }
  
  .data-table {
    font-size: 14px;
  }
  
  .data-table th, .data-table td {
    padding: 10px 12px;
  }
}
</style>
