<template>
  <div class="basic-settings">
    <!-- 页面标题 -->
    <div class="page-header">
      <h1 class="text-2xl font-bold text-neutral-800 mb-2">
        <i class="fa fa-cog mr-3 text-primary"></i>
        基础设置
        <span v-if="loading" class="ml-3 text-sm text-neutral-500">
          <i class="fa fa-spinner fa-spin"></i> 加载中...
        </span>
      </h1>
      <p class="text-neutral-600 text-sm">
        管理个人账户基本信息和偏好设置
      </p>
    </div>

    <!-- 设置表单 -->
    <div class="settings-form">
      <form @submit.prevent="handleSubmit">
        <div class="form-grid">
          <!-- 左列 -->
          <div class="form-column">
            <!-- 登录账号 -->
            <div class="form-group">
              <label class="form-label">登录账号</label>
              <input 
                type="text" 
                class="form-input"
                v-model="form.account"
                placeholder="登录账号"
                disabled
              >
            </div>

            <!-- 默认上传策略 -->
            <div class="form-group">
              <label class="form-label">默认上传策略</label>
              <select class="form-select" v-model="form.defaultStrategy" :disabled="loading">
                <option value="">请选择存储策略</option>
                <option 
                  v-for="strategy in strategies" 
                  :key="strategy.id" 
                  :value="strategy.id.toString()"
                >
                  {{ strategy.name }}
                </option>
              </select>
            </div>

            <!-- 密码 -->
            <div class="form-group">
              <label class="form-label">不修改请留空</label>
              <input 
                type="password" 
                class="form-input"
                v-model="form.password"
                placeholder="请输入新密码"
              >
            </div>

            <!-- 是否自动清除预览 -->
            <div class="form-group">
              <label class="form-label">设置上传时,文件上传完成以后是否自动清除预览图片</label>
              <div class="radio-group">
                <label class="radio-item">
                  <input 
                    type="radio" 
                    name="autoClear" 
                    :value="1"
                    v-model="form.autoClear"
                  >
                  <span>是</span>
                </label>
                <label class="radio-item">
                  <input 
                    type="radio" 
                    name="autoClear" 
                    :value="0"
                    v-model="form.autoClear"
                  >
                  <span>否</span>
                </label>
              </div>
            </div>

            <!-- 图片粘贴后动作 -->
            <div class="form-group">
              <label class="form-label">设置上传页面粘贴图片后的动作</label>
              <div class="radio-group">
                <label class="radio-item">
                  <input 
                    type="radio" 
                    name="pasteAction" 
                    :value="1"
                    v-model="form.pasteAction"
                  >
                  <span>直接上传</span>
                </label>
                <label class="radio-item">
                  <input 
                    type="radio" 
                    name="pasteAction" 
                    :value="0"
                    v-model="form.pasteAction"
                  >
                  <span>等待上传</span>
                </label>
              </div>
            </div>

            <!-- 图片默认权限 -->
            <div class="form-group">
              <label class="form-label">设置上传的图片默认的权限(公开还是私有,公开的图片将会出现在画廊中,你也可以通过图片管理单独设置权限)</label>
              <div class="radio-group">
                <label class="radio-item">
                  <input 
                    type="radio" 
                    name="defaultPermission" 
                    :value="1"
                    v-model="form.defaultPermission"
                  >
                  <span>私有</span>
                </label>
                <label class="radio-item">
                  <input 
                    type="radio" 
                    name="defaultPermission" 
                    :value="0"
                    v-model="form.defaultPermission"
                  >
                  <span>公开</span>
                </label>
              </div>
            </div>
          </div>

          <!-- 右列 -->
          <div class="form-column">
            <!-- 昵称 -->
            <div class="form-group">
              <label class="form-label">昵称</label>
              <input 
                type="text" 
                class="form-input"
                v-model="form.nickname"
                placeholder="请输入昵称"
              >
            </div>
          </div>
        </div>

        <!-- 保存按钮 -->
        <div class="form-actions">
          <button type="submit" class="save-btn" :disabled="loading">
            <i v-if="!loading" class="fa fa-save mr-2"></i>
            <i v-else class="fa fa-spinner fa-spin mr-2"></i>
            {{ loading ? '保存中...' : '保存设置' }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { userAPI } from '@/api'

// 表单数据
const form = reactive({
  account: '',
  nickname: '',
  defaultStrategy: '',
  password: '',
  autoClear: 0,
  pasteAction: 0,
  defaultPermission: 1
})

// 存储策略列表
const strategies = ref<Array<{id: number, name: string, introduction: string, key: string}>>([])
const loading = ref(false)

// 获取用户信息
const fetchUserInfo = async () => {
  try {
    loading.value = true
    const response = await userAPI.getUserInfo()
    if (response.data.status) {
      const { user, config, strategies: userStrategies } = response.data.data
      
      // 更新表单数据
      form.account = user.account
      form.nickname = user.name
      form.defaultStrategy = config.default_strategy.toString()
      form.autoClear = config.is_auto_clear_preview === 1 ? 1 : 0
      form.pasteAction = config.pasted_action === 1 ? 1 : 0
      form.defaultPermission = config.default_permission === 1 ? 1 : 0
      
      // 更新存储策略列表
      strategies.value = userStrategies
    }
  } catch (error) {
    console.error('获取用户信息失败:', error)
  } finally {
    loading.value = false
  }
}

// 提交处理
const handleSubmit = async () => {
  try {
    loading.value = true
    
    // 转换表单数据为API期望的格式
    const apiData: any = {
      name: form.nickname,
      default_strategy: parseInt(form.defaultStrategy) || 1,
      is_auto_clear_preview: form.autoClear,
      pasted_action: form.pasteAction,
      default_permission: form.defaultPermission
    }
    
    // 只有当密码不为空时才包含密码字段
    if (form.password.trim()) {
      apiData.password = form.password
    }
    
    console.log('保存设置:', apiData)
    
    // 调用API保存用户配置
    const response = await userAPI.updateUserInfo(apiData)
    if (response.data.status) {
      // 保存成功后重新获取用户信息
      await fetchUserInfo()
      // 清空密码字段
      form.password = ''
      // 这里可以添加成功提示
      console.log('设置保存成功')
    }
  } catch (error) {
    console.error('保存设置失败:', error)
  } finally {
    loading.value = false
  }
}

// 组件挂载时获取用户信息
onMounted(() => {
  fetchUserInfo()
})
</script>

<style scoped>
.basic-settings {
  max-width: 1000px;
  margin: 0 auto;
}

.page-header {
  margin-bottom: 32px;
}

.settings-form {
  background-color: white;
  border-radius: 12px;
  padding: 32px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.05);
}

.form-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 32px;
  margin-bottom: 32px;
}

.form-column {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.form-label {
  font-weight: 600;
  color: #374151;
  font-size: 14px;
}

.form-input,
.form-select {
  padding: 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 14px;
  background-color: #f9fafb;
  transition: all 0.2s;
}

.form-input:focus,
.form-select:focus {
  outline: none;
  border-color: #3b82f6;
  background-color: white;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.form-input:disabled,
.form-select:disabled {
  background-color: #f3f4f6;
  color: #9ca3af;
  cursor: not-allowed;
}

.radio-group {
  display: flex;
  gap: 24px;
}

.radio-item {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  font-size: 14px;
  color: #374151;
}

.radio-item input[type="radio"] {
  width: 16px;
  height: 16px;
  accent-color: #3b82f6;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  padding-top: 24px;
  border-top: 1px solid #f3f4f6;
}

.save-btn {
  background-color: #3b82f6;
  color: white;
  padding: 12px 24px;
  border: none;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: background-color 0.2s;
  display: flex;
  align-items: center;
}

.save-btn:hover {
  background-color: #2563eb;
}

.save-btn:disabled {
  background-color: #9ca3af;
  cursor: not-allowed;
}

.save-btn:disabled:hover {
  background-color: #9ca3af;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .form-grid {
    grid-template-columns: 1fr;
    gap: 24px;
  }
  
  .settings-form {
    padding: 24px;
  }
  
  .radio-group {
    flex-direction: column;
    gap: 12px;
  }
  
  .form-actions {
    justify-content: center;
  }
  
  .save-btn {
    width: 100%;
    justify-content: center;
  }
}
</style> 