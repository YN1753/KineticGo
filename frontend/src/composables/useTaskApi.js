import { ref } from 'vue'

export function useTaskApi() {
  const taskList = ref([])
  const scheduleList = ref([])
  const runningIds = ref(new Set())
  const loading = ref(false)

  async function getApp() {
    return await import('../../wailsjs/go/wailsapp/App')
  }

  async function fetchTaskList() {
    loading.value = true
    try {
      const app = await getApp()
      taskList.value = (await app.GetTaskList()) ?? []
    } catch {
      console.warn('[mock] GetTaskList')
      taskList.value = [
        { ID: 1, Name: '校园网自动连', Type: 'campus_auth', Description: '检测网络状态并在掉线时自动执行登录认证' },
        { ID: 2, Name: '性能压测', Type: 'load_test', Description: '对指定目标进行高并发HTTP压力测试' },
        { ID: 3, Name: '延迟雷达', Type: 'net_radar', Description: '实时监控网络延迟和丢包率' },
        { ID: 4, Name: '端口杀手', Type: 'port_killer', Description: '扫描并一键关闭占用特定端口的系统进程' },
      ]
    } finally {
      loading.value = false
    }
  }

  async function fetchTaskConfig(taskId) {
    try {
      const app = await getApp()
      return await app.GetTaskConfigById(taskId)
    } catch {
      console.warn('[mock] GetTaskConfigById', taskId)
      return []
    }
  }

  async function fetchScheduleList() {
    loading.value = true
    try {
      const app = await getApp()
      scheduleList.value = (await app.GetTaskScheduleList()) ?? []
      return scheduleList.value
    } catch {
      console.warn('[mock] GetTaskScheduleList')
      scheduleList.value = []
      return []
    } finally {
      loading.value = false
    }
  }

  async function createSchedule(schedule) {
    const app = await getApp()
    await app.CreateTaskSchedule(schedule)
  }

  async function updateSchedule(schedule) {
    const app = await getApp()
    await app.UpdateTaskSchedule(schedule)
  }

  async function deleteSchedule(id) {
    const app = await getApp()
    await app.DeleteTaskSchedule(id)
  }

  async function fetchScheduleById(id) {
    const app = await getApp()
    return await app.GetTaskScheduleById(id)
  }

  async function runTask(scheduleId) {
    try {
      const app = await getApp()
      await app.RunTask(scheduleId)
    } catch (e) {
      console.warn('[mock] RunTask', scheduleId, e)
    }
  }

  async function stopTask(scheduleId) {
    try {
      const app = await getApp()
      await app.StopTask(scheduleId)
    } catch (e) {
      console.warn('[mock] StopTask', scheduleId, e)
    }
  }

  async function fetchRunningIds() {
    try {
      const app = await getApp()
      if (typeof app.GetRunningTaskIds !== 'function') {
        runningIds.value = new Set()
        return
      }
      const ids = await app.GetRunningTaskIds()
      runningIds.value = new Set(ids ?? [])
    } catch (e) {
      console.warn('[mock] GetRunningTaskIds', e)
      runningIds.value = new Set()
    }
  }

  async function fetchExecutions(limit = 50) {
    const app = await getApp()
    return await app.GetTaskExecutions(limit)
  }

  async function fetchLogsByExecution(execId) {
    const app = await getApp()
    return await app.GetTaskLogsByExecution(execId)
  }

  async function getVersion() {
    try {
      const app = await getApp()
      return await app.GetVersion()
    } catch {
      return 'dev'
    }
  }

  async function checkUpdate() {
    try {
      const app = await getApp()
      return await app.CheckUpdate()
    } catch (e) {
      console.warn('CheckUpdate failed', e)
      return null
    }
  }

  async function applyUpdate(url) {
    const app = await getApp()
    return await app.ApplyUpdate(url)
  }

  async function fetchSystemTaskScheduleList() {
    try {
      const app = await getApp()
      return await app.GetSystemTaskScheduleList()
    } catch {
      console.warn('[mock] GetSystemTaskScheduleList')
      return []
    }
  }

  async function enableSystemTask(scheduleID) {
    try {
      const app = await getApp()
      await app.EnableSystemTask(scheduleID)
    } catch (e) {
      console.warn('[mock] EnableSystemTask', scheduleID, e)
    }
  }

  async function disableSystemTask(scheduleID) {
    try {
      const app = await getApp()
      await app.DisableSystemTask(scheduleID)
    } catch (e) {
      console.warn('[mock] DisableSystemTask', scheduleID, e)
    }
  }

  return {
    taskList, scheduleList, runningIds, loading,
    fetchTaskList, fetchTaskConfig, fetchScheduleList, fetchScheduleById,
    createSchedule, updateSchedule, deleteSchedule,
    runTask, stopTask, fetchRunningIds,
    fetchExecutions, fetchLogsByExecution,
    getVersion, checkUpdate, applyUpdate,
    fetchSystemTaskScheduleList, enableSystemTask, disableSystemTask,
  }
}
