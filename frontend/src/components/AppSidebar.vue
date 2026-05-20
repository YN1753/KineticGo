<script setup>
import { ref } from 'vue'
import { LayoutDashboard, ScrollText, Settings, ChevronLeft, ChevronRight } from 'lucide-vue-next'

defineProps({
  currentPage: { type: String, required: true }
})

const emit = defineEmits(['navigate'])

const collapsed = ref(false)
function toggle() {
  collapsed.value = !collapsed.value
}

const navItems = [
  { key: 'dashboard', label: '仪表盘', icon: LayoutDashboard },
  { key: 'loghistory', label: '历史日志', icon: ScrollText },
  { key: 'settings', label: '系统设置', icon: Settings }
]
</script>

<template>
  <aside
    class="h-screen bg-white border-r border-gray-200 flex flex-col shrink-0 transition-[width] duration-300 ease-out shadow-sm"
    :class="collapsed ? 'w-16' : 'w-56'"
  >
    <!-- logo + toggle -->
    <div
      class="h-16 flex items-center gap-2 border-b border-gray-100"
      :class="collapsed ? 'justify-center px-2' : 'px-5'"
    >
      <div class="w-8 h-8 rounded-lg bg-gradient-to-br from-blue-500 to-cyan-500 flex items-center justify-center shrink-0 shadow-sm">
        <span class="text-white font-bold text-sm">K</span>
      </div>
      <span
        v-show="!collapsed"
        class="text-base font-bold text-gray-800 tracking-tight truncate flex-1 transition-opacity duration-200"
      >KineticGo</span>
      <button
        @click="toggle"
        class="w-7 h-7 flex items-center justify-center rounded-lg text-gray-400 hover:text-gray-600 hover:bg-gray-50 transition-colors shrink-0"
        :title="collapsed ? '展开' : '收起'"
      >
        <ChevronLeft v-if="!collapsed" :size="16" />
        <ChevronRight v-else :size="16" />
      </button>
    </div>

    <!-- nav -->
    <nav class="flex-1 py-4 space-y-1" :class="collapsed ? 'px-2' : 'px-3'">
      <button
        v-for="item in navItems"
        :key="item.key"
        @click="emit('navigate', item.key)"
        class="w-full flex items-center rounded-xl text-sm font-medium transition-all duration-200"
        :class="[
          currentPage === item.key
            ? 'bg-blue-50 text-blue-600 shadow-[inset_0_0_0_1px_rgba(59,130,246,0.1)]'
            : 'text-gray-500 hover:text-gray-800 hover:bg-gray-50',
          collapsed ? 'justify-center py-2.5' : 'gap-3 px-3.5 py-2.5'
        ]"
        :title="item.label"
      >
        <component :is="item.icon" :size="18" :stroke-width="currentPage === item.key ? 2.2 : 1.8" />
        <span
          v-show="!collapsed"
          class="whitespace-nowrap transition-opacity duration-200"
        >{{ item.label }}</span>
      </button>
    </nav>

    <!-- footer: version -->
    <div class="px-4 py-4 border-t border-gray-100" :class="collapsed ? 'px-2 text-center' : ''">
      <div
        v-show="!collapsed"
        class="text-[11px] text-gray-400 whitespace-nowrap transition-opacity duration-200"
      >v0.1.0</div>
    </div>
  </aside>
</template>
