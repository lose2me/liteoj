import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { config as mdConfig } from 'md-editor-v3'
import katex from 'katex'
import 'katex/dist/katex.min.css'
import router from './router'
import App from './App.vue'
import 'virtual:uno.css'
import './styles/markdown-overrides.css'
import './styles/data-table-overrides.css'

// md-editor-v3 lazy-loads KaTeX from a CDN by default. On slow / blocked
// networks (common in mainland China deployments) the script can still be
// in flight when the description renders, so formulas show as raw "$...$"
// until the user scrolls or interacts. Bundling katex locally and handing
// md-editor-v3 the instance makes rendering synchronous and fully offline.
//
// The highlight override below intercepts fenced blocks that declare no
// language so highlight.js's auto-detect can't classify them — e.g.
// "-5 7" as a diff "removed" line, which used to render in red. Labeled
// blocks still flow through hljs unchanged.
mdConfig({
  editorExtensions: {
    katex: { instance: katex },
  },
  markdownItConfig(md) {
    md.set({
      highlight(str, lang) {
        if (lang) return ''
        return `<pre><code class="language-">${md.utils.escapeHtml(str)}</code></pre>`
      },
    })
  },
})

const app = createApp(App)
app.use(createPinia())
app.use(router)
app.mount('#app')
