<template>
  <div class="space-y-5">
    <div class="flex flex-wrap items-end justify-between gap-3">
      <div>
        <h1 class="text-2xl font-bold tracking-tight">Editor de cortes</h1>
        <p class="mt-1 text-sm" style="color: var(--color-muted)">Marque o trecho, escolha o enquadramento e salve o clip.</p>
      </div>
      <select v-model="sourceId" class="input max-w-xs" @change="loadSource">
        <option :value="0">Selecione um vídeo</option>
        <option v-for="s in sources" :key="s.id" :value="s.id">{{ s.title }}</option>
      </select>
    </div>

    <p v-if="error" class="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-700">{{ error }}</p>

    <div v-if="!source" class="card p-10 text-center text-sm" style="color: var(--color-muted)">
      Escolha um vídeo na biblioteca para começar.
      <div class="mt-4"><NuxtLink to="/biblioteca" class="btn btn-primary">Ir para Biblioteca</NuxtLink></div>
    </div>

    <div v-else class="grid gap-5 xl:grid-cols-[minmax(0,1fr)_360px]">
      <section class="card p-4">
        <div class="mx-auto grid max-w-[420px] gap-3">
          <div class="relative overflow-hidden rounded-2xl bg-black" style="aspect-ratio: 9 / 16">
            <video
              v-if="layout === 'blur'"
              :src="source.mediaUrl"
              class="absolute inset-0 h-full w-full scale-125 object-cover opacity-45 blur-xl"
              muted
              loop
              autoplay
              playsinline
            />
            <video
              ref="videoEl"
              :src="source.mediaUrl"
              class="absolute left-0 h-full w-full object-contain"
              :style="layout === 'fill' ? 'top:0;object-fit:cover' : 'top:10%;height:80%'"
              controls
              playsinline
              @timeupdate="onTime"
            />
            <div class="pointer-events-none absolute inset-x-0 top-0 p-4 text-center">
              <p class="truncate font-bold text-white">{{ source.authorName || 'Seu perfil' }}</p>
              <p v-if="source.authorHandle" class="text-xs text-white/70">{{ source.authorHandle }}</p>
            </div>
          </div>

          <div class="space-y-3 rounded-xl bg-zinc-50 p-3">
            <div>
              <div class="flex items-center justify-between text-xs font-semibold" style="color: var(--color-muted)">
                <span>Início {{ fmtDuration(startMs) }}</span>
                <span>Fim {{ fmtDuration(endMs) }}</span>
              </div>
              <input v-model.number="startMs" type="range" :min="0" :max="source.durationMs" step="100" class="mt-2 w-full accent-pink-500" />
              <input v-model.number="endMs" type="range" :min="0" :max="source.durationMs" step="100" class="mt-1 w-full accent-pink-600" />
            </div>
            <div class="flex items-center justify-between text-sm">
              <span class="font-semibold">Duração: {{ fmtDuration(endMs - startMs) }}</span>
              <button class="text-xs font-semibold" style="color: var(--color-brand-600)" @click="useCurrent">Usar início atual</button>
            </div>
          </div>

          <div class="grid gap-3 sm:grid-cols-2">
            <div>
              <label class="label">Enquadramento</label>
              <select v-model="layout" class="input">
                <option value="blur">Fundo desfocado (9:16)</option>
                <option value="fill">Preencher (crop)</option>
              </select>
            </div>
            <div>
              <label class="label">Nome do clip</label>
              <input v-model="clipName" class="input" placeholder="Melhor momento" />
            </div>
          </div>

          <button class="btn btn-primary w-full" :disabled="saving" @click="saveClip">
            {{ saving ? 'Salvando...' : 'Salvar corte' }}
          </button>
        </div>
      </section>

      <section class="card h-fit">
        <header class="flex items-center justify-between border-b px-4 py-3" style="border-color: var(--color-line)">
          <h2 class="font-semibold">Cortes salvos</h2>
          <span class="text-xs" style="color: var(--color-muted)">{{ source.clips?.length || 0 }}</span>
        </header>
        <div v-if="!source.clips?.length" class="p-5 text-sm" style="color: var(--color-muted)">
          Nenhum corte ainda. Marque o início e o fim e salve.
        </div>
        <ul v-else class="divide-y">
          <li v-for="clip in source.clips" :key="clip.id" class="px-4 py-3">
            <div class="flex items-start justify-between gap-2">
              <div class="min-w-0">
                <p class="truncate text-sm font-medium">{{ clip.name || `Corte ${clip.id}` }}</p>
                <p class="text-xs" style="color: var(--color-muted)">
                  {{ fmtDuration(clip.startMs) }} → {{ fmtDuration(clip.endMs) }} · {{ clip.layout }}
                </p>
              </div>
              <button class="text-xs text-red-600" @click="deleteClip(clip)">Excluir</button>
            </div>
            <button class="btn btn-ghost mt-2 h-8 w-full text-xs" @click="renderClip(clip)">Gerar variações →</button>
          </li>
        </ul>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
const route = useRoute()
const error = ref('')
const sources = ref<any[]>([])
const sourceId = ref<number>(0)
const source = ref<any>(null)
const videoEl = ref<HTMLVideoElement | null>(null)
const startMs = ref(0)
const endMs = ref(0)
const layout = ref<'blur' | 'fill'>('blur')
const clipName = ref('')
const saving = ref(false)
const currentMs = ref(0)

const MIN_CLIP_MS = 1000

function onTime() {
  currentMs.value = Math.floor((videoEl.value?.currentTime || 0) * 1000)
}

function useCurrent() {
  startMs.value = Math.min(currentMs.value, endMs.value - MIN_CLIP_MS)
}

async function loadSources() {
  const res = await apiGet<{ sources: any[] }>('/api/sources')
  sources.value = res.sources || []
  const q = Number(route.query.source || 0)
  if (q) {
    sourceId.value = q
    await loadSource()
  }
}

async function loadSource() {
  error.value = ''
  if (!sourceId.value) {
    source.value = null
    return
  }
  try {
    const res = await apiGet<{ source: any }>(`/api/sources/${sourceId.value}`)
    source.value = res.source
    startMs.value = 0
    endMs.value = Math.min(res.source.durationMs, 30000)
    clipName.value = ''
  } catch (e) {
    error.value = errorMessage(e)
  }
}

async function saveClip() {
  error.value = ''
  if (endMs.value - startMs.value < MIN_CLIP_MS) {
    error.value = 'O corte precisa ter pelo menos 1 segundo.'
    return
  }
  saving.value = true
  try {
    await apiSend('/api/clips', 'POST', {
      sourceId: sourceId.value,
      name: clipName.value || `Corte ${(source.value.clips?.length || 0) + 1}`,
      startMs: startMs.value,
      endMs: endMs.value,
      layout: layout.value,
      zoom: 1,
    })
    await loadSource()
  } catch (e) {
    error.value = errorMessage(e)
  } finally {
    saving.value = false
  }
}

async function deleteClip(clip: any) {
  if (!confirm('Excluir este corte e seus renders?')) return
  try {
    await apiSend(`/api/clips/${clip.id}`, 'DELETE')
    await loadSource()
  } catch (e) {
    error.value = errorMessage(e)
  }
}

function renderClip(clip: any) {
  navigateTo(`/renders?clip=${clip.id}`)
}

onMounted(loadSources)
</script>
