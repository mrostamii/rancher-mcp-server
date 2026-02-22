package config

// Config holds all server configuration (flags, env, file).
// Env vars use prefix RANCHER_MCP_ (e.g. RANCHER_MCP_RANCHER_SERVER_URL).
type Config struct {
	// Rancher connection
	RancherServerURL string `mapstructure:"rancher_server_url"`
	RancherToken     string `mapstructure:"rancher_token"`
	TLSInsecure      bool   `mapstructure:"tls_insecure"`

	// Optional: default Harvester cluster ID (e.g. c-tx8rn). Used when a tool call omits cluster.
	DefaultHarvesterCluster string `mapstructure:"default_harvester_cluster"`

	// Server
	Port      int    `mapstructure:"port"`
	LogLevel  int    `mapstructure:"log_level"`
	Transport string `mapstructure:"transport"` // stdio | http

	// Security
	ReadOnly           bool     `mapstructure:"read_only"`
	DisableDestructive bool     `mapstructure:"disable_destructive"`
	ShowSensitiveData  bool     `mapstructure:"show_sensitive_data"`
	AllowedNamespaces  []string `mapstructure:"allowed_namespaces"`
	DeniedNamespaces   []string `mapstructure:"denied_namespaces"`

	// Toolsets (enabled set names)
	Toolsets []string `mapstructure:"toolsets"`
}

// DefaultConfig returns defaults for running as stdio MCP server.
func DefaultConfig() *Config {
	return &Config{
		Port:               0,
		LogLevel:           2,
		Transport:           "stdio",
		ReadOnly:           true,
		DisableDestructive: false,
		ShowSensitiveData:  false,
		DeniedNamespaces:   []string{"kube-system", "cattle-system"},
		Toolsets:           []string{"harvester"},
	}
}
