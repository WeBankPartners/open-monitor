package k8s

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type AppInstance struct {
	GUID       string `json:"guid"`
	DeployPath string `json:"deploy_path"`
}
type CMDBClient struct {
	BaseURL string
	Token   func() string
	HTTP    interface {
		Do(*http.Request) (*http.Response, error)
	}
}

func (c CMDBClient) QueryAppInstance(ctx context.Context, guid string) (AppInstance, error) {
	if strings.TrimSpace(guid) == "" {
		return AppInstance{}, errors.New("instance id is required")
	}
	base, err := url.Parse(strings.TrimRight(c.BaseURL, "/"))
	if err != nil || base.Scheme == "" || base.Host == "" {
		return AppInstance{}, errors.New("invalid CMDB URL")
	}
	base.Path = strings.TrimRight(base.Path, "/") + "/wecmdb/api/v1/ci-data/query/app_instance"
	body, err := json.Marshal(map[string]interface{}{"dialect": map[string]string{"queryMode": "new"}, "filters": []map[string]string{{"name": "guid", "operator": "equal", "value": guid}}, "pageable": map[string]int{"pageSize": 1, "startIndex": 0}, "paging": true, "resultColumns": []string{"guid", "deploy_path"}})
	if err != nil {
		return AppInstance{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base.String(), bytes.NewReader(body))
	if err != nil {
		return AppInstance{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.Token != nil {
		req.Header.Set("Authorization", c.Token())
	}
	h := c.HTTP
	if h == nil {
		h = http.DefaultClient
	}
	resp, err := h.Do(req)
	if err != nil {
		return AppInstance{}, errors.New("CMDB query request failed")
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return AppInstance{}, fmt.Errorf("CMDB query status %d", resp.StatusCode)
	}
	var out struct {
		Code       int    `json:"code"`
		StatusCode string `json:"statusCode"`
		Data       struct {
			PageInfo struct {
				TotalRows int `json:"totalRows"`
			} `json:"pageInfo"`
			Contents []AppInstance `json:"contents"`
		} `json:"data"`
	}
	if err = json.NewDecoder(http.MaxBytesReader(nil, resp.Body, 4<<20)).Decode(&out); err != nil {
		return AppInstance{}, errors.New("invalid CMDB response")
	}
	if out.Code != 0 || out.StatusCode != "OK" {
		return AppInstance{}, fmt.Errorf("CMDB query rejected: code=%d", out.Code)
	}
	if out.Data.PageInfo.TotalRows != 1 || len(out.Data.Contents) != 1 {
		return AppInstance{}, fmt.Errorf("CMDB app instance not found: total_rows=%d returned_rows=%d", out.Data.PageInfo.TotalRows, len(out.Data.Contents))
	}
	if strings.TrimSpace(out.Data.Contents[0].DeployPath) == "" {
		return AppInstance{}, errors.New("CMDB deploy path is empty")
	}
	return out.Data.Contents[0], nil
}
