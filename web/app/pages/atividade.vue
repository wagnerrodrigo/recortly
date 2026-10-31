<template>
  <div class="space-y-5">
    <div class="flex flex-wrap items-end justify-between gap-3">
      <div>
        <h1 class="text-2xl font-bold tracking-tight">Atividade</h1>
        <p class="mt-1 text-sm" style="color: var(--color-muted)">
          Logs e erros do render, da publicação e da API. Use isto quando algo travar.
        </p>
      </div>
      <div class="flex items-center gap-2">
        <button class="btn btn-ghost" :disabled="loading" @click="load">{{ loading ? 'Atualizando...' : 'Atualizar' }}</button>
        <button class="btn btn-ghost" @click="cleanupOpen = true">Limpar</button>
      </div>
    </div>

    <section v-if="stuck.length" class="card border-amber-200 p-4" style="background: #fffbeb">
      <div class="flex flex-wrap items-center gap-3">
        <div class="min-w-0 flex-1">
          <p class="text-sm font-semibold text-amber-900">
            {{ stuck.length }} render(s) parecem travados
          </p>
          <p class="text-xs text-amber-800">
            Estão em "running" há mais de 10 minutos. Você pode destravar para gerar de novo.
          </p>
        </div>
        <button class="btn btn-ghost h-9 text-xs" @click="recover">Destravar</button>
      </div>
    </section>

    <div class="card flex flex-wrap items-center gap-2 p-2">
      <button
        v-for="f in levels"
        :key="f.value"
        class="rounded-lg px-3 py-1.5 text-xs font-semibold"
        :class="level === f.value ? 'text-white' : ''"
        :style="level === f.value ? 'background: var(--color-brand-500)' : 'color: var(--color-muted)'"
        @click="level = f.value; load()"
      >{{ f.label }}</button>
      <span class="flex-1" />
      <label class="flex items-center gap-2 text-xs" style="color: var(--color-muted)">
        <input v-model="autoRefresh" type="checkbox" class="h-4 w-4 accent-pink-500" />
        Atualizar automaticamente
      </label>
    </div>

    <p v-if="error" class="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-700">{{ error }}</p>

    <div v-if="!events.length" class="card p-10 text-center text-sm" style="color: var(--color-muted)">
      Nenhum evento registrado{{ level ? ' nesse nível' : '' }}.
    </div>

    <ul v-else class="space-y-2">
      <li v-for="e in events" :key="e.id" class="card p-3">
        <div class="flex flex-wrap items-start gap-2">
          <span :class="levelChip(e.level)">{{ e.level }}</span>
          <span class="chip chip-neutral">{{ sourceLabel(e.source) }}</span>
          <span class="text-[11px]" style="color: var(--color-muted)">{{ fmtDate(e.createdAt) }}</span>
          <span class="flex-1" />
          <button class="text-[11px] underline" style="color: var(--color-muted)" @click="toggle(e.id)">
            {{ expanded === e.id ? 'Ocultar detalhes' : 'Detalhes' }}
          </button>
        </div>
        <p class="mt-1 text-sm">{{ e.message }}</p>
        <pre
          v-if="expanded === e.id && e.details && Object.keys(e.details).length"
          class="mt-2 overflow-x-auto rounded-lg bg-zinc-900 p-3 text-[11px] text-zinc-100"
        >{{ JSON.stringify(e.details, null, 2) }}</pre>
      </li>
    </ul>

    <div v-if="cleanupOpen" class="fixed inset-0 z-30 flex items-center justify-center bg-black/40 p-4" @click.self="cleanupOpen = false">
      <form class="card w-full max-w-md p-6" @submit.prevent="clearEvents">
        <h2 class="text-lg font-semibold">Limpar eventos</h2>
        <p class="mt-1 text-xs" style="color: var(--color-muted)">
          Remove registros antigos. Os renders e arquivos não são afetados.
        </p>
        <div class="mt-4 space-y-3">
          <label class="flex items-center gap-2 text-sm">
            <input v-model="oldDays" type="radio" :value="30" class="accent-pink-500" />
            Com mais de 30 dias
          </label>
          <label class="flex items-center gap-2 text-sm">
            <input v-model="oldDays" type="radio" :value="7" class="accent-pink-500" />
            Com mais de 7 dias
          </label>
          <label class="flex items-center gap-2 text-sm">
            <input v-model="oldDays" type="radio" :value="0" class="accent-pink-500" />
            Todos os eventos
          </label>
        </div>
        <p v-if="clearMessage" class="mt-3 rounded-lg bg-green-50 px-3 py-2 text-sm text-green-700">{{ clearMessage }}</p>
        <div class="mt-5 flex justify-end gap-2">
          <button type="button" class="btn btn-ghost" @click="cleanupOpen = false">Fechar</button>
          <button type="submit" class="btn btn-primary" :disabled="clearing">{{ clearing ? 'Limpando...' : 'Limpar' }}</button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
const events = ref<any[]>([])
const stuck = ref<any[]>([])
const loading = ref(false)
const error = ref('')
const level = ref('')
const expanded = ref(0)
const autoRefresh = ref(true)
const cleanupOpen = ref(false)
const clearing = ref(false)
const clearMessage = ref('')
const oldDays = ref(30)
let timer: ReturnType<typeof setInterval> | null = null

const levels = [
  { label: 'Todos', value: '' },
  { label: 'Erros', value: 'error' },
  { label: 'Avisos', value: 'warn' },
  { label: 'Info', value: 'info' },
]

function levelChip(l: string) {
  if (l === 'error') return 'chip chip-err'
  if (l === 'warn') return 'chip chip-warn'
  return 'chip chip-neutral'
}

function sourceLabel(s: string) {
  const map: Record<string, string> = {
    render: 'render',
    publish: 'publicação',
    video: 'vídeo',
    api: 'api',
    ai: 'ia',
  }
  return map[s] || s
}

function toggle(id: number) {
  expanded.value = expanded.value === id ? 0 : id
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    const q = level.value ? `?level=${level.value}&limit=200` : '?limit=200'
    const [ev, renders] = await Promise.all([
      apiGet<{ events: any[] }>(`/api/events${q}`),
      apiGet<{ renders: any[] }>('/api/renders'),
    ])
    events.value = ev.events || []
    const limit = Date.now() - 10 * 60 * 1000
    stuck.value = (renders.renders || []).filter(
      (r) => r.status === 'running' && new Date(r.createdAt).getTime() < limit,
    )
  } catch (e) {
    error.value = errorMessage(e)
  } finally {
    loading.value = false
  }
}

async function recover() {
  try {
    await apiSend('/api/renders/recover', 'POST')
    await load()
  } catch (e) {
    error.value = errorMessage(e)
  }
}

async function clearEvents() {
  clearing.value = true
  clearMessage.value = ''
  try {
    const res = await apiSend<{ deleted: number }>('/api/events/clear', 'POST', { olderThanDays: oldDays.value })
    clearMessage.value = `${res.deleted} evento(s) removido(s).`
    await load()
  } catch (e) {
    error.value = errorMessage(e)
  } finally {
    clearing.value = false
  }
}

onMounted(() => {
  load()
  timer = setInterval(() => {
    if (autoRefresh.value) load()
  }, 5000)
})

onBeforeUnmount(() => {
  if (timer) clearInterval(timer)
})
</script>
