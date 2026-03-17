package harvester

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

func TestVMListHandler_Success(t *testing.T) {
	// TypeVirtualMachines uses Steve path first
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "virtualmachines") {
			steveResp := rancher.SteveCollection{
				Data: []rancher.SteveResource{
					{ObjectMeta: rancher.ObjectMeta{Name: "vm-1", Namespace: "default"}},
					{ObjectMeta: rancher.ObjectMeta{Name: "vm-2", Namespace: "default"}},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(steveResp)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	client := rancher.NewSteveClient(srv.URL, "token", true)
	toolset := NewToolset(client, &security.Policy{})

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "harvester_vm_list",
			Arguments: map[string]interface{}{
				"cluster":   "c-xxx",
				"namespace": "default",
				"format":    "json",
			},
		},
	}

	result, err := toolset.vmListHandler(context.Background(), req)
	if err != nil {
		t.Fatalf("vmListHandler: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}
	tc, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}
	if !strings.Contains(tc.Text, "vm-1") || !strings.Contains(tc.Text, "vm-2") {
		t.Errorf("output should contain VM names: %s", tc.Text)
	}
}

func TestVMGetHandler_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "virtualmachines") && strings.Contains(r.URL.Path, "my-vm") {
			res := rancher.SteveResource{
				TypeMeta:   rancher.TypeMeta{Kind: "VirtualMachine", APIVersion: "kubevirt.io/v1"},
				ObjectMeta: rancher.ObjectMeta{Name: "my-vm", Namespace: "default"},
				Status:     map[string]interface{}{"ready": true},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(res)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	client := rancher.NewSteveClient(srv.URL, "token", true)
	toolset := NewToolset(client, &security.Policy{})

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "harvester_vm_get",
			Arguments: map[string]interface{}{
				"cluster":   "c-xxx",
				"namespace": "default",
				"name":      "my-vm",
				"format":    "json",
			},
		},
	}

	result, err := toolset.vmGetHandler(context.Background(), req)
	if err != nil {
		t.Fatalf("vmGetHandler: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}
	tc, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}
	if !strings.Contains(tc.Text, "my-vm") {
		t.Errorf("output should contain VM name: %s", tc.Text)
	}
}
