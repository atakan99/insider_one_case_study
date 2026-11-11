<script setup lang="ts">
import type { Content } from '@/lib/content'

import { LIMITS } from '@/lib/limits'

import MessagePreview from './message-preview.vue'

// Read-only view of one language: the SMS next to the email.
// The lengths are the lengths after the longest first name is filled in.
defineProps<{ content: Content, smsLength?: number, subjectLength?: number }>()

const HINT = 'Length with the longest first name in the audience'
</script>

<template>
  <div class="grid gap-6 lg:grid-cols-[minmax(0,22rem)_minmax(0,1fr)]">
    <section class="space-y-2">
      <div class="flex items-baseline justify-between gap-2">
        <h3 class="text-muted-foreground text-xs font-medium tracking-wide uppercase not-italic">
          SMS
        </h3>
        <span
          v-if="smsLength !== undefined"
          data-count="sms"
          :title="HINT"
          class="text-xs tabular-nums not-italic"
          :class="smsLength > LIMITS.sms ? 'text-destructive' : 'text-muted-foreground'"
        >{{ smsLength }} / {{ LIMITS.sms }}</span>
      </div>
      <MessagePreview channel="sms" :content="content" compact />
    </section>
    <section class="space-y-2">
      <div class="flex items-baseline justify-between gap-2">
        <h3 class="text-muted-foreground text-xs font-medium tracking-wide uppercase not-italic">
          Email
        </h3>
        <span
          v-if="subjectLength !== undefined"
          data-count="subject"
          :title="HINT"
          class="text-xs tabular-nums not-italic"
          :class="subjectLength > LIMITS.email_subject ? 'text-destructive' : 'text-muted-foreground'"
        >Subject {{ subjectLength }} / {{ LIMITS.email_subject }}</span>
      </div>
      <MessagePreview channel="email" :content="content" />
    </section>
  </div>
</template>
