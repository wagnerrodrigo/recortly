<template>
  <div class="flex min-h-screen">
    <aside
      class="fixed inset-y-0 left-0 z-20 flex w-[220px] flex-col border-r bg-white transition-transform lg:static lg:translate-x-0"
      :class="open ? 'translate-x-0' : '-translate-x-full'"
      style="border-color: var(--color-line)"
    >
      <div class="flex h-16 items-center gap-2 px-5">
        <div class="flex h-9 w-9 items-center justify-center rounded-[10px] text-white" style="background: var(--color-brand-500)">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" class="h-5 w-5">
            <path d="M4 6h16M4 12h10M4 18h7" stroke-linecap="round" />
          </svg>
        </div>
        <span class="text-lg font-bold tracking-tight">SmashCut</span>
      </div>

      <p class="px-5 pb-2 pt-2 text-[11px] font-semibold uppercase tracking-widest" style="color: var(--color-muted)">Geral</p>
      <nav class="flex-1 space-y-1 overflow-y-auto px-3">
        <NuxtLink v-for="item in nav" :key="item.to" :to="item.to" class="nav-item" active-class="nav-active">
          <span v-html="item.icon" class="h-[18px] w-[18px]" />
          <span>{{ item.label }}</span>
        </NuxtLink>
      </nav>

      <div class="m-3 rounded-xl p-3" style="background: var(--color-brand-50)">
        <p class="text-xs font-semibold">Uso pessoal</p>
        <p class="mt-1 text-[11px] leading-snug" style="color: var(--color-muted)">
          Publique apenas conteúdo próprio ou com permissão do criador.
        </p>
      </div>
    </aside>

    <div v-if="open" class="fixed inset-0 z-10 bg-black/30 lg:hidden" @click="open = false" />

    <div class="flex min-w-0 flex-1 flex-col">
      <header class="sticky top-0 z-10 flex h-16 items-center gap-3 border-b bg-white/90 px-4 backdrop-blur" style="border-color: var(--color-line)">
        <button class="btn btn-ghost h-9 w-9 px-0 lg:hidden" @click="open = !open" aria-label="Menu">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="h-5 w-5"><path d="M4 7h16M4 12h16M4 17h16" /></svg>
        </button>
        <div class="relative w-full max-w-sm">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2" style="color: var(--color-muted)">
            <circle cx="11" cy="11" r="7" /><path d="m20 20-3.5-3.5" />
          </svg>
          <input v-model="query" class="input pl-9" placeholder="Buscar na biblioteca..." @keyup.enter="search" />
        </div>
        <div class="flex-1" />
        <div class="flex items-center gap-2">
          <span class="hidden text-sm sm:block" style="color: var(--color-muted)">{{ session?.name }}</span>
          <div class="flex h-9 w-9 items-center justify-center rounded-full border text-sm font-semibold" style="border-color: var(--color-line); background: var(--color-brand-50)">
            {{ initials }}
          </div>
        </div>
      </header>

      <main class="min-w-0 flex-1 p-4 md:p-6">
        <slot />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
const open = ref(false)
const query = ref('')
const session = useSession()

const initials = computed(() => {
  const name = session.value?.name || '?'
  return name.split(' ').map((p) => p[0]).slice(0, 2).join('').toUpperCase()
})

function search() {
  if (query.value.trim()) navigateTo({ path: '/biblioteca', query: { q: query.value.trim() } })
}

const nav = [
  { to: '/', label: 'Home', icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="m3 10 9-7 9 7v10a1 1 0 0 1-1 1h-5v-7H9v7H4a1 1 0 0 1-1-1Z"/></svg>' },
  { to: '/editor', label: 'Editor', icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M4 6h16M4 12h16M4 18h16"/><circle cx="8" cy="6" r="2" fill="currentColor"/><circle cx="16" cy="12" r="2" fill="currentColor"/><circle cx="10" cy="18" r="2" fill="currentColor"/></svg>' },
  { to: '/biblioteca', label: 'Biblioteca', icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M4 5h6v14H4zM14 5h6v14h-6z"/></svg>' },
  { to: '/templates', label: 'Templates', icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><rect x="3" y="3" width="18" height="18" rx="3"/><path d="M3 9h18M9 21V9"/></svg>' },
  { to: '/renders', label: 'Renders', icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M12 3v12m0 0 4-4m-4 4-4-4"/><path d="M4 17v2a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2v-2"/></svg>' },
  { to: '/publicacoes', label: 'Publicações', icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M4 11a8 8 0 0 1 8-8h8v8a8 8 0 0 1-16 0Z"/><path d="m4 20 8-8"/></svg>' },
  { to: '/integracoes', label: 'Integrações', icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M9 7V3m6 4V3M6 11h12v4a6 6 0 0 1-12 0v-4Z"/><path d="M12 21v-4"/></svg>' },
  { to: '/configuracoes', label: 'Configurações', icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><circle cx="12" cy="12" r="3"/><path d="M12 2v3m0 14v3M2 12h3m14 0h3M4.9 4.9 7 7m10 10 2.1 2.1M19.1 4.9 17 7M7 17l-2.1 2.1"/></svg>' },
]
</script>

<style scoped>
.nav-item {
  display: flex;
  align-items: center;
  gap: 0.7rem;
  padding: 0.55rem 0.75rem;
  border-radius: 8px;
  font-size: 0.875rem;
  color: var(--color-ink);
}
.nav-item:hover { background: #f4f4f5; }
.nav-active { background: var(--color-brand-50); color: var(--color-brand-600); font-weight: 600; }
</style>
