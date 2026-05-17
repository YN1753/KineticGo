<script setup>
import { ref, watch, onMounted } from 'vue'
import { Line } from 'vue-chartjs'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Filler
} from 'chart.js'

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Filler)

const props = defineProps({
  data: { type: Array, default: () => [] }
})

const chartData = ref({
  labels: [],
  datasets: [{
    label: 'QPS',
    data: [],
    borderColor: '#3B82F6',
    backgroundColor: 'rgba(59, 130, 246, 0.1)',
    fill: true,
    tension: 0.3,
    pointRadius: 0,
    borderWidth: 2
  }]
})

const chartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  animation: false,
  scales: {
    x: {
      display: false
    },
    y: {
      beginAtZero: true,
      grid: {
        color: 'rgba(15, 23, 42, 0.06)'
      },
      ticks: {
        color: '#64748B',
        font: { size: 11 }
      }
    }
  },
  plugins: {
    legend: { display: false }
  }
}

watch(() => props.data, (newData) => {
  const last60 = newData.slice(-60)
  chartData.value = {
    labels: last60.map((_, i) => i),
    datasets: [{
      ...chartData.value.datasets[0],
      data: last60.map(d => d.value ?? d)
    }]
  }
}, { deep: true })
</script>

<template>
  <div class="h-48">
    <Line :data="chartData" :options="chartOptions" />
  </div>
</template>
