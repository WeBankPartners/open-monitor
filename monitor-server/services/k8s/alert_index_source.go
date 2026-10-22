package k8s

import "context"

type IndexAlertSnapshotSource struct{ Index *Index }

func (s IndexAlertSnapshotSource) Snapshot(ctx context.Context, cluster, namespace string) (*Snapshot, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if s.Index == nil {
		return nil, ErrConflict
	}
	snapshot, ok := s.Index.Current(cluster, namespace)
	if !ok {
		return nil, ErrConflict
	}
	return snapshot, nil
}
