import { ref, onMounted, onUnmounted } from 'vue'
import { eventsOn } from './useWailsRuntime'

export function useSystemStats() {
  const cpu = ref(0)
  const memory = ref(0)
  const activeTasks = ref(0)
  const netUp = ref(0)
  const netDown = ref(0)
  const qpsHistory = ref([])
  const unsubs = []

  function applyStats(data) {
    if (!data || typeof data !== 'object') return
    if (typeof data.cpuPercent === 'number') cpu.value = data.cpuPercent
    if (typeof data.memPercent === 'number') memory.value = data.memPercent
    if (typeof data.activeTasks === 'number') activeTasks.value = data.activeTasks
    if (typeof data.uploadSpeed === 'number') netUp.value = data.uploadSpeed
    if (typeof data.downloadSpeed === 'number') netDown.value = data.downloadSpeed
  }

  onMounted(async () => {
    const unsubStats = await eventsOn('stats_update', applyStats)
    if (typeof unsubStats === 'function') unsubs.push(unsubStats)

    const unsubQps = await eventsOn('qps_update', (point) => {
      qpsHistory.value.push(point)
      if (qpsHistory.value.length > 120) qpsHistory.value.shift()
    })
    if (typeof unsubQps === 'function') unsubs.push(unsubQps)
  })

  onUnmounted(() => {
    unsubs.forEach((u) => u?.())
  })

  return { cpu, memory, activeTasks, netUp, netDown, qpsHistory }
}
