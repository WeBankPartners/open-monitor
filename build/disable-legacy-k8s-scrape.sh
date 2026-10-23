#!/bin/sh
# Operate on the mounted config; this script itself must live outside the PVC.
set -eu

config=$1
promtool=$2
defaults=$3
candidate=$(mktemp "${config}.legacy-k8s.XXXXXX")
backup_candidate=''
trap 'rm -f "$candidate"; if [ -n "$backup_candidate" ]; then rm -f "$backup_candidate"; fi' EXIT HUP INT TERM

# The marked region is owned by the retired Kubernetes config generator.
# Refuse ambiguous configs rather than removing unrelated jobs or root sections.
awk '
  function jobname(line) {
    sub(/.*job_name:[[:space:]]*/, "", line)
    sub(/[[:space:]]*#.*/, "", line)
    gsub(/["\047[:space:]]/, "", line)
    return line
  }
  BEGIN {
    seconds=ENVIRON["MONITOR_K8S_METRICS_INTERVAL_SECONDS"]
    gsub(/^[[:space:]]+|[[:space:]]+$/, "", seconds)
    sub(/^\+/, "", seconds)
    if (seconds !~ /^[0-9]+$/ || seconds+0 < 15 || seconds+0 > 300) seconds=30
    interval=(seconds+0) "s"
  }
  FILENAME == ARGV[1] {
    if (/^[[:space:]]*-[[:space:]]*job_name:/) current=jobname($0)
    if (/^[^[:space:]#]/) current=""
    if (current == "k8s_monitor" || current == "k8s_log_agent") {
      if (/^[[:space:]]+scrape_interval:/) sub(/scrape_interval:.*/, "scrape_interval: " interval)
      defaults[current]=defaults[current] $0 "\n"
    }
    next
  }
  /^[[:space:]]*#Kubernetes_start[[:space:]]*$/ {
    if (started || ended) { bad=1; exit 1 }
    started=1; inside=1; lines[++n]=$0; next
  }
  /^[[:space:]]*#Kubernetes_end[[:space:]]*$/ {
    if (!inside || ended) { bad=1; exit 1 }
    ended=1; inside=0; lines[++n]=$0; next
  }
  /^[[:space:]]*-[[:space:]]*job_name:/ {
    job=jobname($0)
    legacy=(job ~ /^k8s-(kubelet|cadvisor)-/)
    if ((inside && !legacy) || (!inside && legacy)) { bad=1; exit 1 }
    present[job]=1
  }
  inside && /^[^[:space:]#]/ { bad=1; exit 1 }
  !inside {
    lines[++n]=$0
    if (/^scrape_configs:[[:space:]]*(\[\][[:space:]]*)?(#.*)?$/) header=n
  }
  END {
    if (bad || inside || started != ended) exit 1
    need_monitor=!present["k8s_monitor"]
    need_agent=!present["k8s_log_agent"]
    if ((need_monitor || need_agent) && (!header || !defaults["k8s_monitor"] || !defaults["k8s_log_agent"])) exit 1
    for (i=1; i<=n; i++) {
      if (i == header && (need_monitor || need_agent)) {
        print "scrape_configs:"
        if (need_monitor) printf "%s", defaults["k8s_monitor"]
        if (need_agent) printf "%s", defaults["k8s_log_agent"]
      } else print lines[i]
    }
  }
' "$defaults" "$config" > "$candidate" || {
  echo "Legacy Kubernetes scrape cleanup refused ambiguous config or missing defaults" >&2
  exit 1
}

if cmp -s "$config" "$candidate"; then
  echo "Legacy Kubernetes scrape config already disabled"
  exit 0
fi

if [ ! -x "$promtool" ] || ! "$promtool" check config "$candidate" >/dev/null 2>&1; then
  echo "Legacy Kubernetes scrape cleanup failed config validation" >&2
  exit 1
fi

# Keep credentials in the backup private. Repeated cleanup preserves this backup.
backup="${config}.legacy-k8s.bak"
if [ ! -f "$backup" ]; then
  backup_candidate=$(mktemp "${config}.legacy-k8s-backup.XXXXXX")
  cat "$config" > "$backup_candidate"
  mv -f "$backup_candidate" "$backup"
fi
chmod 600 "$backup"
mv -f "$candidate" "$config"
echo "Legacy Kubernetes scrape config disabled"
