package k8s

import "sync"

var runtimeState struct {
	sync.RWMutex
	index    *Index
	exporter *Exporter
}

func ConfigureRuntime(index *Index, exporter *Exporter) {
	runtimeState.Lock()
	runtimeState.index = index
	runtimeState.exporter = exporter
	runtimeState.Unlock()
}
func RuntimeExporter() *Exporter {
	runtimeState.RLock()
	defer runtimeState.RUnlock()
	return runtimeState.exporter
}
