function normalizeOrigin(value) {
  return String(value || '').trim().replace(/\/$/, '')
}

const apiOrigin = normalizeOrigin(import.meta.env.VITE_API_ORIGIN)

export function publicApiUrl(path) {
  return `${apiOrigin}/api/public${path}`
}
