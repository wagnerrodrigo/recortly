<template>
  <div v-if="loading" class="card p-10 text-center text-sm" style="color: var(--color-muted)">Carregando template...</div>

  <div v-else class="space-y-5">
    <!-- cabeçalho -->
    <div class="card flex flex-wrap items-center gap-3 px-4 py-3">
      <NuxtLink to="/templates" class="btn btn-ghost h-9 w-9 px-0" aria-label="Voltar">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="h-4 w-4"><path d="M19 12H5m0 0 6-6m-6 6 6 6" /></svg>
      </NuxtLink>
      <div class="min-w-0">
        <p class="text-[11px]" style="color: var(--color-muted)">Templates</p>
        <input v-model="spec.name" class="w-44 truncate bg-transparent text-sm font-semibold outline-none" />
      </div>
      <div class="flex flex-1 flex-wrap items-center justify-center gap-1">
        <button
          v-for="(s, i) in steps"
          :key="s"
          class="rounded-lg px-2.5 py-1.5 text-xs font-semibold"
          :class="step === i ? 'text-white' : ''"
          :style="step === i ? 'background: var(--color-brand-500)' : 'color: var(--color-muted)'"
          @click="goTo(i)"
        >
          {{ i + 1 }}. {{ s }}
        </button>
      </div>
      <button class="btn btn-primary h-9 text-xs" :disabled="saving" @click="saveAndContinue">
        {{ saving ? 'Salvando...' : step === steps.length - 1 ? 'Salvar' : 'Salvar e continuar →' }}
      </button>
    </div>

    <p v-if="error" class="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-700">{{ error }}</p>

    <div class="grid gap-5 xl:grid-cols-[minmax(0,1fr)_320px]">
      <section class="card p-5">
        <!-- 1. PERFIL -->
        <div v-if="step === 0" class="space-y-5">
          <div class="flex items-start justify-between gap-3">
            <div>
              <h2 class="font-semibold">Perfil do card</h2>
              <p class="text-xs" style="color: var(--color-muted)">Aparece no topo ou embaixo do vídeo, como um post.</p>
            </div>
            <button class="btn btn-ghost h-8 shrink-0 px-3 text-xs" @click="resetPerfil">Restaurar padrão</button>
          </div>
          <div class="grid gap-4 sm:grid-cols-2">
            <div>
              <label class="label">Nome</label>
              <input v-model="spec.profile.name" class="input" placeholder="testeyoutube" />
            </div>
            <div>
              <label class="label">@handle</label>
              <input v-model="spec.profile.handle" class="input" placeholder="@testeyoutube" />
            </div>
            <div>
              <label class="label">Posição</label>
              <select v-model="spec.profile.position" class="input">
                <option value="top">Topo</option>
                <option value="bottom">Embaixo</option>
                <option value="hidden">Escondido</option>
              </select>
            </div>
            <div>
              <label class="label">Alinhamento</label>
              <select v-model="spec.profile.align" class="input">
                <option value="center">Centralizado</option>
                <option value="left">À esquerda</option>
              </select>
            </div>
          </div>
          <label class="flex items-center gap-2 text-sm">
            <input v-model="spec.profile.verified" type="checkbox" class="h-4 w-4 accent-pink-500" />
            Mostrar selo de verificado
          </label>

          <div class="rounded-xl bg-zinc-50 p-4">
            <div class="flex items-center gap-2">
              <span class="rounded-md px-2 py-0.5 text-[11px] font-bold text-white" style="background: var(--color-brand-500)">AVATAR</span>
              <p class="text-sm font-semibold">Avatar / logo do perfil</p>
            </div>
            <p class="mt-1 text-xs" style="color: var(--color-muted)">
              O círculo do perfil. Sem imagem, aparecem as iniciais do nome.
            </p>
            <div class="mt-3 flex flex-wrap items-center gap-3">
              <input ref="avatarInput" type="file" accept="image/png,image/jpeg,image/webp" class="hidden" @change="uploadAvatar" />
              <div class="flex h-14 w-14 items-center justify-center overflow-hidden rounded-full bg-[#2B2B31] text-sm font-bold text-white">
                <img v-if="spec.profile.avatar" :src="`/api/logos/${spec.profile.avatar}`" alt="avatar" class="h-full w-full object-cover" />
                <span v-else>{{ avatarInitials }}</span>
              </div>
              <button class="btn btn-ghost h-9 text-xs" :disabled="avatarUploading" @click="avatarInput?.click()">
                {{ avatarUploading ? 'Enviando...' : spec.profile.avatar ? 'Trocar imagem' : 'Enviar imagem' }}
              </button>
              <button v-if="spec.profile.avatar" class="text-xs text-red-600" @click="spec.profile.avatar = ''">Remover</button>
            </div>
            <p v-if="avatarError" class="mt-2 rounded-lg bg-red-50 px-3 py-2 text-xs text-red-700">{{ avatarError }}</p>
          </div>

          <div class="rounded-xl bg-zinc-50 p-4">
            <p class="mb-3 text-xs font-semibold uppercase tracking-wide" style="color: var(--color-muted)">Aparência do card</p>
            <div class="grid gap-4 sm:grid-cols-2">
              <div>
                <label class="label">Tamanho do avatar ({{ spec.profile.avatarSize }}px)</label>
                <input v-model.number="spec.profile.avatarSize" type="range" min="90" max="260" class="w-full accent-pink-500" />
              </div>
              <div>
                <label class="label">Fonte do nome ({{ spec.profile.nameSize }}px)</label>
                <input v-model.number="spec.profile.nameSize" type="range" min="28" max="80" class="w-full accent-pink-500" />
              </div>
              <div>
                <label class="label">Fonte do @ ({{ spec.profile.handleSize }}px)</label>
                <input v-model.number="spec.profile.handleSize" type="range" min="24" max="64" class="w-full accent-pink-500" />
              </div>
            </div>
          </div>
        </div>

        <!-- 2. ZONAS -->
        <div v-if="step === 1" class="space-y-5">
          <div>
            <h2 class="font-semibold">Como vai ser o seu layout?</h2>
            <p class="text-xs" style="color: var(--color-muted)">
              As zonas empilham vídeo e espaço vazio. O peso define quanto cada uma ocupa.
            </p>
          </div>

          <div class="grid grid-cols-2 gap-3 sm:grid-cols-4">
            <button
              v-for="p in zonePresets"
              :key="p.id"
              class="rounded-xl border px-2 py-3 text-xs font-semibold transition hover:border-pink-400"
              :style="{ borderColor: 'var(--color-line)' }"
              @click="applyZonePreset(p.id)"
            >
              <span class="block text-lg leading-none" style="color: var(--color-brand-500)">{{ p.icon }}</span>
              <span class="mt-2 block">{{ p.label }}</span>
            </button>
          </div>

          <div class="space-y-3">
            <div
              v-for="(z, i) in spec.zones"
              :key="i"
              class="rounded-xl border p-3"
              :style="{ borderColor: 'var(--color-line)' }"
            >
              <div class="flex items-center gap-3">
                <span class="rounded-md bg-zinc-100 px-2 py-1 text-xs font-semibold">{{ i + 1 }}</span>
                <select v-model="z.type" class="input h-9 max-w-[150px]" @change="onZoneType(z)">
                  <option value="video">Vídeo</option>
                  <option value="image">Imagem</option>
                  <option value="blank">Espaço vazio</option>
                </select>
                <div class="flex-1">
                  <input v-model.number="z.weight" type="range" min="0.1" max="2" step="0.05" class="w-full accent-pink-500" />
                </div>
                <span class="w-12 text-right text-xs" style="color: var(--color-muted)">{{ Math.round(zonePercent(i)) }}%</span>
                <button class="text-xs text-red-600" :disabled="spec.zones.length === 1" @click="removeZone(i)">×</button>
              </div>

              <div v-if="z.type === 'video'" class="mt-3 flex flex-wrap items-center gap-3 pl-10">
                <div class="min-w-[200px] flex-1">
                  <label class="label">Vídeo desta zona</label>
                  <select
                    :value="z.sourceId || 0"
                    class="input h-9"
                    @change="z.sourceId = Number(($event.target as HTMLSelectElement).value) || null"
                  >
                    <option :value="0">Vídeo principal (o corte escolhido)</option>
                    <option v-for="s in otherSources" :key="s.id" :value="s.id">{{ s.title }}</option>
                  </select>
                </div>
                <div class="w-36">
                  <label class="label">Cantos ({{ z.radius }}px)</label>
                  <input v-model.number="z.radius" type="range" min="0" max="48" class="w-full accent-pink-500" />
                </div>
              </div>

              <div v-if="z.type === 'image'" class="mt-3 flex flex-wrap items-center gap-3 pl-10">
                <input
                  :ref="(el) => setImageInput(i, el as HTMLInputElement | null)"
                  type="file"
                  accept="image/png,image/jpeg,image/webp,image/gif"
                  class="hidden"
                  @change="uploadZoneImage(i, $event)"
                />
                <button class="btn btn-ghost h-9 text-xs" :disabled="uploadingZone === i" @click="pickZoneImage(i)">
                  {{ uploadingZone === i ? 'Enviando...' : z.imagePath ? 'Trocar imagem' : 'Enviar imagem' }}
                </button>
                <img v-if="z.imagePath" :src="`/api/images/${z.imagePath}`" alt="" class="h-12 w-12 rounded border object-cover" />
                <button v-if="z.imagePath" class="text-xs text-red-600" @click="z.imagePath = ''">Remover</button>
                <div class="w-36">
                  <label class="label">Ajuste</label>
                  <select v-model="z.fit" class="input h-9">
                    <option value="contain">Caber inteira</option>
                    <option value="cover">Preencher</option>
                  </select>
                </div>
              </div>
            </div>
          </div>

          <div class="flex flex-wrap gap-2">
            <button class="btn btn-ghost text-xs" @click="addZone('video')">+ Zona de vídeo</button>
            <button class="btn btn-ghost text-xs" @click="addZone('image')">+ Zona de imagem</button>
            <button class="btn btn-ghost text-xs" @click="addZone('blank')">+ Espaço vazio</button>
          </div>
          <p v-if="spec.zones.filter((z: any) => z.type === 'video').length === 0" class="text-xs text-red-600">
            O template precisa de pelo menos uma zona de vídeo.
          </p>
        </div>

        <!-- 3. FRASES -->
        <div v-if="step === 2" class="space-y-5">
          <div>
            <h2 class="font-semibold">Frases do projeto</h2>
            <p class="text-xs" style="color: var(--color-muted)">
              Cada variação usa uma frase. Se houver menos frases que variações, elas se repetem em ciclo.
            </p>
          </div>

          <div class="rounded-xl border p-4" :style="{ borderColor: 'var(--color-line)' }">
            <div class="flex items-center gap-2">
              <span class="chip" style="background: var(--color-brand-100); color: var(--color-brand-700)">Recomendado</span>
              <p class="text-sm font-semibold">Gerar automaticamente</p>
            </div>
            <p class="mt-1 text-xs" style="color: var(--color-muted)">
              Descreva a ideia do vídeo. A IA gera headlines curtas para você escolher.
            </p>
            <textarea
              v-model="aiIdea"
              class="input mt-3"
              rows="3"
              maxlength="600"
              placeholder="Ex.: queda de polícia no moto, clube virou no heliporto, NPC bugado no RP"
            />
            <div class="mt-3 flex items-center gap-2">
              <div class="flex items-center gap-2">
                <input v-model.number="aiCount" type="number" min="1" max="10" class="input h-9 w-16" />
                <span class="text-xs" style="color: var(--color-muted)">sugestões</span>
              </div>
              <button class="btn btn-primary ml-auto h-9 text-xs" :disabled="aiLoading || !aiIdea.trim()" @click="generatePhrases">
                {{ aiLoading ? 'Gerando...' : 'Gerar com IA' }}
              </button>
            </div>
            <p v-if="aiHint" class="mt-2 text-[11px]" style="color: var(--color-muted)">{{ aiHint }}</p>
            <p v-if="aiError" class="mt-2 rounded-lg bg-red-50 px-3 py-2 text-xs text-red-700">{{ aiError }}</p>
          </div>

          <div>
            <p class="mb-2 text-xs font-semibold uppercase tracking-wide" style="color: var(--color-muted)">
              Frases do projeto ({{ spec.phrases.length }})
            </p>
            <div v-if="!spec.phrases.length" class="rounded-xl border border-dashed p-6 text-center text-xs" style="border-color: var(--color-line); color: var(--color-muted)">
              Nenhuma frase ainda. Escreva abaixo ou gere automaticamente.
            </div>
            <div v-else class="space-y-2">
              <div v-for="(_, i) in spec.phrases" :key="i" class="flex gap-2">
                <input v-model="spec.phrases[i]" class="input" maxlength="280" :placeholder="`Frase ${i + 1}`" />
                <button class="btn btn-ghost px-3" @click="spec.phrases.splice(i, 1)">×</button>
              </div>
            </div>
          </div>

          <div>
            <label class="label">Escrever frase</label>
            <div class="flex gap-2">
              <input
                v-model="newPhrase"
                class="input"
                maxlength="280"
                placeholder="Ex.: Ninguém tava pronto pra isso"
                @keyup.enter="addPhrase"
              />
              <button class="btn btn-primary px-4" :disabled="!newPhrase.trim()" @click="addPhrase">+</button>
            </div>
            <p class="mt-1 text-right text-[11px]" style="color: var(--color-muted)">{{ newPhrase.length }}/280</p>
          </div>
        </div>

        <!-- 4. MÍDIA -->
        <div v-if="step === 3" class="space-y-5">
          <div>
            <h2 class="font-semibold">Mídia</h2>
            <p class="text-xs" style="color: var(--color-muted)">Escolha o vídeo da biblioteca e marque o corte.</p>
          </div>

          <div>
            <label class="label">Vídeo</label>
            <select v-model.number="sourceId" class="input" @change="loadSource">
              <option :value="0">Selecione um vídeo</option>
              <option v-for="s in sources" :key="s.id" :value="s.id">{{ s.title }}</option>
            </select>
            <p v-if="!sources.length" class="mt-2 text-xs" style="color: var(--color-muted)">
              Nenhum vídeo na biblioteca. <NuxtLink to="/biblioteca" class="font-semibold" style="color: var(--color-brand-600)">Subir agora</NuxtLink>.
            </p>
          </div>

          <div v-if="source" class="grid gap-4 sm:grid-cols-[200px_minmax(0,1fr)]">
            <div class="overflow-hidden rounded-lg bg-black" style="aspect-ratio: 9/16">
              <video :src="source.mediaUrl" class="h-full w-full object-contain" controls playsinline />
            </div>
            <div class="space-y-3">
              <div v-if="clips.length">
                <label class="label">Cortes existentes</label>
                <div class="space-y-2">
                  <label
                    v-for="c in clips"
                    :key="c.id"
                    class="flex cursor-pointer items-center gap-2 rounded-lg border px-3 py-2 text-sm"
                    :style="{ borderColor: clipId === c.id ? 'var(--color-brand-400)' : 'var(--color-line)' }"
                  >
                    <input v-model.number="clipId" type="radio" :value="c.id" class="accent-pink-500" />
                    <span class="truncate">{{ c.name || `Corte ${c.id}` }}</span>
                    <span class="ml-auto text-[11px]" style="color: var(--color-muted)">
                      {{ fmtDuration(c.startMs) }}–{{ fmtDuration(c.endMs) }}
                    </span>
                  </label>
                </div>
              </div>

              <div class="rounded-xl bg-zinc-50 p-3">
                <p class="text-xs font-semibold">Novo corte</p>
                <div class="mt-2 flex items-center justify-between text-[11px]" style="color: var(--color-muted)">
                  <span>Início {{ fmtDuration(startMs) }}</span>
                  <span>Fim {{ fmtDuration(endMs) }} · {{ fmtDuration(endMs - startMs) }}</span>
                </div>
                <input v-model.number="startMs" type="range" min="0" :max="source.durationMs" step="100" class="mt-1 w-full accent-pink-500" />
                <input v-model.number="endMs" type="range" min="0" :max="source.durationMs" step="100" class="mt-1 w-full accent-pink-600" />
                <div class="mt-2 flex gap-2">
                  <input v-model="clipName" class="input h-9" placeholder="Nome do corte" />
                  <button class="btn btn-primary h-9 text-xs" :disabled="creatingClip" @click="createClip">
                    {{ creatingClip ? 'Salvando...' : 'Salvar corte' }}
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 5. ESTILO -->
        <div v-if="step === 4" class="space-y-5">
          <div class="flex items-start justify-between gap-3">
            <div>
              <h2 class="font-semibold">Estilo do texto</h2>
              <p class="text-xs" style="color: var(--color-muted)">Fonte, cor, posição da headline e logo.</p>
            </div>
            <button class="btn btn-ghost h-8 shrink-0 px-3 text-xs" @click="resetEstilo">Restaurar padrão</button>
          </div>
          <div class="grid gap-4 sm:grid-cols-2">
            <div>
              <label class="label">Tamanho ({{ spec.text.size }}px)</label>
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
              <label class="label">Posição da headline</label>
              <select v-model="spec.text.position" class="input">
                <option value="top">Topo</option>
                <option value="bottom">Embaixo</option>
                <option value="hidden">Escondida</option>
              </select>
            </div>
            <div>
              <label class="label">Alinhamento</label>
              <select v-model="spec.text.align" class="input">
                <option value="center">Centralizado</option>
                <option value="left">À esquerda</option>
              </select>
            </div>
            <div>
              <label class="label">Espaçamento lateral ({{ spec.text.sideSpacing }}px)</label>
              <input v-model.number="spec.text.sideSpacing" type="range" min="0" max="360" class="w-full accent-pink-500" />
            </div>
            <div>
              <label class="label">Fundo do canvas</label>
              <div class="flex items-center gap-2">
                <input v-model="spec.canvas.background" type="color" class="h-10 w-12 rounded border" />
                <input v-model="spec.canvas.background" class="input" />
              </div>
            </div>
          </div>

          <div class="rounded-xl bg-zinc-50 p-4">
            <label class="flex items-center gap-2 text-sm font-medium">
              <input v-model="spec.text.band" type="checkbox" class="h-4 w-4 accent-pink-500" />
              Faixa de fundo atrás do texto (estilo meme)
            </label>
            <div v-if="spec.text.band" class="mt-3">
              <label class="label">Cor da faixa</label>
              <input v-model="spec.text.bandColor" type="color" class="h-10 w-16 rounded border" />
            </div>
          </div>

          <div class="rounded-xl bg-zinc-50 p-4">
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

        <!-- 6. VARIAÇÕES -->
        <div v-if="step === 5" class="space-y-5">
          <div class="flex items-start justify-between gap-3">
            <div>
              <h2 class="font-semibold">Variações</h2>
              <p class="text-xs" style="color: var(--color-muted)">
                Cada geração usa uma seed nova, então as variações nunca saem iguais.
              </p>
            </div>
            <button class="btn btn-ghost h-8 shrink-0 px-3 text-xs" @click="resetVariacoes">Restaurar padrão</button>
          </div>
          <div class="grid gap-4 sm:grid-cols-3">
            <div>
              <label class="label">Quantidade ({{ spec.variations.count }})</label>
              <input v-model.number="spec.variations.count" type="range" min="1" max="30" class="w-full accent-pink-500" />
            </div>
            <div>
              <label class="label">Seed base</label>
              <input v-model.number="spec.variations.seedBase" type="number" class="input" />
            </div>
            <label class="flex items-end gap-2 pb-2 text-sm">
              <input v-model="spec.variations.mirror" type="checkbox" class="h-4 w-4 accent-pink-500" />
              Espelhar algumas
            </label>
          </div>

          <div class="grid gap-4 sm:grid-cols-2">
            <div v-for="field in rangeFields" :key="field.key">
              <label class="label">{{ field.label }} ({{ spec.variations[field.key][0] }} – {{ spec.variations[field.key][1] }})</label>
              <div class="flex items-center gap-2">
                <input
                  v-model.number="spec.variations[field.key][0]"
                  type="number"
                  :step="field.step"
                  :min="field.min"
                  :max="field.max"
                  class="input"
                />
                <input
                  v-model.number="spec.variations[field.key][1]"
                  type="number"
                  :step="field.step"
                  :min="field.min"
                  :max="field.max"
                  class="input"
                />
              </div>
            </div>
          </div>
          <p class="text-[11px]" style="color: var(--color-muted)">
            Zoom acima de 1 aproxima, abaixo de 1 afasta. Velocidade em torno de 1 mantém o áudio natural.
          </p>
        </div>

        <!-- 7. GERAR -->
        <div v-if="step === 6" class="space-y-5">
          <div>
            <h2 class="font-semibold">Gerar</h2>
            <p class="text-xs" style="color: var(--color-muted)">Confira o resumo e envie para a fila de render.</p>
          </div>

          <div class="grid gap-3 sm:grid-cols-3">
            <div class="rounded-xl bg-zinc-50 p-3">
              <p class="text-[11px]" style="color: var(--color-muted)">Vídeo</p>
              <p class="truncate text-sm font-semibold">{{ source?.title || 'nenhum selecionado' }}</p>
            </div>
            <div class="rounded-xl bg-zinc-50 p-3">
              <p class="text-[11px]" style="color: var(--color-muted)">Corte</p>
              <p class="text-sm font-semibold">
                {{ selectedClip ? fmtDuration(selectedClip.startMs) + '–' + fmtDuration(selectedClip.endMs) : 'nenhum' }}
              </p>
            </div>
            <div class="rounded-xl bg-zinc-50 p-3">
              <p class="text-[11px]" style="color: var(--color-muted)">Variações</p>
              <p class="text-sm font-semibold">{{ spec.variations.count }}</p>
            </div>
          </div>

          <div class="flex flex-wrap items-center gap-2">
            <input v-model.number="spec.variations.count" type="number" min="1" max="30" class="input h-10 w-20" />
            <button class="btn btn-primary" :disabled="!clipId || generating" @click="generate">
              {{ generating ? 'Enviando...' : 'Gerar variações' }}
            </button>
            <button
              v-if="render"
              class="btn btn-ghost"
              :disabled="generating"
              @click="generate"
            >Gerar novo conjunto</button>
            <NuxtLink v-if="render?.status === 'done'" to="/renders" class="btn btn-ghost">Ver todos em Renders</NuxtLink>
          </div>

          <div v-if="render" class="rounded-xl border p-4" :style="{ borderColor: 'var(--color-line)' }">
            <div class="flex items-center justify-between">
              <p class="text-sm font-semibold">Render #{{ render.id }}</p>
              <span :class="statusChip(render.status)">{{ render.status }}</span>
            </div>
            <div class="mt-2 h-2 overflow-hidden rounded-full bg-zinc-100">
              <div class="h-full transition-all" :style="{ width: render.progress + '%', background: 'var(--color-brand-500)' }" />
            </div>
            <p v-if="render.error" class="mt-2 text-xs text-red-600">{{ render.error }}</p>

            <div v-if="render.outputs?.length" class="mt-3">
              <a :href="`/api/renders/${render.id}/zip`" :download="`render-${render.id}.zip`" class="btn btn-ghost h-8 text-xs">
                Baixar tudo (.zip)
              </a>
            </div>
            <div v-if="render.outputs?.length" class="mt-4 grid gap-3 sm:grid-cols-3 lg:grid-cols-4">
              <div v-for="(o, oi) in render.outputs" :key="o.id">
                <div class="overflow-hidden rounded-lg bg-black" style="aspect-ratio: 9/16">
                  <video :src="o.mediaUrl" class="h-full w-full object-contain" controls preload="metadata" playsinline />
                </div>
                <p class="mt-1 line-clamp-2 text-[11px]" style="color: var(--color-muted)">{{ o.headline }}</p>
                <div class="mt-1 flex gap-1">
                  <a
                    :href="o.mediaUrl"
                    :download="videoFileName(o.headline, oi + 1, render.outputs.length)"
                    class="btn btn-ghost h-7 flex-1 px-1 text-[11px]"
                  >Baixar</a>
                  <NuxtLink :to="`/publicacoes?output=${o.id}`" class="btn btn-primary h-7 flex-1 px-1 text-[11px]">Publicar</NuxtLink>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      <aside class="space-y-3 xl:sticky xl:top-20 xl:h-fit">
        <div class="card p-4">
          <p class="mb-3 text-xs font-semibold uppercase tracking-wide" style="color: var(--color-muted)">Preview</p>
          <TemplatePreview
            :spec="spec"
            :layout="layout"
            :headline="previewHeadline"
            :media-url="source?.mediaUrl"
            :media-by-id="mediaById"
          />
          <p class="mt-3 text-center text-[11px]" style="color: var(--color-muted)">
            Aproximado. O arquivo final é renderizado pelo FFmpeg.
          </p>
        </div>
        <div class="card p-4 text-xs" style="color: var(--color-muted)">
          <p class="font-semibold" style="color: var(--color-ink)">Resumo</p>
          <p class="mt-1">{{ spec.zones.length }} zona(s) · {{ spec.variations.count }} variação(ões)</p>
          <p class="mt-1 truncate">{{ selectedClip ? 'Corte ' + fmtDuration(selectedClip.startMs) + '–' + fmtDuration(selectedClip.endMs) : 'Sem corte selecionado' }}</p>
          <p class="mt-1">{{ source?.title || 'Sem vídeo selecionado' }}</p>
        </div>
      </aside>
    </div>
  </div>
</template>

<script setup lang="ts">
const route = useRoute()
const clipFromQuery = Number(route.query.clip || 0)
const steps = ['Perfil', 'Zonas', 'Frases', 'Mídia', 'Estilo', 'Variações', 'Gerar']

const templateId = computed(() => Number(route.params.id))
const loading = ref(true)
const saving = ref(false)
const error = ref('')
const step = ref(0)

const spec = ref<any>(defaultSpec())
const layout = ref<any>(null)
const sources = ref<any[]>([])
const sourceId = ref(0)
const source = ref<any>(null)
const clipId = ref(0)
const startMs = ref(0)
const endMs = ref(30000)
const clipName = ref('')
const creatingClip = ref(false)

const aiIdea = ref('')
const aiCount = ref(5)
const aiLoading = ref(false)
const aiError = ref('')
const aiHint = ref('')
const newPhrase = ref('')
const avatarInput = ref<HTMLInputElement | null>(null)
const avatarUploading = ref(false)
const avatarError = ref('')
const avatarInitials = computed(() => {
  const name = spec.value.profile?.name || 'P'
  return name.split(' ').map((p: string) => p[0]).slice(0, 2).join('').toUpperCase()
})

const uploadingZone = ref(-1)
const zoneImageInputs = new Map<number, HTMLInputElement>()

const otherSources = computed(() => sources.value)
const mediaById = computed<Record<string, string>>(() => {
  const map: Record<string, string> = {}
  for (const s of sources.value) map[String(s.id)] = s.mediaUrl
  return map
})

const render = ref<any>(null)
const generating = ref(false)
let pollTimer: ReturnType<typeof setInterval> | null = null

const rangeFields = [
  { key: 'zoom', label: 'Zoom', step: 0.01, min: 1, max: 2 },
  { key: 'saturation', label: 'Saturação', step: 0.01, min: 0.5, max: 2 },
  { key: 'brightness', label: 'Brilho', step: 0.01, min: 0.8, max: 1.2 },
  { key: 'speed', label: 'Velocidade', step: 0.01, min: 0.5, max: 2 },
  { key: 'offsetX', label: 'Recorte horizontal', step: 0.01, min: 0, max: 1 },
  { key: 'offsetY', label: 'Recorte vertical', step: 0.01, min: 0, max: 1 },
]

const clips = computed<any[]>(() => source.value?.clips || [])
const selectedClip = computed(() => clips.value.find((c: any) => c.id === clipId.value))
const previewHeadline = computed(() => spec.value.phrases.find((p: string) => p.trim()) || '')

function zonePercent(i: number) {
  const total = spec.value.zones.reduce((acc: number, z: any) => acc + (Number(z.weight) || 0), 0)
  if (!total) return 0
  return (spec.value.zones[i].weight / total) * 100
}

function applyZonePreset(id: string) {
  const preset = zonePresets.find((p) => p.id === id)
  if (!preset) return
  spec.value.zones = preset.zones.map((z) => ({ ...z }))
}

function addZone(type: 'video' | 'image' | 'blank') {
  spec.value.zones.push({
    type,
    weight: 0.5,
    radius: type === 'blank' ? 0 : 16,
    fit: type === 'image' ? 'contain' : 'cover',
    sourceId: null,
    imagePath: '',
  })
}

function onZoneType(z: any) {
  if (z.type !== 'video') z.sourceId = null
  if (z.type !== 'image') z.imagePath = ''
  if (z.type === 'image' && !z.fit) z.fit = 'contain'
}

function setImageInput(i: number, el: HTMLInputElement | null) {
  if (el) zoneImageInputs.set(i, el)
  else zoneImageInputs.delete(i)
}

function pickZoneImage(i: number) {
  zoneImageInputs.get(i)?.click()
}

async function uploadZoneImage(i: number, ev: Event) {
  const input = ev.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  uploadingZone.value = i
  error.value = ''
  try {
    const form = new FormData()
    form.append('image', file)
    const res = await $fetch<{ path: string }>('/api/uploads/image', { method: 'POST', body: form })
    spec.value.zones[i].imagePath = res.path
    spec.value.zones[i].type = 'image'
  } catch (e) {
    error.value = errorMessage(e)
  } finally {
    uploadingZone.value = -1
    input.value = ''
  }
}

function clone<T>(v: T): T {
  return JSON.parse(JSON.stringify(v))
}

function resetPerfil() {
  spec.value.profile = clone(defaultSpec().profile)
}

function resetEstilo() {
  const d = defaultSpec()
  spec.value.text = clone(d.text)
  spec.value.canvas = clone(d.canvas)
}

function resetVariacoes() {
  spec.value.variations = clone(defaultSpec().variations)
}

function removeZone(i: number) {
  if (spec.value.zones.length === 1) return
  spec.value.zones.splice(i, 1)
}

async function refreshLayout() {
  try {
    const res = await apiSend<{ layout: any }>('/api/templates/layout', 'POST', { spec: spec.value })
    layout.value = res.layout
  } catch {
    // mantém o último layout válido
  }
}

let layoutTimer: ReturnType<typeof setTimeout> | null = null
watch(
  spec,
  () => {
    if (layoutTimer) clearTimeout(layoutTimer)
    layoutTimer = setTimeout(refreshLayout, 350)
  },
  { deep: true },
)

async function load() {
  loading.value = true
  try {
    const [tpl, srcs, ai] = await Promise.all([
      apiGet<{ template: any }>(`/api/templates/${templateId.value}`),
      apiGet<{ sources: any[] }>('/api/sources'),
      apiGet<{ configured: boolean; model: string; baseUrl: string }>('/api/ai/status').catch(() => null),
    ])
    spec.value = { ...defaultSpec(), ...tpl.template.spec }
    if (!spec.value.logo) spec.value.logo = defaultSpec().logo
    sources.value = srcs.sources || []
    if (ai?.configured) {
      aiHint.value = `IA local: ${ai.model} em ${ai.baseUrl}`
    } else {
      aiHint.value = 'IA não configurada. Defina SMASHCUT_AI_BASE_URL e SMASHCUT_AI_MODEL no .env para gerar frases.'
    }
    await refreshLayout()
    if (sources.value.length) {
      sourceId.value = sources.value[0].id
      await loadSource()
      // Continuidade: se veio do editor com um corte, seleciona o vídeo e o corte.
      if (clipFromQuery) {
        for (const s of sources.value) {
          const detail = await apiGet<{ source: any }>(`/api/sources/${s.id}`)
          const found = (detail.source.clips || []).find((c: any) => c.id === clipFromQuery)
          if (found) {
            sourceId.value = s.id
            source.value = detail.source
            clipId.value = found.id
            break
          }
        }
      }
    }
  } catch (e) {
    error.value = errorMessage(e)
  } finally {
    loading.value = false
  }
}

async function loadSource() {
  if (!sourceId.value) {
    source.value = null
    return
  }
  const res = await apiGet<{ source: any }>(`/api/sources/${sourceId.value}`)
  source.value = res.source
  endMs.value = Math.min(res.source.durationMs, 30000)
  startMs.value = 0
  if (res.source.clips?.length) clipId.value = res.source.clips[0].id
  else clipId.value = 0
}

async function createClip() {
  if (!source.value) return
  creatingClip.value = true
  error.value = ''
  try {
    const res = await apiSend<{ clip: any }>('/api/clips', 'POST', {
      sourceId: sourceId.value,
      name: clipName.value || `Corte ${(source.value.clips?.length || 0) + 1}`,
      startMs: startMs.value,
      endMs: Math.max(startMs.value + 1000, endMs.value),
      layout: 'blur',
      zoom: 1,
    })
    clipName.value = ''
    await loadSource()
    clipId.value = res.clip.id
  } catch (e) {
    error.value = errorMessage(e)
  } finally {
    creatingClip.value = false
  }
}

function addPhrase() {
  const value = newPhrase.value.trim()
  if (!value || spec.value.phrases.length >= 30) return
  spec.value.phrases.push(value.slice(0, 280))
  newPhrase.value = ''
}

async function uploadAvatar(ev: Event) {
  const file = (ev.target as HTMLInputElement).files?.[0]
  if (!file) return
  avatarUploading.value = true
  avatarError.value = ''
  try {
    const form = new FormData()
    form.append('logo', file)
    const res = await $fetch<{ path: string }>('/api/uploads/logo', { method: 'POST', body: form })
    spec.value.profile.avatar = res.path
  } catch (e) {
    avatarError.value = errorMessage(e)
  } finally {
    avatarUploading.value = false
    if (avatarInput.value) avatarInput.value.value = ''
  }
}

async function generatePhrases() {
  aiLoading.value = true
  aiError.value = ''
  try {
    const res = await apiSend<{ phrases: string[] }>('/api/ai/phrases', 'POST', {
      idea: aiIdea.value,
      count: aiCount.value,
      platform: 'vertical',
    })
    const cleaned = (res.phrases || []).filter((p: string) => p.trim())
    spec.value.phrases = cleaned
    if (!cleaned.length) aiError.value = 'A IA não retornou frases.'
  } catch (e) {
    aiError.value = errorMessage(e)
  } finally {
    aiLoading.value = false
  }
}

async function save(): Promise<boolean> {
  saving.value = true
  error.value = ''
  try {
    if (!spec.value.phrases.some((p: string) => p.trim())) {
      throw new Error('Informe pelo menos uma frase antes de salvar.')
    }
    if (!spec.value.zones.some((z: any) => z.type === 'video')) {
      throw new Error('O template precisa de pelo menos uma zona de vídeo.')
    }
    await apiSend(`/api/templates/${templateId.value}`, 'PUT', { name: spec.value.name, spec: spec.value })
    return true
  } catch (e) {
    error.value = errorMessage(e)
    return false
  } finally {
    saving.value = false
  }
}

async function goTo(i: number) {
  if (i <= step.value) {
    step.value = i
    return
  }
  const ok = await save()
  if (ok) step.value = i
}

async function saveAndContinue() {
  const ok = await save()
  if (!ok) return
  if (step.value < steps.length - 1) step.value++
}

async function generate() {
  if (!clipId.value) {
    error.value = 'Selecione ou crie um corte na etapa Mídia.'
    return
  }
  generating.value = true
  error.value = ''
  try {
    const ok = await save()
    if (!ok) return
    const res = await apiSend<{ render: any }>('/api/renders', 'POST', {
      clipId: clipId.value,
      templateId: templateId.value,
      variantCount: spec.value.variations.count,
    })
    render.value = res.render
    startPolling()
  } catch (e) {
    error.value = errorMessage(e)
  } finally {
    generating.value = false
  }
}

function startPolling() {
  if (pollTimer) clearInterval(pollTimer)
  pollTimer = setInterval(async () => {
    if (!render.value) return
    try {
      const res = await apiGet<{ render: any }>(`/api/renders/${render.value.id}`)
      render.value = res.render
      if (['done', 'failed', 'cancelled'].includes(res.render.status)) {
        if (pollTimer) clearInterval(pollTimer)
        pollTimer = null
      }
    } catch {
      // mantém o estado atual
    }
  }, 1800)
}

onMounted(load)
onBeforeUnmount(() => {
  if (pollTimer) clearInterval(pollTimer)
})
</script>
