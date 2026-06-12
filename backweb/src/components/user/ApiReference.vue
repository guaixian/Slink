<template>
  <div class="api-reference">
    <div class="page-header">
      <h1 class="text-2xl font-bold text-neutral-800">
        <i class="fa fa-plug mr-2 text-primary"></i>
        开放 API 说明
      </h1>
      <p class="text-neutral-600 text-sm mt-1">
        供脚本、App、其他服务调用。所有路径均相对于服务根地址；生产环境请使用 <code class="mono-inline">https://</code> 与反代后的域名。
      </p>
    </div>

    <div class="doc-card">
      <h2 class="section-title">服务根与路径前缀</h2>
      <p class="text-sm text-neutral-600 mb-2">当前页所在站点即 API 同域，基础路径为：</p>
      <div class="code-line">
        <code>{{ apiRoot }}</code>
        <button type="button" class="copy-btn" @click="copy(apiRoot)">复制</button>
      </div>
      <p class="text-xs text-neutral-500 mt-2">第三方客户端将请求发到：<strong>{{ apiRoot }}</strong>，后接下表中的路径（以 <code class="mono-inline">/api</code> 开头）。</p>
    </div>

    <div class="doc-card">
      <h2 class="section-title">认证</h2>
      <p class="text-sm text-neutral-600 mb-3">除「公开」与「初始化」外，请求须带请求头：</p>
      <pre class="code-block">Authorization: Bearer &lt;令牌&gt;</pre>
      <ul class="list-dots text-sm text-neutral-700">
        <li><strong>JWT</strong>：通过 <code class="mono-inline">POST /api/login</code> 取得，含用户身份，与网页登录一致。</li>
        <li><strong>个人访问令牌（PAT）</strong>：在后台先登录网页，在「开放 API 说明」同账号下通过 <code class="mono-inline">POST /api/tokens</code> 创建（需 JWT），仅用于调用图片等接口，与 JWT 二选一即可。</li>
      </ul>
    </div>

    <div class="doc-card" v-for="(group, gi) in groups" :key="gi">
      <h2 class="section-title">{{ group.title }}</h2>
      <p v-if="group.note" class="text-sm text-neutral-500 mb-3">{{ group.note }}</p>
      <div class="overflow-x-auto">
        <table class="api-table">
          <thead>
            <tr>
              <th>方法</th>
              <th>路径</th>
              <th>说明</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(row, i) in group.rows" :key="i">
              <td>
                <span :class="['method-tag', 'm-' + row.method]">{{ row.method }}</span>
              </td>
              <td>
                <code class="path-code">{{ row.path }}</code>
              </td>
              <td class="text-sm text-neutral-700">{{ row.desc }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <div class="doc-card">
      <h2 class="section-title">调用示例</h2>
      <p class="text-sm text-neutral-600 mb-2">登录（返回 JSON 中含 <code class="mono-inline">data.token</code>）：</p>
      <pre class="code-block">{{ exampleLogin }}</pre>
      <p class="text-sm text-neutral-600 mb-2 mt-4">上传图片（<code class="mono-inline">multipart</code> 字段名 <code class="mono-inline">image</code>，可选 <code class="mono-inline">strategy_id</code>）：</p>
      <pre class="code-block">{{ exampleUpload }}</pre>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useMessage } from '@/composables/useMessage'

const { toast } = useMessage()
const origin = ref('')

onMounted(() => {
  origin.value = typeof window !== 'undefined' ? window.location.origin : ''
})

const base = import.meta.env.VITE_API_BASE
const apiRoot = computed(() => {
  if (base && String(base).trim() !== '') {
    return String(base).replace(/\/$/, '')
  }
  return origin.value ? `${origin.value}/api` : '/api'
})

const exampleLogin = computed(() => {
  const o = origin.value || 'https://你的域名'
  return `curl -s -X POST "${o}/api/login" \\\
  -H "Content-Type: application/json" \\\
  -d '{"username":"你的账号","password":"你的密码"}'`
})

const exampleUpload = computed(() => {
  const o = origin.value || 'https://你的域名'
  return `curl -X POST "${o}/api/image/upload" \\\
  -H "Authorization: Bearer <JWT或PAT>" \\\
  -F "image=@/path/to/photo.png"`
})

async function copy(text: string) {
  try {
    await navigator.clipboard.writeText(text)
    toast.success('已复制')
  } catch {
    toast.error('复制失败')
  }
}

type Row = { method: string; path: string; desc: string }
type Group = { title: string; note?: string; rows: Row[] }

const groups: Group[] = [
  {
    title: '初始化与公开（无 Bearer）',
    rows: [
      { method: 'GET', path: '/api/init/status', desc: '是否已初始化、站长账号名（admin_account）' },
      { method: 'POST', path: '/api/init/setup', desc: '首次部署初始化（仅未初始化时）' },
      { method: 'POST', path: '/api/init/test-db', desc: '测试数据库连接（向导用）' },
      { method: 'GET', path: '/api/base_config', desc: '公开基础配置项列表' },
      { method: 'POST', path: '/api/login', desc: '登录，body: { username, password }' },
      { method: 'POST', path: '/api/register', desc: '个人图床会返回拒绝（不开放注册）' },
      { method: 'POST', path: '/api/share/:code', desc: '按分享码取分享内容（:code 为路径参数）' },
    ],
  },
  {
    title: '图片（JWT 或 PAT）',
    note: '需 Authorization: Bearer。上传为 multipart，文件字段名 image；支持 upload-v2、URL 转存等。',
    rows: [
      { method: 'POST', path: '/api/image/upload', desc: '表单上传单图' },
      { method: 'POST', path: '/api/image/upload-v2', desc: '上传 v2' },
      { method: 'POST', path: '/api/image/upload-url', desc: '从 URL 拉取图片' },
      { method: 'POST', path: '/api/image/upload-url-v2', desc: 'URL 转存 v2' },
      { method: 'GET', path: '/api/image/list', desc: '图片列表' },
      { method: 'DELETE', path: '/api/image/:id', desc: '删除单张' },
      { method: 'POST', path: '/api/image/batch-delete', desc: '批量删除' },
      { method: 'PUT', path: '/api/image/:id/rename', desc: '重命名' },
      { method: 'GET', path: '/api/image/:id/qrcode', desc: '图片二维码' },
      { method: 'GET', path: '/api/image/:id/qrcode-base64', desc: '二维码 Base64' },
      { method: 'GET', path: '/api/image/config', desc: '用户组/上传相关配置' },
      { method: 'GET', path: '/api/image/rate-limit', desc: '限流信息（展示用）' },
    ],
  },
  {
    title: '用户（需 JWT）',
    note: '个人中心类接口，使用登录获得的 JWT，一般不用 PAT。',
    rows: [
      { method: 'GET', path: '/api/user/info', desc: '当前用户与策略列表' },
      { method: 'GET', path: '/api/user/config', desc: '用户配置' },
      { method: 'PUT', path: '/api/user/info', desc: '更新资料、偏好、密码等' },
      { method: 'GET', path: '/api/user/dashboard', desc: '仪表盘用量' },
    ],
  },
  {
    title: '个人访问令牌（需 JWT）',
    rows: [
      { method: 'POST', path: '/api/tokens', desc: '创建 PAT，body: { "api_name": "名称" }' },
      { method: 'GET', path: '/api/tokens', desc: '列出 PAT' },
      { method: 'DELETE', path: '/api/tokens/:id', desc: '删除 PAT' },
    ],
  },
  {
    title: '分享（需 JWT）',
    rows: [
      { method: 'POST', path: '/api/shares', desc: '创建分享' },
      { method: 'GET', path: '/api/shares/my', desc: '我的分享' },
      { method: 'DELETE', path: '/api/shares/:id', desc: '删除分享' },
      { method: 'PUT', path: '/api/shares/:id/status', desc: '更新分享状态' },
    ],
  },
  {
    title: '管理员（需 JWT 且管理员账号）',
    rows: [
      { method: 'GET', path: '/api/admin/configs', desc: '系统配置' },
      { method: 'PUT', path: '/api/admin/configs', desc: '更新系统配置' },
      { method: 'GET', path: '/api/admin/upload-policy', desc: '全局上传策略' },
      { method: 'PUT', path: '/api/admin/upload-policy', desc: '更新上传策略' },
      { method: 'GET', path: '/api/admin/stats', desc: '系统统计' },
    ],
  },
  {
    title: '存储策略（需 JWT 且管理员）',
    rows: [
      { method: 'GET', path: '/api/storages', desc: '存储概览' },
      { method: 'GET', path: '/api/storages/strategies', desc: '策略列表' },
      { method: 'POST', path: '/api/storages/strategies', desc: '创建策略' },
      { method: 'GET', path: '/api/storages/strategies/:id', desc: '策略详情' },
      { method: 'PUT', path: '/api/storages/strategies/:id', desc: '更新策略' },
      { method: 'DELETE', path: '/api/storages/strategies/:id', desc: '删除策略' },
    ],
  },
]
</script>

<style scoped>
.api-reference {
  max-width: 1000px;
  margin: 0 auto;
  padding: 0 0 32px;
}

.page-header {
  margin-bottom: 24px;
}

.doc-card {
  background: white;
  border-radius: 12px;
  padding: 24px 28px;
  margin-bottom: 20px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.06);
}

.section-title {
  font-size: 1.1rem;
  font-weight: 600;
  color: #1f2937;
  margin-bottom: 12px;
  border-bottom: 1px solid #e5e7eb;
  padding-bottom: 8px;
}

.mono-inline {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 0.9em;
  background: #f3f4f6;
  padding: 0 4px;
  border-radius: 4px;
}

.code-line {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  background: #0f172a;
  color: #e2e8f0;
  padding: 12px 14px;
  border-radius: 8px;
  font-size: 13px;
}

.code-line code {
  flex: 1;
  min-width: 0;
  word-break: break-all;
}

.copy-btn {
  flex-shrink: 0;
  font-size: 12px;
  padding: 4px 10px;
  background: #334155;
  color: #f8fafc;
  border: none;
  border-radius: 6px;
  cursor: pointer;
}
.copy-btn:hover {
  background: #475569;
}

.code-block {
  background: #0f172a;
  color: #e2e8f0;
  padding: 14px 16px;
  border-radius: 8px;
  font-size: 12px;
  line-height: 1.5;
  overflow-x: auto;
  white-space: pre-wrap;
  word-break: break-all;
}

.list-dots {
  margin: 0;
  padding-left: 1.2rem;
}
.list-dots li {
  margin: 6px 0;
}

.api-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}
.api-table th,
.api-table td {
  text-align: left;
  padding: 10px 12px;
  border-bottom: 1px solid #e5e7eb;
  vertical-align: top;
}
.api-table th {
  background: #f9fafb;
  font-weight: 600;
  color: #4b5563;
}

.path-code {
  font-size: 12px;
  color: #0f172a;
  background: #f1f5f9;
  padding: 2px 6px;
  border-radius: 4px;
  word-break: break-all;
}

.method-tag {
  display: inline-block;
  min-width: 4rem;
  text-align: center;
  font-size: 11px;
  font-weight: 700;
  padding: 2px 8px;
  border-radius: 4px;
  color: #fff;
}
.m-GET { background: #059669; }
.m-POST { background: #2563eb; }
.m-PUT { background: #d97706; }
.m-DELETE { background: #dc2626; }
</style>
