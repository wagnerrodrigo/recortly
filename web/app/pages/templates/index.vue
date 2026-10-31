<template>
  <div class="space-y-8">
    <div class="flex flex-wrap items-end justify-between gap-3">
      <div>
        <h1 class="text-2xl font-bold tracking-tight">Templates</h1>
        <p class="mt-1 text-sm" style="color: var(--color-muted)">
          Escolha um modelo, ajuste em 7 passos e gere as variações do vídeo.
        </p>
      </div>
      <NuxtLink to="/editor" class="btn btn-ghost">Editor livre</NuxtLink>
    </div>

    <section class="card p-4">
      <div class="grid gap-3 text-xs sm:grid-cols-4" style="color: var(--color-muted)">
        <div class="flex gap-2">
          <span class="flex h-6 w-6 shrink-0 items-center justify-center rounded-full text-[11px] font-bold" style="background: var(--color-brand-50); color: var(--color-brand-600)">1</span>
          <p><strong style="color: var(--color-ink)">Escolha um modelo</strong><br />ou comece em branco.</p>
        </div>
        <div class="flex gap-2">
          <span class="flex h-6 w-6 shrink-0 items-center justify-center rounded-full text-[11px] font-bold" style="background: var(--color-brand-50); color: var(--color-brand-600)">2</span>
          <p><strong style="color: var(--color-ink)">Ajuste em 7 passos</strong><br />perfil, zonas, frases, mídia, estilo, variações.</p>
        </div>
        <div class="flex gap-2">
          <span class="flex h-6 w-6 shrink-0 items-center justify-center rounded-full text-[11px] font-bold" style="background: var(--color-brand-50); color: var(--color-brand-600)">3</span>
          <p><strong style="color: var(--color-ink)">Gere as variações</strong><br />o progresso aparece na última etapa.</p>
        </div>
        <div class="flex gap-2">
          <span class="flex h-6 w-6 shrink-0 items-center justify-center rounded-full text-[11px] font-bold" style="background: var(--color-brand-50); color: var(--color-brand-600)">4</span>
          <p><strong style="color: var(--color-ink)">Baixe ou publique</strong><br />tudo fica salvo em Renders e Publicações.</p>
        </div>
      </div>
    </section>

    <section v-if="presets.length">
      <h2 class="mb-3 text-sm font-semibold uppercase tracking-wide" style="color: var(--color-muted)">Escolha um template</h2>
      <div class="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <article
          v-for="preset in presets"
          :key="preset.id"
          class="card flex flex-col p-4"
        >
          <div class="mx-auto w-full max-w-[180px]">
            <TemplatePreview :spec="preset.spec" :layout="layouts['p' + preset.id]" />
          </div>
          <p class="mt-3 text-center text-sm font-semibold">{{ preset.name }}</p>
          <p class="mt-1 flex-1 text-center text-[11px]" style="color: var(--color-muted)">{{ preset.description }}</p>
          <button class="btn btn-primary mt-3 w-full text-xs" :disabled="creating === preset.id" @click="createFromPreset(preset)">
            {{ creating === preset.id ? 'Criando...' : 'Usar este template →' }}
          </button>
        </article>
      </div>
    </section>

    <section>
      <div class="mb-3 flex items-center justify-between">
        <h2 class="text-sm font-semibold uppercase tracking-wide" style="color: var(--color-muted)">Meus templates</h2>
        <button class="btn btn-ghost h-9 text-xs" :disabled="creating === 'blank'" @click="createBlank">+ Em branco</button>
      </div>

      <div v-if="loading" class="card p-8 text-center text-sm" style="color: var(--color-muted)">Carregando...</div>
      <div v-else-if="!templates.length" class="card p-8 text-center text-sm" style="color: var(--color-muted)">
        Nenhum template salvo ainda. Escolha um modelo acima.
      </div>
      <div v-else class="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <article v-for="t in templates" :key="t.id" class="card flex flex-col p-4">
          <div class="mx-auto w-full max-w-[180px]">
            <TemplatePreview :spec="t.spec" :layout="layouts['t' + t.id]" :headline="t.spec.phrases?.[0]" />
          </div>
          <p class="mt-3 truncate text-center text-sm font-semibold">{{ t.name }}</p>
          <p class="text-center text-[11px]" style="color: var(--color-muted)">
            {{ t.spec.zones?.length || 0 }} zona(s) · {{ t.spec.variations?.count || 1 }} variação(ões)
          </p>
          <div class="mt-3 flex gap-2">
            <NuxtLink :to="`/templates/${t.id}`" class="btn btn-primary h-8 flex-1 text-xs">Abrir</NuxtLink>
            <button class="btn btn-danger h-8 px-3 text-xs" @click="remove(t)">Excluir</button>
          </div>
        </article>
      </div>
    </section>

    <p v-if="error" class="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-700">{{ error }}</p>
  </div>
</template>

<script setup lang="ts">
const route = useRoute()
const clipQuery = computed(() => (typeof route.query.clip === 'string' ? route.query.clip : ''))
const presets = ref<any[]>([])
const templates = ref<any[]>([])
const layouts = reactive<Record<string, any>>({})
const loading = ref(true)
const creating = ref('')
const error = ref('')

async function loadLayouts() {
  const jobs: Promise<void>[] = []
  for (const p of presets.value) {
    jobs.push(fetchLayout('p' + p.id, p.spec))
  }
  for (const t of templates.value) {
    jobs.push(fetchLayout('t' + t.id, t.spec))
  }
  await Promise.allSettled(jobs)
}

async function fetchLayout(key: string, spec: any) {
  try {
    const res = await apiSend<{ layout: any }>('/api/templates/layout', 'POST', { spec })
    layouts[key] = res.layout
  } catch {
    // preview quebrado não deve bloquear a tela
  }
}

async function load() {
  loading.value = true
  try {
    const [p, t] = await Promise.all([
      apiGet<{ presets: any[] }>('/api/templates/presets'),
      apiGet<{ templates: any[] }>('/api/templates'),
    ])
    presets.value = p.presets || []
    templates.value = t.templates || []
    await loadLayouts()
  } catch (e) {
    error.value = errorMessage(e)
  } finally {
    loading.value = false
  }
}

async function createFromPreset(preset: any) {
  creating.value = preset.id
  error.value = ''
  try {
    const spec = { ...preset.spec, name: preset.name }
    const res = await apiSend<{ template: any }>('/api/templates', 'POST', { name: preset.name, spec })
    await navigateTo(`/templates/${res.template.id}${clipQuery.value ? '?clip=' + clipQuery.value : ''}`)
  } catch (e) {
    error.value = errorMessage(e)
  } finally {
    creating.value = ''
  }
}

async function createBlank() {
  creating.value = 'blank'
  error.value = ''
  try {
    const res = await apiSend<{ template: any }>('/api/templates', 'POST', {
      name: 'Novo template',
      spec: defaultSpec(),
    })
    await navigateTo(`/templates/${res.template.id}${clipQuery.value ? '?clip=' + clipQuery.value : ''}`)
  } catch (e) {
    error.value = errorMessage(e)
  } finally {
    creating.value = ''
  }
}

async function remove(t: any) {
  if (!confirm(`Excluir o template "${t.name}"?`)) return
  try {
    await apiSend(`/api/templates/${t.id}`, 'DELETE')
    await load()
  } catch (e) {
    error.value = errorMessage(e)
  }
}

onMounted(load)
</script>
