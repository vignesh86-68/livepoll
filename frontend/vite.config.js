import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// During development, the Vite dev server runs on :5173 and proxies API/WS
// requests to the Go backend on :8080. In production, the Go binary serves
// the built React app directly — no proxy needed.
export default defineConfig({
  plugins: [react()],
  build: {
    outDir: 'dist',
  },
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/ws': {
        target: 'ws://localhost:8080',
        ws: true,
      },
    },
  },
})
