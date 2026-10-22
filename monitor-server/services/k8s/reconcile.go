package k8s

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"
)

type Scanner interface {
	Snapshot(context.Context, string) (*Snapshot, error)
}
type scope struct{ cluster, namespace string }
type entry struct {
	version       uint64
	enabled, busy bool
	snapshot      *Snapshot
}

// Index is a process-local inventory. Database task ownership must fence writers
// before publishing this inventory to shared storage or external configuration.
type Index struct {
	mu      sync.RWMutex
	entries map[scope]*entry
}

func NewIndex() *Index { return &Index{entries: make(map[scope]*entry)} }

// Configure rejects stale versions. Versions must increase for every configuration mutation.
func (i *Index) Configure(cluster, namespace string, version uint64, enabled bool) {
	i.mu.Lock()
	defer i.mu.Unlock()
	key := scope{cluster, namespace}
	previous := i.entries[key]
	if previous != nil && version <= previous.version {
		return
	}
	i.entries[key] = &entry{version: version, enabled: enabled}
}
func (i *Index) Reconcile(ctx context.Context, cluster, namespace string, version uint64, scanner Scanner) (err error) {
	key := scope{cluster, namespace}
	i.mu.Lock()
	e := i.entries[key]
	if e == nil || !e.enabled || e.version != version || e.busy {
		i.mu.Unlock()
		return errors.New("namespace unavailable or reconciliation already running")
	}
	e.busy = true
	i.mu.Unlock()
	defer func() {
		i.mu.Lock()
		e.busy = false
		if err != nil && e.snapshot != nil {
			e.snapshot.DiscoveryFailed = true
		}
		i.mu.Unlock()
	}()
	snapshot, err := scanner.Snapshot(ctx, cluster)
	if err != nil {
		return err
	}
	if snapshot == nil {
		return errors.New("empty snapshot")
	}
	if err = ctx.Err(); err != nil {
		return err
	}
	owned, err := cloneSnapshot(snapshot)
	if err != nil {
		return err
	}
	i.mu.Lock()
	defer i.mu.Unlock()
	if current := i.entries[key]; current != e || !e.enabled || e.version != version {
		return errors.New("namespace configuration changed during reconciliation")
	}
	owned.ObservedAt = time.Now()
	owned.DiscoveryFailed = false
	e.snapshot = owned
	return nil
}

func (i *Index) markDiscoveryFailed(cluster, namespace string, version uint64) {
	i.mu.Lock()
	defer i.mu.Unlock()
	if e := i.entries[scope{cluster, namespace}]; e != nil && e.version == version && e.snapshot != nil {
		e.snapshot.DiscoveryFailed = true
	}
}

// Current returns an isolated copy so readers cannot mutate the published inventory.
func (i *Index) Current(cluster, namespace string) (*Snapshot, bool) {
	i.mu.RLock()
	defer i.mu.RUnlock()
	e := i.entries[scope{cluster, namespace}]
	if e == nil || !e.enabled || e.snapshot == nil {
		return nil, false
	}
	copy, err := cloneSnapshot(e.snapshot)
	return copy, err == nil
}
func cloneSnapshot(snapshot *Snapshot) (*Snapshot, error) {
	data, err := json.Marshal(snapshot)
	if err != nil {
		return nil, err
	}
	var copy Snapshot
	err = json.Unmarshal(data, &copy)
	return &copy, err
}

func (i *Index) Remove(cluster, namespace string) {
	i.mu.Lock()
	delete(i.entries, scope{cluster, namespace})
	i.mu.Unlock()
}
