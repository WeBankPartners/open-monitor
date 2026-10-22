package db

import (
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/k8s"
)

// KubernetesMonitoringStore bypasses ORM caching for version and lease checks.
func KubernetesMonitoringStore() k8s.Store {
	return k8s.Store{DB: x.DB().DB, Seed: models.Config().EncryptSeed}
}

func KubernetesAlertStore() k8s.AlertStore { return k8s.AlertStore{DB: x.DB().DB} }
