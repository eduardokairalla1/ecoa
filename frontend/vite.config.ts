/**
 * Vite configuration file.
 */

// --- IMPORTS ---
import tailwindcss from '@tailwindcss/vite'
import react from '@vitejs/plugin-react'
import path from 'node:path'
import { defineConfig } from 'vite'

// --- CODE ---
export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: { '@': path.resolve(import.meta.dirname, 'src') },
  },
  // the API answers on the same origin in dev, so the session cookies
  // travel without any cross-origin setup
  server: {
    proxy: { '/api': 'http://localhost:8000' },
  },
})
