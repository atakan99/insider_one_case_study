<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'

import type { DashboardHealth, DashboardStats } from '@/lib/dashboard'

import CampaignsTable from '@/components/dashboard/campaigns-table.vue'
import FailuresCard from '@/components/dashboard/failures-card.vue'
import HealthPanel from '@/components/dashboard/health-panel.vue'
import ThroughputChart from '@/components/dashboard/throughput-chart.vue'
import TotalsCards from '@/components/dashboard/totals-cards.vue'
import { DASHBOARD_POLL_MS, getDashboardHealth, getDashboardStats } from '@/lib/dashboard'

const stats = ref<DashboardStats | null>(null)
const health = ref<DashboardHealth | null>(null)
const failed = ref(false)
const lastUpdate = ref('')

let inFlight = false
let unmounted = false
let timer: ReturnType<typeof setTimeout> | undefined

const pad = (n: number) => String(n).padStart(2, '0')
function clock(d: Date): string {
  return `${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

function schedule() {
  clearTimeout(timer)
  if (!unmounted)
    timer = setTimeout(refresh, DASHBOARD_POLL_MS)
}

async function refresh() {
  if (inFlight || unmounted || document.visibilityState === 'hidden')
    return
  inFlight = true
  try {
    // Each call sets its own part, so a slow call does not delay the other.
    const statsCall = getDashboardStats().then((v) => {
      if (!unmounted)
        stats.value = v
    })
    const healthCall = getDashboardHealth().then((v) => {
      if (!unmounted)
        health.value = v
    })
    const [s, h] = await Promise.allSettled([statsCall, healthCall])
    if (unmounted)
      return
    failed.value = s.status === 'rejected' || h.status === 'rejected'
    if (!failed.value)
      lastUpdate.value = clock(new Date())
  }
  finally {
    inFlight = false
    schedule()
  }
}

function onVisibility() {
  if (document.visibilityState === 'visible')
    refresh()
}

onMounted(() => {
  document.addEventListener('visibilitychange', onVisibility)
  refresh()
})

onUnmounted(() => {
  unmounted = true
  clearTimeout(timer)
  document.removeEventListener('visibilitychange', onVisibility)
})
</script>

<template>
  <div class="space-y-6">
    <div class="flex flex-wrap items-baseline justify-between gap-2">
      <h1 class="text-2xl font-bold tracking-tight">
        Dashboard
      </h1>
      <p v-if="failed" class="text-muted-foreground text-sm">
        Could not refresh. Last update {{ lastUpdate || '–' }}
      </p>
    </div>

    <template v-if="stats">
      <TotalsCards :totals="stats.totals" />
      <div class="grid grid-cols-1 gap-6 lg:grid-cols-2">
        <ThroughputChart class="min-w-0" :throughput="stats.throughput" />
        <FailuresCard class="min-w-0" :failures="stats.failures" :mean-attempts="stats.mean_attempts" />
      </div>
      <CampaignsTable :campaigns="stats.campaigns" />
    </template>
    <p v-else-if="!failed" class="text-muted-foreground text-sm">
      Loading…
    </p>
    <HealthPanel v-if="health" :health="health" />
  </div>
</template>
