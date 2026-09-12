import { fileURLToPath, URL } from 'node:url'
import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'
import Components from 'unplugin-vue-components/vite'
import { VantResolver } from '@vant/auto-import-resolver'
import autoprefixer from 'autoprefixer'
import pxtorem from 'postcss-pxtorem'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), 'VITE_')

  return {
    // 与遗留部署路径一致，App WebView 打开的是 /mop-saas/#/...
    base: env.VITE_BASE_PATH || '/mop-saas/',
    plugins: [
      vue(),
      Components({ resolvers: [VantResolver()], dts: 'src/components.d.ts' }),
    ],
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url)),
      },
    },
    css: {
      postcss: {
        // 设计稿 375，与 amfe-flexible 的 1rem = 屏宽/10 对应；Vant 4 同为 375 基准
        plugins: [autoprefixer(), pxtorem({ rootValue: 37.5, propList: ['*'] })],
      },
    },
    server: {
      host: true,
      port: 5180,
    },
    build: {
      sourcemap: false,
      chunkSizeWarningLimit: 1200,
      rollupOptions: {
        output: {
          // Rollup 5 起 manualChunks 只接受函数形式。ECharts 单独成块，
          // 避免它和业务代码一起失效缓存。
          manualChunks(id) {
            if (id.includes('node_modules/echarts') || id.includes('node_modules/zrender')) {
              return 'echarts'
            }
            if (/node_modules\/(vue|vue-router|pinia|@intlify|vue-i18n)\//.test(id)) {
              return 'vendor'
            }
            return null
          },
        },
      },
    },
  }
})
