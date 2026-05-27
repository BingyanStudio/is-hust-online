import axios from 'axios'
import { ref } from 'vue'

const AUTH_KEY = 'admin_auth'

let authUsername = ''
let authPassword = ''
let onUnauthorized: (() => void) | null = null

export const loggedIn = ref(false)

// Restore persisted auth (verified in a previous session)
try {
  const saved = localStorage.getItem(AUTH_KEY)
  if (saved) {
    const parsed = JSON.parse(saved)
    if (parsed.username && parsed.password) {
      authUsername = parsed.username
      authPassword = parsed.password
      loggedIn.value = true
    }
  }
} catch {}

export function setAuth(username: string, password: string) {
  authUsername = username
  authPassword = password
  if (username && password) {
    localStorage.setItem(AUTH_KEY, JSON.stringify({ username, password }))
  } else {
    localStorage.removeItem(AUTH_KEY)
    loggedIn.value = false
  }
}

export function markLoggedIn() {
  loggedIn.value = true
}

export function clearAuth() {
  authUsername = ''
  authPassword = ''
  loggedIn.value = false
  localStorage.removeItem(AUTH_KEY)
}

export function isAuthenticated() {
  return authUsername !== '' && authPassword !== ''
}

export function on401(handler: () => void) {
  onUnauthorized = handler
}

const api = axios.create({
  baseURL: '/api',
})

api.interceptors.request.use((config) => {
  if (authUsername && authPassword) {
    config.auth = { username: authUsername, password: authPassword }
  }
  return config
})

api.interceptors.response.use(
  (res) => {
    if (res.data.code !== 0) {
      return Promise.reject(new Error(res.data.message || 'Request failed'))
    }
    return res.data.data
  },
  (error) => {
    if (error.response?.status === 401) {
      clearAuth()
      onUnauthorized?.()
    }
    return Promise.reject(error)
  },
)

export default api
