<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import type { Campaign } from '@/lib/campaigns'
import type { Progress } from '@/lib/delivery'

import DeliverySection from '@/components/campaign/delivery-section.vue'
import ReviewDialog from '@/components/campaign/review-dialog.vue'
import SendingDialog from '@/components/campaign/sending-dialog.vue'
import StatusBadge from '@/components/campaign/status-badge.vue'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { CampaignRequestError, getCampaign, languageName } from '@/lib/campaigns'
import { getProgress, watchProgress } from '@/lib/delivery'

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
let stopWatch: (() => void) | null = null
let loadToken = 0

const fmt = (n: number) => n.toLocaleString('en-US')

function stopWatching() {
  stopWatch?.()
  stopWatch = null
}

// The page owns the progress stream, so that the status, the badge and the buttons follow it.
function startWatching() {
  stopWatching()
  if (!progress.value || !['approved', 'queued', 'sending'].includes(progress.value.status))
    return
  stopWatch = watchProgress(id.value, (p) => {
    progress.value = p
    progressError.value = ''
    if (campaign.value)
      campaign.value.status = p.status
    if (p.status === 'completed')
      stopWatching()
  })
}

async function refreshProgress() {
  try {
    const p = await getProgress(id.value)
    progress.value = p
    progressError.value = ''
    if (campaign.value)
      campaign.value.status = p.status
    startWatching()
  }
  catch (err) {
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
    progress.value = null
    progressError.value = ''
    if (c.status !== 'draft')
      await refreshProgress()
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
      <div class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
        <div class="space-y-2">
          <div class="flex items-center gap-2">
            <h1 class="text-2xl font-bold tracking-tight">
              Campaign #{{ campaign.id }}
            </h1>
            <StatusBadge :status="campaign.status" />
          </div>
          <p class="text-muted-foreground break-words">
            {{ campaign.description }}
          </p>
        </div>
        <Button v-if="campaign.status === 'draft'" data-test="review-button" class="w-full sm:w-auto" @click="reviewOpen = true">
          Review &amp; approve
        </Button>
      </div>

      <p v-for="s in campaign.skipped_languages" :key="s" class="text-muted-foreground text-sm">
        {{ languageName(s) }} skipped: content failed.
      </p>

      <Card v-for="c in campaign.contents" :key="c.language" :data-language="c.language">
        <CardHeader>
          <CardTitle class="flex flex-wrap items-baseline gap-2">
            {{ languageName(c.language) }}
            <span class="text-muted-foreground text-sm font-normal tabular-nums">{{ fmt(c.sms_count) }} SMS · {{ fmt(c.email_count) }} email</span>
          </CardTitle>
        </CardHeader>
        <CardContent class="space-y-4">
          <div>
            <h4 class="text-sm font-medium">
              SMS
            </h4>
            <p class="break-words whitespace-pre-wrap">
              {{ c.sms }}
            </p>
          </div>
          <div>
            <h4 class="text-sm font-medium">
              Email subject
            </h4>
            <p class="break-words">
              {{ c.email_subject }}
            </p>
          </div>
          <div>
            <h4 class="text-sm font-medium">
              Email body
            </h4>
            <p class="break-words whitespace-pre-wrap">
              {{ c.email_body }}
            </p>
          </div>
        </CardContent>
      </Card>

      <div v-if="campaign.status !== 'draft'" ref="deliveryEl" class="scroll-mt-4">
        <DeliverySection
          v-if="progress"
          :campaign-id="campaign.id"
          :progress="progress"
          @view-progress="sendingOpen = true"
          @retried="onRetried"
          @refresh="refreshProgress"
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
  </div>
</template>
