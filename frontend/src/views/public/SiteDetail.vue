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
