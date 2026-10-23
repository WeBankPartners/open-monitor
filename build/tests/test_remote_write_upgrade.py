"""Exercise automatic repair of duplicate managed remote-write entries on a PVC."""
from pathlib import Path
import os
import shutil
import subprocess
import tempfile
import unittest

ROOT = Path(__file__).resolve().parents[2]
REPAIR = ROOT / 'build/deduplicate-remote-write.sh'
CONFIG = """scrape_configs:
  - job_name: k8s_monitor
  - job_name: k8s_log_agent
#Remote write configuration
remote_write:
#Remote_write_start
  - url: 'https://receiver.example.invalid/write'
  - url: 'https://receiver.example.invalid/write'
  - url: 'https://receiver.example.invalid/write'
#Remote_write_end
rule_files: []
"""
CLEAN = CONFIG.replace("  - url: 'https://receiver.example.invalid/write'\n" * 3,
                       "  - url: 'https://receiver.example.invalid/write'\n")


class RemoteWriteUpgradeTest(unittest.TestCase):
    def repair(self, directory, content, valid=True):
        config = directory / 'prometheus.yml'
        config.write_text(content)
        validator = directory / 'promtool'
        validator.write_text('#!/bin/sh\nexit ' + ('0' if valid else '1') + '\n')
        validator.chmod(0o755)
        result = subprocess.run(['sh', str(REPAIR), str(config), str(validator)],
                                capture_output=True, text=True)
        return config, validator, result

    def test_pvc_duplicates_repaired_and_original_backup_retained(self):
        with tempfile.TemporaryDirectory() as tmp:
            directory = Path(tmp)
            config, validator, result = self.repair(directory, CONFIG)
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertEqual(config.read_text(), CLEAN)
            self.assertIn('removed=2', result.stdout)
            backup = directory / 'prometheus.yml.remote-write-duplicates.bak'
            self.assertEqual(backup.read_text(), CONFIG)
            self.assertEqual(backup.stat().st_mode & 0o777, 0o600)
            result = subprocess.run(['sh', str(REPAIR), str(config), str(validator)],
                                    capture_output=True, text=True)
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertEqual(config.read_text(), CLEAN)
            self.assertEqual(backup.read_text(), CONFIG)

    def test_same_url_with_different_options_is_preserved(self):
        content = CONFIG.replace("  - url: 'https://receiver.example.invalid/write'\n" * 3,
                                 "  - url: 'https://receiver.example.invalid/write'\n    name: first\n"
                                 "  - url: 'https://receiver.example.invalid/write'\n    name: second\n")
        with tempfile.TemporaryDirectory() as tmp:
            config, _, result = self.repair(Path(tmp), content)
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertEqual(config.read_text(), content)

    def test_complete_duplicate_blocks_are_removed(self):
        entry = "  - url: 'https://receiver.example.invalid/write'\n    queue_config:\n      capacity: 1000\n"
        content = CONFIG.replace("  - url: 'https://receiver.example.invalid/write'\n" * 3, entry * 3)
        with tempfile.TemporaryDirectory() as tmp:
            config, _, result = self.repair(Path(tmp), content)
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertEqual(config.read_text(), content.replace(entry * 3, entry))

    def test_unmanaged_and_empty_configs_are_unchanged(self):
        for content in [CONFIG.replace('#Remote_write_start\n', '').replace('#Remote_write_end\n', ''),
                        'remote_write: []\n',
                        'remote_write:\n#Remote_write_start\n#Remote_write_end\n']:
            with self.subTest(content=content), tempfile.TemporaryDirectory() as tmp:
                config, _, result = self.repair(Path(tmp), content)
                self.assertEqual(result.returncode, 0, result.stderr)
                self.assertEqual(config.read_text(), content)

    def test_invalid_markers_and_validation_failure_leave_original(self):
        cases = [(CONFIG.replace('#Remote_write_end', ''), True),
                 (CONFIG + '#Remote_write_start\n#Remote_write_end\n', True),
                 (CONFIG.replace('#Remote_write_start', '#Remote_write_end'), True),
                 (CONFIG, False)]
        for content, valid in cases:
            with self.subTest(valid=valid), tempfile.TemporaryDirectory() as tmp:
                config, _, result = self.repair(Path(tmp), content, valid)
                self.assertNotEqual(result.returncode, 0)
                self.assertEqual(config.read_text(), content)
                self.assertNotIn('receiver.example.invalid', result.stdout + result.stderr)

    def test_concurrent_repairs_keep_one_private_original_backup(self):
        with tempfile.TemporaryDirectory() as tmp:
            directory = Path(tmp)
            config = directory / 'prometheus.yml'
            config.write_text(CONFIG)
            validator = directory / 'promtool'
            validator.write_text('#!/bin/sh\nexit 0\n')
            validator.chmod(0o755)
            processes = [subprocess.Popen(['sh', str(REPAIR), str(config), str(validator)],
                                          stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
                         for _ in range(4)]
            for process in processes:
                stdout, stderr = process.communicate(timeout=10)
                self.assertEqual(process.returncode, 0, stdout + stderr)
            self.assertEqual(config.read_text(), CLEAN)
            self.assertEqual((directory / 'prometheus.yml.remote-write-duplicates.bak').read_text(), CONFIG)

    def test_repair_runs_with_busybox_applets(self):
        busybox = shutil.which('busybox')
        if busybox is None:
            self.skipTest('BusyBox is not installed')
        with tempfile.TemporaryDirectory() as tmp:
            directory = Path(tmp)
            config, validator, _ = self.repair(directory, CONFIG)
            config.write_text(CONFIG)
            applets = directory / 'bin'
            applets.mkdir()
            for name in ['awk', 'cat', 'ln', 'mv', 'chmod', 'mktemp', 'rm']:
                (applets / name).symlink_to(busybox)
            result = subprocess.run([busybox, 'sh', str(REPAIR), str(config), str(validator)],
                                    env={**os.environ, 'PATH': str(applets)},
                                    capture_output=True, text=True)
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertEqual(config.read_text(), CLEAN)

    def test_startup_does_not_launch_prometheus_when_repair_fails(self):
        start = (ROOT / 'build/start.sh').read_text()
        block = start[start.index('# Repair persisted duplicates'):start.index('# Disable retired jobs')]
        with tempfile.TemporaryDirectory() as tmp:
            directory = Path(tmp)
            config, _, _ = self.repair(directory, CONFIG, valid=False)
            script = block.replace('/app/monitor/deduplicate-remote-write.sh', str(REPAIR))
            script = script.replace('/app/monitor/prometheus', str(directory))
            launched = directory / 'launched'
            result = subprocess.run(['sh', '-c', script + '\ntouch "' + str(launched) + '"\n'],
                                    capture_output=True, text=True)
            self.assertNotEqual(result.returncode, 0)
            self.assertFalse(launched.exists())
            self.assertEqual(config.read_text(), CONFIG)

    def test_startup_repairs_before_prometheus_launch_and_images_ship_helper(self):
        start = (ROOT / 'build/start.sh').read_text()
        self.assertIn('sh /app/monitor/deduplicate-remote-write.sh', start)
        self.assertLess(start.index('sh /app/monitor/deduplicate-remote-write.sh'),
                        start.index('nohup ./prometheus --config.file='))
        for name in ['Dockerfile', 'Dockerfile-arm64']:
            self.assertIn('COPY build/deduplicate-remote-write.sh $BASE_HOME/',
                          (ROOT / name).read_text())

    def test_both_upgrade_states_recover_in_one_startup_and_remain_idempotent(self):
        start = (ROOT / 'build/start.sh').read_text()
        block = start[start.index('# 启动前校验'):start.index('\ncd ../ping_exporter/')]
        old_jobs = """#Kubernetes_start
  - job_name: k8s-kubelet-cluster-a
  - job_name: k8s-cadvisor-cluster-a
#Kubernetes_end
"""
        for legacy_active, duplicated in [(True, False), (True, True), (False, True)]:
            with self.subTest(legacy_active=legacy_active, duplicated=duplicated), tempfile.TemporaryDirectory() as tmp:
                directory = Path(tmp)
                target = directory / 'prometheus'
                target.mkdir()
                (target / 'logs').mkdir()
                (target / 'data').mkdir()
                (target / 'data/retained-block').write_text('retained metrics')
                content = (CONFIG if duplicated else CLEAN).replace('#Remote write configuration',
                                         (old_jobs if legacy_active else '#Kubernetes_start\n#Kubernetes_end\n')
                                         + '#Remote write configuration')
                (target / 'prometheus.yml').write_text(content)
                old_backup = target / 'prometheus.yml.legacy-k8s.bak'
                if not legacy_active:
                    old_backup.write_text('historical legacy config backup')
                (directory / 'prometheus_scrape_defaults.yml').write_text(
                    'scrape_configs:\n  - job_name: k8s_monitor\n  - job_name: k8s_log_agent\n')
                validator = target / 'promtool'
                # Reproduce the reported case: syntax validation accepts duplicates.
                validator.write_text('#!/bin/sh\nexit 0\n')
                validator.chmod(0o755)
                prometheus = target / 'prometheus'
                prometheus.write_text('''#!/usr/bin/env python3
from pathlib import Path
config = Path('prometheus.yml').read_text()
assert config.count('https://receiver.example.invalid/write') == 1
assert 'k8s-kubelet-' not in config and 'k8s-cadvisor-' not in config
assert 'job_name: k8s_monitor' in config and 'job_name: k8s_log_agent' in config
Path('ready').write_text('ready')
''')
                prometheus.chmod(0o755)
                script = block.replace('/app/monitor/deduplicate-remote-write.sh', str(REPAIR))
                script = script.replace('/app/monitor/disable-legacy-k8s-scrape.sh',
                                        str(ROOT / 'build/disable-legacy-k8s-scrape.sh'))
                script = script.replace('/app/monitor/prometheus', str(target))
                script = script.replace('/tmp/promtool_check.log', str(directory / 'check.log'))
                script = 'set -eu\nensure_log_file() { touch "$1"; }\narchive_day=30d\n' + script + '\nwait\n'
                for _ in range(2):
                    result = subprocess.run(['sh', '-c', script], capture_output=True, text=True)
                    self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
                    self.assertTrue((target / 'ready').exists(), (target / 'logs/prometheus.log').read_text())
                    self.assertEqual((target / 'data/retained-block').read_text(), 'retained metrics')
                if duplicated:
                    self.assertEqual((target / 'prometheus.yml.remote-write-duplicates.bak').read_text(), content)
                if not legacy_active:
                    self.assertEqual(old_backup.read_text(), 'historical legacy config backup')


if __name__ == '__main__':
    unittest.main()
