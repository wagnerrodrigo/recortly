<template>
  <div class="grid min-h-screen place-items-center px-4" style="background: radial-gradient(circle at 20% 10%, #fff1f7, #f7f7f8 55%)">
    <div class="w-full max-w-sm">
      <div class="mb-8 flex items-center gap-3">
        <div class="flex h-11 w-11 items-center justify-center rounded-xl text-white" style="background: var(--color-brand-500)">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" class="h-6 w-6">
            <path d="M4 6h16M4 12h10M4 18h7" stroke-linecap="round" />
          </svg>
        </div>
        <div>
          <p class="text-xl font-bold leading-tight">SmashCut</p>
          <p class="text-xs" style="color: var(--color-muted)">Cortes verticais, do bruto ao publicado</p>
        </div>
      </div>

      <form class="card space-y-4 p-6" @submit.prevent="submit">
        <h1 class="text-lg font-semibold">Entrar</h1>
        <div>
          <label class="label" for="email">E-mail</label>
          <input id="email" v-model="email" type="email" class="input" autocomplete="email" required />
        </div>
        <div>
          <label class="label" for="password">Senha</label>
          <input id="password" v-model="password" type="password" class="input" autocomplete="current-password" required />
        </div>
        <p v-if="error" class="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-700">{{ error }}</p>
        <button type="submit" class="btn btn-primary w-full" :disabled="loading">
          {{ loading ? 'Entrando...' : 'Entrar' }}
        </button>
        <p class="text-center text-xs" style="color: var(--color-muted)">
          Acesso restrito aos dois usuários configurados no servidor.
        </p>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'auth' })

const route = useRoute()
const session = useSession()
const email = ref('eu@smashcut.local')
const password = ref('')
const loading = ref(false)
const error = ref('')

async function submit() {
  error.value = ''
  loading.value = true
  try {
    await session.login(email.value, password.value)
    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : '/'
    await navigateTo(redirect)
  } catch (e) {
    error.value = errorMessage(e)
  } finally {
    loading.value = false
  }
}
</script>
