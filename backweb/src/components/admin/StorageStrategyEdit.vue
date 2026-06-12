<template>
  <div class="storage-edit">
    <!-- 页面标题 -->
    <div class="page-header">
      <div class="header-left">
        <button class="back-btn" @click="goBack">
          <i class="fa fa-arrow-left"></i>
        </button>
        <h1 class="page-title">
          <i class="fa fa-database mr-2"></i>
          {{ isEditing ? '编辑储存策略' : '创建储存策略' }}
        </h1>
      </div>
    </div>

    <!-- 信息提示 -->
    <div class="info-banner">
      <i class="fa fa-info-circle"></i>
      <span>存储策略对所有登录用户可用；上传限额、命名规则、允许后缀等由侧边栏「上传策略」统一配置。</span>
    </div>

    <!-- 编辑表单 -->
    <form @submit.prevent="handleSubmit" class="edit-form">
      <!-- 名称 -->
      <div class="form-group">
        <label class="form-label required">名称</label>
        <input 
          type="text" 
          v-model="formData.name"
          placeholder="请输入策略名称"
          class="form-input"
          :class="{ 'error': errors.name }"
          @blur="validateField('name')"
        >
        <div v-if="errors.name" class="error-message">{{ errors.name }}</div>
      </div>

      <!-- 简介 -->
      <div class="form-group">
        <label class="form-label">简介</label>
        <textarea 
          v-model="formData.description"
          placeholder="请输入简介,可为空"
          class="form-textarea"
          rows="3"
        ></textarea>
      </div>

      <!-- 储存策略 -->
      <div class="form-group">
        <label class="form-label required">储存策略</label>
        <select
          v-model="formData.storageType"
          class="form-select"
          :class="{ 'error': errors.storageType }"
          @change="handleStorageTypeChange"
        >
          <option value="">请选择储存策略</option>
          <option value="local">本地</option>
          <option value="ftp">FTP</option>
          <option value="sftp">SFTP</option>
          <option value="s3">AWS S3</option>
          <option value="minio">MinIO</option>
          <option value="oss">阿里云OSS</option>
          <option value="cos">腾讯云COS</option>
          <option value="qiniu">七牛云</option>
          <option value="upyun">又拍云</option>
          <option value="webdav">WebDAV</option>
        </select>
        <div v-if="errors.storageType" class="error-message">{{ errors.storageType }}</div>
      </div>

      <!-- 访问网址 -->
      <div class="form-group">
        <label class="form-label required">访问网址</label>
        <input
          type="url"
          v-model="formData.accessUrl"
          placeholder="请输入图片访问域名,需要加 http(s)://"
          class="form-input"
          :class="{ 'error': errors.accessUrl }"
          @blur="validateField('accessUrl'); handleAccessUrlChange()"
        >
        <div v-if="errors.accessUrl" class="error-message">{{ errors.accessUrl }}</div>
        <div v-if="formData.storageType === 'local'" class="warning-message">
          <i class="fa fa-exclamation-triangle"></i>
          <span>访问网址用于生成外链；其 URL 中的路径为「对外访问根」——须与实际一致（如本程序通过 <code class="mono-inline">/static/…</code> 读图，则请填 <code class="mono-inline">http://主机:端口/static</code> 等）。与下方「本机子目录」不是同一概念。</span>
        </div>
      </div>

      <!-- URL Queries -->
      <div class="form-group">
        <label class="form-label">URL Queries</label>
        <input
          type="text"
          v-model="formData.urlQueries"
          placeholder="请输入 url 额外参数"
          class="form-input"
        >
      </div>

      <div v-if="formData.storageType === 'local'" class="form-group">
        <label class="form-label">访问根路径 (URL)</label>
        <input
          type="text"
          v-model="formData.customPath"
          placeholder="例如: / 或 /static，需与「访问网址」中路径一致"
          class="form-input"
        >
        <p class="form-hint">与访问网址里的路径段一致，用于 root 字段。勿把此处当成磁盘上的文件夹。</p>
      </div>
      <div v-if="formData.storageType === 'local'" class="form-group">
        <label class="form-label">本机子目录 (磁盘)</label>
        <input
          type="text"
          v-model="formData.localDiskPath"
          placeholder="static"
          class="form-input"
        >
        <div class="warning-message">
          <i class="fa fa-exclamation-triangle"></i>
          <span>相对「程序工作目录」的相对路径，默认 <code class="mono-inline">static</code>（与接口 <code class="mono-inline">/static/…</code> 及 Docker 中挂载的 <code class="mono-inline">…/static</code> 对应）。勿填 <code class="mono-inline">/</code> 或把 URL 根写在这里。可填绝对路径（需有写权限），一般保持默认即可。</span>
        </div>
      </div>

      <!-- FTP/SFTP 配置 -->
      <template v-if="formData.storageType === 'ftp' || formData.storageType === 'sftp'">
        <div class="form-group">
          <label class="form-label required">服务器地址</label>
          <input
            type="text"
            v-model="formData.ftpHost"
            placeholder="例如: ftp.example.com"
            class="form-input"
          >
        </div>
        <div class="form-group">
          <label class="form-label required">端口</label>
          <input
            type="number"
            v-model="formData.ftpPort"
            :placeholder="formData.storageType === 'ftp' ? '21' : '22'"
            class="form-input"
          >
        </div>
        <div class="form-group">
          <label class="form-label required">用户名</label>
          <input
            type="text"
            v-model="formData.ftpUsername"
            placeholder="请输入用户名"
            class="form-input"
          >
        </div>
        <div class="form-group">
          <label class="form-label required">密码</label>
          <input
            type="password"
            v-model="formData.ftpPassword"
            placeholder="请输入密码"
            class="form-input"
          >
        </div>
        <div class="form-group">
          <label class="form-label">远程路径</label>
          <input
            type="text"
            v-model="formData.ftpPath"
            placeholder="例如: /uploads"
            class="form-input"
          >
        </div>
      </template>

      <!-- WebDAV 配置 -->
      <template v-if="formData.storageType === 'webdav'">
        <div class="form-group">
          <label class="form-label required">WebDAV 服务器地址</label>
          <input
            type="url"
            v-model="formData.webdavUrl"
            placeholder="例如: https://dav.example.com"
            class="form-input"
          >
        </div>
        <div class="form-group">
          <label class="form-label required">用户名</label>
          <input
            type="text"
            v-model="formData.webdavUsername"
            placeholder="请输入用户名"
            class="form-input"
          >
        </div>
        <div class="form-group">
          <label class="form-label required">密码</label>
          <input
            type="password"
            v-model="formData.webdavPassword"
            placeholder="请输入密码"
            class="form-input"
          >
        </div>
        <div class="form-group">
          <label class="form-label">基础路径</label>
          <input
            type="text"
            v-model="formData.webdavPath"
            placeholder="例如: /images"
            class="form-input"
          >
        </div>
      </template>

      <!-- 阿里云 OSS 配置 -->
      <template v-if="formData.storageType === 'oss'">
        <div class="form-group">
          <label class="form-label required">AccessKey ID</label>
          <input
            type="text"
            v-model="formData.ossAccessKeyId"
            placeholder="请输入 AccessKey ID"
            class="form-input"
          >
        </div>
        <div class="form-group">
          <label class="form-label required">AccessKey Secret</label>
          <input
            type="password"
            v-model="formData.ossAccessKeySecret"
            placeholder="请输入 AccessKey Secret"
            class="form-input"
          >
        </div>
        <div class="form-group">
          <label class="form-label required">Bucket 名称</label>
          <input
            type="text"
            v-model="formData.ossBucket"
            placeholder="请输入 Bucket 名称"
            class="form-input"
          >
        </div>
        <div class="form-group">
          <label class="form-label required">Endpoint</label>
          <input
            type="text"
            v-model="formData.ossEndpoint"
            placeholder="例如: oss-cn-hangzhou.aliyuncs.com"
            class="form-input"
          >
        </div>
        <div class="form-group">
          <label class="form-label">存储路径</label>
          <input
            type="text"
            v-model="formData.ossPath"
            placeholder="例如: images/"
            class="form-input"
          >
        </div>
      </template>

      <!-- 腾讯云 COS 配置 -->
      <template v-if="formData.storageType === 'cos'">
        <div class="form-group">
          <label class="form-label required">SecretId</label>
          <input
            type="text"
            v-model="formData.cosSecretId"
            placeholder="请输入 SecretId"
            class="form-input"
          >
        </div>
        <div class="form-group">
          <label class="form-label required">SecretKey</label>
          <input
            type="password"
            v-model="formData.cosSecretKey"
            placeholder="请输入 SecretKey"
            class="form-input"
          >
        </div>
        <div class="form-group">
          <label class="form-label required">Bucket 名称</label>
          <input
            type="text"
            v-model="formData.cosBucket"
            placeholder="例如: mybucket-1250000000"
            class="form-input"
          >
        </div>
        <div class="form-group">
          <label class="form-label required">地域</label>
          <input
            type="text"
            v-model="formData.cosRegion"
            placeholder="例如: ap-guangzhou"
            class="form-input"
          >
        </div>
        <div class="form-group">
          <label class="form-label">存储路径</label>
          <input
            type="text"
            v-model="formData.cosPath"
            placeholder="例如: images/"
            class="form-input"
          >
        </div>
      </template>

      <!-- 七牛云配置 -->
      <template v-if="formData.storageType === 'qiniu'">
        <div class="form-group">
          <label class="form-label required">AccessKey</label>
          <input
            type="text"
            v-model="formData.qiniuAccessKey"
            placeholder="请输入 AccessKey"
            class="form-input"
          >
        </div>
        <div class="form-group">
          <label class="form-label required">SecretKey</label>
          <input
            type="password"
            v-model="formData.qiniuSecretKey"
            placeholder="请输入 SecretKey"
            class="form-input"
          >
        </div>
        <div class="form-group">
          <label class="form-label required">Bucket 名称</label>
          <input
            type="text"
            v-model="formData.qiniuBucket"
            placeholder="请输入 Bucket 名称"
            class="form-input"
          >
        </div>
        <div class="form-group">
          <label class="form-label required">存储区域</label>
          <select v-model="formData.qiniuRegion" class="form-select">
            <option value="">请选择存储区域</option>
            <option value="z0">华东-浙江</option>
            <option value="z1">华北-河北</option>
            <option value="z2">华南-广东</option>
            <option value="na0">北美-洛杉矶</option>
            <option value="as0">亚太-新加坡</option>
          </select>
        </div>
        <div class="form-group">
          <label class="form-label">存储路径</label>
          <input
            type="text"
            v-model="formData.qiniuPath"
            placeholder="例如: images/"
            class="form-input"
          >
        </div>
      </template>

      <!-- AWS S3 配置 -->
      <template v-if="formData.storageType === 's3'">
        <div class="form-group">
          <label class="form-label required">Access Key ID</label>
          <input
            type="text"
            v-model="formData.s3AccessKeyId"
            placeholder="请输入 Access Key ID"
            class="form-input"
          >
        </div>
        <div class="form-group">
          <label class="form-label required">Secret Access Key</label>
          <input
            type="password"
            v-model="formData.s3SecretAccessKey"
            placeholder="请输入 Secret Access Key"
            class="form-input"
          >
        </div>
        <div class="form-group">
          <label class="form-label required">Bucket 名称</label>
          <input
            type="text"
            v-model="formData.s3Bucket"
            placeholder="请输入 Bucket 名称"
            class="form-input"
          >
        </div>
        <div class="form-group">
          <label class="form-label required">Region</label>
          <input
            type="text"
            v-model="formData.s3Region"
            placeholder="例如: us-east-1"
            class="form-input"
          >
        </div>
        <div class="form-group">
          <label class="form-label">Endpoint (可选)</label>
          <input
            type="text"
            v-model="formData.s3Endpoint"
            placeholder="自定义 Endpoint，留空使用默认"
            class="form-input"
          >
        </div>
        <div class="form-group">
          <label class="form-label">存储路径</label>
          <input
            type="text"
            v-model="formData.s3Path"
            placeholder="例如: images/"
            class="form-input"
          >
        </div>
      </template>

      <!-- MinIO 配置 -->
      <template v-if="formData.storageType === 'minio'">
        <div class="form-group">
          <label class="form-label required">Endpoint</label>
          <input
            type="text"
            v-model="formData.minioEndpoint"
            placeholder="例如: minio.example.com:9000"
            class="form-input"
          >
        </div>
        <div class="form-group">
          <label class="form-label required">Access Key</label>
          <input
            type="text"
            v-model="formData.minioAccessKey"
            placeholder="请输入 Access Key"
            class="form-input"
          >
        </div>
        <div class="form-group">
          <label class="form-label required">Secret Key</label>
          <input
            type="password"
            v-model="formData.minioSecretKey"
            placeholder="请输入 Secret Key"
            class="form-input"
          >
        </div>
        <div class="form-group">
          <label class="form-label required">Bucket 名称</label>
          <input
            type="text"
            v-model="formData.minioBucket"
            placeholder="请输入 Bucket 名称"
            class="form-input"
          >
        </div>
        <div class="form-group">
          <label class="form-label">使用 SSL</label>
          <div class="checkbox-wrapper">
            <label class="checkbox-label">
              <input type="checkbox" v-model="formData.minioUseSSL" />
              <span>启用 HTTPS 连接</span>
            </label>
          </div>
        </div>
        <div class="form-group">
          <label class="form-label">存储路径</label>
          <input
            type="text"
            v-model="formData.minioPath"
            placeholder="例如: images/"
            class="form-input"
          >
        </div>
      </template>

      <!-- 又拍云配置 -->
      <template v-if="formData.storageType === 'upyun'">
        <div class="form-group">
          <label class="form-label required">服务名称</label>
          <input
            type="text"
            v-model="formData.upyunBucket"
            placeholder="请输入服务名称"
            class="form-input"
          >
        </div>
        <div class="form-group">
          <label class="form-label required">操作员账号</label>
          <input
            type="text"
            v-model="formData.upyunOperator"
            placeholder="请输入操作员账号"
            class="form-input"
          >
        </div>
        <div class="form-group">
          <label class="form-label required">操作员密码</label>
          <input
            type="password"
            v-model="formData.upyunPassword"
            placeholder="请输入操作员密码"
            class="form-input"
          >
        </div>
        <div class="form-group">
          <label class="form-label">存储路径</label>
          <input
            type="text"
            v-model="formData.upyunPath"
            placeholder="例如: /images"
            class="form-input"
          >
        </div>
      </template>

      <!-- 操作按钮 -->
      <div class="form-actions">
        <button type="button" class="cancel-btn" @click="goBack">
          取消
        </button>
        <button type="submit" class="submit-btn" :disabled="isSubmitting">
          <i v-if="isSubmitting" class="fa fa-spinner fa-spin mr-2"></i>
          {{ isSubmitting ? '保存中...' : '保存' }}
        </button>
      </div>
    </form>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { systemAPI } from '../../api'
import { useMessage } from '../../composables/useMessage'

const route = useRoute()
const router = useRouter()

const { toast } = useMessage()

// 响应式数据
const isEditing = ref(false)
const isSubmitting = ref(false)

// 表单数据
const formData = reactive({
  id: null as number | null,
  name: '',
  description: '',
  storageType: '',
  accessUrl: '',
  urlQueries: '',
  customPath: '',
  localDiskPath: 'static',
  // FTP/SFTP
  ftpHost: '',
  ftpPort: 21,
  ftpUsername: '',
  ftpPassword: '',
  ftpPath: '/',
  // WebDAV
  webdavUrl: '',
  webdavUsername: '',
  webdavPassword: '',
  webdavPath: '/',
  // 阿里云 OSS
  ossAccessKeyId: '',
  ossAccessKeySecret: '',
  ossBucket: '',
  ossEndpoint: '',
  ossPath: '',
  // 腾讯云 COS
  cosSecretId: '',
  cosSecretKey: '',
  cosBucket: '',
  cosRegion: '',
  cosPath: '',
  // 七牛云
  qiniuAccessKey: '',
  qiniuSecretKey: '',
  qiniuBucket: '',
  qiniuRegion: '',
  qiniuPath: '',
  // AWS S3
  s3AccessKeyId: '',
  s3SecretAccessKey: '',
  s3Bucket: '',
  s3Region: '',
  s3Endpoint: '',
  s3Path: '',
  // MinIO
  minioEndpoint: '',
  minioAccessKey: '',
  minioSecretKey: '',
  minioBucket: '',
  minioUseSSL: true,
  minioPath: '',
  // 又拍云
  upyunBucket: '',
  upyunOperator: '',
  upyunPassword: '',
  upyunPath: ''
})

// 表单验证错误
const errors = reactive<{
  name: string;
  storageType: string;
  accessUrl: string;
  [key: string]: string;
}>({
  name: '',
  storageType: '',
  accessUrl: ''
})

// 页面初始化
onMounted(async () => {
  const id = route.params.id
  isEditing.value = !!id

  if (isEditing.value) {
    await loadStrategyData()
  }
})

// 加载策略数据
const loadStrategyData = async () => {
  try {
    const response = await systemAPI.getStorageStrategyDetail(Number(route.params.id))
    console.log('API响应:', response)
    if (response.data && response.data.status) {
      const strategy = response.data.data
      console.log('加载的策略数据:', strategy)

      // 获取configs配置
      let configs: any = {}
      if (typeof strategy.configs === 'string') {
        try {
          configs = JSON.parse(strategy.configs)
          console.log('解析的配置:', configs)
        } catch (e) {
          console.error('解析configs失败:', e)
          configs = {}
        }
      } else {
        configs = strategy.configs || {}
        console.log('配置对象:', configs)
      }

      // 映射数据到表单
      formData.id = strategy.id
      formData.name = strategy.name
      formData.description = strategy.introduction || ''

      // 映射存储策略类型
      const keyToTypeMap: { [key: string]: string } = {
        '1': 'local',
        '2': 'ftp',
        '3': 'sftp',
        '4': 'oss',
        '5': 'cos',
        '6': 'qiniu',
        '7': 'webdav',
        '8': 's3',
        '9': 'minio',
        '10': 'upyun'
      }
      formData.storageType = keyToTypeMap[strategy.key] || strategy.key
      formData.accessUrl = configs.url || ''
      formData.urlQueries = typeof configs.queries === 'string' ? configs.queries : ''
      formData.customPath =
        configs.root || (formData.storageType === 'local' ? '/static' : '/')
      formData.localDiskPath = (configs.base_path as string) || 'static'

      // 根据存储类型加载特定配置
      if (formData.storageType === 'ftp' || formData.storageType === 'sftp') {
        formData.ftpHost = configs.host || ''
        formData.ftpPort = configs.port || (formData.storageType === 'ftp' ? 21 : 22)
        formData.ftpUsername = configs.username || ''
        formData.ftpPassword = configs.password || ''
        formData.ftpPath = configs.path || '/'
      } else if (formData.storageType === 'webdav') {
        formData.webdavUrl = configs.webdav_url || ''
        formData.webdavUsername = configs.username || ''
        formData.webdavPassword = configs.password || ''
        formData.webdavPath = configs.path || '/'
      } else if (formData.storageType === 'oss') {
        formData.ossAccessKeyId = configs.access_key_id || ''
        formData.ossAccessKeySecret = configs.access_key_secret || ''
        formData.ossBucket = configs.bucket || ''
        formData.ossEndpoint = configs.endpoint || ''
        formData.ossPath = configs.path || ''
      } else if (formData.storageType === 'cos') {
        formData.cosSecretId = configs.secret_id || ''
        formData.cosSecretKey = configs.secret_key || ''
        formData.cosBucket = configs.bucket || ''
        formData.cosRegion = configs.region || ''
        formData.cosPath = configs.path || ''
      } else if (formData.storageType === 'qiniu') {
        formData.qiniuAccessKey = configs.access_key || ''
        formData.qiniuSecretKey = configs.secret_key || ''
        formData.qiniuBucket = configs.bucket || ''
        formData.qiniuRegion = configs.region || ''
        formData.qiniuPath = configs.path || ''
      } else if (formData.storageType === 's3') {
        formData.s3AccessKeyId = configs.access_key_id || ''
        formData.s3SecretAccessKey = configs.secret_access_key || ''
        formData.s3Bucket = configs.bucket || ''
        formData.s3Region = configs.region || ''
        formData.s3Endpoint = configs.endpoint || ''
        formData.s3Path = configs.path || ''
      } else if (formData.storageType === 'minio') {
        formData.minioEndpoint = configs.endpoint || ''
        formData.minioAccessKey = configs.access_key || ''
        formData.minioSecretKey = configs.secret_key || ''
        formData.minioBucket = configs.bucket || ''
        formData.minioUseSSL = configs.use_ssl !== false
        formData.minioPath = configs.path || ''
      } else if (formData.storageType === 'upyun') {
        formData.upyunBucket = configs.bucket || ''
        formData.upyunOperator = configs.operator || ''
        formData.upyunPassword = configs.password || ''
        formData.upyunPath = configs.path || ''
      }

      console.log('表单数据已更新:', formData)
    } else {
      console.error('API返回失败:', response.data?.message)
      toast.error(`加载策略失败: ${response.data?.message || '未知错误'}`)
    }
  } catch (error: any) {
    console.error('加载策略数据失败:', error)
    const errorMsg = error.response?.data?.error || error.message || '未知错误'
    toast.error(`加载策略失败: ${errorMsg}`)
  }
}

// 返回上一页
const goBack = () => {
  router.back()
}

// 验证字段
const validateField = (field: string) => {
  errors[field] = ''
  
  switch (field) {
    case 'name':
      if (!formData.name.trim()) {
        errors.name = '策略名称不能为空'
      } else if (formData.name.length > 50) {
        errors.name = '策略名称不能超过50个字符'
      }
      break
    case 'storageType':
      if (!formData.storageType) {
        errors.storageType = '请选择储存策略'
      }
      break
    case 'accessUrl':
      if (!formData.accessUrl.trim()) {
        errors.accessUrl = '访问网址不能为空'
      } else if (!isValidUrl(formData.accessUrl)) {
        errors.accessUrl = '请输入有效的URL地址'
      }
      break
  }
}

// URL验证
const isValidUrl = (url: string) => {
  try {
    new URL(url)
    return true
  } catch {
    return false
  }
}

// 从URL中提取路径作为root
const extractRootFromUrl = (url: string) => {
  try {
    const urlObj = new URL(url)
    const pathname = urlObj.pathname
    // 如果路径不是根路径，则使用该路径
    if (pathname && pathname !== '/') {
      return pathname
    }
    return '/'
  } catch {
    return '/'
  }
}

// 监听访问网址变化，自动提取root
const handleAccessUrlChange = () => {
  if (formData.accessUrl && isValidUrl(formData.accessUrl)) {
    const extractedRoot = extractRootFromUrl(formData.accessUrl)
    if (extractedRoot !== '/') {
      formData.customPath = extractedRoot
      console.log('从URL提取的root路径:', extractedRoot)
    }
  }
}

// 储存策略类型变化处理
const handleStorageTypeChange = () => {
  // 根据存储类型设置默认值
  if (formData.storageType === 'local') {
    if (!formData.accessUrl || formData.accessUrl === 'https://') {
      formData.accessUrl =
        typeof window !== 'undefined' && window.location?.origin
          ? `${window.location.origin}/static`
          : 'http://127.0.0.1:8080/static'
    }
    if (!formData.localDiskPath) {
      formData.localDiskPath = 'static'
    }
    if (formData.customPath === undefined || formData.customPath === '' || formData.customPath === '/') {
      formData.customPath = '/static'
    }
  } else if (formData.storageType === 'ftp') {
    if (!formData.ftpPort) {
      formData.ftpPort = 21
    }
    if (!formData.ftpPath) {
      formData.ftpPath = '/'
    }
  } else if (formData.storageType === 'sftp') {
    if (!formData.ftpPort) {
      formData.ftpPort = 22
    }
    if (!formData.ftpPath) {
      formData.ftpPath = '/'
    }
  } else if (formData.storageType === 'webdav') {
    if (!formData.webdavPath) {
      formData.webdavPath = '/'
    }
  }
}

// 表单提交
const handleSubmit = async () => {
  validateField('name')
  validateField('storageType')
  validateField('accessUrl')

  if (Object.values(errors).some(error => error)) {
    toast.warning('请检查表单填写是否正确')
    return
  }

  isSubmitting.value = true

  try {
    // 基础配置
    const configs: any = {
      url: formData.accessUrl,
      root: formData.customPath || '/',
      queries: formData.urlQueries || null,
      auth_type: '1'
    }

    if (formData.storageType === 'local') {
      const d = (formData.localDiskPath || 'static').trim() || 'static'
      configs.base_path = d
    }

    // 根据存储类型添加特定配置
    if (formData.storageType === 'ftp' || formData.storageType === 'sftp') {
      configs.host = formData.ftpHost
      configs.port = formData.ftpPort
      configs.username = formData.ftpUsername
      configs.password = formData.ftpPassword
      configs.path = formData.ftpPath
    } else if (formData.storageType === 'webdav') {
      configs.webdav_url = formData.webdavUrl
      configs.username = formData.webdavUsername
      configs.password = formData.webdavPassword
      configs.path = formData.webdavPath
    } else if (formData.storageType === 'oss') {
      configs.access_key_id = formData.ossAccessKeyId
      configs.access_key_secret = formData.ossAccessKeySecret
      configs.bucket = formData.ossBucket
      configs.endpoint = formData.ossEndpoint
      configs.path = formData.ossPath
    } else if (formData.storageType === 'cos') {
      configs.secret_id = formData.cosSecretId
      configs.secret_key = formData.cosSecretKey
      configs.bucket = formData.cosBucket
      configs.region = formData.cosRegion
      configs.path = formData.cosPath
    } else if (formData.storageType === 'qiniu') {
      configs.access_key = formData.qiniuAccessKey
      configs.secret_key = formData.qiniuSecretKey
      configs.bucket = formData.qiniuBucket
      configs.region = formData.qiniuRegion
      configs.path = formData.qiniuPath
    } else if (formData.storageType === 's3') {
      configs.access_key_id = formData.s3AccessKeyId
      configs.secret_access_key = formData.s3SecretAccessKey
      configs.bucket = formData.s3Bucket
      configs.region = formData.s3Region
      configs.endpoint = formData.s3Endpoint
      configs.path = formData.s3Path
    } else if (formData.storageType === 'minio') {
      configs.endpoint = formData.minioEndpoint
      configs.access_key = formData.minioAccessKey
      configs.secret_key = formData.minioSecretKey
      configs.bucket = formData.minioBucket
      configs.use_ssl = formData.minioUseSSL
      configs.path = formData.minioPath
    } else if (formData.storageType === 'upyun') {
      configs.bucket = formData.upyunBucket
      configs.operator = formData.upyunOperator
      configs.password = formData.upyunPassword
      configs.path = formData.upyunPath
    }

    // 映射存储策略类型到key
    const typeToKeyMap: { [key: string]: string } = {
      'local': '1',
      'ftp': '2',
      'sftp': '3',
      'oss': '4',
      'cos': '5',
      'qiniu': '6',
      'webdav': '7',
      's3': '8',
      'minio': '9',
      'upyun': '10'
    }
    const strategyKey = typeToKeyMap[formData.storageType] || formData.storageType

  const requestData = {
    name: formData.name,
    introduction: formData.description,
    key: strategyKey,
    configs
  }

    console.log('提交数据:', requestData)

    let response
    if (isEditing.value && formData.id) {
      response = await systemAPI.updateStorageStrategy(formData.id, requestData)
    } else {
      response = await systemAPI.createStorageStrategy(requestData)
    }

    console.log('API响应:', response)

    if (response.data && response.data.status) {
      toast.success(isEditing.value ? '更新成功！' : '创建成功！')
      // 返回上一页并刷新数据
      router.push('/admin/storage')
    } else {
      toast.error(`操作失败: ${response.data?.message || '未知错误'}`)
    }
  } catch (error: any) {
    console.error('保存失败:', error)
    const errorMsg = error.response?.data?.error || error.message || '未知错误'
    toast.error(`保存失败: ${errorMsg}`)
  } finally {
    isSubmitting.value = false
  }
}
</script>

<style scoped>
.storage-edit {
  max-width: 800px;
  margin: 0 auto;
  padding: 24px;
}

.page-header {
  margin-bottom: 24px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.back-btn {
  background: none;
  border: none;
  font-size: 18px;
  color: #6b7280;
  cursor: pointer;
  padding: 8px;
  border-radius: 6px;
  transition: background-color 0.2s;
}

.back-btn:hover {
  background-color: #f3f4f6;
}

.page-title {
  font-size: 24px;
  font-weight: 600;
  color: #374151;
  margin: 0;
}

.info-banner {
  background-color: #dbeafe;
  color: #1e40af;
  padding: 12px 16px;
  border-radius: 6px;
  margin-bottom: 24px;
  font-size: 14px;
  line-height: 1.5;
}

.info-banner i {
  margin-right: 8px;
}

.edit-form {
  background-color: white;
  border-radius: 8px;
  padding: 24px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.05);
}

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

.form-label.required::after {
  content: '*';
  color: #ef4444;
  margin-left: 4px;
}

.form-input,
.form-select,
.form-textarea {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 14px;
  transition: border-color 0.2s;
}

.form-input:focus,
.form-select:focus,
.form-textarea:focus {
  outline: none;
  border-color: #3b82f6;
}

.form-input.error {
  border-color: #ef4444;
}

.form-textarea {
  resize: vertical;
  min-height: 80px;
}

.error-message {
  color: #ef4444;
  font-size: 12px;
  margin-top: 4px;
}

.form-hint {
  font-size: 12px;
  color: #6b7280;
  margin-top: 6px;
  line-height: 1.4;
}

.warning-message {
  background-color: #fef3c7;
  color: #92400e;
  padding: 8px 12px;
  border-radius: 4px;
  font-size: 12px;
  line-height: 1.4;
  margin-top: 8px;
}

.warning-message i {
  margin-right: 6px;
}

.role-groups-display {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  background-color: #f9fafb;
}

.selected-groups {
  flex: 1;
  color: #6b7280;
  font-size: 14px;
}

.edit-btn {
  background: none;
  border: none;
  color: #3b82f6;
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
  transition: background-color 0.2s;
}

.edit-btn:hover {
  background-color: #eff6ff;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 32px;
  padding-top: 24px;
  border-top: 1px solid #e5e7eb;
}

.cancel-btn {
  background-color: #f3f4f6;
  color: #374151;
  border: 1px solid #d1d5db;
  padding: 10px 20px;
  border-radius: 6px;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s;
}

.cancel-btn:hover {
  background-color: #e5e7eb;
}

.submit-btn {
  background-color: #3b82f6;
  color: white;
  border: none;
  padding: 10px 20px;
  border-radius: 6px;
  font-size: 14px;
  cursor: pointer;
  transition: background-color 0.2s;
}

.submit-btn:hover:not(:disabled) {
  background-color: #2563eb;
}

.submit-btn:disabled {
  background-color: #9ca3af;
  cursor: not-allowed;
}

/* 模态框样式 */
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
  border-radius: 8px;
  width: 90%;
  max-width: 500px;
  max-height: 80vh;
  display: flex;
  flex-direction: column;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid #e5e7eb;
}

.modal-header h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
}

.close-btn {
  background: none;
  border: none;
  font-size: 20px;
  color: #6b7280;
  cursor: pointer;
  padding: 4px;
  line-height: 1;
}

.close-btn:hover {
  color: #374151;
}

.modal-body {
  padding: 20px;
  overflow-y: auto;
  flex: 1;
}

.role-group-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.role-group-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px;
  border: 1px solid #e5e7eb;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
}

.role-group-item:hover {
  background-color: #f9fafb;
  border-color: #3b82f6;
}

.role-group-item input[type="checkbox"] {
  width: 18px;
  height: 18px;
  cursor: pointer;
}

.role-group-item span {
  font-size: 14px;
  color: #374151;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 20px;
  border-top: 1px solid #e5e7eb;
}

.help-text {
  font-size: 12px;
  color: #6b7280;
  margin-top: 6px;
}

.checkbox-wrapper {
  margin-top: 8px;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  font-size: 14px;
  color: #374151;
}

.checkbox-label input[type="checkbox"] {
  width: 18px;
  height: 18px;
  cursor: pointer;
}

.checkbox-label span {
  user-select: none;
}
</style> 
