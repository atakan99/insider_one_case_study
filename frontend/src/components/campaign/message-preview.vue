<script setup lang="ts">
import { computed } from 'vue'

import type { Content } from '@/lib/content'

const props = defineProps<{ channel: 'sms' | 'email', content: Content, name: string }>()

const VAR = '{{first_name}}'
const render = (s: string) => s.split(VAR).join(props.name)

const sms = computed(() => render(props.content.sms))
const subject = computed(() => render(props.content.email_subject))
const body = computed(() => render(props.content.email_body))
</script>

<template>
  <div v-if="channel === 'sms'" class="bg-muted min-h-56 rounded-2xl border p-4">
    <p class="text-muted-foreground mb-3 text-center text-xs">
      Your brand · now
    </p>
    <p data-test="sms-preview" class="bg-secondary text-secondary-foreground max-w-[85%] rounded-2xl rounded-bl-sm px-3 py-2.5 text-sm leading-snug break-words whitespace-pre-wrap">
      {{ sms }}
    </p>
  </div>
  <div v-else data-test="email-preview" class="overflow-hidden rounded-lg border">
    <div class="bg-muted border-b px-4 py-3">
      <p class="text-muted-foreground text-xs">
        From: Your brand
      </p>
      <p class="mt-1 font-medium break-words">
        {{ subject }}
      </p>
    </div>
    <p class="px-4 py-4 text-sm leading-relaxed break-words whitespace-pre-wrap">
      {{ body }}
    </p>
  </div>
</template>
