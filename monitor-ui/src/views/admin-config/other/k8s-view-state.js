const storageKey = 'monitor.k8s.view'

export function readViewState(page) {
  try {
    const state = JSON.parse(sessionStorage.getItem(storageKey) || '{}')
    const value = state && state[page]
    return value && typeof value === 'object' && !Array.isArray(value) ? value : {}
  } catch (e) { return {} }
}

export function writeViewState(page, value) {
  let state = {}
  try {
    const parsed = JSON.parse(sessionStorage.getItem(storageKey) || '{}')
    if (parsed && typeof parsed === 'object' && !Array.isArray(parsed)) state = parsed
  } catch (e) { /* Ignore invalid or unavailable session data. */ }
  try {
    state[page] = value
    sessionStorage.setItem(storageKey, JSON.stringify(state))
  } catch (e) { /* Browsing remains available when session storage is disabled. */ }
}

export function resolveScope(clusters, namespaces, preferred = {}) {
  const clusterId = preferred.clusterId || ((clusters.find(row => row.enabled) || {}).clusterId || '')
  if (!clusters.some(row => row.clusterId === clusterId && row.enabled)) {
    return {clusterId:'',namespace:'',unavailable:!!preferred.clusterId}
  }
  const available = namespaces.filter(row => row.clusterId === clusterId && row.enabled)
  const namespace = preferred.namespace || ((available[0] || {}).namespace || '')
  if (namespace && !available.some(row => row.namespace === namespace)) {
    return {clusterId,namespace:'',unavailable:true}
  }
  return {clusterId,namespace,unavailable:false}
}
