<script setup>
import { ref, onMounted, onUnmounted, nextTick, computed } from 'vue'
import { 
  Cpu, MemoryStick, Activity, ArrowUp, ArrowDown, Plus, X, 
  ScrollText, Wifi, Zap, Radar, Terminal, Clock, Hand, 
  ClipboardCheck, ChevronDown, ChevronRight, Search, Scan, 
  Cpu as CpuIcon, Code, Terminal as ShellIcon, Database, 
  BrainCircuit, AlertCircle, Info, RefreshCw, Power
} from 'lucide-vue-next'
import { useSystemStats } from '../composables/useSystemStats'
import { useTaskApi } from '../composables/useTaskApi'
import { eventsOn } from '../composables/useWailsRuntime'
import TaskCard from './TaskCard.vue'
import TaskConfigForm from './TaskConfigForm.vue'

const { cpu, memory, activeTasks, netUp, netDown } = useSystemStats()
const {
  taskList, scheduleList, runningIds, loading,
  fetchTaskList, fetchTaskConfig, fetchScheduleList, fetchScheduleById,
  createSchedule, updateSchedule, deleteSchedule,
  runTask, stopTask, fetchRunningIds,
  fetchPortMessages, runPortKiller,
} = useTaskApi()

const configForm = ref(null)

const typeIcons = {
  campus_auth: Wifi,
  '652_signin': ClipboardCheck,
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

const collapsedGroups = ref(new Set())
const isLauncherCollapsed = ref(false)
const isPortKillerCollapsed = ref(false)

const toggleGroup = (type) => {
  if (collapsedGroups.value.has(type)) {
    collapsedGroups.value.delete(type)
  } else {
    collapsedGroups.value.add(type)
  }
}

// 分组逻辑
const groupedSchedules = computed(() => {
  const groups = {}
  scheduleList.value.forEach(s => {
    const type = s.TaskType
    if (!groups[type]) {
      let template = taskList.value.find(t => t.Type === type)
      if (!template && type.startsWith('system-')) {
        const sysName = type.replace('system-', '')
        template = taskList.value.find(t => t.Type === 'system' && t.Name === sysName)
      }
      groups[type] = {
        type: type,
        name: template?.Name || type,
        icon: typeIcons[type] || (type.startsWith('system-') ? (type.includes('cpu') ? Cpu : type.includes('memory') ? MemoryStick : type.includes('network') ? Activity : ScrollText) : ScrollText),
        schedules: []
      }
    }
    groups[type].schedules.push(s)
  })
  return Object.values(groups)
})

// --- Port Killer Implementation ---
const portSearch = ref('')
const isScanning = ref(false)
const scannedPorts = ref([])

async function doScanPorts() {
  isScanning.value = true
  try {
    const ports = await fetchPortMessages()
    scannedPorts.value = ports || []
  } finally {
    isScanning.value = false
  }
}

const filteredPorts = computed(() => {
  let list = scannedPorts.value
  if (portSearch.value.trim()) {
    const s = portSearch.value.toLowerCase()
    list = list.filter(p => 
      String(p.port).includes(s) || 
      p.name.toLowerCase().includes(s) || 
      p.path.toLowerCase().includes(s)
    )
  }
  // 排序逻辑：将用户应用（isCritical 为 false）排在最上面
  return [...list].sort((a, b) => {
    if (a.isCritical === b.isCritical) return a.port - b.port // 同类按端口号升序
    return a.isCritical ? 1 : -1 // a 是系统核心则往后排，否则往前排
  })
})

async function onKillPort(port) {
  try {
    await runPortKiller(port.pid)
    // 延迟刷新列表
    setTimeout(doScanPorts, 1500)
  } catch (e) {
    console.error('释放端口失败', e)
  }
}

// --- App Launcher Mock ---
const apps = ref([
  { name: 'VS Code', path: 'Code.exe', icon: Code, color: 'text-blue-500' },
  { name: 'GoLand', path: 'goland64.exe', icon: CpuIcon, color: 'text-cyan-500' },
  { name: 'Terminal', path: 'PowerShell', icon: ShellIcon, color: 'text-indigo-500' },
  { name: 'DBeaver', path: 'dbeaver.exe', icon: Database, color: 'text-blue-600' },
  { name: 'Postman', path: 'postman.exe', icon: Zap, color: 'text-orange-500' },
  { name: 'Chrome', path: 'chrome.exe', icon: Activity, color: 'text-green-500' },
])

// --- Centralized Logs ---
const allLogs = ref([])
const terminalRef = ref(null)
const MAX_LOGS = 200
let unsubscribeLogs = null

function formatTime(unix) {
  const d = unix ? new Date(unix * 1000) : new Date()
  const pad = (n) => String(n).padStart(2, '0')
  return `${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

function appendGlobalLog(payload) {
  let tag = ''
  if (payload.scheduleId === 0) {
    tag = '[系统即时任务] '
  } else {
    const schedule = scheduleList.value.find(s => Number(s.ID) === Number(payload.scheduleId))
    tag = schedule ? `[${schedule.Name}${schedule.Option ? ' - ' + schedule.Option : ''}] ` : ''
  }
  
  allLogs.value.push({
    time: formatTime(payload.time),
    level: payload.level || 'info',
    message: payload.message || '',
    tag: tag,
    scheduleId: payload.scheduleId
  })
  if (allLogs.value.length > MAX_LOGS) allLogs.value.shift()
  nextTick(() => { if (terminalRef.value) terminalRef.value.scrollTop = terminalRef.value.scrollHeight })
}

// --- Logic methods ---
async function openPicker() { await fetchTaskList(); showPicker.value = true }
async function selectTask(task) {
  selectedTask.value = task; showPicker.value = false
  const raw = await fetchTaskConfig(task.ID); configFields.value = Array.isArray(raw) ? raw : []
  configInitial.value = {}; editingSchedule.value = null
  const mode = task.ExecMode || 'manual'
  if (mode === 'both') { chosenExecMode.value = 'manual'; lockExecMode.value = false } 
  else { chosenExecMode.value = mode === 'schedule' ? 'schedule' : 'manual'; lockExecMode.value = true }
  cronExpr.value = ''; cronError.value = ''; showConfig.value = true
}
async function editConfig(schedule) {
  const fresh = await fetchScheduleById(schedule.ID); editingSchedule.value = fresh
  if (taskList.value.length === 0) await fetchTaskList()
  const task = taskList.value.find(t => t.Type === fresh.TaskType)
  selectedTask.value = task || { ID: 0, Name: fresh.Name, Type: fresh.TaskType }
  if (task) { const raw = await fetchTaskConfig(task.ID); configFields.value = Array.isArray(raw) ? raw : [] }
  else configFields.value = []
  try { configInitial.value = typeof fresh.Config === 'string' ? JSON.parse(fresh.Config) : (fresh.Config || {}) } catch { configInitial.value = {} }
  cronExpr.value = fresh.CronExpr || ''; cronError.value = ''
  chosenExecMode.value = fresh.CronExpr ? 'schedule' : 'manual'; lockExecMode.value = true
  taskOption.value = fresh.Option || ''; showConfig.value = true
}
async function submitConfig() {
  cronError.value = ''; const isSchedule = chosenExecMode.value === 'schedule'
  if (isSchedule && !cronExpr.value.trim()) { cronError.value = '请填写 cron 表达式'; return }
  const values = configForm.value?.getValues() ?? {}; const cronToSend = isSchedule ? cronExpr.value.trim() : ''
  if (editingSchedule.value) await updateSchedule({ ...editingSchedule.value, Config: JSON.stringify(values), CronExpr: cronToSend, Option: taskOption.value.trim() })
  else await createSchedule({ Name: selectedTask.value.Name, TaskType: selectedTask.value.Type, Config: JSON.stringify(values), CronExpr: cronToSend, IsEnabled: true, Option: taskOption.value.trim() })
  closeConfig(); await fetchScheduleList()
}
function closeConfig() { showConfig.value = false; cronExpr.value = ''; cronError.value = ''; chosenExecMode.value = 'manual'; lockExecMode.value = false; editingSchedule.value = null; taskOption.value = '' }
function selectExecMode(mode) { if (lockExecMode.value && mode !== chosenExecMode.value) return; chosenExecMode.value = mode; cronError.value = '' }
function applyCronPreset(expr) { cronExpr.value = expr; cronError.value = '' }
function confirmDelete(schedule) { scheduleToDelete.value = schedule; showDeleteConfirm.value = true }
async function doDelete() { if (!scheduleToDelete.value) return; await deleteSchedule(scheduleToDelete.value.ID); showDeleteConfirm.value = false; scheduleToDelete.value = null; await fetchScheduleList() }
async function onRun(schedule) { await runTask(schedule.ID); await fetchRunningIds() }
async function onStop(schedule) { await stopTask(schedule.ID); await fetchRunningIds() }
function formatSpeed(bytesPerSec) {
  const n = Number(bytesPerSec) || 0
  if (n < 1024) return n.toFixed(0) + ' b/s'
  if (n < 1024 * 1024) return (n / 1024).toFixed(1) + ' kb/s'
  return (n / (1024 * 1024)).toFixed(2) + ' MB/s'
}
function getExecMode(st) {
  if (!st) return 'both'
  if (st.startsWith('system-')) return taskList.value.find(t => t.Type === 'system' && t.Name === st.substring(7))?.ExecMode || 'both'
  return taskList.value.find(t => t.Type === st)?.ExecMode || 'both'
}
function execModeLabel(m) { return m === 'manual' ? '手动' : m === 'schedule' ? '定时' : '手动+定时' }

let pollTimer = null
onMounted(async () => {
  await fetchTaskList(); await fetchScheduleList(); await fetchRunningIds(); pollTimer = setInterval(fetchRunningIds, 2000)
  unsubscribeLogs = await eventsOn('task_log', (payload) => { if (payload) appendGlobalLog(payload) })
})
onUnmounted(() => { if (pollTimer) clearInterval(pollTimer); if (unsubscribeLogs) unsubscribeLogs() })

const chosenExecMode = ref('manual')
const lockExecMode = ref(false)
const cronExpr = ref('')
const taskOption = ref('')
const cronError = ref('')
</script>

<template>
  <div class="h-full bg-gray-50 text-gray-800 p-4 flex flex-col gap-4 overflow-hidden">
    
    <!-- Main Grid Content -->
    <div class="flex-1 min-h-0 overflow-y-auto lg:overflow-hidden custom-scrollbar pb-2 lg:pb-0">
      <div class="grid grid-cols-12 gap-4 lg:h-full">
      
        <!-- Column 1: Task Management -->
        <div class="col-span-12 lg:col-span-3 flex flex-col min-h-0 lg:h-full">
          <div class="flex items-center justify-between mb-2 px-1 shrink-0">
            <h2 class="text-[11px] font-bold text-gray-400 uppercase tracking-widest">任务实例</h2>
            <button @click="openPicker" class="p-1 rounded-md bg-blue-50 text-blue-600 hover:bg-blue-100 transition-all border border-blue-100 active:scale-95">
              <Plus :size="14" :stroke-width="3" />
            </button>
          </div>

          <div class="flex-1 overflow-visible lg:overflow-y-auto lg:pr-1 custom-scrollbar space-y-4 min-h-0">
            <div v-if="loading" class="py-12 text-center text-gray-400 text-xs italic">Loading...</div>
            <div v-else-if="groupedSchedules.length === 0" class="py-12 text-center bg-white rounded-2xl border border-dashed border-gray-200">
              <p class="text-xs text-gray-300">暂无实例</p>
            </div>
            <div v-else v-for="group in groupedSchedules" :key="group.type" class="space-y-2">
              <div @click="toggleGroup(group.type)" class="flex items-center justify-between group cursor-pointer sticky top-0 bg-gray-50/95 backdrop-blur-sm py-1.5 z-10">
                <div class="flex items-center gap-2">
                  <div class="p-1.5 rounded-lg bg-white shadow-sm border border-gray-100 text-gray-400 group-hover:text-blue-500 transition-colors">
                    <component :is="group.icon" :size="14" />
                  </div>
                  <h3 class="text-[11px] font-bold text-gray-600">{{ group.name }}</h3>
                </div>
                <component :is="collapsedGroups.has(group.type) ? ChevronRight : ChevronDown" :size="12" class="text-gray-300" />
              </div>
              <div v-if="!collapsedGroups.has(group.type)" class="grid grid-cols-1 gap-3 animate-fade-in pb-2">
                <TaskCard v-for="s in group.schedules" :key="s.ID" :schedule="s" :exec-mode="getExecMode(s.TaskType)" :is-running="runningIds.has(s.ID)" @run="onRun" @stop="onStop" @edit="editConfig" @delete="confirmDelete" />
              </div>
            </div>
          </div>
        </div>

        <!-- Column 2: App Launcher -->
        <div class="col-span-12 lg:col-span-6 flex flex-col min-h-0 lg:h-full">
          <div @click="isLauncherCollapsed = !isLauncherCollapsed" class="flex items-center justify-between mb-2 px-1 shrink-0 cursor-pointer group">
            <h2 class="text-[11px] font-bold text-gray-400 uppercase tracking-widest">应用启动舱</h2>
            <component :is="isLauncherCollapsed ? ChevronRight : ChevronDown" :size="12" class="text-gray-300 group-hover:text-gray-600" />
          </div>
          
          <div v-if="!isLauncherCollapsed" class="bg-white rounded-2xl border border-gray-200 p-5 flex flex-col flex-1 min-h-0 shadow-sm relative overflow-hidden animate-fade-in">
            <div class="flex items-center justify-between mb-6 shrink-0">
              <div class="flex items-center gap-3">
                <div class="p-2.5 bg-cyan-50 rounded-xl text-cyan-600 shadow-sm border border-cyan-100">
                  <Zap :size="20" />
                </div>
                <div>
                  <h3 class="text-sm font-bold text-gray-800">本地应用快捷拉起</h3>
                  <p class="text-[10px] text-gray-400 mt-0.5 leading-relaxed">一键异步调起开发常用程序</p>
                </div>
              </div>
              <button class="bg-gradient-to-r from-blue-600 to-indigo-600 text-white px-4 py-1.5 rounded-xl text-[10px] font-bold shadow-md shadow-blue-100 active:scale-95 flex items-center gap-2">
                <BrainCircuit :size="14" />
                AI 编排
              </button>
            </div>
            
            <div class="grid grid-cols-2 md:grid-cols-3 gap-4 overflow-visible lg:overflow-y-auto lg:pr-1 custom-scrollbar content-start flex-1">
              <div v-for="app in apps" :key="app.name" class="p-4 bg-gray-50 border border-gray-100 rounded-2xl hover:border-blue-400/50 hover:bg-white hover:shadow-md transition-all cursor-pointer group flex flex-col items-center text-center gap-3">
                <div class="p-3 rounded-2xl bg-white shadow-sm group-hover:scale-110 transition-transform shadow-inner" :class="app.color">
                  <component :is="app.icon" :size="22" />
                </div>
                <div class="min-w-0">
                  <div class="text-xs font-bold text-gray-800 truncate">{{ app.name }}</div>
                  <div class="text-[9px] text-gray-400 font-mono truncate mt-0.5 opacity-60">{{ app.path }}</div>
                </div>
              </div>
              <button class="aspect-square flex flex-col items-center justify-center gap-1.5 border-2 border-dashed border-gray-100 rounded-2xl hover:bg-gray-50 text-gray-300 hover:text-gray-400 transition-colors">
                <Plus :size="20" />
                <span class="text-[10px] font-bold">新增入口</span>
              </button>
            </div>
          </div>
        </div>

        <!-- Column 3: Port Killer -->
        <div class="col-span-12 lg:col-span-3 flex flex-col min-h-0 lg:h-full">
          <div @click="isPortKillerCollapsed = !isPortKillerCollapsed" class="flex items-center justify-between mb-2 px-1 shrink-0 cursor-pointer group">
            <h2 class="text-[11px] font-bold text-gray-400 uppercase tracking-widest">端口大盘</h2>
            <component :is="isPortKillerCollapsed ? ChevronRight : ChevronDown" :size="12" class="text-gray-300 group-hover:text-gray-600" />
          </div>
          
          <div v-if="!isPortKillerCollapsed" class="bg-white rounded-2xl border border-gray-200 p-5 flex flex-col flex-1 min-h-0 shadow-sm animate-fade-in">
            <div class="flex items-center justify-between mb-5 shrink-0">
              <div class="flex items-center gap-3">
                <div class="p-2 bg-blue-50 rounded-xl text-blue-600 shadow-sm border border-blue-100">
                  <AlertCircle :size="18" />
                </div>
                <div class="text-xs font-bold text-gray-800">端口扫描</div>
              </div>
              <button @click="doScanPorts" :disabled="isScanning" class="p-1.5 rounded-lg bg-blue-50 text-blue-600 hover:bg-blue-100 transition-all active:scale-95 disabled:opacity-50">
                <RefreshCw :size="14" :class="{'animate-spin': isScanning}" />
              </button>
            </div>
            
            <div class="relative mb-4 shrink-0">
              <Search :size="12" class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
              <input v-model="portSearch" type="text" placeholder="搜索端口..." class="w-full pl-9 pr-3 py-2 bg-gray-50 border border-gray-100 rounded-xl text-xs focus:outline-none focus:border-blue-400 transition-all shadow-inner" />
            </div>

            <div class="flex-1 overflow-y-auto pr-1 custom-scrollbar min-h-[150px]">
              <div v-if="filteredPorts.length === 0" class="h-full border-2 border-dashed border-gray-100 rounded-2xl flex flex-col items-center justify-center text-gray-300 p-4 text-center bg-gray-50/30">
                <Scan :size="24" class="mb-2 opacity-30" />
                <p class="text-[10px] leading-relaxed font-medium">点击刷新图标<br/>开始全量扫描</p>
              </div>
              <div v-else class="space-y-2">
                <div v-for="p in filteredPorts" :key="p.port" class="p-3 bg-gray-50 border border-gray-100 rounded-xl flex items-center justify-between group/item transition-all hover:bg-white hover:shadow-sm">
                  <div class="min-w-0">
                    <div class="flex items-center gap-2">
                      <span class="text-xs font-black font-mono" :class="p.isCritical ? 'text-gray-400' : 'text-blue-600'">:{{ p.port }}</span>
                      <span class="text-[11px] font-bold text-gray-700 truncate max-w-[80px]">{{ p.name }}</span>
                      <!-- Safety Badge -->
                      <span v-if="p.isCritical" class="text-[8px] px-1 py-0.5 rounded bg-gray-200 text-gray-500 font-bold uppercase tracking-tighter shrink-0">系统核心</span>
                      <span v-else class="text-[8px] px-1 py-0.5 rounded bg-green-100 text-green-600 font-bold uppercase tracking-tighter shrink-0">用户应用</span>
                    </div>
                    <div class="text-[9px] text-gray-400 font-mono truncate max-w-[120px]">PID: {{ p.pid }}</div>
                  </div>
                  <button 
                    v-if="!p.isCritical"
                    @click="onKillPort(p)" 
                    class="p-1.5 rounded-lg bg-red-50 text-red-500 hover:bg-red-500 hover:text-white transition-all opacity-0 group-hover/item:opacity-100 shadow-sm"
                    title="释放端口"
                  >
                    <Power :size="12" />
                  </button>
                  <div v-else class="p-1.5 text-gray-200 shrink-0">
                    <AlertCircle :size="12" />
                  </div>
                </div>
              </div>
            </div>
            
            <button @click="doScanPorts" :disabled="isScanning" class="mt-4 w-full py-2.5 bg-blue-600 text-white rounded-xl text-xs font-bold shadow-lg shadow-blue-100 active:scale-95 hover:bg-blue-700 transition-all flex items-center justify-center gap-2">
              <RefreshCw v-if="isScanning" :size="14" class="animate-spin" />
              {{ isScanning ? '正在排查...' : '全量排查' }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Bottom Section: Terminal Logs -->
    <div class="h-[260px] shrink-0 flex flex-col min-h-0 mt-2">
      <div class="flex items-center justify-between mb-2.5 px-1">
        <h2 class="text-[11px] font-bold text-gray-400 uppercase tracking-widest flex items-center gap-2">
          会话终端日志 <span class="text-[9px] font-normal opacity-40">SESSION TERMINAL LOGS</span>
        </h2>
        <div class="flex items-center gap-2">
          <div class="flex items-center gap-1.5 px-2 py-0.5 rounded-md bg-blue-50 text-blue-600 text-[9px] font-bold border border-blue-100 shadow-sm">
            <Zap :size="10" /> 内存实时推流
          </div>
        </div>
      </div>
      
      <div class="bg-white border border-gray-200 rounded-2xl overflow-hidden shadow-sm flex flex-col flex-1">
        <div ref="terminalRef" class="flex-1 p-5 font-mono text-[13px] overflow-y-auto space-y-2.5 custom-scrollbar bg-gray-50/30">
          <div v-if="allLogs.length === 0" class="text-gray-400 text-xs italic flex items-center gap-2">
            <Info :size="14"/> 等待任务流水推入...
          </div>
          <div v-for="(log, i) in allLogs" :key="i" class="flex items-start gap-4">
            <span class="text-blue-700 font-bold bg-blue-100/60 px-2 py-0.5 rounded text-[11px] shrink-0 select-none border border-blue-100/50">
              {{ log.time }}
            </span>
            <div class="flex-1 flex flex-wrap items-baseline gap-x-3">
              <span class="font-bold shrink-0 text-[12px] tracking-tight" :class="{
                'text-blue-600': log.level === 'info',
                'text-amber-600': log.level === 'warn',
                'text-red-600': log.level === 'error',
                'text-green-600': log.level === 'success'
              }">
                [{{ log.level.toUpperCase() }}]
              </span>
              <span class="text-cyan-700 italic shrink-0 font-bold text-[12px] opacity-90">{{ log.tag }}</span>
              <span class="text-gray-700 break-all leading-relaxed font-medium">{{ log.message }}</span>
            </div>
          </div>
        </div>
        
        <div class="px-4 py-2 bg-gray-50 border-t border-gray-100 flex items-center justify-between text-[10px] text-gray-400 font-mono font-bold">
          <div class="flex gap-4">
            <span class="flex items-center gap-1.5"><Cpu :size="12" class="text-blue-400" /> CPU: {{ cpu.toFixed(1) }}%</span>
            <span class="flex items-center gap-1.5"><MemoryStick :size="12" class="text-cyan-400" /> MEM: {{ memory.toFixed(1) }}%</span>
          </div>
          <div class="flex items-center gap-5">
            <span class="flex items-center gap-1.5 text-gray-400">ACTIVE: {{ activeTasks }}</span>
            <span class="flex items-center gap-1.5"><ArrowUp :size="10" class="text-orange-400"/> {{ formatSpeed(netUp) }}</span>
            <span class="flex items-center gap-1.5"><ArrowDown :size="10" class="text-blue-400"/> {{ formatSpeed(netDown) }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Modals Logic (Light Theme) -->
    <Teleport to="body">
      <div v-if="showPicker" class="fixed inset-0 z-50 flex items-center justify-center">
        <div class="absolute inset-0 bg-black/30 backdrop-blur-sm" @click="showPicker = false" />
        <div class="relative bg-white w-full max-w-lg mx-4 overflow-hidden rounded-3xl border border-gray-200 shadow-2xl animate-slide-up">
          <div class="flex items-center justify-between px-7 py-5 border-b border-gray-100">
            <h2 class="text-base font-bold text-gray-800">注册新任务实例</h2>
            <button @click="showPicker = false" class="text-gray-400 hover:text-gray-600 p-1.5 rounded-xl hover:bg-gray-50 transition-colors"><X :size="20" /></button>
          </div>
          <div class="p-4 max-h-[500px] overflow-y-auto space-y-2 custom-scrollbar">
            <button v-for="task in taskList" :key="task.ID" @click="selectTask(task)" class="w-full flex items-start gap-4 p-4 rounded-2xl border border-transparent hover:border-blue-200 hover:bg-blue-50/50 text-left transition-all group">
              <div class="w-12 h-12 rounded-xl bg-gray-50 flex items-center justify-center text-gray-400 group-hover:text-blue-600 transition-all shadow-inner group-hover:bg-white"><component :is="typeIcons[task.Type] || ScrollText" :size="24" /></div>
              <div class="flex-1 min-w-0">
                <div class="font-bold text-sm text-gray-800">{{ task.Name }}</div>
                <div class="text-[11px] text-gray-400 mt-1 line-clamp-2 leading-relaxed">{{ task.Description }}</div>
              </div>
              <span class="text-[10px] px-2 py-0.5 rounded-lg bg-blue-50 text-blue-600 font-bold border border-blue-100">{{ execModeLabel(task.ExecMode) }}</span>
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <Teleport to="body">
      <div v-if="showConfig" class="fixed inset-0 z-50 flex items-center justify-center">
        <div class="absolute inset-0 bg-black/30 backdrop-blur-sm" @click="closeConfig" />
        <div class="relative bg-white w-full max-w-md mx-4 overflow-hidden rounded-3xl border border-gray-200 shadow-2xl animate-slide-up max-h-[85vh] flex flex-col">
          <div class="flex items-center justify-between px-7 py-5 border-b border-gray-100 shrink-0">
            <h2 class="text-base font-bold text-gray-800">{{ editingSchedule ? '修改实例配置' : '初始化任务实例' }}</h2>
            <button @click="closeConfig" class="text-gray-400 hover:text-gray-600 p-1.5 rounded-xl hover:bg-gray-50 transition-colors"><X :size="20" /></button>
          </div>
          <div class="p-6 space-y-7 overflow-y-auto custom-scrollbar">
            <div class="flex items-center gap-3 p-3 bg-gray-50 rounded-2xl border border-gray-100 shadow-inner">
              <div class="p-2 bg-white rounded-lg shadow-sm text-blue-500"><component :is="typeIcons[selectedTask?.Type] || ScrollText" :size="18" /></div>
              <div class="text-sm font-bold text-gray-700">{{ selectedTask?.Name }}</div>
            </div>
            <div class="space-y-2">
              <label class="block text-[10px] font-bold text-gray-400 uppercase tracking-[0.2em]">实例备注 (Alias)</label>
              <input v-model="taskOption" type="text" placeholder="用于辨识实例..." class="w-full px-4 py-2.5 bg-gray-50 border border-gray-200 rounded-xl text-sm text-gray-800 placeholder:text-gray-300 focus:outline-none focus:border-blue-400 transition-all shadow-inner" />
            </div>
            <div class="space-y-2">
              <label class="block text-[10px] font-bold text-gray-400 uppercase tracking-[0.2em]">运行策略</label>
              <div class="grid grid-cols-2 gap-2 p-1 bg-gray-50 rounded-xl border border-gray-200 shadow-inner">
                <button @click="selectExecMode('manual')" :disabled="lockExecMode && chosenExecMode !== 'manual'" class="py-2 rounded-lg text-[11px] font-bold transition-all" :class="[chosenExecMode === 'manual' ? 'bg-white text-blue-600 shadow-sm border border-gray-100' : 'text-gray-400 opacity-40']">仅手动执行</button>
                <button @click="selectExecMode('schedule')" :disabled="lockExecMode && chosenExecMode !== 'schedule'" class="py-2 rounded-lg text-[11px] font-bold transition-all" :class="[chosenExecMode === 'schedule' ? 'bg-white text-blue-600 shadow-sm border border-gray-100' : 'text-gray-400 opacity-40']">手动+定时</button>
              </div>
            </div>
            <div v-if="chosenExecMode === 'schedule'" class="space-y-3 animate-fade-in">
              <label class="block text-[10px] font-bold text-gray-400 uppercase tracking-[0.2em]">Cron 表达式</label>
              <input v-model="cronExpr" type="text" placeholder="分 时 日 月 周" class="w-full px-4 py-2.5 bg-gray-50 border border-gray-200 rounded-xl text-sm font-mono text-gray-700 focus:outline-none focus:border-blue-400 shadow-inner" />
              <div v-if="cronError" class="text-[10px] text-red-500 font-bold px-1">{{ cronError }}</div>
              <div class="flex flex-wrap gap-2">
                <button v-for="p in CRON_PRESETS" :key="p.expr" @click="applyCronPreset(p.expr)" class="px-2.5 py-1 text-[10px] rounded-lg bg-gray-50 text-gray-500 border border-gray-200 hover:bg-blue-50 hover:text-blue-600 transition-colors shadow-sm">{{ p.label }}</button>
              </div>
            </div>
            <div class="h-px bg-gray-100" />
            <TaskConfigForm v-if="configFields.length > 0" ref="configForm" :fields="configFields" :initial-values="configInitial" />
            <button @click="submitConfig" class="w-full py-3.5 bg-blue-600 hover:bg-blue-700 text-white rounded-2xl text-sm font-bold shadow-lg shadow-blue-100 transition-all active:scale-95 mt-4">保存更新并同步</button>
          </div>
        </div>
      </div>
    </Teleport>

    <Teleport to="body">
      <div v-if="showDeleteConfirm" class="fixed inset-0 z-50 flex items-center justify-center">
        <div class="absolute inset-0 bg-black/30 backdrop-blur-sm" @click="showDeleteConfirm = false" />
        <div class="relative bg-white w-full max-w-sm mx-4 overflow-hidden rounded-3xl border border-gray-200 shadow-2xl animate-slide-up p-8 text-center space-y-5">
          <div class="w-16 h-16 bg-red-50 text-red-500 rounded-full flex items-center justify-center mx-auto border border-red-100 shadow-inner"><AlertCircle :size="32" /></div>
          <div>
            <h2 class="text-base font-bold text-gray-800">确认注销实例？</h2>
            <p class="text-[11px] text-gray-400 mt-2 leading-relaxed">删除「{{ scheduleToDelete?.Name }}」将永久移除其配置</p>
          </div>
          <div class="flex gap-4">
            <button @click="showDeleteConfirm = false" class="flex-1 py-3 bg-gray-50 text-gray-500 rounded-2xl text-sm font-bold border border-gray-200 hover:bg-gray-100 transition-all">取消</button>
            <button @click="doDelete" class="flex-1 py-3 bg-red-500 text-white rounded-2xl text-sm font-bold shadow-lg shadow-red-100 active:scale-95 hover:bg-red-600 transition-all">确认注销</button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.custom-scrollbar::-webkit-scrollbar { width: 3px; }
.custom-scrollbar::-webkit-scrollbar-track { background: transparent; }
.custom-scrollbar::-webkit-scrollbar-thumb { background: rgba(0, 0, 0, 0.05); border-radius: 10px; }
.custom-terminal-scrollbar::-webkit-scrollbar { width: 4px; }
.custom-terminal-scrollbar::-webkit-scrollbar-track { background: transparent; }
.custom-terminal-scrollbar::-webkit-scrollbar-thumb { background: rgba(0, 0, 0, 0.05); border-radius: 10px; }
@keyframes fade-in { from { opacity: 0; transform: translateY(5px); } to { opacity: 1; transform: translateY(0); } }
.animate-fade-in { animation: fade-in 0.3s ease-out forwards; }
</style>
