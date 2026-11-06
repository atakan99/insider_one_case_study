import { MegaphoneIcon, SquarePenIcon } from '@lucide/vue'

import type { NavGroup, NavItem } from '@/components/app-sidebar/types'

type SidebarLink = Extract<NavItem, { url: string }>

export const navData: NavGroup[] = [
  {
    title: 'Campaigns',
    items: [
      { title: 'New campaign', url: '/campaigns/new', icon: SquarePenIcon },
      { title: 'Campaigns', url: '/campaigns', icon: MegaphoneIcon },
    ],
  },
]

/** The pages that the page search (command menu) lists. */
export const pageCommands: SidebarLink[] = navData
  .flatMap(group => group.items)
  .filter((item): item is SidebarLink => typeof item.url === 'string')
