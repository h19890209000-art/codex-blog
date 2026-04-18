function normalizeOrigin(value) {
  return String(value || '').trim().replace(/\/$/, '')
}

const apiOrigin = normalizeOrigin(import.meta.env.VITE_API_ORIGIN)

export function publicApiUrl(path) {
  return `${apiOrigin}/api/public${path}`
}

export async function publicApiRequest(path, options = {}) {
  const response = await fetch(publicApiUrl(path), options)
  const rawText = await response.text()

  if (!rawText.trim()) {
    throw new Error('服务器响应为空，通常是首次抓取原文或生成译文超时，请稍后重试一次')
  }

  let result
  try {
    result = JSON.parse(rawText)
  } catch (error) {
    throw new Error(`接口返回了非 JSON 内容：${rawText.slice(0, 120)}`)
  }

  if (!response.ok || result.success === false) {
    throw new Error(result.error || '请求失败')
  }

  return result.data
}
