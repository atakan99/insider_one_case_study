<script setup lang="ts">
import { CommandGroup, CommandItem } from '@/components/ui/command'
import { pageCommands } from '@/constants/sidebar-data'

import CommandItemHasIcon from './command-item-has-icon.vue'

const emit = defineEmits<{
  click: []
}>()

const commands = pageCommands

const router = useRouter()
const route = useRoute()
function commandItemClick(url: string) {
  emit('click')
  if (route.fullPath !== url) {
    router.push(url)
  }
}
</script>

<template>
  <CommandGroup heading="Pages">
    <CommandItem
      v-for="command in commands"
      :key="command.title"
      :value="command.title"
      @click="commandItemClick(command.url)"
    >
      <CommandItemHasIcon :name="command.title" :icon="command.icon" />
    </CommandItem>
  </CommandGroup>
</template>
