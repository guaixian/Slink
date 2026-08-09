<template>
  <Teleport to="body">
    <div class="preview-overlay" @click.self="close" @keydown.esc="close">
      <!-- 顶部工具栏 -->
      <div class="preview-toolbar">
        <span class="preview-name" :title="name">{{ name }}</span>
        <button class="preview-close" @click="close" title="关闭 (Esc)">
          <i class="fa fa-times"></i>
        </button>
      </div>

      <!-- 图片主体 -->
      <div class="preview-body" @click.self="close">
        <img
          v-if="url"
          :src="url"
          :alt="name"
          class="preview-image"
          @click.stop
        >
        <div v-else class="preview-empty">图片地址不可用</div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'

defineProps<{
  url: string
  name: string
}>()

const emit = defineEmits<{
  close: []
}>()

const close = () => emit('close')

const onKeydown = (e: KeyboardEvent) => {
  if (e.key === 'Escape') close()
}

onMounted(() => {
  document.addEventListener('keydown', onKeydown)
  document.body.style.overflow = 'hidden'
})

onUnmounted(() => {
  document.removeEventListener('keydown', onKeydown)
  document.body.style.overflow = ''
})
</script>

<style scoped>
.preview-overlay {
  position: fixed;
  inset: 0;
  background-color: rgba(0, 0, 0, 0.85);
  z-index: 2000;
  display: flex;
  flex-direction: column;
}

.preview-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 20px;
  color: #fff;
  flex-shrink: 0;
}

.preview-name {
  font-size: 14px;
  color: #e5e7eb;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 80vw;
}

.preview-close {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  border: none;
  background-color: rgba(255, 255, 255, 0.15);
  color: #fff;
  font-size: 18px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background-color 0.2s;
  flex-shrink: 0;
}

.preview-close:hover {
  background-color: rgba(255, 255, 255, 0.3);
}

.preview-body {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: auto;
  padding: 0 24px 24px;
  cursor: zoom-out;
}

.preview-image {
  max-width: 92vw;
  max-height: calc(100vh - 120px);
  object-fit: contain;
  border-radius: 4px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.5);
  cursor: default;
  background-color: #fff;
}

.preview-empty {
  color: #9ca3af;
  font-size: 14px;
}
</style>
