<script setup lang="ts">
import { ArrowLeftIcon } from '@lucide/vue'
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import type { Campaign } from '@/lib/campaigns'
import type { Progress } from '@/lib/delivery'

import ContentView from '@/components/campaign/content-view.vue'
import DeliveryBar from '@/components/campaign/delivery-bar.vue'
import DeliverySection from '@/components/campaign/delivery-section.vue'
import ReviewDialog from '@/components/campaign/review-dialog.vue'
import SendingDialog from '@/components/campaign/sending-dialog.vue'
import StatusBadge from '@/components/campaign/status-badge.vue'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { CampaignRequestError, getCampaign, languageName } from '@/lib/campaigns'
import { getProgress, watchProgress } from '@/lib/delivery'
import { formatCount as fmt, formatDateTime } from '@/lib/format'
import { renderedLength } from '@/lib/limits'

const route = useRoute('/campaigns/[id]')
const router = useRouter()
const id = computed(() => Number(route.params.id))
const campaign = ref<Campaign | null>(null)
const notFound = ref(false)
const error = ref('')
const reviewOpen = ref(false)
const sendingOpen = ref(false)
const progress = ref<Progress | null>(null)
const progressError = ref('')
const deliveryEl = ref<HTMLElement | null>(null)
const FALLBACK_MESSAGE = 'Something went wrong. Try again.'
const lang = ref('')
let stopWatch: (() => void) | null = null
let loadToken = 0

const CHANNELS = [
  { key: 'sms', label: 'SMS' },
  { key: 'email', label: 'Email' },
] as const

const current = computed(() => campaign.value?.contents.find(c => c.language === lang.value) ?? campaign.value?.contents[0])
const totals = computed(() => (campaign.value?.contents ?? []).reduce(
  (t, c) => ({ sms: t.sms + c.sms_count, email: t.email + c.email_count }),
  { sms: 0, email: 0 },
))
// The short summary at the top. The full list is in the delivery section.
const summary = computed(() => {
  const p = progress.value
  if (!p)
    return []
  return CHANNELS.map(c => ({ ...c, p: p.channels[c.key] })).filter(c => c.p.total > 0)
})

function stopWatching() {
  stopWatch?.()
  stopWatch = null
}

// The page owns the progress stream, so that the status, the badge and the buttons follow it.
function startWatching() {
  stopWatching()
  if (!progress.value || !['approved', 'queued', 'sending'].includes(progress.value.status))
    return
  const watchedId = id.value
  stopWatch = watchProgress(watchedId, (p) => {
    if (watchedId !== id.value)
      return
    progress.value = p
    progressError.value = ''
    if (campaign.value)
      campaign.value.status = p.status
    if (p.status === 'completed')
      stopWatching()
  })
}

// A reload bumps loadToken, so a response from an older load or id is dropped.
async function refreshProgress(mine = loadToken) {
  try {
    const p = await getProgress(id.value)
    if (mine !== loadToken)
      return
    progress.value = p
    progressError.value = ''
    if (campaign.value)
      campaign.value.status = p.status
    startWatching()
  }
  catch (err) {
    if (mine !== loadToken)
      return
    progressError.value = err instanceof CampaignRequestError ? err.message : FALLBACK_MESSAGE
  }
}

async function load() {
  const mine = ++loadToken
  stopWatching()
  error.value = ''
  notFound.value = false
  if (!Number.isInteger(id.value) || id.value <= 0) {
    notFound.value = true
    return
  }
  try {
    const c = await getCampaign(id.value)
    if (mine !== loadToken)
      return
    campaign.value = c
    if (!c.contents.some(x => x.language === lang.value))
      lang.value = c.contents[0]?.language ?? ''
    progress.value = null
    progressError.value = ''
    if (c.status !== 'draft')
      await refreshProgress(mine)
  }
  catch (err) {
    if (mine !== loadToken)
      return
    if (err instanceof CampaignRequestError && (err.status === 404 || err.status === 400))
      notFound.value = true
    else
      error.value = err instanceof CampaignRequestError ? err.message : FALLBACK_MESSAGE
  }
}

onBeforeUnmount(() => {
  loadToken++
  stopWatching()
})

watch(id, load, { immediate: true })

// The approval routes here with ?sending=1. Open the window once, then drop the query.
watch(() => route.query.sending, async (v) => {
  if (v !== '1')
    return
  sendingOpen.value = true
  await router.replace({ query: {} })
  if (campaign.value)
    await load()
}, { immediate: true })

function onApproved(approvedId: number) {
  void router.push(`/campaigns/${approvedId}?sending=1`)
}

function onRetried(p: Progress) {
  progress.value = p
  if (campaign.value)
    campaign.value.status = p.status
  startWatching()
  sendingOpen.value = true
}

async function showList() {
  await nextTick()
  deliveryEl.value?.scrollIntoView({ behavior: 'smooth', block: 'start' })
}

function onReviewOpen(v: boolean) {
  reviewOpen.value = v
  if (!v)
    void load()
}
</script>

<template>
  <div class="space-y-6">
    <div v-if="notFound" class="space-y-3">
      <h1 class="text-2xl font-bold tracking-tight">
        Campaign not found
      </h1>
      <Button as-child variant="outline">
        <RouterLink to="/campaigns">
          Back to campaigns
        </RouterLink>
      </Button>
    </div>

    <Alert v-else-if="error" variant="destructive">
      <AlertDescription>{{ error }}</AlertDescription>
    </Alert>

    <template v-else-if="campaign">
      <div class="space-y-3">
        <RouterLink to="/campaigns" class="text-muted-foreground hover:text-foreground inline-flex items-center gap-1 text-sm">
          <ArrowLeftIcon class="size-4" />
          Campaigns
        </RouterLink>
        <div class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
          <div class="space-y-1">
            <div class="flex items-center gap-2">
              <h1 class="text-2xl font-bold tracking-tight">
                Campaign #{{ campaign.id }}
              </h1>
              <StatusBadge :status="campaign.status" />
            </div>
            <p data-test="meta" class="text-muted-foreground text-sm tabular-nums">
              Created {{ formatDateTime(campaign.created_at) }} · {{ fmt(totals.sms) }} SMS · {{ fmt(totals.email) }} email
            </p>
          </div>
          <Button v-if="campaign.status === 'draft'" data-test="review-button" class="w-full sm:w-auto" @click="reviewOpen = true">
            Review &amp; approve
          </Button>
        </div>
      </div>

      <div class="space-y-1">
        <h2 class="text-muted-foreground text-xs font-medium tracking-wide uppercase">
          Audience
        </h2>
        <p class="max-w-prose break-words">
          {{ campaign.description }}
        </p>
      </div>

      <Card v-if="summary.length > 0" data-test="delivery-summary" class="py-4">
        <CardContent class="space-y-3 px-4">
          <div v-for="c in summary" :key="c.key" :data-summary-row="c.key" class="space-y-1.5">
            <div class="flex flex-wrap items-baseline justify-between gap-x-4 text-sm tabular-nums">
              <span class="font-medium">{{ c.label }}</span>
              <span class="text-muted-foreground">
                <span class="text-foreground">{{ fmt(c.p.sent) }}</span> of {{ fmt(c.p.total) }} sent
                <template v-if="c.p.failed > 0">
                  · <span class="text-destructive">{{ fmt(c.p.failed) }} failed</span>
                </template>
                <template v-if="c.p.unknown > 0">
                  · {{ fmt(c.p.unknown) }} unknown
                </template>
                <template v-if="c.p.skipped > 0">
                  · {{ fmt(c.p.skipped) }} skipped
                </template>
              </span>
            </div>
            <DeliveryBar :progress="c.p" />
          </div>
          <button type="button" data-test="to-delivery" class="text-muted-foreground hover:text-foreground text-sm underline underline-offset-4" @click="showList">
            See each message
          </button>
        </CardContent>
      </Card>

      <section class="space-y-3">
        <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <h2 class="text-lg font-semibold tracking-tight">
            Messages
          </h2>
          <Tabs v-if="campaign.contents.length > 1" v-model="lang">
            <TabsList class="w-full sm:w-auto">
              <TabsTrigger
                v-for="c in campaign.contents"
                :key="c.language"
                data-test="lang-tab"
                :data-value="c.language"
                :value="c.language"
                class="flex-1 sm:flex-none"
              >
                {{ languageName(c.language) }}
              </TabsTrigger>
            </TabsList>
          </Tabs>
        </div>

        <p v-for="s in campaign.skipped_languages" :key="s" class="text-muted-foreground text-sm">
          {{ languageName(s) }} skipped: content failed.
        </p>

        <Card v-if="current" :key="current.language" :data-language="current.language" class="py-5">
          <CardContent class="space-y-4 px-5">
            <p class="flex flex-wrap items-baseline gap-2 font-semibold">
              {{ languageName(current.language) }}
              <span class="text-muted-foreground text-sm font-normal tabular-nums">{{ fmt(current.sms_count) }} SMS · {{ fmt(current.email_count) }} email</span>
            </p>
            <ContentView
              :content="current"
              :sms-length="renderedLength(current.sms, current.first_name_runes)"
              :subject-length="renderedLength(current.email_subject, current.first_name_runes)"
            />
          </CardContent>
        </Card>
      </section>

      <div v-if="campaign.status !== 'draft'" ref="deliveryEl" class="scroll-mt-16">
        <DeliverySection
          v-if="progress"
          :campaign-id="campaign.id"
          :progress="progress"
          @view-progress="sendingOpen = true"
          @retried="onRetried"
          @refresh="refreshProgress()"
        />
        <p v-if="progressError" data-test="progress-error" class="text-destructive text-sm">
          {{ progressError }}
        </p>
      </div>

      <SendingDialog
        v-if="campaign.status !== 'draft'"
        v-model:open="sendingOpen"
        :campaign-id="campaign.id"
        :progress="progress"
        @open-list="showList"
      />

      <ReviewDialog
        v-if="campaign.status === 'draft'"
        :open="reviewOpen"
        :campaign-id="campaign.id"
        @update:open="onReviewOpen"
        @approved="onApproved"
      />
    </template>

    <div v-else data-test="loading" class="space-y-4" role="status" aria-label="Loading the campaign">
      <Skeleton class="h-8 w-56" />
      <Skeleton class="h-4 w-80 max-w-full" />
      <Skeleton class="h-64 w-full" />
    </div>
  </div>
</template>
