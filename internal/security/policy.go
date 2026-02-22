package security

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

// CanShowSecret returns true if sensitive data (e.g. Secret data) may be shown.
func (p *Policy) CanShowSecret() bool {
	return p.ShowSensitiveData
}
