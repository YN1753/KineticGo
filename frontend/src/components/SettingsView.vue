<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import {
  Settings, Info, Gift, RefreshCw, CheckCircle, XCircle, Loader2,
  Download, ExternalLink, X, AlertCircle, ChevronDown, ChevronRight, Cpu, MemoryStick, Wifi, Activity
} from 'lucide-vue-next'
import { useTaskApi } from '../composables/useTaskApi'

const { getVersion, checkUpdate, applyUpdate, fetchSystemTaskScheduleList, enableSystemTask, disableSystemTask, runningIds, fetchRunningIds } = useTaskApi()

const appInfo = ref({
  name: 'KineticGo',
  version: 'dev',
  author: '迟暮',
  description: '基于 Wails + Vue 的本地任务调度桌面端。仪表盘实时展示 CPU / 内存 / 网络速率与活跃任务等系统指标;主面板支持按需创建、配置并定时调度各类常驻或一次性任务。',
  features: [
    { title: '系统监控', desc: 'CPU、内存、上下行网速与活跃任务数实时刷新' },
    { title: '校园网自动连', desc: '检测网络状态,断线后自动完成 Portal 认证' },
    { title: '性能压测', desc: '对指定 URL 发起高并发 HTTP 压力测试' },
    { title: '延迟雷达', desc: '按频率探测目标 IP/域名的延迟与丢包率' },
    { title: '端口杀手', desc: '扫描并一键终结占用指定端口的进程' },
  ]
})

// 版本检测
const updateStatus = ref({
  checking: false,
  hasUpdate: false,
  currentVersion: 'dev',
  latestVersion: '',
  message: ''
})

const showUpdateModal = ref(false)
const updateInfo = ref(null)
const updating = ref(false)
const updateError = ref('')

async function fetchVersion() {
  try {
    const ver = await getVersion()
    appInfo.value.version = ver || 'dev'
    updateStatus.value.currentVersion = appInfo.value.version
  } catch {
    appInfo.value.version = 'dev'
  }
}

async function doCheckUpdate() {
  updateStatus.value.checking = true
  updateStatus.value.message = ''
  updateError.value = ''

  const info = await checkUpdate()
  updateStatus.value.checking = false

  if (!info) {
    updateStatus.value.message = '检查失败，请稍后再试'
    return
  }

  updateStatus.value.currentVersion = info.CurrentVer || appInfo.value.version
  updateStatus.value.latestVersion = info.LatestVer || ''

  if (!info.HasUpdate) {
    updateStatus.value.hasUpdate = false
    updateStatus.value.message = '已是最新版本'
    return
  }

  updateStatus.value.hasUpdate = true
  updateInfo.value = info
  showUpdateModal.value = true
}

async function doApplyUpdate() {
  if (!updateInfo.value?.DownloadURL) {
    updateError.value = '未找到下载链接'
    return
  }
  updating.value = true
  updateError.value = ''
  try {
    await applyUpdate(updateInfo.value.DownloadURL)
  } catch (e) {
    updateError.value = e?.message || '打开下载页失败，请手动访问'
  } finally {
    updating.value = false
  }
}

// 兑换码
const redeemCode = ref('')
const redeemResult = ref({ show: false, success: false, message: '' })
const redeeming = ref(false)

async function redeem() {
  if (!redeemCode.value.trim()) {
    redeemResult.value = { show: true, success: false, message: '请输入兑换码' }
    return
  }
  redeeming.value = true
  redeemResult.value.show = false
  await new Promise(r => setTimeout(r, 800))
  const code = redeemCode.value.trim().toUpperCase()
  if (['DEV2026', 'VIP1Y', 'WAILSDEV'].includes(code)) {
    redeemResult.value = { show: true, success: true, message: '兑换成功！获得365天高级功能' }
    redeemCode.value = ''
  } else {
    redeemResult.value = { show: true, success: false, message: '无效的兑换码' }
  }
  redeeming.value = false
}

onMounted(async () => {
  fetchVersion()
  await fetchSystemTasks()
  await fetchRunningIds()
  pollTimer = setInterval(fetchRunningIds, 2000)
})

onUnmounted(() => {
  if (pollTimer) clearInterval(pollTimer)
})

// 系统任务
const systemTasksExpanded = ref(false)
const systemTaskList = ref([])
const systemTasksLoading = ref(false)
let pollTimer = null

const systemTaskIcons = {
  'system-local_cpu': Cpu,
  'system-local_memory': MemoryStick,
  'system-local_network': Wifi,
  'system-active_tasks': Activity,
}

async function fetchSystemTasks() {
  systemTasksLoading.value = true
  try {
    systemTaskList.value = await fetchSystemTaskScheduleList()
  } finally {
    systemTasksLoading.value = false
  }
}

async function toggleSystemTask(task) {
  const isRunning = runningIds.value.has(task.ID)
  if (isRunning) {
    await disableSystemTask(task.ID)
  } else {
    await enableSystemTask(task.ID)
  }
  await fetchRunningIds()
  await fetchSystemTasks()
}

function isSystemTaskRunning(taskId) {
  return runningIds.value.has(taskId)
}
</script>

<template>
  <div class="min-h-full bg-gray-50">
    <div class="sticky top-0 z-30 bg-white/80 backdrop-blur-xl border-b border-gray-200">
      <div class="max-w-7xl mx-auto px-6 py-3">
        <h1 class="text-base font-semibold text-gray-800">系统设置</h1>
      </div>
    </div>

    <div class="max-w-3xl mx-auto px-6 py-6 space-y-6">
      <!-- 项目介绍 -->
      <div class="bg-white rounded-2xl border border-gray-200 shadow-sm p-6 transition-all hover:border-blue-200">
        <div class="flex items-center gap-3 mb-4">
          <div class="inline-flex items-center justify-center w-10 h-10 rounded-xl bg-gray-50 border border-gray-100">
            <Info :size="20" class="text-blue-500" />
          </div>
          <div>
            <h2 class="text-sm font-bold text-gray-800">项目介绍</h2>
            <p class="text-[10px] text-gray-400 uppercase tracking-widest font-mono">About KineticGo</p>
          </div>
        </div>

        <div class="space-y-3">
          <div class="flex items-center justify-between py-2 border-b border-gray-50">
            <span class="text-sm text-gray-500">项目名称</span>
            <span class="text-sm font-bold text-gray-700">{{ appInfo.name }}</span>
          </div>
          <div class="flex items-center justify-between py-2 border-b border-gray-50">
            <span class="text-sm text-gray-500">当前版本</span>
            <span class="text-[11px] font-mono px-2 py-0.5 rounded bg-gray-100 text-gray-600">v{{ appInfo.version }}</span>
          </div>
          <div class="flex items-center justify-between py-2 border-b border-gray-50">
            <span class="text-sm text-gray-500">作者</span>
            <span class="text-sm font-bold text-gray-700">{{ appInfo.author }}</span>
          </div>
          <div class="pt-2 pb-3 border-b border-gray-50">
            <span class="text-sm text-gray-500">项目描述</span>
            <p class="text-sm text-gray-600 mt-1.5 leading-relaxed">{{ appInfo.description }}</p>
          </div>
          <div class="pt-2">
            <span class="text-sm text-gray-500">主要功能</span>
            <ul class="mt-2.5 space-y-2">
              <li
                v-for="(feat, idx) in appInfo.features"
                :key="idx"
                class="flex items-start gap-2.5 text-sm leading-relaxed"
              >
                <span class="mt-1.5 w-1.5 h-1.5 rounded-full bg-blue-400 shrink-0 shadow-[0_0_8px_rgba(96,165,250,0.5)]"></span>
                <span class="text-gray-600">
                  <span class="font-bold text-gray-800">{{ feat.title }}</span>
                  <span class="text-gray-300 mx-1.5">/</span>{{ feat.desc }}
                </span>
              </li>
            </ul>
          </div>
        </div>
      </div>

      <!-- 版本检测 -->
      <div class="bg-white rounded-2xl border border-gray-200 shadow-sm p-6 transition-all hover:border-blue-200">
        <div class="flex items-center gap-3 mb-4">
          <div class="inline-flex items-center justify-center w-10 h-10 rounded-xl bg-gray-50 border border-gray-100">
            <RefreshCw :size="20" class="text-blue-500" />
          </div>
          <div>
            <h2 class="text-sm font-bold text-gray-800">版本检测</h2>
            <p class="text-[10px] text-gray-400 uppercase tracking-widest font-mono">Check for Updates</p>
          </div>
        </div>

        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm text-gray-600">
              当前运行版本: <span class="font-bold text-gray-800">v{{ updateStatus.currentVersion }}</span>
            </p>
            <p
              v-if="updateStatus.message"
              class="text-xs mt-1"
              :class="updateStatus.hasUpdate ? 'text-amber-600 font-bold' : 'text-gray-400'"
            >
              {{ updateStatus.message }}
            </p>
          </div>
          <button
            @click="doCheckUpdate"
            :disabled="updateStatus.checking"
            class="inline-flex items-center gap-2 px-4 py-2 rounded-lg bg-gray-50 hover:bg-gray-100 text-gray-600 text-sm font-bold transition-all disabled:opacity-50 border border-gray-100 shadow-sm active:scale-95"
          >
            <Loader2 v-if="updateStatus.checking" :size="16" class="animate-spin text-blue-500" />
            <RefreshCw v-else :size="16" />
            {{ updateStatus.checking ? '检测中...' : '检查更新' }}
          </button>
        </div>
      </div>

      <!-- 系统任务 -->
      <div class="bg-white rounded-2xl border border-gray-200 shadow-sm p-6 transition-all hover:border-blue-200">
        <button
          @click="systemTasksExpanded = !systemTasksExpanded"
          class="w-full flex items-center gap-3 mb-4"
        >
          <div class="inline-flex items-center justify-center w-10 h-10 rounded-xl bg-gray-50 border border-gray-100">
            <Cpu :size="20" class="text-blue-500" />
          </div>
          <div class="flex-1 text-left">
            <h2 class="text-sm font-bold text-gray-800">系统监控任务</h2>
            <p class="text-[10px] text-gray-400 uppercase tracking-widest font-mono">System Background Tasks</p>
          </div>
          <component :is="systemTasksExpanded ? ChevronDown : ChevronRight" :size="18" class="text-gray-300" />
        </button>

        <div v-if="systemTasksExpanded" class="space-y-3 animate-fade-in">
          <div v-if="systemTasksLoading" class="py-4 text-center text-gray-400 text-sm italic">正在检索...</div>
          <div v-else-if="systemTaskList.length === 0" class="py-4 text-center text-gray-400 text-sm italic">暂无可用系统级任务</div>
          <div
            v-for="task in systemTaskList"
            :key="task.ID"
            class="flex items-center justify-between p-4 rounded-xl bg-gray-50 border border-gray-100 transition-colors hover:bg-white"
          >
            <div class="flex items-center gap-3">
              <div class="w-8 h-8 rounded-lg bg-white shadow-sm flex items-center justify-center border border-gray-100">
                <component :is="systemTaskIcons[task.TaskType] || Settings" :size="16" class="text-gray-500" />
              </div>
              <div>
                <div class="text-sm font-bold text-gray-700">{{ task.Name }}</div>
                <div class="text-[10px] text-gray-400 font-mono">{{ task.TaskType }}</div>
              </div>
            </div>
            <button
              @click="toggleSystemTask(task)"
              class="px-4 py-1.5 rounded-lg text-xs font-bold transition-all shadow-sm active:scale-95"
              :class="isSystemTaskRunning(task.ID)
                ? 'bg-red-50 text-red-600 hover:bg-red-100 border border-red-100'
                : 'bg-green-50 text-green-600 hover:bg-green-100 border border-green-100'"
            >
              {{ isSystemTaskRunning(task.ID) ? '停止' : '启动' }}
            </button>
          </div>
        </div>
      </div>

      <!-- 兑换码 -->
      <div class="bg-white rounded-2xl border border-gray-200 shadow-sm p-6 transition-all hover:border-blue-200">
        <div class="flex items-center gap-3 mb-4">
          <div class="inline-flex items-center justify-center w-10 h-10 rounded-xl bg-gray-50 border border-gray-100">
            <Gift :size="20" class="text-blue-500" />
          </div>
          <div>
            <h2 class="text-sm font-bold text-gray-800">高级功能兑换</h2>
            <p class="text-[10px] text-gray-400 uppercase tracking-widest font-mono">Redeem Activation Code</p>
          </div>
        </div>

        <div class="space-y-4">
          <div class="flex gap-3">
            <input
              v-model="redeemCode"
              type="text"
              placeholder="请输入激活码..."
              class="flex-1 px-4 py-2.5 rounded-xl bg-gray-50 border border-gray-100 text-sm text-gray-800 placeholder:text-gray-300 focus:outline-none focus:border-blue-300 transition-all shadow-inner"
              @keyup.enter="redeem"
            />
            <button
              @click="redeem"
              :disabled="redeeming"
              class="px-6 py-2.5 rounded-xl bg-blue-600 hover:bg-blue-700 text-white text-sm font-bold transition-all disabled:opacity-50 shadow-md shadow-blue-100 active:scale-95"
            >
              <Loader2 v-if="redeeming" :size="16" class="animate-spin" />
              <span v-else>立即兑换</span>
            </button>
          </div>

          <div
            v-if="redeemResult.show"
            class="flex items-center gap-3 p-4 rounded-xl border animate-fade-in"
            :class="redeemResult.success ? 'bg-green-50 border-green-100' : 'bg-red-50 border-red-100'"
          >
            <CheckCircle v-if="redeemResult.success" :size="18" class="text-green-500" />
            <XCircle v-else :size="18" class="text-red-500" />
            <span class="text-sm font-medium" :class="redeemResult.success ? 'text-green-600' : 'text-red-600'">
              {{ redeemResult.message }}
            </span>
          </div>
        </div>
      </div>
    </div>

    <!-- Update Modal -->
    <Teleport to="body">
      <div v-if="showUpdateModal" class="fixed inset-0 z-50 flex items-center justify-center">
        <div class="absolute inset-0 bg-black/30 backdrop-blur-sm" @click="showUpdateModal = false" />
        <div class="relative bg-white w-full max-w-lg mx-4 overflow-hidden rounded-2xl border border-gray-200 shadow-2xl animate-slide-up">
          <div class="flex items-center justify-between px-6 py-5 border-b border-gray-100">
            <div class="flex items-center gap-2">
              <AlertCircle :size="18" class="text-amber-500" />
              <h2 class="text-base font-bold text-gray-800">发现核心版本更新</h2>
            </div>
            <button @click="showUpdateModal = false" class="text-gray-400 hover:text-gray-600 transition-colors p-1.5 rounded-lg hover:bg-gray-50">
              <X :size="18" />
            </button>
          </div>

          <div class="p-6 space-y-6">
            <div class="flex items-center gap-4 text-sm">
              <div class="flex-1 p-3 rounded-xl bg-gray-50 border border-gray-100 text-center">
                <p class="text-[10px] text-gray-400 uppercase font-bold">当前版本</p>
                <p class="text-base font-mono font-bold text-gray-600">v{{ updateInfo?.CurrentVer }}</p>
              </div>
              <div class="p-2 rounded-full bg-blue-50 text-blue-500">
                <ChevronRight :size="16" />
              </div>
              <div class="flex-1 p-3 rounded-xl bg-blue-50 border border-blue-100 text-center">
                <p class="text-[10px] text-blue-400 uppercase font-bold">最新版本</p>
                <p class="text-base font-mono font-bold text-blue-600">v{{ updateInfo?.LatestVer }}</p>
              </div>
            </div>

            <div class="rounded-xl border border-gray-100 bg-gray-50 overflow-hidden shadow-inner">
              <div class="px-4 py-2 border-b border-gray-100 bg-white/50 text-[10px] font-bold text-gray-400 uppercase tracking-widest">
                更新日志 / Release Notes
              </div>
              <div class="px-4 py-3 text-sm text-gray-600 whitespace-pre-wrap max-h-48 overflow-y-auto leading-relaxed custom-scrollbar">
                {{ updateInfo?.ReleaseNotes || '暂无详细更新说明' }}
              </div>
            </div>

            <div v-if="updateError" class="p-3 rounded-lg bg-red-50 text-red-600 text-xs flex items-center gap-2 border border-red-100">
              <XCircle :size="14" />
              {{ updateError }}
            </div>

            <div class="flex gap-4 pt-2">
              <button
                @click="showUpdateModal = false"
                class="flex-1 py-3 bg-gray-100 hover:bg-gray-200 text-gray-600 rounded-xl text-sm font-bold transition-all active:scale-95"
              >
                稍后再说
              </button>
              <button
                @click="doApplyUpdate"
                :disabled="updating || !updateInfo?.DownloadURL"
                class="flex-1 py-3 bg-blue-600 hover:bg-blue-700 text-white rounded-xl text-sm font-bold shadow-lg shadow-blue-100 transition-all disabled:opacity-50 flex items-center justify-center gap-2 active:scale-95"
              >
                <ExternalLink v-if="!updating" :size="14" />
                <Loader2 v-else :size="14" class="animate-spin" />
                {{ updating ? '正在尝试访问...' : '立即前往更新' }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
@keyframes fade-in {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}

.animate-fade-in {
  animation: fade-in 0.4s ease-out forwards;
}

.custom-scrollbar::-webkit-scrollbar {
  width: 4px;
}
.custom-scrollbar::-webkit-scrollbar-track {
  background: transparent;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background: rgba(0, 0, 0, 0.05);
  border-radius: 10px;
}
</style>
