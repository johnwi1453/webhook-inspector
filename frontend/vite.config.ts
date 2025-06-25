import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      '/status': 'http://localhost:8080',
      '/logs': 'http://localhost:8080',
      '/reset': 'http://localhost:8080',
      '/create': 'http://localhost:8080',
      '/api': 'http://localhost:8080',
      '/auth': 'http://localhost:8080',
      '/token': 'http://localhost:8080',
      '/me': 'http://localhost:8080',
      '/logout': 'http://localhost:8080',
    },
  },
})
