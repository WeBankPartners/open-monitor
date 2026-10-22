package k8s

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
)

// FetchCAdvisor must be called on a client created with the separate node-proxy credential.
func (c *Client) FetchCAdvisor(ctx context.Context, node string) (string, error) {
	if len(node) == 0 || len(node) > 253 {
		return "", errors.New("invalid node name")
	}
	for _, label := range strings.Split(node, ".") {
		if len(label) > 63 || !namespacePattern.MatchString(label) {
			return "", errors.New("invalid node name")
		}
	}
	u := *c.base
	u.Path = "/api/v1/nodes/" + node + "/proxy/metrics/cadvisor"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return "", errors.New("invalid node proxy request")
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "text/plain; version=0.0.4")
	resp, err := c.http.Do(req)
	if err != nil {
		return "", errors.New("node metrics request failed")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", &StatusError{Code: resp.StatusCode}
	}
	const limit = 32 * 1024 * 1024
	raw, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return "", errors.New("cannot read node metrics")
	}
	if len(raw) > limit {
		return "", errors.New("node metrics response too large")
	}
	return string(raw), nil
}

type CAdvisorSource struct{ Client *Client }

func (s CAdvisorSource) Fetch(ctx context.Context, node string) (string, error) {
	if s.Client == nil {
		return "", errors.New("node metrics client unavailable")
	}
	return s.Client.FetchCAdvisor(ctx, node)
}
