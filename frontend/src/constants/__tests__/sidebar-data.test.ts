import { describe, expect, it } from 'vitest'

import { navData, pageCommands } from '../sidebar-data'

describe('sidebar data', () => {
  it('has the dashboard and campaign links', () => {
    const links = navData
      .flatMap(group => group.items)
      .map(item => ({ title: item.title, url: item.url }))

    expect(links).toEqual([
      { title: 'Dashboard', url: '/dashboard' },
      { title: 'New campaign', url: '/campaigns/new' },
      { title: 'Campaigns', url: '/campaigns' },
    ])
  })

  it('page search lists only the sidebar pages', () => {
    expect(pageCommands.map(command => command.url)).toEqual(['/dashboard', '/campaigns/new', '/campaigns'])
  })
})
