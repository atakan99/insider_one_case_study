<script setup lang="ts">
import { useMutation } from '@tanstack/vue-query'
import { computed, ref } from 'vue'

import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Textarea } from '@/components/ui/textarea'
import { apiErrorMessage, previewSegment } from '@/lib/segments'

import ContentCards from './content-cards.vue'

const MAX_BRIEF = 1000

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
      <div class="flex items-center justify-between">
        <span class="text-sm" :class="length > MAX_BRIEF ? 'text-destructive' : 'text-muted-foreground'">
          {{ length }} / {{ MAX_BRIEF }}
        </span>
        <Button :disabled="!canSubmit || pending || running" @click="submit">
          Find audience
        </Button>
      </div>
    </div>

    <p v-if="pending" class="text-sm text-muted-foreground" role="status">
      Generating segment…
    </p>

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
              Email
            </dt>
            <dd class="text-2xl font-semibold">
              {{ result.audience.email }}
            </dd>
          </div>
          <div>
            <dt class="text-sm text-muted-foreground">
              Reachable
            </dt>
            <dd class="text-2xl font-semibold">
              {{ result.audience.reachable }}
            </dd>
          </div>
          <div>
            <dt class="text-sm text-muted-foreground">
              SMS
            </dt>
            <dd class="text-2xl font-semibold">
              {{ result.audience.sms }}
            </dd>
          </div>
        </dl>
        <p class="text-sm text-muted-foreground">
          <Badge variant="secondary">
            excluded: no consent
          </Badge>
          {{ result.audience.no_consent }}
          of {{ result.audience.matched }} matched
        </p>

        <table v-if="result.audience.languages.length" class="w-full text-sm">
          <thead>
            <tr class="text-left text-muted-foreground">
              <th class="py-1 font-medium">
                Language
              </th>
              <th class="py-1 font-medium">
                Email
              </th>
              <th class="py-1 font-medium">
                SMS
              </th>
              <th class="py-1 font-medium">
                Excluded: no consent
              </th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="lang in result.audience.languages" :key="lang.code">
              <td class="py-1">
                {{ lang.code }}
              </td>
              <td class="py-1">
                {{ lang.email }}
              </td>
              <td class="py-1">
                {{ lang.sms }}
              </td>
              <td class="py-1">
                {{ lang.no_consent }}
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
