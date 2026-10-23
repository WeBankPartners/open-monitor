#!/bin/sh
# Run on the kubectl host. Only job names and counts are printed, never tokens.
set -eu
python3 - "$@" <<'PY'
import json
import re
import subprocess
import sys

if len(sys.argv) != 4 or sys.argv[3] not in ('before', 'after'):
    sys.exit('Usage: sh verify-legacy-k8s-scrape.sh <namespace> <monitor-pod> <before|after>')
namespace, pod, phase = sys.argv[1:]
base = ['kubectl', '-n', namespace, 'exec', pod, '--']

def read(*command):
    result = subprocess.run(base + list(command), capture_output=True, text=True)
    if result.returncode:
        sys.exit('ERROR: kubectl exec failed; verify namespace, pod, container and command availability')
    if not result.stdout.strip():
        sys.exit('ERROR: empty verification response')
    return result.stdout

def api(path):
    result = json.loads(read('wget', '-qO-', 'http://localhost:9090/api/v1/' + path))
    if result.get('status') != 'success':
        sys.exit('ERROR: Prometheus API did not return success')
    return result['data']

def jobs(config):
    return re.findall(r'''(?m)^\s*-\s*job_name:\s*['"]?([^'"\s#]+)''', config)

def legacy(job):
    return job.startswith(('k8s-kubelet-', 'k8s-cadvisor-'))

disk = jobs(read('cat', '/app/monitor/prometheus/prometheus.yml'))
loaded = jobs(api('status/config')['yaml'])
targets = api('targets?state=active')['activeTargets']
old_disk = sorted(job for job in disk if legacy(job))
old_loaded = sorted(job for job in loaded if legacy(job))
old_targets = [target for target in targets if legacy(target.get('labels', {}).get('job', ''))]
new_jobs = sorted(job for job in loaded if job in ('k8s_monitor', 'k8s_log_agent'))
new_targets = [target for target in targets
               if target.get('labels', {}).get('job') in ('k8s_monitor', 'k8s_log_agent')]
print('PVC_LEGACY_JOBS=' + str(len(old_disk)), ','.join(old_disk))
print('LOADED_LEGACY_JOBS=' + str(len(old_loaded)), ','.join(old_loaded))
print('ACTIVE_LEGACY_TARGETS=' + str(len(old_targets)))
print('HEALTHY_LEGACY_TARGETS=' + str(sum(target.get('health') == 'up' for target in old_targets)))
print('LOADED_NEW_JOBS=' + ','.join(new_jobs))
print('ACTIVE_NEW_TARGETS=' + str(len(new_targets)))
if phase == 'before':
    if not old_disk or not old_loaded:
        sys.exit('CHECK: legacy jobs were not found in both disk and loaded config')
    print('PASS: legacy scrape config exists before upgrade')
else:
    if old_disk or old_loaded or old_targets:
        sys.exit('FAIL: legacy scrape is still configured or active')
    print('PASS: legacy scrape config and active targets are absent after upgrade')
PY
