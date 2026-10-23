package k8s

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func testClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	server := httptest.NewTLSServer(handler)
	t.Cleanup(server.Close)
	cert, err := x509.ParseCertificate(server.TLS.Certificates[0].Certificate[0])
	if err != nil {
		t.Fatal(err)
	}
	client, err := NewClient(server.URL, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw}), "test-token", "sit-a")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(client.Close)
	return client
}

func TestSnapshotPaginationAndOwnerIdentity(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Error("missing authentication")
		}
		switch r.URL.Path {
		case "/apis/apps/v1/namespaces/sit-a/statefulsets":
			fmt.Fprint(w, `{"metadata":{"resourceVersion":"sts-rv"},"items":[{"metadata":{"name":"orders","namespace":"sit-a","uid":"sts-new"},"spec":{"replicas":1,"selector":{"matchLabels":{"app":"orders"}}}}]}`)
		case "/api/v1/namespaces/sit-a/pods":
			if r.URL.Query().Get("continue") == "" {
				fmt.Fprint(w, `{"metadata":{"resourceVersion":"pod-rv","continue":"next+page"},"items":[{"metadata":{"name":"orders-0","namespace":"sit-a","uid":"pod-new","labels":{"app":"orders"},"ownerReferences":[{"kind":"StatefulSet","name":"orders","uid":"sts-new","controller":true}]},"spec":{"nodeName":"node-b","containers":[{"name":"app","resources":{"requests":{"cpu":"100m"}}}]},"status":{"phase":"Running","containerStatuses":[{"name":"app","containerID":"runtime://new","ready":true,"restartCount":2}]}}]}`)
			} else {
				if r.URL.Query().Get("continue") != "next+page" {
					t.Error("continuation token was not encoded correctly")
				}
				// A pod still owned by an old same-name StatefulSet must not attach to the new one.
				fmt.Fprint(w, `{"metadata":{"resourceVersion":"pod-rv"},"items":[{"metadata":{"name":"orders-1","namespace":"sit-a","uid":"pod-old","labels":{"app":"orders"},"ownerReferences":[{"kind":"StatefulSet","name":"orders","uid":"sts-old","controller":true}]},"spec":{"containers":[{"name":"app"}]}}]}`)
			}
		default:
			t.Errorf("unexpected access: %s", r.URL.Path)
			http.Error(w, "unexpected", 404)
		}
	})
	snapshot, err := client.Snapshot(context.Background(), "cluster-a")
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.PodVersion != "pod-rv" || snapshot.WorkloadVersion != "sts-rv" {
		t.Fatal("list versions lost")
	}
	if len(snapshot.Containers) != 1 {
		t.Fatalf("unexpected containers: %+v", snapshot.Containers)
	}
	c := snapshot.Containers[0]
	if c.PodUID != "pod-new" || c.Node != "node-b" || c.ContainerID != "runtime://new" || c.RestartCount != 2 {
		t.Fatalf("runtime identity lost: %+v", c)
	}
	key := c.Key
	c.PodUID = "another"
	c.ContainerID = "runtime://replacement"
	if c.Key != key || key.PodName != "orders-0" || key.WorkloadName != "orders" {
		t.Fatal("logical key must use names")
	}
}

func TestSnapshotFailureDoesNotReturnPartialInventory(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "statefulsets") {
			fmt.Fprint(w, `{"metadata":{"resourceVersion":"1"},"items":[]}`)
			return
		}
		http.Error(w, "credential-must-not-appear", 403)
	})
	snapshot, err := client.Snapshot(context.Background(), "cluster-a")
	if err == nil || snapshot != nil {
		t.Fatal("failed scan returned usable inventory")
	}
	if strings.Contains(err.Error(), "credential-must-not-appear") {
		t.Fatal("response body leaked")
	}
	if !IsStatus(err, 403) {
		t.Fatalf("status not preserved: %v", err)
	}
}

func TestClientRejectsUnsafeConfiguration(t *testing.T) {
	for _, endpoint := range []string{"http://api.example.invalid", "https://user:secret@api.example.invalid", "https://api.example.invalid/path", "https://api.example.invalid?token=x"} {
		if _, err := NewClient(endpoint, nil, "test", "sit-a"); err == nil {
			t.Errorf("accepted unsafe endpoint %s", endpoint)
		}
	}
	if _, err := NewClient("https://api.example.invalid", nil, "test", "../other"); err == nil {
		t.Fatal("accepted invalid namespace")
	}
}

func TestWatchUsesListVersionAndReportsExpiredVersion(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("resourceVersion") != "pod-version" || r.URL.Query().Get("watch") != "true" {
			t.Error("watch did not resume list version")
		}
		fmt.Fprintln(w, `{"type":"BOOKMARK","object":{"metadata":{"resourceVersion":"newer"}}}`)
		fmt.Fprintln(w, `{"type":"ERROR","object":{"code":410}}`)
	})
	err := client.Watch(context.Background(), "pods", "pod-version", func() { t.Error("bookmark must not trigger reconciliation") })
	if !IsStatus(err, 410) {
		t.Fatalf("expected expired resourceVersion, got %v", err)
	}
}

func TestWatchTreatsNormalStreamCloseAsResync(t *testing.T) {
	client := testClient(t, func(http.ResponseWriter, *http.Request) {})
	if err := client.Watch(context.Background(), "pods", "pod-version", func() {}); err != nil {
		t.Fatalf("normal watch close should trigger resync without an error, got %v", err)
	}
}

func TestClientRejectsUntrustedCertificateAndRedirect(t *testing.T) {
	var foreignRequests atomic.Int32
	foreign := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		foreignRequests.Add(1)
		http.Error(w, "direct test endpoint", http.StatusNotFound)
	}))
	defer foreign.Close()
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, foreign.URL, http.StatusFound) })
	if _, err := client.Snapshot(context.Background(), "cluster-a"); err == nil {
		t.Fatal("redirect accepted")
	}
	if foreignRequests.Load() != 0 {
		t.Fatal("request followed redirect")
	}
	untrusted, err := NewClient(foreign.URL, nil, "test-token", "sit-a")
	if err != nil {
		t.Fatal(err)
	}
	defer untrusted.Close()
	if _, err = untrusted.Snapshot(context.Background(), "cluster-a"); err == nil || foreignRequests.Load() == 0 {
		t.Fatal("optional CA behavior was not exercised")
	}
}
