<script setup lang="ts">
import type { CampaignStatus } from '@/lib/campaigns'

import { Badge } from '@/components/ui/badge'

defineProps<{ status: CampaignStatus }>()

const LABELS: Record<CampaignStatus, string> = {
  draft: 'Draft',
  approved: 'Approved',
  queued: 'Queued',
  sending: 'Sending',
  completed: 'Completed',
}

const CLASSES: Record<CampaignStatus, string> = {
  draft: '',
  approved: '',
  queued: 'bg-secondary text-secondary-foreground border-transparent',
  sending: 'bg-info/15 text-info border-transparent',
  completed: 'bg-success/15 text-success border-transparent',
}
</script>

<template>
  <Badge :variant="status === 'draft' ? 'outline' : 'default'" :class="CLASSES[status]" :data-status="status">
    {{ LABELS[status] }}
  </Badge>
</template>
