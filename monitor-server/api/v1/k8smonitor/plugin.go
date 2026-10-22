package k8smonitor

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/WeBankPartners/open-monitor/monitor-server/services/k8s"
	"github.com/gin-gonic/gin"
)

type pluginInput struct {
	APIServer         string `json:"apiServer"`
	CA                string `json:"ca"`
	Token             string `json:"token"`
	Guid              string `json:"guid"`
	CallbackParameter string `json:"callbackParameter"`
	ClusterID         string `json:"clusterId"`
	Namespace         string `json:"namespace"`
}

type pluginOutput struct {
	Guid              string `json:"guid"`
	CallbackParameter string `json:"callbackParameter"`
	ErrorCode         string `json:"errorCode"`
	ErrorMessage      string `json:"errorMessage"`
}

type pluginResult struct {
	ResultCode    string `json:"resultCode"`
	ResultMessage string `json:"resultMessage"`
	Results       struct {
		Outputs []pluginOutput `json:"outputs"`
	} `json:"results"`
}

// Plugin handles the WeCube batch envelope. Versioning is internal: callers
// send the desired registration and retries are made idempotent by comparing
// it with the current stored registration before saving.
func (h Handler) Plugin(c *gin.Context) {
	result := pluginResult{ResultCode: "0", ResultMessage: "OK"}
	result.Results.Outputs = []pluginOutput{}
	defer func() { c.JSON(http.StatusOK, result) }()
	fail := func(message string) { result.ResultCode = "1"; result.ResultMessage = message }
	var request struct {
		RequestID string        `json:"requestId"`
		Inputs    []pluginInput `json:"inputs"`
	}
	decoder := json.NewDecoder(http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20))
	if err := decoder.Decode(&request); err != nil {
		fail("invalid plugin request")
		return
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF || request.RequestID == "" || len(request.Inputs) == 0 || len(request.Inputs) > 100 {
		fail("invalid plugin request")
		return
	}
	action := c.Param("action")
	if action != "put" && action != "delete" {
		fail("unsupported action")
		return
	}
	for _, input := range request.Inputs {
		output := pluginOutput{Guid: input.Guid, CallbackParameter: input.CallbackParameter, ErrorCode: "0"}
		var applyErr error
		if h.Store == nil {
			applyErr = errors.New("registration store unavailable")
		} else {
			applyErr = h.applyPluginInput(c, action, input)
		}
		if applyErr != nil {
			output.ErrorCode = "1"
			output.ErrorMessage = applyErr.Error()
			fail("one or more inputs failed")
		}
		result.Results.Outputs = append(result.Results.Outputs, output)
	}
}

func (h Handler) applyPluginInput(c *gin.Context, action string, in pluginInput) error {
	in.ClusterID = strings.TrimSpace(in.ClusterID)
	in.Namespace = strings.TrimSpace(in.Namespace)
	in.Guid = strings.TrimSpace(in.Guid)
	if in.Guid == "" || len(in.Guid) > 128 {
		return errors.New("guid is required")
	}
	if h.ClusterMode {
		if err := h.Policy.ValidateCluster(in.ClusterID); err != nil {
			return err
		}
	} else if err := h.Policy.Validate(k8s.Registration{ClusterID: in.ClusterID, Namespace: in.Namespace}); err != nil {
		return err
	}
	rows, err := h.Store.List(c)
	if err != nil {
		return errors.New("registration lookup failed")
	}
	var current *k8s.NamespaceState
	for i := range rows {
		if rows[i].ClusterID == in.ClusterID && (h.ClusterMode || rows[i].Namespace == in.Namespace) {
			current = &rows[i]
			break
		}
	}
	if action == "delete" {
		if current == nil || !current.Enabled {
			return nil
		}
		registration := current.Registration
		registration.Enabled = false
		if err := h.Store.Save(c, registration); err != nil {
			return errors.New("registration delete failed")
		}
		return nil
	}

	registration := k8s.Registration{ClusterID: in.ClusterID, Namespace: in.Namespace, Enabled: true}
	if current != nil {
		registration.Version = current.Version
	}
	if h.ClusterMode {
		if current != nil && strings.TrimSpace(in.Token) == "" {
			registration.Credentials = current.Credentials
			if strings.TrimSpace(in.APIServer) != "" {
				registration.Credentials.APIServer = strings.TrimSpace(in.APIServer)
			}
			registration.Credentials.CA = in.CA
		} else {
			credentials, err := k8s.PrepareCredentials(in.Guid, h.Seed, strings.TrimSpace(in.APIServer), in.CA, in.Token)
			if err != nil {
				return err
			}
			registration.Credentials = credentials
		}
		if registration.Credentials.APIServer == "" {
			return errors.New("apiServer is required")
		}
		plain, err := registration.Credentials.PlainToken(h.Seed)
		if err != nil || plain == "" {
			return errors.New("cluster credential is invalid")
		}
		client, err := k8s.NewClient(registration.Credentials.APIServer, []byte(registration.Credentials.CA), plain, "default")
		if err != nil {
			return err
		}
		client.Close()
	} else {
		credentials, err := k8s.PrepareOptionalToken(in.Guid, h.Seed, in.Token)
		if err != nil {
			return err
		}
		registration.Credentials = credentials
	}
	if current != nil && samePluginRegistration(*current, registration, h.Seed) {
		return nil
	}
	if err := h.Store.Save(c, registration); err != nil {
		return errors.New("registration save failed")
	}
	return nil
}

func samePluginRegistration(current k8s.NamespaceState, desired k8s.Registration, seed string) bool {
	if !current.Enabled || current.ClusterID != desired.ClusterID || current.Namespace != desired.Namespace {
		return false
	}
	if current.Credentials.APIServer != desired.Credentials.APIServer || current.Credentials.CA != desired.Credentials.CA {
		return false
	}
	if current.Credentials.EncryptedToken == "" && desired.Credentials.EncryptedToken == "" {
		return true
	}
	a, errA := current.Credentials.PlainToken(seed)
	b, errB := desired.Credentials.PlainToken(seed)
	return errA == nil && errB == nil && a == b
}
