package security

import (
	"fmt"
	"slices"
	"strings"
)

// Policy is the central security policy. Every tool checks it before executing.
// Checks happen at registration time (write tools not registered when read-only)
// and at runtime in handlers as defense-in-depth.
type Policy struct {
	ReadOnly           bool
	DisableDestructive bool
	ShowSensitiveData  bool
	AllowedNamespaces  []string // Empty = all allowed (except denied)
	DeniedNamespaces   []string // Always blocked
}

// CanWrite returns true if write operations (create, update, action) are allowed.
func (p *Policy) CanWrite() bool {
	return !p.ReadOnly
}

// CanDelete returns true if delete operations are allowed.
func (p *Policy) CanDelete() bool {
	return !p.ReadOnly && !p.DisableDestructive
}

// CheckWrite returns an error if write operations are not allowed.
func (p *Policy) CheckWrite() error {
	if p.ReadOnly {
		return fmt.Errorf("write operations are disabled (read-only mode)")
	}
	return nil
}

// CheckDestructive returns an error if delete/destructive operations are not allowed.
func (p *Policy) CheckDestructive() error {
	if p.ReadOnly {
		return fmt.Errorf("destructive operations are disabled (read-only mode)")
	}
	if p.DisableDestructive {
		return fmt.Errorf("destructive operations are disabled (disable-destructive)")
	}
	return nil
}

// CanShowSecret returns true if sensitive data (e.g. Secret data) may be shown.
func (p *Policy) CanShowSecret() bool {
	return p.ShowSensitiveData
}

// CheckNamespace returns an error if the namespace is not allowed by policy.
// Empty namespace is allowed (cluster-scoped resources or list-all).
// When AllowedNamespaces is non-empty, only those namespaces are allowed.
// DeniedNamespaces always blocks, regardless of AllowedNamespaces.
func (p *Policy) CheckNamespace(namespace string) error {
	if namespace == "" {
		return nil
	}
	// Denied always wins
	for _, denied := range p.DeniedNamespaces {
		if strings.EqualFold(namespace, denied) {
			return fmt.Errorf("namespace %q is denied by security policy", namespace)
		}
	}
	// If allowed list is set, namespace must be in it
	if len(p.AllowedNamespaces) > 0 {
		if !slices.ContainsFunc(p.AllowedNamespaces, func(a string) bool { return strings.EqualFold(a, namespace) }) {
			return fmt.Errorf("namespace %q is not in allowed namespaces", namespace)
		}
	}
	return nil
}

// FilterListByNamespace filters items to only include those in allowed namespaces.
// For cluster-scoped resources (empty namespace), items are included.
// When AllowedNamespaces is non-empty, only items in those namespaces pass.
// DeniedNamespaces always exclude.
func (p *Policy) FilterListByNamespace(items []map[string]interface{}) []map[string]interface{} {
	if len(items) == 0 {
		return items
	}
	// No restrictions
	if len(p.AllowedNamespaces) == 0 && len(p.DeniedNamespaces) == 0 {
		return items
	}
	filtered := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		ns, _ := item["namespace"].(string)
		if p.CheckNamespace(ns) == nil {
			filtered = append(filtered, item)
		}
	}
	return filtered
}
