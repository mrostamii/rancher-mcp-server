package rancher

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const (
	// Steve VM type (KubeVirt)
	TypeVirtualMachines         = "kubevirt.io.virtualmachines"
	TypeVirtualMachineInstances = "kubevirt.io.virtualmachineinstances"
	// Harvester CRDs
	TypeVirtualMachineImages = "harvesterhci.io.virtualmachineimages"
	TypePersistentVolumeClaims = "v1.persistentvolumeclaims"
	// NetworkAttachmentDefinition for Harvester networks
	TypeNetworkAttachmentDefinition = "k8s.cni.cncf.io.networkattachmentdefinitions"
)

// SteveClient talks to Rancher Steve API (K8s proxy).
type SteveClient struct {
	baseURL    string
	token      string
	insecure   bool
	httpClient *http.Client
}

// NewSteveClient creates a Steve API client. baseURL is the Rancher server URL (e.g. https://rancher.example.com).
func NewSteveClient(baseURL, token string, insecure bool) *SteveClient {
	u, _ := url.Parse(baseURL)
	if u.Scheme == "" {
		u.Scheme = "https"
	}
	base := u.String()
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure},
	}
	return &SteveClient{
		baseURL:    base,
		token:      token,
		insecure:   insecure,
		httpClient: &http.Client{Transport: tr},
	}
}

// SteveCollection is the list response from Steve API.
type SteveCollection struct {
	Data   []SteveResource `json:"data"`
	Continue string        `json:"continue,omitempty"`
}

// SteveResource is a single resource (type + metadata + spec/status).
type SteveResource struct {
	TypeMeta   TypeMeta   `json:"typeMeta"`
	ObjectMeta ObjectMeta `json:"metadata"`
	Spec       interface{} `json:"spec,omitempty"`
	Status     interface{} `json:"status,omitempty"`
}

type TypeMeta struct {
	Kind       string `json:"kind"`
	APIVersion string `json:"apiVersion"`
}

type ObjectMeta struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	Annotations map[string]string `json:"annotations,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
}

// ListOpts for list requests.
type ListOpts struct {
	Namespace string
	LabelSelector string
	Limit     int
	Continue  string
}

// List resources in a cluster.
func (c *SteveClient) List(ctx context.Context, clusterID, resourceType string, opts ListOpts) (*SteveCollection, error) {
	path := fmt.Sprintf("/k8s/clusters/%s/v1/%s", clusterID, resourceType)
	if opts.Namespace != "" {
		path = fmt.Sprintf("/k8s/clusters/%s/v1/namespaces/%s/%s", clusterID, opts.Namespace, resourceType)
	}
	u, err := url.Parse(c.baseURL + path)
	if err != nil {
		return nil, fmt.Errorf("steve list url: %w", err)
	}
	q := u.Query()
	if opts.Limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", opts.Limit))
	}
	if opts.Continue != "" {
		q.Set("continue", opts.Continue)
	}
	if opts.LabelSelector != "" {
		q.Set("labelSelector", opts.LabelSelector)
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("steve list request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("steve list %s: %s", resp.Status, string(body))
	}
	var col SteveCollection
	if err := json.NewDecoder(resp.Body).Decode(&col); err != nil {
		return nil, fmt.Errorf("steve list decode: %w", err)
	}
	return &col, nil
}

// Get a single resource.
func (c *SteveClient) Get(ctx context.Context, clusterID, resourceType, namespace, name string) (*SteveResource, error) {
	var path string
	if namespace != "" {
		path = fmt.Sprintf("/k8s/clusters/%s/v1/namespaces/%s/%s/%s", clusterID, namespace, resourceType, name)
	} else {
		path = fmt.Sprintf("/k8s/clusters/%s/v1/%s/%s", clusterID, resourceType, name)
	}
	u := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("steve get request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("steve get %s: %s", resp.Status, string(body))
	}
	var res SteveResource
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("steve get decode: %w", err)
	}
	return &res, nil
}

// Action calls a subresource action (e.g. start, stop on a VM).
func (c *SteveClient) Action(ctx context.Context, clusterID, resourceType, namespace, name, action string, body interface{}) error {
	path := fmt.Sprintf("/k8s/clusters/%s/v1/namespaces/%s/%s/%s?action=%s", clusterID, namespace, resourceType, name, action)
	u := c.baseURL + path
	var buf io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		buf = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, buf)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("steve action request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("steve action %s: %s", resp.Status, string(b))
	}
	return nil
}
