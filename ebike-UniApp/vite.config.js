import { defineConfig } from 'vite'
import uni from '@dcloudio/vite-plugin-uni'
import path from 'node:path'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [uni()],
  resolve: {
    alias: {
      // Full build includes message compiler so `{n}` / `{time}` interpolate on MP.
      'vue-i18n': path.resolve(
        process.cwd(),
        'node_modules/vue-i18n/dist/vue-i18n.esm-bundler.js',
      ),
    },
  },
  define: {
    __VUE_I18N_FULL_INSTALL__: true,
    __VUE_I18N_LEGACY_API__: false,
    __INTLIFY_PROD_DEVTOOLS__: false,
  },
  css: {
    preprocessorOptions: {
      scss: {
        // Use Dart Sass modern API to avoid legacy-js-api deprecation noise
        api: 'modern-compiler',
        silenceDeprecations: ['legacy-js-api', 'import'],
      },
    },
  },
})
