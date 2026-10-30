export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig()
  const target = (config.apiProxyTarget as string).replace(/\/$/, '')
  const query = getQuery(event)
  const params = new URLSearchParams()
  for (const [k, v] of Object.entries(query)) {
    if (typeof v === 'string') params.set(k, v)
  }
  const res = await fetch(`${target}/api/oauth/youtube/callback?${params.toString()}`, {
    method: 'GET',
    redirect: 'manual',
  })
  const location = res.headers.get('location')
  if (location) {
    await sendRedirect(event, location.startsWith('http') ? '/integracoes?connected=youtube' : location, 302)
    return
  }
  setResponseStatus(event, res.status)
  return await res.text()
})
