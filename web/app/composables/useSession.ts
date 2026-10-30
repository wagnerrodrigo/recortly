interface SessionUser {
  id: number
  name: string
  email: string
}

export const useSession = () => {
  const session = useState<SessionUser | null>('session-user', () => null)

  // Durante o SSR o cookie do navegador precisa ser repassado manualmente.
  const cookieHeaders = () => {
    if (import.meta.server) {
      return useRequestHeaders(['cookie'])
    }
    return {}
  }

  const refresh = async () => {
    try {
      const res = await $fetch<{ user: SessionUser }>('/api/auth/me', { headers: cookieHeaders() })
      session.value = res.user
    } catch {
      session.value = null
    }
    return session.value
  }

  const login = async (email: string, password: string) => {
    const res = await $fetch<{ user: SessionUser }>('/api/auth/login', {
      method: 'POST',
      body: { email, password },
    })
    session.value = res.user
  }

  const logout = async () => {
    await $fetch('/api/auth/logout', { method: 'POST' })
    session.value = null
    await navigateTo('/login')
  }

  return Object.assign(session, { refresh, login, logout })
}
