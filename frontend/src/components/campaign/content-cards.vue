<script setup lang="ts">
import { onBeforeUnmount, reactive, ref, watch } from 'vue'
import { useRouter } from 'vue-router'

import type { Content, ContentEvent, ContentLimits, RetryReason } from '@/lib/content'

import { Alert, AlertDescription } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { languageName } from '@/lib/campaigns'
import { ContentRequestError, generateContent } from '@/lib/content'
import { formatCount } from '@/lib/format'

import ContentView from './content-view.vue'
import ReviewDialog from './review-dialog.vue'

interface LangCard {
  code: string
  email: number
  sms: number
  draft?: Content
  final?: { content: Content, limits: ContentLimits }
  // retry is set from a retry event until the next draft, final or error.
  retry?: RetryReason
  error?: string
  closed: boolean
}

const props = defineProps<{ brief: string, rule: unknown, runId: number }>()
const emit = defineEmits<{ 'update:running': [running: boolean] }>()

const CLOSED_MESSAGE = 'The connection closed early. Try again.'

const cards = ref<LangCard[]>([])
const running = ref(false)
const requestError = ref('')
// campaignId is the draft of the current run; the review button needs it.
// draftId is the draft that "Regenerate content" replaces.
const campaignId = ref<number | null>(null)
let draftId: number | null = null
const reviewOpen = ref(false)
const router = useRouter()
let controller: AbortController | null = null
let token = 0

function setRunning(v: boolean) {
  running.value = v
  emit('update:running', v)
}

function cardFor(code: string) {
  return cards.value.find(c => c.code === code)
}

function handle(e: ContentEvent) {
  if (e.event === 'audience') {
    cards.value = e.data.languages.map(l => reactive({
      code: l.code,
      email: l.email,
      sms: l.sms,
      closed: false,
    }))
    return
  }
  if (e.event === 'done')
    return
  if (e.event === 'campaign') {
    campaignId.value = e.data.id
    draftId = e.data.id
    return
  }
  if (e.event === 'error' && e.data.language === undefined) {
    if (e.data.code === 'campaign_conflict')
      draftId = null
    requestError.value = e.data.message
    return
  }
  const card = cardFor(e.data.language!)
  if (!card)
    return
  if (e.event === 'draft') {
    card.draft = e.data.content
    card.retry = undefined
  }
  else if (e.event === 'retry') {
    card.draft = undefined
    card.retry = e.data.reason ?? 'invalid'
  }
  else if (e.event === 'final') {
    card.final = { content: e.data.content, limits: e.data.limits }
    card.retry = undefined
    card.error = undefined
  }
  else if (e.event === 'error') {
    card.error = e.data.message
    card.retry = undefined
  }
}

// A regenerate keeps the draft id, so the backend replaces that draft.
// A new run (a new "Find audience") starts a new campaign.
async function start(regenerate = false) {
  if (!regenerate)
    draftId = null
  campaignId.value = null
  reviewOpen.value = false
  const replaceId = draftId ?? undefined
  controller?.abort()
  const ctl = new AbortController()
  controller = ctl
  const mine = ++token
  cards.value = []
  requestError.value = ''
  setRunning(true)
  try {
    await generateContent(props.brief, props.rule, (e) => {
      if (mine === token)
        handle(e)
    }, ctl.signal, replaceId)
  }
  catch (err) {
    if (mine !== token)
      return
    if (err instanceof ContentRequestError && err.code !== 'stream_closed') {
      requestError.value = err.message
    }
    else {
      for (const c of cards.value) {
        if (!c.final && !c.error)
          c.closed = true
      }
    }
  }
  finally {
    if (mine === token)
      setRunning(false)
  }
}

watch(() => props.runId, (id) => {
  if (id > 0)
    start()
}, { immediate: true })

onBeforeUnmount(() => {
  token++
  controller?.abort()
})
</script>

<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <p v-if="running" class="text-sm text-muted-foreground" role="status">
        Generating content…
      </p>
      <span v-else />
      <Button variant="outline" :disabled="running" @click="start(true)">
        Regenerate content
      </Button>
    </div>

    <Alert v-if="requestError" variant="destructive">
      <AlertDescription>{{ requestError }}</AlertDescription>
    </Alert>

    <Card v-for="card in cards" :key="card.code" :data-language="card.code">
      <CardHeader>
        <CardTitle class="flex flex-wrap items-baseline gap-2">
          {{ languageName(card.code) }}
          <span class="text-sm font-normal text-muted-foreground tabular-nums">{{ formatCount(card.sms) }} SMS · {{ formatCount(card.email) }} email</span>
          <Badge v-if="!card.final && !card.error && card.draft" variant="secondary">
            Draft
          </Badge>
        </CardTitle>
        <p v-if="card.retry" class="text-sm text-muted-foreground" role="status">
          {{ card.retry === 'restart' ? 'Starting again…' : 'Fixing content…' }}
        </p>
      </CardHeader>
      <CardContent class="space-y-4">
        <Alert v-if="card.error" variant="destructive">
          <AlertDescription>{{ card.error }}</AlertDescription>
        </Alert>
        <Alert v-else-if="card.closed" variant="destructive">
          <AlertDescription>{{ CLOSED_MESSAGE }}</AlertDescription>
        </Alert>
        <ContentView
          v-if="!card.error && (card.final || card.draft)"
          :content="(card.final?.content ?? card.draft)!"
          :sms-length="card.final?.limits.sms_runes"
          :subject-length="card.final?.limits.email_subject_runes"
          :class="card.final ? '' : 'text-muted-foreground italic'"
        />
        <div v-else-if="!card.error && !card.closed" data-test="card-skeleton" class="grid gap-6 lg:grid-cols-[minmax(0,22rem)_minmax(0,1fr)]">
          <Skeleton class="h-24" />
          <Skeleton class="h-40" />
        </div>
      </CardContent>
    </Card>

    <div v-if="campaignId !== null" class="flex justify-end">
      <Button data-test="review-button" class="w-full sm:w-auto" :disabled="running" @click="reviewOpen = true">
        Review &amp; approve
      </Button>
    </div>
    <ReviewDialog
      v-if="campaignId !== null"
      v-model:open="reviewOpen"
      :campaign-id="campaignId"
      @approved="id => router.push(`/campaigns/${id}?sending=1`)"
    />
  </div>
</template>
