// Package k8s implements namespace-scoped monitoring without node discovery permissions.
package k8s

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var namespacePattern = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)

type Client struct {
	base             *url.URL
	http             *http.Client
	token, namespace string
}
type StatusError struct{ Code int }

func (e *StatusError) Error() string {
	return fmt.Sprintf("kubernetes request failed with status %d", e.Code)
}
func IsStatus(err error, code int) bool {
	var status *StatusError
	return errors.As(err, &status) && status.Code == code
}

// NewClient accepts an administrator-approved HTTPS origin. The caller owns the origin allowlist.
func NewClient(endpoint string, ca []byte, token, namespace string) (*Client, error) {
	u, err := url.Parse(endpoint)
	if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || (u.Path != "" && u.Path != "/") {
		return nil, errors.New("invalid Kubernetes HTTPS origin")
	}
	if len(namespace) > 63 || !namespacePattern.MatchString(namespace) {
		return nil, errors.New("invalid namespace")
	}
	if strings.TrimSpace(token) == "" || strings.ContainsAny(token, "\r\n") {
		return nil, errors.New("invalid Kubernetes credential")
	}
	roots, err := x509.SystemCertPool()
	if err != nil {
		roots = x509.NewCertPool()
	}
	if len(ca) > 0 && !roots.AppendCertsFromPEM(ca) {
		return nil, errors.New("invalid Kubernetes CA")
	}
	tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12, RootCAs: roots}
	if len(ca) == 0 {
		tlsConfig.InsecureSkipVerify = true // configured cluster has no CA; accept its certificate
	}
	transport := &http.Transport{TLSClientConfig: tlsConfig, TLSHandshakeTimeout: 10 * time.Second, ResponseHeaderTimeout: 15 * time.Second, IdleConnTimeout: 90 * time.Second, MaxIdleConnsPerHost: 4}
	return &Client{base: u, token: token, namespace: namespace, http: &http.Client{Transport: transport, Timeout: 75 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return errors.New("Kubernetes redirects are disabled") }}}, nil
}
func (c *Client) Close() { c.http.CloseIdleConnections() }
func (c *Client) resourcePath(resource string) (string, error) {
	switch resource {
	case "pods":
		return "/api/v1/namespaces/" + c.namespace + "/pods", nil
	case "statefulsets":
		return "/apis/apps/v1/namespaces/" + c.namespace + "/statefulsets", nil
	}
	return "", errors.New("unsupported Kubernetes resource")
}
func (c *Client) get(ctx context.Context, resource string, query url.Values) (*http.Response, error) {
	path, err := c.resourcePath(resource)
	if err != nil {
		return nil, err
	}
	u := *c.base
	u.Path = path
	u.RawQuery = query.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Kubernetes request failed: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, &StatusError{Code: resp.StatusCode}
	}
	return resp, nil
}

type listMetadata struct {
	ResourceVersion string `json:"resourceVersion"`
	Continue        string `json:"continue"`
}

func (c *Client) list(ctx context.Context, resource string) ([]json.RawMessage, string, error) {
	var items []json.RawMessage
	version, continuation := "", ""
	seen := map[string]bool{}
	for {
		q := url.Values{"limit": {"500"}}
		if continuation != "" {
			q.Set("continue", continuation)
		}
		resp, err := c.get(ctx, resource, q)
		if err != nil {
			return nil, "", err
		}
		data, err := io.ReadAll(io.LimitReader(resp.Body, 16*1024*1024+1))
		resp.Body.Close()
		if err != nil {
			return nil, "", fmt.Errorf("read Kubernetes list: %w", err)
		}
		if len(data) > 16*1024*1024 {
			return nil, "", errors.New("Kubernetes list page too large")
		}
		var page struct {
			Metadata listMetadata      `json:"metadata"`
			Items    []json.RawMessage `json:"items"`
		}
		if json.Unmarshal(data, &page) != nil || page.Metadata.ResourceVersion == "" || page.Items == nil {
			return nil, "", errors.New("invalid Kubernetes list response")
		}
		if version != "" && version != page.Metadata.ResourceVersion {
			return nil, "", errors.New("inconsistent Kubernetes list version")
		}
		version = page.Metadata.ResourceVersion
		items = append(items, page.Items...)
		if len(items) > 100000 {
			return nil, "", errors.New("namespace inventory limit exceeded")
		}
		continuation = page.Metadata.Continue
		if continuation == "" {
			return items, version, nil
		}
		if seen[continuation] {
			return nil, "", errors.New("repeated Kubernetes continuation token")
		}
		seen[continuation] = true
	}
}

// Watch reports changes; the caller reconciles a full snapshot instead of publishing partial events.
func (c *Client) Watch(ctx context.Context, resource, version string, changed func()) error {
	if version == "" {
		return errors.New("watch requires a list resourceVersion")
	}
	resp, err := c.get(ctx, resource, url.Values{"watch": {"true"}, "resourceVersion": {version}, "allowWatchBookmarks": {"true"}, "timeoutSeconds": {"60"}})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(io.LimitReader(resp.Body, 32*1024*1024))
	for {
		var event struct {
			Type   string `json:"type"`
			Object struct {
				Code int `json:"code"`
			} `json:"object"`
		}
		if err = decoder.Decode(&event); err != nil {
			return err
		}
		switch event.Type {
		case "BOOKMARK":
			continue
		case "ERROR":
			return &StatusError{Code: event.Object.Code}
		case "ADDED", "MODIFIED", "DELETED":
			changed()
		default:
			return errors.New("invalid Kubernetes watch event")
		}
	}
}
