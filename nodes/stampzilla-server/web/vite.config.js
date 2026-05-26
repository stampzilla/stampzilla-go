import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import basicSsl from '@vitejs/plugin-basic-ssl';

export default defineConfig(({ mode }) => {
  return {
    plugins: [
      react(),
      basicSsl(),
    ],
    esbuild: {
      loader: 'jsx',
      include: /src\/.*\.jsx?$/,
      exclude: [],
    },
    optimizeDeps: {
      esbuildOptions: {
        loader: {
          '.js': 'jsx',
        },
      },
    },
    resolve: {
      alias: [
        { find: /^~/, replacement: '' },
      ],
    },
    server: {
      host: '0.0.0.0',
      port: 8085,
      proxy: {
        '/ws': {
          target: 'https://localhost:6443',
          secure: false,
          ws: true,
        },
        '/register': {
          target: 'https://localhost:6443',
          secure: false,
        },
        '/login': {
          target: 'https://localhost:6443',
          secure: false,
        },
        '/logout': {
          target: 'https://localhost:6443',
          secure: false,
        },
        '/cert': {
          target: 'https://localhost:6443',
          secure: false,
        },
      },
    },
    define: {
      NODE_ENV: JSON.stringify(mode === 'production' ? 'production' : 'development'),
    },
  };
});
