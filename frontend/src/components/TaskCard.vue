<script setup>
import { computed } from 'vue'
import { Play, Square, Settings, X, Wifi, Zap, Radar, Terminal, ScrollText, Clock, ClipboardCheck, Trash2, Rocket, ChevronDown } from 'lucide-vue-next'

const props = defineProps({
  schedule: { type: Object, required: true },
  execMode: { type: String, default: 'both' },
  isRunning: { type: Boolean, default: false },
})

const emit = defineEmits(['run', 'stop', 'edit', 'delete'])

const typeIcons = {
  campus_auth: Wifi,
  '652_signin': ClipboardCheck,
  load_test: Zap,
  net_radar: Radar,
  port_killer: Terminal,
  app_launcher: Rocket,
}

const icon = computed(() => typeIcons[props.schedule.TaskType] || ScrollText)

const isLauncher = computed(() => props.schedule.TaskType === 'app_launcher')

const launcherPaths = computed(() => {
  if (!isLauncher.value) return []
  try {
    const config = typeof props.schedule.Config === 'string' 
      ? JSON.parse(props.schedule.Config) 
      : props.schedule.Config
    const paths = config?.paths || ''
    return paths.split('\n').map(p => p.trim()).filter(p => p)
  } catch (e) {
    return []
  }
})

const execModeText = computed(() => {
  if (isLauncher.value) return '快捷启动'
  if (props.execMode === 'manual') return '手动'
  if (props.execMode === 'schedule') return '定时'
  return props.schedule.CronExpr ? '定时' : '手动'
})

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

function handleRun() {
  emit('run', props.schedule)
}

function handleStop() {
  emit('stop', props.schedule)
}
</script>

<template>
  <div
    class="relative rounded-2xl border border-gray-100 bg-white overflow-hidden transition-all duration-300 hover:border-blue-300 group animate-fade-in shadow-sm hover:shadow-md p-4"
    :class="{ 'ring-1 ring-green-500/20 border-green-100': isRunning }"
  >
    <!-- Header -->
    <div class="flex items-start justify-between mb-3">
      <div class="flex items-center gap-3">
        <div 
          class="w-10 h-10 rounded-xl flex items-center justify-center transition-all shadow-inner"
          :class="isRunning ? 'bg-green-50 text-green-500' : 'bg-gray-50 text-gray-300'"
        >
          <component :is="icon" :size="20" :stroke-width="2.5" />
        </div>
        <div class="min-w-0">
          <h3 class="font-bold text-[13px] text-gray-800 truncate leading-tight">
            {{ schedule.Name }}
          </h3>
          <div v-if="schedule.Option" class="text-[10px] text-gray-400 truncate mt-0.5 font-medium">{{ schedule.Option }}</div>
        </div>
      </div>
      
      <button
        @click="emit('delete', schedule)"
        class="p-1 rounded-full text-gray-200 hover:text-red-500 hover:bg-red-50 transition-all opacity-0 group-hover:opacity-100"
      >
        <X :size="14" />
      </button>
    </div>

    <!-- Status -->
    <div class="flex items-center gap-2 mb-4 text-[10px]">
      <div class="flex items-center gap-1.5 px-1.5 py-0.5 rounded-lg" :class="isLauncher ? 'bg-blue-50 border border-blue-100/50' : 'bg-gray-50 border border-gray-100/50'">
        <span class="w-1.5 h-1.5 rounded-full" :class="isRunning ? 'bg-green-500 animate-pulse' : (isLauncher ? 'bg-blue-400' : 'bg-gray-300')" />
        <span :class="isRunning ? 'text-green-600 font-bold' : (isLauncher ? 'text-blue-600 font-bold' : 'text-gray-400 font-medium')">{{ isRunning ? '运行中' : (isLauncher ? '快捷' : '就绪') }}</span>
      </div>
      <span class="text-gray-400 font-bold">{{ execModeText }}</span>
      <div v-if="hasNextRun" class="flex items-center gap-1 px-1.5 py-0.5 rounded-lg bg-blue-50 text-blue-500 font-bold border border-blue-100/50">
        <Clock :size="10" />
        {{ formatNextRun(props.schedule.NextRunTime) }}
      </div>
    </div>

    <!-- Path Preview for App Launcher -->
    <div v-if="isLauncher && launcherPaths.length > 0" class="mb-4 space-y-1">
      <div v-for="(path, idx) in launcherPaths.slice(0, 3)" :key="idx" class="flex items-center gap-2 px-2 py-1 bg-gray-50/50 rounded-md border border-gray-100/50">
        <span class="text-[9px] font-black text-blue-300 font-mono">#{{ idx + 1 }}</span>
        <span class="text-[9px] text-gray-500 truncate font-medium">{{ path }}</span>
      </div>
      <div v-if="launcherPaths.length > 3" class="text-[8px] text-gray-300 italic px-2">+ {{ launcherPaths.length - 3 }} 更多目标...</div>
    </div>

    <!-- Actions -->
    <div class="flex gap-2">
      <button
        v-if="!isRunning"
        @click="handleRun"
        class="flex-1 flex items-center justify-center gap-1.5 py-1.5 rounded-xl text-[10px] font-bold transition-all active:scale-95 bg-gray-50 text-gray-500 hover:bg-gray-100 border border-gray-100 shadow-sm"
        :class="{'bg-blue-50 text-blue-600 border-blue-100 hover:bg-blue-100': isLauncher}"
      >
        <Rocket v-if="isLauncher" :size="10" />
        {{ isLauncher ? '一键唤醒' : (props.schedule.CronExpr ? '待触发' : '启动') }}
      </button>
      <button
        v-else
        @click="handleStop"
        class="flex-1 flex items-center justify-center gap-1.5 py-1.5 rounded-xl text-[10px] font-bold transition-all active:scale-95 bg-red-50 text-red-600 hover:bg-red-100 border border-red-100 shadow-sm"
      >
        <Square :size="10" :fill="true" />
        停止
      </button>
      
      <button
        @click="emit('edit', schedule)"
        class="px-2.5 flex items-center justify-center rounded-xl text-[10px] font-bold transition-all active:scale-95 bg-gray-50 text-gray-400 hover:bg-gray-100 border border-gray-100 shadow-sm"
        title="配置"
      >
        <Settings :size="12" />
      </button>
    </div>

    <!-- Log Tip -->
    <div class="mt-3 pt-2.5 border-t border-gray-50 flex items-center gap-2 text-[9px] text-gray-300 italic font-medium">
      <span class="text-blue-300 font-black">&gt;</span>
      <span>实时流水推流中...</span>
    </div>
  </div>
</template>
