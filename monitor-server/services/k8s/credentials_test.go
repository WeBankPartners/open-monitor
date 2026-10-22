package k8s

import (
	"encoding/json"
	"github.com/WeBankPartners/go-common-lib/cipher"
	"strings"
	"testing"
)

func TestCredentialsReuseWeCubeCipherAndNeverSerialize(t *testing.T) {
	const seed = "test-encryption-seed"
	encrypted, err := cipher.AesEnPasswordByGuid("object-a", seed, "test-token", "")
	if err != nil {
		t.Fatal(err)
	}
	credentials, err := PrepareCredentials("object-a", seed, "https://api.example.invalid", "test-ca", encrypted)
	if err != nil {
		t.Fatal(err)
	}
	plain, err := credentials.PlainToken(seed)
	if err != nil || plain != "test-token" {
		t.Fatal("platform encrypted credential did not round trip")
	}
	raw, err := json.Marshal(credentials)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "test-token") || strings.Contains(string(raw), encrypted) {
		t.Fatal("credential leaked into response")
	}
	if _, err = PrepareCredentials("", seed, "https://api.example.invalid", "test-ca", "test-token"); err == nil {
		t.Fatal("missing guid accepted")
	}
}

func TestResolveUsesClusterEndpointAndFallsBackToClusterToken(t *testing.T) {
	cluster := Credentials{Guid: "cluster-guid", APIServer: "https://cluster.example", CA: "cluster-ca", EncryptedToken: "cluster-token"}
	resolved := Resolve(Credentials{}, cluster)
	if resolved.Guid != "cluster-guid" || resolved.APIServer != cluster.APIServer || resolved.CA != cluster.CA || resolved.EncryptedToken != cluster.EncryptedToken {
		t.Fatalf("cluster credentials were not inherited: %+v", resolved)
	}
	override := Resolve(Credentials{Guid: "namespace-guid", EncryptedToken: "namespace-token", APIServer: "ignored", CA: "ignored"}, cluster)
	if override.Guid != "namespace-guid" || override.EncryptedToken != "namespace-token" || override.APIServer != cluster.APIServer || override.CA != cluster.CA {
		t.Fatalf("namespace token override was not applied: %+v", override)
	}
}

func TestPrepareOptionalTokenAllowsClusterFallback(t *testing.T) {
	credentials, err := PrepareOptionalToken("namespace-guid", "seed", "")
	if err != nil || credentials.EncryptedToken != "" {
		t.Fatalf("empty namespace token should remain empty: %+v, %v", credentials, err)
	}
}
