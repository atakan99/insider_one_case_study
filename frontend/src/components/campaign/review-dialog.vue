<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { toast } from 'vue-sonner'

import type { Campaign, ContentInput } from '@/lib/campaigns'
import type { Content } from '@/lib/content'

import { Alert, AlertDescription } from '@/components/ui/alert'
import {
  AlertDialog,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Spinner } from '@/components/ui/spinner'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Textarea } from '@/components/ui/textarea'
import { approveCampaign, CampaignRequestError, getCampaign, getSampleNames, languageName, saveContents } from '@/lib/campaigns'
import { checkContent, LIMITS, renderedLength } from '@/lib/limits'

import MessagePreview from './message-preview.vue'

type Field = keyof Content
type Channel = 'sms' | 'email'

const props = defineProps<{ campaignId: number }>()
const emit = defineEmits<{ approved: [id: number] }>()
const open = defineModel<boolean>('open', { default: false })

const FALLBACK_NAME = 'Alex'
const VAR_LABEL = '{{first_name}}'
const FIELDS: Field[] = ['sms', 'email_subject', 'email_body']

const campaign = ref<Campaign | null>(null)
const loadError = ref('')
const edits = reactive<Record<string, Content>>({})
const saved = ref<Record<string, Content>>({})
const samples = ref<Record<string, string[]>>({})
const sampleFailed = ref(false)
const chosenName = reactive<Record<string, string>>({})
const lang = ref('')
const channel = ref<Channel>('sms')
const serverErrors = reactive<Record<string, Partial<Record<Field, string>>>>({})
const banner = ref('')
const busy = ref(false)
const confirmOpen = ref(false)
const discardOpen = ref(false)
let loadedFor = 0
let loadToken = 0

const isDraft = computed(() => campaign.value?.status === 'draft')
const contents = computed(() => campaign.value?.contents ?? [])
const current = computed(() => contents.value.find(c => c.language === lang.value))
const nameRunes = computed(() => current.value?.first_name_runes ?? 0)

function copyContent(c: Content): Content {
  return { sms: c.sms, email_subject: c.email_subject, email_body: c.email_body }
}

function reset(c: Campaign) {
  campaign.value = c
  for (const k of Object.keys(edits)) delete edits[k]
  for (const k of Object.keys(serverErrors)) delete serverErrors[k]
  const snap: Record<string, Content> = {}
  for (const lc of c.contents) {
    edits[lc.language] = copyContent(lc)
    snap[lc.language] = copyContent(lc)
  }
  saved.value = snap
  if (!c.contents.some(lc => lc.language === lang.value))
    lang.value = c.contents[0]?.language ?? ''
  banner.value = ''
}

async function load() {
  const id = props.campaignId
  const mine = ++loadToken
  loadError.value = ''
  loadedFor = id
  try {
    const c = await getCampaign(id)
    if (mine !== loadToken)
      return
    reset(c)
  }
  catch (err) {
    if (mine === loadToken)
      loadError.value = err instanceof CampaignRequestError ? err.message : 'Something went wrong. Try again.'
    return
  }
  try {
    const names = await getSampleNames(id)
    if (mine !== loadToken)
      return
    samples.value = names
    sampleFailed.value = false
  }
  catch {
    if (mine !== loadToken)
      return
    samples.value = {}
    sampleFailed.value = true
  }
}

// Reloads the counts and limits after a failed check, and keeps the edits.
async function refreshLimits() {
  const id = props.campaignId
  const mine = loadToken
  try {
    const c = await getCampaign(id)
    if (mine === loadToken && id === props.campaignId)
      campaign.value = c
  }
  catch {
    // The old limits stay. The server check still decides.
  }
}

const dirty = computed(() => Object.keys(edits).some(l =>
  FIELDS.some(f => edits[l]![f] !== saved.value[l]?.[f])))

// Load on each open, so that a regenerated draft shows its new text. Unsaved
// edits of the same campaign stay.
watch([open, () => props.campaignId], ([o, id]) => {
  if (o && (loadedFor !== id || !campaign.value || !dirty.value))
    void load()
}, { immediate: true })

const clientErrors = computed(() => {
  const out: Record<string, Partial<Record<Field, string>>> = {}
  for (const lc of contents.value) {
    const e = edits[lc.language]
    if (e)
      out[lc.language] = checkContent(e, lc.first_name_runes)
  }
  return out
})

function fieldError(l: string, f: Field): string | undefined {
  return clientErrors.value[l]?.[f] ?? serverErrors[l]?.[f]
}

function langHasError(l: string): boolean {
  return FIELDS.some(f => fieldError(l, f) !== undefined)
}

const hasClientErrors = computed(() => Object.values(clientErrors.value).some(e => Object.keys(e).length > 0))
const canSubmit = computed(() => isDraft.value && !busy.value && !hasClientErrors.value)

function counter(f: Field) {
  const n = renderedLength(edits[lang.value]?.[f] ?? '', nameRunes.value)
  const limit = LIMITS[f]
  const tone = n > limit ? 'text-destructive font-medium' : n > limit * 0.9 ? 'text-warning' : 'text-muted-foreground'
  return { label: `${n.toLocaleString('en-US')} / ${limit.toLocaleString('en-US')}`, tone }
}

function onInput(f: Field, v: string | number) {
  const e = edits[lang.value]
  if (!e)
    return
  e[f] = String(v)
  if (serverErrors[lang.value])
    delete serverErrors[lang.value]![f]
}

const names = computed(() => {
  const list = samples.value[lang.value]
  return list && list.length > 0 ? list : [FALLBACK_NAME]
})
const previewName = computed({
  get: () => {
    const n = chosenName[lang.value]
    return n && names.value.includes(n) ? n : names.value[0]!
  },
  set: (v: string) => {
    chosenName[lang.value] = v
  },
})
function nameLabel(n: string, i: number) {
  return i === 0 && samples.value[lang.value]?.length ? `${n} (longest)` : n
}

const fmt = (n: number) => n.toLocaleString('en-US')
const totals = computed(() => contents.value.reduce((t, c) => ({ sms: t.sms + c.sms_count, email: t.email + c.email_count }), { sms: 0, email: 0 }))
const skipped = computed(() => campaign.value?.skipped_languages ?? [])
const languagesLabel = computed(() => contents.value.map(c => languageName(c.language)).join(', '))

function payload(): ContentInput[] {
  return contents.value.map(c => ({ language: c.language, ...copyContent(edits[c.language]!) }))
}

async function applyError(err: unknown, sent: ContentInput[]) {
  if (err instanceof CampaignRequestError && err.code === 'content_invalid' && err.details.length > 0) {
    const rest: string[] = []
    for (const d of err.details) {
      const m = /^contents\[(\d+)\]\.(sms|email_subject|email_body)$/.exec(d.path)
      const l = m ? sent[Number(m[1])]?.language : undefined
      if (m && l) {
        serverErrors[l] = { ...serverErrors[l], [m[2] as Field]: d.message }
      }
      else {
        rest.push(d.message)
      }
    }
    banner.value = rest.length > 0 ? `${err.message} ${rest.join(' ')}` : ''
    await refreshLimits()
    return
  }
  banner.value = err instanceof CampaignRequestError ? err.message : 'Something went wrong. Try again.'
}

async function saveDraft() {
  if (!canSubmit.value)
    return
  busy.value = true
  banner.value = ''
  const sent = payload()
  try {
    const c = await saveContents(props.campaignId, sent)
    // The fields stay editable during the request. Text typed after the
    // submit is newer than the response, so it stays and counts as unsaved.
    const newer: Record<string, Partial<Content>> = {}
    for (const s of sent) {
      for (const f of FIELDS) {
        const now = edits[s.language]?.[f]
        if (now !== undefined && now !== s[f])
          (newer[s.language] ??= {})[f] = now
      }
    }
    reset(c)
    for (const [l, fields] of Object.entries(newer)) {
      if (edits[l])
        Object.assign(edits[l], fields)
    }
    toast.success('Draft saved.')
  }
  catch (err) {
    await applyError(err, sent)
  }
  finally {
    busy.value = false
  }
}

async function sendNow() {
  if (busy.value || !isDraft.value)
    return
  busy.value = true
  banner.value = ''
  const sent = payload()
  try {
    const c = await approveCampaign(props.campaignId, sent)
    reset(c)
    confirmOpen.value = false
    open.value = false
    emit('approved', c.id)
  }
  catch (err) {
    confirmOpen.value = false
    await applyError(err, sent)
  }
  finally {
    busy.value = false
  }
}

// Every close path (close button, Cancel, Escape, outside click) comes here.
function requestClose(v: boolean) {
  if (v) {
    open.value = true
    return
  }
  if (busy.value)
    return
  if (isDraft.value && dirty.value) {
    discardOpen.value = true
    return
  }
  open.value = false
}

function discard() {
  if (campaign.value)
    reset(campaign.value)
  discardOpen.value = false
  open.value = false
}
</script>

<template>
  <Dialog :open="open" @update:open="requestClose">
    <DialogContent
      data-test="review"
      class="flex h-dvh w-screen max-w-none flex-col gap-0 overflow-hidden rounded-none p-0 sm:max-w-none md:h-auto md:max-h-[calc(100dvh-2rem)] md:w-[min(920px,calc(100%-2rem))] md:max-w-[920px] md:rounded-lg"
    >
      <DialogHeader class="border-b px-4 py-4 text-left md:px-6">
        <DialogTitle>Review campaign</DialogTitle>
        <DialogDescription>Check and edit the messages. Nothing is sent until you confirm.</DialogDescription>
      </DialogHeader>

      <div v-if="loadError" class="p-6">
        <Alert variant="destructive">
          <AlertDescription>{{ loadError }}</AlertDescription>
        </Alert>
      </div>

      <template v-else-if="campaign">
        <div class="flex flex-col gap-3 border-b px-4 py-3 md:flex-row md:items-center md:justify-between md:px-6">
          <Tabs v-model="channel">
            <TabsList class="w-full md:w-auto">
              <TabsTrigger data-test="channel-tab" data-value="sms" value="sms" class="flex-1 md:flex-none">
                SMS <span class="text-muted-foreground text-xs tabular-nums">{{ fmt(current?.sms_count ?? 0) }}</span>
              </TabsTrigger>
              <TabsTrigger data-test="channel-tab" data-value="email" value="email" class="flex-1 md:flex-none">
                Email <span class="text-muted-foreground text-xs tabular-nums">{{ fmt(current?.email_count ?? 0) }}</span>
              </TabsTrigger>
            </TabsList>
          </Tabs>
          <Tabs v-model="lang">
            <TabsList class="w-full md:w-auto">
              <TabsTrigger
                v-for="c in contents"
                :key="c.language"
                data-test="lang-tab"
                :data-value="c.language"
                :value="c.language"
                class="flex-1 md:flex-none"
              >
                {{ languageName(c.language) }}<span v-if="langHasError(c.language)" data-test="tab-error" class="bg-destructive ml-1.5 inline-block size-1.5 rounded-full" aria-label="has errors" />
              </TabsTrigger>
            </TabsList>
          </Tabs>
        </div>

        <div class="grid min-h-0 flex-1 grid-cols-1 gap-4 overflow-y-auto px-4 py-4 md:grid-cols-2 md:gap-6 md:px-6">
          <Alert v-if="banner" data-test="banner" variant="destructive" class="md:col-span-2">
            <AlertDescription>{{ banner }}</AlertDescription>
          </Alert>

          <div v-if="isDraft && edits[lang]" class="min-w-0 space-y-4">
            <p class="text-muted-foreground text-xs font-semibold tracking-wide uppercase">
              Edit
            </p>
            <template v-if="channel === 'sms'">
              <div class="space-y-1.5">
                <div class="flex items-baseline justify-between">
                  <Label for="review-sms">Message</Label>
                  <span data-test="sms-counter" class="text-xs tabular-nums" :class="counter('sms').tone">{{ counter('sms').label }}</span>
                </div>
                <Textarea id="review-sms" data-test="sms-input" class="min-h-24" :model-value="edits[lang]!.sms" :aria-invalid="!!fieldError(lang, 'sms')" @update:model-value="v => onInput('sms', v)" />
                <p v-if="fieldError(lang, 'sms')" data-test="sms-error" class="text-destructive text-xs">
                  {{ fieldError(lang, 'sms') }}
                </p>
                <p class="text-muted-foreground text-xs">
                  Counted with the longest first name ({{ nameRunes }} characters). Use <code class="bg-muted rounded px-1" v-text="VAR_LABEL" /> for the name.
                </p>
              </div>
            </template>
            <template v-else>
              <div class="space-y-1.5">
                <div class="flex items-baseline justify-between">
                  <Label for="review-subject">Subject</Label>
                  <span data-test="subject-counter" class="text-xs tabular-nums" :class="counter('email_subject').tone">{{ counter('email_subject').label }}</span>
                </div>
                <Input id="review-subject" data-test="subject-input" :model-value="edits[lang]!.email_subject" :aria-invalid="!!fieldError(lang, 'email_subject')" @update:model-value="v => onInput('email_subject', v)" />
                <p v-if="fieldError(lang, 'email_subject')" data-test="subject-error" class="text-destructive text-xs">
                  {{ fieldError(lang, 'email_subject') }}
                </p>
              </div>
              <div class="space-y-1.5">
                <div class="flex items-baseline justify-between">
                  <Label for="review-body">Body</Label>
                  <span data-test="body-counter" class="text-xs tabular-nums" :class="counter('email_body').tone">{{ counter('email_body').label }}</span>
                </div>
                <Textarea id="review-body" data-test="body-input" class="min-h-48" :model-value="edits[lang]!.email_body" :aria-invalid="!!fieldError(lang, 'email_body')" @update:model-value="v => onInput('email_body', v)" />
                <p v-if="fieldError(lang, 'email_body')" data-test="body-error" class="text-destructive text-xs">
                  {{ fieldError(lang, 'email_body') }}
                </p>
                <p class="text-muted-foreground text-xs">
                  Plain text. No HTML.
                </p>
              </div>
            </template>
          </div>

          <div class="min-w-0 space-y-2" :class="{ 'md:col-span-2': !isDraft }">
            <div class="flex items-center justify-between gap-2">
              <p class="text-muted-foreground text-xs font-semibold tracking-wide uppercase">
                Preview
              </p>
              <Select v-model="previewName">
                <SelectTrigger data-test="name-select" size="sm" class="max-w-[60%]" aria-label="Preview as">
                  <SelectValue>{{ nameLabel(previewName, names.indexOf(previewName)) }}</SelectValue>
                </SelectTrigger>
                <SelectContent>
                  <SelectItem v-for="(n, i) in names" :key="n" :value="n">
                    {{ nameLabel(n, i) }}
                  </SelectItem>
                </SelectContent>
              </Select>
            </div>
            <p v-if="sampleFailed" data-test="sample-note" class="text-muted-foreground text-xs">
              Sample names could not be loaded. The preview uses "{{ FALLBACK_NAME }}".
            </p>
            <MessagePreview v-if="edits[lang]" :channel="channel" :content="edits[lang]!" :name="previewName" />
          </div>
        </div>

        <div class="flex flex-wrap items-center gap-3 border-t px-4 py-3 md:flex-nowrap md:px-6">
          <p data-test="footer-note" class="text-muted-foreground w-full text-xs md:mr-auto md:w-auto">
            {{ contents.length }} {{ contents.length === 1 ? 'language' : 'languages' }} · {{ fmt(totals.sms) }} SMS · {{ fmt(totals.email) }} email
            <span v-for="s in skipped" :key="s" class="block">{{ languageName(s) }} skipped: content failed.</span>
          </p>
          <template v-if="isDraft">
            <Button data-test="cancel" variant="outline" class="flex-1 md:flex-none" :disabled="busy" @click="requestClose(false)">
              Cancel
            </Button>
            <Button data-test="save" variant="outline" class="flex-1 md:flex-none" :disabled="!canSubmit" @click="saveDraft">
              Save draft
            </Button>
            <Button data-test="approve" class="w-full md:w-auto" :disabled="!canSubmit" @click="confirmOpen = true">
              Approve &amp; send
            </Button>
          </template>
          <Button v-else variant="outline" class="w-full md:w-auto" @click="requestClose(false)">
            Close
          </Button>
        </div>
      </template>
    </DialogContent>
  </Dialog>

  <AlertDialog :open="confirmOpen" @update:open="v => { if (!busy) confirmOpen = v }">
    <AlertDialogContent data-test="confirm">
      <AlertDialogHeader>
        <AlertDialogTitle>Send this campaign?</AlertDialogTitle>
        <AlertDialogDescription>
          Messages go out now. Only users who gave consent for a channel get that channel. You cannot undo this.
        </AlertDialogDescription>
      </AlertDialogHeader>
      <dl class="divide-y rounded-md border text-sm tabular-nums">
        <div class="flex justify-between px-3 py-2">
          <dt>SMS</dt><dd>{{ fmt(totals.sms) }}</dd>
        </div>
        <div class="flex justify-between px-3 py-2">
          <dt>Email</dt><dd>{{ fmt(totals.email) }}</dd>
        </div>
        <div class="flex justify-between gap-4 px-3 py-2">
          <dt>Languages</dt><dd class="text-right">
            {{ languagesLabel }}
          </dd>
        </div>
        <div v-for="s in skipped" :key="s" class="text-muted-foreground px-3 py-2">
          {{ languageName(s) }} skipped: content failed.
        </div>
      </dl>
      <AlertDialogFooter>
        <Button data-test="confirm-back" variant="outline" :disabled="busy" @click="confirmOpen = false">
          Back
        </Button>
        <Button data-test="send-now" :disabled="busy" @click="sendNow">
          <Spinner v-if="busy" />
          Send now
        </Button>
      </AlertDialogFooter>
    </AlertDialogContent>
  </AlertDialog>

  <AlertDialog :open="discardOpen" @update:open="v => discardOpen = v">
    <AlertDialogContent data-test="discard">
      <AlertDialogHeader>
        <AlertDialogTitle>Discard your changes?</AlertDialogTitle>
        <AlertDialogDescription>Your edits are not saved. Save the draft to keep them.</AlertDialogDescription>
      </AlertDialogHeader>
      <AlertDialogFooter>
        <Button variant="outline" @click="discardOpen = false">
          Keep editing
        </Button>
        <Button data-test="discard-confirm" variant="destructive" @click="discard">
          Discard
        </Button>
      </AlertDialogFooter>
    </AlertDialogContent>
  </AlertDialog>
</template>
