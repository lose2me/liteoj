import { zh } from './zh'

// `t` is an alias for the Chinese message pack. We only ship one locale, so
// there's no runtime lookup — property access gives full TS autocomplete.
export const t = zh
export type { Messages } from './zh'
