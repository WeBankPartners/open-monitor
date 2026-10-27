#!/bin/sh
# Read-only checks from a kubectl host; never print tokens or metric labels.
set -eu
python3 - "$@" <<'PY'
import json
import re
import subprocess
import sys
import time
from urllib.parse import urlsplit

if len(sys.argv) not in (7, 8) or sys.argv[6] not in ('before', 'off', 'on'):
    sys.exit('Usage: sh verify-k8s-local-collection.sh <monitor-namespace> <monitor-pod> <agent-namespace> <agent-pod> <prometheus-base-url> <before|off|on> [observe-seconds]')
monitor_ns, monitor_pod, agent_ns, agent_pod, prometheus, phase = sys.argv[1:7]
try:
    observe = int(sys.argv[7]) if len(sys.argv) == 8 else 35
    if observe < 1 or observe > 3600:
        raise ValueError()
    parts = urlsplit(prometheus)
    if parts.scheme not in ('http', 'https') or not parts.hostname or parts.username or parts.password or parts.query or parts.fragment:
        raise ValueError()
except ValueError:
    sys.exit('ERROR: provide an HTTP(S) base URL without credentials and observe-seconds from 1 to 3600')
prometheus = prometheus.rstrip('/')
jobs = ('k8s_monitor', 'k8s_log_agent')

def read(namespace, pod, *args):
    try:
        result = subprocess.run(['kubectl', '-n', namespace, 'exec', pod, '--', *args],
                                capture_output=True, text=True, timeout=25)
    except (OSError, subprocess.TimeoutExpired):
        sys.exit('ERROR: kubectl exec unavailable or timed out')
    if result.returncode or not result.stdout.strip():
        sys.exit('ERROR: kubectl exec failed or returned an empty response; check Pod, container, port and command availability')
    return result.stdout

def monitor_read(*args):
    return read(monitor_ns, monitor_pod, *args)

def api(path):
    result = json.loads(monitor_read('wget', '-q', '-O', '-', '-T', '10', prometheus + '/api/v1/' + path))
    if result.get('status') != 'success':
        sys.exit('ERROR: Prometheus API did not return success')
    return result['data']

def inspect_targets():
    for job, filename in zip(jobs, ('k8s_monitor.json', 'k8s_log_agents.json')):
        groups = json.loads(monitor_read('cat', '/app/monitor/prometheus/sd_file/' + filename))
        if not isinstance(groups, list):
            sys.exit('FAIL: target file is not a JSON array')
        count = sum(len(group.get('targets', [])) for group in groups)
        print('FILE_TARGETS ' + job + '=' + str(count), flush=True)
        if phase == 'off' and groups != []:
            sys.exit('FAIL: disabled mode requires an empty target array')
        if phase == 'on' and count == 0:
            sys.exit('FAIL: enabled mode requires populated target files')
    targets = api('targets?state=active')['activeTargets']
    loaded = re.findall(r'''(?m)^\s*-\s*job_name:\s*['"]?([^'"\s#]+)''', api('status/config')['yaml'])
    for job in jobs:
        matching = [target for target in targets if target.get('labels', {}).get('job') == job]
        print('ACTIVE_TARGETS ' + job + '=' + str(len(matching)), flush=True)
        print('HEALTHY_TARGETS ' + job + '=' + str(sum(target.get('health') == 'up' for target in matching)), flush=True)
        if phase != 'before' and job not in loaded:
            sys.exit('FAIL: expected Kubernetes job definition is missing')
        if phase == 'off' and matching:
            sys.exit('FAIL: disabled mode still has active Kubernetes targets')
        if phase == 'on' and (not matching or any(target.get('health') != 'up' for target in matching)):
            sys.exit('FAIL: enabled mode requires healthy targets for both Kubernetes jobs')

try:
    inspect_targets()
    if phase == 'off':
        print('Observing Agent heartbeat target synchronization for ' + str(observe) + ' seconds', flush=True)
        time.sleep(observe)
        inspect_targets()
    # Read only the non-secret listening address, supporting custom Agent ports.
    address = read(agent_ns, agent_pod, 'sh', '-c', 'printf "%s" "${LOG_AGENT_HTTP_ADDR:-:9105}"').strip()
    port = address.rsplit(':', 1)[-1]
    if not port.isdigit() or not 1 <= int(port) <= 65535:
        sys.exit('ERROR: cannot determine Agent metrics port')
    metrics = read(agent_ns, agent_pod, 'wget', '-q', '-O', '-', '-T', '10', 'http://localhost:' + port + '/metrics')
    samples = [line for line in metrics.splitlines() if line and not line.startswith('#')]
    print('AGENT_EXPORTER_SAMPLES=' + str(len(samples)))
    print('AGENT_LOG_MATCH_SERIES=' + str(sum(line.startswith('k8s_log_match_total{') for line in samples)))
except (ValueError, TypeError, KeyError):
    sys.exit('ERROR: unexpected JSON or metrics response')
print('PASS: ' + phase + ' target checks and Agent exporter request completed')
print('CHECK SEPARATELY: outbox cancellation, unchanged historical alerts, cAdvisor request counts, all Agent Pods and actual external Prometheus scraping')
PY
