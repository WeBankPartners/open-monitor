package k8smonitor

import (
	"encoding/json"
	mid "github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"net/http"
	"strings"

	"github.com/WeBankPartners/open-monitor/monitor-server/services/k8s"
	"github.com/gin-gonic/gin"
)

type sourceInput struct {
	Guid      string `json:"guid"`
	ClusterID string `json:"clusterId"`
	Namespace string `json:"namespace"`
	APIServer string `json:"apiServer"`
	CA        string `json:"ca"`
	Token     string `json:"token"`
	Enabled   *bool  `json:"enabled"`
}

func (h Handler) PutSource(c *gin.Context) {
	var in sourceInput
	if json.NewDecoder(http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20)).Decode(&in) != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	enabled := true
	if in.Enabled != nil {
		enabled = *in.Enabled
	}
	h.saveSource(c, in, enabled)
}

func (h Handler) saveSource(c *gin.Context, in sourceInput, enabled bool) {
	r := k8s.Registration{ClusterID: strings.TrimSpace(in.ClusterID), Namespace: strings.TrimSpace(in.Namespace), Enabled: enabled}
	if h.Store == nil {
		c.AbortWithStatus(http.StatusServiceUnavailable)
		return
	}
	if (h.ClusterMode && h.Policy.ValidateCluster(in.ClusterID) != nil) || (!h.ClusterMode && h.Policy.Validate(k8s.Registration{ClusterID: in.ClusterID, Namespace: in.Namespace}) != nil) {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	rows, err := h.Store.List(c)
	if err != nil {
		c.AbortWithStatus(http.StatusServiceUnavailable)
		return
	}
	var existing *k8s.NamespaceState
	for i := range rows {
		if rows[i].ClusterID == r.ClusterID && (h.ClusterMode || rows[i].Namespace == r.Namespace) {
			existing = &rows[i]
			r.Version = rows[i].Version
			break
		}
	}
	if h.ClusterMode && strings.TrimSpace(in.Token) == "" && existing != nil {
		r.Credentials = existing.Credentials
		if h.ClusterMode {
			r.Credentials.APIServer = strings.TrimSpace(in.APIServer)
			r.Credentials.CA = in.CA
		}
	} else {
		if strings.TrimSpace(in.Guid) == "" {
			if h.ClusterMode {
				in.Guid = r.ClusterID
			} else {
				in.Guid = r.ClusterID + "/" + r.Namespace
			}
		}
		if h.ClusterMode {
			creds, err := k8s.PrepareCredentials(in.Guid, h.Seed, in.APIServer, in.CA, in.Token)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			r.Credentials = creds
		} else {
			creds, err := k8s.PrepareOptionalToken(in.Guid, h.Seed, in.Token)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			r.Credentials = creds
		}
	}
	if h.ClusterMode {
		token, decryptErr := r.Credentials.PlainToken(h.Seed)
		if decryptErr != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		client, clientErr := k8s.NewClient(r.Credentials.APIServer, []byte(r.Credentials.CA), token, "default")
		if clientErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": clientErr.Error()})
			return
		}
		client.Close()
	}
	if err = h.Store.Save(c, r); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "source update failed"})
		return
	}
	mid.ReturnSuccess(c)
}

func (h Handler) DeleteSource(c *gin.Context) {
	var in sourceInput
	if c.Request.ContentLength > 0 {
		if json.NewDecoder(http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20)).Decode(&in) != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
	} else {
		in.ClusterID, in.Namespace = c.Query("clusterId"), c.Query("namespace")
	}
	cluster, namespace := strings.TrimSpace(in.ClusterID), strings.TrimSpace(in.Namespace)
	if cluster == "" || (!h.ClusterMode && namespace == "") {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	if (h.ClusterMode && h.Policy.ValidateCluster(cluster) != nil) || (!h.ClusterMode && h.Policy.Validate(k8s.Registration{ClusterID: cluster, Namespace: namespace}) != nil) {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	rows, err := h.Store.List(c)
	if err != nil {
		c.AbortWithStatus(http.StatusServiceUnavailable)
		return
	}
	for _, row := range rows {
		if row.ClusterID == cluster && (h.ClusterMode || row.Namespace == namespace) {
			if !row.Enabled {
				mid.ReturnSuccess(c)
				return
			}
			row.Enabled = false
			if err = h.Store.Save(c, row.Registration); err != nil {
				c.AbortWithStatus(http.StatusConflict)
				return
			}
			mid.ReturnSuccess(c)
			return
		}
	}
	mid.ReturnSuccess(c)
}
