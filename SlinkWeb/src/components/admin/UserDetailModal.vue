<template>
  <div class="modal-overlay" @click="handleClose">
    <div class="modal-content" @click.stop>
      <!-- 关闭按钮 -->
      <button class="close-btn" @click="handleClose">
        <i class="fa fa-times"></i>
      </button>
      
      <!-- 用户头像 -->
      <div class="user-avatar">
        <i class="fa fa-user"></i>
      </div>
      
      <!-- 用户信息 -->
      <div class="user-info">
        <div class="info-item">
          <span class="label">用户ID:</span>
          <span class="value">{{ user.id }}</span>
        </div>
        <div class="info-item">
          <span class="label">用户名:</span>
          <span class="value">{{ user.name }}</span>
        </div>
        <div class="info-item">
          <span class="label">登录账号:</span>
          <span class="value">{{ user.account }}</span>
        </div>
        <div class="info-item">
          <span class="label">管理员:</span>
          <span class="value">
            <span class="badge" :class="user.is_admin ? 'badge-admin' : 'badge-user'">
              {{ user.is_admin ? '是' : '否' }}
            </span>
          </span>
        </div>
        <div class="info-item">
          <span class="label">用户组ID:</span>
          <span class="value">{{ user.group_id }}</span>
        </div>
        <div class="info-item">
          <span class="label">图片数量:</span>
          <span class="value">{{ user.image_nums }}</span>
        </div>
        <div class="info-item">
          <span class="label">容量限制:</span>
          <span class="value">{{ formatCapacity(user.capacity) }}</span>
        </div>
        <div class="info-item">
          <span class="label">已使用:</span>
          <span class="value">{{ formatCapacity(user.used_size || 0) }}</span>
        </div>
        <div class="info-item" v-if="user.registered_ip">
          <span class="label">注册IP:</span>
          <span class="value">{{ user.registered_ip }}</span>
        </div>
        <div class="info-item" v-if="user.created_at">
          <span class="label">注册时间:</span>
          <span class="value">{{ formatDate(user.created_at) }}</span>
        </div>
        <div class="info-item" v-if="user.updated_at">
          <span class="label">更新时间:</span>
          <span class="value">{{ formatDate(user.updated_at) }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { defineProps, defineEmits } from 'vue'

interface User {
  id: number
  name: string
  account: string
  is_admin: boolean
  group_id: number
  image_nums: number
  capacity: number
  used_size?: number
  registered_ip?: string
  created_at?: string
  updated_at?: string
}

const props = defineProps<{
  user: User
}>()

const emit = defineEmits<{
  close: []
}>()

const handleClose = () => {
  emit('close')
}

// capacity 单位为字节，0 表示无限制
const formatCapacity = (capacity: number) => {
  if (!capacity) return '无限制'
  const mb = capacity / (1024 * 1024)
  if (mb < 1024) return `${mb.toFixed(2)} MB`
  return `${(mb / 1024).toFixed(2)} GB`
}

const formatDate = (dateStr: string) => {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
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

.user-avatar {
  display: flex;
  justify-content: center;
  margin-bottom: 24px;
}

.user-avatar i {
  width: 80px;
  height: 80px;
  background-color: #f3f4f6;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 32px;
  color: #9ca3af;
}

.user-info {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.info-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 0;
  border-bottom: 1px solid #f3f4f6;
}

.info-item:last-child {
  border-bottom: none;
}

.label {
  font-weight: 500;
  color: #374151;
  font-size: 14px;
}

.value {
  color: #6b7280;
  font-size: 14px;
  text-align: right;
  word-break: break-all;
}

.badge {
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
}

.badge-admin {
  background-color: #dc2626;
  color: white;
}

.badge-user {
  background-color: #3b82f6;
  color: white;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .modal-content {
    padding: 24px;
    margin: 16px;
  }
  
  .info-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 4px;
  }
  
  .value {
    text-align: left;
  }
}
</style> 