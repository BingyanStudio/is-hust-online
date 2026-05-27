<script setup lang="ts">
import { listCheckConfigs } from '@/api/checkConfigs'
import { listClients } from '@/api/clients'
import { getSite } from '@/api/sites'
import CheckHistoryTable from '@/components/CheckHistoryTable.vue'
import LatencyChart from '@/components/LatencyChart.vue'
import SiteStatusBadge from '@/components/SiteStatusBadge.vue'
import UptimeChart from '@/components/UptimeChart.vue'
import type { Client, Site } from '@/types'
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()
const site = ref<Site | null>(null)
const loading = ref(true)
const id = route.params.id as string

const clientsMap = ref<Record<string, Client>>({})
const activeTab = ref<string>('')

interface ClientGroup {
  client: Client
  checkConfigIds: string[]
}

const clientGroups = ref<ClientGroup[]>([])
const chartGranularity = ref<number>(0)

onMounted(async () => {
  try {
    site.value = await getSite(id)

    const ccRes = await listCheckConfigs({ site_id: id, page: 1, page_size: 100 })

    const ccByClient = new Map<string, string[]>()
    for (const cc of ccRes.items) {
      const list = ccByClient.get(cc.client_id) || []
      list.push(cc.id)
      ccByClient.set(cc.client_id, list)
    }

    if (ccByClient.size > 0) {
      const allClients = await listClients({ page: 1, page_size: 200 })
      for (const c of allClients.items) {
        clientsMap.value[c.id] = c
      }

      for (const [clientId, ccIds] of ccByClient.entries()) {
        const client = clientsMap.value[clientId]
        if (client) {
          clientGroups.value.push({ client, checkConfigIds: ccIds })
        }
      }

      if (clientGroups.value.length > 0) {
        activeTab.value = clientGroups.value[0]!.client.id
      }
    }
  } catch {
    site.value = null
  } finally {
    loading.value = false
  }
})

const clientStatusLabel = (status: number) => {
  if (status === 1) return 'Online'
  if (status === 4) return 'Offline'
  return 'Disabled'
}

const clientStatusColor = (status: number) => {
  if (status === 1) return '#22c55e'
  if (status === 4) return '#f59e0b'
  return '#9ca3af'
}
</script>

<template>
  <div style="max-width: 960px; margin: 40px auto; padding: 0 20px; font-family: system-ui, sans-serif;">
    <el-button text @click="router.push('/')" style="margin-bottom: 16px;">
      &larr; 回到状态页
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

        <template v-if="clientGroups.length > 0">
          <!-- Shared granularity control for both charts -->
          <div style="margin-bottom: 16px;">
            <el-radio-group v-model="chartGranularity" size="small">
              <el-radio-button :value="0">每小时</el-radio-button>
              <el-radio-button :value="1">每日</el-radio-button>
              <el-radio-button :value="2">每月</el-radio-button>
            </el-radio-group>
          </div>

          <!-- Charts: all clients overlaid -->
          <h2 style="font-size: 16px; font-weight: 600; margin-bottom: 12px;">在线状态</h2>
          <UptimeChart
            :site-id="id"
            :groups="clientGroups.map(g => ({ name: g.client.name, checkConfigIds: g.checkConfigIds }))"
            :granularity="chartGranularity"
            :show-controls="false"
          />

          <h2 style="font-size: 16px; font-weight: 600; margin: 24px 0 12px;">延迟</h2>
          <LatencyChart
            :site-id="id"
            :groups="clientGroups.map(g => ({ name: g.client.name, checkConfigIds: g.checkConfigIds }))"
            :granularity="chartGranularity"
            :show-controls="false"
          />

          <!-- Client tabs: info + check records -->
          <h2 style="font-size: 16px; font-weight: 600; margin: 24px 0 12px;">最近检测</h2>
          <el-tabs v-model="activeTab">
            <el-tab-pane
              v-for="group in clientGroups"
              :key="group.client.id"
              :label="group.client.name"
              :name="group.client.id"
            >
              <div style="display: flex; align-items: center; gap: 12px; margin-bottom: 16px; padding: 12px 16px; background: var(--bg-secondary); border-radius: 6px;">
                <div>
                  <div style="font-weight: 500; font-size: 15px;">{{ group.client.name }}</div>
                  <div style="font-size: 12px; color: var(--text-dim);">{{ group.client.location || 'No location' }}</div>
                </div>
                <el-tag size="small" :color="clientStatusColor(group.client.status)" style="color: white; border: none;">
                  {{ clientStatusLabel(group.client.status) }}
                </el-tag>
              </div>

              <CheckHistoryTable
                :site-id="id"
                :client-id="group.client.id"
                :client-name="group.client.name"
                :clients="clientsMap"
              />
            </el-tab-pane>
          </el-tabs>
        </template>

        <!-- Fallback: no clients configured -->
        <template v-else>
          <p style="color: var(--text-faint); font-size: 13px; margin-bottom: 20px;">No monitoring clients configured for this site.</p>
          <h2 style="font-size: 16px; font-weight: 600; margin-bottom: 12px;">Uptime</h2>
          <UptimeChart :site-id="id" />

          <h2 style="font-size: 16px; font-weight: 600; margin: 24px 0 12px;">Recent Checks</h2>
          <CheckHistoryTable :site-id="id" :clients="clientsMap" />
        </template>
      </template>

      <el-empty v-if="!loading && !site" description="Site not found" />
    </div>
  </div>
</template>
