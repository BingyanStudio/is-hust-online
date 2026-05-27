<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import type { Check, PaginatedResponse, Client } from '@/types'
import { listChecks } from '@/api/checks'

const props = defineProps<{
  siteId: string
  clientId?: string
  clientName?: string
  clients?: Record<string, Client>
}>()

const data = ref<Check[]>([])
const loading = ref(false)
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)

const fetchData = async () => {
  loading.value = true
  try {
    const params: { site_id: string; client_id?: string; page: number; page_size: number } = {
      site_id: props.siteId,
      page: page.value,
      page_size: pageSize.value,
    }
    if (props.clientId) {
      params.client_id = props.clientId
    }
    const res: PaginatedResponse<Check> = await listChecks(params as any)
    data.value = res.items
    total.value = res.paging.total_pages * pageSize.value
  } catch {
    data.value = []
  } finally {
    loading.value = false
  }
}

onMounted(fetchData)
watch(() => props.clientId, fetchData)

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

const getClientName = (clientId: string) => {
  if (props.clients && props.clients[clientId]) {
    return props.clients[clientId]!.name
  }
  return clientId.substring(0, 8) + '...'
}
</script>

<template>
  <div>
    <div v-if="clientName" style="font-size: 14px; font-weight: 500; color: #666; margin-bottom: 8px;">
      {{ clientName }}
    </div>
    <el-table :data="data" v-loading="loading" stripe>
      <el-table-column label="Time" width="200">
        <template #default="{ row }">{{ formatTime(row.timestamp) }}</template>
      </el-table-column>
      <el-table-column v-if="!clientId && clients" label="Client" width="120">
        <template #default="{ row }">
          <span style="font-size: 12px;">{{ getClientName(row.client_id) }}</span>
        </template>
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
  </div>
</template>
