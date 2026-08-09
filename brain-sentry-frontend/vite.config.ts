/// <reference types="vitest/config" />
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import path from 'path';
import { gzipSync } from 'node:zlib';

const KIB = 1024;
const ENTRY_BUDGET_GZIP = 250 * KIB;
const ASYNC_CHUNK_BUDGET_GZIP = 190 * KIB;

function bundleBudget() {
  return {
    name: 'bundle-budget',
    generateBundle(_options: unknown, bundle: Record<string, { type: string; code?: string; isEntry?: boolean }>) {
      for (const [fileName, output] of Object.entries(bundle)) {
        if (output.type !== 'chunk' || !output.code) continue;

        const compressedSize = gzipSync(output.code).byteLength;
        const budget = output.isEntry ? ENTRY_BUDGET_GZIP : ASYNC_CHUNK_BUDGET_GZIP;
        if (compressedSize > budget) {
          throw new Error(
            `${fileName} is ${(compressedSize / KIB).toFixed(1)} KiB gzip; budget is ${budget / KIB} KiB`,
          );
        }
      }
    },
  };
}

export default defineConfig({
  plugins: [react(), bundleBudget()],
  build: {
    // The gzip budget above is the release gate; this raw-size limit only
    // suppresses Vite's less representative default warning.
    chunkSizeWarningLimit: 700,
  },
  test: {
    environment: 'jsdom',
    include: ['src/**/*.test.{ts,tsx}'],
    exclude: ['e2e/**'],
    setupFiles: ['./src/test/setup.ts'],
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json-summary', 'html'],
      reportsDirectory: './coverage',
      include: ['src/**/*.{ts,tsx}'],
      exclude: ['src/**/*.d.ts', 'src/**/*.test.{ts,tsx}', 'src/test/**'],
      thresholds: {
        statements: 1,
        branches: 25,
        functions: 10,
        lines: 1,
      },
    },
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
      '@components': path.resolve(__dirname, './src/components'),
      '@pages': path.resolve(__dirname, './src/pages'),
      '@lib': path.resolve(__dirname, './src/lib'),
      '@hooks': path.resolve(__dirname, './src/hooks'),
      '@types': path.resolve(__dirname, './src/types'),
      '@styles': path.resolve(__dirname, './src/styles'),
    },
  },
  server: {
    port: 3000,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, '')
      },
    },
  },
});
