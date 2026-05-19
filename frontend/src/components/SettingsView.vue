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

const isWindows = navigator.platform.toLowerCase().includes('win')

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
    // Windows 成功会触发 os.Exit，不会走到这里
  } catch (e) {
    // macOS 或其他错误：fallback 到浏览器打开
    if (!isWindows) {
      window.open(updateInfo.value.DownloadURL, '_blank')
    } else {
      updateError.value = e?.message || '更新失败，请手动下载'
    }
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
  <div class="min-h-full">
    <div class="sticky top-0 z-30 bg-dark-bg/85 backdrop-blur-xl border-b border-dark-border">
      <div class="max-w-7xl mx-auto px-6 py-3">
        <h1 class="text-base font-semibold">系统设置</h1>
      </div>
    </div>

    <div class="max-w-3xl mx-auto px-6 py-6 space-y-6">
      <!-- 项目介绍 -->
      <div class="glass-panel p-6">
        <div class="flex items-center gap-3 mb-4">
          <div class="inline-flex items-center justify-center w-10 h-10 rounded-xl bg-black/[0.04]">
            <Info :size="20" class="text-dark-muted" />
          </div>
          <div>
            <h2 class="text-sm font-medium">项目介绍</h2>
            <p class="text-xs text-dark-muted">About KineticGo</p>
          </div>
        </div>

        <div class="space-y-3">
          <div class="flex items-center justify-between py-2 border-b border-dark-border/60">
            <span class="text-sm text-dark-muted/80">项目名称</span>
            <span class="text-sm font-medium">{{ appInfo.name }}</span>
          </div>
          <div class="flex items-center justify-between py-2 border-b border-dark-border/60">
            <span class="text-sm text-dark-muted/80">当前版本</span>
            <span class="text-sm font-medium">v{{ appInfo.version }}</span>
          </div>
          <div class="flex items-center justify-between py-2 border-b border-dark-border/60">
            <span class="text-sm text-dark-muted/80">作者</span>
            <span class="text-sm font-medium">{{ appInfo.author }}</span>
          </div>
          <div class="pt-2 pb-3 border-b border-dark-border/60">
            <span class="text-sm text-dark-muted/80">项目描述</span>
            <p class="text-sm text-dark-text/80 mt-1.5 leading-relaxed">{{ appInfo.description }}</p>
          </div>
          <div class="pt-2">
            <span class="text-sm text-dark-muted/80">主要功能</span>
            <ul class="mt-2.5 space-y-2">
              <li
                v-for="(feat, idx) in appInfo.features"
                :key="idx"
                class="flex items-start gap-2.5 text-sm leading-relaxed"
              >
                <span class="mt-1.5 w-1 h-1 rounded-full bg-accent-blue/60 shrink-0"></span>
                <span class="text-dark-muted">
                  <span class="font-medium text-dark-text">{{ feat.title }}</span>
                  <span class="text-dark-muted/60 mx-1.5">·</span>{{ feat.desc }}
                </span>
              </li>
            </ul>
          </div>
        </div>
      </div>

      <!-- 版本检测 -->
      <div class="glass-panel p-6">
        <div class="flex items-center gap-3 mb-4">
          <div class="inline-flex items-center justify-center w-10 h-10 rounded-xl bg-black/[0.04]">
            <RefreshCw :size="20" class="text-dark-muted" />
          </div>
          <div>
            <h2 class="text-sm font-medium">版本检测</h2>
            <p class="text-xs text-dark-muted">Check for Updates</p>
          </div>
        </div>

        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm">
              当前版本: <span class="font-medium">v{{ updateStatus.currentVersion }}</span>
            </p>
            <p
              v-if="updateStatus.message"
              class="text-xs mt-1"
              :class="updateStatus.hasUpdate ? 'text-accent-amber' : 'text-dark-muted/80'"
            >
              {{ updateStatus.message }}
            </p>
          </div>
          <button
            @click="doCheckUpdate"
            :disabled="updateStatus.checking"
            class="inline-flex items-center gap-2 px-4 py-2 rounded-lg bg-black/[0.04] hover:bg-black/[0.07] text-dark-text text-sm font-medium transition-all disabled:opacity-50"
          >
            <Loader2 v-if="updateStatus.checking" :size="16" class="animate-spin" />
            <RefreshCw v-else :size="16" />
            {{ updateStatus.checking ? '检测中...' : '检查更新' }}
          </button>
        </div>
      </div>

      <!-- 系统任务 -->
      <div class="glass-panel p-6">
        <button
          @click="systemTasksExpanded = !systemTasksExpanded"
          class="w-full flex items-center gap-3 mb-4"
        >
          <div class="inline-flex items-center justify-center w-10 h-10 rounded-xl bg-black/[0.04]">
            <Cpu :size="20" class="text-dark-muted" />
          </div>
          <div class="flex-1 text-left">
            <h2 class="text-sm font-medium">系统任务</h2>
            <p class="text-xs text-dark-muted">System Tasks</p>
          </div>
          <component :is="systemTasksExpanded ? ChevronDown : ChevronRight" :size="18" class="text-dark-muted" />
        </button>

        <div v-if="systemTasksExpanded" class="space-y-3">
          <div v-if="systemTasksLoading" class="py-4 text-center text-dark-muted text-sm">加载中...</div>
          <div v-else-if="systemTaskList.length === 0" class="py-4 text-center text-dark-muted text-sm">暂无系统任务</div>
          <div
            v-for="task in systemTaskList"
            :key="task.ID"
            class="flex items-center justify-between p-4 rounded-xl bg-black/[0.03] border border-dark-border/60"
          >
            <div class="flex items-center gap-3">
              <div class="w-8 h-8 rounded-lg bg-black/[0.04] flex items-center justify-center">
                <component :is="systemTaskIcons[task.TaskType] || Settings" :size="16" class="text-dark-muted" />
              </div>
              <div>
                <div class="text-sm font-medium text-dark-text">{{ task.Name }}</div>
                <div class="text-xs text-dark-muted">{{ task.TaskType }}</div>
              </div>
            </div>
            <button
              @click="toggleSystemTask(task)"
              class="px-3 py-1.5 rounded-lg text-xs font-medium transition-all"
              :class="isSystemTaskRunning(task.ID)
                ? 'bg-accent-red/20 text-accent-red hover:bg-accent-red/30'
                : 'bg-accent-green/20 text-accent-green hover:bg-accent-green/30'"
            >
              {{ isSystemTaskRunning(task.ID) ? '停止' : '启动' }}
            </button>
          </div>
        </div>
      </div>

      <!-- 兑换码 -->
      <div class="glass-panel p-6">
        <div class="flex items-center gap-3 mb-4">
          <div class="inline-flex items-center justify-center w-10 h-10 rounded-xl bg-black/[0.04]">
            <Gift :size="20" class="text-dark-muted" />
          </div>
          <div>
            <h2 class="text-sm font-medium">兑换码</h2>
            <p class="text-xs text-dark-muted">Redeem Code</p>
          </div>
        </div>

        <div class="space-y-4">
          <div class="flex gap-3">
            <input
              v-model="redeemCode"
              type="text"
              placeholder="请输入兑换码"
              class="flex-1 px-4 py-2.5 rounded-lg bg-black/[0.03] border border-dark-border text-sm text-dark-text placeholder:text-dark-muted focus:outline-none focus:border-accent-blue/40 transition-colors"
              @keyup.enter="redeem"
            />
            <button
              @click="redeem"
              :disabled="redeeming"
              class="px-5 py-2.5 rounded-lg bg-black/[0.06] hover:bg-black/[0.1] text-dark-text text-sm font-medium transition-all disabled:opacity-50"
            >
              <Loader2 v-if="redeeming" :size="16" class="animate-spin" />
              <span v-else>兑换</span>
            </button>
          </div>

          <div
            v-if="redeemResult.show"
            class="flex items-center gap-3 p-4 rounded-lg"
            :class="redeemResult.success ? 'bg-green-500/10' : 'bg-red-500/10'"
          >
            <CheckCircle v-if="redeemResult.success" :size="18" class="text-green-400" />
            <XCircle v-else :size="18" class="text-red-400" />
            <span class="text-sm" :class="redeemResult.success ? 'text-green-400' : 'text-red-400'">
              {{ redeemResult.message }}
            </span>
          </div>
        </div>
      </div>
    </div>

    <!-- Update Modal -->
    <Teleport to="body">
      <div v-if="showUpdateModal" class="fixed inset-0 z-50 flex items-center justify-center">
        <div class="absolute inset-0 bg-black/40 backdrop-blur-sm" @click="showUpdateModal = false" />
        <div class="relative glass-panel w-full max-w-lg mx-4 overflow-hidden animate-slide-up">
          <div class="flex items-center justify-between px-6 py-4 border-b border-dark-border">
            <div class="flex items-center gap-2">
              <AlertCircle :size="18" class="text-accent-amber" />
              <h2 class="text-base font-semibold">发现新版本</h2>
            </div>
            <button @click="showUpdateModal = false" class="text-dark-muted hover:text-dark-text transition-colors">
              <X :size="18" />
            </button>
          </div>

          <div class="p-6 space-y-4">
            <div class="flex items-center gap-3 text-sm">
              <span class="text-dark-muted">当前版本</span>
              <span class="text-dark-text font-medium">v{{ updateInfo?.CurrentVer }}</span>
              <span class="text-dark-muted/40">→</span>
              <span class="text-accent-blue font-medium">v{{ updateInfo?.LatestVer }}</span>
            </div>

            <div class="rounded-xl border border-dark-border bg-black/[0.02] overflow-hidden">
              <div class="px-4 py-2 border-b border-dark-border/60 text-xs font-medium text-dark-muted uppercase tracking-wider">
                更新内容
              </div>
              <div class="px-4 py-3 text-sm text-dark-text/80 whitespace-pre-wrap max-h-48 overflow-y-auto leading-relaxed">
                {{ updateInfo?.ReleaseNotes || '暂无更新说明' }}
              </div>
            </div>

            <div v-if="updateError" class="text-xs text-accent-red flex items-center gap-1.5">
              <XCircle :size="12" />
              {{ updateError }}
            </div>

            <div class="flex gap-3 pt-1">
              <button
                @click="showUpdateModal = false"
                class="flex-1 py-2.5 bg-black/[0.04] hover:bg-black/[0.07] text-dark-text rounded-xl text-sm font-medium transition-colors"
              >
                稍后再说
              </button>
              <button
                v-if="isWindows"
                @click="doApplyUpdate"
                :disabled="updating || !updateInfo?.DownloadURL"
                class="flex-1 py-2.5 bg-accent-blue hover:bg-accent-blue/90 text-white rounded-xl text-sm font-medium transition-all disabled:opacity-50 flex items-center justify-center gap-1.5"
              >
                <Loader2 v-if="updating" :size="14" class="animate-spin" />
                <Download v-else :size="14" />
                {{ updating ? '正在更新...' : '立即更新并重启' }}
              </button>
              <button
                v-else
                @click="window.open(updateInfo?.DownloadURL, '_blank')"
                class="flex-1 py-2.5 bg-accent-blue hover:bg-accent-blue/90 text-white rounded-xl text-sm font-medium transition-all flex items-center justify-center gap-1.5"
              >
                <ExternalLink :size="14" />
                前往下载页
              </button>
            </div>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>
