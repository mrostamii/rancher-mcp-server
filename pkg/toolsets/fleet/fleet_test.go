package fleet

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mrostamii/rancher-mcp-server/internal/security"
	"github.com/mrostamii/rancher-mcp-server/pkg/client/rancher"
)

func TestGitrepoListHandler_Success(t *testing.T) {
	// TypeFleetGitRepos uses native API first
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "gitrepos") {
			k8sResp := map[string]interface{}{
				"items": []map[string]interface{}{
					{"apiVersion": "fleet.cattle.io/v1alpha1", "kind": "GitRepo", "metadata": map[string]interface{}{"name": "repo-1", "namespace": "fleet-default"}},
					{"apiVersion": "fleet.cattle.io/v1alpha1", "kind": "GitRepo", "metadata": map[string]interface{}{"name": "repo-2", "namespace": "fleet-default"}},
				},
				"metadata": map[string]interface{}{"continue": ""},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(k8sResp)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	client := rancher.NewSteveClient(srv.URL, "token", true)
	toolset := NewToolset(client, &security.Policy{})

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "fleet_gitrepo_list",
			Arguments: map[string]interface{}{
				"namespace": "fleet-default",
				"format":    "json",
			},
		},
	}

	result, err := toolset.gitrepoListHandler(context.Background(), req)
	if err != nil {
		t.Fatalf("gitrepoListHandler: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}
	tc, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}
	if !strings.Contains(tc.Text, "repo-1") || !strings.Contains(tc.Text, "repo-2") {
		t.Errorf("output should contain repo names: %s", tc.Text)
	}
}
