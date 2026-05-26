# is-hust-online Vue 3 Frontend Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a Vue 3 frontend with a public uptime status page and an admin CRUD dashboard for the is-hust-online monitoring system.

**Architecture:** Two-section SPA — public pages at `/` (status list, site detail with ECharts uptime chart) and admin pages at `/admin` (sidebar layout with Sites/Clients/CheckConfigs CRUD tables). API calls go through an Axios instance with response envelope unwrapping. Browser-native Basic Auth for admin mutations.

**Tech Stack:** Vue 3, TypeScript, Vite, vue-router, Element Plus, ECharts (vue-echarts), Axios

---

## File Structure

| File | Responsibility |
|---|---|
| `frontend/package.json` | Add dependencies: element-plus, echarts, vue-echarts, axios |
| `frontend/vite.config.ts` | Add `/api` dev proxy to Go backend |
| `frontend/src/types/index.ts` | TypeScript interfaces for all API models |
| `frontend/src/api/client.ts` | Axios instance with response interceptor |
| `frontend/src/api/sites.ts` | Site CRUD API functions |
| `frontend/src/api/clients.ts` | Client CRUD API functions |
| `frontend/src/api/checkConfigs.ts` | CheckConfig CRUD API functions |
| `frontend/src/api/checks.ts` | Check list API (read-only) |
| `frontend/src/api/reports.ts` | Report list API (read-only) |
| `frontend/src/router/index.ts` | Route definitions with lazy-loaded views |
| `frontend/src/components/SiteStatusBadge.vue` | Green/red status indicator dot |
| `frontend/src/components/UptimeChart.vue` | ECharts line chart for uptime time-series |
| `frontend/src/components/CheckHistoryTable.vue` | Paginated check results table |
| `frontend/src/views/public/StatusPage.vue` | Homepage — list of sites with status |
| `frontend/src/views/public/SiteDetail.vue` | Site detail — chart + check history |
| `frontend/src/views/admin/AdminLayout.vue` | Sidebar nav + router-view |
| `frontend/src/views/admin/Sites.vue` | Sites CRUD table + dialog |
| `frontend/src/views/admin/Clients.vue` | Clients CRUD table + dialog |
| `frontend/src/views/admin/CheckConfigs.vue` | CheckConfigs CRUD table + dialog |
| `frontend/src/main.ts` | Register Element Plus, mount app |
| `frontend/src/App.vue` | Root template with router-view |

---

### Task 1: Install Dependencies and Configure Vite

**Files:**
- Modify: `frontend/package.json`
- Modify: `frontend/vite.config.ts`

- [ ] **Step 1: Install dependencies**

Run from `frontend/`:
```bash
bun add element-plus echarts vue-echarts axios
```

- [ ] **Step 2: Add Vite dev proxy**

Edit `frontend/vite.config.ts` to add proxy configuration:

```typescript
import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueJsx from '@vitejs/plugin-vue-jsx'
import vueDevTools from 'vite-plugin-vue-devtools'

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    vueJsx(),
    vueDevTools(),
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    },
  },
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
})
```

- [ ] **Step 3: Commit**

```bash
cd frontend && git add package.json bun.lock vite.config.ts && git commit -m "chore: add element-plus, echarts, axios deps and vite proxy"
```

---

### Task 2: TypeScript Types

**Files:**
- Create: `frontend/src/types/index.ts`

- [ ] **Step 1: Create type definitions**

Create `frontend/src/types/index.ts` with all API model interfaces:

```typescript
export interface Site {
  id: string
  name: string
  url: string
  type: string
  description: string
  logo: string
  status: number // 0=enabled, 1=disabled
  created_at: number
}

export interface Client {
  id: string
  name: string
  location: string
  capabilities: number
  status: number // 0=online, 1=offline, 2=disabled
  token: string
  ip: string
  last_online: number
  created_at: number
  labels: string[]
}

export interface CheckConfig {
  id: string
  site_id: string
  client_id: string
  status: number // 0=enabled, 1=disabled
  check_type: number // 0=UNKNOWN, 1=HTTP, 2=PING, 4=TCP, 8=OTHER
  check_interval: string
  check_extra: unknown
}

export interface Check {
  id: string
  site_id: string
  client_id: string
  check_config_id: string
  timestamp: number
  type: number
  status: number // ErrorType enum
  result: string
  delay: number
  extra: unknown
}

export interface Report {
  site_id: string
  check_config_id: string
  timeframe: string
  type: number // 0=hourly, 1=daily, 2=monthly
  checks: number
  successes: number
  uptime: number
  avg_delay: number
}

export interface PagingMeta {
  page: number
  page_size: number
  total_pages: number
  has_more: boolean
}

export interface PaginatedResponse<T> {
  items: T[]
  paging: PagingMeta
}

export interface CreateSiteRequest {
  name: string
  url: string
  type?: string
  description?: string
  logo?: string
}

export interface UpdateSiteRequest {
  name?: string
  url?: string
  type?: string
  description?: string
  status?: number
  logo?: string
}

export interface CreateClientRequest {
  name: string
  location?: string
  capabilities?: number
  labels?: string[]
}

export interface UpdateClientRequest {
  name?: string
  location?: string
  capabilities?: number
  labels?: string[]
  status?: number
}

export interface CreateCheckConfigRequest {
  site_id: string
  client_id: string
  check_type: number
  check_interval: string
  check_extra?: unknown
}

export interface UpdateCheckConfigRequest {
  client_id?: string
  check_type?: number
  check_interval?: string
  check_extra?: unknown
  status?: number
}

export const CheckType = {
  UNKNOWN: 0,
  HTTP: 1,
  PING: 2,
  TCP: 4,
  OTHER: 8,
} as const

export const ErrorType = {
  NO_ERROR: 0,
  HTTP_UNREACHABLE: 1,
  HTTP_TIMEOUT: 2,
  HTTP_OTHER: 3,
  PING_TIMEOUT: 4,
  PING_UNREACHABLE: 5,
  TCP_TIMEOUT: 6,
  TCP_UNREACHABLE: 7,
} as const

export const ReportType = {
  HOURLY: 0,
  DAILY: 1,
  MONTHLY: 2,
} as const
```

- [ ] **Step 2: Commit**

```bash
cd frontend && git add src/types/index.ts && git commit -m "feat: add TypeScript type definitions for API models"
```

---

### Task 3: API Client and Modules

**Files:**
- Create: `frontend/src/api/client.ts`
- Create: `frontend/src/api/sites.ts`
- Create: `frontend/src/api/clients.ts`
- Create: `frontend/src/api/checkConfigs.ts`
- Create: `frontend/src/api/checks.ts`
- Create: `frontend/src/api/reports.ts`

- [ ] **Step 1: Create Axios client**

Create `frontend/src/api/client.ts`:

```typescript
import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
})

api.interceptors.response.use(
  (res) => {
    if (res.data.code !== 0) {
      return Promise.reject(new Error(res.data.message || 'Request failed'))
    }
    return res.data.data
  },
  (error) => {
    return Promise.reject(error)
  },
)

export default api
```

- [ ] **Step 2: Create Sites API**

Create `frontend/src/api/sites.ts`:

```typescript
import api from './client'
import type {
  Site,
  PaginatedResponse,
  CreateSiteRequest,
  UpdateSiteRequest,
} from '@/types'

export function listSites(params?: { page?: number; page_size?: number }) {
  return api.get<any, PaginatedResponse<Site>>('/sites', { params })
}

export function getSite(id: string) {
  return api.get<any, Site>(`/sites/${id}`)
}

export function createSite(data: CreateSiteRequest) {
  return api.post<any, Site>('/sites', data)
}

export function updateSite(id: string, data: UpdateSiteRequest) {
  return api.put<any, Site>(`/sites/${id}`, data)
}

export function deleteSite(id: string) {
  return api.delete(`/sites/${id}`)
}
```

- [ ] **Step 3: Create Clients API**

Create `frontend/src/api/clients.ts`:

```typescript
import api from './client'
import type {
  Client,
  PaginatedResponse,
  CreateClientRequest,
  UpdateClientRequest,
} from '@/types'

export function listClients(params?: { page?: number; page_size?: number }) {
  return api.get<any, PaginatedResponse<Client>>('/clients', { params })
}

export function getClient(id: string) {
  return api.get<any, Client>(`/clients/${id}`)
}

export function createClient(data: CreateClientRequest) {
  return api.post<any, Client>('/clients', data)
}

export function updateClient(id: string, data: UpdateClientRequest) {
  return api.put<any, Client>(`/clients/${id}`, data)
}

export function deleteClient(id: string) {
  return api.delete(`/clients/${id}`)
}
```

- [ ] **Step 4: Create CheckConfigs API**

Create `frontend/src/api/checkConfigs.ts`:

```typescript
import api from './client'
import type {
  CheckConfig,
  PaginatedResponse,
  CreateCheckConfigRequest,
  UpdateCheckConfigRequest,
} from '@/types'

export function listCheckConfigs(params?: {
  site_id?: string
  client_id?: string
  page?: number
  page_size?: number
}) {
  return api.get<any, PaginatedResponse<CheckConfig>>('/check-configs', { params })
}

export function getCheckConfig(id: string) {
  return api.get<any, CheckConfig>(`/check-configs/${id}`)
}

export function createCheckConfig(data: CreateCheckConfigRequest) {
  return api.post<any, CheckConfig>('/check-configs', data)
}

export function updateCheckConfig(id: string, data: UpdateCheckConfigRequest) {
  return api.put<any, CheckConfig>(`/check-configs/${id}`, data)
}

export function deleteCheckConfig(id: string) {
  return api.delete(`/check-configs/${id}`)
}
```

- [ ] **Step 5: Create Checks API**

Create `frontend/src/api/checks.ts`:

```typescript
import api from './client'
import type { Check, PaginatedResponse } from '@/types'

export function listChecks(params?: {
  site_id?: string
  client_id?: string
  page?: number
  page_size?: number
}) {
  return api.get<any, PaginatedResponse<Check>>('/checks', { params })
}
```

- [ ] **Step 6: Create Reports API**

Create `frontend/src/api/reports.ts`:

```typescript
import api from './client'
import type { Report } from '@/types'

export function listReports(params: {
  site_id: string
  type?: number
  page?: number
  page_size?: number
}) {
  return api.get<any, Report[]>('/reports', { params })
}
```

- [ ] **Step 7: Commit**

```bash
cd frontend && git add src/api/ && git commit -m "feat: add API client and module functions"
```

---

### Task 4: Router Setup

**Files:**
- Modify: `frontend/src/router/index.ts`

- [ ] **Step 1: Define routes**

Replace `frontend/src/router/index.ts`:

```typescript
import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      component: () => import('@/views/public/StatusPage.vue'),
    },
    {
      path: '/:id',
      component: () => import('@/views/public/SiteDetail.vue'),
    },
    {
      path: '/admin',
      component: () => import('@/views/admin/AdminLayout.vue'),
      children: [
        { path: '', redirect: '/admin/sites' },
        {
          path: 'sites',
          component: () => import('@/views/admin/Sites.vue'),
        },
        {
          path: 'clients',
          component: () => import('@/views/admin/Clients.vue'),
        },
        {
          path: 'check-configs',
          component: () => import('@/views/admin/CheckConfigs.vue'),
        },
      ],
    },
  ],
})

export default router
```

- [ ] **Step 2: Commit**

```bash
cd frontend && git add src/router/index.ts && git commit -m "feat: define routes for public and admin sections"
```

---

### Task 5: Shared Components

**Files:**
- Create: `frontend/src/components/SiteStatusBadge.vue`
- Create: `frontend/src/components/UptimeChart.vue`
- Create: `frontend/src/components/CheckHistoryTable.vue`

- [ ] **Step 1: Create SiteStatusBadge**

Create `frontend/src/components/SiteStatusBadge.vue`:

```vue
<script setup lang="ts">
defineProps<{
  status: number
}>()

const label = (status: number) => {
  switch (status) {
    case 0:
      return 'Enabled'
    case 1:
      return 'Disabled'
    default:
      return 'Unknown'
  }
}
</script>

<template>
  <el-tag :type="status === 0 ? 'success' : 'danger'" size="small" effect="dark">
    {{ label(status) }}
  </el-tag>
</template>
```

- [ ] **Step 2: Create UptimeChart**

Create `frontend/src/components/UptimeChart.vue`:

```vue
<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import VChart from 'vue-echarts'
import { use } from 'echarts/core'
import { LineChart } from 'echarts/charts'
import {
  TitleComponent,
  TooltipComponent,
  GridComponent,
  DataZoomComponent,
} from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'
import type { Report } from '@/types'
import { listReports } from '@/api/reports'

use([LineChart, TitleComponent, TooltipComponent, GridComponent, DataZoomComponent, CanvasRenderer])

const props = defineProps<{
  siteId: string
}>()

const reports = ref<Report[]>([])
const granularity = ref<number>(0)

const loadData = async () => {
  try {
    reports.value = await listReports({
      site_id: props.siteId,
      type: granularity.value,
      page_size: 50,
    })
  } catch {
    reports.value = []
  }
}

onMounted(loadData)

const option = computed(() => ({
  tooltip: {
    trigger: 'axis' as const,
  },
  grid: {
    left: '3%',
    right: '4%',
    bottom: '3%',
    containLabel: true,
  },
  dataZoom: [
    {
      type: 'slider' as const,
      start: 0,
      end: 100,
    },
  ],
  xAxis: {
    type: 'category' as const,
    data: reports.value.map((r) => r.timeframe),
  },
  yAxis: {
    type: 'value' as const,
    min: 0,
    max: 100,
    axisLabel: {
      formatter: '{value}%',
    },
  },
  series: [
    {
      name: 'Uptime',
      type: 'line' as const,
      data: reports.value.map((r) => r.uptime),
      smooth: true,
      areaStyle: {
        opacity: 0.15,
      },
      itemStyle: {
        color: '#22c55e',
      },
    },
  ],
}))
</script>

<template>
  <div>
    <div style="margin-bottom: 12px;">
      <el-radio-group v-model="granularity" @change="loadData" size="small">
        <el-radio-button :value="0">Hourly</el-radio-button>
        <el-radio-button :value="1">Daily</el-radio-button>
        <el-radio-button :value="2">Monthly</el-radio-button>
      </el-radio-group>
    </div>
    <v-chart :option="option" style="height: 300px;" autoresize />
  </div>
</template>
```

- [ ] **Step 3: Create CheckHistoryTable**

Create `frontend/src/components/CheckHistoryTable.vue`:

```vue
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import type { Check, PaginatedResponse } from '@/types'
import { listChecks } from '@/api/checks'

const props = defineProps<{
  siteId: string
}>()

const data = ref<Check[]>([])
const loading = ref(false)
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)

const fetchData = async () => {
  loading.value = true
  try {
    const res: PaginatedResponse<Check> = await listChecks({
      site_id: props.siteId,
      page: page.value,
      page_size: pageSize.value,
    })
    data.value = res.items
    total.value = res.paging.total_pages * pageSize.value
  } catch {
    data.value = []
  } finally {
    loading.value = false
  }
}

onMounted(fetchData)

const handlePageChange = (newPage: number) => {
  page.value = newPage
  fetchData()
}

const formatTime = (ts: number) => {
  return new Date(ts * 1000).toLocaleString()
}

const statusLabel = (status: number) => {
  if (status === 0) return 'OK'
  return `Error (${status})`
}
</script>

<template>
  <el-table :data="data" v-loading="loading" stripe>
    <el-table-column label="Time" width="200">
      <template #default="{ row }">{{ formatTime(row.timestamp) }}</template>
    </el-table-column>
    <el-table-column label="Status">
      <template #default="{ row }">
        <el-tag :type="row.status === 0 ? 'success' : 'danger'" size="small">
          {{ statusLabel(row.status) }}
        </el-tag>
      </template>
    </el-table-column>
    <el-table-column label="Delay (ms)">
      <template #default="{ row }">{{ row.delay }}</template>
    </el-table-column>
    <el-table-column label="Type">
      <template #default="{ row }">{{ row.type }}</template>
    </el-table-column>
  </el-table>
  <el-pagination
    style="margin-top: 12px; justify-content: center;"
    :current-page="page"
    :page-size="pageSize"
    :total="total"
    layout="prev, pager, next"
    @current-change="handlePageChange"
  />
</template>
```

- [ ] **Step 4: Commit**

```bash
cd frontend && git add src/components/ && git commit -m "feat: add shared components (SiteStatusBadge, UptimeChart, CheckHistoryTable)"
```

---

### Task 6: Public Pages

**Files:**
- Create: `frontend/src/views/public/StatusPage.vue`
- Create: `frontend/src/views/public/SiteDetail.vue`

- [ ] **Step 1: Create StatusPage**

Create `frontend/src/views/public/StatusPage.vue`:

```vue
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import type { Site } from '@/types'
import { listSites } from '@/api/sites'
import SiteStatusBadge from '@/components/SiteStatusBadge.vue'

const router = useRouter()
const sites = ref<Site[]>([])
const loading = ref(false)

onMounted(async () => {
  loading.value = true
  try {
    const res = await listSites({ page_size: 20 })
    sites.value = res.items
  } catch {
    sites.value = []
  } finally {
    loading.value = false
  }
})

const goToDetail = (id: string) => {
  router.push(`/${id}`)
}
</script>

<template>
  <div style="max-width: 720px; margin: 40px auto; padding: 0 20px; font-family: system-ui, sans-serif;">
    <h1 style="font-size: 24px; font-weight: 600; margin-bottom: 8px;">Is HUST Online?</h1>
    <p style="color: #666; margin-bottom: 24px;">Real-time status monitoring for HUST websites</p>

    <div v-loading="loading">
      <div
        v-for="site in sites"
        :key="site.id"
        @click="goToDetail(site.id)"
        style="display: flex; align-items: center; gap: 12px; padding: 14px 16px; border: 1px solid #e5e7eb; border-radius: 8px; margin-bottom: 8px; cursor: pointer; transition: background 0.15s;"
        @mouseenter="($event.currentTarget as HTMLElement).style.background = '#f9fafb'"
        @mouseleave="($event.currentTarget as HTMLElement).style.background = 'white'"
      >
        <SiteStatusBadge :status="site.status" />
        <div style="flex: 1;">
          <div style="font-weight: 500;">{{ site.name }}</div>
          <div style="font-size: 13px; color: #888;">{{ site.url }}</div>
        </div>
      </div>

      <p v-if="!loading && sites.length === 0" style="color: #999; text-align: center; padding: 40px 0;">
        No sites configured yet.
      </p>
    </div>
  </div>
</template>
```

- [ ] **Step 2: Create SiteDetail**

Create `frontend/src/views/public/SiteDetail.vue`:

```vue
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import type { Site } from '@/types'
import { getSite } from '@/api/sites'
import UptimeChart from '@/components/UptimeChart.vue'
import CheckHistoryTable from '@/components/CheckHistoryTable.vue'
import SiteStatusBadge from '@/components/SiteStatusBadge.vue'

const route = useRoute()
const router = useRouter()
const site = ref<Site | null>(null)
const loading = ref(true)
const id = route.params.id as string

onMounted(async () => {
  try {
    site.value = await getSite(id)
  } catch {
    site.value = null
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div style="max-width: 900px; margin: 40px auto; padding: 0 20px; font-family: system-ui, sans-serif;">
    <el-button text @click="router.push('/')" style="margin-bottom: 16px;">
      &larr; Back to status page
    </el-button>

    <div v-loading="loading">
      <template v-if="site">
        <div style="display: flex; align-items: center; gap: 12px; margin-bottom: 24px;">
          <SiteStatusBadge :status="site.status" />
          <div>
            <h1 style="font-size: 22px; font-weight: 600; margin: 0;">{{ site.name }}</h1>
            <a :href="site.url" target="_blank" style="color: #2563eb; font-size: 14px;">{{ site.url }}</a>
          </div>
        </div>

        <h2 style="font-size: 16px; font-weight: 600; margin-bottom: 12px;">Uptime</h2>
        <UptimeChart :site-id="id" />

        <h2 style="font-size: 16px; font-weight: 600; margin: 24px 0 12px;">Recent Checks</h2>
        <CheckHistoryTable :site-id="id" />
      </template>

      <el-empty v-if="!loading && !site" description="Site not found" />
    </div>
  </div>
</template>
```

- [ ] **Step 3: Commit**

```bash
cd frontend && git add src/views/public/ && git commit -m "feat: add public status page and site detail views"
```

---

### Task 7: Admin Layout and CRUD Pages

**Files:**
- Create: `frontend/src/views/admin/AdminLayout.vue`
- Create: `frontend/src/views/admin/Sites.vue`
- Create: `frontend/src/views/admin/Clients.vue`
- Create: `frontend/src/views/admin/CheckConfigs.vue`

- [ ] **Step 1: Create AdminLayout**

Create `frontend/src/views/admin/AdminLayout.vue`:

```vue
<script setup lang="ts">
import { useRouter, useRoute } from 'vue-router'

const router = useRouter()
const route = useRoute()

const navItems = [
  { path: '/admin/sites', label: 'Sites' },
  { path: '/admin/clients', label: 'Clients' },
  { path: '/admin/check-configs', label: 'Check Configs' },
]
</script>

<template>
  <div style="display: flex; min-height: 100vh; font-family: system-ui, sans-serif;">
    <aside style="width: 220px; background: #f8fafc; border-right: 1px solid #e5e7eb; padding: 20px 12px;">
      <div style="font-size: 11px; color: #999; text-transform: uppercase; letter-spacing: 0.5px; margin-bottom: 12px; padding: 0 8px;">
        Admin
      </div>
      <nav>
        <div
          v-for="item in navItems"
          :key="item.path"
          @click="router.push(item.path)"
          style="padding: 8px 12px; border-radius: 6px; cursor: pointer; font-size: 14px; margin-bottom: 2px; transition: background 0.15s;"
          :style="{
            background: route.path === item.path ? '#eff6ff' : 'transparent',
            color: route.path === item.path ? '#2563eb' : '#374151',
            fontWeight: route.path === item.path ? 500 : 400,
          }"
        >
          {{ item.label }}
        </div>
      </nav>
    </aside>
    <main style="flex: 1; padding: 24px 32px;">
      <router-view />
    </main>
  </div>
</template>
```

- [ ] **Step 2: Create Sites CRUD page**

Create `frontend/src/views/admin/Sites.vue`:

```vue
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { Site, PaginatedResponse } from '@/types'
import { listSites, createSite, updateSite, deleteSite } from '@/api/sites'
import SiteStatusBadge from '@/components/SiteStatusBadge.vue'

const data = ref<Site[]>([])
const loading = ref(false)
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)

const dialogVisible = ref(false)
const editingSite = ref<Site | null>(null)
const form = ref({ name: '', url: '', type: '', description: '', logo: '', status: 0 })

const fetchData = async () => {
  loading.value = true
  try {
    const res: PaginatedResponse<Site> = await listSites({ page: page.value, page_size: pageSize.value })
    data.value = res.items
    total.value = res.paging.total_pages * pageSize.value
  } catch (e: any) {
    ElMessage.error(e.message || 'Failed to load sites')
  } finally {
    loading.value = false
  }
}

onMounted(fetchData)

const openCreate = () => {
  editingSite.value = null
  form.value = { name: '', url: '', type: '', description: '', logo: '', status: 0 }
  dialogVisible.value = true
}

const openEdit = (site: Site) => {
  editingSite.value = site
  form.value = {
    name: site.name,
    url: site.url,
    type: site.type,
    description: site.description,
    logo: site.logo,
    status: site.status,
  }
  dialogVisible.value = true
}

const handleSubmit = async () => {
  try {
    if (editingSite.value) {
      await updateSite(editingSite.value.id, form.value)
      ElMessage.success('Site updated')
    } else {
      await createSite(form.value)
      ElMessage.success('Site created')
    }
    dialogVisible.value = false
    fetchData()
  } catch (e: any) {
    ElMessage.error(e.message || 'Failed to save site')
  }
}

const handleDelete = async (site: Site) => {
  try {
    await ElMessageBox.confirm(`Delete site "${site.name}"?`, 'Confirm', { type: 'warning' })
    await deleteSite(site.id)
    ElMessage.success('Site deleted')
    fetchData()
  } catch {
    // cancelled
  }
}

const formatTime = (ts: number) => new Date(ts * 1000).toLocaleDateString()
</script>

<template>
  <div>
    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
      <h2 style="font-size: 18px; font-weight: 600; margin: 0;">Sites</h2>
      <el-button type="primary" @click="openCreate">+ Add Site</el-button>
    </div>

    <el-table :data="data" v-loading="loading" stripe>
      <el-table-column prop="name" label="Name" />
      <el-table-column prop="url" label="URL" />
      <el-table-column label="Status" width="120">
        <template #default="{ row }"><SiteStatusBadge :status="row.status" /></template>
      </el-table-column>
      <el-table-column label="Created" width="140">
        <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
      </el-table-column>
      <el-table-column label="Actions" width="160" align="right">
        <template #default="{ row }">
          <el-button text type="primary" size="small" @click="openEdit(row)">Edit</el-button>
          <el-button text type="danger" size="small" @click="handleDelete(row)">Delete</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-pagination
      style="margin-top: 12px; justify-content: center;"
      :current-page="page"
      :page-size="pageSize"
      :total="total"
      layout="prev, pager, next"
      @current-change="(p: number) => { page = p; fetchData() }"
    />

    <el-dialog v-model="dialogVisible" :title="editingSite ? 'Edit Site' : 'Create Site'" width="500px">
      <el-form label-position="top">
        <el-form-item label="Name" required>
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="URL" required>
          <el-input v-model="form.url" />
        </el-form-item>
        <el-form-item label="Type">
          <el-input v-model="form.type" />
        </el-form-item>
        <el-form-item label="Description">
          <el-input v-model="form.description" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item label="Logo URL">
          <el-input v-model="form.logo" />
        </el-form-item>
        <el-form-item label="Status" v-if="editingSite">
          <el-select v-model="form.status">
            <el-option :value="0" label="Enabled" />
            <el-option :value="1" label="Disabled" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">Cancel</el-button>
        <el-button type="primary" @click="handleSubmit">{{ editingSite ? 'Save' : 'Create' }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>
```

- [ ] **Step 3: Create Clients CRUD page**

Create `frontend/src/views/admin/Clients.vue`:

```vue
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { Client, PaginatedResponse } from '@/types'
import { listClients, createClient, updateClient, deleteClient } from '@/api/clients'

const data = ref<Client[]>([])
const loading = ref(false)
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)

const dialogVisible = ref(false)
const editingClient = ref<Client | null>(null)
const form = ref({ name: '', location: '', capabilities: 0, labels: [] as string[], status: 0 })

const fetchData = async () => {
  loading.value = true
  try {
    const res: PaginatedResponse<Client> = await listClients({ page: page.value, page_size: pageSize.value })
    data.value = res.items
    total.value = res.paging.total_pages * pageSize.value
  } catch (e: any) {
    ElMessage.error(e.message || 'Failed to load clients')
  } finally {
    loading.value = false
  }
}

onMounted(fetchData)

const openCreate = () => {
  editingClient.value = null
  form.value = { name: '', location: '', capabilities: 0, labels: [], status: 0 }
  dialogVisible.value = true
}

const openEdit = (client: Client) => {
  editingClient.value = client
  form.value = {
    name: client.name,
    location: client.location,
    capabilities: client.capabilities,
    labels: [...client.labels],
    status: client.status,
  }
  dialogVisible.value = true
}

const handleSubmit = async () => {
  try {
    if (editingClient.value) {
      await updateClient(editingClient.value.id, form.value)
      ElMessage.success('Client updated')
    } else {
      await createClient(form.value)
      ElMessage.success('Client created')
    }
    dialogVisible.value = false
    fetchData()
  } catch (e: any) {
    ElMessage.error(e.message || 'Failed to save client')
  }
}

const handleDelete = async (client: Client) => {
  try {
    await ElMessageBox.confirm(`Delete client "${client.name}"?`, 'Confirm', { type: 'warning' })
    await deleteClient(client.id)
    ElMessage.success('Client deleted')
    fetchData()
  } catch {
    // cancelled
  }
}

const statusLabel = (s: number) => {
  if (s === 0) return 'Online'
  if (s === 1) return 'Offline'
  return 'Disabled'
}

const statusType = (s: number): 'success' | 'warning' | 'danger' => {
  if (s === 0) return 'success'
  if (s === 1) return 'warning'
  return 'danger'
}

const formatTime = (ts: number) => ts ? new Date(ts * 1000).toLocaleString() : '-'
</script>

<template>
  <div>
    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
      <h2 style="font-size: 18px; font-weight: 600; margin: 0;">Clients</h2>
      <el-button type="primary" @click="openCreate">+ Add Client</el-button>
    </div>

    <el-table :data="data" v-loading="loading" stripe>
      <el-table-column prop="name" label="Name" />
      <el-table-column prop="location" label="Location" />
      <el-table-column label="Status" width="120">
        <template #default="{ row }">
          <el-tag :type="statusType(row.status)" size="small">{{ statusLabel(row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="Labels">
        <template #default="{ row }">
          <el-tag v-for="label in row.labels" :key="label" size="small" style="margin-right: 4px;">{{ label }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="Last Online" width="180">
        <template #default="{ row }">{{ formatTime(row.last_online) }}</template>
      </el-table-column>
      <el-table-column label="Actions" width="160" align="right">
        <template #default="{ row }">
          <el-button text type="primary" size="small" @click="openEdit(row)">Edit</el-button>
          <el-button text type="danger" size="small" @click="handleDelete(row)">Delete</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-pagination
      style="margin-top: 12px; justify-content: center;"
      :current-page="page"
      :page-size="pageSize"
      :total="total"
      layout="prev, pager, next"
      @current-change="(p: number) => { page = p; fetchData() }"
    />

    <el-dialog v-model="dialogVisible" :title="editingClient ? 'Edit Client' : 'Create Client'" width="500px">
      <el-form label-position="top">
        <el-form-item label="Name" required>
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="Location">
          <el-input v-model="form.location" />
        </el-form-item>
        <el-form-item label="Capabilities (bitmask)">
          <el-input-number v-model="form.capabilities" :min="0" />
        </el-form-item>
        <el-form-item label="Status" v-if="editingClient">
          <el-select v-model="form.status">
            <el-option :value="0" label="Online" />
            <el-option :value="1" label="Offline" />
            <el-option :value="2" label="Disabled" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">Cancel</el-button>
        <el-button type="primary" @click="handleSubmit">{{ editingClient ? 'Save' : 'Create' }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>
```

- [ ] **Step 4: Create CheckConfigs CRUD page**

Create `frontend/src/views/admin/CheckConfigs.vue`:

```vue
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { CheckConfig, Site, Client, PaginatedResponse } from '@/types'
import { listCheckConfigs, createCheckConfig, updateCheckConfig, deleteCheckConfig } from '@/api/checkConfigs'
import { listSites } from '@/api/sites'
import { listClients } from '@/api/clients'

const data = ref<CheckConfig[]>([])
const sites = ref<Site[]>([])
const clients = ref<Client[]>([])
const loading = ref(false)
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)

const dialogVisible = ref(false)
const editingConfig = ref<CheckConfig | null>(null)
const form = ref({
  site_id: '',
  client_id: '',
  check_type: 1,
  check_interval: '',
  check_extra: '',
  status: 0,
})

const fetchData = async () => {
  loading.value = true
  try {
    const res: PaginatedResponse<CheckConfig> = await listCheckConfigs({ page: page.value, page_size: pageSize.value })
    data.value = res.items
    total.value = res.paging.total_pages * pageSize.value
  } catch (e: any) {
    ElMessage.error(e.message || 'Failed to load check configs')
  } finally {
    loading.value = false
  }
}

const loadOptions = async () => {
  try {
    const [siteRes, clientRes] = await Promise.all([
      listSites({ page_size: 50 }),
      listClients({ page_size: 50 }),
    ])
    sites.value = siteRes.items
    clients.value = clientRes.items
  } catch {
    // ignore
  }
}

onMounted(() => {
  fetchData()
  loadOptions()
})

const siteName = (id: string) => sites.value.find((s) => s.id === id)?.name || id
const clientName = (id: string) => clients.value.find((c) => c.id === id)?.name || id

const openCreate = () => {
  editingConfig.value = null
  form.value = { site_id: '', client_id: '', check_type: 1, check_interval: '', check_extra: '', status: 0 }
  dialogVisible.value = true
}

const openEdit = (config: CheckConfig) => {
  editingConfig.value = config
  form.value = {
    site_id: config.site_id,
    client_id: config.client_id,
    check_type: config.check_type,
    check_interval: config.check_interval,
    check_extra: String(config.check_extra || ''),
    status: config.status,
  }
  dialogVisible.value = true
}

const handleSubmit = async () => {
  try {
    if (editingConfig.value) {
      const { site_id: _, client_id: __, ...updateData } = form.value
      await updateCheckConfig(editingConfig.value.id, updateData)
      ElMessage.success('Check config updated')
    } else {
      await createCheckConfig({
        site_id: form.value.site_id,
        client_id: form.value.client_id,
        check_type: form.value.check_type,
        check_interval: form.value.check_interval,
        check_extra: form.value.check_extra || undefined,
      })
      ElMessage.success('Check config created')
    }
    dialogVisible.value = false
    fetchData()
  } catch (e: any) {
    ElMessage.error(e.message || 'Failed to save check config')
  }
}

const handleDelete = async (config: CheckConfig) => {
  try {
    await ElMessageBox.confirm('Delete this check config?', 'Confirm', { type: 'warning' })
    await deleteCheckConfig(config.id)
    ElMessage.success('Check config deleted')
    fetchData()
  } catch {
    // cancelled
  }
}

const checkTypeLabel = (t: number) => {
  const map: Record<number, string> = { 0: 'Unknown', 1: 'HTTP', 2: 'PING', 4: 'TCP', 8: 'Other' }
  return map[t] || 'Unknown'
}
</script>

<template>
  <div>
    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
      <h2 style="font-size: 18px; font-weight: 600; margin: 0;">Check Configs</h2>
      <el-button type="primary" @click="openCreate">+ Add Check Config</el-button>
    </div>

    <el-table :data="data" v-loading="loading" stripe>
      <el-table-column label="Site">
        <template #default="{ row }">{{ siteName(row.site_id) }}</template>
      </el-table-column>
      <el-table-column label="Client">
        <template #default="{ row }">{{ clientName(row.client_id) }}</template>
      </el-table-column>
      <el-table-column label="Type" width="100">
        <template #default="{ row }">{{ checkTypeLabel(row.check_type) }}</template>
      </el-table-column>
      <el-table-column prop="check_interval" label="Interval" />
      <el-table-column label="Status" width="100">
        <template #default="{ row }">
          <el-tag :type="row.status === 0 ? 'success' : 'info'" size="small">
            {{ row.status === 0 ? 'Enabled' : 'Disabled' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="Actions" width="160" align="right">
        <template #default="{ row }">
          <el-button text type="primary" size="small" @click="openEdit(row)">Edit</el-button>
          <el-button text type="danger" size="small" @click="handleDelete(row)">Delete</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-pagination
      style="margin-top: 12px; justify-content: center;"
      :current-page="page"
      :page-size="pageSize"
      :total="total"
      layout="prev, pager, next"
      @current-change="(p: number) => { page = p; fetchData() }"
    />

    <el-dialog v-model="dialogVisible" :title="editingConfig ? 'Edit Check Config' : 'Create Check Config'" width="500px">
      <el-form label-position="top">
        <el-form-item label="Site" required>
          <el-select v-model="form.site_id" filterable placeholder="Select site" :disabled="!!editingConfig">
            <el-option v-for="s in sites" :key="s.id" :value="s.id" :label="s.name" />
          </el-select>
        </el-form-item>
        <el-form-item label="Client" required>
          <el-select v-model="form.client_id" filterable placeholder="Select client" :disabled="!!editingConfig">
            <el-option v-for="c in clients" :key="c.id" :value="c.id" :label="c.name" />
          </el-select>
        </el-form-item>
        <el-form-item label="Check Type" required>
          <el-select v-model="form.check_type">
            <el-option :value="1" label="HTTP" />
            <el-option :value="2" label="PING" />
            <el-option :value="4" label="TCP" />
            <el-option :value="8" label="Other" />
          </el-select>
        </el-form-item>
        <el-form-item label="Check Interval (cron)" required>
          <el-input v-model="form.check_interval" placeholder="e.g. */5 * * * *" />
        </el-form-item>
        <el-form-item label="Status" v-if="editingConfig">
          <el-select v-model="form.status">
            <el-option :value="0" label="Enabled" />
            <el-option :value="1" label="Disabled" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">Cancel</el-button>
        <el-button type="primary" @click="handleSubmit">{{ editingConfig ? 'Save' : 'Create' }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>
```

- [ ] **Step 5: Commit**

```bash
cd frontend && git add src/views/admin/ && git commit -m "feat: add admin layout and CRUD pages for sites, clients, check configs"
```

---

### Task 8: Wire Up main.ts and App.vue

**Files:**
- Modify: `frontend/src/main.ts`
- Modify: `frontend/src/App.vue`

- [ ] **Step 1: Update main.ts**

Replace `frontend/src/main.ts`:

```typescript
import { createApp } from 'vue'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import App from './App.vue'
import router from './router'

const app = createApp(App)

app.use(router)
app.use(ElementPlus)

app.mount('#app')
```

- [ ] **Step 2: Update App.vue**

Replace `frontend/src/App.vue`:

```vue
<script setup lang="ts"></script>

<template>
  <router-view />
</template>

<style>
body {
  margin: 0;
  font-family: system-ui, -apple-system, sans-serif;
  color: #1f2937;
}
</style>
```

- [ ] **Step 3: Commit**

```bash
cd frontend && git add src/main.ts src/App.vue && git commit -m "feat: wire up Element Plus and root App component"
```

---

### Task 9: Verify Build

**Files:** None (verification only)

- [ ] **Step 1: Install dependencies**

Run from `frontend/`:
```bash
bun install
```

- [ ] **Step 2: Run type check**

```bash
cd frontend && bun run type-check
```

Expected: No errors. Fix any TypeScript errors that appear.

- [ ] **Step 3: Run build**

```bash
cd frontend && bun run build
```

Expected: Build succeeds with no errors.

- [ ] **Step 4: Final commit (if any fixes were needed)**

```bash
cd frontend && git add -A && git commit -m "fix: resolve type-check and build errors"
```
