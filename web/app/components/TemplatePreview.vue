<template>
  <div>
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
            v-if="zoneMedia(z)"
            ref="videos"
            :src="zoneMedia(z)"
            class="h-full w-full object-cover"
            muted
            loop
            playsinline
            preload="metadata"
            @timeupdate="onTime"
            @loadedmetadata="onMeta"
            @durationchange="onMeta"
          />
          <div v-else class="flex h-full w-full items-center justify-center bg-zinc-800">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" class="h-6 w-6 text-white/40">
              <path d="m10 8 6 4-6 4V8Z" />
            </svg>
          </div>
        </template>
        <img
          v-else-if="z.type === 'image' && zoneImage(z)"
          :src="zoneImage(z)"
          class="h-full w-full"
          :style="{ objectFit: z.fit === 'cover' ? 'cover' : 'contain' }"
          alt=""
        />
        <div v-else class="h-full w-full" style="background: #1a1a1e" />
      </div>

      <!-- perfil -->
      <template v-if="spec.profile.position !== 'hidden'">
        <div class="absolute flex items-center gap-2" :style="profileStyle">
          <div
            class="flex shrink-0 items-center justify-center overflow-hidden rounded-full bg-[#2B2B31] font-bold text-white"
            :style="avatarStyle"
          >
            <img
              v-if="spec.profile.avatar"
              :src="`/api/logos/${spec.profile.avatar}`"
              alt="avatar"
              class="h-full w-full object-cover"
            />
            <span v-else>{{ initials }}</span>
          </div>
          <div :class="spec.profile.align === 'left' ? 'text-left' : 'text-center'">
            <p class="font-bold leading-tight" :style="{ color: spec.text.color, fontSize: font(spec.profile.nameSize) }">
              {{ spec.profile.name || 'Seu perfil' }}
              <span v-if="spec.profile.verified" style="color: #3897f0">✓</span>
            </p>
            <p class="leading-tight" :style="{ color: spec.text.color, opacity: 0.72, fontSize: font(spec.profile.handleSize) }">
              {{ spec.profile.handle }}
            </p>
          </div>
        </div>
      </template>

      <!-- headline -->
      <p
        v-if="spec.text.position !== 'hidden'"
        class="absolute inset-x-[6%] font-bold leading-tight"
        :style="headlineStyle"
      >
        {{ headline || 'Sua headline' }}
      </p>
    </div>

    <!-- controles do preview -->
    <div v-if="showControls && hasVideo" class="mt-2 flex items-center gap-2">
      <button class="btn btn-ghost h-8 w-8 px-0" :aria-label="playing ? 'Pausar' : 'Tocar'" @click="toggle">
        <svg v-if="playing" viewBox="0 0 24 24" fill="currentColor" class="h-4 w-4"><path d="M8 5h3v14H8zM13 5h3v14h-3z" /></svg>
        <svg v-else viewBox="0 0 24 24" fill="currentColor" class="h-4 w-4"><path d="m8 5 11 7-11 7V5Z" /></svg>
      </button>
      <input
        :value="time"
        type="range"
        min="0"
        :max="duration || 0"
        step="0.1"
        class="flex-1 accent-pink-500"
        :disabled="!duration"
        @input="seek"
      />
      <span class="w-24 whitespace-nowrap text-right text-[11px] tabular-nums" style="color: var(--color-muted)">
        {{ fmtTime(time) }} / {{ fmtTime(duration) }}
      </span>
    </div>
  </div>
</template>

<script setup lang="ts">
const props = withDefaults(
  defineProps<{
    spec: any
    layout: any
    headline?: string
    mediaUrl?: string
    mediaById?: Record<string, string>
    showControls?: boolean
  }>(),
  { showControls: true },
)

const videos = ref<HTMLVideoElement[]>([])
const time = ref(0)
const duration = ref(0)
const playing = ref(false)
const hasVideo = computed(() => (props.layout?.zones || []).some((z: any) => z.type === 'video' && zoneMedia(z)))

function zoneMedia(z: any) {
  const specZone = props.spec.zones?.[z.index]
  const srcId = specZone?.sourceId
  if (srcId && props.mediaById?.[String(srcId)]) return props.mediaById[String(srcId)]
  if (srcId) return ''
  return props.mediaUrl || ''
}

function zoneImage(z: any) {
  const specZone = props.spec.zones?.[z.index]
  if (!specZone?.imagePath) return ''
  return `/api/images/${specZone.imagePath}`
}

function onMeta() {
  const primary = videos.value.find((v) => v && v.duration > 0)
  if (!primary) return
  duration.value = primary.duration || 0
  time.value = primary.currentTime || 0
  playing.value = !primary.paused
}

function onTime() {
  const primary = videos.value.find((v) => v && v.duration > 0)
  if (!primary) return
  time.value = primary.currentTime
  duration.value = primary.duration || 0
  playing.value = !primary.paused
}

function toggle() {
  const el = videos.value.find((v) => v && v.duration > 0)
  if (!el) return
  if (el.paused) {
    void el.play()
    playing.value = true
  } else {
    el.pause()
    playing.value = false
  }
}

function seek(ev: Event) {
  const value = Number((ev.target as HTMLInputElement).value)
  time.value = value
  for (const v of videos.value) {
    if (v && v.duration > 0) v.currentTime = value
  }
}

function fmtTime(s: number) {
  if (!s || !Number.isFinite(s)) return '0:00'
  const m = Math.floor(s / 60)
  const sec = Math.floor(s % 60)
  return `${m}:${String(sec).padStart(2, '0')}`
}

watch(
  () => props.mediaUrl,
  () => {
    const el = videos.value.find((v) => v && v.duration > 0)
    if (el) {
      duration.value = el.duration || 0
      playing.value = false
    }
  },
)

// 1 unit = 1080 px do canvas. Usa cqw para escalar com o container.
const font = (px: number) => `calc(${(px / 1080) * 100}cqw)`
const pctW = (px: number) => `${(px / 1080) * 100}%`
const pctH = (px: number) => `${(px / 1920) * 100}%`

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
    WebkitTextStroke: t.strokeEnabled && !t.band ? `${(t.strokeWidth / 1080) * 100}cqw ${t.strokeColor}` : '0px transparent',
    background: t.band ? `${t.bandColor}e6` : 'transparent',
    padding: t.band ? `${(t.size / 3 / 1080) * 100}cqw 0` : '0',
  }
})


</script>
