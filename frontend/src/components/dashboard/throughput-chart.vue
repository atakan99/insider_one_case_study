<script lang="ts">
import { VisAxis, VisLine, VisXYContainer } from '@unovis/vue'
import { computed } from 'vue'

import type { ChartConfig } from '@/components/ui/chart'
import type { DashboardStats } from '@/lib/dashboard'

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { ChartContainer, ChartCrosshair, ChartTooltipContent, componentToString } from '@/components/ui/chart'

export const SEND_LIMIT = 100

export interface ThroughputPoint { x: number, sms: number, email: number, limit: number }

export function buildPoints(t: DashboardStats['throughput']): ThroughputPoint[] {
  const start = Date.parse(t.start)
  const step = t.bucket_seconds * 1000
  return t.sms.map((sms, i) => ({ x: start + i * step, sms, email: t.email[i] ?? 0, limit: SEND_LIMIT }))
}
</script>

<script setup lang="ts">
const props = defineProps<{ throughput: DashboardStats['throughput'] }>()

const config: ChartConfig = {
  sms: { label: 'SMS', color: 'var(--chart-1)' },
  email: { label: 'Email', color: 'var(--chart-2)' },
  limit: { label: 'Limit', color: 'var(--muted-foreground)' },
}

const data = computed(() => buildPoints(props.throughput))

const pad = (n: number) => String(n).padStart(2, '0')
function formatTime(x: number | Date): string {
  const d = new Date(x)
  return `${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

const x = (d: ThroughputPoint) => d.x
const colors = [config.sms.color, config.email.color, config.limit.color]
const tooltip = componentToString(config, ChartTooltipContent, { labelFormatter: formatTime })
</script>

<template>
  <Card>
    <CardHeader>
      <CardTitle>Throughput</CardTitle>
      <CardDescription>Sends per second, last 10 minutes. Dashed line is the {{ SEND_LIMIT }} per second limit.</CardDescription>
    </CardHeader>
    <CardContent>
      <ChartContainer :config="config" class="aspect-auto h-64 w-full" cursor>
        <VisXYContainer :data="data" :y-domain="[0, undefined]">
          <VisLine :x="x" :y="(d: ThroughputPoint) => d.sms" :color="colors[0]" />
          <VisLine :x="x" :y="(d: ThroughputPoint) => d.email" :color="colors[1]" />
          <VisLine :x="x" :y="(d: ThroughputPoint) => d.limit" :color="colors[2]" :line-dash-array="[4, 4]" />
          <VisAxis type="x" :tick-format="formatTime" :num-ticks="6" :grid-line="false" />
          <VisAxis type="y" :num-ticks="4" />
          <ChartCrosshair :template="tooltip" :color="(_d: ThroughputPoint, i: number) => colors[i]" />
        </VisXYContainer>
      </ChartContainer>
    </CardContent>
  </Card>
</template>
