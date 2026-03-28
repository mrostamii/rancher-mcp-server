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

func TestClusterListHandler_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "local") && strings.Contains(r.URL.Path, "clusters") {
			steveResp := rancherclient.SteveCollection{
				Data: []rancherclient.SteveResource{
					{ObjectMeta: rancherclient.ObjectMeta{Name: "cluster-a"}},
					{ObjectMeta: rancherclient.ObjectMeta{Name: "cluster-b"}},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(steveResp)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	client := rancherclient.NewSteveClient(srv.URL, "token", true)
	toolset := NewToolset(client, nil, &security.Policy{})

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "rancher_cluster_list",
			Arguments: map[string]interface{}{
				"format": "json",
			},
		},
	}

	result, err := toolset.clusterListHandler(context.Background(), req)
	if err != nil {
		t.Fatalf("clusterListHandler: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}
	tc, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}
	if !strings.Contains(tc.Text, "cluster-a") || !strings.Contains(tc.Text, "cluster-b") {
		t.Errorf("output should contain cluster names: %s", tc.Text)
	}
}
