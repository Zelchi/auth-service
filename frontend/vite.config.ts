import { defineConfig } from 'vite'
import solid from 'vite-plugin-solid'

export default defineConfig({
    plugins: [solid()],
    server: {
        proxy: {
            '/api': {
                target: 'http://127.0.0.1:8888',
                changeOrigin: true,
                configure: proxy => {
                    proxy.on('proxyReq', proxyReq => {
                        if (proxyReq.getHeader('origin')) {
                            proxyReq.setHeader('origin', 'http://127.0.0.1:8888')
                        }
                    })
                },
            },
        }
    }
})
