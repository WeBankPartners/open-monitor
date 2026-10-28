export function filterObjects(rows, filters, alertsActive = true) {
  return rows.filter(row => (!filters.workload || row.workloadName === filters.workload)
    && (!filters.pod || (row.podName || '').includes(filters.pod.trim()))
    && (!filters.instanceId || (row.instanceId || '').includes(filters.instanceId.trim()))
    && (!filters.onlyAbnormal || (row.statusKnown && (!row.ready || row.restartCount > 0))
      || (alertsActive && (row.alerts || []).some(alert => ['pending','firing'].includes(alert.Status)))))
}

export function metricRange(duration, interval, now = Math.floor(Date.now() / 1000)) {
  const step = Math.max(interval, Math.ceil(duration / (600 * interval)) * interval)
  return {
    start: now - duration,
    end: now,
    step
  }
}

export function chartSeries(response, range) {
  const rows = Array.isArray(response) ? response : ((response.data || response).result || [])
  const contextLabels = ['__name__', 'cluster_id', 'namespace', 'workload_type', 'workload_name', 'pod', 'pod_name', 'container', 'job', 'instance', 'instance_id', 'node', 'pod_uid', 'source_id', 'series_id']
  const series = rows.map((row, index) => {
    const labels = row.metric || {}
    const keys = Object.keys(labels).sort()
    const name = keys.length ? JSON.stringify(keys.map(key => [key, labels[key]])) : String(index + 1)
    const displayName = keys.filter(key => !contextLabels.includes(key))
      .map(key => key + '=' + labels[key])
      .join(', ') || '#' + (index + 1)
    const values = new Map((row.values || []).map(value => [Number(value[0]), value[1]]))
    const data = []
    for (let time = range.start; time <= range.end; time += range.step) {
      const raw = values.get(time)
      const value = raw === undefined || raw === null || raw === '' ? NaN : Number(raw)
      data.push([time * 1000, Number.isFinite(value) ? value : null])
    }
    return {
      name,
      displayName,
      labels: {...labels},
      type: 'line',
      showSymbol: false,
      connectNulls: false,
      data
    }
  }).filter(series => series.data.some(point => point[1] !== null))
  const counts = new Map()
  series.forEach(item => counts.set(item.displayName, (counts.get(item.displayName) || 0) + 1))
  series.forEach((item, index) => {
    if (counts.get(item.displayName) > 1) {item.displayName += ' · #' + (index + 1)}
  })
  return series
}

export function formatMetricValue(value, unit) {
  if (value === undefined || value === null || !Number.isFinite(Number(value))) {return '-'}
  const number = Number(value)
  const rounded = n => Number(n.toPrecision(4)).toString()
  if (unit === 'bytes' || unit === 'bytes/s') {
    const power = Math.min(4, Math.max(0, Math.floor(Math.log(Math.max(1, Math.abs(number))) / Math.log(1024))))
    return rounded(number / Math.pow(1024, power)) + ' ' + ['B', 'KiB', 'MiB', 'GiB', 'TiB'][power] + (unit === 'bytes/s' ? '/s' : '')
  }
  if (unit === 'ratio') {return rounded(number * 100) + '%'}
  if (unit === 'seconds' && number !== 0 && Math.abs(number) < 1) {return rounded(number * 1000) + ' ms'}
  return rounded(number) + (unit === 'seconds' ? ' s' : (unit === 'count' || unit === 'state' ? '' : ' ' + unit))
}
