<template>
  <div class="space-y-6">
    <div class="flex flex-wrap items-end justify-between gap-3">
      <div>
        <h1 class="text-2xl font-bold tracking-tight">Olá, {{ session?.name || 'criador' }}</h1>
        <p class="mt-1 text-sm" style="color: var(--color-muted)">Do vídeo bruto aos cortes publicados.</p>
      </div>
      <NuxtLink to="/biblioteca" class="btn btn-primary">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="h-4 w-4"><path d="M12 5v14M5 12h14" /></svg>
        Subir vídeo
      </NuxtLink>
    </div>

    <div class="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
      <div v-for="stat in stats" :key="stat.label" class="card p-5">
        <p class="text-xs font-semibold uppercase tracking-wide" style="color: var(--color-muted)">{{ stat.label }}</p>
        <p class="mt-2 text-3xl font-bold">{{ stat.value }}</p>
        <p class="mt-1 text-xs" style="color: var(--color-muted)">{{ stat.hint }}</p>
      </div>
    </div>

    <div class="grid gap-4 lg:grid-cols-3">
      <section class="card lg:col-span-2">
        <header class="flex items-center justify-between border-b px-5 py-4" style="border-color: var(--color-line)">
          <h2 class="font-semibold">Renders recentes</h2>
          <NuxtLink to="/renders" class="text-sm font-semibold" style="color: var(--color-brand-600)">Ver todos</NuxtLink>
        </header>
        <div v-if="loading" class="p-5 text-sm" style="color: var(--color-muted)">Carregando...</div>
        <div v-else-if="!renders.length" class="p-8 text-center text-sm" style="color: var(--color-muted)">
          Nenhum render ainda. Suba um vídeo, crie um corte e gere as variações.
        </div>
        <ul v-else class="divide-y">
          <li v-for="r in renders.slice(0, 6)" :key="r.id" class="flex items-center gap-4 px-5 py-3">
            <div class="flex h-10 w-10 items-center justify-center rounded-lg" style="background: var(--color-brand-50)">
              <svg viewBox="0 0 24 24" fill="none" stroke="var(--color-brand-600)" stroke-width="2" class="h-5 w-5"><path d="m10 8 6 4-6 4V8Z" /></svg>
            </div>
            <div class="min-w-0 flex-1">
              <p class="truncate text-sm font-medium">Render #{{ r.id }} · {{ r.variantCount }} variação(ões)</p>
              <p class="text-xs" style="color: var(--color-muted)">{{ fmtDate(r.createdAt) }}</p>
            </div>
            <span :class="statusChip(r.status)">{{ r.status }}</span>
            <div class="w-24">
              <div class="h-1.5 overflow-hidden rounded-full bg-zinc-100">
                <div class="h-full rounded-full" :style="{ width: r.progress + '%', background: 'var(--color-brand-500)' }" />
              </div>
            </div>
          </li>
        </ul>
      </section>

      <section class="card">
        <header class="border-b px-5 py-4" style="border-color: var(--color-line)">
          <h2 class="font-semibold">Próximos passos</h2>
        </header>
        <ol class="space-y-4 p-5 text-sm">
          <li v-for="(step, i) in steps" :key="step.title" class="flex gap-3">
            <span class="flex h-6 w-6 shrink-0 items-center justify-center rounded-full text-xs font-bold" style="background: var(--color-brand-50); color: var(--color-brand-600)">{{ i + 1 }}</span>
            <div>
              <p class="font-medium">{{ step.title }}</p>
              <p class="text-xs" style="color: var(--color-muted)">{{ step.hint }}</p>
            </div>
          </li>
        </ol>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
const session = useSession()
const loading = ref(true)
const renders = ref<any[]>([])
const sources = ref<any[]>([])
const posts = ref<any[]>([])

const stats = computed(() => [
  { label: 'Vídeos na biblioteca', value: sources.value.length, hint: 'originais importados' },
  { label: 'Renders', value: renders.value.length, hint: 'variações geradas' },
  { label: 'Prontos', value: renders.value.filter((r) => r.status === 'done').length, hint: 'arquivos para baixar' },
  { label: 'Publicações', value: posts.value.length, hint: 'agendadas ou publicadas' },
])

const steps = [
  { title: 'Suba o vídeo', hint: 'Biblioteca → escolha o arquivo e registre a origem' },
  { title: 'Marque os cortes', hint: 'Editor → defina início, fim e enquadramento' },
  { title: 'Escolha um template', hint: 'Templates → headline, estilo e variações' },
  { title: 'Gere e publique', hint: 'Renders → baixe ou agende no YouTube, Instagram e TikTok' },
]

onMounted(async () => {
  try {
    const [s, r, p] = await Promise.all([
      apiGet<{ sources: any[] }>('/api/sources'),
      apiGet<{ renders: any[] }>('/api/renders'),
      apiGet<{ posts: any[] }>('/api/posts'),
    ])
    sources.value = s.sources || []
    renders.value = r.renders || []
    posts.value = p.posts || []
  } catch {
    // sessão expirada é tratada pelo middleware
  } finally {
    loading.value = false
  }
})
</script>
