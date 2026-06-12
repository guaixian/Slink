<template>
  <Teleport to="body">
    <Transition name="confirm-fade">
      <div v-if="visible" class="confirm-overlay" @click.self="handleCancel">
        <Transition name="confirm-scale">
          <div v-if="visible" class="confirm-dialog">
            <div class="confirm-header">
              <div class="confirm-icon" :class="[`icon-${type}`]">
                <i :class="iconClass"></i>
              </div>
              <h3 class="confirm-title">{{ title }}</h3>
            </div>
            <div class="confirm-body">
              <p class="confirm-message">{{ message }}</p>
            </div>
            <div class="confirm-footer">
              <button class="btn btn-cancel" @click="handleCancel">
                {{ cancelText }}
              </button>
              <button class="btn btn-confirm" :class="[`btn-${type}`]" @click="handleConfirm">
                {{ confirmText }}
              </button>
            </div>
          </div>
        </Transition>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'

interface Props {
  title?: string
  message: string
  type?: 'warning' | 'danger' | 'info'
  confirmText?: string
  cancelText?: string
  onConfirm?: () => void
  onCancel?: () => void
}

const props = withDefaults(defineProps<Props>(), {
  title: '确认操作',
  type: 'warning',
  confirmText: '确定',
  cancelText: '取消'
})

const visible = ref(false)

const iconClass = computed(() => {
  const iconMap: Record<typeof props.type, string> = {
    warning: 'fa fa-exclamation-triangle',
    danger: 'fa fa-trash',
    info: 'fa fa-question-circle'
  }
  return iconMap[props.type]
})

const handleConfirm = () => {
  visible.value = false
  setTimeout(() => {
    props.onConfirm?.()
  }, 200)
}

const handleCancel = () => {
  visible.value = false
  setTimeout(() => {
    props.onCancel?.()
  }, 200)
}

onMounted(() => {
  visible.value = true
})
</script>

<style scoped>
.confirm-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 10001;
}

.confirm-dialog {
  background: white;
  border-radius: 12px;
  width: 90%;
  max-width: 400px;
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
  overflow: hidden;
}

.confirm-header {
  padding: 24px 24px 16px;
  text-align: center;
}

.confirm-icon {
  width: 56px;
  height: 56px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 16px;
  font-size: 24px;
}

.confirm-icon.icon-warning {
  background-color: #fef3c7;
  color: #f59e0b;
}

.confirm-icon.icon-danger {
  background-color: #fee2e2;
  color: #ef4444;
}

.confirm-icon.icon-info {
  background-color: #dbeafe;
  color: #3b82f6;
}

.confirm-title {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #1f2937;
}

.confirm-body {
  padding: 0 24px 24px;
  text-align: center;
}

.confirm-message {
  margin: 0;
  font-size: 14px;
  color: #6b7280;
  line-height: 1.6;
}

.confirm-footer {
  display: flex;
  gap: 12px;
  padding: 16px 24px;
  background-color: #f9fafb;
  border-top: 1px solid #e5e7eb;
}

.btn {
  flex: 1;
  padding: 10px 16px;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  border: none;
}

.btn-cancel {
  background-color: white;
  border: 1px solid #d1d5db;
  color: #374151;
}

.btn-cancel:hover {
  background-color: #f3f4f6;
}

.btn-confirm {
  color: white;
}

.btn-confirm.btn-warning {
  background-color: #f59e0b;
}

.btn-confirm.btn-warning:hover {
  background-color: #d97706;
}

.btn-confirm.btn-danger {
  background-color: #ef4444;
}

.btn-confirm.btn-danger:hover {
  background-color: #dc2626;
}

.btn-confirm.btn-info {
  background-color: #3b82f6;
}

.btn-confirm.btn-info:hover {
  background-color: #2563eb;
}

/* Animations */
.confirm-fade-enter-active,
.confirm-fade-leave-active {
  transition: opacity 0.2s ease;
}

.confirm-fade-enter-from,
.confirm-fade-leave-to {
  opacity: 0;
}

.confirm-scale-enter-active,
.confirm-scale-leave-active {
  transition: all 0.2s ease;
}

.confirm-scale-enter-from {
  opacity: 0;
  transform: scale(0.9);
}

.confirm-scale-leave-to {
  opacity: 0;
  transform: scale(0.9);
}
</style>
