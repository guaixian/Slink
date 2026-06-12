<template>
  <div class="page-layout">
    <!-- 页面标题区域 -->
    <div class="page-header">
      <div class="header-content">
        <div class="title-section">
          <h1 class="page-title">
            <i :class="iconClass" class="title-icon"></i>
            {{ title }}
          </h1>
          <p class="page-description">{{ description }}</p>
        </div>
        <div class="header-actions" v-if="$slots.actions">
          <slot name="actions"></slot>
        </div>
      </div>
    </div>

    <!-- 页面内容区域 -->
    <div class="page-content">
      <slot></slot>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  title: string
  description: string
  iconClass: string
}>()
</script>

<style scoped>
.page-layout {
  width: 100%;
  max-width: 1400px;
  margin: 0 auto;
  padding: 0;
}

.page-header {
  margin-bottom: 32px;
  padding-bottom: 24px;
  border-bottom: 1px solid #e5e7eb;
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 24px;
}

.title-section {
  flex: 1;
}

.page-title {
  font-size: 28px;
  font-weight: 700;
  color: #111827;
  margin: 0 0 8px 0;
  display: flex;
  align-items: center;
  line-height: 1.2;
}

.title-icon {
  margin-right: 12px;
  font-size: 24px;
  color: #3b82f6;
}

.page-description {
  font-size: 16px;
  color: #6b7280;
  margin: 0;
  line-height: 1.5;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.page-content {
  min-height: calc(100vh - 200px);
}

/* 响应式设计 */
@media (max-width: 768px) {
  .page-layout {
    padding: 0 16px;
  }
  
  .header-content {
    flex-direction: column;
    gap: 16px;
  }
  
  .page-title {
    font-size: 24px;
  }
  
  .title-icon {
    font-size: 20px;
  }
  
  .page-description {
    font-size: 14px;
  }
}
</style> 