import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  clearScreen: false,
  server: {
    port: 5173,
    strictPort: true,
    host: '0.0.0.0', // Listen on all interfaces so Wails IPv4 proxy can connect
  },
  envPrefix: ['VITE_', 'WAILS_'],
  build: {
    outDir: 'dist',
    emptyOutDir: true,
  },
})
