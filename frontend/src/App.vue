<script setup>
import { ref, computed } from 'vue'
import AppSidebar from './components/AppSidebar.vue'
import DashboardView from './components/DashboardView.vue'
import LogHistoryView from './components/LogHistoryView.vue'
import SettingsView from './components/SettingsView.vue'

const currentPage = ref('dashboard')

const views = {
  dashboard: DashboardView,
  loghistory: LogHistoryView,
  settings: SettingsView
}

const currentView = computed(() => views[currentPage.value])
</script>

<template>
  <div class="flex h-screen bg-dark-bg overflow-hidden">
    <AppSidebar :current-page="currentPage" @navigate="currentPage = $event" />
    <main class="flex-1 overflow-y-auto">
      <component :is="currentView" />
    </main>
  </div>
</template>
