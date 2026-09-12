declare module 'postcss-pxtorem' {
  import type { PluginCreator } from 'postcss'

  interface PxToRemOptions {
    rootValue?: number | ((input: { file?: string }) => number)
    unitPrecision?: number
    propList?: string[]
    selectorBlackList?: Array<string | RegExp>
    replace?: boolean
    mediaQuery?: boolean
    minPixelValue?: number
    exclude?: string | RegExp | ((file: string) => boolean)
  }

  const pxtorem: PluginCreator<PxToRemOptions>
  export default pxtorem
}
