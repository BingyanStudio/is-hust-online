<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import type { Site, Check, CheckConfig, Client, Report } from '@/types'
import { listSites } from '@/api/sites'
import { listReports } from '@/api/reports'
import { listChecks } from '@/api/checks'
import { listCheckConfigs } from '@/api/checkConfigs'
import { listClients } from '@/api/clients'
import SiteStatusBadge from '@/components/SiteStatusBadge.vue'

const router = useRouter()
const sites = ref<Site[]>([])
const loading = ref(false)

// Per-client data
interface ClientSiteInfo {
  clientName: string
  monthlyUptime: number
  recentChecks: Check[]
}

// Key: "siteId:clientId"
const clientSiteInfos = ref<Record<string, ClientSiteInfo>>({})

// check_configs grouped by site_id -> client_id -> check_config_ids
const siteClientCCs = ref<Record<string, Record<string, string[]>>>({})

onMounted(async () => {
  loading.value = true
  try {
    const [sitesRes, ccRes, clientsRes] = await Promise.all([
      listSites({ page: 1, page_size: 50 }),
      listCheckConfigs({ page: 1, page_size: 500 }),
      listClients({ page: 1, page_size: 200 }),
    ])
    sites.value = sitesRes.items

    // Build client map
    const clientMap: Record<string, Client> = {}
    for (const c of clientsRes.items) {
      clientMap[c.id] = c
    }

    // Group check_configs: site_id -> client_id -> [check_config_ids]
    const grouping: Record<string, Record<string, string[]>> = {}
    for (const cc of ccRes.items) {
      if (!grouping[cc.site_id]) grouping[cc.site_id] = {}
      if (!grouping[cc.site_id]![cc.client_id]) grouping[cc.site_id]![cc.client_id] = []
      grouping[cc.site_id]![cc.client_id]!.push(cc.id)
    }
    siteClientCCs.value = grouping

    // Fetch per-client data for all sites
    const infoPromises: Promise<void>[] = []
    for (const site of sites.value) {
      const clientGroups = grouping[site.id]
      if (!clientGroups) continue

      for (const [clientId, ccIds] of Object.entries(clientGroups)) {
        const key = `${site.id}:${clientId}`
        const client = clientMap[clientId]
        const clientName = client ? client.name : clientId.substring(0, 8) + '...'

        infoPromises.push(
          (async () => {
            try {
              // Fetch reports for each check_config and merge
              const reportsArr = await Promise.all(
                ccIds.map((ccId) =>
                  listReports({ site_id: site.id, type: 1, check_config_id: ccId, page: 1, page_size: 1 })
                    .catch(() => [] as Report[]),
                ),
              )
              let combinedChecks = 0
              let combinedSuccesses = 0
              for (const arr of reportsArr) {
                if (arr.length > 0) {
                  combinedChecks += arr[0]!.checks
                  combinedSuccesses += arr[0]!.successes
                }
              }
              const uptime = combinedChecks > 0 ? (combinedSuccesses / combinedChecks) * 100 : 0

              // Fetch recent checks
              const checksRes = await listChecks({
                site_id: site.id,
                client_id: clientId,
                page: 1,
                page_size: 10,
              })

              clientSiteInfos.value[key] = {
                clientName,
                monthlyUptime: uptime,
                recentChecks: checksRes.items,
              }
            } catch {
              clientSiteInfos.value[key] = { clientName, monthlyUptime: 0, recentChecks: [] }
            }
          })(),
        )
      }
    }
    await Promise.all(infoPromises)
  } catch {
    sites.value = []
  } finally {
    loading.value = false
  }
})

const groupedSites = computed(() => {
  const groups: Record<string, Site[]> = {}
  for (const site of sites.value) {
    const type = site.type || 'Other'
    if (!groups[type]) groups[type] = []
    groups[type].push(site)
  }
  return groups
})

const getSiteClientKeys = (siteId: string): string[] => {
  const groups = siteClientCCs.value[siteId]
  if (!groups) return []
  return Object.keys(groups).map((clientId) => `${siteId}:${clientId}`)
}

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
        <div v-for="(groupSites, type) in groupedSites" :key="type" style="margin-bottom: 32px;">
          <h2 style="font-size: 15px; font-weight: 600; color: #374151; text-transform: uppercase; letter-spacing: 0.5px; margin-bottom: 12px; padding-bottom: 8px; border-bottom: 1px solid #e5e7eb;">
            {{ type }}
          </h2>

          <div
            v-for="site in groupSites"
            :key="site.id"
            style="border: 1px solid #e5e7eb; border-radius: 8px; padding: 16px; margin-bottom: 8px;"
          >
            <!-- Site header: clickable -->
            <div
              @click="goToDetail(site.id)"
              style="display: flex; align-items: center; gap: 12px; margin-bottom: 12px; cursor: pointer;"
            >
              <SiteStatusBadge :status="site.status" />
              <div style="flex: 1;">
                <div style="font-weight: 500;">{{ site.name }}</div>
                <div style="font-size: 12px; color: #888;">{{ site.url }}</div>
              </div>
            </div>

            <!-- Per-client stats -->
            <div v-if="getSiteClientKeys(site.id).length > 0">
              <div
                v-for="clientKey in getSiteClientKeys(site.id)"
                :key="clientKey"
                style="padding: 8px 0; border-top: 1px solid #f3f4f6;"
              >
                <div style="display: flex; align-items: center; justify-content: space-between; margin-bottom: 6px;">
                  <span style="font-size: 13px; font-weight: 500; color: #4b5563;">
                    {{ clientSiteInfos[clientKey]?.clientName || '...' }}
                  </span>
                  <div v-if="clientSiteInfos[clientKey]" style="text-align: right;">
                    <span style="font-size: 11px; color: #888;">SLA</span>
                    <span style="font-size: 14px; font-weight: 600; margin-left: 4px;" :style="{ color: slaColor(clientSiteInfos[clientKey]!.monthlyUptime) }">
                      {{ clientSiteInfos[clientKey]!.monthlyUptime.toFixed(1) }}%
                    </span>
                  </div>
                </div>
                <!-- Recent check bars -->
                <div v-if="clientSiteInfos[clientKey]" style="display: flex; gap: 3px; align-items: flex-end; height: 20px;">
                  <div
                    v-for="(check, idx) in clientSiteInfos[clientKey]!.recentChecks"
                    :key="idx"
                    :style="{
                      width: '100%',
                      flex: 1,
                      height: check.status === 0 ? '100%' : '50%',
                      background: checkBarColor(check),
                      borderRadius: '2px',
                      opacity: check.status === 0 ? 0.8 : 1,
                    }"
                    :title="`${new Date(check.timestamp * 1000).toLocaleString()} - ${check.status === 0 ? 'OK' : 'Error'}`"
                  />
                  <div
                    v-if="clientSiteInfos[clientKey] && clientSiteInfos[clientKey]!.recentChecks.length === 0"
                    style="flex: 1; height: 4px; background: #e5e7eb; border-radius: 2px;"
                  />
                </div>
              </div>
            </div>

            <!-- Fallback when no clients -->
            <div v-else style="padding-top: 8px; border-top: 1px solid #f3f4f6;">
              <p style="font-size: 12px; color: #999; margin-bottom: 8px;">No clients configured</p>
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
