import { onMounted, onUnmounted } from 'vue'

export function useAutoRefresh(fetchFn: () => void, intervalMs = 60_000) {
  let timer: ReturnType<typeof setInterval> | null = null

  onMounted(() => {
    timer = setInterval(fetchFn, intervalMs)
  })

  onUnmounted(() => {
    if (timer) clearInterval(timer)
  })
}
