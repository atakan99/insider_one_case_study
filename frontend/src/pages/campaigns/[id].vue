<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute } from 'vue-router'

import type { Campaign } from '@/lib/campaigns'

import ReviewDialog from '@/components/campaign/review-dialog.vue'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { CampaignRequestError, getCampaign, languageName } from '@/lib/campaigns'

const route = useRoute('/campaigns/[id]')
const id = computed(() => Number(route.params.id))
const campaign = ref<Campaign | null>(null)
const notFound = ref(false)
const error = ref('')
const reviewOpen = ref(false)

const fmt = (n: number) => n.toLocaleString('en-US')

async function load() {
  error.value = ''
  notFound.value = false
  if (!Number.isInteger(id.value) || id.value <= 0) {
    notFound.value = true
    return
  }
  try {
    campaign.value = await getCampaign(id.value)
  }
  catch (err) {
    if (err instanceof CampaignRequestError && (err.status === 404 || err.status === 400))
      notFound.value = true
    else
      error.value = err instanceof CampaignRequestError ? err.message : 'Something went wrong. Try again.'
  }
}

watch(id, load, { immediate: true })

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
            <Badge :variant="campaign.status === 'draft' ? 'outline' : 'default'">
              {{ campaign.status === 'draft' ? 'Draft' : 'Approved' }}
            </Badge>
          </div>
          <p class="text-muted-foreground break-words">
            {{ campaign.description }}
          </p>
        </div>
        <Button v-if="campaign.status === 'draft'" data-test="review-button" class="w-full sm:w-auto" @click="reviewOpen = true">
          Review &amp; approve
        </Button>
      </div>

      <p v-if="campaign.status === 'approved'" class="text-muted-foreground text-sm">
        Sending starts in a later step.
      </p>
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

      <ReviewDialog
        v-if="campaign.status === 'draft'"
        :open="reviewOpen"
        :campaign-id="campaign.id"
        @update:open="onReviewOpen"
        @approved="load"
      />
    </template>
  </div>
</template>
