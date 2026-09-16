/** App uses Amap. Mini programs omit provider (built-in map). */
export function mapProvider(): string | undefined {
  // #ifdef APP-PLUS
  return 'amap'
  // #endif
  // #ifndef APP-PLUS
  return undefined
  // #endif
}
