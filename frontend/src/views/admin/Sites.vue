<script setup lang="ts">
import { loggedIn } from '@/api/client'
import { createSite, deleteSite, listSites, updateSite } from '@/api/sites'
import SiteStatusBadge from '@/components/SiteStatusBadge.vue'
import type { PaginatedResponse, Site } from '@/types'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ref, watch } from 'vue'

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

watch(loggedIn, (v) => { if (v) fetchData() }, { immediate: true })

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
