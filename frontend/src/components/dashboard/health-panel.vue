<script setup lang="ts">
import type { DashboardHealth } from '@/lib/dashboard'

import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

defineProps<{ health: DashboardHealth }>()

const SERVICES = [
  { key: 'db', label: 'Database' },
  { key: 'redis', label: 'Redis' },
  { key: 'worker', label: 'Worker' },
  { key: 'provider', label: 'Provider' },
] as const

const num = (v: number | null) => (v === null ? '–' : String(v))
</script>

<template>
  <Card>
    <CardHeader>
      <CardTitle>System health</CardTitle>
    </CardHeader>
    <CardContent class="space-y-4">
      <div class="flex flex-wrap gap-2">
        <Badge
          v-for="s in SERVICES"
          :key="s.key"
          :data-service="s.key"
          :data-state="health.services[s.key]"
          :variant="health.services[s.key] === 'down' ? 'destructive' : 'default'"
          :class="health.services[s.key] === 'ok' ? 'bg-success/15 text-success border-transparent' : ''"
        >
          {{ s.label }} {{ health.services[s.key] }}
        </Badge>
      </div>
      <div class="space-y-1 text-sm">
        <div v-for="s in health.streams" :key="s.channel" :data-stream="s.channel" class="flex flex-wrap gap-x-4">
          <span class="w-14 font-medium uppercase">{{ s.channel }}</span>
          <span>Length: <span class="tabular-nums">{{ num(s.length) }}</span></span>
          <span>Pending: <span class="tabular-nums">{{ num(s.pending) }}</span></span>
          <span>In progress: <span class="tabular-nums">{{ s.in_progress }}</span></span>
        </div>
      </div>
    </CardContent>
  </Card>
</template>
