<script setup>
import { ref } from 'vue'
import { ScrollText, ChevronDown, ChevronRight, Clock, AlertCircle, CheckCircle2, XCircle } from 'lucide-vue-next'

const executions = ref([
  { id: 1, taskName: '校园网自动连', triggerType: 'auto', status: 'success', resultSummary: '登录成功', startTime: '2026-05-16 14:02:13', endTime: '2026-05-16 14:02:15', logs: [
    { time: '14:02:13', level: 'info', message: '检测网络连接状态' },
    { time: '14:02:13', level: 'warn', message: '发现网络掉线' },
    { time: '14:02:14', level: 'info', message: '正在执行自动登录...' },
    { time: '14:02:15', level: 'info', message: '认证成功，网络已恢复' },
  ]},
  { id: 2, taskName: '性能压测', triggerType: 'manual', status: 'success', resultSummary: 'QPS avg=342 p99=89ms', startTime: '2026-05-16 13:30:00', endTime: '2026-05-16 13:31:00', logs: [
    { time: '13:30:00', level: 'info', message: '初始化压测引擎，目标: https://api.example.com' },
    { time: '13:30:01', level: 'info', message: '建立连接池，并发数: 50' },
    { time: '13:30:15', level: 'info', message: 'QPS: 342, 延迟 avg=23ms p99=89ms' },
    { time: '13:31:00', level: 'info', message: '压测完成，总请求: 20520, 失败: 0' },
  ]},
  { id: 3, taskName: '延迟雷达', triggerType: 'auto', status: 'warning', resultSummary: '检测到延迟波动', startTime: '2026-05-16 12:00:00', endTime: '2026-05-16 12:05:00', logs: [
    { time: '12:00:00', level: 'info', message: '启动延迟探测: 114.114.114.114' },
    { time: '12:01:30', level: 'warn', message: '延迟升高: avg=120ms (正常 <30ms)' },
    { time: '12:03:00', level: 'warn', message: '丢包率: 2.3%' },
    { time: '12:05:00', level: 'info', message: '延迟恢复正常: avg=15ms' },
  ]},
])

const expandedId = ref(null)

function toggleExpand(id) {
  expandedId.value = expandedId.value === id ? null : id
}

function statusIcon(status) {
  if (status === 'success') return CheckCircle2
  if (status === 'warning') return AlertCircle
  return XCircle
}

function statusClass(status) {
  if (status === 'success') return 'text-accent-green'
  if (status === 'warning') return 'text-accent-amber'
  return 'text-accent-red'
}

function levelClass(level) {
  if (level === 'warn') return 'text-accent-amber'
  if (level === 'error') return 'text-accent-red'
  return 'text-dark-muted'
}
</script>

<template>
  <div class="min-h-full">
    <div class="sticky top-0 z-30 bg-dark-bg/85 backdrop-blur-xl border-b border-dark-border">
      <div class="max-w-7xl mx-auto px-6 py-3">
        <h1 class="text-base font-semibold">历史日志</h1>
      </div>
    </div>

    <div class="max-w-7xl mx-auto px-6 py-6">
      <div v-if="executions.length === 0" class="py-24 text-center">
        <div class="inline-flex items-center justify-center w-16 h-16 rounded-2xl bg-black/[0.04] mb-4">
          <ScrollText :size="28" class="text-dark-muted/50" />
        </div>
        <div class="text-dark-muted text-sm">暂无执行记录</div>
      </div>

      <div v-else class="glass-panel overflow-hidden">
        <table class="w-full text-sm">
          <thead>
            <tr class="border-b border-dark-border text-dark-muted text-xs uppercase tracking-wider">
              <th class="text-left px-5 py-3 font-medium w-8"></th>
              <th class="text-left px-5 py-3 font-medium">任务</th>
              <th class="text-left px-5 py-3 font-medium">触发</th>
              <th class="text-left px-5 py-3 font-medium">状态</th>
              <th class="text-left px-5 py-3 font-medium">结果</th>
              <th class="text-left px-5 py-3 font-medium">开始时间</th>
              <th class="text-left px-5 py-3 font-medium">耗时</th>
            </tr>
          </thead>
          <tbody>
            <template v-for="exec in executions" :key="exec.id">
              <tr
                @click="toggleExpand(exec.id)"
                class="border-b border-dark-border hover:bg-black/[0.03] cursor-pointer transition-colors"
              >
                <td class="px-5 py-3 text-dark-muted">
                  <component :is="expandedId === exec.id ? ChevronDown : ChevronRight" :size="14" />
                </td>
                <td class="px-5 py-3 font-medium text-dark-text">{{ exec.taskName }}</td>
                <td class="px-5 py-3">
                  <span class="text-xs px-2 py-0.5 rounded-full bg-black/[0.04] text-dark-muted">
                    {{ exec.triggerType === 'manual' ? '手动' : '自动' }}
                  </span>
                </td>
                <td class="px-5 py-3">
                  <span class="flex items-center gap-1.5" :class="statusClass(exec.status)">
                    <component :is="statusIcon(exec.status)" :size="14" />
                    <span class="text-xs">{{ exec.status === 'success' ? '成功' : exec.status === 'warning' ? '警告' : '失败' }}</span>
                  </span>
                </td>
                <td class="px-5 py-3 text-dark-muted text-xs">{{ exec.resultSummary }}</td>
                <td class="px-5 py-3 text-dark-muted text-xs font-mono">{{ exec.startTime }}</td>
                <td class="px-5 py-3 text-dark-muted text-xs font-mono">
                  {{ exec.endTime && exec.startTime ? Math.round((new Date(exec.endTime) - new Date(exec.startTime)) / 1000) + 's' : '-' }}
                </td>
              </tr>
              <tr v-if="expandedId === exec.id">
                <td colspan="7" class="px-5 py-0">
                  <div class="mini-console rounded-xl border border-dark-border my-3 overflow-hidden">
                    <div class="px-4 py-3 space-y-1">
                      <div
                        v-for="(log, i) in exec.logs"
                        :key="i"
                        class="flex gap-3 text-xs animate-fade-in"
                      >
                        <span class="text-dark-muted/50 shrink-0 font-mono">{{ log.time }}</span>
                        <span class="shrink-0 w-8 text-right" :class="levelClass(log.level)">
                          {{ log.level.toUpperCase() }}
                        </span>
                        <span class="text-dark-text/80">{{ log.message }}</span>
                      </div>
                    </div>
                  </div>
                </td>
              </tr>
            </template>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>
