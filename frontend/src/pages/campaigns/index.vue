<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import type { CampaignSummary } from '@/lib/campaigns'

import ReviewDialog from '@/components/campaign/review-dialog.vue'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { CampaignRequestError, languageName, listCampaigns } from '@/lib/campaigns'

const router = useRouter()
const campaigns = ref<CampaignSummary[] | null>(null)
const error = ref('')
const reviewId = ref<number | null>(null)
const reviewOpen = ref(false)

async function load() {
  error.value = ''
  try {
    campaigns.value = await listCampaigns()
  }
  catch (err) {
    error.value = err instanceof CampaignRequestError ? err.message : 'Something went wrong. Try again.'
  }
}

onMounted(load)

function languages(c: CampaignSummary) {
  return c.languages.map(languageName).join(', ')
}

function when(iso: string) {
  return new Date(iso).toLocaleString()
}

function openCampaign(c: CampaignSummary) {
  if (c.status === 'draft') {
    reviewId.value = c.id
    reviewOpen.value = true
  }
  else {
    router.push(`/campaigns/${c.id}`)
  }
}

function onReviewOpen(v: boolean) {
  reviewOpen.value = v
  if (!v)
    void load()
}

function onApproved(id: number) {
  router.push(`/campaigns/${id}`)
}
</script>

<template>
  <div class="space-y-6">
    <div class="space-y-2">
      <h1 class="text-2xl font-bold tracking-tight">
        Campaigns
      </h1>
      <p class="text-muted-foreground">
        Drafts wait for your review. Approved campaigns are ready to send.
      </p>
    </div>

    <Alert v-if="error" variant="destructive">
      <AlertDescription>{{ error }}</AlertDescription>
    </Alert>

    <div v-else-if="campaigns && campaigns.length === 0" class="space-y-3 rounded-lg border border-dashed p-8 text-center">
      <p class="font-medium">
        No campaigns yet
      </p>
      <Button as-child variant="outline">
        <RouterLink to="/campaigns/new">
          Create a campaign
        </RouterLink>
      </Button>
    </div>

    <template v-else-if="campaigns">
      <div class="hidden rounded-lg border md:block">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead class="w-16">
                #
              </TableHead>
              <TableHead>Audience</TableHead>
              <TableHead>Languages</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Created</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow
              v-for="c in campaigns"
              :key="c.id"
              data-test="campaign-row"
              class="cursor-pointer"
              tabindex="0"
              @click="openCampaign(c)"
              @keydown.enter="openCampaign(c)"
            >
              <TableCell class="font-medium tabular-nums">
                #{{ c.id }}
              </TableCell>
              <TableCell class="max-w-md truncate">
                {{ c.description }}
              </TableCell>
              <TableCell>{{ languages(c) }}</TableCell>
              <TableCell>
                <Badge :variant="c.status === 'draft' ? 'outline' : 'default'">
                  {{ c.status === 'draft' ? 'Draft' : 'Approved' }}
                </Badge>
              </TableCell>
              <TableCell class="text-muted-foreground">
                {{ when(c.created_at) }}
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </div>

      <div class="space-y-3 md:hidden">
        <Card
          v-for="c in campaigns"
          :key="c.id"
          class="cursor-pointer py-4"
          tabindex="0"
          @click="openCampaign(c)"
          @keydown.enter="openCampaign(c)"
        >
          <CardContent class="space-y-2 px-4">
            <div class="flex items-center justify-between gap-2">
              <span class="font-medium tabular-nums">#{{ c.id }}</span>
              <Badge :variant="c.status === 'draft' ? 'outline' : 'default'">
                {{ c.status === 'draft' ? 'Draft' : 'Approved' }}
              </Badge>
            </div>
            <p class="text-sm break-words">
              {{ c.description }}
            </p>
            <p class="text-muted-foreground text-xs">
              {{ languages(c) }} · {{ when(c.created_at) }}
            </p>
          </CardContent>
        </Card>
      </div>
    </template>

    <ReviewDialog
      v-if="reviewId !== null"
      :open="reviewOpen"
      :campaign-id="reviewId"
      @update:open="onReviewOpen"
      @approved="onApproved"
    />
  </div>
</template>
