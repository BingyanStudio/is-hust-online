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
