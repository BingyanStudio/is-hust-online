<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useAutoRefresh } from '@/composables/useAutoRefresh'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { Client, PaginatedResponse } from '@/types'
import { listClients, createClient, updateClient, deleteClient } from '@/api/clients'
import { loggedIn } from '@/api/client'

const data = ref<Client[]>([])
const loading = ref(false)
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)

const capabilityOptions = [
  { value: 1, label: 'HTTP' },
  { value: 2, label: 'PING' },
  { value: 4, label: 'TCP' },
  { value: 8, label: 'Other' },
]

const dialogVisible = ref(false)
const editingClient = ref<Client | null>(null)
const form = ref({ name: '', location: '', capabilities: 0, labels: [] as string[], status: 1 })
const capabilityChecks = ref<number[]>([])

const syncCapabilitiesToChecks = () => {
  capabilityChecks.value = capabilityOptions
    .filter((o) => (form.value.capabilities & o.value) !== 0)
    .map((o) => o.value)
}

const syncChecksToCapabilities = () => {
  form.value.capabilities = capabilityChecks.value.reduce((sum, v) => sum + v, 0)
}

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

watch(loggedIn, (v) => { if (v) fetchData() }, { immediate: true })
useAutoRefresh(fetchData)

const openCreate = () => {
  editingClient.value = null
  form.value = { name: '', location: '', capabilities: 0, labels: [], status: 1 }
  syncCapabilitiesToChecks()
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
  syncCapabilitiesToChecks()
  dialogVisible.value = true
}

const handleSubmit = async () => {
  try {
    if (editingClient.value) {
      await updateClient(editingClient.value.id, form.value)
      ElMessage.success('Client updated')
    } else {
      const client = await createClient(form.value)
      ElMessage.success('Client created')
      ElMessageBox.alert(
        `The client token is: ${client.token}`,
        'Client Created',
        {
          confirmButtonText: 'Copy Token',
          callback: () => copyToClipboard(client.token),
        },
      )
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
  if (s === 1) return 'Online'
  if (s === 4) return 'Offline'
  return 'Disabled'
}

const statusType = (s: number): 'success' | 'warning' | 'danger' => {
  if (s === 1) return 'success'
  if (s === 4) return 'warning'
  return 'danger'
}

const formatTime = (ts: number) => ts ? new Date(ts * 1000).toLocaleString() : '-'

const copyToClipboard = async (text: string) => {
  try {
    await navigator.clipboard.writeText(text)
    ElMessage.success('Copied to clipboard')
  } catch {
    ElMessage.error('Failed to copy')
  }
}
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
      <el-table-column prop="ip" label="IP" width="210" />
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
        <el-form-item v-if="editingClient" label="Token">
          <div style="display: flex; gap: 8px;">
            <el-input :model-value="editingClient.token" readonly style="flex: 1;" />
            <el-button @click="copyToClipboard(editingClient!.token)">Copy</el-button>
          </div>
        </el-form-item>
        <el-form-item label="Capabilities">
          <el-checkbox-group v-model="capabilityChecks" @change="syncChecksToCapabilities">
            <el-checkbox v-for="opt in capabilityOptions" :key="opt.value" :value="opt.value">
              {{ opt.label }}
            </el-checkbox>
          </el-checkbox-group>
        </el-form-item>
        <el-form-item label="Labels">
          <el-select
            v-model="form.labels"
            multiple
            filterable
            allow-create
            default-first-option
            placeholder="Add labels..."
            style="width: 100%;"
          />
        </el-form-item>
        <el-form-item label="Status" v-if="editingClient">
          <el-select v-model="form.status">
            <el-option :value="1" label="Online" />
            <el-option :value="4" label="Offline" />
            <el-option :value="5" label="Disabled" />
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
