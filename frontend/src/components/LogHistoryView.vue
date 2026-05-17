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
    return { icon: CheckCircle2, cls: 'text-accent-green', text: '成功' }
  }
  if (status === 'running') {
    return { icon: Clock, cls: 'text-accent-blue', text: '运行中' }
  }
  return { icon: XCircle, cls: 'text-accent-red', text: '失败' }
}

function triggerBadge(type) {
  if (type === 'manual') return { text: '手动', cls: 'bg-black/[0.04] text-dark-muted' }
  return { text: '定时', cls: 'bg-accent-blue/10 text-accent-blue' }
}

function levelClass(level) {
  if (level === 'error') return 'text-accent-red'
  if (level === 'warn') return 'text-accent-amber'
  return 'text-dark-muted'
}

onMounted(load)
</script>

<template>
  <div class="min-h-full">
    <div class="sticky top-0 z-30 bg-dark-bg/85 backdrop-blur-xl border-b border-dark-border">
      <div class="max-w-7xl mx-auto px-6 py-3">
        <h1 class="text-base font-semibold">历史日志</h1>
      </div>
    </div>

    <div class="max-w-7xl mx-auto px-6 py-6">
      <div v-if="loading" class="py-24 text-center text-dark-muted text-sm">加载中...</div>

      <div v-else-if="executions.length === 0" class="py-24 text-center">
        <div class="inline-flex items-center justify-center w-16 h-16 rounded-2xl bg-black/[0.04] mb-4">
          <ScrollText :size="28" class="text-dark-muted/50" />
        </div>
        <div class="text-dark-muted text-sm">暂无执行记录</div>
      </div>

      <div v-else class="space-y-3">
        <div
          v-for="exec in executions"
          :key="exec.ID"
          class="rounded-2xl border border-dark-border bg-dark-card overflow-hidden transition-all duration-200 hover:border-dark-border/80"
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
                <div class="text-sm font-medium text-dark-text truncate">
                  {{ scheduleName(exec.OptionID) }}
                </div>
                <div class="text-[11px] text-dark-muted mt-0.5 flex items-center gap-2">
                  <span>{{ formatTime(exec.StartTime) }}</span>
                  <span class="text-dark-muted/40">·</span>
                  <span>耗时 {{ duration(exec.StartTime, exec.EndTime) }}</span>
                  <span v-if="exec.ResultSummary" class="text-dark-muted/40">·</span>
                  <span v-if="exec.ResultSummary" class="truncate max-w-[200px]">{{ exec.ResultSummary }}</span>
                </div>
              </div>
            </div>

            <div class="flex items-center gap-2 shrink-0 ml-3">
              <span
                class="text-[10px] px-2 py-0.5 rounded-full"
                :class="triggerBadge(exec.TriggerType).cls"
              >
                {{ triggerBadge(exec.TriggerType).text }}
              </span>
              <span
                class="text-[10px] px-2 py-0.5 rounded-full flex items-center gap-1"
                :class="statusConfig(exec.Status).cls"
              >
                <component :is="statusConfig(exec.Status).icon" :size="10" />
                {{ statusConfig(exec.Status).text }}
              </span>
              <component
                :is="expandedId === exec.ID ? ChevronDown : ChevronRight"
                :size="14"
                class="text-dark-muted/60"
              />
            </div>
          </div>

          <!-- expanded logs -->
          <div v-if="expandedId === exec.ID" class="px-5 pb-4">
            <div class="rounded-xl border border-dark-border overflow-hidden">
              <div class="px-3 py-2 space-y-0.5 max-h-48 overflow-y-auto text-[11px] font-mono">
                <div v-if="!exec._logs || exec._logs.length === 0" class="text-dark-muted/70 flex gap-2">
                  <span class="text-accent-cyan/60 select-none">&gt;</span>
                  <span>暂无日志</span>
                </div>
                <div
                  v-for="log in exec._logs"
                  :key="log.ID"
                  class="flex gap-2"
                >
                  <span class="text-dark-muted/50 shrink-0">{{ formatTime(log.CreatedAt).split(' ')[1] }}</span>
                  <span class="shrink-0 w-10 text-right" :class="levelClass(log.Level)">
                    {{ (log.Level || 'info').toUpperCase() }}
                  </span>
                  <span class="text-dark-text/80 break-all">{{ log.Message }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
