<template>
  <Teleport to="body">
    <Transition name="toast-slide">
      <div v-if="visible" class="toast-container" :class="[`toast-${type}`]">
        <div class="toast-content">
          <div class="toast-icon">
            <i :class="iconClass"></i>
          </div>
          <div class="toast-message">{{ message }}</div>
          <button class="toast-close" @click="close">
            <i class="fa fa-times"></i>
          </button>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'

interface Props {
  message: string
  type?: 'success' | 'error' | 'warning' | 'info'
  duration?: number
  onClose?: () => void
}

const props = withDefaults(defineProps<Props>(), {
  type: 'info',
  duration: 3000
})

const visible = ref(false)
let timer: ReturnType<typeof setTimeout> | null = null

const iconClass = computed(() => {
  const iconMap: Record<typeof props.type, string> = {
    success: 'fa fa-check-circle',
    error: 'fa fa-times-circle',
    warning: 'fa fa-exclamation-triangle',
    info: 'fa fa-info-circle'
  }
  return iconMap[props.type]
})

const close = () => {
  visible.value = false
  if (timer) {
    clearTimeout(timer)
    timer = null
  }
  setTimeout(() => {
    props.onClose?.()
  }, 300)
}

onMounted(() => {
  visible.value = true
  if (props.duration > 0) {
    timer = setTimeout(() => {
      close()
    }, props.duration)
  }
})

onUnmounted(() => {
  if (timer) {
    clearTimeout(timer)
  }
})
</script>

<style scoped>
.toast-container {
  position: fixed;
  top: 20px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 10000;
  min-width: 300px;
  max-width: 500px;
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  overflow: hidden;
}

.toast-content {
  display: flex;
  align-items: center;
  padding: 14px 16px;
  gap: 12px;
}

.toast-icon {
  flex-shrink: 0;
  font-size: 20px;
}

.toast-message {
  flex: 1;
  font-size: 14px;
  line-height: 1.5;
  word-break: break-word;
}

.toast-close {
  flex-shrink: 0;
  background: none;
  border: none;
  cursor: pointer;
  padding: 4px;
  opacity: 0.6;
  transition: opacity 0.2s;
  font-size: 14px;
}

.toast-close:hover {
  opacity: 1;
}

/* Success */
.toast-success {
  background-color: #f0fdf4;
  border: 1px solid #86efac;
}

.toast-success .toast-icon {
  color: #22c55e;
}

.toast-success .toast-message {
  color: #166534;
}

.toast-success .toast-close {
  color: #166534;
}

/* Error */
.toast-error {
  background-color: #fef2f2;
  border: 1px solid #fca5a5;
}

.toast-error .toast-icon {
  color: #ef4444;
}

.toast-error .toast-message {
  color: #991b1b;
}

.toast-error .toast-close {
  color: #991b1b;
}

/* Warning */
.toast-warning {
  background-color: #fffbeb;
  border: 1px solid #fcd34d;
}

.toast-warning .toast-icon {
  color: #f59e0b;
}

.toast-warning .toast-message {
  color: #92400e;
}

.toast-warning .toast-close {
  color: #92400e;
}

/* Info */
.toast-info {
  background-color: #eff6ff;
  border: 1px solid #93c5fd;
}

.toast-info .toast-icon {
  color: #3b82f6;
}

.toast-info .toast-message {
  color: #1e40af;
}

.toast-info .toast-close {
  color: #1e40af;
}

/* Animation */
.toast-slide-enter-active,
.toast-slide-leave-active {
  transition: all 0.3s ease;
}

.toast-slide-enter-from {
  opacity: 0;
  transform: translateX(-50%) translateY(-20px);
}

.toast-slide-leave-to {
  opacity: 0;
  transform: translateX(-50%) translateY(-20px);
}
</style>
