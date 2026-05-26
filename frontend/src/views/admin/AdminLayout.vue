<script setup lang="ts">
import { useRouter, useRoute } from 'vue-router'

const router = useRouter()
const route = useRoute()

const navItems = [
  { path: '/admin/sites', label: 'Sites' },
  { path: '/admin/clients', label: 'Clients' },
  { path: '/admin/check-configs', label: 'Check Configs' },
]
</script>

<template>
  <div style="display: flex; min-height: 100vh; font-family: system-ui, sans-serif;">
    <aside style="width: 220px; background: #f8fafc; border-right: 1px solid #e5e7eb; padding: 20px 12px;">
      <div style="font-size: 11px; color: #999; text-transform: uppercase; letter-spacing: 0.5px; margin-bottom: 12px; padding: 0 8px;">
        Admin
      </div>
      <nav>
        <div
          v-for="item in navItems"
          :key="item.path"
          @click="router.push(item.path)"
          style="padding: 8px 12px; border-radius: 6px; cursor: pointer; font-size: 14px; margin-bottom: 2px; transition: background 0.15s;"
          :style="{
            background: route.path === item.path ? '#eff6ff' : 'transparent',
            color: route.path === item.path ? '#2563eb' : '#374151',
            fontWeight: route.path === item.path ? 500 : 400,
          }"
        >
          {{ item.label }}
        </div>
      </nav>
    </aside>
    <main style="flex: 1; padding: 24px 32px;">
      <router-view />
    </main>
  </div>
</template>
