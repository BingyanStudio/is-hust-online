<script setup lang="ts">
import { useDarkMode } from '@/composables/useDarkMode'

const { isDark, toggle } = useDarkMode()
</script>

<template>
  <button
    @click="toggle"
    :title="isDark ? '切到白昼模式' : '切到黑暗模式'"
    style="
      position: fixed;
      bottom: 20px;
      right: 16px;
      z-index: 9999;
      width: 36px;
      height: 36px;
      border-radius: 50%;
      border: 1px solid var(--border);
      background: var(--bg-card);
      color: var(--text-primary);
      cursor: pointer;
      font-size: 18px;
      display: flex;
      align-items: center;
      justify-content: center;
      transition: background 0.3s, border-color 0.3s;
      line-height: 1;
      box-shadow: 0 2px 8px rgba(0,0,0,0.15);
    "
  >
    <span v-if="isDark">&#9728;</span>
    <span v-else>&#9790;</span>
  </button>
</template>
