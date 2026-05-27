<script setup lang="ts">
import ThemeToggle from '@/components/ThemeToggle.vue'
</script>

<template>
  <ThemeToggle />
  <router-view />
</template>

<style>
:root {
  --bg-primary: #ffffff;
  --bg-secondary: #f8fafc;
  --bg-card: #ffffff;
  --border: #e5e7eb;
  --border-light: #f3f4f6;
  --text-primary: #111827;
  --text-secondary: #374151;
  --text-muted: #666666;
  --text-dim: #888888;
  --text-faint: #999999;
}

.dark {
  --bg-primary: #121212;
  --bg-secondary: #1e1e1e;
  --bg-card: #1e1e1e;
  --border: #333333;
  --border-light: #2a2a2a;
  --text-primary: #e0e0e0;
  --text-secondary: #b0b0b0;
  --text-muted: #909090;
  --text-dim: #707070;
  --text-faint: #505050;
}

body {
  margin: 0;
  font-family: system-ui, -apple-system, sans-serif;
  background: var(--bg-primary);
  color: var(--text-primary);
  transition: background 0.3s, color 0.3s;
}
</style>
