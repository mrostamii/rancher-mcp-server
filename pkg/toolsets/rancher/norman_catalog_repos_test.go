package rancher

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mrostamii/rancher-mcp-server/internal/security"
	rancherclient "github.com/mrostamii/rancher-mcp-server/pkg/client/rancher"
)

func TestNormanCatalogList_UnavailableJSON_NotHardError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/v3/catalogs") {
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"type":"error"}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	steve := rancherclient.NewSteveClient(srv.URL, "tok", true)
	norman := rancherclient.NewNormanClient(srv.URL, "tok", true)
	toolset := NewToolset(steve, norman, &security.Policy{})

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "rancher_catalog_list",
			Arguments: map[string]interface{}{"limit": float64(20)},
		},
	}
	res, err := toolset.normanCatalogListHandler(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if res.IsError {
		t.Fatalf("expected soft JSON, hard error: %v", res.Content)
	}
	tc := res.Content[0].(mcp.TextContent)
	var out map[string]interface{}
	if err := json.Unmarshal([]byte(tc.Text), &out); err != nil {
		t.Fatalf("parse JSON: %v\n%s", err, tc.Text)
	}
	if out["_source"] != "unavailable" {
		t.Fatalf("_source = %v", out["_source"])
	}
	if int(out["_http_status"].(float64)) != http.StatusForbidden {
		t.Fatalf("_http_status = %v", out["_http_status"])
	}
}

func TestNormanClusterRepoList_KubernetesFallback_NotHardError(t *testing.T) {
	steveResp := rancherclient.SteveCollection{
		Data: []rancherclient.SteveResource{
			{
				TypeMeta:   rancherclient.TypeMeta{Kind: "ClusterRepo", APIVersion: "catalog.cattle.io/v1"},
				ObjectMeta: rancherclient.ObjectMeta{Name: "rancher-charts", Namespace: clusterRepoSteveNamespace},
			},
		},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/v3/clusterrepos") {
			http.NotFound(w, r)
			return
		}
		want := "/k8s/clusters/local/v1/namespaces/" + clusterRepoSteveNamespace + "/" + rancherclient.TypeCatalogClusterRepos
		if r.URL.Path == want {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(steveResp)
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	steve := rancherclient.NewSteveClient(srv.URL, "tok", true)
	norman := rancherclient.NewNormanClient(srv.URL, "tok", true)
	toolset := NewToolset(steve, norman, &security.Policy{})

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "rancher_cluster_repo_list",
			Arguments: map[string]interface{}{"limit": float64(20)},
		},
	}
	res, err := toolset.normanClusterRepoListHandler(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if res.IsError {
		t.Fatalf("expected soft JSON: %v", res.Content)
	}
	tc := res.Content[0].(mcp.TextContent)
	var out map[string]interface{}
	if err := json.Unmarshal([]byte(tc.Text), &out); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if out["_source"] != "kubernetes_api_fallback" {
		t.Fatalf("_source = %v", out["_source"])
	}
	data, ok := out["data"].([]interface{})
	if !ok || len(data) != 1 {
		t.Fatalf("data = %v", out["data"])
	}
}

func TestNormanClusterRepoList_UnavailableWithAttemptErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/v3/clusterrepos") {
			http.NotFound(w, r)
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	steve := rancherclient.NewSteveClient(srv.URL, "tok", true)
	norman := rancherclient.NewNormanClient(srv.URL, "tok", true)
	toolset := NewToolset(steve, norman, &security.Policy{})

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "rancher_cluster_repo_list",
			Arguments: map[string]interface{}{"limit": float64(20)},
		},
	}
	res, err := toolset.normanClusterRepoListHandler(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if res.IsError {
		t.Fatal("expected unavailable JSON")
	}
	tc := res.Content[0].(mcp.TextContent)
	var out map[string]interface{}
	if err := json.Unmarshal([]byte(tc.Text), &out); err != nil {
		t.Fatal(err)
	}
	if out["_source"] != "unavailable" {
		t.Fatalf("_source = %v", out["_source"])
	}
	ae, ok := out["attempt_errors"].([]interface{})
	if !ok || len(ae) < 1 {
		t.Fatalf("attempt_errors = %v", out["attempt_errors"])
	}
}

func TestClusterRepoNamespaceAndName(t *testing.T) {
	ns, name := clusterRepoNamespaceAndName("foo")
	if ns != clusterRepoSteveNamespace || name != "foo" {
		t.Fatalf("default ns: got %q %q", ns, name)
	}
	ns, name = clusterRepoNamespaceAndName("cattle-global-data:release-name")
	if ns != "cattle-global-data" || name != "release-name" {
		t.Fatalf("colon: got %q %q", ns, name)
	}
	ns, name = clusterRepoNamespaceAndName("cattle-global-data/release-name")
	if ns != "cattle-global-data" || name != "release-name" {
		t.Fatalf("slash: got %q %q", ns, name)
	}
}
