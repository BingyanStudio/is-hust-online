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
    const res = await listSites({ page: 1, page_size: 20 })
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
