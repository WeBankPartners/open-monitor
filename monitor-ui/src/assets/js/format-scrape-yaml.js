const TOP_LEVEL_KEYS = {
  job_name: true,
  scrape_interval: true,
  scrape_timeout: true,
  scrape_protocols: true,
  fallback_scrape_protocol: true,
  metrics_path: true,
  honor_labels: true,
  honor_timestamps: true,
  track_timestamps_staleness: true,
  scheme: true,
  params: true,
  basic_auth: true,
  authorization: true,
  oauth2: true,
  tls_config: true,
  proxy_url: true,
  proxy_from_environment: true,
  follow_redirects: true,
  enable_http2: true,
  http_headers: true,
  static_configs: true,
  file_sd_configs: true,
  http_sd_configs: true,
  dns_sd_configs: true,
  consul_sd_configs: true,
  kubernetes_sd_configs: true,
  ec2_sd_configs: true,
  azure_sd_configs: true,
  gce_sd_configs: true,
  docker_sd_configs: true,
  relabel_configs: true,
  metric_relabel_configs: true,
  body_size_limit: true,
  sample_limit: true,
  label_limit: true,
  target_limit: true,
  bearer_token: true,
  bearer_token_file: true,
  keep_dropped_targets: true
}

function isTopLevelKey(key) {
  return !!TOP_LEVEL_KEYS[key]
}

export function formatScrapeYaml(input) {
  if (input === null || input === undefined) {
    return ''
  }
  const text = String(input)
    .replace(/\t/g, '  ')
    .replace(/\r\n/g, '\n')
    .replace(/\r/g, '\n')
  if (text.trim() === '') {
    return text
  }
  const lines = text.split('\n')
  const hasJobDash = lines.some(line => /^(\s*)-\s*job_name\s*:/.test(line))
  if (!hasJobDash) {
    for (let i = 0; i < lines.length; i++) {
      if (/^\s*job_name\s*:/.test(lines[i]) && !lines[i].trim().startsWith('#')) {
        lines[i] = '- ' + lines[i].trim()
        break
      }
    }
  }
  const result = []
  let lastTopLevelOrigIndent = 2
  const jobKeyIndent = 2
  for (let i = 0; i < lines.length; i++) {
    const line = lines[i]
    if (line.trim() === '') {
      result.push('')
      continue
    }
    const origIndent = line.match(/^ */)[0].length
    const trimmed = line.trimStart()
    if (trimmed.startsWith('#')) {
      result.push(line)
      continue
    }
    if (/^-\s*job_name\s*:/.test(trimmed)) {
      lastTopLevelOrigIndent = origIndent + 2
      result.push('- ' + trimmed.replace(/^-\s*/, ''))
      continue
    }
    const keyMatch = trimmed.match(/^([A-Za-z0-9_]+)\s*:/)
    const key = keyMatch ? keyMatch[1] : ''
    const isListItem = trimmed.startsWith('- ')
    if (key && !isListItem && (isTopLevelKey(key) || origIndent <= jobKeyIndent + 1)) {
      lastTopLevelOrigIndent = origIndent
      result.push('  ' + trimmed)
      continue
    }
    let rel = origIndent - lastTopLevelOrigIndent
    if (rel < 2) {
      rel = 2
    }
    rel = Math.round(rel / 2) * 2
    result.push(' '.repeat(2 + rel) + trimmed)
  }
  const formatted = result
    .join('\n')
    .replace(/\s+$/g, '')
  return formatted + (text.endsWith('\n') ? '\n' : '')
}

export default {
  formatScrapeYaml
}
