<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import VChart from 'vue-echarts'
import { use } from 'echarts/core'
import { LineChart } from 'echarts/charts'
import {
  TitleComponent,
  TooltipComponent,
  GridComponent,
  DataZoomComponent,
  LegendComponent,
} from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'
import type { Report } from '@/types'
import { listReports } from '@/api/reports'
import { useDarkMode } from '@/composables/useDarkMode'

use([LineChart, TitleComponent, TooltipComponent, GridComponent, DataZoomComponent, LegendComponent, CanvasRenderer])

const props = withDefaults(defineProps<{
  siteId: string
  groups?: { name: string; checkConfigIds: string[] }[]
  checkConfigIds?: string[]
  clientName?: string
  granularity?: number
  showControls?: boolean
}>(), {
  showControls: true,
})

// Palette for multiple clients
const PALETTE = [
  '#5470c6', '#91cc75', '#fac858', '#ee6666', '#73c0de',
  '#3ba272', '#fc8452', '#9a60b4', '#ea7ccc', '#48b8d0',
]

interface MergedPoint {
  timeframe: string
  checks: number
  successes: number
  uptime: number
  avg_delay: number
}

const seriesData = ref<{ name: string; color: string; data: MergedPoint[] }[]>([])
const allTimeframes = ref<string[]>([])
const internalGranularity = ref<number>(0)
const isControlled = computed(() => props.granularity !== undefined)
const effectiveGranularity = computed(() => isControlled.value ? props.granularity! : internalGranularity.value)

const { isDark } = useDarkMode()

const mergeReports = async (configIds: string[]): Promise<MergedPoint[]> => {
  const results = await Promise.all(
    configIds.map((ccId) =>
      listReports({
        site_id: props.siteId,
        type: effectiveGranularity.value,
        check_config_id: ccId,
        page: 1,
        page_size: 50,
      }).catch(() => [] as Report[]),
    ),
  )

  const merged = new Map<string, { checks: number; successes: number; avg_delay: number }>()
  for (const arr of results) {
    for (const r of arr) {
      const existing = merged.get(r.timeframe)
      if (existing) {
        const totalChecks = existing.checks + r.checks
        existing.avg_delay =
          totalChecks > 0
            ? (existing.avg_delay * existing.checks + r.avg_delay * r.checks) / totalChecks
            : 0
        existing.checks = totalChecks
        existing.successes += r.successes
      } else {
        merged.set(r.timeframe, {
          checks: r.checks,
          successes: r.successes,
          avg_delay: r.avg_delay,
        })
      }
    }
  }

  return Array.from(merged.entries())
    .map(([timeframe, data]) => ({
      timeframe,
      checks: data.checks,
      successes: data.successes,
      uptime: data.checks > 0 ? (data.successes / data.checks) * 100 : 0,
      avg_delay: data.avg_delay,
    }))
    .sort((a, b) => a.timeframe.localeCompare(b.timeframe))
}

const loadData = async () => {
  try {
    if (props.groups && props.groups.length > 0) {
      // Multi-client mode: fetch per group
      const results = await Promise.all(
        props.groups.map((g, i) =>
          mergeReports(g.checkConfigIds).then((data) => ({
            name: g.name,
            color: PALETTE[i % PALETTE.length]!,
            data,
          })),
        ),
      )
      seriesData.value = results

      // Compute union of all timeframes
      const tfSet = new Set<string>()
      for (const s of results) {
        for (const p of s.data) tfSet.add(p.timeframe)
      }
      allTimeframes.value = Array.from(tfSet).sort()
    } else {
      // Single-line mode (backward compat)
      const configIds =
        props.checkConfigIds && props.checkConfigIds.length > 0 ? props.checkConfigIds : ['']
      const merged = await mergeReports(configIds)
      const name = props.clientName || 'Uptime'
      seriesData.value = [{ name, color: '#22c55e', data: merged }]
      allTimeframes.value = merged.map((p) => p.timeframe)
    }
  } catch {
    seriesData.value = []
    allTimeframes.value = []
  }
}

onMounted(loadData)
watch([() => props.groups, () => props.checkConfigIds, () => props.granularity], loadData)

const option = computed(() => {
  const textColor = isDark.value ? '#b0b0b0' : '#333'
  const gridColor = isDark.value ? '#333' : '#e5e7eb'
  return {
  tooltip: {
    trigger: 'axis' as const,
  },
  legend: {
    data: seriesData.value.map((s) => s.name),
    bottom: 0,
    textStyle: { color: textColor },
  },
  grid: {
    left: '3%',
    right: '4%',
    bottom: '12%',
    containLabel: true,
  },
  dataZoom: [
    {
      type: 'slider' as const,
      start: 0,
      end: 100,
      textStyle: { color: textColor },
    },
  ],
  xAxis: {
    type: 'category' as const,
    data: allTimeframes.value,
    axisLine: { lineStyle: { color: gridColor } },
    axisLabel: { color: textColor },
    splitLine: { show: false },
  },
  yAxis: {
    type: 'value' as const,
    min: 0,
    max: 100,
    axisLabel: {
      formatter: '{value}%',
      color: textColor,
    },
    splitLine: { lineStyle: { color: gridColor } },
  },
  series: seriesData.value.map((s) => ({
    name: s.name,
    type: 'line' as const,
    data: allTimeframes.value.map((tf) => {
      const pt = s.data.find((p) => p.timeframe === tf)
      return pt ? pt.uptime : null
    }),
    smooth: true,
    areaStyle: {
      opacity: 0.15,
    },
    itemStyle: {
      color: s.color,
    },
  })),
}})
</script>

<template>
  <div>
    <div v-if="showControls && !isControlled" style="margin-bottom: 12px;">
      <el-radio-group v-model="internalGranularity" @change="loadData" size="small">
        <el-radio-button :value="0">Hourly</el-radio-button>
        <el-radio-button :value="1">Daily</el-radio-button>
        <el-radio-button :value="2">Monthly</el-radio-button>
      </el-radio-group>
    </div>
    <v-chart :option="option" style="height: 300px;" autoresize />
  </div>
</template>
