<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { Cpu, MemoryStick, Activity, ArrowUp, ArrowDown, Plus, X, ScrollText, Wifi, Zap, Radar, Terminal } from 'lucide-vue-next'
import { useSystemStats } from '../composables/useSystemStats'
import { useTaskApi } from '../composables/useTaskApi'
import TaskCard from './TaskCard.vue'
import TaskConfigForm from './TaskConfigForm.vue'

const { cpu, memory, activeTasks, netUp, netDown } = useSystemStats()
const {
  taskList, scheduleList, runningIds, loading,
  fetchTaskList, fetchTaskConfig, fetchScheduleList, fetchScheduleById,
  createSchedule, updateSchedule, deleteSchedule,
  runTask, stopTask, fetchRunningIds,
} = useTaskApi()

const configForm = ref(null)

const typeIcons = {
  campus_auth: Wifi,
  load_test: Zap,
  net_radar: Radar,
  port_killer: Terminal,
}

// modal state
const showPicker = ref(false)
const showConfig = ref(false)
const showDeleteConfirm = ref(false)

const selectedTask = ref(null)
const configFields = ref([])
const configInitial = ref({})
const editingSchedule = ref(null)
const scheduleToDelete = ref(null)

async function openPicker() {
  await fetchTaskList()
  showPicker.value = true
}

async function selectTask(task) {
  selectedTask.value = task
  showPicker.value = false
  const raw = await fetchTaskConfig(task.ID)
  configFields.value = Array.isArray(raw) ? raw : []
  configInitial.value = {}
  editingSchedule.value = null
  showConfig.value = true
}

async function editConfig(schedule) {
  // fetch fresh schedule data
  const fresh = await fetchScheduleById(schedule.ID)
  editingSchedule.value = fresh

  // find matching task template for config schema
  if (taskList.value.length === 0) await fetchTaskList()
  const task = taskList.value.find(t => t.Type === fresh.TaskType)
  selectedTask.value = task || { ID: 0, Name: fresh.Name, Type: fresh.TaskType }

  // get config schema from task template
  if (task) {
    const raw = await fetchTaskConfig(task.ID)
    configFields.value = Array.isArray(raw) ? raw : []
  } else {
    configFields.value = []
  }

  try {
    configInitial.value = typeof fresh.Config === 'string'
      ? JSON.parse(fresh.Config)
      : (fresh.Config || {})
  } catch {
    configInitial.value = {}
  }
  showConfig.value = true
}

async function submitConfig() {
  const values = configForm.value?.getValues() ?? {}
  if (editingSchedule.value) {
    await updateSchedule({ ...editingSchedule.value, Config: JSON.stringify(values) })
  } else {
    await createSchedule({
      Name: selectedTask.value.Name,
      TaskType: selectedTask.value.Type,
      Config: JSON.stringify(values),
      IsEnabled: true,
    })
  }
  showConfig.value = false
  await fetchScheduleList()
}

function confirmDelete(schedule) {
  scheduleToDelete.value = schedule
  showDeleteConfirm.value = true
}

async function doDelete() {
  if (!scheduleToDelete.value) return
  await deleteSchedule(scheduleToDelete.value.ID)
  showDeleteConfirm.value = false
  scheduleToDelete.value = null
  await fetchScheduleList()
}

async function onRun(schedule) {
  await runTask(schedule.ID)
  await fetchRunningIds()
}

async function onStop(schedule) {
  await stopTask(schedule.ID)
  await fetchRunningIds()
}

function formatSpeed(bytesPerSec) {
  const n = Number(bytesPerSec) || 0
  if (n < 1024) return n.toFixed(0) + ' b/s'
  if (n < 1024 * 1024) return (n / 1024).toFixed(1) + ' kb/s'
  return (n / (1024 * 1024)).toFixed(2) + ' MB/s'
}

function getExecMode(scheduleTaskType) {
  if (!scheduleTaskType) return 'both'
  if (scheduleTaskType.startsWith('system-')) {
    const name = scheduleTaskType.substring('system-'.length)
    return taskList.value.find(t => t.Type === 'system' && t.Name === name)?.ExecMode || 'both'
  }
  return taskList.value.find(t => t.Type === scheduleTaskType)?.ExecMode || 'both'
}

function execModeLabel(mode) {
  if (mode === 'manual') return '手动'
  if (mode === 'schedule') return '定时'
  return '手动+定时'
}

let pollTimer = null
onMounted(async () => {
  await fetchTaskList()
  await fetchScheduleList()
  await fetchRunningIds()
  pollTimer = setInterval(fetchRunningIds, 2000)
})

onUnmounted(() => {
  if (pollTimer) clearInterval(pollTimer)
})
</script>

<template>
  <div class="min-h-full">
    <!-- top stats bar -->
    <div class="sticky top-0 z-30 bg-dark-bg/85 backdrop-blur-xl border-b border-dark-border">
      <div class="max-w-7xl mx-auto px-6 py-3 flex items-center justify-between">
        <div class="flex items-center gap-3 min-w-0 overflow-hidden">
          <div class="flex items-center gap-2 text-sm shrink-0">
            <Cpu :size="15" class="text-accent-blue shrink-0" :stroke-width="2" />
            <span class="text-dark-muted shrink-0">CPU</span>
            <span class="font-mono font-semibold text-dark-text tabular-nums w-14 text-right">{{ cpu.toFixed(1) }}%</span>
          </div>
          <div class="w-px h-4 bg-dark-border shrink-0" />
          <div class="flex items-center gap-2 text-sm shrink-0">
            <MemoryStick :size="15" class="text-accent-cyan shrink-0" :stroke-width="2" />
            <span class="text-dark-muted shrink-0">内存</span>
            <span class="font-mono font-semibold text-dark-text tabular-nums w-14 text-right">{{ memory.toFixed(1) }}%</span>
          </div>
          <div class="w-px h-4 bg-dark-border shrink-0" />
          <div class="flex items-center gap-2 text-sm shrink-0">
            <Activity :size="15" class="text-accent-green shrink-0" :stroke-width="2" />
            <span class="text-dark-muted shrink-0">活跃任务</span>
            <span class="font-mono font-semibold text-dark-text tabular-nums w-6 text-right">{{ activeTasks }}</span>
          </div>
          <div class="w-px h-4 bg-dark-border shrink-0 hidden sm:block" />
          <div class="hidden sm:flex items-center gap-3 text-sm shrink-0">
            <div class="flex items-center gap-0.5 w-24 shrink-0">
              <ArrowUp :size="13" class="text-accent-amber shrink-0" :stroke-width="2.5" />
              <span class="font-mono text-dark-muted tabular-nums text-xs truncate min-w-0">{{ formatSpeed(netUp) }}</span>
            </div>
            <div class="flex items-center gap-0.5 w-24 shrink-0">
              <ArrowDown :size="13" class="text-accent-cyan shrink-0" :stroke-width="2.5" />
              <span class="font-mono text-dark-muted tabular-nums text-xs truncate min-w-0">{{ formatSpeed(netDown) }}</span>
            </div>
          </div>
        </div>
        <button
          @click="openPicker"
          class="flex items-center gap-2 px-4 py-2 bg-accent-blue hover:bg-accent-blue/90 text-white rounded-xl text-sm font-medium transition-all duration-200 active:scale-[0.97]"
        >
          <Plus :size="16" :stroke-width="2.5" />
          添加任务
        </button>
      </div>
    </div>

    <!-- content -->
    <div class="max-w-7xl mx-auto px-6 py-6">
      <div v-if="loading" class="py-24 text-center text-dark-muted text-sm">加载中...</div>

      <div v-else-if="scheduleList.length === 0" class="py-24 text-center">
        <div class="inline-flex items-center justify-center w-16 h-16 rounded-2xl bg-black/[0.03] mb-4">
          <ScrollText :size="28" class="text-dark-muted/40" />
        </div>
        <div class="text-dark-muted text-sm mb-1">暂无任务</div>
        <div class="text-dark-muted/60 text-xs">点击右上角「添加任务」开始</div>
      </div>

      <div v-else class="grid gap-4 grid-cols-1 sm:grid-cols-2 xl:grid-cols-3">
        <TaskCard
          v-for="s in scheduleList"
          :key="s.ID"
          :schedule="s"
          :exec-mode="getExecMode(s.TaskType)"
          :is-running="runningIds.has(s.ID)"
          @run="onRun"
          @stop="onStop"
          @edit="editConfig"
          @delete="confirmDelete"
        />
      </div>
    </div>

    <!-- Task Picker Modal -->
    <Teleport to="body">
      <div v-if="showPicker" class="fixed inset-0 z-50 flex items-center justify-center">
        <div class="absolute inset-0 bg-black/40 backdrop-blur-sm" @click="showPicker = false" />
        <div class="relative glass-panel w-full max-w-lg mx-4 overflow-hidden animate-slide-up">
          <div class="flex items-center justify-between px-6 py-4 border-b border-dark-border">
            <h2 class="text-base font-semibold">选择任务类型</h2>
            <button @click="showPicker = false" class="text-dark-muted hover:text-dark-text transition-colors p-1 rounded-lg hover:bg-black/[0.04]">
              <X :size="18" />
            </button>
          </div>
          <div class="p-4 max-h-96 overflow-y-auto space-y-2">
            <button
              v-for="task in taskList"
              :key="task.ID"
              @click="selectTask(task)"
              class="w-full flex items-start gap-4 p-4 rounded-xl border border-transparent hover:border-accent-blue/30 hover:bg-accent-blue/[0.04] text-left transition-all group"
            >
              <div class="flex-shrink-0 w-10 h-10 rounded-xl bg-black/[0.04] flex items-center justify-center text-dark-muted group-hover:text-accent-blue transition-colors">
                <component :is="typeIcons[task.Type] || ScrollText" :size="20" />
              </div>
              <div class="flex-1 min-w-0">
                <div class="font-medium text-sm text-dark-text">{{ task.Name }}</div>
                <div class="text-xs text-dark-muted mt-0.5">{{ task.Description }}</div>
              </div>
              <div class="shrink-0 self-center flex flex-col items-end gap-1">
                <span class="text-[10px] px-2 py-0.5 rounded-full bg-black/[0.04] text-dark-muted font-mono">{{ task.Type }}</span>
                <span class="text-[10px] px-2 py-0.5 rounded-full bg-accent-blue/10 text-accent-blue">{{ execModeLabel(task.ExecMode) }}</span>
              </div>
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Config Modal -->
    <Teleport to="body">
      <div v-if="showConfig" class="fixed inset-0 z-50 flex items-center justify-center">
        <div class="absolute inset-0 bg-black/40 backdrop-blur-sm" @click="showConfig = false" />
        <div class="relative glass-panel w-full max-w-md mx-4 overflow-hidden animate-slide-up">
          <div class="flex items-center justify-between px-6 py-4 border-b border-dark-border">
            <h2 class="text-base font-semibold">{{ editingSchedule ? '修改配置' : '配置任务' }}</h2>
            <button @click="showConfig = false" class="text-dark-muted hover:text-dark-text transition-colors p-1 rounded-lg hover:bg-black/[0.04]">
              <X :size="18" />
            </button>
          </div>
          <div class="p-6 space-y-5">
            <div class="flex items-center gap-2 text-sm">
              <span class="text-dark-muted">任务</span>
              <span class="text-dark-text font-medium px-2 py-0.5 rounded-lg bg-black/[0.04]">{{ selectedTask?.Name }}</span>
            </div>
            <TaskConfigForm
              v-if="configFields.length > 0"
              ref="configForm"
              :fields="configFields"
              :initial-values="configInitial"
            />
            <div v-else class="py-4 text-center text-dark-muted text-sm">该任务无需配置</div>
            <button
              @click="submitConfig"
              class="w-full py-2.5 bg-accent-blue hover:bg-accent-blue/90 text-white rounded-xl text-sm font-medium transition-all duration-200 active:scale-[0.98]"
            >
              {{ editingSchedule ? '保存修改' : '创建任务' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Delete Confirm Modal -->
    <Teleport to="body">
      <div v-if="showDeleteConfirm" class="fixed inset-0 z-50 flex items-center justify-center">
        <div class="absolute inset-0 bg-black/40 backdrop-blur-sm" @click="showDeleteConfirm = false" />
        <div class="relative glass-panel w-full max-w-sm mx-4 overflow-hidden animate-slide-up">
          <div class="p-6 space-y-4">
            <h2 class="text-base font-semibold">确认删除</h2>
            <p class="text-sm text-dark-muted leading-relaxed">
              确定要删除任务「<span class="text-dark-text font-medium">{{ scheduleToDelete?.Name }}</span>」吗？此操作不可撤销。
            </p>
            <div class="flex gap-3 pt-2">
              <button
                @click="showDeleteConfirm = false"
                class="flex-1 py-2.5 bg-black/[0.04] hover:bg-black/[0.07] text-dark-text rounded-xl text-sm font-medium transition-colors"
              >
                取消
              </button>
              <button
                @click="doDelete"
                class="flex-1 py-2.5 bg-accent-red/90 hover:bg-accent-red text-white rounded-xl text-sm font-medium transition-colors"
              >
                删除
              </button>
            </div>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>
