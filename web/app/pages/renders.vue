<template>
  <div class="space-y-6">
    <div class="flex flex-wrap items-end justify-between gap-3">
      <div>
        <h1 class="text-2xl font-bold tracking-tight">Renders</h1>
        <p class="mt-1 text-sm" style="color: var(--color-muted)">Cada corte vira uma ou mais variações verticais prontas para postar.</p>
      </div>
      <div class="flex gap-2">
        <button class="btn btn-ghost" :disabled="refreshing" @click="load">{{ refreshing ? 'Atualizando...' : 'Atualizar' }}</button>
        <button class="btn btn-ghost" @click="cleanupOpen = true">Apagar antigos</button>
      </div>
    </div>

    <section v-if="clips.length" class="card p-5">
      <h2 class="font-semibold">Gerar novas variações</h2>
      <div class="mt-3 grid gap-3 sm:grid-cols-4">
        <div class="sm:col-span-2">
          <label class="label">Corte</label>
          <select v-model="form.clipId" class="input">
            <option :value="0">Selecione</option>
            <option v-for="c in clips" :key="c.id" :value="c.id">
              #{{ c.id }} · {{ c.name }} ({{ fmtDuration(c.startMs) }}–{{ fmtDuration(c.endMs) }})
            </option>
          </select>
        </div>
        <div>
          <label class="label">Template</label>
          <select v-model="form.templateId" class="input">
            <option :value="0">Sem template</option>
            <option v-for="t in templates" :key="t.id" :value="t.id">{{ t.name }}</option>
          </select>
        </div>
        <div>
          <label class="label">Variações</label>
          <input v-model.number="form.variantCount" type="number" min="1" max="30" class="input" />
        </div>
      </div>
      <button class="btn btn-primary mt-3" :disabled="!form.clipId || creating" @click="create">
        {{ creating ? 'Enviando...' : 'Gerar' }}
      </button>
      <p v-if="message" class="mt-3 rounded-lg bg-green-50 px-3 py-2 text-sm text-green-700">{{ message }}</p>
    </section>
    <section v-else class="card p-6 text-sm" style="color: var(--color-muted)">
      Nenhum corte encontrado. Crie um corte no
      <NuxtLink to="/editor" class="font-semibold" style="color: var(--color-brand-600)">editor</NuxtLink>.
    </section>

    <p v-if="error" class="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-700">{{ error }}</p>

    <div v-if="cleanupOpen" class="fixed inset-0 z-30 flex items-center justify-center bg-black/40 p-4" @click.self="cleanupOpen = false">
      <form class="card w-full max-w-md p-6" @submit.prevent="purge">
        <h2 class="text-lg font-semibold">Apagar renders antigos</h2>
        <p class="mt-1 text-xs" style="color: var(--color-muted)">
          Os arquivos são removidos do disco. Publicações já feitas continuam no histórico.
        </p>
        <div class="mt-4 space-y-3">
          <label class="flex items-center gap-2 text-sm">
            <input v-model="cleanup.failed" type="checkbox" class="h-4 w-4 accent-pink-500" />
            Renders com falha ou cancelados
          </label>
          <label class="flex items-center gap-2 text-sm">
            <input v-model="cleanup.finished" type="checkbox" class="h-4 w-4 accent-pink-500" />
            Renders concluídos com mais de
            <input v-model.number="cleanup.days" type="number" min="1" max="365" class="input h-9 w-20" />
            dias
          </label>
        </div>
        <p v-if="cleanupError" class="mt-3 rounded-lg bg-red-50 px-3 py-2 text-sm text-red-700">{{ cleanupError }}</p>
        <p v-if="cleanupMessage" class="mt-3 rounded-lg bg-green-50 px-3 py-2 text-sm text-green-700">{{ cleanupMessage }}</p>
        <div class="mt-5 flex justify-end gap-2">
          <button type="button" class="btn btn-ghost" @click="cleanupOpen = false">Fechar</button>
          <button type="submit" class="btn btn-primary" :disabled="purging || (!cleanup.failed && !cleanup.finished)">
            {{ purging ? 'Apagando...' : 'Apagar' }}
          </button>
        </div>
      </form>
    </div>

    <section v-if="loading" class="card p-8 text-center text-sm" style="color: var(--color-muted)">Carregando renders...</section>
    <section v-else-if="!renders.length" class="card p-10 text-center text-sm" style="color: var(--color-muted)">
      Nada renderizado ainda.
    </section>
    <div v-else class="space-y-5">
      <article v-for="r in renders" :key="r.id" class="card overflow-hidden">
        <header class="flex flex-wrap items-center gap-3 border-b px-5 py-3" style="border-color: var(--color-line)">
          <div class="min-w-0 flex-1">
            <p class="font-semibold">Render #{{ r.id }}</p>
            <p class="text-xs" style="color: var(--color-muted)">
              {{ r.outputs?.length || 0 }} de {{ r.variantCount }} variação(ões) · {{ fmtDate(r.createdAt) }}
            </p>
          </div>
          <span :class="statusChip(r.status)">{{ r.status }}</span>
          <button
            v-if="r.status === 'failed' || r.status === 'cancelled'"
            class="btn btn-ghost h-8 px-3 text-xs"
            :disabled="retrying === r.id"
            @click="retry(r)"
          >{{ retrying === r.id ? 'Enviando...' : 'Tentar de novo' }}</button>
          <button
            class="btn btn-danger h-8 px-3 text-xs"
            :disabled="r.status === 'queued' || r.status === 'running' || deleting === r.id"
            :title="r.status === 'queued' || r.status === 'running' ? 'Aguarde o render terminar' : 'Apagar render'"
            @click="removeRender(r)"
          >{{ deleting === r.id ? '...' : 'Apagar' }}</button>
          <a
            v-if="r.outputs?.length"
            :href="`/api/outputs/${r.outputs[0].id}/media`"
            :download="`render-${r.id}.mp4`"
            class="btn btn-ghost h-8 px-3 text-xs"
          >Baixar tudo</a>
          <div class="w-32">
            <div class="h-1.5 overflow-hidden rounded-full bg-zinc-100">
              <div class="h-full rounded-full transition-all" :style="{ width: r.progress + '%', background: 'var(--color-brand-500)' }" />
            </div>
            <p class="mt-1 text-right text-[11px]" style="color: var(--color-muted)">{{ r.progress }}%</p>
          </div>
        </header>
        <p v-if="r.error" class="bg-red-50 px-5 py-2 text-sm text-red-700">{{ r.error }}</p>
        <div v-if="r.outputs?.length" class="flex gap-4 overflow-x-auto p-5 scrollbar-thin">
          <div v-for="o in r.outputs" :key="o.id" class="w-[200px] shrink-0">
            <div class="overflow-hidden rounded-lg bg-black" style="aspect-ratio: 9/16">
              <video :src="o.mediaUrl" class="h-full w-full object-contain" controls preload="metadata" playsinline />
            </div>
            <p class="mt-2 line-clamp-2 text-xs" style="color: var(--color-muted)">{{ o.headline || 'Sem headline' }}</p>
            <div class="mt-2 flex gap-2">
              <a :href="o.mediaUrl" download class="btn btn-ghost h-8 flex-1 text-xs">Baixar</a>
              <NuxtLink :to="`/publicacoes?output=${o.id}`" class="btn btn-primary h-8 flex-1 text-xs">Publicar</NuxtLink>
            </div>
          </div>
        </div>
      </article>
    </div>
  </div>
</template>

<script setup lang="ts">
const route = useRoute()
const renders = ref<any[]>([])
const templates = ref<any[]>([])
const clips = ref<any[]>([])
const loading = ref(true)
const refreshing = ref(false)
const creating = ref(false)
const error = ref('')
const message = ref('')

const form = reactive({ clipId: 0, templateId: 0, variantCount: 3 })
const retrying = ref(0)
const deleting = ref(0)
const cleanupOpen = ref(false)
const purging = ref(false)
const cleanupError = ref('')
const cleanupMessage = ref('')
const cleanup = reactive({ failed: true, finished: false, days: 30 })
let timer: ReturnType<typeof setInterval> | null = null

async function load() {
  refreshing.value = true
  try {
    const res = await apiGet<{ renders: any[] }>('/api/renders')
    renders.value = res.renders || []
  } catch (e) {
    error.value = errorMessage(e)
  } finally {
    loading.value = false
    refreshing.value = false
  }
}

async function loadMeta() {
  try {
    const [t, s] = await Promise.all([
      apiGet<{ templates: any[] }>('/api/templates'),
      apiGet<{ sources: any[] }>('/api/sources'),
    ])
    templates.value = t.templates || []
    const all: any[] = []
    for (const src of s.sources || []) {
      const detail = await apiGet<{ source: any }>(`/api/sources/${src.id}`)
      for (const c of detail.source.clips || []) all.push(c)
    }
    clips.value = all
    const q = Number(route.query.clip || 0)
    if (q) form.clipId = q
  } catch (e) {
    error.value = errorMessage(e)
  }
}

async function create() {
  creating.value = true
  error.value = ''
  message.value = ''
  try {
    await apiSend('/api/renders', 'POST', {
      clipId: form.clipId,
      templateId: form.templateId || undefined,
      variantCount: form.variantCount,
    })
    message.value = 'Render na fila.'
    await load()
    startPolling()
  } catch (e) {
    error.value = errorMessage(e)
  } finally {
    creating.value = false
  }
}

async function removeRender(r: any) {
  const aviso = r.status === 'done'
    ? 'Apagar este render e os arquivos? Publicações já feitas continuam no histórico.'
    : 'Apagar este render e os arquivos?'
  if (!confirm(aviso)) return
  deleting.value = r.id
  error.value = ''
  try {
    await apiSend(`/api/renders/${r.id}`, 'DELETE')
    await load()
  } catch (e) {
    error.value = errorMessage(e)
  } finally {
    deleting.value = 0
  }
}

async function purge() {
  purging.value = true
  cleanupError.value = ''
  cleanupMessage.value = ''
  try {
    const statuses: string[] = []
    if (cleanup.failed) statuses.push('failed', 'cancelled')
    if (cleanup.finished) statuses.push('done')
    const res = await apiSend<{ deleted: number }>('/api/renders/purge', 'POST', {
      statuses,
      olderThanDays: cleanup.finished ? cleanup.days : 0,
    })
    cleanupMessage.value = res.deleted === 0
      ? 'Nenhum render para apagar com esses filtros.'
      : `${res.deleted} render(s) apagado(s).`
    await load()
  } catch (e) {
    cleanupError.value = errorMessage(e)
  } finally {
    purging.value = false
  }
}

async function retry(r: any) {
  retrying.value = r.id
  error.value = ''
  try {
    await apiSend(`/api/renders/${r.id}/retry`, 'POST')
    await load()
    startPolling()
  } catch (e) {
    error.value = errorMessage(e)
  } finally {
    retrying.value = 0
  }
}

function startPolling() {
  if (timer) return
  timer = setInterval(async () => {
    await load()
    if (!renders.value.some((r) => r.status === 'running' || r.status === 'queued')) {
      if (timer) clearInterval(timer)
      timer = null
    }
  }, 2000)
}

onMounted(async () => {
  await Promise.all([load(), loadMeta()])
  if (renders.value.some((r) => r.status === 'running' || r.status === 'queued')) startPolling()
})

onBeforeUnmount(() => {
  if (timer) clearInterval(timer)
})
</script>
