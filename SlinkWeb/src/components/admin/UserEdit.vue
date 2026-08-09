<template>
  <div class="user-edit-page">
    <!-- 编辑表单 -->
    <div class="edit-container">
      <div class="edit-form">
        <!-- 基本信息 -->
        <div class="form-section">
          <div class="form-group">
            <label class="form-label">登录账号</label>
            <input type="text" v-model="formData.account" class="form-input" placeholder="登录账号" />
          </div>

          <div class="form-group">
            <label class="form-label">*用户名</label>
            <input type="text" v-model="formData.name" class="form-input" placeholder="请输入用户名" required />
          </div>
        </div>

        <!-- 容量设置 -->
        <div class="form-section">
          <div class="form-group">
            <label class="form-label">*总容量(MB)</label>
            <input type="number" v-model.number="formData.capacity" class="form-input" placeholder="请输入容量限制" min="0" />
            <small class="form-hint">0 表示无限制</small>
          </div>
        </div>

        <!-- 安全设置 -->
        <div class="form-section">
          <div class="form-group">
            <label class="form-label">新密码</label>
            <input type="password" v-model="formData.newPassword" class="form-input" placeholder="不修改请留空" />
            <small class="form-hint">留空表示不修改密码</small>
          </div>

          <div class="form-group">
            <label class="form-label">管理员权限</label>
            <p class="status-description">设置用户是否拥有管理员权限</p>
            <div class="status-toggle">
              <label class="toggle-item" :class="{ active: !formData.isAdmin }">
                <input type="radio" :value="false" v-model="formData.isAdmin" />
                <span class="toggle-text">普通用户</span>
              </label>
              <label class="toggle-item" :class="{ active: formData.isAdmin }">
                <input type="radio" :value="true" v-model="formData.isAdmin" />
                <span class="toggle-text">管理员</span>
              </label>
            </div>
          </div>
        </div>

        <!-- 上传策略组 -->
        <div class="form-section">
          <div class="form-group">
            <label class="form-label">上传策略组</label>
            <select v-model.number="selectedPolicyGroupId" class="form-input">
              <option :value="0">-- 请选择 --</option>
              <option v-for="pg in policyGroups" :key="pg.id" :value="pg.id">
                {{ pg.name }} {{ pg.is_default === 1 ? '(默认)' : '' }}
              </option>
            </select>
          </div>
        </div>

        <!-- 存储策略分配 -->
        <div class="form-section" v-if="allStrategies.length > 0">
          <div class="form-group">
            <label class="form-label">分配存储策略</label>
            <p class="status-description">选择该用户可使用的上传策略</p>
            <div v-if="loadingStrategies" class="text-sm text-gray-500">加载策略中...</div>
            <div v-else class="strategy-list">
              <label
                v-for="strategy in allStrategies"
                :key="strategy.id"
                class="strategy-item"
              >
                <input
                  type="checkbox"
                  :checked="assignedStrategyIds.includes(strategy.id)"
                  @change="toggleStrategy(strategy.id)"
                />
                <span class="strategy-name">{{ strategy.name }}</span>
                <span class="strategy-desc">{{ strategy.introduction }}</span>
              </label>
            </div>
          </div>
        </div>

        <!-- 底部操作按钮 -->
        <div class="form-actions">
          <button class="btn btn-secondary" @click="handleCancel">
            <i class="fa fa-times mr-1"></i>
            取消
          </button>
          <button class="btn btn-primary" @click="handleSave" :disabled="loading">
            <i v-if="loading" class="fa fa-spinner fa-spin mr-1"></i>
            <i v-else class="fa fa-save mr-1"></i>
            {{ loading ? '保存中...' : '保存' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { adminAPI, systemAPI } from '../../api'
import { http } from '../../utils/request'
import { useMessage } from '../../composables/useMessage'

const route = useRoute()
const router = useRouter()
const loading = ref(false)

const { toast, confirm } = useMessage()

// 判断是创建还是编辑模式
const isCreateMode = computed(() => !route.params.id)

// 用户数据接口
interface User {
  id: number
  name: string
  account: string
  capacity: number
  is_admin: number
}

// 表单数据
const formData = reactive({
  account: '',
  name: '',
  capacity: 0,
  newPassword: '',
  isAdmin: false
})

// 策略组
interface PolicyGroup { id: number; name: string; is_default: number }
const policyGroups = ref<PolicyGroup[]>([])
const selectedPolicyGroupId = ref(0)

const loadPolicyGroups = async () => {
  try {
    const res = await adminAPI.getPolicyGroups()
    if (res.data?.status) policyGroups.value = res.data.data || []
  } catch { /* ignore */ }
}

// 策略管理
interface Strategy {
  id: number
  name: string
  introduction: string
  key: string
}
const allStrategies = ref<Strategy[]>([])
const assignedStrategyIds = ref<number[]>([])
const loadingStrategies = ref(false)

const loadStrategies = async () => {
  loadingStrategies.value = true
  try {
    const [strategiesRes, userStrategiesRes] = await Promise.all([
      systemAPI.getStorageStrategies(),
      !isCreateMode.value && route.params.id
        ? adminAPI.getUserStrategies(Number(route.params.id))
        : Promise.resolve(null)
    ])
    if (strategiesRes.data.status) {
      allStrategies.value = strategiesRes.data.data || []
    }
    if (userStrategiesRes?.data?.status) {
      assignedStrategyIds.value = userStrategiesRes.data.data?.strategy_ids || []
    }
  } catch (e) {
    console.error('加载策略失败:', e)
  } finally {
    loadingStrategies.value = false
  }
}

const toggleStrategy = (strategyId: number) => {
  const idx = assignedStrategyIds.value.indexOf(strategyId)
  if (idx >= 0) {
    assignedStrategyIds.value.splice(idx, 1)
  } else {
    assignedStrategyIds.value.push(strategyId)
  }
}

// 加载用户数据
const loadUser = async () => {
  await loadPolicyGroups()
  if (isCreateMode.value) {
    await loadStrategies()
    return
  }

  const userId = route.params.id
  if (!userId) {
    toast.error('用户ID不能为空')
    router.push('/admin/users')
    return
  }

  loading.value = true
  try {
    const response = await adminAPI.getUser(Number(userId))
    if (response.data && response.data.status) {
      const userData = response.data.data
      Object.assign(formData, {
        account: userData.account ?? '',
        name: userData.name ?? '',
        // 后端 capacity 单位为字节，界面以 MB 展示
        capacity: Math.round(Number((userData as any).capacity ?? 0) / (1024 * 1024)),
        isAdmin: Number((userData as any).is_admin ?? 0) === 1
      })
      selectedPolicyGroupId.value = (userData as any).policy_group_id ?? 0
      await loadStrategies()
    } else {
      toast.error(`加载用户数据失败: ${response.data?.message || '未知错误'}`)
    }
  } catch (error: any) {
    console.error('加载用户数据失败:', error)
    const errorMsg = error.response?.data?.error || error.message || '未知错误'
    toast.error(`加载用户数据失败: ${errorMsg}`)
  } finally {
    loading.value = false
  }
}

// 保存用户数据
const handleSave = async () => {
  if (!formData.name.trim()) {
    toast.warning('用户名不能为空')
    return
  }

  if (!formData.account.trim()) {
    toast.warning('登录账号不能为空')
    return
  }

  if (isCreateMode.value && !formData.newPassword.trim()) {
    toast.warning('创建用户需要设置密码')
    return
  }

  loading.value = true
  try {
    const userData: any = {
      name: formData.name,
      account: formData.account,
      // 界面单位为 MB，后端存储单位为字节
      capacity: Math.round(Number(formData.capacity || 0) * 1024 * 1024),
      is_admin: formData.isAdmin ? 1 : 0
    }

    if (formData.newPassword.trim()) {
      userData.password = formData.newPassword
    }

    let userId: number

    if (isCreateMode.value) {
      // 创建新用户
      const response = await http.post('/api/admin/users', userData)
      if (response.data && response.data.status) {
        userId = response.data.data.id
        toast.success('创建用户成功')
      } else {
        toast.error(`创建失败：${response.data?.message || '未知错误'}`)
        return
      }
    } else {
      // 更新现有用户
      userId = Number(route.params.id)
      const response = await adminAPI.updateUser(userId, userData)
      if (response.data && response.data.status) {
        toast.success('保存成功')
      } else {
        toast.error(`保存失败：${response.data?.message || '未知错误'}`)
        return
      }
    }

    // 保存策略组分配
    if (userId && selectedPolicyGroupId.value > 0) {
      try {
        await adminAPI.setUserPolicyGroup(userId, selectedPolicyGroupId.value)
      } catch (e) { console.error('保存策略组失败:', e) }
    }

    // 保存策略分配（编辑模式下始终提交，允许清空全部分配）
    if (userId && (!isCreateMode.value || assignedStrategyIds.value.length > 0)) {
      try {
        await adminAPI.assignStrategiesToUser(userId, assignedStrategyIds.value)
      } catch (e) {
        console.error('保存策略失败:', e)
      }
    }

    router.push('/admin/users')
  } catch (error: any) {
    console.error('保存用户数据失败:', error)
    const errorMsg = error.response?.data?.error || error.message || '未知错误'
    toast.error(`保存失败: ${errorMsg}`)
  } finally {
    loading.value = false
  }
}

// 取消编辑
const handleCancel = async () => {
  const msg = isCreateMode.value ? '确定要取消创建吗？' : '确定要取消编辑吗？未保存的更改将丢失。'
  const title = isCreateMode.value ? '取消创建' : '取消编辑'
  const confirmed = await confirm.info(msg, title)
  if (confirmed) {
    router.push('/admin/users')
  }
}

// 页面初始化
onMounted(async () => {
  await loadUser()
})
</script>

<style scoped>
.user-edit-page {
  min-height: 100vh;
  background-color: #f8fafc;
}



/* 编辑容器 */
.edit-container {
  max-width: 800px;
  margin: 8px auto 0;
  padding: 16px;
}

.edit-form {
  background: white;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
  overflow: hidden;
}

/* 表单区域 */
.form-section {
  padding: 16px 24px;
  border-bottom: 1px solid #f3f4f6;
}

.form-section:last-child {
  border-bottom: none;
}

/* 表单组件 */
.form-group {
  margin-bottom: 20px;
}

.form-label {
  display: block;
  font-weight: 500;
  color: #374151;
  margin-bottom: 8px;
  font-size: 14px;
}

.form-input {
  width: 100%;
  padding: 12px 16px;
  border: 1px solid #d1d5db;
  border-radius: 8px;
  font-size: 14px;
  background-color: #f9fafb;
  transition: all 0.2s;
}

.form-input:focus {
  outline: none;
  border-color: #3b82f6;
  background-color: white;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.form-hint {
  display: block;
  margin-top: 4px;
  font-size: 12px;
  color: #6b7280;
}

/* 状态切换 */
.status-description {
  margin: 8px 0 0 0;
  color: #6b7280;
  font-size: 13px;
  line-height: 1.4;
  font-style: italic;
}

.status-toggle {
  display: flex;
  background: #f3f4f6;
  border-radius: 8px;
  padding: 4px;
  margin-top: 8px;
  border: 1px solid #e5e7eb;
}

.toggle-item {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  padding: 10px 16px;
  border-radius: 6px;
  transition: all 0.3s ease;
  position: relative;
  color: #6b7280;
  border: 1px solid transparent;
}

.toggle-item:hover {
  color: #374151;
  background-color: rgba(255, 255, 255, 0.5);
}

.toggle-item.active {
  background-color: white;
  color: #3b82f6;
  border-color: #e5e7eb;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.toggle-item input[type="radio"] {
  display: none;
}

.toggle-text {
  position: relative;
  z-index: 1;
}

/* 底部操作按钮 */
.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 24px;
  border-top: 1px solid #e5e7eb;
  background-color: #f9fafb;
}

/* 按钮样式 */
.btn {
  padding: 10px 20px;
  border: none;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
  display: flex;
  align-items: center;
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-secondary {
  background-color: #f3f4f6;
  color: #374151;
}

.btn-secondary:hover:not(:disabled) {
  background-color: #e5e7eb;
}

.btn-primary {
  background-color: #3b82f6;
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background-color: #2563eb;
}

.strategy-list {
  margin-top: 8px;
  max-height: 200px;
  overflow-y: auto;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  padding: 8px;
  background: #f9fafb;
}

.strategy-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px;
  cursor: pointer;
  border-radius: 4px;
  transition: background-color 0.2s;
}

.strategy-item:hover {
  background-color: #fff;
}

.strategy-item input[type="checkbox"] {
  flex-shrink: 0;
}

.strategy-name {
  font-weight: 500;
  font-size: 14px;
  color: #374151;
  flex-shrink: 0;
}

.strategy-desc {
  font-size: 12px;
  color: #9ca3af;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .edit-container {
    padding: 12px;
  }
  
  .form-section {
    padding: 12px 16px;
  }
  
  .form-actions {
    padding: 12px 16px;
    flex-direction: column;
  }
  
  .status-toggle {
    flex-direction: column;
    gap: 4px;
  }
  
  .toggle-item {
    padding: 12px 16px;
  }
}
</style> 
