package k8s

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"path/filepath"
	"strings"
)

// CMDBLogTargetResolver creates one node-local log target from a discovered
// container and its app_instance record.
type CMDBLogTargetResolver struct {
	Client CMDBClient
}

func (r CMDBLogTargetResolver) Resolve(ctx context.Context, c Container) (LogTarget, error) {
	if strings.TrimSpace(c.InstanceID) == "" {
		return LogTarget{}, errors.New("instance id is required")
	}
	app, err := r.Client.QueryAppInstance(ctx, c.InstanceID)
	if err != nil {
		return LogTarget{}, err
	}
	deployPath := filepath.Clean(strings.TrimSpace(app.DeployPath))
	if !safeAbsolutePath(deployPath) {
		return LogTarget{}, errors.New("invalid CMDB deploy path")
	}
	path := filepath.Join(deployPath, "logs")
	return LogTarget{
		ID:           automaticLogTargetID(c),
		Namespace:    c.Key.Namespace,
		WorkloadType: c.Key.WorkloadType,
		WorkloadName: c.Key.WorkloadName,
		WorkloadUID:  c.WorkloadUID,
		PodName:      c.Key.PodName,
		PodUID:       c.PodUID,
		Container:    c.Key.ContainerName,
		InstanceID:   c.InstanceID,
		Directory:    path,
		Glob:         "*.log",
		Start:        "end",
		Rules: []LogRule{
			{ID: "default-error", Version: 1, Mode: "regex", Pattern: `(?i)\b(ERROR|FATAL|PANIC)\b`},
		},
	}, nil
}

func automaticLogTargetID(c Container) string {
	h := sha256.New()
	for _, value := range []string{c.Key.ClusterID, c.Key.Namespace, c.Key.WorkloadType, c.Key.WorkloadName, c.Key.PodName, c.Key.ContainerName, c.InstanceID} {
		var size [4]byte
		binary.BigEndian.PutUint32(size[:], uint32(len(value)))
		h.Write(size[:])
		h.Write([]byte(value))
	}
	return "cmdb-" + hex.EncodeToString(h.Sum(nil))
}

func safeAbsolutePath(path string) bool {
	if path == "" || !filepath.IsAbs(path) || path == "/" {
		return false
	}
	for _, part := range strings.Split(filepath.Clean(path), string(filepath.Separator)) {
		if part == ".." {
			return false
		}
	}
	return true
}
