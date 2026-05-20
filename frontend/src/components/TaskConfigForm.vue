<script setup>
import { ref, watch } from 'vue'

const props = defineProps({
  fields: { type: Array, default: () => [] },
  initialValues: { type: Object, default: () => ({}) },
})

const form = ref({})

watch(() => props.fields, (fields) => {
  const next = {}
  for (const f of fields) {
    next[f.field] = props.initialValues[f.field] ?? f.default_val ?? ''
  }
  form.value = next
}, { immediate: true })

watch(() => props.initialValues, (vals) => {
  for (const f of props.fields) {
    if (vals[f.field] !== undefined) {
      form.value[f.field] = vals[f.field]
    }
  }
}, { deep: true })

function getValues() {
  return { ...form.value }
}

defineExpose({ getValues })
</script>

<template>
  <div class="space-y-4">
    <div v-for="field in fields" :key="field.field" class="space-y-1.5">
      <label class="block text-[10px] font-bold text-gray-400 uppercase tracking-widest">{{ field.label }}</label>
      <div class="relative">
        <select
          v-if="field.input_type === 'select'"
          v-model="form[field.field]"
          class="w-full px-4 py-2.5 bg-gray-50 border border-gray-200 rounded-xl text-sm text-gray-800 focus:outline-none focus:border-blue-400 transition-all appearance-none bg-[length:14px_14px] bg-[right_12px_center] bg-no-repeat pr-9"
          :style="{ backgroundImage: 'url(\'data:image/svg+xml;utf8,<svg xmlns=%22http://www.w3.org/2000/svg%22 width=%2214%22 height=%2214%22 viewBox=%220 0 24 24%22 fill=%22none%22 stroke=%22%2394A3B8%22 stroke-width=%222%22 stroke-linecap=%22round%22 stroke-linejoin=%22round%22><polyline points=%226 9 12 15 18 9%22/></svg>\')' }"
        >
          <option v-if="!form[field.field]" value="" disabled>{{ field.placeholder || '请选择' }}</option>
          <option v-for="opt in (field.options || [])" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
        </select>
        <input
          v-else
          v-model="form[field.field]"
          :type="field.input_type === 'password' ? 'password' : field.input_type === 'number' ? 'number' : 'text'"
          :placeholder="field.placeholder"
          class="w-full px-4 py-2.5 bg-gray-50 border border-gray-200 rounded-xl text-sm text-gray-800 placeholder:text-gray-300 focus:outline-none focus:border-blue-400 transition-all"
        />
      </div>
    </div>
  </div>
</template>
