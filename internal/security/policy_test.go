package security

import (
	"strings"
	"testing"
)

func TestPolicy_CheckNamespace(t *testing.T) {
	t.Run("empty namespace allowed", func(t *testing.T) {
		p := &Policy{}
		if err := p.CheckNamespace(""); err != nil {
			t.Errorf("empty namespace should be allowed: %v", err)
		}
	})

	t.Run("denied namespace", func(t *testing.T) {
		p := &Policy{DeniedNamespaces: []string{"kube-system", "default"}}
		if err := p.CheckNamespace("kube-system"); err == nil {
			t.Error("kube-system should be denied")
		}
		if err := p.CheckNamespace("default"); err == nil {
			t.Error("default should be denied")
		}
		if err := p.CheckNamespace("other"); err != nil {
			t.Errorf("other should be allowed: %v", err)
		}
	})

	t.Run("denied case insensitive", func(t *testing.T) {
		p := &Policy{DeniedNamespaces: []string{"Kube-System"}}
		if err := p.CheckNamespace("kube-system"); err == nil {
			t.Error("kube-system should be denied (case insensitive)")
		}
	})

	t.Run("allowed namespaces only", func(t *testing.T) {
		p := &Policy{AllowedNamespaces: []string{"default", "app"}}
		if err := p.CheckNamespace("default"); err != nil {
			t.Errorf("default should be allowed: %v", err)
		}
		if err := p.CheckNamespace("app"); err != nil {
			t.Errorf("app should be allowed: %v", err)
		}
		if err := p.CheckNamespace("other"); err == nil {
			t.Error("other should not be allowed")
		}
	})

	t.Run("denied overrides allowed", func(t *testing.T) {
		p := &Policy{
			AllowedNamespaces: []string{"default", "app"},
			DeniedNamespaces:   []string{"default"},
		}
		if err := p.CheckNamespace("default"); err == nil {
			t.Error("default should be denied even when in allowed")
		}
		if err := p.CheckNamespace("app"); err != nil {
			t.Errorf("app should be allowed: %v", err)
		}
	})
}

func TestPolicy_FilterListByNamespace(t *testing.T) {
	items := []map[string]interface{}{
		{"name": "a", "namespace": "ns1"},
		{"name": "b", "namespace": "ns2"},
		{"name": "c", "namespace": "ns1"},
	}

	t.Run("no restrictions", func(t *testing.T) {
		p := &Policy{}
		got := p.FilterListByNamespace(items)
		if len(got) != 3 {
			t.Errorf("expected 3 items, got %d", len(got))
		}
	})

	t.Run("allowed namespaces", func(t *testing.T) {
		p := &Policy{AllowedNamespaces: []string{"ns1"}}
		got := p.FilterListByNamespace(items)
		if len(got) != 2 {
			t.Errorf("expected 2 items (ns1 only), got %d", len(got))
		}
		for _, m := range got {
			if m["namespace"] != "ns1" {
				t.Errorf("unexpected namespace: %v", m["namespace"])
			}
		}
	})

	t.Run("allowed namespaces with cluster-scoped items", func(t *testing.T) {
		itemsWithCluster := []map[string]interface{}{
			{"name": "node-1", "namespace": ""},
			{"name": "pod-a", "namespace": "ns1"},
		}
		p := &Policy{AllowedNamespaces: []string{"ns1"}}
		got := p.FilterListByNamespace(itemsWithCluster)
		if len(got) != 2 {
			t.Errorf("expected 2 items (cluster-scoped + ns1), got %d", len(got))
		}
	})

	t.Run("denied namespaces", func(t *testing.T) {
		p := &Policy{DeniedNamespaces: []string{"ns1"}}
		got := p.FilterListByNamespace(items)
		if len(got) != 1 {
			t.Errorf("expected 1 item (ns2 only), got %d", len(got))
		}
		if got[0]["namespace"] != "ns2" {
			t.Errorf("expected ns2, got %v", got[0]["namespace"])
		}
	})

	t.Run("empty list", func(t *testing.T) {
		p := &Policy{AllowedNamespaces: []string{"ns1"}}
		got := p.FilterListByNamespace([]map[string]interface{}{})
		if len(got) != 0 {
			t.Errorf("expected empty, got %d", len(got))
		}
	})
}

func TestPolicy_CanWrite_CanDelete(t *testing.T) {
	t.Run("default allows write and delete", func(t *testing.T) {
		p := &Policy{}
		if !p.CanWrite() {
			t.Error("expected CanWrite true")
		}
		if !p.CanDelete() {
			t.Error("expected CanDelete true")
		}
	})

	t.Run("read only disables write and delete", func(t *testing.T) {
		p := &Policy{ReadOnly: true}
		if p.CanWrite() {
			t.Error("expected CanWrite false")
		}
		if p.CanDelete() {
			t.Error("expected CanDelete false")
		}
	})

	t.Run("disable destructive keeps write", func(t *testing.T) {
		p := &Policy{DisableDestructive: true}
		if !p.CanWrite() {
			t.Error("expected CanWrite true")
		}
		if p.CanDelete() {
			t.Error("expected CanDelete false")
		}
	})
}

func TestPolicy_CheckWrite_CheckDestructive(t *testing.T) {
	t.Run("CheckWrite read only", func(t *testing.T) {
		p := &Policy{ReadOnly: true}
		err := p.CheckWrite()
		if err == nil {
			t.Error("expected error")
		}
		if !strings.Contains(err.Error(), "read-only") {
			t.Errorf("expected read-only in error: %v", err)
		}
	})

	t.Run("CheckDestructive read only", func(t *testing.T) {
		p := &Policy{ReadOnly: true}
		err := p.CheckDestructive()
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("CheckDestructive disable destructive", func(t *testing.T) {
		p := &Policy{DisableDestructive: true}
		err := p.CheckDestructive()
		if err == nil {
			t.Error("expected error")
		}
		if !strings.Contains(err.Error(), "disable-destructive") {
			t.Errorf("expected disable-destructive in error: %v", err)
		}
	})
}
