<script setup lang="ts">
defineProps<{
  status: number
}>()

const label = (status: number) => {
  switch (status) {
    case 0:
      return '已启用'
    case 1:
      return '已禁用'
    default:
      return '未知'
  }
}
</script>

<template>
  <el-tag :type="status === 0 ? 'success' : 'danger'" size="small" effect="dark">
    {{ label(status) }}
  </el-tag>
</template>
