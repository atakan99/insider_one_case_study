<script setup lang="ts">
import { onBeforeUnmount, reactive, ref, watch } from 'vue'
import { useRouter } from 'vue-router'

import type { Content, ContentEvent, ContentLimits } from '@/lib/content'

import { Alert, AlertDescription } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { ContentRequestError, generateContent } from '@/lib/content'

import ReviewDialog from './review-dialog.vue'

interface LangCard {
  code: string
  email: number
  sms: number
  draft?: Content
  final?: { content: Content, limits: ContentLimits }
  retrying: boolean
  error?: string
  closed: boolean
}

const props = defineProps<{ brief: string, rule: unknown, runId: number }>()
const emit = defineEmits<{ 'update:running': [running: boolean] }>()

const NAMES: Record<string, string> = { tr: 'Turkish', en: 'English' }
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

function langName(code: string) {
  return NAMES[code] ?? code.toUpperCase()
}

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
      retrying: false,
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
    card.retrying = false
  }
  else if (e.event === 'retry') {
    card.draft = undefined
    card.retrying = true
  }
  else if (e.event === 'final') {
    card.final = { content: e.data.content, limits: e.data.limits }
    card.retrying = false
    card.error = undefined
  }
  else if (e.event === 'error') {
    card.error = e.data.message
    card.retrying = false
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
        <CardTitle class="flex items-center gap-2">
          {{ langName(card.code) }}
          <span class="text-sm font-normal text-muted-foreground">
            {{ card.email }} email, {{ card.sms }} SMS
          </span>
          <Badge v-if="!card.final && !card.error && card.draft" variant="secondary">
            Draft
          </Badge>
        </CardTitle>
        <p v-if="card.retrying" class="text-sm text-muted-foreground" role="status">
          Fixing content…
        </p>
      </CardHeader>
      <CardContent class="space-y-4">
        <Alert v-if="card.error" variant="destructive">
          <AlertDescription>{{ card.error }}</AlertDescription>
        </Alert>
        <Alert v-else-if="card.closed" variant="destructive">
          <AlertDescription>{{ CLOSED_MESSAGE }}</AlertDescription>
        </Alert>
        <template v-if="!card.error && (card.final || card.draft)">
          <div class="space-y-4" :class="card.final ? '' : 'text-muted-foreground italic'">
            <div>
              <h4 class="text-sm font-medium">
                SMS
              </h4>
              <p class="whitespace-pre-wrap" data-field="sms">
                {{ (card.final?.content ?? card.draft)!.sms }}
              </p>
              <p v-if="card.final" class="text-xs text-muted-foreground">
                {{ card.final.limits.sms_runes }} / 160
              </p>
            </div>
            <div>
              <h4 class="text-sm font-medium">
                Email subject
              </h4>
              <p class="whitespace-pre-wrap" data-field="subject">
                {{ (card.final?.content ?? card.draft)!.email_subject }}
              </p>
              <p v-if="card.final" class="text-xs text-muted-foreground">
                {{ card.final.limits.email_subject_runes }} / 80
              </p>
            </div>
            <div>
              <h4 class="text-sm font-medium">
                Email body
              </h4>
              <p class="whitespace-pre-wrap" data-field="body">
                {{ (card.final?.content ?? card.draft)!.email_body }}
              </p>
            </div>
          </div>
        </template>
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
      @approved="id => router.push(`/campaigns/${id}`)"
    />
  </div>
</template>
