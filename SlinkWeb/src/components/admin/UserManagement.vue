<template>
  <PageLayout 
    title="用户管理"
    description="管理系统用户账户和权限"
    icon-class="fa fa-users"
  >

    <!-- 操作栏 -->
    <div class="action-bar">
      <div class="search-box">
        <input
          type="text"
          placeholder="输入关键字回车搜索..."
          class="search-input"
          v-model="searchKeyword"
          @keyup.enter="handleSearch"
        >
        <i class="fa fa-search search-icon"></i>
      </div>
      <div class="action-right">
        <button class="create-btn" @click="handleCreate">
          <i class="fa fa-plus"></i>
          创建用户
        </button>
        <div class="loading-indicator" v-if="loading">
          <i class="fa fa-spinner fa-spin"></i>
          加载中...
        </div>
      </div>
    </div>

    <!-- 数据表格 -->
    <div class="table-container">
      <table class="data-table">
        <thead>
          <tr>
            <th>ID</th>
            <th>用户名</th>
            <th>登录账号</th>
            <th>管理员</th>
            <th>图片数量</th>
            <th>已用 / 容量</th>
            <th>策略组</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="user in filteredUsers" :key="user.id">
            <td>{{ user.id }}</td>
            <td>{{ user.name }}</td>
            <td>{{ user.account }}</td>
            <td>
              <span class="role-badge" :class="user.is_admin ? 'admin-badge' : 'user-badge'">
                {{ user.is_admin ? '管理员' : '普通用户' }}
              </span>
            </td>
            <td>{{ user.image_nums }}</td>
            <td>{{ formatUsedSize(user.used_size) }} / {{ formatCapacity(user.capacity) }}</td>
            <td>{{ getPolicyGroupName(user.policy_group_id) }}</td>
            <td>
              <div class="action-buttons">
                <button class="detail-btn" @click="handleDetail(user)">
                  详细
                </button>
                <button class="edit-btn" @click="handleEdit(user)">
                  编辑
                </button>
                <button class="delete-btn" @click="handleDelete(user)">
                  删除
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 用户详情弹窗 -->
    <UserDetailModal 
      v-if="showDetailModal" 
      :user="selectedUser" 
      @close="showDetailModal = false"
    />



  </PageLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { adminAPI } from '../../api'
import PageLayout from '../admin/PageLayout.vue'
import UserDetailModal from '../admin/UserDetailModal.vue'
import { useMessage } from '../../composables/useMessage'

const router = useRouter()

// 响应式数据 
const searchKeyword = ref('')
const showDetailModal = ref(false)
const selectedUser = ref<any>(null)
const loading = ref(false)

const { toast, confirm } = useMessage()

// 用户数据
const users = ref<any[]>([])

// 过滤后的用户
const filteredUsers = computed(() => {
  if (!searchKeyword.value) return users.value
  return users.value.filter(user => 
    user.name.toLowerCase().includes(searchKeyword.value.toLowerCase()) ||
    user.account.toLowerCase().includes(searchKeyword.value.toLowerCase())
  )
})

// 策略组名称映射
const policyGroupNames = ref<Record<number, string>>({})

const loadPolicyGroups = async () => {
  try {
    const res = await adminAPI.getPolicyGroups()
    if (res.data?.status) {
      const map: Record<number, string> = {}
      for (const g of res.data.data || []) {
        map[g.id] = g.name
      }
      policyGroupNames.value = map
    }
  } catch { /* ignore */ }
}

const getPolicyGroupName = (id: number) => {
  if (!id) return '—'
  return policyGroupNames.value[id] || `ID:${id}`
}

// 加载用户列表
const loadUsers = async () => {
  loading.value = true
  try {
    const response = await adminAPI.getAllUsers()
    if (response.data.status) {
      users.value = response.data.data
      console.log('用户数据:', users.value)
    }
  } catch (error) {
    console.error('加载用户列表失败:', error)
  } finally {
    loading.value = false
  }
}

// 格式化已用空间（used_size 单位为 MB）
const formatUsedSize = (mb: number) => {
  if (!mb) return '0 MB'
  if (mb < 1024) return `${mb.toFixed(2)} MB`
  return `${(mb / 1024).toFixed(2)} GB`
}

// 格式化容量上限（capacity 单位为字节，0 表示无限制）
const formatCapacity = (bytes: number) => {
  if (!bytes) return '无限制'
  const mb = bytes / (1024 * 1024)
  if (mb < 1024) return `${mb.toFixed(0)} MB`
  return `${(mb / 1024).toFixed(2)} GB`
}

// 创建用户
const handleCreate = () => {
  router.push('/admin/users/create')
}

// 页面初始化
onMounted(() => {
  loadUsers()
  loadPolicyGroups()
})

// 搜索处理
const handleSearch = () => {
  console.log('搜索关键词:', searchKeyword.value)
}

// 详细处理
const handleDetail = (user: any) => {
  selectedUser.value = user
  showDetailModal.value = true
}

// 编辑处理
const handleEdit = (user: any) => {
  router.push(`/admin/users/edit/${user.id}`)
}

// 删除处理
const handleDelete = async (user: any) => {
  const confirmed = await confirm.danger(`确定要删除用户 "${user.name}" 吗？此操作不可恢复。`, '删除用户')
  if (!confirmed) {
    return
  }

  try {
    const response = await adminAPI.deleteUser(user.id)

    if (response.data.status) {
      const index = users.value.findIndex((u: any) => u.id === user.id)
      if (index !== -1) {
        users.value.splice(index, 1)
      }
      toast.success('删除成功')
    } else {
      toast.error(`删除失败: ${response.data.message}`)
    }
  } catch (error: any) {
    console.error('删除用户失败:', error)
    toast.error(`删除失败: ${error.response?.data?.message || error.message}`)
  }
}
</script>

<style scoped>
.user-management {
  max-width: 1400px;
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

.action-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.create-btn {
  padding: 10px 20px;
  background-color: #3b82f6;
  color: white;
  border: none;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  gap: 8px;
}

.create-btn:hover {
  background-color: #2563eb;
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
  font-size: 14px;
}

.data-table th {
  background-color: #f9fafb;
  padding: 12px 8px;
  text-align: left;
  font-weight: 600;
  color: #374151;
  border-bottom: 1px solid #e5e7eb;
  white-space: nowrap;
}

.data-table td {
  padding: 12px 8px;
  border-bottom: 1px solid #f3f4f6;
  color: #374151;
  white-space: nowrap;
}

.data-table tr:hover {
  background-color: #f9fafb;
}

.role-badge {
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
}

.admin-badge {
  background-color: #dc2626;
  color: white;
}

.user-badge {
  background-color: #3b82f6;
  color: white;
}

/* 加载指示器 */
.loading-indicator {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #6b7280;
  font-size: 14px;
}

.action-buttons {
  display: flex;
  gap: 8px;
}

.detail-btn, .edit-btn, .delete-btn {
  background: none;
  border: none;
  font-size: 14px;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 4px;
  transition: background-color 0.2s;
}

.detail-btn {
  color: #3b82f6;
}

.edit-btn {
  color: #059669;
}

.delete-btn {
  color: #dc2626;
}

.detail-btn:hover, .edit-btn:hover, .delete-btn:hover {
  background-color: #f3f4f6;
}

/* 响应式设计 */
@media (max-width: 1200px) {
  .data-table {
    font-size: 12px;
  }
  
  .data-table th,
  .data-table td {
    padding: 8px 4px;
  }
}

@media (max-width: 768px) {
  .action-bar {
    flex-direction: column;
    gap: 16px;
    align-items: stretch;
  }
  
  .search-box {
    width: 100%;
  }
  
  .table-container {
    overflow-x: auto;
  }
}
</style> 
