<script setup>
import { ref, onMounted } from 'vue'
import { Settings, Info, Gift, RefreshCw, CheckCircle, XCircle, Loader2 } from 'lucide-vue-next'

const appInfo = ref({
  name: 'KineticGo',
  version: '1.0.0',
  author: '迟暮',
  description: '基于 Wails + Vue 的本地任务调度桌面端。仪表盘实时展示 CPU / 内存 / 网络速率与活跃任务等系统指标;主面板支持按需创建、配置并定时调度各类常驻或一次性任务。',
  buildTime: '2026-05-17',
  features: [
    { title: '系统监控', desc: 'CPU、内存、上下行网速与活跃任务数实时刷新' },
    { title: '校园网自动连', desc: '检测网络状态,断线后自动完成 Portal 认证' },
    { title: '性能压测', desc: '对指定 URL 发起高并发 HTTP 压力测试' },
    { title: '延迟雷达', desc: '按频率探测目标 IP/域名的延迟与丢包率' },
    { title: '端口杀手', desc: '扫描并一键终结占用指定端口的进程' },
    { title: '652 定时签到', desc: '按计划自动完成 652 平台签到打卡' },
  ]
})

const updateStatus = ref({
  checking: false,
  hasUpdate: false,
  currentVersion: '1.0.0',
  latestVersion: '1.0.0',
  message: ''
})

const redeemCode = ref('')
const redeemResult = ref({ show: false, success: false, message: '' })
const redeeming = ref(false)

async function checkUpdate() {
  updateStatus.value.checking = true
  updateStatus.value.message = ''

  await new Promise(r => setTimeout(r, 1000))

  updateStatus.value.currentVersion = appInfo.value.version
  updateStatus.value.latestVersion = appInfo.value.version
  updateStatus.value.hasUpdate = false
  updateStatus.value.message = '已是最新版本'
  updateStatus.value.checking = false
}

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
    redeemResult.value = {
      show: true,
      success: true,
      message: '兑换成功！获得365天高级功能'
    }
    redeemCode.value = ''
  } else {
    redeemResult.value = {
      show: true,
      success: false,
      message: '无效的兑换码'
    }
  }
  redeeming.value = false
}

onMounted(() => {
  checkUpdate()
})
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
          <div class="flex items-center justify-between py-2 border-b border-dark-border/60">
            <span class="text-sm text-dark-muted/80">构建时间</span>
            <span class="text-sm font-medium">{{ appInfo.buildTime }}</span>
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
            <p v-if="updateStatus.message" class="text-xs mt-1" :class="updateStatus.hasUpdate ? 'text-yellow-400' : 'text-dark-muted/80'">
              {{ updateStatus.message }}
            </p>
          </div>
          <button
            @click="checkUpdate"
            :disabled="updateStatus.checking"
            class="inline-flex items-center gap-2 px-4 py-2 rounded-lg bg-black/[0.04] hover:bg-black/[0.07] text-dark-text text-sm font-medium transition-all disabled:opacity-50"
          >
            <Loader2 v-if="updateStatus.checking" :size="16" class="animate-spin" />
            <RefreshCw v-else :size="16" />
            {{ updateStatus.checking ? '检测中...' : '检查更新' }}
          </button>
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

          <!-- 兑换结果 -->
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
  </div>
</template>