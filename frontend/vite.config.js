import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'https://localhost:8080',
        changeOrigin: true,
        secure: false,
        ws: true, // ← THIS WAS MISSING. Forwards WebSocket upgrades.
        configure: (proxy) => {
          // Browser can't set Authorization on a WebSocket. We send the
          // token as ?token=…, and Vite's proxy rewrites it into a Bearer
          // header before the upgrade reaches Go.
          proxy.on('proxyReqWs', (proxyReq, req) => {
            try {
              const url = new URL(req.url, 'http://localhost')
              const token = url.searchParams.get('token')
              if (token) {
                proxyReq.setHeader('Authorization', `Bearer ${token}`)
              }
            } catch {
              // ignore malformed URLs
            }
          })
        },
      },
    },
  },
})