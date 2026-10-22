package k8s

import (
	"errors"
	"sort"
)

// RebindLogAgentConfigs reconciles all node-scoped configurations against one
// discovery snapshot. Targets follow the discovered node when a Pod moves.
func RebindLogAgentConfigs(configs map[string]LogAgentConfig, snapshot Snapshot) (map[string]LogAgentConfig, bool, error) {
	return rebindLogAgentConfigs(configs, snapshot, "")
}

// RebindLogAgentConfigsForNamespace updates only targets belonging to the
// scanned namespace and preserves targets owned by other registered namespaces.
func RebindLogAgentConfigsForNamespace(configs map[string]LogAgentConfig, snapshot Snapshot, namespace string) (map[string]LogAgentConfig, bool, error) {
	return rebindLogAgentConfigs(configs, snapshot, namespace)
}

func rebindLogAgentConfigs(configs map[string]LogAgentConfig, snapshot Snapshot, namespace string) (map[string]LogAgentConfig, bool, error) {
	out := make(map[string]LogAgentConfig, len(configs))
	for node, config := range configs {
		if err := ValidateLogAgentConfig(config); err != nil {
			return nil, false, err
		}
		out[node] = LogAgentConfig{Version: config.Version, Targets: nil}
		for _, target := range config.Targets {
			if namespace != "" && target.Namespace != namespace {
				preserved := out[node]
				preserved.Targets = append(preserved.Targets, target)
				out[node] = preserved
			}
		}
	}
	byLogical := make(map[string]Container, len(snapshot.Containers))
	for _, item := range snapshot.Containers {
		if namespace != "" && item.Key.Namespace != namespace {
			continue
		}
		key := item.Key.Namespace + "\x00" + item.Key.WorkloadType + "\x00" + item.Key.WorkloadName + "\x00" + item.Key.PodName + "\x00" + item.Key.ContainerName
		byLogical[key] = item
	}
	changed := false
	for sourceNode, config := range configs {
		for _, target := range config.Targets {
			if namespace != "" && target.Namespace != namespace {
				continue
			}
			key := target.Namespace + "\x00" + target.WorkloadType + "\x00" + target.WorkloadName + "\x00" + target.PodName + "\x00" + target.Container
			item, ok := byLogical[key]
			if !ok || item.Node == "" {
				changed = true
				continue
			}
			oldPodUID, oldWorkloadUID := target.PodUID, target.WorkloadUID
			target.PodUID, target.WorkloadUID = item.PodUID, item.WorkloadUID
			destination := out[item.Node]
			destination.Targets = append(destination.Targets, target)
			out[item.Node] = destination
			if item.Node != sourceNode || oldPodUID != target.PodUID || oldWorkloadUID != target.WorkloadUID {
				changed = true
			}
		}
	}
	for node, config := range out {
		original, existed := configs[node]
		if !existed {
			if len(config.Targets) > 0 {
				config.Version = 1
				out[node] = config
			}
			continue
		}
		if !sameLogTargets(original.Targets, config.Targets) {
			config.Version = original.Version + 1
			out[node] = config
			changed = true
		}
	}
	return out, changed, nil
}

func sameLogTargets(a, b []LogTarget) bool {
	a = append([]LogTarget(nil), a...)
	b = append([]LogTarget(nil), b...)
	logTargetOrder := func(targets []LogTarget) {
		sort.Slice(targets, func(i, j int) bool { return targets[i].ID < targets[j].ID })
	}
	logTargetOrder(a)
	logTargetOrder(b)
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].ID != b[i].ID || a[i].Namespace != b[i].Namespace || a[i].WorkloadType != b[i].WorkloadType || a[i].WorkloadName != b[i].WorkloadName || a[i].PodName != b[i].PodName || a[i].PodUID != b[i].PodUID || a[i].WorkloadUID != b[i].WorkloadUID || a[i].Container != b[i].Container || a[i].InstanceID != b[i].InstanceID || a[i].Directory != b[i].Directory || a[i].Glob != b[i].Glob || a[i].Start != b[i].Start || a[i].Format != b[i].Format {
			return false
		}
		if len(a[i].Rules) != len(b[i].Rules) {
			return false
		}
		for j := range a[i].Rules {
			if a[i].Rules[j].ID != b[i].Rules[j].ID || a[i].Rules[j].Version != b[i].Rules[j].Version || !sameLogRuleDefinition(a[i].Rules[j], b[i].Rules[j]) {
				return false
			}
		}
	}
	return true
}

// RebindLogAgentConfig reconciles manually registered log targets with the
// latest discovery snapshot. Directory/glob/rules remain operator supplied;
// only runtime Pod identity is refreshed. Targets for deleted pods are removed.
func RebindLogAgentConfig(config LogAgentConfig, snapshot Snapshot, node string) (LogAgentConfig, bool, error) {
	if err := ValidateLogAgentConfig(config); err != nil {
		return config, false, err
	}
	if node == "" {
		return config, false, errors.New("node is required")
	}
	byLogical := make([]Container, 0, len(snapshot.Containers))
	for _, item := range snapshot.Containers {
		if item.Node == node {
			byLogical = append(byLogical, item)
		}
	}
	out := LogAgentConfig{Version: config.Version, Targets: make([]LogTarget, 0, len(config.Targets))}
	changed := false
	for _, target := range config.Targets {
		key := ObjectKey{ClusterID: "", Namespace: target.Namespace, WorkloadType: target.WorkloadType, WorkloadName: target.WorkloadName, PodName: target.PodName, ContainerName: target.Container}
		var item Container
		ok := false
		for _, candidate := range byLogical {
			if candidate.Key.Namespace == key.Namespace && candidate.Key.WorkloadType == key.WorkloadType && candidate.Key.WorkloadName == key.WorkloadName && candidate.Key.PodName == key.PodName && candidate.Key.ContainerName == key.ContainerName {
				item, ok = candidate, true
				break
			}
		}
		if !ok {
			changed = true
			continue
		}
		if target.PodUID != item.PodUID || target.WorkloadUID != item.WorkloadUID {
			target.PodUID, target.WorkloadUID = item.PodUID, item.WorkloadUID
			changed = true
		}
		out.Targets = append(out.Targets, target)
	}
	if changed {
		out.Version++
	}
	return out, changed, nil
}
