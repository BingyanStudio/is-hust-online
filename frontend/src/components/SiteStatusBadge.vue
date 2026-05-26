<script setup lang="ts">
defineProps<{
  status: number
}>()

const label = (status: number) => {
  switch (status) {
    case 0:
      return 'Enabled'
    case 1:
      return 'Disabled'
    default:
      return 'Unknown'
  }
}
</script>

<template>
  <el-tag :type="status === 0 ? 'success' : 'danger'" size="small" effect="dark">
    {{ label(status) }}
  </el-tag>
</template>
