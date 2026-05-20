<script setup>
import { ref, onMounted } from 'vue'
import {
  ScrollText, ChevronDown, ChevronRight,
  CheckCircle2, AlertCircle, XCircle, Clock
} from 'lucide-vue-next'
import { useTaskApi } from '../composables/useTaskApi'

const { fetchExecutions, fetchLogsByExecution, fetchScheduleList } = useTaskApi()

const executions = ref([])
const scheduleMap = ref({})
const expandedId = ref(null)
const loading = ref(false)

async function load() {
  loading.value = true
  try {
    const [execs, schedules] = await Promise.all([
      fetchExecutions(50),
      fetchScheduleList(),
    ])
    executions.value = execs ?? []
    const map = {}
    for (const s of (schedules ?? [])) {
      map[s.ID] = s.Name
    }
    scheduleMap.value = map
  } catch (e) {
    console.warn('加载历史日志失败', e)
    executions.value = []
  } finally {
    loading.value = false
  }
}

async function toggleExpand(exec) {
  if (expandedId.value === exec.ID) {
    expandedId.value = null
    return
  }
  if (!exec._logs) {
    try {
      exec._logs = await fetchLogsByExecution(exec.ID)
    } catch {
      exec._logs = []
    }
  }
  expandedId.value = exec.ID
}

function scheduleName(optionId) {
  return scheduleMap.value[optionId] ?? '未知任务'
}

function formatTime(val) {
  if (!val) return '-'
  const d = new Date(val)
  if (isNaN(d.getTime())) return String(val)
  const pad = (n) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

function duration(start, end) {
  if (!start || !end) return '-'
  const s = new Date(start), e = new Date(end)
  const ms = e - s
  if (ms < 1000) return `${ms}ms`
  return `${(ms / 1000).toFixed(1)}s`
}

function statusConfig(status) {
  if (status === 'success') {
    return { icon: CheckCircle2, cls: 'text-green-600', text: '成功' }
  }
  if (status === 'running') {
    return { icon: Clock, cls: 'text-blue-600', text: '运行中' }
  }
  return { icon: XCircle, cls: 'text-red-600', text: '失败' }
}

function triggerBadge(type) {
  if (type === 'manual') return { text: '手动', cls: 'bg-gray-100 text-gray-500' }
  return { text: '定时', cls: 'bg-blue-50 text-blue-600' }
}

function levelClass(level) {
  if (level === 'error') return 'text-red-600'
  if (level === 'warn') return 'text-amber-600'
  return 'text-gray-500'
}

onMounted(load)
</script>

<template>
  <div class="min-h-full bg-gray-50">
    <div class="sticky top-0 z-30 bg-white/80 backdrop-blur-xl border-b border-gray-200">
      <div class="max-w-7xl mx-auto px-6 py-3">
        <h1 class="text-base font-semibold text-gray-800">历史日志</h1>
      </div>
    </div>

    <div class="max-w-7xl mx-auto px-6 py-6">
      <div v-if="loading" class="py-24 text-center text-gray-400 text-sm italic">加载中...</div>

      <div v-else-if="executions.length === 0" class="py-24 text-center">
        <div class="inline-flex items-center justify-center w-16 h-16 rounded-2xl bg-gray-100 mb-4 border border-gray-200">
          <ScrollText :size="28" class="text-gray-300" />
        </div>
        <div class="text-gray-400 text-sm">暂无执行记录</div>
      </div>

      <div v-else class="space-y-3">
        <div
          v-for="exec in executions"
          :key="exec.ID"
          class="rounded-2xl border border-gray-200 bg-white overflow-hidden transition-all duration-200 hover:border-blue-300 shadow-sm"
        >
          <!-- card header -->
          <div
            @click="toggleExpand(exec)"
            class="px-5 py-4 flex items-center justify-between cursor-pointer select-none"
          >
            <div class="flex items-center gap-3 min-w-0">
              <component
                :is="statusConfig(exec.Status).icon"
                :size="18"
                :class="statusConfig(exec.Status).cls"
                class="shrink-0"
              />
              <div class="min-w-0">
                <div class="text-sm font-bold text-gray-700 truncate">
                  {{ scheduleName(exec.OptionID) }}
                </div>
                <div class="text-[11px] text-gray-400 mt-0.5 flex items-center gap-2">
                  <span>{{ formatTime(exec.StartTime) }}</span>
                  <span class="text-gray-200">·</span>
                  <span>耗时 {{ duration(exec.StartTime, exec.EndTime) }}</span>
                  <span v-if="exec.ResultSummary" class="text-gray-200">·</span>
                  <span v-if="exec.ResultSummary" class="truncate max-w-[200px]">{{ exec.ResultSummary }}</span>
                </div>
              </div>
            </div>

            <div class="flex items-center gap-2 shrink-0 ml-3">
              <span
                class="text-[10px] px-2 py-0.5 rounded-full font-medium"
                :class="triggerBadge(exec.TriggerType).cls"
              >
                {{ triggerBadge(exec.TriggerType).text }}
              </span>
              <span
                class="text-[10px] px-2 py-0.5 rounded-full flex items-center gap-1 font-bold"
                :class="statusConfig(exec.Status).cls"
              >
                <component :is="statusConfig(exec.Status).icon" :size="10" />
                {{ statusConfig(exec.Status).text }}
              </span>
              <component
                :is="expandedId === exec.ID ? ChevronDown : ChevronRight"
                :size="14"
                class="text-gray-300"
              />
            </div>
          </div>

          <!-- expanded logs -->
          <div v-if="expandedId === exec.ID" class="px-5 pb-4">
            <div class="rounded-xl border border-gray-100 bg-gray-50 overflow-hidden">
              <div class="px-3 py-2 space-y-0.5 max-h-48 overflow-y-auto text-[11px] font-mono">
                <div v-if="!exec._logs || exec._logs.length === 0" class="text-gray-400 flex gap-2">
                  <span class="text-blue-400 select-none">&gt;</span>
                  <span>暂无日志</span>
                </div>
                <div
                  v-for="log in exec._logs"
                  :key="log.ID"
                  class="flex gap-2"
                >
                  <span class="text-gray-400 shrink-0">{{ formatTime(log.CreatedAt).split(' ')[1] }}</span>
                  <span class="shrink-0 w-10 text-right font-bold" :class="levelClass(log.Level)">
                    {{ (log.Level || 'info').toUpperCase() }}
                  </span>
                  <span class="text-gray-600 break-all">{{ log.Message }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
