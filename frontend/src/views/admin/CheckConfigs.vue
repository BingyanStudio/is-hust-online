<script setup lang="ts">
import { createCheckConfig, deleteCheckConfig, listCheckConfigs, updateCheckConfig } from '@/api/checkConfigs'
import { loggedIn } from '@/api/client'
import { listClients } from '@/api/clients'
import { listSites } from '@/api/sites'
import type { CheckConfig, Client, PaginatedResponse, Site } from '@/types'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ref, watch } from 'vue'
import { useAutoRefresh } from '@/composables/useAutoRefresh'

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
      listSites({ page: 1, page_size: 50 }),
      listClients({ page: 1, page_size: 50 }),
    ])
    sites.value = siteRes.items
    clients.value = clientRes.items
  } catch {
    // ignore
  }
}

watch(loggedIn, (v) => {
  if (v) {
    fetchData()
    loadOptions()
  }
}, { immediate: true })
useAutoRefresh(fetchData)

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
