package kubernetes

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

func newTestToolset(handler http.Handler) (*Toolset, *httptest.Server) {
	srv := httptest.NewServer(handler)
	client := rancher.NewSteveClient(srv.URL, "test-token", true)
	policy := &security.Policy{}
	toolset := NewToolset(client, policy)
	return toolset, srv
}

func callToolRequest(args map[string]interface{}) mcp.CallToolRequest {
	return mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "kubernetes_list",
			Arguments: args,
		},
	}
}

func TestListHandler_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// kubernetes_list uses SteveType(v1, Pod) = core.v1.pods; core types try native K8s first
		// Native path: /k8s/clusters/xxx/api/v1/namespaces/default/pods
		if strings.Contains(r.URL.Path, "/api/v1/") {
			k8sResp := map[string]interface{}{
				"items": []map[string]interface{}{
					{"apiVersion": "v1", "kind": "Pod", "metadata": map[string]interface{}{"name": "pod-a", "namespace": "default"}},
					{"apiVersion": "v1", "kind": "Pod", "metadata": map[string]interface{}{"name": "pod-b", "namespace": "default"}},
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

	req := callToolRequest(map[string]interface{}{
		"cluster":     "c-xxx",
		"api_version": "v1",
		"kind":        "Pod",
		"namespace":   "default",
		"format":      "json",
	})

	result, err := toolset.listHandler(context.Background(), req)
	if err != nil {
		t.Fatalf("listHandler: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}
	if len(result.Content) == 0 {
		t.Fatal("expected content")
	}
	tc, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}
	if !strings.Contains(tc.Text, "pod-a") || !strings.Contains(tc.Text, "pod-b") {
		t.Errorf("output should contain pod names: %s", tc.Text)
	}
}

func TestListHandler_MissingRequiredParam(t *testing.T) {
	toolset, srv := newTestToolset(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not call API when params missing")
	}))
	defer srv.Close()

	req := callToolRequest(map[string]interface{}{
		"cluster": "c-xxx",
		"kind":    "Pod",
		// missing api_version
	})

	result, err := toolset.listHandler(context.Background(), req)
	if err != nil {
		t.Fatalf("listHandler: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for missing required param")
	}
}

func TestListHandler_DeniedNamespace(t *testing.T) {
	policy := &security.Policy{DeniedNamespaces: []string{"kube-system"}}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not call API when namespace denied")
	}))
	defer srv.Close()

	client := rancher.NewSteveClient(srv.URL, "token", true)
	toolset := NewToolset(client, policy)

	req := callToolRequest(map[string]interface{}{
		"cluster":     "c-xxx",
		"api_version": "v1",
		"kind":        "Pod",
		"namespace":   "kube-system",
		"format":      "json",
	})

	result, err := toolset.listHandler(context.Background(), req)
	if err != nil {
		t.Fatalf("listHandler: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for denied namespace")
	}
}

func TestGetHandler_Success(t *testing.T) {
	// Use apps/v1 Deployment so we hit Steve path (simpler to mock)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "apps.v1.deployments") {
			res := rancher.SteveResource{
				TypeMeta:   rancher.TypeMeta{Kind: "Deployment", APIVersion: "apps/v1"},
				ObjectMeta: rancher.ObjectMeta{Name: "my-dep", Namespace: "default"},
				Spec:       map[string]interface{}{"replicas": float64(2)},
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
			Name: "kubernetes_get",
			Arguments: map[string]interface{}{
				"cluster":     "c-xxx",
				"api_version": "apps/v1",
				"kind":        "Deployment",
				"namespace":   "default",
				"name":        "my-dep",
				"format":      "json",
			},
		},
	}

	result, err := toolset.getHandler(context.Background(), req)
	if err != nil {
		t.Fatalf("getHandler: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}
	tc, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}
	if !strings.Contains(tc.Text, "my-dep") {
		t.Errorf("output should contain resource name: %s", tc.Text)
	}
}
