<script setup lang="ts">
import { ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { setAuth, markLoggedIn, isAuthenticated, on401, clearAuth } from '@/api/client'
import { listSites } from '@/api/sites'

const router = useRouter()
const route = useRoute()

const navItems = [
  { path: '/admin/sites', label: 'Sites' },
  { path: '/admin/clients', label: 'Clients' },
  { path: '/admin/check-configs', label: 'Check Configs' },
]

const loginVisible = ref(!isAuthenticated())
const loginForm = ref({ username: '', password: '' })

on401(() => {
  loginVisible.value = true
})

const loginLoading = ref(false)

const handleLogin = async () => {
  if (!loginForm.value.username || !loginForm.value.password) {
    ElMessage.warning('Please enter username and password')
    return
  }
  loginLoading.value = true
  try {
    setAuth(loginForm.value.username, loginForm.value.password)
    await listSites({ page: 1, page_size: 1 })
    markLoggedIn()
    loginVisible.value = false
    ElMessage.success('Logged in')
  } catch {
    clearAuth()
    ElMessage.error('Invalid username or password')
  } finally {
    loginLoading.value = false
  }
}

const handleCancel = () => {
  router.push('/')
}
</script>

<template>
  <el-dialog v-model="loginVisible" title="Admin Login" :close-on-click-modal="false" :close-on-press-escape="false" :show-close="false" width="380px">
    <el-form label-position="top">
      <el-form-item label="Username">
        <el-input v-model="loginForm.username" autofocus />
      </el-form-item>
      <el-form-item label="Password">
        <el-input v-model="loginForm.password" type="password" show-password @keyup.enter="handleLogin" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="handleCancel">Cancel</el-button>
      <el-button type="primary" :loading="loginLoading" @click="handleLogin">Login</el-button>
    </template>
  </el-dialog>

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
