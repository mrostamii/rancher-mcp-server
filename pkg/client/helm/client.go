package helm

import (
	"fmt"
	"net/url"
	"strings"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// RancherRESTClientGetter implements genericclioptions.RESTClientGetter for Rancher's cluster proxy.
// It builds rest.Config for a downstream cluster via Rancher's proxy URL.
type RancherRESTClientGetter struct {
	baseURL  string
	token    string
	insecure bool
	cluster  string
}

// NewRancherRESTClientGetter creates a getter for cluster operations through Rancher proxy.
// baseURL is the Rancher server URL (e.g. https://rancher.example.com).
func NewRancherRESTClientGetter(baseURL, token string, insecure bool, clusterID string) (*RancherRESTClientGetter, error) {
	u, err := url.Parse(strings.TrimSuffix(baseURL, "/"))
	if err != nil {
		return nil, fmt.Errorf("parse base URL: %w", err)
	}
	if u.Scheme == "" {
		u.Scheme = "https"
	}
	return &RancherRESTClientGetter{
		baseURL:  u.String(),
		token:    token,
		insecure: insecure,
		cluster:  clusterID,
	}, nil
}

// ToRESTConfig returns rest.Config for the cluster via Rancher proxy.
func (g *RancherRESTClientGetter) ToRESTConfig() (*rest.Config, error) {
	host := fmt.Sprintf("%s/k8s/clusters/%s", g.baseURL, g.cluster)
	cfg := &rest.Config{
		Host: host,
		// Rancher token works as Bearer for downstream cluster API
		BearerToken:     g.token,
		TLSClientConfig: rest.TLSClientConfig{Insecure: g.insecure},
	}
	cfg.ContentType = "application/json"
	return cfg, nil
}

// ToDiscoveryClient returns a cached discovery client.
func (g *RancherRESTClientGetter) ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	cfg, err := g.ToRESTConfig()
	if err != nil {
		return nil, err
	}
	dc, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return nil, err
	}
	return memory.NewMemCacheClient(dc), nil
}

// ToRESTMapper returns a REST mapper for the cluster.
func (g *RancherRESTClientGetter) ToRESTMapper() (meta.RESTMapper, error) {
	discoveryClient, err := g.ToDiscoveryClient()
	if err != nil {
		return nil, err
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(discoveryClient)
	expander := restmapper.NewShortcutExpander(mapper, discoveryClient, func(string) {})
	return expander, nil
}

// ToRawKubeConfigLoader returns a ClientConfig that produces our rest.Config.
// Helm's action.Configuration uses this for namespace resolution.
func (g *RancherRESTClientGetter) ToRawKubeConfigLoader() clientcmd.ClientConfig {
	return &restConfigLoader{getter: g}
}

// restConfigLoader implements clientcmd.ClientConfig for in-process rest.Config.
type restConfigLoader struct {
	getter *RancherRESTClientGetter
}

func (r *restConfigLoader) RawConfig() (clientcmdapi.Config, error) {
	return clientcmdapi.Config{}, nil
}

func (r *restConfigLoader) ClientConfig() (*rest.Config, error) {
	return r.getter.ToRESTConfig()
}

func (r *restConfigLoader) Namespace() (string, bool, error) {
	return "default", false, nil
}

func (r *restConfigLoader) ConfigAccess() clientcmd.ConfigAccess {
	return clientcmd.NewDefaultClientConfigLoadingRules()
}

var _ genericclioptions.RESTClientGetter = (*RancherRESTClientGetter)(nil)
