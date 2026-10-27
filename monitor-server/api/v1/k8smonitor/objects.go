package k8smonitor

import (
	"errors"
	mid "github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/k8s"
	"github.com/gin-gonic/gin"
	"strings"
)

// Objects returns the discovered logical monitoring objects and their alert
// states in one namespace-scoped view. Resource and log metrics remain
// queryable through Prometheus using the same object labels.
func (h Handler) Objects(c *gin.Context) {
	cluster, namespace := strings.TrimSpace(c.Query("clusterId")), strings.TrimSpace(c.Query("namespace"))
	if h.Policy.Validate(k8s.Registration{ClusterID: cluster, Namespace: namespace}) != nil || h.AlertStore == nil {
		mid.ReturnValidateError(c, "invalid object scope")
		return
	}
	rows, err := h.Store.List(c.Request.Context())
	if err != nil {
		mid.ReturnServerHandleError(c, errors.New("object scope lookup failed"))
		return
	}
	var snapshot *k8s.Snapshot
	state := ""
	for _, row := range rows {
		if row.Enabled && row.ClusterID == cluster && row.Namespace == namespace {
			snapshot = row.Snapshot
			state = row.State
			break
		}
	}
	if state == "" {
		mid.ReturnValidateError(c, "namespace is not enabled")
		return
	}
	if snapshot == nil {
		snapshot = &k8s.Snapshot{}
	}
	states, err := h.AlertStore.ListStates(c, cluster, namespace)
	if err != nil {
		mid.ReturnServerHandleError(c, errors.New("object alert lookup failed"))
		return
	}
	byObject := map[k8s.ObjectKey][]k8s.AlertStateView{}
	for _, state := range states {
		key := k8s.ObjectKey{ClusterID: state.Key.ClusterID, Namespace: state.Key.Namespace, WorkloadType: state.Key.WorkloadType, WorkloadName: state.Key.WorkloadName, PodName: state.Key.PodName, ContainerName: state.Key.Container}
		byObject[key] = append(byObject[key], state)
	}
	objects := make([]map[string]interface{}, 0, len(snapshot.Containers))
	for _, container := range snapshot.Containers {
		key := container.Key
		objects = append(objects, map[string]interface{}{"clusterId": cluster, "namespace": namespace, "workloadType": container.Key.WorkloadType, "workloadName": container.Key.WorkloadName, "podName": container.Key.PodName, "container": container.Key.ContainerName, "workloadUid": container.WorkloadUID, "podUid": container.PodUID, "instanceId": container.InstanceID, "node": container.Node, "phase": container.Phase, "ready": container.Ready, "statusKnown": container.StatusKnown, "restartCount": container.RestartCount, "alerts": byObject[key]})
	}
	var observedAt interface{}
	if !snapshot.ObservedAt.IsZero() {
		observedAt = snapshot.ObservedAt
	}
	mid.ReturnSuccessData(c, map[string]interface{}{"objects": objects, "state": state, "observedAt": observedAt, "localCollectionEnabled": h.Policy.LocalCollectionActive(), "alertEvaluationEnabled": h.Policy.LocalCollectionActive()})
}
