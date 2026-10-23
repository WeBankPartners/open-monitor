#!/bin/sh
# Repair only complete, identical entries in Monitor's managed remote-write region.
set -eu

config=$1
promtool=$2
candidate=$(mktemp "${config}.remote-write.XXXXXX")
stats=$(mktemp "${config}.remote-write-count.XXXXXX")
backup_candidate=''
trap 'rm -f "$candidate" "$stats"; if [ -n "$backup_candidate" ]; then rm -f "$backup_candidate"; fi' EXIT HUP INT TERM

awk -v stats="$stats" '
  function flush() {
    if (entry == "") return
    if (entry in seen) removed++
    else { seen[entry]=1; printf "%s", entry }
    entry=""
  }
  /^remote_write:[[:space:]]*/ { remote=1; print; next }
  /^[[:alnum:]_]+:/ { remote=0 }
  /^[[:space:]]*#Remote_write_start[[:space:]]*$/ {
    if (!remote || started || ended) { bad=1; exit 1 }
    started=1; inside=1; print; next
  }
  /^[[:space:]]*#Remote_write_end[[:space:]]*$/ {
    if (!inside || ended) { bad=1; exit 1 }
    flush(); inside=0; ended=1; print; next
  }
  inside {
    if ($0 ~ /^[[:space:]]*(#.*)?$/) {
      if (entry == "") print
      else entry=entry $0 "\n"
      next
    }
    if ($0 ~ /^[[:space:]]*-[[:space:]]/) {
      indent=match($0, /[^[:space:]]/)-1
      if (!have_indent) { base=indent; have_indent=1 }
      if (indent < base) { bad=1; exit 1 }
      if (indent == base) flush()
    } else if (entry == "" || $0 ~ /^[^[:space:]]/) {
      bad=1; exit 1
    }
    entry=entry $0 "\n"
    next
  }
  { print }
  END {
    if (bad || inside || started != ended) exit 1
    print removed+0 > stats
  }
' "$config" > "$candidate" || {
  echo "Remote write duplicate repair refused ambiguous managed config" >&2
  exit 1
}

removed=$(cat "$stats")
if [ "$removed" = 0 ]; then
  echo "Managed remote write configs already unique"
  exit 0
fi
if [ ! -x "$promtool" ] || ! "$promtool" check config "$candidate" >/dev/null 2>&1; then
  echo "Remote write duplicate repair failed config validation" >&2
  exit 1
fi

# Preserve the original privately; concurrent or repeated repairs keep the first backup.
backup="${config}.remote-write-duplicates.bak"
if [ ! -f "$backup" ]; then
  backup_candidate=$(mktemp "${config}.remote-write-backup.XXXXXX")
  cat "$config" > "$backup_candidate"
  if ! ln "$backup_candidate" "$backup" && [ ! -f "$backup" ]; then
    echo "Remote write duplicate repair failed to preserve original config" >&2
    exit 1
  fi
fi
chmod 600 "$backup"
mv -f "$candidate" "$config"
echo "Duplicate managed remote write configs repaired: removed=$removed"
