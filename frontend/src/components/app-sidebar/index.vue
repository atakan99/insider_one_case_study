<script setup lang="ts">
import { MegaphoneIcon } from '@lucide/vue'

import { Sidebar, SidebarContent, SidebarHeader, SidebarMenu, SidebarMenuButton, SidebarMenuItem, SidebarRail } from '@/components/ui/sidebar'

import { sidebarData } from './data/sidebar-data'
import NavTeam from './nav-team.vue'
</script>

<template>
  <Sidebar collapsible="icon" class="z-50">
    <SidebarHeader class="border-b h-14 py-1 group-data-[collapsible=icon]:justify-center">
      <SidebarMenu>
        <SidebarMenuItem>
          <SidebarMenuButton size="lg" as-child>
            <router-link to="/campaigns/new">
              <div class="flex aspect-square size-8 items-center justify-center rounded-lg bg-sidebar-primary text-sidebar-primary-foreground">
                <MegaphoneIcon class="size-4" />
              </div>
              <span class="truncate font-semibold">Campaign Composer</span>
            </router-link>
          </SidebarMenuButton>
        </SidebarMenuItem>
      </SidebarMenu>
    </SidebarHeader>

    <SidebarContent>
      <NavTeam :nav-main="sidebarData.navMain" />
    </SidebarContent>

    <SidebarRail />
  </Sidebar>
</template>
