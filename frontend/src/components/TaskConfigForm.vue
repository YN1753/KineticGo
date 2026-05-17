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
      <label class="block text-xs font-medium text-dark-muted uppercase tracking-wider">{{ field.label }}</label>
      <input
        v-model="form[field.field]"
        :type="field.input_type === 'password' ? 'password' : field.input_type === 'number' ? 'number' : 'text'"
        :placeholder="field.placeholder"
        class="w-full px-3.5 py-2.5 bg-black/[0.03] border border-dark-border rounded-xl text-sm text-dark-text placeholder-dark-muted/60 focus:outline-none focus:border-accent-blue/50 focus:ring-1 focus:ring-accent-blue/20 transition-all"
      />
    </div>
  </div>
</template>
