package k8smonitor

import (
	"encoding/json"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/k8s"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"strconv"
)

type pluginInput struct {
	APIServer         string `json:"apiServer"` // cluster plugin only
	CA                string `json:"ca"`        // cluster plugin only
	Token             string `json:"token"`
	Guid              string `json:"guid"`
	CallbackParameter string `json:"callbackParameter"`
	ClusterID         string `json:"clusterId"`
	Namespace         string `json:"namespace"`
	Version           string `json:"version"`
}
type pluginOutput struct {
	Version           string `json:"version,omitempty"`
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

// Plugin follows the existing WeCube batch invocation envelope. Extra platform
// metadata is intentionally accepted; only declared inputs affect registration.
func (h Handler) Plugin(c *gin.Context) {
	result := pluginResult{ResultCode: "0", ResultMessage: "OK"}
	result.Results.Outputs = []pluginOutput{}
	defer func() { c.JSON(http.StatusOK, result) }()
	fail := func(message string) { result.ResultCode = "1"; result.ResultMessage = message }
	var input struct {
		RequestID string        `json:"requestId"`
		Inputs    []pluginInput `json:"inputs"`
	}
	decoder := json.NewDecoder(http.MaxBytesReader(c.Writer, c.Request.Body, 1024*1024))
	if err := decoder.Decode(&input); err != nil {
		fail("invalid plugin request")
		return
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		fail("invalid plugin request")
		return
	}
	if input.RequestID == "" || len(input.Inputs) == 0 || len(input.Inputs) > 100 {
		fail("requestId and 1 to 100 inputs are required")
		return
	}
	action := c.Param("action")
	if action != "add" && action != "update" && action != "delete" {
		fail("unsupported action")
		return
	}
	for _, row := range input.Inputs {
		output := pluginOutput{Guid: row.Guid, CallbackParameter: row.CallbackParameter, ErrorCode: "0"}
		version, err := strconv.ParseUint(row.Version, 10, 64)
		if err != nil || row.Guid == "" || len(row.Guid) > 128 || version == ^uint64(0) {
			output.ErrorMessage = "guid and numeric version are required"
		} else if (action == "add" && version != 0) || (action != "add" && version == 0) {
			output.ErrorMessage = "add requires version zero; update and delete require an existing version"
		} else {
			registration := k8s.Registration{ClusterID: row.ClusterID, Namespace: row.Namespace, Version: version, Enabled: action != "delete"}
			if action != "delete" {
				var credentials k8s.Credentials
				var credentialErr error
				if h.ClusterMode {
					credentials, credentialErr = k8s.PrepareCredentials(row.Guid, h.Seed, row.APIServer, row.CA, row.Token)
				} else {
					credentials, credentialErr = k8s.PrepareOptionalToken(row.Guid, h.Seed, row.Token)
				}
				if credentialErr != nil {
					output.ErrorMessage = credentialErr.Error()
				} else {
					registration.Credentials = credentials
					plain, decryptErr := credentials.PlainToken(h.Seed)
					if decryptErr != nil {
						output.ErrorMessage = "cannot decrypt credential"
					} else {
						namespace := row.Namespace
						if h.ClusterMode {
							namespace = "default"
						}
						if h.ClusterMode {
							client, clientErr := k8s.NewClient(row.APIServer, []byte(row.CA), plain, namespace)
							if clientErr != nil {
								output.ErrorMessage = clientErr.Error()
							} else {
								client.Close()
							}
						}
					}
				}
			}
			if output.ErrorMessage != "" {
				output.ErrorCode = "1"
				fail("one or more inputs failed")
				result.Results.Outputs = append(result.Results.Outputs, output)
				continue
			}
			if h.ClusterMode {
				err = h.Policy.ValidateCluster(registration.ClusterID)
			} else {
				err = h.Policy.Validate(registration)
			}
			if err != nil {
				output.ErrorMessage = err.Error()
			} else if err = h.Store.Save(c, registration); err != nil {
				if err == k8s.ErrConflict {
					output.ErrorMessage = "namespace version conflict"
				} else {
					output.ErrorMessage = "namespace registration failed"
				}
			}
		}
		if output.ErrorMessage != "" {
			output.ErrorCode = "1"
			fail("one or more inputs failed")
		}
		if output.ErrorCode == "0" {
			output.Version = strconv.FormatUint(version+1, 10)
		}
		result.Results.Outputs = append(result.Results.Outputs, output)
	}
}
