<template>
  <div class="min-h-screen bg-neutral-100 flex items-center justify-center p-4">
    <div class="max-w-2xl w-full bg-white rounded-2xl shadow-xl p-8">
      <!-- 头部 -->
      <div class="text-center mb-8">
        <div class="w-16 h-16 bg-primary/10 rounded-2xl flex items-center justify-center mx-auto mb-4">
          <i class="fa fa-cloud-upload text-primary text-2xl"></i>
        </div>
        <h1 class="text-2xl font-bold text-neutral-800">Slink 图床初始化</h1>
        <p class="text-neutral-500 text-sm mt-1">首次运行，请创建站长账号并配置数据库和缓存</p>
      </div>

      <!-- 步骤指示器 -->
      <div class="flex justify-center mb-8">
        <template v-for="(step, index) in steps" :key="index">
          <div class="flex items-center">
            <div class="flex flex-col items-center">
              <div
                :class="[
                  'w-9 h-9 rounded-full flex items-center justify-center text-sm font-semibold transition-all',
                  currentStep > index + 1
                    ? 'bg-emerald-500 text-white'
                    : currentStep === index + 1
                      ? 'bg-primary text-white'
                      : 'bg-neutral-200 text-neutral-500'
                ]"
              >
                {{ currentStep > index + 1 ? '✓' : index + 1 }}
              </div>
              <span class="text-xs mt-1.5" :class="currentStep >= index + 1 ? 'text-neutral-700 font-medium' : 'text-neutral-400'">
                {{ step }}
              </span>
            </div>
            <div
              v-if="index < steps.length - 1"
              class="w-16 h-0.5 mx-3 mt-[-16px] rounded"
              :class="currentStep > index + 1 ? 'bg-emerald-500' : 'bg-neutral-200'"
            ></div>
          </div>
        </template>
      </div>

      <!-- 表单 -->
      <form @submit.prevent="handleSubmit">
        <!-- 步骤1: 管理员账号 -->
        <div v-show="currentStep === 1" class="space-y-4">
          <h2 class="text-lg font-semibold text-neutral-800 mb-2">站长账号（唯一管理员）</h2>

          <div>
            <label class="block text-sm font-medium text-neutral-700 mb-1.5">登录账号</label>
            <input
              v-model="formData.adminAccount"
              type="text" required autocomplete="username"
              class="w-full px-4 py-2.5 border border-neutral-300 rounded-lg focus:ring-2 focus:ring-primary/20 focus:border-primary outline-none transition-colors text-sm"
              placeholder="任意账号，如 admin"
            />
          </div>

          <div>
            <label class="block text-sm font-medium text-neutral-700 mb-1.5">登录密码</label>
            <input
              v-model="formData.adminPassword"
              type="password" required minlength="6"
              class="w-full px-4 py-2.5 border border-neutral-300 rounded-lg focus:ring-2 focus:ring-primary/20 focus:border-primary outline-none transition-colors text-sm"
              placeholder="至少6位字符"
            />
            <div class="mt-1.5 flex gap-1">
              <div v-for="i in 4" :key="i" class="h-1 flex-1 rounded-full transition-colors"
                :class="passwordStrength >= i ? (passwordStrength <= 2 ? 'bg-orange-400' : 'bg-emerald-500') : 'bg-neutral-200'">
              </div>
              <span class="text-xs text-neutral-400 ml-2">{{ ['弱','一般','强','很强'][passwordStrength - 1] || '' }}</span>
            </div>
          </div>

          <div>
            <label class="block text-sm font-medium text-neutral-700 mb-1.5">确认密码</label>
            <input
              v-model="confirmPassword"
              type="password" required
              :class="[
                'w-full px-4 py-2.5 border rounded-lg focus:ring-2 focus:ring-primary/20 focus:border-primary outline-none transition-colors text-sm',
                passwordMismatch ? 'border-red-400' : 'border-neutral-300'
              ]"
              placeholder="再次输入密码"
            />
            <p v-if="passwordMismatch" class="text-red-500 text-xs mt-1">两次输入的密码不一致</p>
          </div>
        </div>

        <!-- 步骤2: 数据库配置 -->
        <div v-show="currentStep === 2" class="space-y-4">
          <h2 class="text-lg font-semibold text-neutral-800 mb-2">数据库配置</h2>

          <div>
            <label class="block text-sm font-medium text-neutral-700 mb-1.5">数据库类型</label>
            <div class="grid grid-cols-4 gap-2">
              <button
                v-for="db in dbTypes" :key="db.value" type="button"
                @click="formData.databaseConfig.type = db.value"
                :class="[
                  'py-2.5 px-2 rounded-lg text-xs font-medium transition-all border text-center',
                  formData.databaseConfig.type === db.value
                    ? 'bg-primary/5 border-primary text-primary'
                    : 'bg-white border-neutral-200 text-neutral-500 hover:border-neutral-300'
                ]"
              >
                <i :class="db.icon + ' block text-base mb-1'"></i>
                {{ db.label }}
              </button>
            </div>
          </div>

          <!-- SQLite -->
          <div v-if="formData.databaseConfig.type === 'sqlite'" class="bg-neutral-50 rounded-lg p-4 border border-neutral-200">
            <div class="flex items-center gap-2 text-emerald-600 text-xs mb-3">
              <i class="fa fa-check-circle"></i> 无需额外配置，即开即用
            </div>
            <label class="block text-sm font-medium text-neutral-700 mb-1.5">数据库文件路径</label>
            <input
              v-model="formData.databaseConfig.dbname"
              type="text"
              class="w-full px-4 py-2.5 border border-neutral-300 rounded-lg focus:ring-2 focus:ring-primary/20 focus:border-primary outline-none transition-colors text-sm"
              placeholder="data.db"
            />
          </div>

          <!-- 远程数据库 -->
          <template v-else>
            <div class="grid grid-cols-2 gap-3">
              <div>
                <label class="block text-sm font-medium text-neutral-700 mb-1.5">主机地址</label>
                <input v-model="formData.databaseConfig.host" type="text"
                  class="w-full px-4 py-2.5 border border-neutral-300 rounded-lg focus:ring-2 focus:ring-primary/20 focus:border-primary outline-none transition-colors text-sm"
                  placeholder="localhost" />
              </div>
              <div>
                <label class="block text-sm font-medium text-neutral-700 mb-1.5">端口</label>
                <input
                  :value="formData.databaseConfig.port != null ? String(formData.databaseConfig.port) : ''"
                  @input="onPortInput($event as InputEvent)"
                  type="text" inputmode="numeric"
                  class="w-full px-4 py-2.5 border border-neutral-300 rounded-lg focus:ring-2 focus:ring-primary/20 focus:border-primary outline-none transition-colors text-sm"
                  :placeholder="String(getDefaultPort())" />
              </div>
            </div>
            <div class="mt-3">
              <label class="block text-sm font-medium text-neutral-700 mb-1.5">数据库名</label>
              <input v-model="formData.databaseConfig.dbname" type="text"
                class="w-full px-4 py-2.5 border border-neutral-300 rounded-lg focus:ring-2 focus:ring-primary/20 focus:border-primary outline-none transition-colors text-sm"
                placeholder="slink" />
            </div>
            <div class="grid grid-cols-2 gap-3 mt-3">
              <div>
                <label class="block text-sm font-medium text-neutral-700 mb-1.5">用户名</label>
                <input v-model="formData.databaseConfig.user" type="text"
                  class="w-full px-4 py-2.5 border border-neutral-300 rounded-lg focus:ring-2 focus:ring-primary/20 focus:border-primary outline-none transition-colors text-sm"
                  placeholder="root" />
              </div>
              <div>
                <label class="block text-sm font-medium text-neutral-700 mb-1.5">密码</label>
                <input v-model="formData.databaseConfig.password" type="password"
                  class="w-full px-4 py-2.5 border border-neutral-300 rounded-lg focus:ring-2 focus:ring-primary/20 focus:border-primary outline-none transition-colors text-sm"
                  placeholder="数据库密码" />
              </div>
            </div>
          </template>

          <!-- 测试连接 -->
          <button type="button" @click="testDatabaseConnection" :disabled="testing"
            class="w-full py-2.5 rounded-lg text-sm font-medium border transition-all flex items-center justify-center gap-2"
            :class="testResult
              ? (testResult.success ? 'bg-emerald-50 border-emerald-300 text-emerald-700' : 'bg-red-50 border-red-300 text-red-700')
              : 'bg-neutral-50 border-neutral-200 text-neutral-600 hover:bg-neutral-100'"
          >
            <i v-if="testing" class="fa fa-spinner fa-spin"></i>
            <i v-else-if="testResult?.success" class="fa fa-check-circle"></i>
            <i v-else-if="testResult && !testResult.success" class="fa fa-times-circle"></i>
            <i v-else class="fa fa-plug"></i>
            {{ testing ? '测试中...' : testResult?.success ? '连接成功' : testResult && !testResult.success ? '连接失败，点击重试' : '测试数据库连接' }}
          </button>
        </div>

        <!-- 步骤3: 缓存配置 -->
        <div v-show="currentStep === 3" class="space-y-4">
          <h2 class="text-lg font-semibold text-neutral-800 mb-2">缓存配置</h2>

          <div>
            <label class="block text-sm font-medium text-neutral-700 mb-1.5">缓存类型</label>
            <div class="grid grid-cols-3 gap-2">
              <button
                v-for="c in cacheTypes" :key="c.value" type="button"
                @click="formData.cacheType = c.value"
                :class="[
                  'py-3 px-2 rounded-lg text-center transition-all border',
                  formData.cacheType === c.value
                    ? 'bg-primary/5 border-primary text-primary'
                    : 'bg-white border-neutral-200 text-neutral-500 hover:border-neutral-300'
                ]"
              >
                <i :class="c.icon + ' block text-lg mb-1'"></i>
                <div class="text-xs font-medium">{{ c.label }}</div>
              </button>
            </div>
          </div>

          <div v-if="formData.cacheType === 'file'" class="bg-neutral-50 rounded-lg p-4 border border-neutral-200">
            <label class="block text-sm font-medium text-neutral-700 mb-1.5">缓存目录</label>
            <input v-model="formData.cachePath" type="text"
              class="w-full px-4 py-2.5 border border-neutral-300 rounded-lg focus:ring-2 focus:ring-primary/20 focus:border-primary outline-none transition-colors text-sm"
              placeholder="cache" />
          </div>

          <div v-if="formData.cacheType === 'redis'" class="bg-neutral-50 rounded-lg p-4 border border-neutral-200 space-y-3">
            <div class="grid grid-cols-2 gap-3">
              <div>
                <label class="block text-sm font-medium text-neutral-700 mb-1.5">Redis 主机</label>
                <input v-model="formData.redisHost" type="text"
                  class="w-full px-4 py-2.5 border border-neutral-300 rounded-lg focus:ring-2 focus:ring-primary/20 focus:border-primary outline-none transition-colors text-sm"
                  placeholder="localhost" />
              </div>
              <div>
                <label class="block text-sm font-medium text-neutral-700 mb-1.5">端口</label>
                <input
                  :value="formData.redisPort != null ? String(formData.redisPort) : ''"
                  @input="onRedisPortInput($event as InputEvent)" type="text" inputmode="numeric"
                  class="w-full px-4 py-2.5 border border-neutral-300 rounded-lg focus:ring-2 focus:ring-primary/20 focus:border-primary outline-none transition-colors text-sm"
                  placeholder="6379" />
              </div>
            </div>
            <div>
              <label class="block text-sm font-medium text-neutral-700 mb-1.5">密码 <span class="text-neutral-400 font-normal">(可选)</span></label>
              <input v-model="formData.redisPassword" type="password"
                class="w-full px-4 py-2.5 border border-neutral-300 rounded-lg focus:ring-2 focus:ring-primary/20 focus:border-primary outline-none transition-colors text-sm"
                placeholder="留空表示无密码" />
            </div>
            <div>
              <label class="block text-sm font-medium text-neutral-700 mb-1.5">数据库编号</label>
              <input
                :value="formData.redisDB != null ? String(formData.redisDB) : ''"
                @input="onRedisDBInput($event as InputEvent)" type="text" inputmode="numeric"
                class="w-full px-4 py-2.5 border border-neutral-300 rounded-lg focus:ring-2 focus:ring-primary/20 focus:border-primary outline-none transition-colors text-sm"
                placeholder="0" />
            </div>
            <button type="button" @click="testRedisConnection" :disabled="testingRedis"
              class="w-full py-2.5 rounded-lg text-sm font-medium border transition-all flex items-center justify-center gap-2"
              :class="redisTestResult
                ? (redisTestResult.success ? 'bg-emerald-50 border-emerald-300 text-emerald-700' : 'bg-red-50 border-red-300 text-red-700')
                : 'bg-white border-neutral-200 text-neutral-600 hover:bg-neutral-100'"
            >
              <i v-if="testingRedis" class="fa fa-spinner fa-spin"></i>
              <i v-else-if="redisTestResult?.success" class="fa fa-check-circle"></i>
              <i v-else-if="redisTestResult && !redisTestResult.success" class="fa fa-times-circle"></i>
              <i v-else class="fa fa-plug"></i>
              {{ testingRedis ? '测试中...' : redisTestResult?.success ? '连接成功' : redisTestResult && !redisTestResult.success ? '连接失败，点击重试' : '测试 Redis 连接' }}
            </button>
          </div>

          <!-- 配置摘要 -->
          <div class="bg-blue-50 border border-blue-200 rounded-lg p-4">
            <h3 class="font-semibold text-blue-800 text-sm mb-2">配置摘要</h3>
            <ul class="text-sm text-blue-700 space-y-1">
              <li>登录账号: {{ formData.adminAccount || '—' }}</li>
              <li>数据库类型: {{ getDatabaseTypeName() }}</li>
              <li>缓存类型: {{ getCacheTypeName() }}</li>
              <li v-if="formData.cacheType === 'redis'">Redis: {{ formData.redisHost }}:{{ formData.redisPort }}</li>
            </ul>
          </div>
        </div>

        <!-- 按钮组 -->
        <div class="flex justify-between mt-8">
          <button
            v-if="currentStep > 1"
            type="button" @click="previousStep"
            class="px-6 py-2.5 bg-neutral-100 text-neutral-700 rounded-lg hover:bg-neutral-200 transition-colors text-sm font-medium"
          >
            上一步
          </button>
          <div v-else></div>

          <button
            v-if="currentStep < 3"
            type="button" @click="nextStep" :disabled="!canProceed"
            class="px-6 py-2.5 bg-primary text-white rounded-lg hover:bg-primary/90 disabled:bg-neutral-300 disabled:cursor-not-allowed transition-colors text-sm font-medium"
          >
            下一步
          </button>
          <button
            v-else
            type="submit" :disabled="submitting"
            class="px-6 py-2.5 bg-primary text-white rounded-lg hover:bg-primary/90 disabled:bg-neutral-300 disabled:cursor-not-allowed transition-colors text-sm font-medium flex items-center gap-2"
          >
            <i v-if="submitting" class="fa fa-spinner fa-spin"></i>
            {{ submitting ? '初始化中...' : '完成初始化' }}
          </button>
        </div>
      </form>

      <!-- 错误提示 -->
      <div v-if="error" class="mt-4 p-3 bg-red-50 border border-red-200 rounded-lg flex items-start gap-2">
        <i class="fa fa-exclamation-circle text-red-500 mt-0.5"></i>
        <p class="text-red-700 text-sm">{{ error }}</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { http } from '../utils/request'
import { useMessage } from '../composables/useMessage'

const router = useRouter()
const { toast } = useMessage()

const steps = ['站长账号', '数据库配置', '缓存配置']
const currentStep = ref(1)
const submitting = ref(false)
const testing = ref(false)
const error = ref('')
const confirmPassword = ref('')
const testResult = ref<{ success: boolean; message: string } | null>(null)

const dbTypes = [
  { value: 'sqlite', label: 'SQLite', icon: 'fa fa-file' },
  { value: 'mysql', label: 'MySQL', icon: 'fa fa-database' },
  { value: 'postgres', label: 'PostgreSQL', icon: 'fa fa-server' },
  { value: 'sqlserver', label: 'SQL Server', icon: 'fa fa-windows' },
]

const cacheTypes = [
  { value: 'memory', label: '内存缓存', icon: 'fa fa-microchip' },
  { value: 'file', label: '文件缓存', icon: 'fa fa-file-archive-o' },
  { value: 'redis', label: 'Redis', icon: 'fa fa-database' },
]

const formData = ref({
  adminAccount: '',
  adminPassword: '',
  databaseConfig: {
    type: 'sqlite', host: 'localhost', port: 3306,
    user: '', password: '', dbname: 'data.db', sslmode: 'disable', charset: 'utf8mb4'
  },
  cacheType: 'memory',
  cachePath: 'cache',
  redisHost: 'localhost',
  redisPort: 6379,
  redisPassword: '',
  redisDB: 0
})

const testingRedis = ref(false)
const redisTestResult = ref<{ success: boolean; message: string } | null>(null)

const passwordStrength = computed(() => {
  const pwd = formData.value.adminPassword
  let score = 0
  if (pwd.length >= 6) score++
  if (pwd.length >= 10) score++
  if (/[A-Z]/.test(pwd) && /[a-z]/.test(pwd)) score++
  if (/[0-9]/.test(pwd)) score++
  return Math.min(4, score)
})

const passwordMismatch = computed(() => {
  return !!confirmPassword.value && formData.value.adminPassword !== confirmPassword.value
})

const canProceed = computed(() => {
  if (currentStep.value === 1) {
    return !!formData.value.adminAccount && formData.value.adminPassword.length >= 6 && !passwordMismatch.value
  }
  if (currentStep.value === 2) {
    const c = formData.value.databaseConfig
    if (c.type === 'sqlite') return !!c.dbname
    return !!(c.host && c.port && c.dbname && c.user)
  }
  return true
})

const getDefaultPort = () => {
  return { mysql: 3306, postgres: 5432, sqlserver: 1433 }[formData.value.databaseConfig.type] || 3306
}

const getDatabaseTypeName = () => {
  return { sqlite: 'SQLite', mysql: 'MySQL', postgres: 'PostgreSQL', sqlserver: 'SQL Server' }[formData.value.databaseConfig.type] || 'SQLite'
}

const getCacheTypeName = () => {
  return { memory: '内存缓存', file: '文件缓存', redis: 'Redis 缓存' }[formData.value.cacheType] || '内存缓存'
}

const nextStep = () => {
  if (canProceed.value && currentStep.value < 3) { currentStep.value++; error.value = '' }
}

const previousStep = () => {
  if (currentStep.value > 1) { currentStep.value--; error.value = '' }
}

const testDatabaseConnection = async () => {
  testing.value = true; testResult.value = null; error.value = ''
  try {
    const response = await http.post('/api/init/test-db', formData.value.databaseConfig)
    testResult.value = response.data.code === 200
      ? { success: true, message: '连接成功' }
      : { success: false, message: response.data.msg || '连接失败' }
  } catch (err: any) {
    testResult.value = { success: false, message: err.response?.data?.msg || '连接失败' }
  } finally { testing.value = false }
}

const testRedisConnection = async () => {
  testingRedis.value = true; redisTestResult.value = null; error.value = ''
  try {
    const response = await http.post('/api/init/test-redis', {
      host: formData.value.redisHost || 'localhost',
      port: formData.value.redisPort || 6379,
      password: formData.value.redisPassword,
      db: formData.value.redisDB || 0
    })
    redisTestResult.value = response.data.code === 200
      ? { success: true, message: '连接成功' }
      : { success: false, message: response.data.msg || '连接失败' }
  } catch (err: any) {
    redisTestResult.value = { success: false, message: err.response?.data?.msg || '连接失败' }
  } finally { testingRedis.value = false }
}

const handleSubmit = async () => {
  if (!canProceed.value) return
  submitting.value = true; error.value = ''
  try {
    const payload: any = {
      admin_account: formData.value.adminAccount,
      admin_password: formData.value.adminPassword,
      database_config: formData.value.databaseConfig,
      cache_type: formData.value.cacheType,
      cache_path: formData.value.cachePath
    }
    if (formData.value.cacheType === 'redis') {
      payload.redis_host = formData.value.redisHost || 'localhost'
      payload.redis_port = formData.value.redisPort || 6379
      payload.redis_password = formData.value.redisPassword || ''
      payload.redis_db = formData.value.redisDB || 0
    }
    const response = await http.post('/api/init/setup', payload)
    if (response.data.code === 200) {
      toast.success('初始化成功！请使用管理员账号登录。')
      window.location.href = '/'
    } else {
      error.value = response.data.msg || '初始化失败'
    }
  } catch (err: any) {
    error.value = err.response?.data?.msg || '初始化失败，请检查配置'
  } finally { submitting.value = false }
}

const onPortInput = (event: InputEvent) => {
  const target = event.target as HTMLInputElement
  target.value = target.value.replace(/[^0-9]/g, '')
  formData.value.databaseConfig.port = target.value ? Number(target.value) : getDefaultPort()
}

const onRedisPortInput = (event: InputEvent) => {
  const target = event.target as HTMLInputElement
  target.value = target.value.replace(/[^0-9]/g, '')
  formData.value.redisPort = target.value ? Number(target.value) : 6379
}

const onRedisDBInput = (event: InputEvent) => {
  const target = event.target as HTMLInputElement
  target.value = target.value.replace(/[^0-9]/g, '')
  formData.value.redisDB = target.value ? Number(target.value) : 0
}
</script>
