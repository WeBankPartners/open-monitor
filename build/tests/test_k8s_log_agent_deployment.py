"""Check deployment wiring and run the real PVC initialization script locally."""
import os
from pathlib import Path
import re
import subprocess
import tempfile
import unittest
import xml.etree.ElementTree as ET

import yaml


DEPLOY = Path(__file__).resolve().parents[2] / "monitor-agent/k8s-log-agent/deploy"


class KubernetesLogAgentDeploymentTest(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        with (DEPLOY / "daemonset.yaml").open() as stream:
            cls.pvc, cls.config, cls.daemonset = yaml.safe_load_all(stream)
        cls.pod = cls.daemonset["spec"]["template"]["spec"]
        cls.agent = cls.pod["containers"][0]
        cls.init = cls.pod["initContainers"][0]

    def test_configmap_contains_only_non_secret_runtime_values(self):
        self.assertEqual("ConfigMap", self.config["kind"])
        self.assertEqual("monitor-k8s-log-agent-config", self.config["metadata"]["name"])
        self.assertEqual("DaemonSet", self.daemonset["kind"])
        self.assertEqual("monitor-k8s-log-agent", self.daemonset["metadata"]["name"])
        self.assertEqual(self.config["metadata"]["namespace"], self.daemonset["metadata"]["namespace"])
        data = self.config["data"]
        for key in [
            "TZ",
            "LOG_AGENT_MONITOR_URL",
            "LOG_AGENT_CLUSTER_ID",
            "LOG_AGENT_ROOT",
            "LOG_AGENT_STATE_DIR",
            "LOG_AGENT_HTTP_ADDR",
            "LOG_AGENT_POLL_INTERVAL",
        ]:
            self.assertIsInstance(data[key], str)
        self.assertFalse(any(re.search(r"TOKEN|PASSWORD|SECRET|PRIVATE_KEY", key) for key in data))
        self.assertEqual("REPLACE_MONITOR_URL", data["LOG_AGENT_MONITOR_URL"])
        self.assertNotIn("REPLACE_MONITOR_SERVICE", (DEPLOY / "daemonset.yaml").read_text())
        self.assertNotIn("REPLACE_MONITOR_SERVICE_PORT", (DEPLOY / "daemonset.yaml").read_text())

    def test_register_xml_exposes_fixed_monitor_service_port(self):
        register = ET.parse(Path(__file__).resolve().parents[1] / "register.xml")
        resource = register.find(".//docker")
        self.assertIsNotNone(resource)
        mappings = resource.attrib["portBindings"].split(",")
        self.assertIn("19191:8080", mappings)

    def test_daemonset_consumes_configmap_and_matches_monitor_token_default(self):
        self.assertEqual("RollingUpdate", self.daemonset["spec"]["updateStrategy"]["type"])
        self.assertEqual(1, self.daemonset["spec"]["updateStrategy"]["rollingUpdate"]["maxUnavailable"])
        self.assertEqual("REPLACE_INIT_IMAGE_PREFIX/REPLACE_AGENT_IMAGE", self.agent["image"])
        self.assertTrue(self.agent["image"].startswith("REPLACE_INIT_IMAGE_PREFIX/"))
        self.assertEqual("monitor-k8s-log-agent", self.daemonset["spec"]["selector"]["matchLabels"]["app"])
        self.assertEqual(self.daemonset["spec"]["selector"]["matchLabels"], self.daemonset["spec"]["template"]["metadata"]["labels"])
        self.assertIn("IMAGE ?= monitor-k8s-log-agent:dev", (DEPLOY.parents[0] / "Makefile").read_text())
        self.assertEqual(self.config["metadata"]["name"], self.agent["envFrom"][0]["configMapRef"]["name"])
        env = {item["name"]: item for item in self.agent["env"]}
        register = ET.parse(Path(__file__).resolve().parents[1] / "register.xml")
        parameter = register.find(".//systemParameter[@name='MONITOR_K8S_LOG_AGENT_TOKEN']")
        self.assertIsNotNone(parameter)
        self.assertEqual(parameter.attrib["defaultValue"], env["LOG_AGENT_AUTH_TOKEN"]["value"])
        self.assertEqual("Y", parameter.attrib["sensitiveData"])
        self.assertEqual("status.podIP", env["LOG_AGENT_ADVERTISE_ADDRESS"]["valueFrom"]["fieldRef"]["fieldPath"])
        mounts = {item["name"]: item for item in self.agent["volumeMounts"]}
        volumes = {item["name"]: item for item in self.pod["volumes"]}
        self.assertEqual(self.config["data"]["LOG_AGENT_ROOT"], mounts["logs"]["mountPath"])
        self.assertEqual(mounts["logs"]["mountPath"], volumes["logs"]["hostPath"]["path"])
        self.assertTrue(mounts["logs"]["readOnly"])
        self.assertIn("persistentVolumeClaim", volumes["state"])
        self.assertNotIn("hostPath", volumes["state"])
        self.assertEqual(self.pvc["metadata"]["name"], volumes["state"]["persistentVolumeClaim"]["claimName"])

    def test_yaml_creates_rw_nfs_state_claim_before_daemonset(self):
        self.assertEqual("PersistentVolumeClaim", self.pvc["kind"])
        self.assertEqual("monitor-k8s-log-agent-state", self.pvc["metadata"]["name"])
        self.assertEqual(self.daemonset["metadata"]["namespace"], self.pvc["metadata"]["namespace"])
        self.assertEqual("nfs", self.pvc["spec"]["storageClassName"])
        self.assertEqual(["ReadWriteMany"], self.pvc["spec"]["accessModes"])
        self.assertEqual("1Gi", self.pvc["spec"]["resources"]["requests"]["storage"])

    def test_networkpolicy_targets_monitor_agent_resources(self):
        with (DEPLOY / "networkpolicy.yaml").open() as stream:
            policy = yaml.safe_load(stream)
        self.assertEqual("monitor-k8s-log-agent", policy["metadata"]["name"])
        self.assertEqual(self.daemonset["spec"]["selector"]["matchLabels"], policy["spec"]["podSelector"]["matchLabels"])

    def test_init_uses_existing_environment_image_and_overrides_entrypoint(self):
        self.assertEqual(
            "REPLACE_INIT_IMAGE_PREFIX/package-init-container:1.0.0",
            self.init["image"],
        )
        self.assertEqual(["sh", "-c"], self.init["command"][:2])
        self.assertIn("mkdir -p", self.init["command"][2])
        self.assertNotIn("busybox", (DEPLOY / "daemonset.yaml").read_text())

    def test_init_creates_the_mounted_subpath_and_preserves_state_on_restart(self):
        state_mount = next(item for item in self.agent["volumeMounts"] if item["name"] == "state")
        init_mount = next(item for item in self.init["volumeMounts"] if item["name"] == "state")
        self.assertEqual(self.config["data"]["LOG_AGENT_STATE_DIR"], state_mount["mountPath"])
        self.assertNotEqual(0, self.init["securityContext"]["runAsUser"])
        self.assertEqual(10001, self.pod["securityContext"]["fsGroup"])
        self.assertEqual("OnRootMismatch", self.pod["securityContext"]["fsGroupChangePolicy"])
        cases = [("monitor-sit", "node-a"), ("monitor-sit", "node-b"), ("monitor-uat", "node-a")]
        paths = []
        with tempfile.TemporaryDirectory() as root:
            for namespace, node in cases:
                fields = {"metadata.namespace": namespace, "spec.nodeName": node}
                container_envs = []
                for container in [self.init, self.agent]:
                    container_envs.append({item["name"]: fields[item["valueFrom"]["fieldRef"]["fieldPath"]]
                                          for item in container["env"] if "valueFrom" in item and "fieldRef" in item["valueFrom"]
                                          and item["valueFrom"]["fieldRef"]["fieldPath"] in fields})
                init_env, agent_env = container_envs
                for variable in re.findall(r"\$\(([^)]+)\)", state_mount["subPathExpr"]):
                    self.assertIn(variable, agent_env, "subPathExpr variable must be injected into the Agent container")
                subpath = re.sub(r"\$\(([^)]+)\)", lambda match: agent_env[match[1]], state_mount["subPathExpr"])
                self.assertNotIn("..", Path(subpath).parts)
                self.assertFalse(Path(subpath).is_absolute())
                state = Path(root) / subpath
                command = list(self.init["command"])
                command[-1] = command[-1].replace(init_mount["mountPath"], root)
                result = subprocess.run(command, env={**os.environ, **init_env}, capture_output=True, text=True)
                self.assertEqual(0, result.returncode, result.stderr)
                self.assertTrue(state.is_dir())
                self.assertEqual(0o700, state.stat().st_mode & 0o777)
                checkpoint = state / "checkpoint.json"
                checkpoint.write_text("committed checkpoint")
                result = subprocess.run(command, env={**os.environ, **init_env}, capture_output=True, text=True)
                self.assertEqual(0, result.returncode, result.stderr)
                self.assertEqual("committed checkpoint", checkpoint.read_text())
                paths.append(state)
            self.assertEqual(len(cases), len(set(paths)))


if __name__ == "__main__":
    unittest.main()
