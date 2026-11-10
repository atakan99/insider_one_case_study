<script setup lang="ts">
import type { DashboardStats } from '@/lib/dashboard'

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'

defineProps<{ failures: DashboardStats['failures'], meanAttempts: DashboardStats['mean_attempts'] }>()

function attempts(v: number | null): string {
  return v === null ? '–' : v.toFixed(2)
}
</script>

<template>
  <Card>
    <CardHeader>
      <CardTitle>Failures</CardTitle>
    </CardHeader>
    <CardContent class="space-y-4">
      <div class="flex gap-6 text-sm">
        <span data-test="mean-sms">Mean attempts SMS: <strong class="tabular-nums">{{ attempts(meanAttempts.sms) }}</strong></span>
        <span data-test="mean-email">Mean attempts email: <strong class="tabular-nums">{{ attempts(meanAttempts.email) }}</strong></span>
      </div>
      <p v-if="failures.length === 0" class="text-muted-foreground text-sm">
        No failures.
      </p>
      <div v-else class="overflow-x-auto">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Status</TableHead>
              <TableHead>Reason</TableHead>
              <TableHead class="text-right">
                SMS
              </TableHead>
              <TableHead class="text-right">
                Email
              </TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="f in failures" :key="`${f.status}-${f.reason}`" data-test="failure-row">
              <TableCell class="capitalize">
                {{ f.status }}
              </TableCell>
              <TableCell>{{ f.reason }}</TableCell>
              <TableCell class="text-right tabular-nums">
                {{ f.sms }}
              </TableCell>
              <TableCell class="text-right tabular-nums">
                {{ f.email }}
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </div>
    </CardContent>
  </Card>
</template>
