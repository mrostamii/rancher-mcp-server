package security

import "fmt"

// Policy is the central security policy. Every tool checks it before executing.
// Checks happen at registration time (write tools not registered when read-only)
// and at runtime in handlers as defense-in-depth.
type Policy struct {
	ReadOnly           bool
	DisableDestructive bool
	ShowSensitiveData  bool
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
