<script setup>
import { ref, onMounted } from 'vue'
import { ScrollText, Plus, X, Play, Settings, Trash2, Zap, Wifi, Radar, Terminal } from 'lucide-vue-next'
import { useTaskApi } from '../composables/useTaskApi'
import TaskConfigForm from './TaskConfigForm.vue'

const {
  taskList, scheduleList, loading,
  fetchTaskList, fetchTaskConfig, fetchScheduleList,
  createSchedule, updateSchedule, deleteSchedule,
} = useTaskApi()

const configForm = ref(null)

// modal state
const showPicker = ref(false)
const showConfig = ref(false)
const showDeleteConfirm = ref(false)

const selectedTask = ref(null)
const configFields = ref([])
const configInitial = ref({})
const editingSchedule = ref(null)
const scheduleToDelete = ref(null)

const typeIcons = {
  campus_auth: Wifi,
  load_test: Zap,
  net_radar: Radar,
  port_killer: Terminal,
  system: ScrollText,
}

function getTypeIcon(type) {
  return typeIcons[type] || ScrollText
}

// --- flow: add task ---

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

// --- flow: edit config ---

async function editConfig(schedule) {
  editingSchedule.value = schedule
  selectedTask.value = { ID: 0, Name: schedule.Name, Type: schedule.TaskType }
  const raw = await fetchTaskConfig(schedule.ID)
  configFields.value = Array.isArray(raw) ? raw : []
  // parse existing biz config as initial values
  try {
    configInitial.value = typeof schedule.Config === 'string'
      ? JSON.parse(schedule.Config)
      : (schedule.Config || {})
  } catch {
    configInitial.value = {}
  }
  showConfig.value = true
}

// --- flow: submit config ---

async function submitConfig() {
  const values = configForm.value?.getValues() ?? {}
  if (editingSchedule.value) {
    const updated = {
      ...editingSchedule.value,
      Config: JSON.stringify(values),
    }
    await updateSchedule(updated)
  } else {
    const schedule = {
      Name: selectedTask.value.Name,
      TaskType: selectedTask.value.Type,
      Config: JSON.stringify(values),
      IsEnabled: true,
    }
    await createSchedule(schedule)
  }
  showConfig.value = false
  await fetchScheduleList()
}

// --- flow: delete ---

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

// --- flow: run (placeholder) ---

function runTask(schedule) {
  console.log('run task:', schedule)
  // TODO: implement run
}

onMounted(() => {
  fetchScheduleList()
})
</script>

<template>
  <div class="space-y-6">
    <!-- header -->
    <div class="flex items-center justify-between">
      <h1 class="text-xl font-bold">任务管理</h1>
      <button
        @click="openPicker"
        class="flex items-center gap-2 px-4 py-2 bg-primary hover:bg-primary/90 text-white rounded-lg text-sm font-medium transition-colors"
      >
        <Plus :size="16" />
        添加任务
      </button>
    </div>

    <!-- schedule list -->
    <div v-if="loading" class="py-16 text-center text-dark-muted">加载中...</div>

    <div v-else-if="scheduleList.length === 0" class="py-16 text-center text-dark-muted">
      <ScrollText :size="32" class="mx-auto mb-3 opacity-40" />
      <div>暂无任务，点击上方按钮添加</div>
    </div>

    <div v-else class="grid gap-4 sm:grid-cols-2">
      <div
        v-for="s in scheduleList"
        :key="s.ID"
        class="relative bg-dark-surface border border-dark-border rounded-xl p-4 group hover:border-dark-border/80 transition-colors"
      >
        <!-- delete button -->
        <button
          @click="confirmDelete(s)"
          class="absolute top-3 right-3 w-6 h-6 flex items-center justify-center rounded-full text-dark-muted hover:text-red-400 hover:bg-red-400/10 opacity-0 group-hover:opacity-100 transition-all"
        >
          <X :size="14" />
        </button>

        <div class="flex items-start gap-3 mb-4">
          <div class="flex-shrink-0 w-10 h-10 rounded-lg bg-dark-bg flex items-center justify-center text-dark-muted">
            <component :is="getTypeIcon(s.TaskType)" :size="20" />
          </div>
          <div class="min-w-0 flex-1">
            <div class="font-medium text-sm truncate pr-6">{{ s.Name }}</div>
            <div class="text-xs text-dark-muted mt-0.5">{{ s.TaskType }}</div>
          </div>
        </div>

        <div class="flex gap-2">
          <button
            @click="runTask(s)"
            class="flex-1 flex items-center justify-center gap-1.5 px-3 py-1.5 bg-primary/10 hover:bg-primary/20 text-primary text-xs font-medium rounded-lg transition-colors"
          >
            <Play :size="14" />
            运行
          </button>
          <button
            @click="editConfig(s)"
            class="flex-1 flex items-center justify-center gap-1.5 px-3 py-1.5 bg-dark-bg hover:bg-dark-border text-dark-muted hover:text-dark-text text-xs font-medium rounded-lg transition-colors"
          >
            <Settings :size="14" />
            修改配置
          </button>
        </div>
      </div>
    </div>

    <!-- Task Picker Modal -->
    <Teleport to="body">
      <div v-if="showPicker" class="fixed inset-0 z-50 flex items-center justify-center">
        <div class="absolute inset-0 bg-black/60 backdrop-blur-sm" @click="showPicker = false" />
        <div class="relative bg-dark-surface border border-dark-border rounded-2xl w-full max-w-lg mx-4 shadow-2xl overflow-hidden">
          <div class="flex items-center justify-between px-6 py-4 border-b border-dark-border">
            <h2 class="text-lg font-semibold">选择任务</h2>
            <button @click="showPicker = false" class="text-dark-muted hover:text-white transition-colors">
              <X :size="20" />
            </button>
          </div>
          <div class="p-4 max-h-96 overflow-y-auto">
            <div v-if="loading" class="py-12 text-center text-dark-muted">加载中...</div>
            <div v-else class="grid gap-3">
              <button
                v-for="task in taskList"
                :key="task.ID"
                @click="selectTask(task)"
                class="flex items-start gap-4 p-4 rounded-xl border border-dark-border hover:border-primary/50 hover:bg-primary/5 text-left transition-all group"
              >
                <div class="flex-shrink-0 w-10 h-10 rounded-lg bg-dark-bg flex items-center justify-center text-dark-muted group-hover:text-primary transition-colors">
                  <component :is="getTypeIcon(task.Type)" :size="20" />
                </div>
                <div class="flex-1 min-w-0">
                  <div class="font-medium text-sm">{{ task.Name }}</div>
                  <div class="text-xs text-dark-muted mt-1">{{ task.Description }}</div>
                </div>
                <div class="flex-shrink-0 self-center">
                  <span class="text-xs px-2 py-0.5 rounded-full bg-dark-bg text-dark-muted">{{ task.Type }}</span>
                </div>
              </button>
            </div>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Config Modal -->
    <Teleport to="body">
      <div v-if="showConfig" class="fixed inset-0 z-50 flex items-center justify-center">
        <div class="absolute inset-0 bg-black/60 backdrop-blur-sm" @click="showConfig = false" />
        <div class="relative bg-dark-surface border border-dark-border rounded-2xl w-full max-w-md mx-4 shadow-2xl overflow-hidden">
          <div class="flex items-center justify-between px-6 py-4 border-b border-dark-border">
            <h2 class="text-lg font-semibold">{{ editingSchedule ? '修改配置' : '配置任务' }}</h2>
            <button @click="showConfig = false" class="text-dark-muted hover:text-white transition-colors">
              <X :size="20" />
            </button>
          </div>
          <div class="p-6 space-y-5">
            <div class="text-sm text-dark-muted">
              任务: <span class="text-dark-text font-medium">{{ selectedTask?.Name }}</span>
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
              class="w-full py-2.5 bg-primary hover:bg-primary/90 text-white rounded-lg text-sm font-medium transition-colors"
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
        <div class="absolute inset-0 bg-black/60 backdrop-blur-sm" @click="showDeleteConfirm = false" />
        <div class="relative bg-dark-surface border border-dark-border rounded-2xl w-full max-w-sm mx-4 shadow-2xl overflow-hidden">
          <div class="p-6 space-y-4">
            <h2 class="text-lg font-semibold">确认删除</h2>
            <p class="text-sm text-dark-muted">
              确定要删除任务「<span class="text-dark-text">{{ scheduleToDelete?.Name }}</span>」吗？此操作不可撤销。
            </p>
            <div class="flex gap-3 pt-2">
              <button
                @click="showDeleteConfirm = false"
                class="flex-1 py-2 bg-dark-bg hover:bg-dark-border text-dark-text rounded-lg text-sm font-medium transition-colors"
              >
                取消
              </button>
              <button
                @click="doDelete"
                class="flex-1 py-2 bg-red-500/90 hover:bg-red-500 text-white rounded-lg text-sm font-medium transition-colors"
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
