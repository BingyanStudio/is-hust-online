<script setup lang="ts">
import { listCheckConfigs } from '@/api/checkConfigs'
import { listChecks } from '@/api/checks'
import { listClients } from '@/api/clients'
import { listReports } from '@/api/reports'
import { listSites } from '@/api/sites'
import SiteStatusBadge from '@/components/SiteStatusBadge.vue'
import type { Check, Client, Report, Site } from '@/types'
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()
const sites = ref<Site[]>([])
const searchQuery = ref('')
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

const filteredSites = computed(() => {
  if (!searchQuery.value.trim()) return sites.value
  const q = searchQuery.value.toLowerCase()
  return sites.value.filter(
    (s) => s.name.toLowerCase().includes(q) || s.url.toLowerCase().includes(q),
  )
})

const groupedSites = computed(() => {
  const groups: Record<string, Site[]> = {}
  for (const site of filteredSites.value) {
    const type = site.type || 'Other'
    if (!groups[type]) groups[type] = []
    groups[type].push(site)
  }
  return groups
})

const typeKeys = computed(() => Object.keys(groupedSites.value))

const scrollToType = (type: string) => {
  document.getElementById(`type-${type}`)?.scrollIntoView({ behavior: 'smooth', block: 'start' })
}

const siteStatuses = computed(() => {
  const online = new Set<string>()
  const offline = new Set<string>()

  for (const site of sites.value) {
    const clientKeys = getSiteClientKeys(site.id)
    if (clientKeys.length === 0) {
      offline.add(site.id)
      continue
    }
    const hasSuccess = clientKeys.some((key) => {
      const info = clientSiteInfos.value[key]
      return info?.recentChecks.some((c) => c.status === 0)
    })
    if (hasSuccess) {
      online.add(site.id)
    } else {
      offline.add(site.id)
    }
  }
  return { online, offline }
})

const overallStats = computed(() => {
  const total = sites.value.length
  const onlineCount = siteStatuses.value.online.size
  const offlineCount = total - onlineCount

  let slaSum = 0
  let slaCount = 0
  for (const info of Object.values(clientSiteInfos.value)) {
    if (info.recentChecks.length > 0) {
      slaSum += info.monthlyUptime
      slaCount++
    }
  }
  const overallSLA = slaCount > 0 ? slaSum / slaCount : 0

  return { total, onlineCount, offlineCount, overallSLA }
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
    <h1 style="font-size: 24px; font-weight: 600; margin-bottom: 8px;">华中大在线吗？</h1>
    <p style="color: #666; margin-bottom: 32px;">实时检测华科各类网络服务状态</p>

    <!-- Global stats -->
    <div v-if="sites.length > 0" style="display: flex; flex-wrap: wrap; gap: 16px 24px; margin-bottom: 24px; padding: 16px 20px; background: #f8fafc; border-radius: 8px; border: 1px solid #e5e7eb;">
      <div style="text-align: center; flex: 1; min-width: 60px;">
        <div style="font-size: 24px; font-weight: 700;">{{ overallStats.total }}</div>
        <div style="font-size: 12px; color: #888;">总数</div>
      </div>
      <div style="text-align: center; flex: 1; min-width: 60px;">
        <div style="font-size: 24px; font-weight: 700; color: #22c55e;">{{ overallStats.onlineCount }}</div>
        <div style="font-size: 12px; color: #888;">在线</div>
      </div>
      <div style="text-align: center; flex: 1; min-width: 60px;">
        <div style="font-size: 24px; font-weight: 700; color: #ef4444;">{{ overallStats.offlineCount }}</div>
        <div style="font-size: 12px; color: #888;">离线</div>
      </div>
      <div style="text-align: center; flex: 1; min-width: 60px;">
        <div style="font-size: 24px; font-weight: 700;" :style="{ color: slaColor(overallStats.overallSLA) }">
          {{ overallStats.overallSLA.toFixed(1) }}%
        </div>
        <div style="font-size: 12px; color: #888;">SLA</div>
      </div>
    </div>

    <div v-loading="loading">
      <!-- Search -->
      <el-input
        v-model="searchQuery"
        placeholder="搜索站点..."
        clearable
        style="margin-bottom: 16px;"
      >
        <template #prefix>
          <span style="color: #a8abb2;">&#128269;</span>
        </template>
      </el-input>

      <!-- Type navigation -->
      <div v-if="typeKeys.length > 1" style="display: flex; gap: 8px; flex-wrap: wrap; margin-bottom: 24px;">
        <el-tag
          v-for="type in typeKeys"
          :key="type"
          type="primary"
          style="cursor: pointer;"
          @click="scrollToType(type)"
        >
          {{ type }}
        </el-tag>
      </div>

      <template v-if="Object.keys(groupedSites).length > 0">
        <div v-for="(groupSites, type) in groupedSites" :key="type" :id="'type-' + type" style="margin-bottom: 32px;">
          <h2 style="font-size: 15px; font-weight: 600; color: #374151; text-transform: uppercase; letter-spacing: 0.5px; margin-bottom: 12px; padding-bottom: 8px; border-bottom: 1px solid #e5e7eb;">
            {{ type }}
          </h2>

          <div
            v-for="site in groupSites"
            :key="site.id"
                        style="border: 1px solid #e5e7eb; border-radius: 8px; padding: 16px; margin-bottom: 8px; overflow: hidden; min-width: 0;"
          >
            <!-- Site header: clickable -->
            <div
              @click="goToDetail(site.id)"
              style="display: flex; align-items: center; gap: 12px; margin-bottom: 12px; cursor: pointer;"
            >
              <SiteStatusBadge :status="site.status" />
              <div style="flex: 1; min-width: 0;">
                <div style="font-weight: 500; overflow-wrap: break-word;">{{ site.name }}</div>
                <div style="font-size: 12px; color: #888; overflow-wrap: break-word; word-break: break-all;">{{ site.url }}</div>
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
                  <span style="font-size: 13px; font-weight: 500; color: #4b5563; flex: 1; min-width: 0; overflow-wrap: break-word;">
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
              <p style="font-size: 12px; color: #999; margin-bottom: 8px;">还没有配置客户端</p>
            </div>
          </div>
        </div>
      </template>

      <p v-if="!loading && sites.length === 0" style="color: #999; text-align: center; padding: 40px 0;">
        还没有配置站点
      </p>
      <p v-if="!loading && sites.length > 0 && Object.keys(groupedSites).length === 0" style="color: #999; text-align: center; padding: 40px 0;">
        没有匹配您搜索的站点
      </p>
    </div>
  </div>
</template>
