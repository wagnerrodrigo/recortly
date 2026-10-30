export default defineNuxtRouteMiddleware(async (to) => {
  const session = useSession()
  if (!session.value) {
    await session.refresh()
  }
  if (to.path === '/login') {
    if (session.value) return navigateTo('/')
    return
  }
  if (!session.value) {
    return navigateTo(`/login?redirect=${encodeURIComponent(to.fullPath)}`)
  }
})
