import path from 'node:path'
import tailwindcss from '@tailwindcss/vite'
import react, { reactCompilerPreset } from '@vitejs/plugin-react'
import babel from '@rolldown/plugin-babel'
import { defineConfig } from 'vite'

export default defineConfig(({ mode }) => ({
  plugins: [
    react(),
    babel({ presets: [reactCompilerPreset()] }),
    tailwindcss(),
  ],
  base: mode === 'capacitor' ? './' : '/',
  resolve: {
    alias: {
      'react-native': 'react-native-web',
      '@': path.resolve(import.meta.dirname, 'src'),
    },
    extensions: ['.web.tsx', '.web.ts', '.tsx', '.ts', '.web.jsx', '.web.js', '.jsx', '.js'],
  },
  define: {
    __DEV__: JSON.stringify(process.env.NODE_ENV !== 'production'),
    global: 'globalThis',
  },
  optimizeDeps: {
    include: ['react-native-web', 'react', 'react-dom', 'motion', 'framer-motion'],
  },
  server: {
    port: 3017,
    host: '0.0.0.0',
    strictPort: true,
  },
  preview: {
    port: 3017,
    host: '0.0.0.0',
  },
  build: {
    sourcemap: false,
    target: 'es2022',
    assetsInlineLimit: 4096,
  },
}))

