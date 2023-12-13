import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react-swc'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  optimizeDeps: {
    exclude: ["@vite/client", "@vite/env", "@vlcn.io/crsqlite-wasm"],
    esbuildOptions: {
      target: "esnext",
    },
  },
})
