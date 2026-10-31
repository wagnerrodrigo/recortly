export interface ApiError {
  error: string
}

export async function apiGet<T>(path: string): Promise<T> {
  return await $fetch<T>(path)
}

export async function apiSend<T>(path: string, method: 'POST' | 'PUT' | 'PATCH' | 'DELETE', body?: unknown): Promise<T> {
  return await $fetch<T>(path, { method, body })
}

export function errorMessage(err: unknown): string {
  const e = err as { data?: ApiError; message?: string }
  return e?.data?.error || e?.message || 'Erro inesperado'
}

export function fmtDuration(ms: number): string {
  if (!ms || ms < 0) return '00:00'
  const total = Math.floor(ms / 1000)
  const m = Math.floor(total / 60)
  const s = total % 60
  return `${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
}

export function fmtBytes(bytes: number): string {
  if (!bytes) return '0 MB'
  const mb = bytes / (1024 * 1024)
  if (mb >= 1024) return `${(mb / 1024).toFixed(2)} GB`
  return `${mb.toFixed(1)} MB`
}

export function fmtDate(iso?: string | null): string {
  if (!iso) return '—'
  return new Date(iso).toLocaleString('pt-BR', { dateStyle: 'short', timeStyle: 'short' })
}

export const platformLabels: Record<string, string> = {
  youtube: 'YouTube',
  instagram: 'Instagram',
  tiktok: 'TikTok',
}

export function statusChip(status: string): string {
  switch (status) {
    case 'done':
    case 'published':
      return 'chip chip-ok'
    case 'running':
    case 'publishing':
    case 'scheduled':
      return 'chip chip-warn'
    case 'failed':
    case 'cancelled':
      return 'chip chip-err'
    default:
      return 'chip chip-neutral'
  }
}

// Nome do arquivo baixado, igual ao usado no ZIP do backend.
export function videoFileName(headline: string | undefined, index: number, total: number): string {
  const valid = (headline || '')
    .normalize('NFD')
    .replace(/[\u0300-\u036f]/g, '')
    .replace(/[\\/:*?"<>|\x00-\x1f]+/g, ' ')
    .replace(/[^\x20-\x7E]/g, '')
    .replace(/\s+/g, ' ')
    .trim()
    .replace(/^[ .-]+|[ .-]+$/g, '')
    .slice(0, 70)
    .trim()
  const base = valid || 'video'
  return total > 1 ? `${String(index).padStart(2, '0')} - ${base}.mp4` : `${base}.mp4`
}
