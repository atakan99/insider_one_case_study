// @vitest-environment happy-dom

import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import ContentView from '../content-view.vue'
import DeliveryBar from '../delivery-bar.vue'
import MessagePreview from '../message-preview.vue'

const content = { sms: 'Hi {{first_name}}, welcome!', email_subject: '{{first_name}}, 10% off', email_body: 'Hi {{first_name}},\n\n<b>Body</b>' }

describe('content view', () => {
  it('shows {{first_name}} as a chip and keeps the text', () => {
    const w = mount(ContentView, { props: { content } })
    expect(w.find('[data-field="sms"]').text()).toBe('Hi {{first_name}}, welcome!')
    expect(w.find('[data-field="sms"] [data-var]').text()).toBe('{{first_name}}')
    expect(w.findAll('[data-var]').length).toBe(3)
    expect(w.find('[data-field="body"]').text()).toContain('<b>Body</b>')
    expect(w.find('b').exists()).toBe(false)
  })

  it('shows the lengths only when given, and marks a length over the limit', () => {
    const none = mount(ContentView, { props: { content } })
    expect(none.find('[data-count]').exists()).toBe(false)
    const w = mount(ContentView, { props: { content, smsLength: 161, subjectLength: 44 } })
    expect(w.find('[data-count="sms"]').text()).toBe('161 / 160')
    expect(w.find('[data-count="sms"]').classes()).toContain('text-destructive')
    expect(w.find('[data-count="subject"]').text()).toBe('Subject 44 / 80')
    expect(w.find('[data-count="subject"]').classes()).not.toContain('text-destructive')
  })

  it('the preview fills in a name when it gets one', () => {
    const w = mount(MessagePreview, { props: { channel: 'sms', content, name: 'Ayşe' } })
    expect(w.find('[data-test="sms-preview"]').text()).toBe('Hi Ayşe, welcome!')
    expect(w.find('[data-var]').exists()).toBe(false)
  })
})

describe('delivery bar', () => {
  const zero = { total: 0, queued: 0, sending: 0, retrying: 0, sent: 0, failed: 0, unknown: 0, skipped: 0, retryable_failed: 0 }

  it('draws one segment for each result that has messages', () => {
    const w = mount(DeliveryBar, { props: { progress: { ...zero, total: 200, sent: 150, failed: 10, queued: 40 } } })
    expect(w.findAll('[data-segment]').map(s => s.attributes('data-segment'))).toEqual(['sent', 'failed'])
    expect(w.find('[data-segment="sent"]').attributes('style')).toContain('width: 75%')
    expect(w.attributes('aria-label')).toBe('150 sent, 10 failed, 0 unknown, 0 skipped of 200')
  })

  it('draws no segment for an empty channel', () => {
    const w = mount(DeliveryBar, { props: { progress: zero } })
    expect(w.findAll('[data-segment]').length).toBe(0)
  })
})
