package k8s

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

type cmdbRoundTripper func(*http.Request) (*http.Response, error)

func (f cmdbRoundTripper) Do(r *http.Request) (*http.Response, error) { return f(r) }

func TestCMDBClientQueriesGUIDAndReturnsDeployPath(t *testing.T) {
	client := cmdbRoundTripper(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/wecmdb/api/v1/ci-data/query/app_instance" || r.Method != http.MethodPost {
			return nil, errors.New("invalid CMDB request")
		}
		body := make([]byte, r.ContentLength)
		r.Body.Read(body)
		if !strings.Contains(string(body), `"name":"guid"`) || !strings.Contains(string(body), `"operator":"equal"`) {
			return nil, errors.New("missing guid equality filter")
		}
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{"code":0,"statusCode":"OK","data":{"pageInfo":{"totalRows":1},"contents":[{"guid":"app_instance-a","deploy_path":"/var/log/app"}]}}`)), Header: make(http.Header)}, nil
	})
	got, err := (CMDBClient{BaseURL: "https://cmdb.invalid", Token: func() string { return "Bearer test" }, HTTP: client}).QueryAppInstance(context.Background(), "app_instance-a")
	if err != nil || got.DeployPath != "/var/log/app" {
		t.Fatalf("unexpected result: %+v %v", got, err)
	}
}

func TestCMDBLogTargetResolverBuildsStableTarget(t *testing.T) {
	client := cmdbRoundTripper(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{"code":0,"statusCode":"OK","data":{"pageInfo":{"totalRows":1},"contents":[{"guid":"app_instance-a","deploy_path":"/var/log/app/"}]}}`)), Header: make(http.Header)}, nil
	})
	container := Container{Key: ObjectKey{ClusterID: "cluster", Namespace: "sit", WorkloadType: "StatefulSet", WorkloadName: "orders", PodName: "orders-0", ContainerName: "app"}, WorkloadUID: "work", PodUID: "pod-a", InstanceID: "app_instance-a"}
	resolver := CMDBLogTargetResolver{Client: CMDBClient{BaseURL: "https://cmdb.invalid", HTTP: client}}
	first, err := resolver.Resolve(context.Background(), container)
	if err != nil || first.Directory != "/var/log/app/logs" || first.Glob != "*.log" || first.InstanceID != container.InstanceID || first.PodName != container.Key.PodName {
		t.Fatalf("unexpected target: %+v %v", first, err)
	}
	container.PodUID = "pod-b"
	second, err := resolver.Resolve(context.Background(), container)
	if err != nil || first.ID != second.ID {
		t.Fatalf("target ID changed across Pod replacement: %s %s %v", first.ID, second.ID, err)
	}
}
