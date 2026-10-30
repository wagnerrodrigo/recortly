export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig()
  const target = (config.apiProxyTarget as string).replace(/\/$/, '')
  const path = event.path || ''
  await proxyRequest(event, `${target}${path}`)
})
