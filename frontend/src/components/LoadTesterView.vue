<script setup>
import { ref } from 'vue'
import { Play, Square } from 'lucide-vue-next'
import { useLoadTestApi } from '../composables/useLoadTestApi'
import { useSystemStats } from '../composables/useSystemStats'
import QpsChart from './charts/QpsChart.vue'

const { status, startTest, stopTest } = useLoadTestApi()
const { qpsHistory } = useSystemStats()

const url = ref('https://example.com/api')
const concurrency = ref(10)
const duration = ref(30)

async function handleStart() {
  await startTest({ url: url.value, concurrency: concurrency.value, duration: duration.value })
}

async function handleStop() {
  await stopTest()
}
</script>

<template>
  <div class="space-y-6">
    <h1 class="text-xl font-bold">压测配置</h1>

    <div class="bg-dark-surface rounded-xl p-6 border border-dark-border space-y-5">
      <!-- URL -->
      <div>
        <label class="block text-sm text-dark-muted mb-1.5">目标 URL</label>
        <input
          v-model="url"
          type="text"
          placeholder="https://example.com/api"
          class="w-full bg-dark-bg border border-dark-border rounded-lg px-4 py-2.5 text-sm focus:outline-none focus:border-primary transition-colors"
        />
      </div>

      <!-- Concurrency -->
      <div>
        <label class="block text-sm text-dark-muted mb-1.5">
          并发数: <span class="text-primary font-bold">{{ concurrency }}</span>
        </label>
        <input
          v-model.number="concurrency"
          type="range"
          min="1"
          max="1000"
          class="w-full accent-primary"
        />
        <div class="flex justify-between text-xs text-dark-muted mt-1">
          <span>1</span>
          <span>1000</span>
        </div>
      </div>

      <!-- Duration -->
      <div>
        <label class="block text-sm text-dark-muted mb-1.5">执行时长 (秒)</label>
        <input
          v-model.number="duration"
          type="number"
          min="1"
          class="w-full bg-dark-bg border border-dark-border rounded-lg px-4 py-2.5 text-sm focus:outline-none focus:border-primary transition-colors"
        />
      </div>

      <!-- Actions -->
      <div class="flex gap-3 pt-2">
        <button
          @click="handleStart"
          :disabled="status === 'running'"
          class="flex items-center gap-2 px-5 py-2.5 bg-primary hover:bg-primary-hover disabled:opacity-40 disabled:cursor-not-allowed rounded-lg text-sm font-medium transition-colors"
        >
          <Play :size="16" />
          启动
        </button>
        <button
          @click="handleStop"
          :disabled="status !== 'running'"
          class="flex items-center gap-2 px-5 py-2.5 bg-red-600 hover:bg-red-700 disabled:opacity-40 disabled:cursor-not-allowed rounded-lg text-sm font-medium transition-colors"
        >
          <Square :size="16" />
          停止
        </button>
        <span class="flex items-center text-sm ml-2" :class="status === 'running' ? 'text-green-400' : 'text-dark-muted'">
          {{ status === 'running' ? '运行中...' : '就绪' }}
        </span>
      </div>
    </div>

    <!-- Live QPS -->
    <div class="bg-dark-surface rounded-xl p-5 border border-dark-border">
      <h2 class="text-sm text-dark-muted mb-4">实时 QPS</h2>
      <QpsChart :data="qpsHistory" />
    </div>
  </div>
</template>
