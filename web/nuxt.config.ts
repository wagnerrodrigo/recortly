import tailwindcss from '@tailwindcss/vite'

export default defineNuxtConfig({
  compatibilityDate: '2025-07-15',
  devtools: { enabled: false },
  css: ['~/assets/css/main.css'],
  vite: {
    plugins: [tailwindcss()],
  },
  app: {
    head: {
      title: 'Recortly',
      htmlAttrs: { lang: 'pt-BR' },
      meta: [
        { name: 'viewport', content: 'width=device-width, initial-scale=1' },
        { name: 'description', content: 'Cortes verticais e publicação para YouTube, Instagram e TikTok.' },
      ],
    },
  },
  runtimeConfig: {
    apiProxyTarget: process.env.NUXT_API_PROXY_TARGET || 'http://localhost:8080',
  },
})
