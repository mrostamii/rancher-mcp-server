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

func TestNormanSchemaListHandler_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v3/schemas" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"type": "collection",
			"data": []map[string]string{{"id": "user"}},
		})
	}))
	defer srv.Close()

	steve := rancherclient.NewSteveClient(srv.URL, "tok", true)
	norman := rancherclient.NewNormanClient(srv.URL, "tok", true)
	toolset := NewToolset(steve, norman, &security.Policy{})

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "rancher_norman_schema_list",
			Arguments: map[string]interface{}{},
		},
	}
	res, err := toolset.normanSchemaListHandler(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if res.IsError {
		t.Fatalf("error: %v", res.Content)
	}
	tc, ok := res.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("want TextContent")
	}
	if !strings.Contains(tc.Text, "user") {
		t.Fatalf("output: %s", tc.Text)
	}
}
