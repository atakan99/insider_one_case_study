import { defineComponent, h } from 'vue'

const VAR = '{{first_name}}'
const CHIP = 'bg-primary/10 text-primary rounded px-1 py-0.5 font-mono text-[0.85em] not-italic'

// Renders message text. With a name, each {{first_name}} becomes that name.
// Without a name, each {{first_name}} shows as a chip, so it is not read as copy.
export default defineComponent({
  name: 'VarText',
  props: {
    text: { type: String, required: true },
    name: { type: String, default: undefined },
  },
  setup(props) {
    return () => {
      const parts = props.text.split(VAR)
      if (props.name !== undefined)
        return parts.join(props.name)
      return parts.flatMap((part, i) => i === 0 ? [part] : [h('span', { 'class': CHIP, 'data-var': 'first_name' }, VAR), part])
    }
  },
})
