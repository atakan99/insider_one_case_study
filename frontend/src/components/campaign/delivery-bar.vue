<script setup lang="ts">
import { computed } from 'vue'

import type { ChannelProgress } from '@/lib/delivery'

const props = defineProps<{ progress: ChannelProgress }>()

const PARTS = [
  { key: 'sent', label: 'sent', cls: 'bg-success' },
  { key: 'failed', label: 'failed', cls: 'bg-destructive' },
  { key: 'unknown', label: 'unknown', cls: 'bg-warning' },
  { key: 'skipped', label: 'skipped', cls: 'bg-muted-foreground/40' },
] as const

// The part of the track with no segment is the messages that are not finished.
const segments = computed(() => {
  const total = props.progress.total || 0
  return PARTS.map(p => ({ ...p, n: props.progress[p.key] || 0 }))
    .filter(s => total > 0 && s.n > 0)
    .map(s => ({ ...s, width: `${s.n / total * 100}%` }))
})
const label = computed(() =>
  `${PARTS.map(p => `${props.progress[p.key] || 0} ${p.label}`).join(', ')} of ${props.progress.total || 0}`)
</script>

<template>
  <div data-test="delivery-bar" role="img" :aria-label="label" class="bg-muted flex h-2 w-full overflow-hidden rounded-full">
    <div v-for="s in segments" :key="s.key" :data-segment="s.key" :class="s.cls" :style="{ width: s.width }" />
  </div>
</template>
