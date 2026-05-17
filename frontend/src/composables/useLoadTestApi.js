import { ref } from 'vue'

export function useLoadTestApi() {
  const status = ref('idle')

  async function startTest(config) {
    try {
      const mod = await import('../../wailsjs/go/wailsapp/App')
      await mod.StartLoadTest(config.url, config.concurrency, config.duration)
      status.value = 'running'
    } catch {
      console.warn('[mock] StartLoadTest', config)
      status.value = 'running'
    }
  }

  async function stopTest() {
    try {
      const mod = await import('../../wailsjs/go/wailsapp/App')
      await mod.StopLoadTest()
      status.value = 'idle'
    } catch {
      console.warn('[mock] StopLoadTest')
      status.value = 'idle'
    }
  }

  return { status, startTest, stopTest }
}
