<template>
  <div class="space-y-6">
    <div class="flex flex-wrap items-end justify-between gap-3">
      <div>
        <h1 class="text-2xl font-bold tracking-tight">Templates</h1>
        <p class="mt-1 text-sm" style="color: var(--color-muted)">
          Monte o layout do post: perfil, headline, estilo e variações.
        </p>
      </div>
      <div class="flex items-center gap-2">
        <select v-model="selectedId" class="input max-w-[240px]" @change="loadSelected">
          <option :value="0">Novo template</option>
          <option v-for="t in templates" :key="t.id" :value="t.id">{{ t.name }}</option>
        </select>
        <button class="btn btn-primary" :disabled="saving" @click="save">{{ saving ? 'Salvando...' : 'Salvar template' }}</button>
      </div>
    </div>

    <div class="card flex flex-wrap gap-1 p-2">
      <button
        v-for="(step, i) in steps"
        :key="step"
        class="rounded-lg px-3 py-1.5 text-sm font-medium"
        :class="activeStep === i ? 'text-white' : ''"
        :style="activeStep === i ? 'background: var(--color-brand-500)' : 'color: var(--color-muted)'"
        @click="activeStep = i"
      >
        {{ i + 1 }}. {{ step }}
      </button>
      <span class="flex-1" />
      <button class="rounded-lg px-3 py-1.5 text-sm" style="color: var(--color-muted)" @click="activeStep = steps.length - 1">
        Ir para gerar →
      </button>
    </div>

    <div class="grid gap-5 xl:grid-cols-[minmax(0,1fr)_360px]">
      <section class="card p-5">
        <div v-if="activeStep === 0" class="space-y-4">
          <h2 class="font-semibold">Perfil</h2>
          <div class="grid gap-4 sm:grid-cols-2">
            <div>
              <label class="label">Nome exibido</label>
              <input v-model="spec.profile.name" class="input" />
            </div>
            <div>
              <label class="label">@</label>
              <input v-model="spec.profile.handle" class="input" />
            </div>
          </div>
          <label class="flex items-center gap-2 text-sm">
            <input v-model="spec.profile.verified" type="checkbox" class="h-4 w-4 accent-pink-500" />
            Mostrar selo de verificado
          </label>
          <label class="flex items-center gap-2 text-sm">
            <input v-model="spec.text.showHandle" type="checkbox" class="h-4 w-4 accent-pink-500" />
            Mostrar @ abaixo do nome
          </label>
        </div>

        <div v-if="activeStep === 1" class="space-y-4">
          <h2 class="font-semibold">Frases</h2>
          <p class="text-xs" style="color: var(--color-muted)">
            Uma frase por variação. Se houver menos frases que variações, elas se repetem em ciclo.
          </p>
          <div v-for="(_, i) in spec.phrases" :key="i" class="flex gap-2">
            <input v-model="spec.phrases[i]" class="input" :placeholder="`Frase ${i + 1}`" />
            <button class="btn btn-ghost px-3" :disabled="spec.phrases.length === 1" @click="spec.phrases.splice(i, 1)">×</button>
          </div>
          <button class="btn btn-ghost" @click="spec.phrases.push('')">+ Adicionar frase</button>
        </div>

        <div v-if="activeStep === 2" class="space-y-4">
          <h2 class="font-semibold">Estilo do texto</h2>
          <div class="grid gap-4 sm:grid-cols-2">
            <div>
              <label class="label">Tamanho da fonte ({{ spec.text.size }}px)</label>
              <input v-model.number="spec.text.size" type="range" min="24" max="110" class="w-full accent-pink-500" />
            </div>
            <div>
              <label class="label">Cor do texto</label>
              <div class="flex items-center gap-2">
                <input v-model="spec.text.color" type="color" class="h-10 w-12 rounded border" />
                <input v-model="spec.text.color" class="input" />
              </div>
            </div>
            <div>
              <label class="label">Espaçamento lateral ({{ spec.text.sideSpacing }}px)</label>
              <input v-model.number="spec.text.sideSpacing" type="range" min="0" max="360" class="w-full accent-pink-500" />
            </div>
            <div>
              <label class="label">Espaço acima do perfil ({{ spec.text.offsetFromProfile }}px)</label>
              <input v-model.number="spec.text.offsetFromProfile" type="range" min="0" max="500" class="w-full accent-pink-500" />
            </div>
            <div class="sm:col-span-2">
              <label class="label">Fundo do canvas</label>
              <div class="flex items-center gap-2">
                <input v-model="spec.canvas.background" type="color" class="h-10 w-12 rounded border" />
                <input v-model="spec.canvas.background" class="input" />
              </div>
            </div>
          </div>
          <div class="rounded-lg bg-zinc-50 p-3">
            <label class="flex items-center gap-2 text-sm font-medium">
              <input v-model="spec.text.strokeEnabled" type="checkbox" class="h-4 w-4 accent-pink-500" />
              Contorno no texto
            </label>
            <div v-if="spec.text.strokeEnabled" class="mt-3 grid gap-3 sm:grid-cols-2">
              <div>
                <label class="label">Tamanho do contorno ({{ spec.text.strokeWidth }}px)</label>
                <input v-model.number="spec.text.strokeWidth" type="range" min="1" max="12" class="w-full accent-pink-500" />
              </div>
              <div>
                <label class="label">Cor do contorno</label>
                <input v-model="spec.text.strokeColor" type="color" class="h-10 w-16 rounded border" />
              </div>
            </div>
          </div>
        </div>

        <div v-if="activeStep === 3" class="space-y-4">
          <h2 class="font-semibold">Variações</h2>
          <div class="grid gap-4 sm:grid-cols-2">
            <div>
              <label class="label">Quantidade ({{ spec.variations.count }})</label>
              <input v-model.number="spec.variations.count" type="range" min="1" max="30" class="w-full accent-pink-500" />
            </div>
            <div>
              <label class="label">Seed (reproduzível)</label>
              <input v-model.number="spec.variations.seedBase" type="number" class="input" />
            </div>
            <div>
              <label class="label">Zoom mínimo</label>
              <input v-model.number="spec.variations.zoom[0]" type="number" step="0.01" min="1" max="1.5" class="input" />
            </div>
            <div>
              <label class="label">Zoom máximo</label>
              <input v-model.number="spec.variations.zoom[1]" type="number" step="0.01" min="1" max="1.5" class="input" />
            </div>
            <div>
              <label class="label">Saturação mínima</label>
              <input v-model.number="spec.variations.saturation[0]" type="number" step="0.01" min="0.5" max="2" class="input" />
            </div>
            <div>
              <label class="label">Saturação máxima</label>
              <input v-model.number="spec.variations.saturation[1]" type="number" step="0.01" min="0.5" max="2" class="input" />
            </div>
          </div>
          <label class="flex items-center gap-2 text-sm">
            <input v-model="spec.variations.mirror" type="checkbox" class="h-4 w-4 accent-pink-500" />
            Espelhar variações ímpares
          </label>
        </div>

        <div v-if="activeStep === 4" class="space-y-4">
          <h2 class="font-semibold">Gerar</h2>
          <p class="text-sm" style="color: var(--color-muted)">
            Salve o template e gere as variações a partir de um corte no editor.
          </p>
          <div class="grid gap-4 sm:grid-cols-2">
            <div>
              <label class="label">Corte</label>
              <select v-model="genClipId" class="input">
                <option :value="0">Selecione</option>
                <option v-for="c in clips" :key="c.id" :value="c.id">#{{ c.id }} · {{ c.name }}</option>
              </select>
            </div>
            <div>
              <label class="label">Quantidade</label>
              <input v-model.number="spec.variations.count" type="number" min="1" max="30" class="input" />
            </div>
          </div>
          <button class="btn btn-primary" :disabled="!genClipId || generating" @click="generate">
            {{ generating ? 'Enviando...' : 'Gerar variações' }}
          </button>
          <p v-if="message" class="rounded-lg bg-green-50 px-3 py-2 text-sm text-green-700">{{ message }}</p>
        </div>
      </section>

      <aside class="card h-fit p-4">
        <p class="mb-3 text-xs font-semibold uppercase tracking-wide" style="color: var(--color-muted)">Preview</p>
        <TemplatePreview :spec="spec" :headline="spec.phrases[0]" class="mx-auto max-w-[280px]" />
        <p class="mt-3 text-center text-[11px]" style="color: var(--color-muted)">
          Preview aproximado. O resultado final é renderizado pelo FFmpeg.
        </p>
      </aside>
    </div>

    <p v-if="error" class="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-700">{{ error }}</p>
  </div>
</template>

<script setup lang="ts">
const steps = ['Perfil', 'Frases', 'Estilo', 'Variações', 'Gerar']
const activeStep = ref(0)
const templates = ref<any[]>([])
const selectedId = ref(0)
const saving = ref(false)
const generating = ref(false)
const error = ref('')
const message = ref('')
const clips = ref<any[]>([])
const genClipId = ref(0)

const spec = ref<any>(createDefault())

function createDefault() {
  return {
    schemaVersion: 1,
    name: 'Novo template',
    canvas: { width: 1080, height: 1920, fps: 30, background: '#0E0E10' },
    profile: { name: 'Seu perfil', handle: '@seuperfil', verified: true, avatar: '' },
    text: {
      font: 'Inter Bold',
      size: 53,
      color: '#FFFFFF',
      align: 'center',
      sideSpacing: 84,
      offsetFromProfile: 156,
      strokeEnabled: false,
      strokeWidth: 4,
      strokeColor: '#000000',
      showHandle: true,
    },
    phrases: ['Coloque sua headline aqui'],
    variations: { count: 5, zoom: [1.0, 1.06], saturation: [1.0, 1.08], speed: [0.98, 1.02], mirror: false, seedBase: 42 },
  }
}


async function loadTemplates() {
  try {
    const res = await apiGet<{ templates: any[] }>('/api/templates')
    templates.value = res.templates || []
  } catch (e) {
    error.value = errorMessage(e)
  }
}

async function loadSelected() {
  if (!selectedId.value) {
    spec.value = createDefault()
    return
  }
  try {
    const res = await apiGet<{ template: any }>(`/api/templates/${selectedId.value}`)
    spec.value = res.template.spec
  } catch (e) {
    error.value = errorMessage(e)
  }
}

async function save() {
  saving.value = true
  error.value = ''
  try {
    if (selectedId.value) {
      await apiSend(`/api/templates/${selectedId.value}`, 'PUT', { name: spec.value.name, spec: spec.value })
    } else {
      const res = await apiSend<{ template: any }>('/api/templates', 'POST', { name: spec.value.name, spec: spec.value })
      selectedId.value = res.template.id
    }
    await loadTemplates()
    message.value = 'Template salvo.'
  } catch (e) {
    error.value = errorMessage(e)
  } finally {
    saving.value = false
  }
}

async function loadClips() {
  try {
    const res = await apiGet<{ sources: any[] }>('/api/sources')
    const all: any[] = []
    for (const s of res.sources || []) {
      const detail = await apiGet<{ source: any }>(`/api/sources/${s.id}`)
      for (const c of detail.source.clips || []) all.push({ ...c, sourceTitle: s.title })
    }
    clips.value = all
  } catch {
    // silencioso
  }
}

async function generate() {
  if (!genClipId.value) return
  generating.value = true
  error.value = ''
  message.value = ''
  try {
    await apiSend('/api/renders', 'POST', {
      clipId: genClipId.value,
      templateId: selectedId.value || undefined,
      variantCount: spec.value.variations.count,
    })
    message.value = 'Render enviado para a fila. Acompanhe em Renders.'
  } catch (e) {
    error.value = errorMessage(e)
  } finally {
    generating.value = false
  }
}

onMounted(async () => {
  await Promise.all([loadTemplates(), loadClips()])
})
</script>
