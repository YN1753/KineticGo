let runtime = null

async function getRuntime() {
  if (runtime !== null) return runtime
  try {
    const mod = await import('../../wailsjs/runtime/runtime')
    runtime = mod
    return mod
  } catch {
    runtime = null
    return null
  }
}

export async function eventsOn(eventName, callback) {
  const rt = await getRuntime()
  if (rt && rt.EventsOn) {
    return rt.EventsOn(eventName, callback)
  }
  console.warn(`[mock] EventsOn(${eventName})`)
  return () => {}
}

export async function eventsOff(eventName) {
  const rt = await getRuntime()
  if (rt && rt.EventsOff) {
    rt.EventsOff(eventName)
  }
}
