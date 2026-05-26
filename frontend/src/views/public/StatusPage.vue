<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import type { Site, Report, Check } from '@/types'
import { listSites } from '@/api/sites'
import { listReports } from '@/api/reports'
import { listChecks } from '@/api/checks'
import SiteStatusBadge from '@/components/SiteStatusBadge.vue'

const router = useRouter()
const sites = ref<Site[]>([])
const loading = ref(false)

interface SiteInfo {
  site: Site
  monthlyUptime: number
  recentChecks: Check[]
}

const siteInfos = ref<Map<string, SiteInfo>>(new Map())

onMounted(async () => {
  loading.value = true
  try {
    const res = await listSites({ page: 1, page_size: 50 })
    sites.value = res.items

    await Promise.all(
      sites.value.map(async (site) => {
        const [reports, checksRes] = await Promise.all([
          listReports({ site_id: site.id, type: 1, page: 1, page_size: 1 }).catch(() => []),
          listChecks({ site_id: site.id, page: 1, page_size: 10 }).catch(() => ({ items: [] as Check[] })),
        ])
        const monthlyUptime = reports.length > 0 ? reports[0]!.uptime : 0
        siteInfos.value.set(site.id, {
          site,
          monthlyUptime,
          recentChecks: checksRes.items,
        })
      }),
    )
  } catch {
    sites.value = []
  } finally {
    loading.value = false
  }
})

const groupedSites = computed(() => {
  const groups = new Map<string, Site[]>()
  for (const site of sites.value) {
    const type = site.type || 'Other'
    if (!groups.has(type)) groups.set(type, [])
    groups.get(type)!.push(site)
  }
  return groups
})

const goToDetail = (id: string) => {
  router.push(`/${id}`)
}

const checkBarColor = (check: Check) => {
  if (check.status === 0) return '#22c55e'
  return '#ef4444'
}

const slaColor = (uptime: number) => {
  if (uptime >= 99.9) return '#22c55e'
  if (uptime >= 99) return '#f59e0b'
  return '#ef4444'
}
</script>

<template>
  <div style="max-width: 800px; margin: 40px auto; padding: 0 20px; font-family: system-ui, sans-serif;">
    <h1 style="font-size: 24px; font-weight: 600; margin-bottom: 8px;">Is HUST Online?</h1>
    <p style="color: #666; margin-bottom: 32px;">Real-time status monitoring for HUST websites</p>

    <div v-loading="loading">
      <template v-if="sites.length > 0">
        <div v-for="[type, groupSites] in groupedSites" :key="type" style="margin-bottom: 32px;">
          <h2 style="font-size: 15px; font-weight: 600; color: #374151; text-transform: uppercase; letter-spacing: 0.5px; margin-bottom: 12px; padding-bottom: 8px; border-bottom: 1px solid #e5e7eb;">
            {{ type }}
          </h2>

          <div
            v-for="site in groupSites"
            :key="site.id"
            @click="goToDetail(site.id)"
            style="border: 1px solid #e5e7eb; border-radius: 8px; padding: 16px; margin-bottom: 8px; cursor: pointer; transition: background 0.15s;"
            @mouseenter="($event.currentTarget as HTMLElement).style.background = '#f9fafb'"
            @mouseleave="($event.currentTarget as HTMLElement).style.background = 'white'"
          >
            <div style="display: flex; align-items: center; gap: 12px; margin-bottom: 10px;">
              <SiteStatusBadge :status="site.status" />
              <div style="flex: 1;">
                <div style="font-weight: 500;">{{ site.name }}</div>
                <div style="font-size: 12px; color: #888;">{{ site.url }}</div>
              </div>
              <div v-if="siteInfos.get(site.id)" style="text-align: right;">
                <div style="font-size: 12px; color: #888;">Monthly SLA</div>
                <div style="font-size: 18px; font-weight: 600;" :style="{ color: slaColor(siteInfos.get(site.id)!.monthlyUptime) }">
                  {{ siteInfos.get(site.id)!.monthlyUptime.toFixed(2) }}%
                </div>
              </div>
            </div>

            <div v-if="siteInfos.get(site.id)" style="display: flex; gap: 3px; align-items: flex-end; height: 28px;">
              <div
                v-for="(check, idx) in siteInfos.get(site.id)!.recentChecks"
                :key="idx"
                :style="{
                  width: '100%',
                  flex: 1,
                  height: check.status === 0 ? '100%' : '60%',
                  background: checkBarColor(check),
                  borderRadius: '2px',
                  opacity: check.status === 0 ? 0.8 : 1,
                }"
                :title="`${new Date(check.timestamp * 1000).toLocaleString()} - ${check.status === 0 ? 'OK' : 'Error'}`"
              />
              <div
                v-if="siteInfos.get(site.id)!.recentChecks.length === 0"
                style="flex: 1; height: 4px; background: #e5e7eb; border-radius: 2px;"
              />
            </div>
          </div>
        </div>
      </template>

      <p v-if="!loading && sites.length === 0" style="color: #999; text-align: center; padding: 40px 0;">
        No sites configured yet.
      </p>
    </div>
  </div>
</template>
