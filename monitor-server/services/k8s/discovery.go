package k8s

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

type ObjectKey struct {
	ClusterID     string `json:"clusterId"`
	Namespace     string `json:"namespace"`
	WorkloadType  string `json:"workloadType"`
	WorkloadName  string `json:"workloadName"`
	PodName       string `json:"podName"`
	ContainerName string `json:"containerName"`
}
type Container struct {
	Key                   ObjectKey `json:"key"`
	WorkloadUID           string    `json:"workloadUid"`
	PodUID                string    `json:"podUid"`
	ContainerID           string    `json:"containerId"`
	Node                  string    `json:"node"`
	Phase                 string    `json:"phase"`
	StatusKnown           bool      `json:"statusKnown"`
	PodReady              string    `json:"podReady"`
	LastTerminationReason string    `json:"lastTerminationReason"`
	Ready                 bool      `json:"ready"`
	RestartCount          int       `json:"restartCount"`
	Resources             Resources `json:"resources"`
	InstanceID            string    `json:"instanceId"`
}
type Resources struct {
	Requests map[string]string `json:"requests"`
	Limits   map[string]string `json:"limits"`
}
type owner struct {
	Kind, Name, UID string
	Controller      bool
}
type metadata struct {
	Name, Namespace, UID string
	Labels               map[string]string
	OwnerReferences      []owner
}
type selector struct {
	MatchLabels      map[string]string
	MatchExpressions []struct {
		Key, Operator string
		Values        []string
	}
}
type Workload struct {
	Metadata metadata
	Spec     struct {
		Replicas *int
		Selector selector
	}
	Status struct{ Replicas, ReadyReplicas, UpdatedReplicas int }
}
type pod struct {
	Metadata metadata
	Spec     struct {
		NodeName   string
		Containers []struct {
			Name      string
			Resources Resources
		}
	}
	Status struct {
		Phase             string
		Conditions        []struct{ Type, Status string }
		ContainerStatuses []struct {
			Name, ContainerID string
			Ready             bool
			RestartCount      int
			LastState         struct{ Terminated *struct{ Reason string } }
		}
	}
}
type Snapshot struct {
	DiscoveryFailed bool        `json:"discoveryFailed,omitempty"`
	ObservedAt      time.Time   `json:"observedAt"`
	WorkloadVersion string      `json:"workloadVersion"`
	PodVersion      string      `json:"podVersion"`
	Workloads       []Workload  `json:"workloads"`
	Containers      []Container `json:"containers"`
}

func (s selector) matches(labels map[string]string) bool {
	for key, value := range s.MatchLabels {
		if v, ok := labels[key]; !ok || v != value {
			return false
		}
	}
	for _, expr := range s.MatchExpressions {
		value, present := labels[expr.Key]
		found := false
		for _, v := range expr.Values {
			if v == value {
				found = true
			}
		}
		switch expr.Operator {
		case "In":
			if !present || !found {
				return false
			}
		case "NotIn":
			if present && found {
				return false
			}
		case "Exists":
			if !present {
				return false
			}
		case "DoesNotExist":
			if present {
				return false
			}
		default:
			return false
		}
	}
	return true
}
func (c *Client) Snapshot(ctx context.Context, cluster string) (*Snapshot, error) {
	if cluster == "" {
		return nil, errors.New("cluster ID is required")
	}
	workloads, wv, err := c.list(ctx, "statefulsets")
	if err != nil {
		return nil, err
	}
	pods, pv, err := c.list(ctx, "pods")
	if err != nil {
		return nil, err
	}
	result := &Snapshot{WorkloadVersion: wv, PodVersion: pv, Workloads: []Workload{}, Containers: []Container{}}
	byUID := map[string]Workload{}
	for _, raw := range workloads {
		var w Workload
		if json.Unmarshal(raw, &w) != nil || w.Metadata.UID == "" || w.Metadata.Name == "" || w.Metadata.Namespace != c.namespace {
			return nil, errors.New("invalid StatefulSet identity")
		}
		byUID[w.Metadata.UID] = w
		result.Workloads = append(result.Workloads, w)
	}
	for _, raw := range pods {
		var p pod
		if json.Unmarshal(raw, &p) != nil || p.Metadata.UID == "" || p.Metadata.Name == "" || p.Metadata.Namespace != c.namespace {
			return nil, errors.New("invalid Pod identity")
		}
		for _, ref := range p.Metadata.OwnerReferences {
			w, ok := byUID[ref.UID]
			if !ok || !ref.Controller || ref.Kind != "StatefulSet" || ref.Name != w.Metadata.Name || !w.Spec.Selector.matches(p.Metadata.Labels) {
				continue
			}
			for _, spec := range p.Spec.Containers {
				item := Container{Key: ObjectKey{ClusterID: cluster, Namespace: c.namespace, WorkloadType: "StatefulSet", WorkloadName: w.Metadata.Name, PodName: p.Metadata.Name, ContainerName: spec.Name}, WorkloadUID: w.Metadata.UID, PodUID: p.Metadata.UID, Node: p.Spec.NodeName, Phase: p.Status.Phase, Resources: spec.Resources, InstanceID: p.Metadata.Labels["instanceId"]}
				for _, status := range p.Status.ContainerStatuses {
					if status.Name == spec.Name {
						item.StatusKnown = true
						if status.LastState.Terminated != nil {
							item.LastTerminationReason = status.LastState.Terminated.Reason
						}
						item.ContainerID = status.ContainerID
						item.Ready = status.Ready
						item.RestartCount = status.RestartCount
						break
					}
				}
				for _, condition := range p.Status.Conditions {
					if condition.Type == "Ready" {
						item.PodReady = condition.Status
					}
				}
				result.Containers = append(result.Containers, item)
			}
			break
		}
	}
	return result, nil
}
