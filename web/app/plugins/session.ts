import { defineNuxtPlugin } from '#app'

export default defineNuxtPlugin(async () => {
  const session = useSession()
  if (!session.value) {
    await session.refresh()
  }
})
