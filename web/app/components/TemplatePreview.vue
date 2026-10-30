<template>
  <div class="relative w-full overflow-hidden rounded-xl" :style="{ aspectRatio: '9 / 16', background: spec.canvas.background }">
    <!-- perfil -->
    <div class="absolute inset-x-0 top-[9%] flex flex-col items-center px-4 text-center">
      <div class="h-[8%] w-auto">
        <div
          class="flex items-center justify-center rounded-full border-2 text-[10px] font-bold"
          :style="{ width: avatarSize + 'px', height: avatarSize + 'px', borderColor: spec.text.color + '33', background: 'rgba(255,255,255,0.08)', color: spec.text.color }"
        >
          {{ initials }}
        </div>
      </div>
      <p class="mt-2 max-w-full truncate font-bold" :style="{ color: spec.text.color, fontSize: nameSize + 'px' }">
        {{ spec.profile.name || 'Seu perfil' }}
        <span v-if="spec.profile.verified" :style="{ color: '#3897F0' }">✓</span>
      </p>
      <p v-if="spec.text.showHandle" class="max-w-full truncate" :style="{ color: spec.text.color, opacity: 0.7, fontSize: handleSize + 'px' }">
        {{ spec.profile.handle }}
      </p>
    </div>

    <!-- headline -->
    <p
      class="absolute inset-x-0 px-4 text-center font-bold leading-tight"
      :style="{
        color: spec.text.color,
        fontSize: headlineSize + 'px',
        top: headlineTop + '%',
        WebkitTextStroke: spec.text.strokeEnabled ? strokeSize + 'px ' + spec.text.strokeColor : 'none',
      }"
    >
      {{ headline || 'Sua headline' }}
    </p>

    <!-- área do vídeo -->
    <div
      class="absolute left-[6%] right-[6%] overflow-hidden rounded-md bg-zinc-800"
      :style="{ top: videoTop + '%', height: videoHeight + '%' }"
    >
      <video v-if="mediaUrl" :src="mediaUrl" class="h-full w-full object-cover" muted loop autoplay playsinline />
      <div v-else class="flex h-full items-center justify-center text-[10px] text-white/40">vídeo 9:16</div>
    </div>
  </div>
</template>

<script setup lang="ts">
const props = defineProps<{
  spec: any
  headline?: string
  mediaUrl?: string
}>()

// O canvas do template é 1080x1920; aqui convertemos para % do card.
const pct = (px: number, total: number) => Math.round((px / total) * 1000) / 10

const nameSize = computed(() => Math.max(9, pct(props.spec.text.size * 0.86, 1080) * 10))
const handleSize = computed(() => Math.max(7, nameSize.value * 0.72))
const headlineSize = computed(() => Math.max(9, pct(props.spec.text.size, 1080) * 10))
const strokeSize = computed(() => Math.max(0.5, props.spec.text.strokeWidth * 0.1))
const avatarSize = computed(() => Math.max(16, nameSize.value * 1.8))
const headlineTop = computed(() => pct(220 + props.spec.text.offsetFromProfile, 1920))
const videoTop = computed(() => 11 + pct(220 + props.spec.text.offsetFromProfile + props.spec.text.size * 2 + 40, 1920))
const videoHeight = computed(() => Math.max(22, 90 - videoTop.value))
const initials = computed(() => {
  const name = props.spec.profile.name || 'P'
  return name.split(' ').map((p: string) => p[0]).slice(0, 2).join('').toUpperCase()
})
</script>
