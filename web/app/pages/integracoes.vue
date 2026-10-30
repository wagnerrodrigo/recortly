<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-2xl font-bold tracking-tight">Integrações</h1>
      <p class="mt-1 text-sm" style="color: var(--color-muted)">
        Conecte apenas contas próprias. O SmashCut usa as APIs oficiais de cada plataforma.
      </p>
    </div>

    <p v-if="connected" class="rounded-lg bg-green-50 px-3 py-2 text-sm text-green-700">
      Conta do {{ platformLabels[connected] }} conectada.
    </p>
    <p v-if="error" class="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-700">{{ error }}</p>

    <div class="grid gap-4 md:grid-cols-3">
      <article v-for="p in platforms" :key="p.id" class="card flex flex-col p-5">
        <div class="flex items-center justify-between">
          <h2 class="font-semibold">{{ p.name }}</h2>
          <span :class="accountFor(p.id) ? 'chip chip-ok' : 'chip chip-neutral'">
            {{ accountFor(p.id) ? 'conectado' : 'não conectado' }}
          </span>
        </div>
        <p class="mt-2 flex-1 text-sm" style="color: var(--color-muted)">{{ p.description }}</p>
        <button
          v-if="p.id === 'youtube'"
          class="btn btn-primary mt-4"
          :disabled="connecting"
          @click="connectYouTube"
        >
          {{ accountFor('youtube') ? 'Reconectar' : 'Conectar YouTube' }}
        </button>
        <p v-else class="mt-4 rounded-lg bg-amber-50 px-3 py-2 text-xs text-amber-800">
          {{ p.note }}
        </p>
      </article>
    </div>

    <section class="card p-5">
      <h2 class="font-semibold">Como configurar</h2>
      <ol class="mt-3 list-decimal space-y-2 pl-5 text-sm" style="color: var(--color-muted)">
        <li>Crie um projeto no Google Cloud e ative a YouTube Data API v3.</li>
        <li>Crie credenciais OAuth do tipo "Web application".</li>
        <li>Adicione <code class="rounded bg-zinc-100 px-1">http://localhost:3000/api/oauth/youtube/callback</code> como URI de redirecionamento.</li>
        <li>Preencha <code class="rounded bg-zinc-100 px-1">SMASHCUT_YOUTUBE_CLIENT_ID</code> e <code class="rounded bg-zinc-100 px-1">SMASHCUT_YOUTUBE_CLIENT_SECRET</code> no <code class="rounded bg-zinc-100 px-1">.env</code> e reinicie a API.</li>
        <li>Instagram e TikTok exigem app aprovado nas plataformas antes de publicar de verdade.</li>
      </ol>
    </section>
  </div>
</template>

<script setup lang="ts">
const route = useRoute()
const accounts = ref<any[]>([])
const error = ref('')
const connecting = ref(false)

const connected = computed(() => (typeof route.query.connected === 'string' ? route.query.connected : ''))

const platforms = [
  { id: 'youtube', name: 'YouTube Shorts', description: 'Upload direto na sua conta via YouTube Data API.', note: '' },
  {
    id: 'instagram',
    name: 'Instagram Reels',
    description: 'Reels via Instagram Graph API.',
    note: 'Requer conta Business/Creator e app aprovado na Meta.',
  },
  {
    id: 'tiktok',
    name: 'TikTok',
    description: 'Vídeos via Content Posting API.',
    note: 'Apps não auditados só publicam com visibilidade privada.',
  },
]

function accountFor(platform: string) {
  return accounts.value.find((a) => a.platform === platform && a.connected)
}

async function load() {
  try {
    const res = await apiGet<{ accounts: any[] }>('/api/accounts')
    accounts.value = res.accounts || []
  } catch (e) {
    error.value = errorMessage(e)
  }
}

async function connectYouTube() {
  connecting.value = true
  error.value = ''
  try {
    const res = await apiGet<{ url: string }>('/api/oauth/youtube/start')
    window.location.href = res.url
  } catch (e) {
    error.value = errorMessage(e)
  } finally {
    connecting.value = false
  }
}

onMounted(load)
</script>
