"""Exercise upgrade migration against a simulated persistent Prometheus directory."""
from pathlib import Path
import json
import os
import subprocess
import tempfile
import unittest

ROOT = Path(__file__).resolve().parents[2]
MIGRATE = ROOT / 'build/disable-legacy-k8s-scrape.sh'
DEFAULTS = """scrape_configs:
  - job_name: k8s_monitor
    scrape_interval: 30s
  - job_name: k8s_log_agent
    scrape_interval: 30s
"""
LEGACY = """scrape_configs:
  - job_name: k8s_monitor
    scrape_interval: 30s
#Kubernetes_start
  - job_name: 'k8s-kubelet-cluster-a'
    bearer_token_file: /example/token/cluster-a
  - job_name: 'k8s-cadvisor-cluster-a'
    metrics_path: /metrics/cadvisor
#Kubernetes_end
#Snmp_start
  - job_name: snmp-example
    scrape_interval: 60s
#Snmp_end
remote_write: []
"""
CLEAN = """scrape_configs:
  - job_name: k8s_log_agent
    scrape_interval: 30s
  - job_name: k8s_monitor
    scrape_interval: 30s
#Kubernetes_start
#Kubernetes_end
#Snmp_start
  - job_name: snmp-example
    scrape_interval: 60s
#Snmp_end
remote_write: []
"""


class LegacyKubernetesScrapeTest(unittest.TestCase):
    def migrate(self, config, validator):
        defaults = config.parent / 'prometheus_scrape_defaults.yml'
        defaults.write_text(DEFAULTS)
        return subprocess.run(['sh', str(MIGRATE), str(config), str(validator), str(defaults)],
                              capture_output=True, text=True)

    def validator(self, directory, exit_code=0):
        path = directory / 'promtool'
        path.write_text('#!/bin/sh\nexit ' + str(exit_code) + '\n')
        path.chmod(0o755)
        return path

    def test_persistent_config_cleaned_and_repeat_preserves_original_backup(self):
        with tempfile.TemporaryDirectory() as tmp:
            directory = Path(tmp)
            config = directory / 'prometheus.yml'
            config.write_text(LEGACY)
            data = directory / 'data'
            data.mkdir()
            (data / 'retained-block').write_text('retained metrics')
            validator = self.validator(directory)
            result = self.migrate(config, validator)
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertEqual(config.read_text(), CLEAN)
            backup = directory / 'prometheus.yml.legacy-k8s.bak'
            self.assertEqual(backup.read_text(), LEGACY)
            self.assertEqual(backup.stat().st_mode & 0o777, 0o600)
            result = self.migrate(config, validator)
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertEqual(config.read_text(), CLEAN)
            self.assertEqual(backup.read_text(), LEGACY)
            self.assertEqual((data / 'retained-block').read_text(), 'retained metrics')

    def test_invalid_marker_or_unexpected_job_does_not_change_config(self):
        invalid = [LEGACY.replace('#Kubernetes_end', ''),
                   LEGACY.replace('#Kubernetes_start', ''),
                   LEGACY.replace('k8s-cadvisor-cluster-a', 'unrelated-job'),
                   LEGACY + '#Kubernetes_start\n#Kubernetes_end\n',
                   CLEAN + '  - job_name: k8s-cadvisor-outside\n']
        with tempfile.TemporaryDirectory() as tmp:
            directory = Path(tmp)
            validator = self.validator(directory)
            config = directory / 'prometheus.yml'
            for content in invalid:
                with self.subTest(content=content):
                    config.write_text(content)
                    result = self.migrate(config, validator)
                    self.assertNotEqual(result.returncode, 0)
                    self.assertEqual(config.read_text(), content)

    def test_promtool_rejection_leaves_original_config(self):
        with tempfile.TemporaryDirectory() as tmp:
            directory = Path(tmp)
            config = directory / 'prometheus.yml'
            config.write_text(LEGACY)
            result = self.migrate(config, self.validator(directory, 1))
            self.assertNotEqual(result.returncode, 0)
            self.assertEqual(config.read_text(), LEGACY)

    def test_no_markers_and_no_legacy_jobs_is_unchanged(self):
        with tempfile.TemporaryDirectory() as tmp:
            directory = Path(tmp)
            config = directory / 'prometheus.yml'
            content = DEFAULTS
            config.write_text(content)
            result = self.migrate(config, self.validator(directory))
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertEqual(config.read_text(), content)

    def test_startup_preserves_existing_config_but_updates_program(self):
        start = (ROOT / 'build/start.sh').read_text()
        block = start[start.index('# 复制 prometheus_tmp 到 prometheus'):start.index('# 复制 alertmanager_tmp 到 alertmanager')]
        for existing in [True, False]:
            with self.subTest(existing=existing), tempfile.TemporaryDirectory() as tmp:
                directory = Path(tmp)
                source = directory / 'prometheus_tmp'
                target = directory / 'prometheus'
                source.mkdir()
                target.mkdir()
                (source / 'prometheus.yml').write_text(CLEAN)
                (directory / 'prometheus_scrape_defaults.yml').write_text(DEFAULTS)
                (source / 'prometheus').write_text('new executable')
                self.validator(source)
                if existing:
                    (target / 'prometheus.yml').write_text(LEGACY)
                    (target / 'prometheus').write_text('old executable')
                script = block.replace('/app/monitor/prometheus_tmp', str(source))
                script = script.replace('/app/monitor/prometheus', str(target))
                script = 'set -eu\nsudo() { if [ "$1" = chown ]; then return 0; fi; command "$@"; }\n' + script
                result = subprocess.run(['sh', '-c', script], capture_output=True, text=True)
                self.assertEqual(result.returncode, 0, result.stderr)
                self.assertEqual((target / 'prometheus.yml').read_text(),
                                 LEGACY if existing else CLEAN)
                self.assertEqual((target / 'prometheus').read_text(), 'new executable')
                cleanup = start[start.index('# Disable retired jobs'):start.index('\ncd /app/monitor/prometheus')]
                cleanup = cleanup.replace('/app/monitor/disable-legacy-k8s-scrape.sh', str(MIGRATE))
                cleanup = cleanup.replace('/app/monitor/prometheus', str(target))
                for _ in range(2):
                    result = subprocess.run(['sh', '-c', cleanup], capture_output=True, text=True)
                    self.assertEqual(result.returncode, 0, result.stderr)
                    self.assertEqual((target / 'prometheus.yml').read_text(), CLEAN)

    def test_pre_mvp_config_gains_only_missing_managed_jobs(self):
        with tempfile.TemporaryDirectory() as tmp:
            directory = Path(tmp)
            config = directory / 'prometheus.yml'
            config.write_text(LEGACY.replace('  - job_name: k8s_monitor\n    scrape_interval: 30s\n', ''))
            result = self.migrate(config, self.validator(directory))
            self.assertEqual(result.returncode, 0, result.stderr)
            content = config.read_text()
            for name in ['k8s_monitor', 'k8s_log_agent']:
                self.assertEqual(content.count('job_name: ' + name), 1)
            self.assertIn('job_name: snmp-example\n    scrape_interval: 60s', content)
            self.assertNotIn('k8s-cadvisor-', content)
            result = self.migrate(config, self.validator(directory))
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertEqual(config.read_text(), content)

    def test_startup_quarantines_invalid_custom_scrape_before_migration(self):
        start = (ROOT / 'build/start.sh').read_text()
        block = start[start.index('# 启动前校验'):start.index('\ncd /app/monitor/prometheus')]
        with tempfile.TemporaryDirectory() as tmp:
            directory = Path(tmp)
            target = directory / 'prometheus'
            custom = target / 'custom_scrape'
            custom.mkdir(parents=True)
            (custom / 'invalid.yml').write_text('invalid custom config')
            (target / 'prometheus.yml').write_text(LEGACY)
            (directory / 'prometheus_scrape_defaults.yml').write_text(DEFAULTS)
            promtool = target / 'promtool'
            promtool.write_text('#!/bin/sh\n[ ! -f "' + str(custom / 'invalid.yml') + '" ]\n')
            promtool.chmod(0o755)
            block = block.replace('/app/monitor/disable-legacy-k8s-scrape.sh', str(MIGRATE))
            block = block.replace('/app/monitor/deduplicate-remote-write.sh',
                                  str(ROOT / 'build/deduplicate-remote-write.sh'))
            block = block.replace('/app/monitor/prometheus', str(target))
            block = block.replace('/tmp/promtool_check.log', str(directory / 'check.log'))
            block = 'sudo() { if [ "$1" = chown ]; then return 0; fi; command "$@"; }\n' + block
            result = subprocess.run(['sh', '-c', block], capture_output=True, text=True)
            self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
            self.assertEqual((target / 'prometheus.yml').read_text(), CLEAN)
            self.assertFalse((custom / 'invalid.yml').exists())
            self.assertTrue(list(target.glob('custom_scrape_bad_*/invalid.yml')))

    def test_verification_checks_loaded_config_and_active_targets(self):
        old_target = {'labels': {'job': 'k8s-cadvisor-cluster-a'}, 'health': 'up'}
        cases = [('before', LEGACY, LEGACY, [old_target], 0),
                 ('after', CLEAN, CLEAN, [], 0),
                 ('after', CLEAN, LEGACY, [], 1),
                 ('after', LEGACY, CLEAN, [], 1),
                 ('after', CLEAN, CLEAN, [old_target], 1)]
        with tempfile.TemporaryDirectory() as tmp:
            directory = Path(tmp)
            kubectl = directory / 'kubectl'
            kubectl.write_text('''#!/usr/bin/env python3
import os, sys
if 'cat' in sys.argv:
    print(os.environ['TEST_DISK_CONFIG'])
elif sys.argv[-1].endswith('status/config'):
    print(os.environ['TEST_LOADED_RESPONSE'])
else:
    print(os.environ['TEST_TARGET_RESPONSE'])
''')
            kubectl.chmod(0o755)
            for phase, disk, loaded, targets, expected in cases:
                with self.subTest(phase=phase, disk=disk, loaded=loaded, targets=targets):
                    env = dict(os.environ, PATH=str(directory) + os.pathsep + os.environ['PATH'],
                               TEST_DISK_CONFIG=disk,
                               TEST_LOADED_RESPONSE=json.dumps({'status': 'success', 'data': {'yaml': loaded}}),
                               TEST_TARGET_RESPONSE=json.dumps({'status': 'success', 'data': {'activeTargets': targets}}))
                    result = subprocess.run(['sh', str(ROOT / 'build/verify-legacy-k8s-scrape.sh'),
                                             'monitor-scope', 'monitor-pod', phase],
                                            env=env, capture_output=True, text=True)
                    self.assertEqual(result.returncode, expected, result.stdout + result.stderr)
                    self.assertNotIn('bearer_token_file', result.stdout)


if __name__ == '__main__':
    unittest.main()
