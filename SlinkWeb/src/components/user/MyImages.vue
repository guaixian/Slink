<template>
  <PageLayout 
    title="我的图片"
    description="查看和管理您上传的所有图片"
    icon-class="fa fa-images"
  >

    <!-- 操作栏 -->
    <div class="action-bar">
      <div class="search-box">
        <input 
          type="text" 
          placeholder="搜索图片名称..."
          class="search-input"
          v-model="searchKeyword"
          @keyup.enter="handleSearch"
        >
        <i class="fa fa-search search-icon"></i>
      </div>
      <div class="loading-indicator" v-if="loading">
        <i class="fa fa-spinner fa-spin"></i>
        加载中...
      </div>
    </div>

         <!-- 调试信息 -->
     <div v-if="images.length === 0 && !loading" class="debug-info">
       <p>没有图片数据</p>
       <p>总图片数: {{ totalImages }}</p>
       <p>当前页: {{ currentPage }}</p>
     </div>
     
     <!-- 图片网格 -->
     <div class="images-grid">
       <div 
         v-for="image in filteredImages" 
         :key="image.id" 
         class="image-card"
         @contextmenu.prevent="showContextMenu($event, image)"
         @click="handleImageClick(image)"
       >
                 <div class="image-container">
           <img :src="fixImageUrl(image.links.url)" :alt="image.origin_name" class="image-thumbnail">
          <div class="image-overlay">
            <div class="overlay-actions">
              <button class="action-btn" @click.stop="handleView(image)" title="查看原图">
                <i class="fa fa-eye"></i>
              </button>
              <button class="action-btn" @click.stop="handleCopyLink(image, 'url')" title="复制链接">
                <i class="fa fa-link"></i>
              </button>
            </div>
          </div>
          <button class="delete-btn" @click.stop="handleDelete(image)" title="删除图片">
            <i class="fa fa-trash"></i>
          </button>
        </div>
        <div class="image-info">
          <div class="image-name">{{ image.origin_name }}</div>
          <div class="image-meta">
            <span class="image-size">{{ formatFileSize(image.size) }}</span>
            <span class="image-type">{{ image.mimetype }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 右键菜单 -->
    <div v-if="contextMenu.show" class="context-menu" :style="{ left: contextMenu.x + 'px', top: contextMenu.y + 'px' }">
      <div class="menu-item" @click="handleView(contextMenu.image)">
        <i class="fa fa-eye"></i>
        查看原图
      </div>
      <div class="menu-item" @click="handleCopyLink(contextMenu.image, 'url')">
        <i class="fa fa-link"></i>
        复制链接
      </div>
      <div class="menu-item" @click="handleCopyLink(contextMenu.image, 'html')">
        <i class="fa fa-code"></i>
        复制HTML
      </div>
      <div class="menu-item" @click="handleCopyLink(contextMenu.image, 'markdown')">
        <i class="fa fa-file-text"></i>
        复制Markdown
      </div>
      <div class="menu-item" @click="handleCopyLink(contextMenu.image, 'bbcode')">
        <i class="fa fa-brackets"></i>
        复制BBCode
      </div>
      <div class="menu-separator"></div>
      <div class="menu-item delete" @click="handleDelete(contextMenu.image)">
        <i class="fa fa-trash"></i>
        删除图片
      </div>
    </div>

    <!-- 分页 -->
    <div class="pagination">
      <button class="page-btn" :disabled="currentPage === 1" @click="prevPage">
        <i class="fa fa-chevron-left"></i>
      </button>
      <span class="page-info">第 {{ currentPage }} 页，共 {{ totalPages }} 页</span>
      <button class="page-btn" :disabled="currentPage === totalPages" @click="nextPage">
        <i class="fa fa-chevron-right"></i>
      </button>
    </div>
  </PageLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { imageAPI } from '../../api'
import PageLayout from '../admin/PageLayout.vue'
import { useMessage } from '../../composables/useMessage'

// 响应式数据
const searchKeyword = ref('')
const selectedFilter = ref('all')
const currentPage = ref(1)
const totalPages = ref(1)
const loading = ref(false)

const { toast, confirm } = useMessage()

// 图片数据
const images = ref<any[]>([])
const totalImages = ref(0)

// 右键菜单
const contextMenu = ref({
  show: false,
  x: 0,
  y: 0,
  image: null as any
})

// 过滤后的图片
const filteredImages = computed(() => {
  let filtered = images.value

  // 按搜索关键词过滤
  if (searchKeyword.value) {
    filtered = filtered.filter(img => 
      img.origin_name.toLowerCase().includes(searchKeyword.value.toLowerCase())
    )
  }

  return filtered
})

// 格式化文件大小
const formatFileSize = (size: number) => {
  if (size < 1) {
    return (size * 1024).toFixed(2) + ' KB'
  }
  return size.toFixed(2) + ' MB'
}

// 修复URL中的反斜杠问题
const fixImageUrl = (url: string): string => {
  if (!url) return ''
  // 将反斜杠替换为正斜杠，确保URL格式正确
  return url.replace(/\\/g, '/')
}

// 加载图片列表
const loadImages = async () => {
  loading.value = true
  try {
    const response = await imageAPI.getImageList(currentPage.value, 20)
    console.log('API响应:', response.data)
    if (response.data.status) {
      // API直接返回图片数组，而不是包含images字段的对象
      images.value = response.data.data
      totalImages.value = response.data.data.length
      totalPages.value = Math.ceil(totalImages.value / 20)
      console.log('图片数据:', images.value)
      if (images.value.length > 0) {
        console.log('第一张图片:', images.value[0])
        console.log('图片URL:', images.value[0].links?.url)
      }
    }
  } catch (error) {
    console.error('加载图片列表失败:', error)
  } finally {
    loading.value = false
  }
}

// 显示右键菜单
const showContextMenu = (event: MouseEvent, image: any) => {
  event.preventDefault()
  contextMenu.value = {
    show: true,
    x: event.clientX,
    y: event.clientY,
    image
  }
}

// 隐藏右键菜单
const hideContextMenu = () => {
  contextMenu.value.show = false
}

// 页面初始化
onMounted(() => {
  loadImages()
  
  // 添加全局点击事件来隐藏右键菜单
  document.addEventListener('click', hideContextMenu)
})

// 搜索处理
const handleSearch = () => {
  // 搜索功能通过computed自动过滤
  console.log('搜索关键词:', searchKeyword.value)
}

// 图片点击处理
const handleImageClick = (image: any) => {
  console.log('点击图片:', image)
}

// 查看图片
const handleView = (image: any) => {
  window.open(fixImageUrl(image.links.url), '_blank')
}

// 复制链接
const handleCopyLink = (image: any, linkType: string = 'url') => {
  const link = fixImageUrl(image.links[linkType] || image.links.url)
  navigator.clipboard.writeText(link).then(() => {
    toast.success('链接已复制到剪贴板！')
  }).catch(() => {
    // 降级方案
    const textArea = document.createElement('textarea')
    textArea.value = link
    document.body.appendChild(textArea)
    textArea.select()
    document.execCommand('copy')
    document.body.removeChild(textArea)
    toast.success('链接已复制到剪贴板！')
  })
}

// 删除图片
const handleDelete = async (image: any) => {
  const confirmed = await confirm.danger(`确定要删除图片 "${image.origin_name}" 吗？此操作不可恢复。`, '删除图片')
  if (!confirmed) {
    return
  }

  try {
    const response = await imageAPI.deleteImage(image.id)
    if (response.data.status) {
      toast.success('删除成功！')
      loadImages() // 重新加载列表
    } else {
      toast.error(`删除失败：${response.data.message}`)
    }
  } catch (error: any) {
    console.error('删除图片失败:', error)
    toast.error(`删除失败，请重试：${error.response?.data?.message || error.message}`)
  }
}

// 分页处理
const prevPage = () => {
  if (currentPage.value > 1) {
    currentPage.value--
    loadImages()
  }
}

const nextPage = () => {
  if (currentPage.value < totalPages.value) {
    currentPage.value++
    loadImages()
  }
}
</script>

<style scoped>
.my-images {
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

/* 图片网格 */
.images-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
  gap: 20px;
  margin-bottom: 32px;
}

.image-card {
  background-color: white;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.05);
  overflow: hidden;
  cursor: pointer;
  transition: all 0.2s;
}

.image-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0,0,0,0.1);
}

.image-container {
  position: relative;
  height: 150px;
  overflow: hidden;
}

.image-thumbnail {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.image-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0,0,0,0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 0;
  transition: opacity 0.2s;
}

.image-card:hover .image-overlay {
  opacity: 1;
}

.overlay-actions {
  display: flex;
  gap: 8px;
}

.action-btn {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background-color: white;
  border: none;
  color: #374151;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.action-btn:hover {
  background-color: #3b82f6;
  color: white;
}

.delete-btn {
  position: absolute;
  top: 8px;
  right: 8px;
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background-color: rgba(239, 68, 68, 0.9);
  color: white;
  border: none;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
  opacity: 0;
}

.image-card:hover .delete-btn {
  opacity: 1;
}

.delete-btn:hover {
  background-color: #dc2626;
  transform: scale(1.1);
}

/* 右键菜单 */
.context-menu {
  position: fixed;
  background-color: white;
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0,0,0,0.15);
  border: 1px solid #e5e7eb;
  z-index: 1000;
  min-width: 160px;
  padding: 8px 0;
}

.menu-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  cursor: pointer;
  font-size: 14px;
  color: #374151;
  transition: background-color 0.2s;
}

.menu-item:hover {
  background-color: #f9fafb;
}

.menu-item.delete {
  color: #dc2626;
}

.menu-item.delete:hover {
  background-color: #fef2f2;
}

.menu-separator {
  height: 1px;
  background-color: #e5e7eb;
  margin: 4px 0;
}

/* 加载指示器 */
.loading-indicator {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #6b7280;
  font-size: 14px;
}

/* 调试信息 */
.debug-info {
  background-color: #fef3c7;
  border: 1px solid #f59e0b;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 24px;
  color: #92400e;
}

.debug-info p {
  margin: 4px 0;
  font-size: 14px;
}

.image-info {
  padding: 12px;
}

.image-name {
  font-size: 14px;
  font-weight: 500;
  color: #374151;
  margin-bottom: 4px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.image-meta {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  color: #6b7280;
  margin-bottom: 8px;
}



/* 分页 */
.pagination {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 16px;
}

.page-btn {
  width: 36px;
  height: 36px;
  border-radius: 6px;
  background-color: white;
  border: 1px solid #d1d5db;
  color: #374151;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.page-btn:hover:not(:disabled) {
  background-color: #3b82f6;
  color: white;
  border-color: #3b82f6;
}

.page-btn:disabled {
  background-color: #f9fafb;
  color: #9ca3af;
  cursor: not-allowed;
}

.page-info {
  font-size: 14px;
  color: #6b7280;
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
  
  .images-grid {
    grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
    gap: 16px;
  }
  
  .image-container {
    height: 120px;
  }
}
</style> 
