<script setup lang="ts">
const route = useRoute('/campaigns/[id]')
</script>

<template>
  <div class="space-y-2">
    <h1 class="text-2xl font-bold tracking-tight">
      Campaign {{ route.params.id }}
    </h1>
    <p class="text-muted-foreground">
      The live send dashboard comes in a later slice.
    </p>
  </div>
</template>
