"""Exercise the actual startup config rendering with isolated filesystem paths."""
import os
from pathlib import Path
import re
import subprocess
import tempfile
import unittest
import xml.etree.ElementTree as ET

BUILD = Path(__file__).resolve().parents[1]


class KubernetesScrapeIntervalTest(unittest.TestCase):
    def test_interval_updates_only_kubernetes_jobs(self):
        start = (BUILD / "start.sh").read_text()
        block = start[start.index('# Keep Kubernetes exporter persistence'):start.index('# 存量 prometheus.yml')]
        base = (BUILD / "conf/prometheus.yml").read_text()
        base += "\n  - job_name: 'k8s_log_agent_extra'\n    scrape_interval: 17s\n"
        for value, expected in [(None, 30), ('60', 60), ('300', 300), ('15', 15),
                                ('bad', 30), ('0', 30), ('301', 30), ('0060', 60),
                                (' 60 ', 60), ('+60', 60)]:
            with self.subTest(value=value), tempfile.TemporaryDirectory() as directory:
                config = Path(directory) / 'prometheus.yml'
                output = Path(directory) / 'rendered.yml'
                config.write_text(base)
                script = block.replace('/app/monitor/prometheus/prometheus.yml', str(config))
                script = script.replace('/tmp/prometheus.yml.k8s_interval', str(output))
                script = 'set -eu\nsudo() { if [ "$1" = chown ]; then return 0; fi; command "$@"; }\n' + script
                env = dict(os.environ)
                env.pop('MONITOR_K8S_METRICS_INTERVAL_SECONDS', None)
                if value is not None:
                    env['MONITOR_K8S_METRICS_INTERVAL_SECONDS'] = value
                subprocess.run(['bash', '-c', script], env=env, check=True, capture_output=True)
                rendered = config.read_text()
                for job in ['k8s_monitor', 'k8s_log_agent']:
                    base = re.sub(r"(job_name: '" + job + r"'\n    scrape_interval: )\d+s",
                                  lambda m: m[1] + str(expected) + 's', base)
                self.assertEqual(rendered, base)
                subprocess.run(['bash', '-c', script], env=env, check=True, capture_output=True)
                self.assertEqual(config.read_text(), rendered)

    def test_default_template_and_plugin_parameter(self):
        config = (BUILD / 'conf/prometheus.yml').read_text()
        for job in ['k8s_monitor', 'k8s_log_agent']:
            self.assertRegex(config, "job_name: '" + job + "'\n    scrape_interval: 30s")
        root = ET.parse(BUILD / 'register.xml').getroot()
        parameter = next(e for e in root.iter('systemParameter')
                         if e.get('name') == 'MONITOR_K8S_METRICS_INTERVAL_SECONDS')
        self.assertEqual(parameter.get('defaultValue'), '30')
        self.assertTrue(any('MONITOR_K8S_METRICS_INTERVAL_SECONDS={{MONITOR_K8S_METRICS_INTERVAL_SECONDS}}'
                            in e.get('envVariables', '') for e in root.iter('docker')))


if __name__ == '__main__':
    unittest.main()
