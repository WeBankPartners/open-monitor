"""Check the deployment contract and startup behavior of local collection."""
from pathlib import Path
import json
import os
import shutil
import subprocess
import tempfile
import unittest
import xml.etree.ElementTree as ET

BUILD = Path(__file__).resolve().parents[1]


class LocalCollectionDeploymentTest(unittest.TestCase):
    def test_plugin_defaults_off_and_forwards_parameter(self):
        root = ET.parse(BUILD / 'register.xml').getroot()
        parameter = next(e for e in root.iter('systemParameter')
                         if e.get('name') == 'MONITOR_K8S_LOCAL_COLLECTION_ENABLED')
        self.assertEqual(parameter.get('defaultValue'), 'N')
        self.assertTrue(any('MONITOR_K8S_LOCAL_COLLECTION_ENABLED={{MONITOR_K8S_LOCAL_COLLECTION_ENABLED}}'
                            in e.get('envVariables', '') for e in root.iter('docker')))


class LocalCollectionStartupTest(unittest.TestCase):
    def prepare(self, directory):
        target = directory / 'prometheus'
        sd = target / 'sd_file'
        sd.mkdir(parents=True)
        (target / 'logs').mkdir()
        (target / 'custom_scrape').mkdir()
        (target / 'custom_scrape' / 'invalid.yml').write_text('invalid custom config')
        (target / 'data').mkdir()
        (target / 'data' / 'retained-block').write_text('retained metrics')
        for name in ('k8s_monitor.json', 'k8s_log_agents.json', 'sd_file_10.json'):
            (sd / name).write_text('[{"targets":["exporter.example.invalid:9105"]}]')
        (target / 'prometheus.yml').write_text('''scrape_configs:
#Kubernetes_start
  - job_name: k8s-cadvisor-cluster-a
#Kubernetes_end
  - job_name: unrelated-job
remote_write:
#Remote_write_start
  - url: 'https://receiver.example.invalid/write'
  - url: 'https://receiver.example.invalid/write'
#Remote_write_end
''')
        (directory / 'prometheus_scrape_defaults.yml').write_text(
            (BUILD / 'conf' / 'prometheus.yml').read_text())
        promtool = target / 'promtool'
        promtool.write_text('#!/bin/sh\n[ ! -f "' + str(target / 'custom_scrape' / 'invalid.yml') + '" ]\n')
        promtool.chmod(0o755)
        prometheus = target / 'prometheus'
        prometheus.write_text('''#!/usr/bin/env python3
from pathlib import Path
import json
for name in ('k8s_monitor.json', 'k8s_log_agents.json'):
    assert json.loads((Path('sd_file') / name).read_text()) == []
config = Path('prometheus.yml').read_text()
assert config.count('https://receiver.example.invalid/write') == 1
assert 'k8s-cadvisor-' not in config
assert 'k8s_monitor' in config and 'k8s_log_agent' in config
Path('launched').write_text('ready')
''')
        prometheus.chmod(0o755)
        start = (BUILD / 'start.sh').read_text()
        block = start[start.index('# 启动前校验'):start.index('\ncd ../ping_exporter/')]
        for helper in ('deduplicate-remote-write.sh', 'disable-legacy-k8s-scrape.sh'):
            block = block.replace('/app/monitor/' + helper, str(BUILD / helper))
        block = block.replace('/app/monitor/prometheus', str(target))
        block = block.replace('/tmp/promtool_check.log', str(directory / 'check.log'))
        return target, block

    def run_startup(self, directory, block, master=None, local=None, failure='', path_prefix=None):
        env = dict(os.environ)
        if path_prefix is not None:
            env['PATH'] = str(path_prefix) + os.pathsep + env['PATH']
        for key, value in [('MONITOR_K8S_ENABLED', master), ('MONITOR_K8S_LOCAL_COLLECTION_ENABLED', local)]:
            env.pop(key, None)
            if value is not None:
                env[key] = value
        prefix = '''ensure_log_file() { touch "$1"; }
sudo() { if [ "$1" = chown ]; then return 0; fi; command "$@"; }
archive_day=30d
'''
        return subprocess.run(['bash', '-c', prefix + failure + block + '\nwait\n'],
                              env=env, capture_output=True, text=True, timeout=15)

    def test_startup_clears_persistent_targets_independent_of_environment(self):
        for master, local in [(None, None), ('Y', 'N'), ('Y', 'Y'), ('N', 'Y'), ('invalid', 'true')]:
            with self.subTest(master=master, local=local), tempfile.TemporaryDirectory() as tmp:
                target, block = self.prepare(Path(tmp))
                for _ in range(2):
                    result = self.run_startup(Path(tmp), block, master, local)
                    self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
                    self.assertTrue((target / 'launched').exists(), (target / 'logs' / 'prometheus.log').read_text())
                    for name in ('k8s_monitor.json', 'k8s_log_agents.json'):
                        self.assertEqual(json.loads((target / 'sd_file' / name).read_text()), [])
                self.assertIn('exporter.example.invalid', (target / 'sd_file' / 'sd_file_10.json').read_text())
                self.assertEqual((target / 'data' / 'retained-block').read_text(), 'retained metrics')
                self.assertTrue(list(target.glob('custom_scrape_bad_*/invalid.yml')))
                self.assertIn('unrelated-job', (target / 'prometheus.yml').read_text())

    def test_target_file_failures_prevent_prometheus_start_and_can_retry(self):
        for name in ('k8s_monitor.json', 'k8s_log_agents.json'):
            for operation in ('create', 'write', 'replace'):
                with self.subTest(name=name, operation=operation), tempfile.TemporaryDirectory() as tmp:
                    target, block = self.prepare(Path(tmp))
                    if operation == 'create':
                        failure = 'mktemp() { case "$*" in *' + name + '*) return 1;; esac; command mktemp "$@"; }\n'
                    elif operation == 'write':
                        attempt = 1 if name == 'k8s_monitor.json' else 2
                        failure = ('target_write_attempt=0\nprintf() { target_write_attempt=$((target_write_attempt + 1)); '
                                   '[ "$target_write_attempt" -ne ' + str(attempt) + ' ] || return 1; command printf "$@"; }\n')
                    else:
                        failure = 'mv() { case "${@: -1}" in *' + name + ') return 1;; esac; command mv "$@"; }\n'
                    result = self.run_startup(Path(tmp), block, failure=failure)
                    self.assertNotEqual(result.returncode, 0, result.stdout + result.stderr)
                    self.assertFalse((target / 'launched').exists())
                    self.assertIn(name, result.stderr)
                    self.assertEqual(list((target / 'sd_file').glob('.*')), [])
                    result = self.run_startup(Path(tmp), block)
                    self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
                    self.assertTrue((target / 'launched').exists())

    def test_startup_target_replacement_supports_busybox_mv(self):
        busybox = shutil.which('busybox')
        if busybox is None:
            self.skipTest('BusyBox is not installed')
        with tempfile.TemporaryDirectory() as tmp:
            directory = Path(tmp)
            target, block = self.prepare(directory)
            applets = directory / 'bin'
            applets.mkdir()
            (applets / 'mv').symlink_to(busybox)
            result = self.run_startup(directory, block, path_prefix=applets)
            self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
            self.assertTrue((target / 'launched').exists())

    def test_directory_at_target_path_prevents_start(self):
        with tempfile.TemporaryDirectory() as tmp:
            directory = Path(tmp)
            target, block = self.prepare(directory)
            blocked = target / 'sd_file' / 'k8s_monitor.json'
            blocked.unlink()
            blocked.mkdir()
            result = self.run_startup(directory, block)
            self.assertNotEqual(result.returncode, 0, result.stdout + result.stderr)
            self.assertFalse((target / 'launched').exists())
            self.assertIn('k8s_monitor.json', result.stderr)
            self.assertEqual(list(blocked.iterdir()), [])


class LocalCollectionVerificationTest(unittest.TestCase):
    def run_check(self, case, phase):
        with tempfile.TemporaryDirectory() as directory:
            fake = Path(directory) / 'kubectl'
            fake.write_text('''#!/usr/bin/env python3
import json, os, sys
args = sys.argv[sys.argv.index('--') + 1:]
case = os.environ['VERIFY_CASE']
if args[0] == 'cat':
    groups = [] if case != 'on' else [{'targets': ['exporter.example.invalid:9105']}]
    print(json.dumps(groups))
elif args[0] == 'sh':
    print(':9105')
elif args[0] == 'wget':
    url = args[-1]
    if url.endswith('/metrics'):
        print('k8s_log_match_total{private_label="DO_NOT_PRINT"} 1')
    elif 'status/config' in url:
        print(json.dumps({'status': 'success', 'data': {'yaml': "scrape_configs:\\n  - job_name: 'k8s_monitor'\\n  - job_name: 'k8s_log_agent'\\n"}}))
    elif 'targets?' in url:
        targets = [] if case == 'off' else [{'labels': {'job': job}, 'health': 'up'} for job in ('k8s_monitor', 'k8s_log_agent')]
        print(json.dumps({'status': 'success', 'data': {'activeTargets': targets}}))
    else:
        sys.exit(3)
else:
    sys.exit(4)
''')
            fake.chmod(0o755)
            env = dict(os.environ, PATH=directory + os.pathsep + os.environ['PATH'], VERIFY_CASE=case)
            return subprocess.run(['sh', str(BUILD / 'verify-k8s-local-collection.sh'),
                                   'monitor-example', 'monitor-pod', 'agent-example', 'agent-pod',
                                   'http://prometheus.example.invalid:9090', phase, '1'],
                                  env=env, capture_output=True, text=True, timeout=15)

    def test_disabled_checks_files_targets_and_exporter_without_printing_labels(self):
        result = self.run_check('off', 'off')
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(result.stdout.count('ACTIVE_TARGETS k8s_log_agent=0'), 2)
        self.assertIn('AGENT_LOG_MATCH_SERIES=1', result.stdout)
        self.assertNotIn('DO_NOT_PRINT', result.stdout + result.stderr)

    def test_active_target_fails_disabled_verification(self):
        result = self.run_check('stale', 'off')
        self.assertNotEqual(result.returncode, 0)
        self.assertIn('still has active', result.stderr)

    def test_enabled_rejects_empty_files_with_cached_healthy_targets(self):
        result = self.run_check('stale', 'on')
        self.assertNotEqual(result.returncode, 0)
        self.assertIn('requires populated target files', result.stderr)

    def test_enabled_and_before_checks(self):
        for phase in ('on', 'before'):
            result = self.run_check('on', phase)
            self.assertEqual(result.returncode, 0, result.stderr)


if __name__ == '__main__':
    unittest.main()
