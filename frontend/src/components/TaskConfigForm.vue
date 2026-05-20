<script setup>
import { ref, watch } from 'vue'
import { Rocket, Zap, Code, Terminal, Globe, Cpu, Monitor } from 'lucide-vue-next'

const props = defineProps({
  fields: { type: Array, default: () => [] },
  initialValues: { type: Object, default: () => ({}) },
})

const form = ref({})
const iconComponents = { Rocket, Zap, Code, Terminal, Globe, Cpu, Monitor }

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
        <!-- Icon Picker -->
        <div v-if="field.input_type === 'icon_picker'" class="grid grid-cols-7 gap-2">
          <button
            v-for="opt in field.options"
            :key="opt.value"
            @click="form[field.field] = opt.value"
            type="button"
            class="aspect-square flex items-center justify-center rounded-xl border transition-all"
            :class="form[field.field] === opt.value ? 'bg-blue-50 border-blue-400 text-blue-600 shadow-sm' : 'bg-gray-50 border-gray-100 text-gray-400 hover:border-gray-200'"
          >
            <component :is="iconComponents[opt.value]" :size="18" />
          </button>
        </div>

        <!-- Color Picker -->
        <div v-else-if="field.input_type === 'color_picker'" class="grid grid-cols-6 gap-2">
          <button
            v-for="opt in field.options"
            :key="opt.value"
            @click="form[field.field] = opt.value"
            type="button"
            class="aspect-square flex items-center justify-center rounded-xl border transition-all"
            :class="form[field.field] === opt.value ? 'ring-2 ring-blue-400 ring-offset-2 scale-90' : 'border-gray-100 hover:border-gray-200'"
          >
            <div class="w-full h-full rounded-lg" :class="opt.value.split(' ')[1]"></div>
          </button>
        </div>

        <select
          v-else-if="field.input_type === 'select'"
          v-model="form[field.field]"
          class="w-full px-4 py-2.5 bg-gray-50 border border-gray-200 rounded-xl text-sm text-gray-800 focus:outline-none focus:border-blue-400 transition-all appearance-none bg-[length:14px_14px] bg-[right_12px_center] bg-no-repeat pr-9"
          :style="{ backgroundImage: 'url(\'data:image/svg+xml;utf8,<svg xmlns=%22http://www.w3.org/2000/svg%22 width=%2214%22 height=%2214%22 viewBox=%220 0 24 24%22 fill=%22none%22 stroke=%22%2394A3B8%22 stroke-width=%222%22 stroke-linecap=%22round%22 stroke-linejoin=%22round%22><polyline points=%226 9 12 15 18 9%22/></svg>\')' }"
        >
          <option v-if="!form[field.field]" value="" disabled>{{ field.placeholder || '请选择' }}</option>
          <option v-for="opt in (field.options || [])" :key="opt.value" :value="opt.value">{{ opt.label || opt.value }}</option>
        </select>
        <textarea
          v-else-if="field.input_type === 'textarea'"
          v-model="form[field.field]"
          :placeholder="field.placeholder"
          rows="5"
          class="w-full px-4 py-2.5 bg-gray-50 border border-gray-200 rounded-xl text-sm text-gray-800 placeholder:text-gray-300 focus:outline-none focus:border-blue-400 transition-all resize-none custom-scrollbar shadow-inner"
        ></textarea>
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

<style scoped>
.custom-scrollbar::-webkit-scrollbar { width: 3px; }
.custom-scrollbar::-webkit-scrollbar-track { background: transparent; }
.custom-scrollbar::-webkit-scrollbar-thumb { background: rgba(0, 0, 0, 0.05); border-radius: 10px; }
</style>
