package cmd

import (
	"github.com/mrostamii/rancher-mcp-server/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const envPrefix = "RANCHER_MCP"

// NewRootCommand creates the root cobra command and binds config via viper.
func NewRootCommand() *cobra.Command {
	cfg := config.DefaultConfig()
	root := &cobra.Command{
		Use:   "rancher-mcp-server",
		Short: "MCP server for Rancher ecosystem (Harvester, Fleet, Kubernetes)",
		Long:  "Model Context Protocol server providing tools for multi-cluster Kubernetes, Harvester HCI, and Fleet GitOps via Rancher Steve API.",
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			if p, _ := cmd.Flags().GetString("config"); p != "" {
				viper.SetConfigFile(p)
				if err := viper.ReadInConfig(); err != nil {
					return err
				}
			}
			return viper.Unmarshal(cfg)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServe(cfg)
		},
	}

	flags := root.PersistentFlags()
	flags.StringVar(&cfg.RancherServerURL, "rancher-server-url", cfg.RancherServerURL, "Rancher server URL")
	flags.StringVar(&cfg.RancherToken, "rancher-token", cfg.RancherToken, "Rancher bearer token")
	flags.BoolVar(&cfg.TLSInsecure, "tls-insecure", cfg.TLSInsecure, "Skip TLS verification")
	flags.IntVar(&cfg.Port, "port", cfg.Port, "HTTP port (0 = stdio)")
	flags.IntVar(&cfg.LogLevel, "log-level", cfg.LogLevel, "Log level 0-9")
	flags.StringVar(&cfg.Transport, "transport", cfg.Transport, "Transport: stdio or http")
	flags.BoolVar(&cfg.ReadOnly, "read-only", cfg.ReadOnly, "Disable all write operations")
	flags.BoolVar(&cfg.DisableDestructive, "disable-destructive", cfg.DisableDestructive, "Disable delete operations")
	flags.BoolVar(&cfg.ShowSensitiveData, "show-sensitive-data", cfg.ShowSensitiveData, "Show secret data (default: masked)")
	flags.StringSliceVar(&cfg.Toolsets, "toolsets", cfg.Toolsets, "Toolsets to enable (currently: harvester)")

	root.Flags().String("config", "", "Config file (TOML or YAML)")
	_ = viper.BindPFlag("config", root.Flags().Lookup("config"))

	_ = viper.BindPFlag("rancher_server_url", root.PersistentFlags().Lookup("rancher-server-url"))
	_ = viper.BindPFlag("rancher_token", root.PersistentFlags().Lookup("rancher-token"))
	_ = viper.BindPFlag("tls_insecure", root.PersistentFlags().Lookup("tls-insecure"))
	_ = viper.BindPFlag("port", root.PersistentFlags().Lookup("port"))
	_ = viper.BindPFlag("log_level", root.PersistentFlags().Lookup("log-level"))
	_ = viper.BindPFlag("transport", root.PersistentFlags().Lookup("transport"))
	_ = viper.BindPFlag("read_only", root.PersistentFlags().Lookup("read-only"))
	_ = viper.BindPFlag("disable_destructive", root.PersistentFlags().Lookup("disable-destructive"))
	_ = viper.BindPFlag("show_sensitive_data", root.PersistentFlags().Lookup("show-sensitive-data"))
	_ = viper.BindPFlag("toolsets", root.PersistentFlags().Lookup("toolsets"))

	viper.SetEnvPrefix(envPrefix)
	viper.AutomaticEnv()
	return root
}
