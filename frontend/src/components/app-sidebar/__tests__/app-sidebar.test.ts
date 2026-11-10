// @vitest-environment happy-dom

import { mount } from '@vue/test-utils'
import { createPinia } from 'pinia'
import { describe, expect, it } from 'vitest'
import { defineComponent, h } from 'vue'
import { createMemoryHistory, createRouter } from 'vue-router'

import { SidebarProvider } from '@/components/ui/sidebar'

import AppSidebar from '../index.vue'

describe('app sidebar', () => {
  it('shows the app name and no team or user menu', async () => {
    const router = createRouter({
      history: createMemoryHistory(),
      routes: [{ path: '/:path(.*)*', component: { render: () => null } }],
    })
    await router.push('/campaigns/new')

    const Wrapper = defineComponent({
      render: () => h(SidebarProvider, null, () => h(AppSidebar)),
    })
    const wrapper = mount(Wrapper, { global: { plugins: [router, createPinia()] } })
    const text = wrapper.text()

    expect(text).toContain('Campaign Composer')
    const first = wrapper.find('[data-sidebar="content"] a')
    expect(first.attributes('href')).toBe('/dashboard')
    expect(first.text()).toContain('Dashboard')
    expect(wrapper.find('[data-sidebar="content"] a[href="/campaigns/new"]').text()).toContain('New campaign')
    expect(wrapper.find('[data-sidebar="content"] a[href="/campaigns"]').text()).toContain('Campaigns')
    // No team switcher (a dropdown in the header) and no user menu (the footer).
    expect(wrapper.find('[data-sidebar="header"] [aria-haspopup]').exists()).toBe(false)
    expect(wrapper.find('[data-sidebar="footer"]').exists()).toBe(false)
    expect(text).not.toContain('Acme')
    expect(text).not.toContain('shadcn')
    expect(text).not.toContain('m@example.com')
  })
})
