import axios from 'axios'

let authUsername = ''
let authPassword = ''
let onUnauthorized: (() => void) | null = null

export function setAuth(username: string, password: string) {
  authUsername = username
  authPassword = password
}

export function clearAuth() {
  authUsername = ''
  authPassword = ''
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
