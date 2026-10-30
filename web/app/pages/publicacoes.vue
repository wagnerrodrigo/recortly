<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-2xl font-bold tracking-tight">Publicações</h1>
      <p class="mt-1 text-sm" style="color: var(--color-muted)">
        Publique ou agende um render. O crédito da origem é adicionado automaticamente na descrição.
      </p>
    </div>

    <section class="card p-5">
      <h2 class="font-semibold">Nova publicação</h2>
      <div class="mt-3 grid gap-3 sm:grid-cols-2">
        <div class="sm:col-span-2">
          <label class="label">Variação</label>
          <select v-model="form.outputId" class="input">
            <option :value="0">Selecione um render pronto</option>
            <option v-for="o in outputs" :key="o.id" :value="o.id">
              Render #{{ o.renderId }} · variação {{ o.variantIdx + 1 }} · {{ o.headline || 'sem headline' }}
            </option>
          </select>
        </div>
        <div>
          <label class="label">Plataforma</label>
          <select v-model="form.platform" class="input">
            <option value="youtube">YouTube Shorts</option>
            <option value="instagram">Instagram Reels</option>
            <option value="tiktok">TikTok</option>
          </select>
        </div>
        <div>
          <label class="label">Quando</label>
          <select v-model="when" class="input">
            <option value="now">Agora</option>
            <option value="schedule">Agendar</option>
            <option value="draft">Salvar como rascunho</option>
          </select>
        </div>
        <div v-if="when === 'schedule'" class="sm:col-span-2">
          <label class="label">Data e hora</label>
          <input v-model="form.scheduledAt" type="datetime-local" class="input" />
        </div>
        <div class="sm:col-span-2">
          <label class="label">Título</label>
          <input v-model="form.title" class="input" maxlength="100" placeholder="Título do Short/Reel" />
        </div>
        <div class="sm:col-span-2">
          <label class="label">Descrição</label>
          <textarea v-model="form.description" class="input" rows="3" placeholder="Inclua #Shorts e o contexto do corte" />
        </div>
      </div>
      <button class="btn btn-primary mt-3" :disabled="!form.outputId || sending" @click="submit">
        {{ sending ? 'Enviando...' : when === 'schedule' ? 'Agendar' : when === 'draft' ? 'Salvar' : 'Publicar agora' }}
      </button>
      <p v-if="error" class="mt-3 rounded-lg bg-red-50 px-3 py-2 text-sm text-red-700">{{ error }}</p>
      <p v-if="message" class="mt-3 rounded-lg bg-green-50 px-3 py-2 text-sm text-green-700">{{ message }}</p>
    </section>

    <section class="card">
      <header class="flex items-center justify-between border-b px-5 py-3" style="border-color: var(--color-line)">
        <h2 class="font-semibold">Histórico</h2>
        <button class="text-sm font-semibold" style="color: var(--color-brand-600)" @click="loadPosts">Atualizar</button>
      </header>
      <div v-if="!posts.length" class="p-6 text-sm" style="color: var(--color-muted)">Nenhuma publicação ainda.</div>
      <ul v-else class="divide-y">
        <li v-for="p in posts" :key="p.id" class="flex flex-wrap items-center gap-3 px-5 py-3">
          <span class="chip chip-neutral">{{ platformLabels[p.platform] || p.platform }}</span>
          <div class="min-w-0 flex-1">
            <p class="truncate text-sm font-medium">{{ p.title || '(sem título)' }}</p>
            <p class="text-xs" style="color: var(--color-muted)">
              {{ p.status === 'scheduled' ? 'Agendado para ' + fmtDate(p.scheduledAt) : fmtDate(p.publishedAt || p.createdAt) }}
            </p>
            <p v-if="p.error" class="text-xs text-red-600">{{ p.error }}</p>
          </div>
          <span :class="statusChip(p.status)">{{ p.status }}</span>
          <a v-if="p.externalUrl" :href="p.externalUrl" target="_blank" rel="noopener" class="text-xs font-semibold" style="color: var(--color-brand-600)">Abrir</a>
        </li>
      </ul>
    </section>
  </div>
</template>

<script setup lang="ts">
const route = useRoute()
const outputs = ref<any[]>([])
const posts = ref<any[]>([])
const sending = ref(false)
const error = ref('')
const message = ref('')
const when = ref<'now' | 'schedule' | 'draft'>('now')

const form = reactive({
  outputId: 0,
  platform: 'youtube',
  title: '',
  description: '',
  scheduledAt: '',
})

async function loadOutputs() {
  const res = await apiGet<{ renders: any[] }>('/api/renders')
  const all: any[] = []
  for (const r of res.renders || []) {
    if (r.status !== 'done') continue
    for (const o of r.outputs || []) all.push(o)
  }
  outputs.value = all
  const q = Number(route.query.output || 0)
  if (q) form.outputId = q
}

async function loadPosts() {
  const res = await apiGet<{ posts: any[] }>('/api/posts')
  posts.value = res.posts || []
}

async function submit() {
  sending.value = true
  error.value = ''
  message.value = ''
  try {
    let scheduledAt: string | undefined
    if (when.value === 'schedule') {
      if (!form.scheduledAt) throw new Error('Informe a data e hora')
      scheduledAt = new Date(form.scheduledAt).toISOString()
    }
    await apiSend('/api/posts', 'POST', {
      outputId: form.outputId,
      platform: form.platform,
      title: form.title,
      description: form.description,
      scheduledAt,
      publishNow: when.value === 'now',
    })
    message.value = when.value === 'draft' ? 'Rascunho salvo.' : 'Publicação enviada para a fila.'
    form.title = ''
    form.description = ''
    await loadPosts()
  } catch (e) {
    error.value = errorMessage(e)
  } finally {
    sending.value = false
  }
}

onMounted(async () => {
  try {
    await Promise.all([loadOutputs(), loadPosts()])
  } catch (e) {
    error.value = errorMessage(e)
  }
})
</script>
