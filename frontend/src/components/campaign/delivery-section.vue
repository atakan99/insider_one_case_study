<script setup lang="ts">
import { computed, ref, watch } from 'vue'

import type { ChannelProgress, DeliveryChannel, DeliveryStatus, Progress, SendJob } from '@/lib/delivery'

import {
  AlertDialog,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Spinner } from '@/components/ui/spinner'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { CampaignRequestError, languageName } from '@/lib/campaigns'
import { listSendJobs, retryCampaign } from '@/lib/delivery'
import { formatCount as fmt } from '@/lib/format'

import DeliveryBar from './delivery-bar.vue'

type Chip = 'all' | 'sent' | 'retrying' | 'failed' | 'unknown' | 'skipped'

const props = defineProps<{ campaignId: number, progress: Progress }>()

const emit = defineEmits<{ viewProgress: [], retried: [progress: Progress], refresh: [] }>()

const FALLBACK_MESSAGE = 'Something went wrong. Try again.'
const PAGE_SIZE = 50
const CHANNELS: { key: DeliveryChannel, label: string }[] = [
  { key: 'sms', label: 'SMS' },
  { key: 'email', label: 'Email' },
]
const CHIPS: { key: Chip, label: string }[] = [
  { key: 'all', label: 'All' },
  { key: 'sent', label: 'Sent' },
  { key: 'retrying', label: 'Retrying' },
  { key: 'failed', label: 'Failed' },
  { key: 'unknown', label: 'Unknown' },
  { key: 'skipped', label: 'Skipped' },
]
const STATUS_LABEL: Record<DeliveryStatus, string> = {
  queued: 'Queued',
  sending: 'Sending',
  retrying: 'Retrying',
  sent: 'Sent',
  failed: 'Failed',
  unknown: 'Unknown',
  skipped: 'Skipped',
}
const STATUS_CLASS: Record<DeliveryStatus, string> = {
  queued: 'bg-secondary text-secondary-foreground',
  sending: 'bg-info/15 text-info',
  retrying: 'bg-info/15 text-info',
  sent: 'bg-success/15 text-success',
  failed: 'bg-destructive/15 text-destructive',
  unknown: 'bg-warning/15 text-warning',
  skipped: 'bg-muted text-muted-foreground',
}

const status = ref<Chip>('all')
const channel = ref<DeliveryChannel | 'all'>('all')
const jobs = ref<SendJob[]>([])
const nextAfter = ref<number | null>(null)
const loading = ref(false)
const listError = ref('')
const confirmOpen = ref(false)
const retrying = ref(false)
const retryError = ref('')
let token = 0
// Pages loaded with "Load more", and whether one is in flight. A refresh from the progress stream must not discard them.
let extraPages = 0
let loadingMore = false

const channels = computed(() => CHANNELS.map(c => ({ ...c, p: props.progress.channels[c.key] as ChannelProgress })))
const summaryChannels = computed(() => channels.value.filter(c => c.p.total > 0))

function sum(key: keyof ChannelProgress): number {
  return channels.value.reduce((t, c) => t + c.p[key], 0)
}
const retryable = computed(() => sum('retryable_failed'))
const canRetry = computed(() => props.progress.status === 'completed' && retryable.value > 0)
const canView = computed(() => ['approved', 'queued', 'sending'].includes(props.progress.status))

function chipCount(k: Chip): number {
  return k === 'all' ? sum('total') : sum(k)
}

function errorMessage(err: unknown): string {
  return err instanceof CampaignRequestError ? err.message : FALLBACK_MESSAGE
}

async function load(more: boolean) {
  const mine = ++token
  loadingMore = more
  if (!more)
    extraPages = 0
  loading.value = true
  listError.value = ''
  try {
    const res = await listSendJobs(props.campaignId, {
      status: status.value === 'all' ? undefined : status.value,
      channel: channel.value === 'all' ? undefined : channel.value,
      after: more && nextAfter.value !== null ? nextAfter.value : undefined,
      limit: PAGE_SIZE,
    })
    if (mine !== token)
      return
    jobs.value = more ? [...jobs.value, ...res.send_jobs] : res.send_jobs
    if (more)
      extraPages++
    nextAfter.value = res.next_after
  }
  catch (err) {
    if (mine === token)
      listError.value = errorMessage(err)
  }
  finally {
    if (mine === token) {
      loading.value = false
      loadingMore = false
    }
  }
}

function pick(c: Chip) {
  status.value = c
  void load(false)
}

function pickChannel(c: DeliveryChannel | 'all') {
  channel.value = c
  void load(false)
}

// Reload the first page when the campaign or its status changes, also at
// completed. When only the counts change, reload only while no extra page is
// loaded or loading, so that "Load more" keeps its pages.
watch(
  () => [props.campaignId, props.progress.status, JSON.stringify(props.progress.channels)] as const,
  ([id, st], old) => {
    if (!old || id !== old[0] || st !== old[1] || (extraPages === 0 && !loadingMore))
      void load(false)
  },
  { immediate: true },
)

function viewProgress() {
  emit('viewProgress')
}

async function sendAgain() {
  if (retrying.value)
    return
  retrying.value = true
  retryError.value = ''
  try {
    const res = await retryCampaign(props.campaignId)
    confirmOpen.value = false
    emit('retried', res.progress)
  }
  catch (err) {
    confirmOpen.value = false
    retryError.value = errorMessage(err)
    // A 409 means the page state is stale. Ask the page for fresh progress.
    if (err instanceof CampaignRequestError && err.status === 409)
      emit('refresh')
  }
  finally {
    retrying.value = false
  }
}
</script>

<template>
  <section data-test="delivery" class="space-y-4">
    <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
      <h2 class="text-lg font-semibold tracking-tight">
        Delivery
      </h2>
      <Button v-if="canView" data-test="view-progress" variant="outline" class="w-full sm:w-auto" @click="viewProgress">
        View progress
      </Button>
      <Button v-if="canRetry" data-test="retry" class="w-full sm:w-auto" @click="confirmOpen = true">
        Retry {{ fmt(retryable) }} failed
      </Button>
    </div>

    <p v-if="retryError" data-test="retry-error" class="text-destructive text-sm">
      {{ retryError }}
    </p>

    <div class="grid gap-3 sm:grid-cols-2">
      <Card v-for="c in summaryChannels" :key="c.key" :data-summary="c.key" class="py-4">
        <CardHeader class="px-4">
          <CardTitle class="flex items-baseline justify-between text-base">
            {{ c.label }}
            <span class="text-muted-foreground text-sm font-normal tabular-nums">{{ fmt(c.p.total) }} total</span>
          </CardTitle>
        </CardHeader>
        <CardContent class="grid grid-cols-2 gap-x-4 gap-y-1 px-4 text-sm tabular-nums">
          <DeliveryBar :progress="c.p" class="col-span-2 mb-2" />
          <span class="text-muted-foreground">Sent</span><span class="text-right">{{ fmt(c.p.sent) }}</span>
          <span class="text-muted-foreground">Failed</span><span class="text-right">{{ fmt(c.p.failed) }}</span>
          <span class="text-muted-foreground">Unknown</span><span class="text-right">{{ fmt(c.p.unknown) }}</span>
          <span class="text-muted-foreground">Skipped</span><span class="text-right">{{ fmt(c.p.skipped) }}</span>
        </CardContent>
      </Card>
    </div>

    <div class="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
      <div class="flex flex-wrap gap-1.5">
        <button
          v-for="c in CHIPS"
          :key="c.key"
          type="button"
          :data-chip="c.key"
          :aria-pressed="status === c.key"
          class="rounded-full border px-3 py-1 text-xs tabular-nums transition-colors"
          :class="status === c.key ? 'bg-primary text-primary-foreground border-primary' : 'hover:bg-accent'"
          @click="pick(c.key)"
        >
          {{ c.label }} {{ fmt(chipCount(c.key)) }}
        </button>
      </div>
      <div class="flex gap-1.5">
        <button
          v-for="c in [{ key: 'all', label: 'All' }, ...CHANNELS] as { key: DeliveryChannel | 'all', label: string }[]"
          :key="c.key"
          type="button"
          :data-channel-filter="c.key"
          :aria-pressed="channel === c.key"
          class="rounded-md border px-3 py-1 text-xs transition-colors"
          :class="channel === c.key ? 'bg-accent text-accent-foreground' : 'hover:bg-accent/50'"
          @click="pickChannel(c.key)"
        >
          {{ c.label }}
        </button>
      </div>
    </div>

    <p v-if="listError" data-test="list-error" class="text-destructive text-sm">
      {{ listError }}
    </p>
    <p v-else-if="!loading && jobs.length === 0" data-test="empty" class="text-muted-foreground rounded-lg border border-dashed p-6 text-center text-sm">
      No messages match this filter.
    </p>

    <template v-if="jobs.length > 0">
      <div class="hidden rounded-lg border min-[720px]:block">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Recipient</TableHead>
              <TableHead>Channel</TableHead>
              <TableHead>Language</TableHead>
              <TableHead>Status</TableHead>
              <TableHead class="text-right">
                Attempts
              </TableHead>
              <TableHead>Last error</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="j in jobs" :key="j.id" data-test="job">
              <TableCell class="tabular-nums">
                User #{{ j.user_id }}
              </TableCell>
              <TableCell>{{ j.channel === 'sms' ? 'SMS' : 'Email' }}</TableCell>
              <TableCell>{{ languageName(j.language) }}</TableCell>
              <TableCell>
                <Badge variant="outline" class="border-transparent" :class="STATUS_CLASS[j.status]">
                  {{ STATUS_LABEL[j.status] }}
                </Badge>
                <span v-if="j.status === 'unknown'" class="text-muted-foreground ml-1.5 text-xs">may be delivered</span>
              </TableCell>
              <TableCell class="text-right tabular-nums">
                {{ j.attempts }}
              </TableCell>
              <TableCell class="text-muted-foreground max-w-56 truncate">
                {{ j.last_error }}
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </div>

      <div class="space-y-2 min-[720px]:hidden">
        <Card v-for="j in jobs" :key="`c${j.id}`" data-test="job-card" class="py-3">
          <CardContent class="space-y-1.5 px-4 text-sm">
            <div class="flex items-center justify-between gap-2">
              <span class="font-medium tabular-nums">User #{{ j.user_id }}</span>
              <Badge variant="outline" class="border-transparent" :class="STATUS_CLASS[j.status]">
                {{ STATUS_LABEL[j.status] }}
              </Badge>
            </div>
            <p class="text-muted-foreground text-xs">
              {{ j.channel === 'sms' ? 'SMS' : 'Email' }} · {{ languageName(j.language) }} · {{ j.attempts }} {{ j.attempts === 1 ? 'attempt' : 'attempts' }}
            </p>
            <p v-if="j.status === 'unknown'" class="text-muted-foreground text-xs">
              may be delivered
            </p>
            <p v-if="j.last_error" class="text-muted-foreground text-xs break-words">
              {{ j.last_error }}
            </p>
          </CardContent>
        </Card>
      </div>
    </template>

    <div v-if="nextAfter !== null" class="flex justify-center">
      <Button data-test="load-more" variant="outline" :disabled="loading" @click="load(true)">
        <Spinner v-if="loading" />
        Load more
      </Button>
    </div>

    <AlertDialog :open="confirmOpen" @update:open="v => { if (!retrying) confirmOpen = v }">
      <AlertDialogContent data-test="retry-confirm">
        <AlertDialogHeader>
          <AlertDialogTitle>Retry failed messages?</AlertDialogTitle>
          <AlertDialogDescription>
            Send {{ fmt(retryable) }} failed messages again? Messages with an unknown result are not sent again.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <Button data-test="retry-back" variant="outline" :disabled="retrying" @click="confirmOpen = false">
            Back
          </Button>
          <Button data-test="retry-send" :disabled="retrying" @click="sendAgain">
            <Spinner v-if="retrying" />
            Send again
          </Button>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  </section>
</template>
