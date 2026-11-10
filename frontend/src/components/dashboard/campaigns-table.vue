<script setup lang="ts">
import { RouterLink } from 'vue-router'

import type { DashboardStats } from '@/lib/dashboard'

import StatusBadge from '@/components/campaign/status-badge.vue'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { formatDuration, percent } from '@/lib/dashboard'

defineProps<{ campaigns: DashboardStats['campaigns'] }>()
</script>

<template>
  <Card>
    <CardHeader>
      <CardTitle>Campaigns</CardTitle>
    </CardHeader>
    <CardContent>
      <p v-if="campaigns.length === 0" class="text-muted-foreground text-sm">
        No campaigns sent yet.
        <RouterLink to="/campaigns/new" class="text-foreground underline">
          Create a campaign
        </RouterLink>
      </p>
      <div v-else class="overflow-x-auto">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Campaign</TableHead>
              <TableHead>Status</TableHead>
              <TableHead class="text-right">
                Total
              </TableHead>
              <TableHead class="text-right">
                Sent
              </TableHead>
              <TableHead class="text-right">
                Failed
              </TableHead>
              <TableHead class="text-right">
                Success rate
              </TableHead>
              <TableHead class="text-right">
                Duration
              </TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="c in campaigns" :key="c.id">
              <TableCell class="max-w-64 truncate">
                <RouterLink :to="`/campaigns/${c.id}`" class="underline-offset-4 hover:underline">
                  {{ c.description }}
                </RouterLink>
              </TableCell>
              <TableCell>
                <StatusBadge :status="c.status" />
              </TableCell>
              <TableCell class="text-right tabular-nums">
                {{ c.total }}
              </TableCell>
              <TableCell class="text-right tabular-nums">
                {{ c.sent }}
              </TableCell>
              <TableCell class="text-right tabular-nums">
                {{ c.failed }}
              </TableCell>
              <TableCell class="text-right tabular-nums">
                {{ percent(c.sent, c.sent + c.failed + c.unknown) }}
              </TableCell>
              <TableCell class="text-right tabular-nums">
                {{ formatDuration(c.duration_ms) }}
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </div>
    </CardContent>
  </Card>
</template>
