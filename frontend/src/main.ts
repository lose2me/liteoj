import { loader as monacoLoader } from '@guolao/vue-monaco-editor'
import Cropper from 'cropperjs'
import * as echarts from 'echarts'
import hljs from 'highlight.js'
import katex from 'katex'
import { config as mdConfig } from 'md-editor-v3'
import mermaid from 'mermaid'
import * as monaco from 'monaco-editor'
import editorWorker from 'monaco-editor/esm/vs/editor/editor.worker?worker'
import cssWorker from 'monaco-editor/esm/vs/language/css/css.worker?worker'
import htmlWorker from 'monaco-editor/esm/vs/language/html/html.worker?worker'
import jsonWorker from 'monaco-editor/esm/vs/language/json/json.worker?worker'
import tsWorker from 'monaco-editor/esm/vs/language/typescript/ts.worker?worker'
import { createPinia } from 'pinia'
import prettier from 'prettier/standalone'
import prettierMarkdown from 'prettier/plugins/markdown'
import screenfull from 'screenfull'
import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import 'cropperjs/dist/cropper.css'
import 'katex/dist/katex.min.css'
import 'virtual:uno.css'
import './styles/theme-tokens.css'
import './styles/markdown-overrides.css'
import './styles/data-table-overrides.css'

const monacoRuntime = self as typeof globalThis & {
  MonacoEnvironment?: {
    getWorker: (_workerId: string, label: string) => Worker
  }
}

// vue-monaco-editor uses @monaco-editor/loader, whose default source is a CDN.
// Force it onto the local npm package and Vite-managed workers instead.
monacoRuntime.MonacoEnvironment = {
  getWorker(_workerId, label) {
    if (label === 'json') return new jsonWorker()
    if (label === 'css' || label === 'scss' || label === 'less') return new cssWorker()
    if (label === 'html' || label === 'handlebars' || label === 'razor') return new htmlWorker()
    if (label === 'typescript' || label === 'javascript') return new tsWorker()
    return new editorWorker()
  },
}
monacoLoader.config({ monaco })

// md-editor-v3 otherwise injects its helper libraries from unpkg/CDNs during
// runtime, including KaTeX, highlight.js, cropper, screenfull and mermaid.
// Supplying local npm instances keeps markdown edit/preview fully intranet-safe.
mdConfig({
  editorExtensions: {
    highlight: { instance: hljs },
    prettier: {
      prettierInstance: prettier,
      parserMarkdownInstance: prettierMarkdown,
    },
    cropper: { instance: Cropper },
    screenfull: { instance: screenfull },
    mermaid: { instance: mermaid },
    katex: { instance: katex },
    echarts: { instance: echarts },
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
