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

func TestNormanAuditLogList_405_UnavailableJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/v3/auditlogs") && r.Method == http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			_, _ = w.Write([]byte(`allowed`))
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
			Name:      "rancher_audit_log_list",
			Arguments: map[string]interface{}{"limit": float64(5)},
		},
	}
	res, err := toolset.normanAuditLogListHandler(context.Background(), req)
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
	if out["_source"] != "unavailable" {
		t.Fatalf("_source = %v", out["_source"])
	}
	if int(out["_http_status"].(float64)) != http.StatusMethodNotAllowed {
		t.Fatalf("_http_status = %v", out["_http_status"])
	}
	note, _ := out["_note"].(string)
	if !strings.Contains(strings.ToUpper(note), "GET") {
		t.Fatalf("_note should mention GET: %q", note)
	}
}
