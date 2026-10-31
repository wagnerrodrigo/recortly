export interface TemplateZone {
  type: 'video' | 'blank' | 'image'
  weight: number
  radius: number
  fit: string
}

export interface TemplateSpec {
  schemaVersion: number
  name: string
  preset: string
  canvas: { width: number; height: number; fps: number; background: string }
  profile: {
    name: string
    handle: string
    verified: boolean
    avatar: string
    position: string
    align: string
    theme: string
    avatarSize: number
    nameSize: number
    handleSize: number
  }
  text: {
    font: string
    size: number
    color: string
    align: string
    sideSpacing: number
    position: string
    band: boolean
    bandColor: string
    strokeEnabled: boolean
    strokeWidth: number
    strokeColor: string
    offset: number
  }
  logo: {
    enabled: boolean
    path: string
    position: string
    size: number
    opacity: number
    margin: number
  }
  zones: TemplateZone[]
  phrases: string[]
  variations: {
    count: number
    zoom: number[]
    saturation: number[]
    brightness: number[]
    speed: number[]
    offsetX: number[]
    offsetY: number[]
    mirror: boolean
    seedBase: number
  }
}

export function defaultSpec(): TemplateSpec {
  return {
    schemaVersion: 2,
    name: 'Novo template',
    preset: 'blank',
    canvas: { width: 1080, height: 1920, fps: 30, background: '#0E0E10' },
    profile: {
      name: 'Seu perfil',
      handle: '@seuperfil',
      verified: true,
      avatar: '',
      position: 'top',
      align: 'center',
      theme: 'light',
      avatarSize: 170,
      nameSize: 50,
      handleSize: 42,
    },
    text: {
      font: 'Inter Bold',
      size: 53,
      color: '#FFFFFF',
      align: 'center',
      sideSpacing: 84,
      position: 'top',
      band: false,
      bandColor: '#000000',
      strokeEnabled: false,
      strokeWidth: 4,
      strokeColor: '#000000',
      offset: 24,
    },
    logo: { enabled: false, path: '', position: 'top-right', size: 140, opacity: 0.9, margin: 40 },
    zones: [{ type: 'video', weight: 1, radius: 16, fit: 'cover' }],
    phrases: ['Coloque sua headline aqui'],
    variations: {
      count: 5,
      zoom: [1.0, 1.18],
      saturation: [0.9, 1.15],
      brightness: [0.96, 1.06],
      speed: [0.96, 1.04],
      offsetX: [0.35, 0.65],
      offsetY: [0.4, 0.6],
      mirror: false,
      seedBase: 42,
    },
  }
}

export const zonePresets = [
  { id: 'single', label: '1 vídeo', icon: '▮', zones: [{ type: 'video', weight: 1, radius: 16, fit: 'cover' }] },
  {
    id: 'video-blank',
    label: 'Vídeo + espaço',
    icon: '▮▯',
    zones: [
      { type: 'video', weight: 0.7, radius: 16, fit: 'cover' },
      { type: 'blank', weight: 0.3, radius: 0, fit: 'cover' },
    ],
  },
  {
    id: 'stacked',
    label: '2 vídeos empilhados',
    icon: '▮▮',
    zones: [
      { type: 'video', weight: 0.55, radius: 14, fit: 'cover' },
      { type: 'video', weight: 0.45, radius: 14, fit: 'cover' },
    ],
  },
  {
    id: 'blank-video',
    label: 'Espaço + vídeo',
    icon: '▯▮',
    zones: [
      { type: 'blank', weight: 0.28, radius: 0, fit: 'cover' },
      { type: 'video', weight: 0.72, radius: 16, fit: 'cover' },
    ],
  },
] as const
