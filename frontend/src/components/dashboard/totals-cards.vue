<script setup lang="ts">
import type { ChannelTotals } from '@/lib/dashboard'

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { successRate } from '@/lib/dashboard'
import { formatCount } from '@/lib/format'

defineProps<{ totals: { sms: ChannelTotals, email: ChannelTotals } }>()

const CHANNELS = [
  { key: 'sms', label: 'SMS' },
  { key: 'email', label: 'Email' },
] as const

const CARDS = [
  { key: 'sent', label: 'Sent', cls: 'text-success' },
  { key: 'failed', label: 'Failed', cls: 'text-destructive' },
  { key: 'unknown', label: 'Unknown', cls: '', hint: 'may be delivered' },
  { key: 'skipped', label: 'Skipped', cls: 'text-muted-foreground' },
  { key: 'in_progress', label: 'In progress', cls: '' },
] as const
</script>

<template>
  <div class="space-y-4">
    <div v-for="ch in CHANNELS" :key="ch.key" :data-row="ch.key" class="space-y-2">
      <h2 class="text-sm font-medium">
        {{ ch.label }}
      </h2>
      <div class="grid grid-cols-2 gap-3 md:grid-cols-3 xl:grid-cols-6">
        <Card v-for="c in CARDS" :key="c.key" :data-card="c.key" class="gap-1 py-4">
          <CardHeader class="px-4">
            <CardTitle class="text-muted-foreground text-xs font-medium">
              {{ c.label }}
            </CardTitle>
          </CardHeader>
          <CardContent class="px-4">
            <div class="text-2xl font-semibold tabular-nums" :class="c.cls">
              {{ formatCount(totals[ch.key][c.key]) }}
            </div>
            <p v-if="'hint' in c" class="text-muted-foreground text-xs">
              {{ c.hint }}
            </p>
          </CardContent>
        </Card>
        <Card data-card="rate" class="gap-1 py-4">
          <CardHeader class="px-4">
            <CardTitle class="text-muted-foreground text-xs font-medium">
              Success rate
            </CardTitle>
          </CardHeader>
          <CardContent class="px-4">
            <div class="text-2xl font-semibold tabular-nums">
              {{ successRate(totals[ch.key]) }}
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  </div>
</template>
