import { ref, watch } from 'vue'

const STORAGE_KEY = 'theme_mode'

function getInitialMode(): boolean {
  const stored = localStorage.getItem(STORAGE_KEY)
  if (stored === 'dark') return true
  if (stored === 'light') return false
  return window.matchMedia('(prefers-color-scheme: dark)').matches
}

const isDark = ref(getInitialMode())

// Apply dark class on init
if (isDark.value) {
  document.documentElement.classList.add('dark')
}

// Sync to DOM + localStorage
watch(isDark, (dark) => {
  document.documentElement.classList.toggle('dark', dark)
  localStorage.setItem(STORAGE_KEY, dark ? 'dark' : 'light')
})

// Listen for OS theme changes when in auto mode
window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', (e) => {
  const stored = localStorage.getItem(STORAGE_KEY)
  if (!stored) {
    isDark.value = e.matches
  }
})

export function useDarkMode() {
  function toggle() {
    isDark.value = !isDark.value
  }

  function setDark(dark: boolean) {
    isDark.value = dark
  }

  return { isDark, toggle, setDark }
}
