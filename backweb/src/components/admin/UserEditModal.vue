<template>
  <div class="modal-overlay" @click="handleClose">
    <div class="modal-content" @click.stop>
      <!-- 关闭按钮 -->
      <button class="close-btn" @click="handleClose">
        <i class="fa fa-times"></i>
      </button>
      
      <!-- 编辑表单 -->
      <div class="edit-form">
        <div class="form-group">
          <label>选择角色组</label>
          <select v-model="formData.roleGroup" class="form-input">
            <option value="系统默认组&游客组">系统默认组&游客组</option>
            <option value="管理员组">管理员组</option>
            <option value="普通用户组">普通用户组</option>
          </select>
        </div>
        
        <div class="form-group">
          <label>登录账号</label>
          <input type="text" v-model="formData.account" class="form-input" placeholder="登录账号" />
        </div>
        
        <div class="form-group">
          <label>*用户名</label>
          <input type="text" v-model="formData.username" class="form-input" />
        </div>
        
        <div class="form-group">
          <label>*总容量(kb)</label>
          <input type="number" v-model="formData.totalCapacity" class="form-input" />
        </div>
        
        <div class="form-group">
          <label>新密码</label>
          <input type="password" v-model="formData.newPassword" class="form-input" placeholder="不修改请留空" />
        </div>
        
        <div class="form-group">
          <label>账号状态</label>
          <p class="status-description">冻结账号后将无法登录系统</p>
          <div class="status-toggle">
            <label class="toggle-item" :class="{ active: formData.status === '正常' }">
              <input type="radio" v-model="formData.status" value="正常" />
              <span class="toggle-text">正常</span>
            </label>
            <label class="toggle-item" :class="{ active: formData.status === '冻结' }">
              <input type="radio" v-model="formData.status" value="冻结" />
              <span class="toggle-text">冻结</span>
            </label>
          </div>
        </div>
      </div>
      
      <!-- 底部按钮 -->
      <div class="modal-footer">
        <button class="btn btn-secondary" @click="handleClose">取消</button>
        <button class="btn btn-primary" @click="handleSave">确认保存</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, defineProps, defineEmits } from 'vue'

interface User {
  username: string
  account: string
  roleGroup: string
  totalCapacity: string | number
  status: string
}

const props = defineProps<{
  user: User
}>()

const emit = defineEmits<{
  close: []
  save: [user: User]
}>()

// 表单数据
const formData = reactive({
  roleGroup: props.user?.roleGroup || '系统默认组&游客组',
  account: props.user?.account || '',
  username: props.user?.username || '',
  totalCapacity: props.user?.totalCapacity ? 
    (typeof props.user.totalCapacity === 'string' ? 
      parseInt(props.user.totalCapacity.replace(/[^\d]/g, '')) : (props.user.totalCapacity as number)) : 0,
  newPassword: '',
  status: props.user?.status || '正常'
})

const handleClose = () => {
  emit('close')
}

const handleSave = () => {
  const updatedUser = {
    ...(props.user || {}),
    ...formData
  }
  emit('save', updatedUser)
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  background: white;
  border-radius: 12px;
  padding: 32px;
  max-width: 500px;
  width: 90%;
  max-height: 80vh;
  overflow-y: auto;
  position: relative;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.15);
}

.close-btn {
  position: absolute;
  top: 16px;
  right: 16px;
  background: none;
  border: none;
  font-size: 20px;
  color: #6b7280;
  cursor: pointer;
  padding: 8px;
  border-radius: 4px;
  transition: all 0.2s ease;
}

.close-btn:hover {
  background-color: #f3f4f6;
  color: #374151;
}

.edit-form {
  margin-bottom: 24px;
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  margin-bottom: 6px;
  font-weight: 500;
  color: #374151;
  font-size: 14px;
}

.form-input {
  width: 100%;
  padding: 10px 14px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 14px;
  transition: border-color 0.3s ease;
}

.form-input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

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

.toggle-item.active::before {
  content: '';
  position: absolute;
  left: 0;
  top: 0;
  width: 100%;
  height: 100%;
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
  border-radius: 6px;
  opacity: 0.1;
  z-index: 0;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding-top: 20px;
  border-top: 1px solid #e5e7eb;
}

.btn {
  padding: 10px 20px;
  border: none;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
}

.btn-secondary {
  background-color: #f3f4f6;
  color: #374151;
}

.btn-secondary:hover {
  background-color: #e5e7eb;
}

.btn-primary {
  background-color: #3b82f6;
  color: white;
}

.btn-primary:hover {
  background-color: #2563eb;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .modal-content {
    padding: 24px;
    margin: 16px;
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