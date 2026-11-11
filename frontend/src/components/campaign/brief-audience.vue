<script setup lang="ts">
import { useMutation } from '@tanstack/vue-query'
import { computed, ref } from 'vue'

import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { Textarea } from '@/components/ui/textarea'
import { languageName } from '@/lib/campaigns'
import { formatCount } from '@/lib/format'
import { apiErrorMessage, previewSegment } from '@/lib/segments'

import ContentCards from './content-cards.vue'

const MAX_BRIEF = 1000

// Each brief uses only fields that the segment allowlist has.
const SUGGESTIONS = [
  { label: 'Cart reminder', text: 'Remind customers in Istanbul who added to cart in the last 30 days but did not buy to complete their order with 30% off. Friendly and short.' },
  { label: 'Loyal customers', text: 'Thank customers who spent over 5000 and ordered more than 5 times. Offer early access to the new collection.' },
  { label: 'Win-back', text: 'Win back customers who have not bought in 90 days. Offer free shipping on the next order.' },
  { label: 'Welcome', text: 'Welcome customers who signed up in the last 14 days and have no orders yet. Offer 10% off a first order.' },
]

const brief = ref('')
const length = computed(() => brief.value.length)
const canSubmit = computed(() => brief.value.trim() !== '' && length.value <= MAX_BRIEF)

const mutation = useMutation({ mutationFn: previewSegment, meta: { silent: true } })
const pending = computed(() => mutation.isPending.value)
const result = computed(() => (mutation.isError.value ? undefined : mutation.data.value))
const errorMessage = computed(() => (mutation.isError.value ? apiErrorMessage(mutation.error.value) : ''))

const runId = ref(0)
const running = ref(false)
const contentBrief = ref('')
const contentRule = ref<unknown>(null)

function submit() {
  if (!canSubmit.value || pending.value || running.value)
    return
  const sent = brief.value
  // Unmounts the cards, which aborts any open content request.
  runId.value = 0
  running.value = false
  mutation.mutate(sent, {
    onSuccess: (data) => {
      if (data.audience.languages.length === 0)
        return
      contentBrief.value = sent
      contentRule.value = data.rule
      runId.value++
    },
  })
}
</script>

<template>
  <div class="space-y-6">
    <div class="space-y-2">
      <Textarea
        v-model="brief"
        class="min-h-32"
        placeholder="Example: Send a winter sale to customers in Turkey who bought in the last 90 days."
        aria-label="Campaign brief"
      />
      <div class="flex flex-wrap items-center gap-2" role="group" aria-label="Suggested briefs">
        <span class="text-muted-foreground text-sm">Try an example:</span>
        <Button
          v-for="s in SUGGESTIONS"
          :key="s.label"
          type="button"
          variant="outline"
          size="sm"
          class="rounded-full font-normal"
          :title="s.text"
          :disabled="pending || running"
          @click="brief = s.text"
        >
          {{ s.label }}
        </Button>
      </div>
      <div class="flex items-center justify-between">
        <span class="text-sm" :class="length > MAX_BRIEF ? 'text-destructive' : 'text-muted-foreground'">
          {{ length }} / {{ MAX_BRIEF }}
        </span>
        <Button data-submit :disabled="!canSubmit || pending || running" @click="submit">
          Find audience
        </Button>
      </div>
    </div>

    <div v-if="pending" class="space-y-3" role="status">
      <p class="text-sm text-muted-foreground">
        Generating segment…
      </p>
      <Skeleton class="h-40 w-full" />
    </div>

    <Alert v-if="errorMessage" variant="destructive">
      <AlertDescription>{{ errorMessage }}</AlertDescription>
    </Alert>

    <Card v-if="result && !pending">
      <CardHeader>
        <CardTitle>Audience</CardTitle>
        <p class="text-sm text-muted-foreground">
          {{ result.description }}
        </p>
      </CardHeader>
      <CardContent class="space-y-4">
        <dl class="grid grid-cols-3 gap-4">
          <div>
            <dt class="text-sm text-muted-foreground">
              Reachable
            </dt>
            <dd class="text-2xl font-semibold tabular-nums">
              {{ formatCount(result.audience.reachable) }}
            </dd>
          </div>
          <div>
            <dt class="text-sm text-muted-foreground">
              SMS
            </dt>
            <dd class="text-2xl font-semibold tabular-nums">
              {{ formatCount(result.audience.sms) }}
            </dd>
          </div>
          <div>
            <dt class="text-sm text-muted-foreground">
              Email
            </dt>
            <dd class="text-2xl font-semibold tabular-nums">
              {{ formatCount(result.audience.email) }}
            </dd>
          </div>
        </dl>
        <p class="text-sm text-muted-foreground tabular-nums">
          <Badge variant="secondary">
            excluded: no consent
          </Badge>
          {{ formatCount(result.audience.no_consent) }}
          of {{ formatCount(result.audience.matched) }} matched
        </p>

        <table v-if="result.audience.languages.length" class="w-full text-sm tabular-nums">
          <thead>
            <tr class="border-b text-left text-muted-foreground">
              <th class="py-1.5 font-medium">
                Language
              </th>
              <th class="py-1.5 text-right font-medium">
                SMS
              </th>
              <th class="py-1.5 text-right font-medium">
                Email
              </th>
              <th class="py-1.5 text-right font-medium">
                Excluded: no consent
              </th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="lang in result.audience.languages" :key="lang.code">
              <td class="py-1.5">
                {{ languageName(lang.code) }}
              </td>
              <td class="py-1.5 text-right">
                {{ formatCount(lang.sms) }}
              </td>
              <td class="py-1.5 text-right">
                {{ formatCount(lang.email) }}
              </td>
              <td class="py-1.5 text-right">
                {{ formatCount(lang.no_consent) }}
              </td>
            </tr>
          </tbody>
        </table>

        <Alert v-if="result.unmapped.length">
          <AlertTitle>Not used from the brief</AlertTitle>
          <AlertDescription>
            <ul class="list-disc space-y-1 pl-5">
              <li v-for="(note, i) in result.unmapped" :key="i">
                {{ note }}
              </li>
            </ul>
          </AlertDescription>
        </Alert>
      </CardContent>
    </Card>

    <ContentCards
      v-if="runId > 0"
      :brief="contentBrief"
      :rule="contentRule"
      :run-id="runId"
      @update:running="running = $event"
    />
  </div>
</template>
