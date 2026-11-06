import { describe, expect, it } from 'vitest'

import { navData, pageCommands } from '../sidebar-data'

describe('sidebar data', () => {
  it('has only the campaign links', () => {
    const links = navData
      .flatMap(group => group.items)
      .map(item => ({ title: item.title, url: item.url }))

    expect(links).toEqual([
      { title: 'New campaign', url: '/campaigns/new' },
      { title: 'Campaigns', url: '/campaigns' },
    ])
  })

  it('page search lists only the sidebar pages', () => {
    expect(pageCommands.map(command => command.url)).toEqual(['/campaigns/new', '/campaigns'])
  })
})
