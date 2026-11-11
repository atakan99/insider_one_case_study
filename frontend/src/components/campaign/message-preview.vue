<script setup lang="ts">
import type { Content } from '@/lib/content'

import VarText from './var-text'

// Without a name, the preview keeps {{first_name}} and shows it as a chip.
defineProps<{ channel: 'sms' | 'email', content: Content, name?: string, compact?: boolean }>()
</script>

<template>
  <div v-if="channel === 'sms'" class="bg-muted rounded-2xl border p-4" :class="compact ? '' : 'min-h-56'">
    <p class="text-muted-foreground mb-3 text-center text-xs">
      Your brand · now
    </p>
    <p data-test="sms-preview" data-field="sms" class="bg-secondary text-secondary-foreground max-w-[85%] rounded-2xl rounded-bl-sm px-3 py-2.5 text-sm leading-snug break-words whitespace-pre-wrap">
      <VarText :text="content.sms" :name="name" />
    </p>
  </div>
  <div v-else data-test="email-preview" class="overflow-hidden rounded-lg border">
    <div class="bg-muted border-b px-4 py-3">
      <p class="text-muted-foreground text-xs">
        From: Your brand
      </p>
      <p data-field="subject" class="mt-1 font-medium break-words">
        <VarText :text="content.email_subject" :name="name" />
      </p>
    </div>
    <p data-field="body" class="max-w-prose px-4 py-4 text-sm leading-relaxed break-words whitespace-pre-wrap">
      <VarText :text="content.email_body" :name="name" />
    </p>
  </div>
</template>
