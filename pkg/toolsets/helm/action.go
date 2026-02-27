package helm

import (
	"log"
	"os"

	"github.com/mrostamii/rancher-mcp-server/pkg/client/helm"
	"helm.sh/helm/v3/pkg/action"
)

// actionConfigFor creates action.Configuration for the given cluster and namespace.
func (t *Toolset) actionConfigFor(clusterID, namespace string) (*action.Configuration, error) {
	getter, err := helm.NewRancherRESTClientGetter(t.baseURL, t.token, t.insecure, clusterID)
	if err != nil {
		return nil, err
	}
	if namespace == "" {
		namespace = "default"
	}
	cfg := new(action.Configuration)
	// Helm stores release data in Secrets by default
	helmDriver := os.Getenv("HELM_DRIVER")
	if helmDriver == "" {
		helmDriver = "secret"
	}
	if err := cfg.Init(getter, namespace, helmDriver, log.Printf); err != nil {
		return nil, err
	}
	return cfg, nil
}
