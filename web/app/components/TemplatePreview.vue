<template>
  <div
    class="relative w-full overflow-hidden rounded-xl"
    :style="{ aspectRatio: '9 / 16', background: spec.canvas.background, containerType: 'size' }"
  >
    <!-- zonas -->
    <div
      v-for="z in layout?.zones || []"
      :key="z.index"
      class="absolute overflow-hidden"
      :style="zoneStyle(z)"
    >
      <template v-if="z.type === 'video'">
        <video
          v-if="mediaUrl"
          :src="mediaUrl"
          class="h-full w-full object-cover"
          :style="{ transform: `scale(${1.02})` }"
          muted
          loop
          autoplay
          playsinline
        />
        <div v-else class="flex h-full w-full items-center justify-center bg-zinc-800">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" class="h-6 w-6 text-white/40">
            <path d="m10 8 6 4-6 4V8Z" />
          </svg>
        </div>
      </template>
      <div v-else class="h-full w-full" style="background: #1a1a1e" />
    </div>

    <!-- perfil -->
    <template v-if="spec.profile.position !== 'hidden'">
      <div
        class="absolute flex items-center gap-2"
        :style="profileStyle"
      >
        <div
          class="flex shrink-0 items-center justify-center rounded-full bg-[#2B2B31] font-bold text-white"
          :style="avatarStyle"
        >
          {{ initials }}
        </div>
        <div :class="spec.profile.align === 'left' ? 'text-left' : 'text-center'">
          <p
            class="font-bold leading-tight"
            :style="{ color: spec.text.color, fontSize: font(spec.profile.nameSize) }"
          >
            {{ spec.profile.name || 'Seu perfil' }}
            <span v-if="spec.profile.verified" style="color: #3897f0">✓</span>
          </p>
          <p
            class="leading-tight"
            :style="{ color: spec.text.color, opacity: 0.72, fontSize: font(spec.profile.handleSize) }"
          >
            {{ spec.profile.handle }}
          </p>
        </div>
      </div>
    </template>

    <!-- logo -->
    <img
      v-if="spec.logo?.enabled && spec.logo.path && !logoError"
      :src="`/api/logos/${spec.logo.path}`"
      alt="logo"
      :style="logoStyle"
      @error="logoError = true"
      @load="logoError = false"
    />

    <!-- headline -->
    <p
      v-if="spec.text.position !== 'hidden'"
      class="absolute inset-x-[6%] font-bold leading-tight"
      :style="headlineStyle"
    >
      {{ headline || 'Sua headline' }}
    </p>
  </div>
</template>

<script setup lang="ts">
const props = defineProps<{
  spec: any
  layout: any
  headline?: string
  mediaUrl?: string
}>()

// 1 unit = 1080 px do canvas. Usa cqw para escalar com o container.
const font = (px: number) => `calc(${(px / 1080) * 100}cqw)`
const pctW = (px: number) => `${(px / 1080) * 100}%`
const pctH = (px: number) => `${(px / 1920) * 100}%`

const logoError = ref(false)

const logoStyle = computed(() => {
  const l = props.spec.logo
  const m = (l.margin / 1080) * 100
  const base: Record<string, string> = {
    position: 'absolute',
    width: `${(l.size / 1080) * 100}%`,
    height: 'auto',
    opacity: String(l.opacity ?? 1),
    zIndex: '5',
  }
  if (l.position === 'top-left' || l.position === 'bottom-left') base.left = `${m}cqw`
  else base.right = `${m}cqw`
  if (l.position === 'top-left' || l.position === 'top-right') base.top = `${(l.margin / 1920) * 100}%`
  else base.bottom = `${(l.margin / 1920) * 100}%`
  return base
})

const initials = computed(() => {
  const name = props.spec?.profile?.name || 'P'
  return name.split(' ').map((p: string) => p[0]).slice(0, 2).join('').toUpperCase()
})

function zoneStyle(z: any) {
  return {
    left: pctW(z.x),
    top: pctH(z.y),
    width: pctW(z.w),
    height: pctH(z.h),
    borderRadius: `${(z.radius / 1080) * 100}cqw`,
  }
}

const avatarStyle = computed(() => {
  const s = props.spec.profile.avatarSize
  return {
    width: pctW(s),
    height: pctW(s),
    fontSize: font(s * 0.42),
    borderRadius: '50%',
  }
})

const profileStyle = computed(() => {
  const p = props.spec.profile
  const l = props.layout || {}
  const center = p.align !== 'left'
  return {
    left: center ? '6%' : pctW(40),
    right: center ? '6%' : undefined,
    top: pctH(l.profileY ?? 40),
    justifyContent: center ? 'center' : 'flex-start',
    flexDirection: center ? 'column' : 'row',
    textAlign: center ? 'center' : 'left',
  }
})

const headlineStyle = computed(() => {
  const t = props.spec.text
  const l = props.layout || {}
  return {
    top: pctH((l.textY ?? 0) + (t.offset || 0)),
    color: t.color,
    fontSize: font(t.size),
    textAlign: t.align === 'left' ? 'left' : 'center',
    WebkitTextStroke: t.strokeEnabled && !t.band ? `${(t.strokeWidth / 1080) * 100}cqw ${t.strokeColor}` : 'none',
    background: t.band ? `${t.bandColor}e6` : 'transparent',
    padding: t.band ? `${(t.size / 3 / 1080) * 100}cqw 0` : '0',
  }
})
</script>
