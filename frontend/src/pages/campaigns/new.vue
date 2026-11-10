<script setup lang="ts">
import BriefAudience from '@/components/campaign/brief-audience.vue'
</script>

<template>
  <div class="space-y-6">
    <div class="space-y-2">
      <h1 class="text-2xl font-bold tracking-tight">
        New campaign
      </h1>
      <p class="text-muted-foreground">
        Write a brief. Claude finds the audience and writes the messages.
      </p>
    </div>
    <BriefAudience />
  </div>
</template>
