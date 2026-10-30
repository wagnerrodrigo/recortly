<template>
  <div class="space-y-6">
    <div class="flex flex-wrap items-end justify-between gap-3">
      <div>
        <h1 class="text-2xl font-bold tracking-tight">Biblioteca</h1>
        <p class="mt-1 text-sm" style="color: var(--color-muted)">
          Importe vídeos próprios ou com permissão. O crédito de origem é obrigatório para publicar.
        </p>
      </div>
    </div>

    <section class="card p-5">
      <h2 class="font-semibold">Subir vídeo</h2>
      <div
        class="mt-3 flex flex-col items-center justify-center rounded-xl border-2 border-dashed px-6 py-8 text-center transition"
        :class="dragging ? 'border-brand-400 bg-brand-50' : ''"
        style="border-color: var(--color-line)"
        @dragover.prevent="dragging = true"
        @dragleave.prevent="dragging = false"
        @drop.prevent="onDrop"
      >
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6" class="h-8 w-8" style="color: var(--color-muted)">
          <path d="M12 16V4m0 0 4 4m-4-4L8 8" /><path d="M4 16v2a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2v-2" />
        </svg>
        <p class="mt-2 text-sm font-medium">Arraste o vídeo aqui ou selecione um arquivo</p>
        <p class="text-xs" style="color: var(--color-muted)">MP4, MOV ou MKV · até 2 GB</p>
        <input ref="fileInput" type="file" accept="video/*" class="hidden" @change="onPick" />
        <button class="btn btn-ghost mt-3" @click="fileInput?.click()">Selecionar arquivo</button>
      </div>
      <div v-if="uploadName" class="mt-3 flex items-center justify-between rounded-lg bg-zinc-50 px-3 py-2 text-sm">
        <span class="truncate">{{ uploadName }}</span>
        <button v-if="!uploading" class="text-xs font-semibold" style="color: var(--color-brand-600)" @click="startUpload">Enviar</button>
      </div>
      <div v-if="uploading" class="mt-3 h-2 overflow-hidden rounded-full bg-zinc-100">
        <div class="h-full transition-all" :style="{ width: uploadProgress + '%', background: 'var(--color-brand-500)' }" />
      </div>
      <p v-if="error" class="mt-3 rounded-lg bg-red-50 px-3 py-2 text-sm text-red-700">{{ error }}</p>
    </section>

    <section>
      <div class="mb-3 flex items-center justify-between">
        <h2 class="font-semibold">Seus vídeos</h2>
        <span class="text-sm" style="color: var(--color-muted)">{{ filtered.length }} item(ns)</span>
      </div>

      <div v-if="loading" class="card p-8 text-center text-sm" style="color: var(--color-muted)">Carregando...</div>
      <div v-else-if="!filtered.length" class="card p-8 text-center text-sm" style="color: var(--color-muted)">
        Nada por aqui ainda. Suba o primeiro vídeo.
      </div>
      <div v-else class="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
        <article v-for="src in filtered" :key="src.id" class="card overflow-hidden">
          <div class="flex gap-3 p-4">
            <div class="relative h-24 w-16 shrink-0 overflow-hidden rounded-lg bg-zinc-900">
              <video :src="src.mediaUrl" class="h-full w-full object-cover" muted preload="metadata" />
            </div>
            <div class="min-w-0 flex-1">
              <p class="truncate font-semibold">{{ src.title }}</p>
              <p class="mt-0.5 text-xs" style="color: var(--color-muted)">
                {{ fmtDuration(src.durationMs) }} · {{ src.width }}x{{ src.height }} · {{ fmtBytes(src.sizeBytes) }}
              </p>
              <div class="mt-2 flex flex-wrap gap-1.5">
                <span :class="src.licenseType ? 'chip chip-ok' : 'chip chip-warn'">
                  {{ licenseLabel(src.licenseType) }}
                </span>
                <span v-if="src.authorName" class="chip chip-neutral">{{ src.authorName }}</span>
              </div>
            </div>
          </div>
          <div class="flex items-center gap-2 border-t px-4 py-3" style="border-color: var(--color-line)">
            <NuxtLink :to="`/editor?source=${src.id}`" class="btn btn-primary h-8 flex-1 text-xs">Editar cortes</NuxtLink>
            <button class="btn btn-ghost h-8 px-3 text-xs" @click="openMeta(src)">Origem</button>
            <button class="btn btn-danger h-8 px-3 text-xs" @click="remove(src)">Remover</button>
          </div>
        </article>
      </div>
    </section>

    <div v-if="metaSource" class="fixed inset-0 z-30 flex items-center justify-center bg-black/40 p-4" @click.self="metaSource = null">
      <form class="card max-h-[90vh] w-full max-w-lg overflow-y-auto p-6" @submit.prevent="saveMeta">
        <h2 class="text-lg font-semibold">Origem e licença</h2>
        <p class="mt-1 text-xs" style="color: var(--color-muted)">
          Sem esses dados o SmashCut bloqueia a publicação. Use apenas conteúdo próprio ou com permissão.
        </p>
        <div class="mt-4 grid gap-4 sm:grid-cols-2">
          <div class="sm:col-span-2">
            <label class="label">Título</label>
            <input v-model="metaForm.title" class="input" />
          </div>
          <div>
            <label class="label">Autor</label>
            <input v-model="metaForm.authorName" class="input" placeholder="Nome do criador" />
          </div>
          <div>
            <label class="label">@ do autor</label>
            <input v-model="metaForm.authorHandle" class="input" placeholder="@perfil" />
          </div>
          <div class="sm:col-span-2">
            <label class="label">Link original</label>
            <input v-model="metaForm.originalUrl" class="input" placeholder="https://..." />
          </div>
          <div>
            <label class="label">Licença</label>
            <select v-model="metaForm.licenseType" class="input">
              <option value="">Selecione</option>
              <option value="own">Conteúdo próprio</option>
              <option value="permission">Permissão do criador</option>
              <option value="credit">Uso com crédito</option>
            </select>
          </div>
          <div>
            <label class="label">Comprovante da permissão</label>
            <input v-model="metaForm.permissionProof" class="input" placeholder="print, e-mail, data..." />
          </div>
          <div class="sm:col-span-2">
            <label class="label">Notas</label>
            <textarea v-model="metaForm.notes" class="input" rows="2" />
          </div>
        </div>
        <p v-if="metaError" class="mt-3 rounded-lg bg-red-50 px-3 py-2 text-sm text-red-700">{{ metaError }}</p>
        <div class="mt-5 flex justify-end gap-2">
          <button type="button" class="btn btn-ghost" @click="metaSource = null">Cancelar</button>
          <button type="submit" class="btn btn-primary" :disabled="savingMeta">{{ savingMeta ? 'Salvando...' : 'Salvar' }}</button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
const route = useRoute()
const loading = ref(true)
const sources = ref<any[]>([])
const error = ref('')
const dragging = ref(false)
const fileInput = ref<HTMLInputElement | null>(null)
const pendingFile = ref<File | null>(null)
const uploadName = ref('')
const uploading = ref(false)
const uploadProgress = ref(0)

const metaSource = ref<any>(null)
const metaForm = ref<any>({})
const savingMeta = ref(false)
const metaError = ref('')

const q = computed(() => (typeof route.query.q === 'string' ? route.query.q.toLowerCase() : ''))
const filtered = computed(() =>
  !q.value ? sources.value : sources.value.filter((s) => (s.title + ' ' + s.authorName).toLowerCase().includes(q.value)),
)

function licenseLabel(lt?: string) {
  if (lt === 'own') return 'Próprio'
  if (lt === 'permission') return 'Com permissão'
  if (lt === 'credit') return 'Com crédito'
  return 'Licença pendente'
}

async function load() {
  loading.value = true
  try {
    const res = await apiGet<{ sources: any[] }>('/api/sources')
    sources.value = res.sources || []
  } catch (e) {
    error.value = errorMessage(e)
  } finally {
    loading.value = false
  }
}

function onDrop(ev: DragEvent) {
  dragging.value = false
  const file = ev.dataTransfer?.files?.[0]
  if (file) pick(file)
}

function onPick(ev: Event) {
  const file = (ev.target as HTMLInputElement).files?.[0]
  if (file) pick(file)
}

function pick(file: File) {
  error.value = ''
  pendingFile.value = file
  uploadName.value = `${file.name} · ${fmtBytes(file.size)}`
}

async function startUpload() {
  if (!pendingFile.value) return
  error.value = ''
  uploading.value = true
  uploadProgress.value = 0

  const form = new FormData()
  form.append('video', pendingFile.value)
  form.append('title', pendingFile.value.name.replace(/\.[^.]+$/, ''))

  try {
    await new Promise<void>((resolve, reject) => {
      const xhr = new XMLHttpRequest()
      xhr.open('POST', '/api/sources')
      xhr.upload.onprogress = (ev) => {
        if (ev.lengthComputable) uploadProgress.value = Math.round((ev.loaded / ev.total) * 100)
      }
      xhr.onload = () => {
        if (xhr.status >= 200 && xhr.status < 300) resolve()
        else {
          try {
            reject(new Error(JSON.parse(xhr.responseText).error || 'Falha no envio'))
          } catch {
            reject(new Error('Falha no envio'))
          }
        }
      }
      xhr.onerror = () => reject(new Error('Falha de rede'))
      xhr.send(form)
    })
    pendingFile.value = null
    uploadName.value = ''
    await load()
  } catch (e) {
    error.value = errorMessage(e)
  } finally {
    uploading.value = false
  }
}

function openMeta(src: any) {
  metaSource.value = src
  metaError.value = ''
  metaForm.value = {
    title: src.title,
    authorName: src.authorName || '',
    authorHandle: src.authorHandle || '',
    originalUrl: src.originalUrl || '',
    licenseType: src.licenseType || '',
    permissionProof: src.permissionProof || '',
    notes: src.notes || '',
  }
}

async function saveMeta() {
  savingMeta.value = true
  metaError.value = ''
  try {
    await apiSend(`/api/sources/${metaSource.value.id}`, 'PATCH', metaForm.value)
    metaSource.value = null
    await load()
  } catch (e) {
    metaError.value = errorMessage(e)
  } finally {
    savingMeta.value = false
  }
}

async function remove(src: any) {
  if (!confirm(`Remover "${src.title}"? Os clips e renders ligados também serão apagados.`)) return
  try {
    await apiSend(`/api/sources/${src.id}`, 'DELETE')
    await load()
  } catch (e) {
    error.value = errorMessage(e)
  }
}

onMounted(load)
</script>
