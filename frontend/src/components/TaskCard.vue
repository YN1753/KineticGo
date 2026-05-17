<script setup>
import { computed, ref, onMounted, onUnmounted, nextTick } from 'vue'
import { Play, Square, Settings, X, Wifi, Zap, Radar, Terminal, ScrollText, Clock } from 'lucide-vue-next'
import { eventsOn } from '../composables/useWailsRuntime'

const runningGlow = 'shadow-[0_0_24px_rgba(16,185,129,0.18)]'

const props = defineProps({
  schedule: { type: Object, required: true },
  execMode: { type: String, default: 'both' },
  isRunning: { type: Boolean, default: false },
})

const emit = defineEmits(['run', 'stop', 'edit', 'delete'])

const typeIcons = {
  campus_auth: Wifi,
  load_test: Zap,
  net_radar: Radar,
  port_killer: Terminal,
}

const icon = computed(() => typeIcons[props.schedule.TaskType] || ScrollText)

// 系统任务持续 tick，不展示 mini console
const isSystemTask = computed(() => (props.schedule.TaskType || '').startsWith('system-'))

// 模板若是 both，按 schedule 自己的 CronExpr 决定实际模式
const effectiveMode = computed(() => {
  if (props.execMode === 'manual') return 'manual'
  if (props.execMode === 'schedule') return 'schedule'
  return props.schedule.CronExpr ? 'schedule' : 'manual'
})

const canManualStart = computed(
  () => !props.isRunning && effectiveMode.value === 'manual'
)
const canStop = computed(() => props.isRunning)

const execModeText = computed(() => effectiveMode.value === 'schedule' ? '定时' : '手动')

const hasNextRun = computed(() => {
  return !!props.schedule.CronExpr && props.schedule.NextRunTime && props.schedule.NextRunTime !== '0001-01-01T00:00:00Z'
})

function formatNextRun(val) {
  if (!val) return ''
  const d = new Date(val)
  if (isNaN(d.getTime())) return ''
  const now = new Date()
  const pad = (n) => String(n).padStart(2, '0')
  const hm = `${pad(d.getHours())}:${pad(d.getMinutes())}`
  if (d.toDateString() === now.toDateString()) return `今天 ${hm}`
  const tmr = new Date(now)
  tmr.setDate(tmr.getDate() + 1)
  if (d.toDateString() === tmr.toDateString()) return `明天 ${hm}`
  return `${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${hm}`
}

// ===== mini console 日志 =====
const MAX_LOGS = 50
const logs = ref([])
const consoleRef = ref(null)
let unsubscribe = null

function formatTime(unix) {
  const d = unix ? new Date(unix * 1000) : new Date()
  const pad = (n) => String(n).padStart(2, '0')
  return `${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

function levelColor(level) {
  if (level === 'error') return 'text-accent-red'
  if (level === 'warn') return 'text-accent-amber'
  return 'text-dark-text/80'
}

function appendLog(payload) {
  logs.value.push({
    time: formatTime(payload.time),
    level: payload.level || 'info',
    message: payload.message || '',
  })
  if (logs.value.length > MAX_LOGS) {
    logs.value.splice(0, logs.value.length - MAX_LOGS)
  }
  nextTick(() => {
    if (consoleRef.value) consoleRef.value.scrollTop = consoleRef.value.scrollHeight
  })
}

onMounted(async () => {
  if (isSystemTask.value) return
  unsubscribe = await eventsOn('task_log', (payload) => {
    if (!payload || Number(payload.scheduleId) !== Number(props.schedule.ID)) return
    appendLog(payload)
  })
})

onUnmounted(() => {
  if (typeof unsubscribe === 'function') unsubscribe()
})

function handleRun() {
  emit('run', props.schedule)
}

function handleStop() {
  emit('stop', props.schedule)
}
</script>

<template>
  <div
    class="relative rounded-2xl border border-dark-border bg-dark-card overflow-hidden transition-all duration-300 hover:translate-y-[-1px]"
    :class="isRunning ? runningGlow : 'shadow-card hover:shadow-card-hover'"
  >
    <!-- delete button -->
    <button
      @click="emit('delete', schedule)"
      class="absolute top-3 right-3 z-10 w-6 h-6 flex items-center justify-center rounded-full text-dark-muted/60 hover:text-accent-red hover:bg-accent-red/10 transition-all"
    >
      <X :size="14" />
    </button>

    <!-- header -->
    <div class="px-5 pt-4 pb-3 flex items-start gap-3.5">
      <div
        class="flex-shrink-0 w-10 h-10 rounded-xl flex items-center justify-center transition-colors"
        :class="isRunning ? 'bg-accent-green/10 text-accent-green' : 'bg-black/[0.04] text-dark-muted'"
      >
        <component :is="icon" :size="20" :stroke-width="1.8" />
      </div>
      <div class="min-w-0 flex-1 pr-6">
        <div class="font-semibold text-sm text-dark-text truncate">{{ schedule.Name }}</div>
        <div class="flex items-center gap-2 mt-1">
          <span
            class="status-dot"
            :class="isRunning ? 'bg-accent-green text-accent-green' : 'bg-dark-muted/40 text-dark-muted/40'"
          />
          <span class="text-[11px]" :class="isRunning ? 'text-accent-green' : 'text-dark-muted'">
            {{ isRunning ? '运行中' : '已就绪' }}
          </span>
          <span class="text-[10px] px-1.5 py-0.5 rounded-full bg-black/[0.04] text-dark-muted">{{ execModeText }}</span>
          <span
            v-if="hasNextRun"
            class="text-[10px] px-1.5 py-0.5 rounded-full bg-accent-blue/10 text-accent-blue flex items-center gap-1"
          >
            <Clock :size="10" />
            {{ formatNextRun(schedule.NextRunTime) }}
          </span>
          <span class="text-[10px] text-dark-muted/40 font-mono">#{{ schedule.ID }}</span>
        </div>
      </div>
    </div>

    <!-- action buttons -->
    <div class="px-5 pb-3 flex gap-2">
      <button
        v-if="canManualStart"
        @click="handleRun"
        class="flex-1 flex items-center justify-center gap-1.5 px-3 py-2 rounded-xl text-xs font-semibold transition-all duration-200 bg-accent-blue/10 text-accent-blue hover:bg-accent-blue/20 active:scale-[0.97]"
      >
        <Play :size="13" :fill="true" />
        启动
      </button>
      <button
        v-else-if="canStop"
        @click="handleStop"
        class="flex-1 flex items-center justify-center gap-1.5 px-3 py-2 rounded-xl text-xs font-semibold transition-all duration-200 bg-accent-red/10 text-accent-red hover:bg-accent-red/20 active:scale-[0.97]"
      >
        <Square :size="13" :fill="true" />
        结束
      </button>
      <div
        v-else
        class="flex-1 flex items-center justify-center gap-1.5 px-3 py-2 rounded-xl text-xs font-medium bg-black/[0.02] text-dark-muted/60"
      >
        等待定时触发
      </div>
      <button
        @click="emit('edit', schedule)"
        class="flex-1 flex items-center justify-center gap-1.5 px-3 py-2 rounded-xl text-xs font-medium transition-all duration-200 bg-black/[0.04] text-dark-muted hover:text-dark-text hover:bg-black/[0.07] active:scale-[0.97]"
      >
        <Settings :size="13" />
        配置
      </button>
    </div>

    <!-- mini console（非 system 任务展示） -->
    <div
      v-if="!isSystemTask"
      class="mini-console mx-3 mb-3 rounded-xl border border-dark-border overflow-hidden"
    >
      <div ref="consoleRef" class="px-3 py-2 space-y-0.5 overflow-y-auto max-h-32 text-[11px] font-mono">
        <div v-if="logs.length === 0" class="text-dark-muted/70 flex gap-2">
          <span class="text-accent-cyan/60 select-none">&gt;</span>
          <span>{{ isRunning ? '运行中...' : '等待输出...' }}</span>
        </div>
        <div v-for="(line, i) in logs" :key="i" class="flex gap-2">
          <span class="text-dark-muted/50 shrink-0">{{ line.time }}</span>
          <span class="text-dark-text/80 break-all" :class="levelColor(line.level)">{{ line.message }}</span>
        </div>
      </div>
    </div>
  </div>
</template>
