<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'

import type { ChannelProgress, DeliveryChannel, Progress, Step } from '@/lib/delivery'

import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { currentStep, watchProgress } from '@/lib/delivery'

const props = defineProps<{ campaignId: number, open: boolean }>()
// The kebab-case names are the component contract of the plan.

const emit = defineEmits<{ 'update:open': [open: boolean], 'open-list': [] }>()

const STEPS: { key: Step, label: string }[] = [
  { key: 'approved', label: 'Approved' },
  { key: 'queued', label: 'Queued' },
  { key: 'sending', label: 'Sending' },
  { key: 'done', label: 'Done' },
]
const CHANNELS: { key: DeliveryChannel, label: string }[] = [
  { key: 'sms', label: 'SMS' },
  { key: 'email', label: 'Email' },
]
const SEGMENTS = [
  { key: 'sent', label: 'Sent', color: 'bg-success' },
  { key: 'retrying', label: 'Retrying', color: 'bg-info' },
  { key: 'failed', label: 'Failed', color: 'bg-destructive' },
  { key: 'unknown', label: 'Unknown', color: 'bg-warning' },
] as const

const progress = ref<Progress | null>(null)
let stop: (() => void) | null = null

function stopWatching() {
  stop?.()
  stop = null
}

watch(() => [props.open, props.campaignId] as const, ([open, id]) => {
  stopWatching()
  if (!open)
    return
  progress.value = null
  stop = watchProgress(id, (p) => {
    progress.value = p
  })
}, { immediate: true })

onBeforeUnmount(stopWatching)

const fmt = (n: number) => n.toLocaleString('en-US')

const step = computed(() => progress.value ? currentStep(progress.value.status) : 'approved')
const isDone = computed(() => progress.value?.status === 'completed')

function stepState(i: number): 'done' | 'current' | 'todo' {
  const now = STEPS.findIndex(s => s.key === step.value)
  if (isDone.value || i < now)
    return 'done'
  return i === now ? 'current' : 'todo'
}

const rows = computed(() => CHANNELS
  .map(c => ({ ...c, p: progress.value?.channels[c.key] as ChannelProgress | undefined }))
  .filter(r => r.p && r.p.total > 0) as { key: DeliveryChannel, label: string, p: ChannelProgress }[])

const totals = computed(() => {
  const t = { sent: 0, failed: 0, unknown: 0, skipped: 0 }
  for (const r of rows.value) {
    t.sent += r.p.sent
    t.failed += r.p.failed
    t.unknown += r.p.unknown
    t.skipped += r.p.skipped
  }
  return t
})

function width(p: ChannelProgress, key: typeof SEGMENTS[number]['key']): string {
  return `${p[key] / p.total * 100}%`
}

function close() {
  emit('update:open', false)
}

function openList() {
  // eslint-disable-next-line vue/custom-event-name-casing
  emit('open-list')
  close()
}
</script>

<template>
  <Dialog :open="open" @update:open="v => emit('update:open', v)">
    <DialogContent
      data-test="sending"
      class="flex h-dvh w-screen max-w-none flex-col gap-4 overflow-y-auto rounded-none sm:max-w-none md:h-auto md:max-h-[calc(100dvh-2rem)] md:w-[min(480px,calc(100%-2rem))] md:max-w-[480px] md:rounded-lg"
    >
      <DialogHeader class="text-left">
        <DialogTitle>Sending campaign #{{ campaignId }}</DialogTitle>
        <DialogDescription>You can close this. Sending continues in the background.</DialogDescription>
      </DialogHeader>

      <ol class="flex items-start gap-1">
        <template v-for="(s, i) in STEPS" :key="s.key">
          <li
            :data-step="s.key"
            :data-state="stepState(i)"
            class="flex min-w-0 flex-col items-center gap-1.5 text-xs sm:flex-row"
            :class="stepState(i) === 'todo' ? 'text-muted-foreground' : 'text-foreground'"
          >
            <span
              class="flex size-5 shrink-0 items-center justify-center rounded-full border-2 text-[11px]"
              :class="{
                'border-success bg-success text-white': stepState(i) === 'done',
                'border-primary text-primary': stepState(i) === 'current',
                'border-border': stepState(i) === 'todo',
              }"
            >{{ stepState(i) === 'done' ? '✓' : i + 1 }}</span>
            <span :class="{ 'font-medium': stepState(i) === 'current' }">{{ s.label }}</span>
          </li>
          <li v-if="i < STEPS.length - 1" aria-hidden="true" class="mt-2.5 h-0.5 min-w-3 flex-1 sm:mt-2.5" :class="stepState(i) === 'done' ? 'bg-success' : 'bg-border'" />
        </template>
      </ol>

      <div v-for="r in rows" :key="r.key" :data-channel="r.key" class="space-y-1.5 text-sm">
        <div class="flex justify-between">
          <strong>{{ r.label }}</strong>
          <span class="tabular-nums">{{ fmt(r.p.sent) }} / {{ fmt(r.p.total) }}</span>
        </div>
        <div class="bg-muted flex h-2 overflow-hidden rounded-full">
          <i
            v-for="seg in SEGMENTS"
            :key="seg.key"
            :data-segment="seg.key"
            class="block h-full"
            :class="seg.color"
            :style="{ width: width(r.p, seg.key) }"
          />
        </div>
      </div>

      <div v-if="rows.length > 0" data-test="legend" class="text-muted-foreground flex flex-wrap gap-x-3 gap-y-1 text-xs">
        <span v-for="seg in SEGMENTS" :key="seg.key" class="inline-flex items-center gap-1">
          <i class="inline-block size-2 rounded-sm" :class="seg.color" />{{ seg.label }}
        </span>
        <span v-if="totals.skipped > 0">Skipped {{ fmt(totals.skipped) }}</span>
      </div>

      <p v-if="isDone" data-test="done-text" class="text-sm font-medium">
        Done. {{ fmt(totals.sent) }} sent, {{ fmt(totals.failed) }} failed, {{ fmt(totals.unknown) }} unknown.
      </p>

      <div class="mt-auto flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
        <Button data-test="close" variant="outline" @click="close">
          Close
        </Button>
        <Button data-test="open-list" @click="openList">
          Open delivery list
        </Button>
      </div>
    </DialogContent>
  </Dialog>
</template>
