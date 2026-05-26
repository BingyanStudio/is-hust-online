<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import VChart from 'vue-echarts'
import { use } from 'echarts/core'
import { LineChart } from 'echarts/charts'
import {
  TitleComponent,
  TooltipComponent,
  GridComponent,
  DataZoomComponent,
} from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'
import type { Report } from '@/types'
import { listReports } from '@/api/reports'

use([LineChart, TitleComponent, TooltipComponent, GridComponent, DataZoomComponent, CanvasRenderer])

const props = defineProps<{
  siteId: string
}>()

const reports = ref<Report[]>([])
const granularity = ref<number>(0)

const loadData = async () => {
  try {
    reports.value = await listReports({
      site_id: props.siteId,
      type: granularity.value,
      page: 1,
      page_size: 50,
    })
  } catch {
    reports.value = []
  }
}

onMounted(loadData)

const option = computed(() => ({
  tooltip: {
    trigger: 'axis' as const,
  },
  grid: {
    left: '3%',
    right: '4%',
    bottom: '3%',
    containLabel: true,
  },
  dataZoom: [
    {
      type: 'slider' as const,
      start: 0,
      end: 100,
    },
  ],
  xAxis: {
    type: 'category' as const,
    data: reports.value.map((r) => r.timeframe),
  },
  yAxis: {
    type: 'value' as const,
    min: 0,
    max: 100,
    axisLabel: {
      formatter: '{value}%',
    },
  },
  series: [
    {
      name: 'Uptime',
      type: 'line' as const,
      data: reports.value.map((r) => r.uptime),
      smooth: true,
      areaStyle: {
        opacity: 0.15,
      },
      itemStyle: {
        color: '#22c55e',
      },
    },
  ],
}))
</script>

<template>
  <div>
    <div style="margin-bottom: 12px;">
      <el-radio-group v-model="granularity" @change="loadData" size="small">
        <el-radio-button :value="0">Hourly</el-radio-button>
        <el-radio-button :value="1">Daily</el-radio-button>
        <el-radio-button :value="2">Monthly</el-radio-button>
      </el-radio-group>
    </div>
    <v-chart :option="option" style="height: 300px;" autoresize />
  </div>
</template>
