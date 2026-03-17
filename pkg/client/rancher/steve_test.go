package rancher

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSteveType(t *testing.T) {
	tests := []struct {
		apiVersion string
		kind      string
		want      string
	}{
		{"v1", "Pod", "core.v1.pods"},
		{"v1", "Node", "core.v1.nodes"},
		{"", "Pod", "core.v1.pods"},
		{"apps/v1", "Deployment", "apps.v1.deployments"},
		{"batch/v1", "Job", "batch.v1.jobs"},
		{"v1", "Endpoints", "core.v1.endpoints"},
		{"v1", "Event", "core.v1.events"},
		{"v1", "Ingress", "core.v1.ingresses"},
		{"networking.k8s.io/v1", "Ingress", "networking.k8s.io.v1.ingresses"},
	}
	for _, tt := range tests {
		got := SteveType(tt.apiVersion, tt.kind)
		if got != tt.want {
			t.Errorf("SteveType(%q, %q) = %q, want %q", tt.apiVersion, tt.kind, got, tt.want)
		}
	}
}

func TestNewSteveClient(t *testing.T) {
	client := NewSteveClient("rancher.example.com", "token", true)
	if client == nil {
		t.Fatal("NewSteveClient returned nil")
	}
	if client.token != "token" {
		t.Errorf("token = %q, want token", client.token)
	}
	if !client.insecure {
		t.Error("insecure = false, want true")
	}
	// URL without scheme gets https
	if client.baseURL != "https://rancher.example.com" {
		t.Errorf("baseURL = %q, want https://rancher.example.com", client.baseURL)
	}

	client2 := NewSteveClient("https://other.example.com", "t2", false)
	if client2.baseURL != "https://other.example.com" {
		t.Errorf("baseURL = %q, want https://other.example.com", client2.baseURL)
	}
}

func TestSteveClient_List(t *testing.T) {
	// Use apps.v1.deployments so we hit Steve path (core types try native K8s first)
	steveListResp := SteveCollection{
		Data: []SteveResource{
			{
				TypeMeta:   TypeMeta{Kind: "Deployment", APIVersion: "apps/v1"},
				ObjectMeta: ObjectMeta{Name: "dep-1", Namespace: "default"},
			},
			{
				TypeMeta:   TypeMeta{Kind: "Deployment", APIVersion: "apps/v1"},
				ObjectMeta: ObjectMeta{Name: "dep-2", Namespace: "default"},
			},
		},
		Continue: "next-token",
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Error("missing or wrong Authorization header")
		}
		expected := "/k8s/clusters/c-xxx/v1/namespaces/default/apps.v1.deployments"
		if r.URL.Path != expected {
			t.Errorf("unexpected path: %s, want %s", r.URL.Path, expected)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(steveListResp)
	}))
	defer srv.Close()

	client := NewSteveClient(srv.URL, "test-token", true)
	ctx := context.Background()

	col, err := client.List(ctx, "c-xxx", "apps.v1.deployments", ListOpts{
		Namespace: "default",
		Limit:     10,
	})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(col.Data) != 2 {
		t.Errorf("expected 2 items, got %d", len(col.Data))
	}
	if col.Data[0].ObjectMeta.Name != "dep-1" {
		t.Errorf("first item name = %q, want dep-1", col.Data[0].ObjectMeta.Name)
	}
	if col.Continue != "next-token" {
		t.Errorf("Continue = %q, want next-token", col.Continue)
	}
}

func TestSteveClient_List_404_FallbackToNativeK8s(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		// Steve path (/v1/...) returns 404 first; native path (/apis/...) returns 200
		if strings.Contains(r.URL.Path, "/v1/") && !strings.Contains(r.URL.Path, "/apis/") {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"message":"not found"}`))
			return
		}
		// Native K8s path returns 200
		k8sResp := map[string]interface{}{
			"items": []map[string]interface{}{
				{
					"apiVersion": "apps/v1",
					"kind":       "Deployment",
					"metadata":  map[string]interface{}{"name": "p1", "namespace": "default"},
				},
			},
			"metadata": map[string]interface{}{"continue": ""},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(k8sResp)
	}))
	defer srv.Close()

	client := NewSteveClient(srv.URL, "token", true)
	ctx := context.Background()

	// apps.v1.deployments tries Steve first, then native on 404
	col, err := client.List(ctx, "c-xxx", "apps.v1.deployments", ListOpts{Namespace: "default"})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(col.Data) != 1 {
		t.Errorf("expected 1 item after fallback, got %d", len(col.Data))
	}
	if col.Data[0].ObjectMeta.Name != "p1" {
		t.Errorf("name = %q, want p1", col.Data[0].ObjectMeta.Name)
	}
	if col.Data[0].TypeMeta.Kind != "Deployment" {
		t.Errorf("kind = %q, want Deployment", col.Data[0].TypeMeta.Kind)
	}
	if callCount < 2 {
		t.Errorf("expected at least 2 calls (Steve 404 + native), got %d", callCount)
	}
}

func TestSteveClient_Get(t *testing.T) {
	res := SteveResource{
		TypeMeta:   TypeMeta{Kind: "Deployment", APIVersion: "apps/v1"},
		ObjectMeta: ObjectMeta{Name: "my-dep", Namespace: "default"},
		Spec:       map[string]interface{}{"replicas": float64(1)},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expected := "/k8s/clusters/c-xxx/v1/namespaces/default/apps.v1.deployments/my-dep"
		if r.URL.Path != expected {
			t.Errorf("unexpected path: %s, want %s", r.URL.Path, expected)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(res)
	}))
	defer srv.Close()

	client := NewSteveClient(srv.URL, "token", true)
	ctx := context.Background()

	got, err := client.Get(ctx, "c-xxx", "apps.v1.deployments", "default", "my-dep")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.ObjectMeta.Name != "my-dep" {
		t.Errorf("name = %q, want my-dep", got.ObjectMeta.Name)
	}
}

func TestSteveClient_GetPodLogs(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expected := "/k8s/clusters/c-xxx/api/v1/namespaces/default/pods/my-pod/log"
		if r.URL.Path != expected {
			t.Errorf("path = %s, want %s", r.URL.Path, expected)
		}
		if r.URL.Query().Get("tailLines") != "100" {
			t.Errorf("tailLines = %s", r.URL.Query().Get("tailLines"))
		}
		w.Write([]byte("log line 1\nlog line 2\n"))
	}))
	defer srv.Close()

	client := NewSteveClient(srv.URL, "token", true)
	ctx := context.Background()

	logs, err := client.GetPodLogs(ctx, "c-xxx", "default", "my-pod", "", 100, 0)
	if err != nil {
		t.Fatalf("GetPodLogs: %v", err)
	}
	if !strings.Contains(logs, "log line 1") {
		t.Errorf("logs = %q", logs)
	}
}

func TestSteveClient_Delete(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %s, want DELETE", r.Method)
		}
		if r.URL.Path != "/k8s/clusters/c-xxx/v1/namespaces/default/core.v1.pods/foo" {
			t.Errorf("path = %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := NewSteveClient(srv.URL, "token", true)
	ctx := context.Background()

	err := client.Delete(ctx, "c-xxx", "core.v1.pods", "default", "foo")
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
}
