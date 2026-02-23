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
	"strings"
)

const (
	// Steve VM type (KubeVirt)
	TypeVirtualMachines         = "kubevirt.io.virtualmachines"
	TypeVirtualMachineInstances = "kubevirt.io.virtualmachineinstances"
	// Harvester CRDs
	TypeVirtualMachineImages   = "harvesterhci.io.virtualmachineimages"
	TypePersistentVolumeClaims = "v1.persistentvolumeclaims"
	// NetworkAttachmentDefinition for Harvester networks
	TypeNetworkAttachmentDefinition = "k8s.cni.cncf.io.networkattachmentdefinitions"
	// Rancher management (use clusterID = "local")
	TypeManagementClusters = "management.cattle.io.clusters"
	TypeManagementProjects = "management.cattle.io.projects"
	// Core K8s (Rancher Steve uses "core" as the API group name for core/v1 resources)
	TypeEvents = "core.v1.events"
	TypeNodes  = "core.v1.nodes"
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
	Data     []SteveResource `json:"data"`
	Continue string          `json:"continue,omitempty"`
}

// SteveResource is a single resource (type + metadata + spec/status).
type SteveResource struct {
	TypeMeta   TypeMeta    `json:"typeMeta"`
	ObjectMeta ObjectMeta  `json:"metadata"`
	Spec       interface{} `json:"spec,omitempty"`
	Status     interface{} `json:"status,omitempty"`
}

type TypeMeta struct {
	Kind       string `json:"kind"`
	APIVersion string `json:"apiVersion"`
}

type ObjectMeta struct {
	Name            string            `json:"name"`
	Namespace       string            `json:"namespace"`
	ResourceVersion string            `json:"resourceVersion,omitempty"`
	Annotations     map[string]string `json:"annotations,omitempty"`
	Labels          map[string]string `json:"labels,omitempty"`
}

// ListOpts for list requests.
type ListOpts struct {
	Namespace     string
	LabelSelector string
	FieldSelector string
	Limit         int
	Continue      string
}

// steveTypeFallback returns the alternate Steve type for 404 fallback (some Rancher versions use v1.X instead of core.v1.X for core resources).
func steveTypeFallback(resourceType string) string {
	if strings.HasPrefix(resourceType, "core.v1.") {
		return "v1." + strings.TrimPrefix(resourceType, "core.v1.")
	}
	return ""
}

// steveTypeToK8sCore returns the native K8s resource name for core v1 types (e.g. "pods", "nodes"), or "" if not a supported core type.
// Used when Steve /v1/ path returns 404; Rancher may expose the raw K8s API at /k8s/clusters/<id>/api/v1/<resource>.
func steveTypeToK8sCore(resourceType string) string {
	// Normalize to "v1.X" for lookup
	t := resourceType
	if strings.HasPrefix(t, "core.v1.") {
		t = "v1." + strings.TrimPrefix(t, "core.v1.")
	}
	switch t {
	case "v1.pods", "v1.nodes", "v1.events", "v1.namespaces", "v1.services", "v1.configmaps", "v1.secrets", "v1.persistentvolumeclaims", "v1.persistentvolumes":
		return strings.TrimPrefix(t, "v1.")
	}
	return ""
}

// k8sAPIPath holds native Kubernetes API path components (group, version, resource).
// Core group is "" and uses /api/v1/; other groups use /apis/<group>/<version>/.
type k8sAPIPath struct {
	group    string
	version  string
	resource string
}

// steveTypeToK8sAPIPath parses a Steve resource type (e.g. "apps.v1.deployments", "core.v1.pods")
// into native K8s path components. Returns nil if the type cannot be parsed.
func steveTypeToK8sAPIPath(resourceType string) *k8sAPIPath {
	parts := strings.Split(resourceType, ".")
	if len(parts) < 2 {
		return nil
	}
	resource := parts[len(parts)-1]
	version := parts[len(parts)-2]
	group := strings.Join(parts[:len(parts)-2], ".")
	if group == "core" {
		group = ""
	}
	return &k8sAPIPath{group: group, version: version, resource: resource}
}

// basePath returns the path prefix for this API: /api/v1 or /apis/<group>/<version>.
func (p *k8sAPIPath) basePath() string {
	if p.group == "" {
		return "/api/v1"
	}
	return fmt.Sprintf("/apis/%s/%s", p.group, p.version)
}

// k8sListResponse is the native Kubernetes list response (e.g. PodList).
type k8sListResponse struct {
	Items []json.RawMessage `json:"items"`
}

// List resources in a cluster. For core v1 types we try the native K8s API first, then Steve on 404.
// For other types (e.g. apps/v1 Deployment) we try Steve first, then native K8s API on 404.
func (c *SteveClient) List(ctx context.Context, clusterID, resourceType string, opts ListOpts) (*SteveCollection, error) {
	if k8sRes := steveTypeToK8sCore(resourceType); k8sRes != "" {
		col, err := c.listK8sNative(ctx, clusterID, k8sRes, opts)
		if err == nil {
			return col, nil
		}
		if !strings.Contains(err.Error(), "404") && !strings.Contains(err.Error(), "403") {
			return nil, err
		}
	}
	col, err := c.list(ctx, clusterID, resourceType, opts)
	if err != nil {
		if !strings.Contains(err.Error(), "404") {
			return nil, err
		}
		if path := steveTypeToK8sAPIPath(resourceType); path != nil {
			return c.listK8sNativeByPath(ctx, clusterID, path, opts)
		}
		if alt := steveTypeFallback(resourceType); alt != "" {
			return c.list(ctx, clusterID, alt, opts)
		}
		return nil, err
	}
	return col, nil
}

func (c *SteveClient) list(ctx context.Context, clusterID, resourceType string, opts ListOpts) (*SteveCollection, error) {
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
	if opts.FieldSelector != "" {
		q.Set("fieldSelector", opts.FieldSelector)
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

// listK8sNative lists using the native Kubernetes API path (/api/v1/<resource>) and converts the response to SteveCollection.
func (c *SteveClient) listK8sNative(ctx context.Context, clusterID, k8sResource string, opts ListOpts) (*SteveCollection, error) {
	var path string
	if opts.Namespace != "" {
		path = fmt.Sprintf("/k8s/clusters/%s/api/v1/namespaces/%s/%s", clusterID, opts.Namespace, k8sResource)
	} else {
		path = fmt.Sprintf("/k8s/clusters/%s/api/v1/%s", clusterID, k8sResource)
	}
	u, err := url.Parse(c.baseURL + path)
	if err != nil {
		return nil, fmt.Errorf("k8s list url: %w", err)
	}
	q := u.Query()
	if opts.Limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", opts.Limit))
	}
	if opts.LabelSelector != "" {
		q.Set("labelSelector", opts.LabelSelector)
	}
	if opts.FieldSelector != "" {
		q.Set("fieldSelector", opts.FieldSelector)
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
		return nil, fmt.Errorf("k8s list request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("k8s list %s: %s", resp.Status, string(body))
	}

	var list k8sListResponse
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return nil, fmt.Errorf("k8s list decode: %w", err)
	}
	// Convert each K8s item to SteveResource (metadata, spec, status; typeMeta from item)
	data := make([]SteveResource, 0, len(list.Items))
	for _, raw := range list.Items {
		var item struct {
			APIVersion string      `json:"apiVersion"`
			Kind       string      `json:"kind"`
			Metadata   ObjectMeta  `json:"metadata"`
			Spec       interface{} `json:"spec,omitempty"`
			Status     interface{} `json:"status,omitempty"`
		}
		if err := json.Unmarshal(raw, &item); err != nil {
			continue
		}
		data = append(data, SteveResource{
			TypeMeta:   TypeMeta{Kind: item.Kind, APIVersion: item.APIVersion},
			ObjectMeta: item.Metadata,
			Spec:       item.Spec,
			Status:     item.Status,
		})
	}
	return &SteveCollection{Data: data}, nil
}

// listK8sNativeByPath lists using the native Kubernetes API path (/apis/<group>/<version>/... or /api/v1/...).
func (c *SteveClient) listK8sNativeByPath(ctx context.Context, clusterID string, path *k8sAPIPath, opts ListOpts) (*SteveCollection, error) {
	var urlPath string
	if opts.Namespace != "" {
		urlPath = fmt.Sprintf("/k8s/clusters/%s%s/namespaces/%s/%s", clusterID, path.basePath(), opts.Namespace, path.resource)
	} else {
		urlPath = fmt.Sprintf("/k8s/clusters/%s%s/%s", clusterID, path.basePath(), path.resource)
	}
	u, err := url.Parse(c.baseURL + urlPath)
	if err != nil {
		return nil, fmt.Errorf("k8s list url: %w", err)
	}
	q := u.Query()
	if opts.Limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", opts.Limit))
	}
	if opts.LabelSelector != "" {
		q.Set("labelSelector", opts.LabelSelector)
	}
	if opts.FieldSelector != "" {
		q.Set("fieldSelector", opts.FieldSelector)
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
		return nil, fmt.Errorf("k8s list request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("k8s list %s: %s", resp.Status, string(body))
	}

	var list k8sListResponse
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return nil, fmt.Errorf("k8s list decode: %w", err)
	}
	data := make([]SteveResource, 0, len(list.Items))
	for _, raw := range list.Items {
		var item struct {
			APIVersion string      `json:"apiVersion"`
			Kind       string      `json:"kind"`
			Metadata   ObjectMeta  `json:"metadata"`
			Spec       interface{} `json:"spec,omitempty"`
			Status     interface{} `json:"status,omitempty"`
		}
		if err := json.Unmarshal(raw, &item); err != nil {
			continue
		}
		data = append(data, SteveResource{
			TypeMeta:   TypeMeta{Kind: item.Kind, APIVersion: item.APIVersion},
			ObjectMeta: item.Metadata,
			Spec:       item.Spec,
			Status:     item.Status,
		})
	}
	return &SteveCollection{Data: data}, nil
}

// getK8sNative gets a single resource via the native Kubernetes API and converts it to SteveResource.
func (c *SteveClient) getK8sNative(ctx context.Context, clusterID, k8sResource, namespace, name string) (*SteveResource, error) {
	var path string
	if namespace != "" {
		path = fmt.Sprintf("/k8s/clusters/%s/api/v1/namespaces/%s/%s/%s", clusterID, namespace, k8sResource, name)
	} else {
		path = fmt.Sprintf("/k8s/clusters/%s/api/v1/%s/%s", clusterID, k8sResource, name)
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
		return nil, fmt.Errorf("k8s get request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("k8s get %s: %s", resp.Status, string(body))
	}
	var item struct {
		APIVersion string      `json:"apiVersion"`
		Kind       string      `json:"kind"`
		Metadata   ObjectMeta  `json:"metadata"`
		Spec       interface{} `json:"spec,omitempty"`
		Status     interface{} `json:"status,omitempty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return nil, fmt.Errorf("k8s get decode: %w", err)
	}
	return &SteveResource{
		TypeMeta:   TypeMeta{Kind: item.Kind, APIVersion: item.APIVersion},
		ObjectMeta: item.Metadata,
		Spec:       item.Spec,
		Status:     item.Status,
	}, nil
}

// getK8sNativeByPath gets a single resource via the native Kubernetes API path.
func (c *SteveClient) getK8sNativeByPath(ctx context.Context, clusterID string, path *k8sAPIPath, namespace, name string) (*SteveResource, error) {
	var urlPath string
	if namespace != "" {
		urlPath = fmt.Sprintf("/k8s/clusters/%s%s/namespaces/%s/%s/%s", clusterID, path.basePath(), namespace, path.resource, name)
	} else {
		urlPath = fmt.Sprintf("/k8s/clusters/%s%s/%s/%s", clusterID, path.basePath(), path.resource, name)
	}
	u := c.baseURL + urlPath
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("k8s get request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("k8s get %s: %s", resp.Status, string(body))
	}
	var item struct {
		APIVersion string      `json:"apiVersion"`
		Kind       string      `json:"kind"`
		Metadata   ObjectMeta  `json:"metadata"`
		Spec       interface{} `json:"spec,omitempty"`
		Status     interface{} `json:"status,omitempty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return nil, fmt.Errorf("k8s get decode: %w", err)
	}
	return &SteveResource{
		TypeMeta:   TypeMeta{Kind: item.Kind, APIVersion: item.APIVersion},
		ObjectMeta: item.Metadata,
		Spec:       item.Spec,
		Status:     item.Status,
	}, nil
}

// Get a single resource. For core v1 types we try the native K8s API first, then Steve on 404.
// For other types we try Steve first, then native K8s API on 404.
func (c *SteveClient) Get(ctx context.Context, clusterID, resourceType, namespace, name string) (*SteveResource, error) {
	if k8sRes := steveTypeToK8sCore(resourceType); k8sRes != "" {
		res, err := c.getK8sNative(ctx, clusterID, k8sRes, namespace, name)
		if err == nil {
			return res, nil
		}
		if !strings.Contains(err.Error(), "404") && !strings.Contains(err.Error(), "403") {
			return nil, err
		}
	}
	res, err := c.get(ctx, clusterID, resourceType, namespace, name)
	if err != nil {
		if !strings.Contains(err.Error(), "404") {
			return nil, err
		}
		if path := steveTypeToK8sAPIPath(resourceType); path != nil {
			return c.getK8sNativeByPath(ctx, clusterID, path, namespace, name)
		}
		if alt := steveTypeFallback(resourceType); alt != "" {
			return c.get(ctx, clusterID, alt, namespace, name)
		}
		return nil, err
	}
	return res, nil
}

func (c *SteveClient) get(ctx context.Context, clusterID, resourceType, namespace, name string) (*SteveResource, error) {
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

// Create a resource. namespace empty for cluster-scoped.
func (c *SteveClient) Create(ctx context.Context, clusterID, resourceType, namespace string, body interface{}) (*SteveResource, error) {
	var path string
	if namespace != "" {
		path = fmt.Sprintf("/k8s/clusters/%s/v1/namespaces/%s/%s", clusterID, namespace, resourceType)
	} else {
		path = fmt.Sprintf("/k8s/clusters/%s/v1/%s", clusterID, resourceType)
	}
	b, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("steve create marshal: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("steve create request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("steve create %s: %s", resp.Status, string(body))
	}
	var res SteveResource
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("steve create decode: %w", err)
	}
	return &res, nil
}

// Update a resource (PUT). namespace empty for cluster-scoped.
// Tries Steve first; on 404 falls back to native Kubernetes API.
func (c *SteveClient) Update(ctx context.Context, clusterID, resourceType, namespace, name string, body interface{}) (*SteveResource, error) {
	res, err := c.update(ctx, clusterID, resourceType, namespace, name, body)
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			if path := steveTypeToK8sAPIPath(resourceType); path != nil {
				return c.updateK8sNativeByPath(ctx, clusterID, path, namespace, name, body)
			}
		}
		return nil, err
	}
	return res, nil
}

func (c *SteveClient) update(ctx context.Context, clusterID, resourceType, namespace, name string, body interface{}) (*SteveResource, error) {
	var path string
	if namespace != "" {
		path = fmt.Sprintf("/k8s/clusters/%s/v1/namespaces/%s/%s/%s", clusterID, namespace, resourceType, name)
	} else {
		path = fmt.Sprintf("/k8s/clusters/%s/v1/%s/%s", clusterID, resourceType, name)
	}
	b, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("steve update marshal: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.baseURL+path, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("steve update request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("steve update %s: %s", resp.Status, string(body))
	}
	var res SteveResource
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("steve update decode: %w", err)
	}
	return &res, nil
}

// updateK8sNativeByPath updates a resource via the native Kubernetes API (PUT).
func (c *SteveClient) updateK8sNativeByPath(ctx context.Context, clusterID string, path *k8sAPIPath, namespace, name string, body interface{}) (*SteveResource, error) {
	var urlPath string
	if namespace != "" {
		urlPath = fmt.Sprintf("/k8s/clusters/%s%s/namespaces/%s/%s/%s", clusterID, path.basePath(), namespace, path.resource, name)
	} else {
		urlPath = fmt.Sprintf("/k8s/clusters/%s%s/%s/%s", clusterID, path.basePath(), path.resource, name)
	}
	b, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("k8s update marshal: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.baseURL+urlPath, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("k8s update request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("k8s update %s: %s", resp.Status, string(body))
	}
	var item struct {
		APIVersion string      `json:"apiVersion"`
		Kind       string      `json:"kind"`
		Metadata   ObjectMeta  `json:"metadata"`
		Spec       interface{} `json:"spec,omitempty"`
		Status     interface{} `json:"status,omitempty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return nil, fmt.Errorf("k8s update decode: %w", err)
	}
	return &SteveResource{
		TypeMeta:   TypeMeta{Kind: item.Kind, APIVersion: item.APIVersion},
		ObjectMeta: item.Metadata,
		Spec:       item.Spec,
		Status:     item.Status,
	}, nil
}

// Delete a resource. namespace empty for cluster-scoped.
// Tries Steve first; on 404 falls back to native Kubernetes API.
func (c *SteveClient) Delete(ctx context.Context, clusterID, resourceType, namespace, name string) error {
	err := c.delete(ctx, clusterID, resourceType, namespace, name)
	if err != nil && strings.Contains(err.Error(), "404") {
		if path := steveTypeToK8sAPIPath(resourceType); path != nil {
			return c.deleteK8sNativeByPath(ctx, clusterID, path, namespace, name)
		}
	}
	return err
}

func (c *SteveClient) delete(ctx context.Context, clusterID, resourceType, namespace, name string) error {
	var path string
	if namespace != "" {
		path = fmt.Sprintf("/k8s/clusters/%s/v1/namespaces/%s/%s/%s", clusterID, namespace, resourceType, name)
	} else {
		path = fmt.Sprintf("/k8s/clusters/%s/v1/%s/%s", clusterID, resourceType, name)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.baseURL+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("steve delete request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("steve delete %s: %s", resp.Status, string(body))
	}
	return nil
}

// deleteK8sNativeByPath deletes a resource via the native Kubernetes API.
func (c *SteveClient) deleteK8sNativeByPath(ctx context.Context, clusterID string, path *k8sAPIPath, namespace, name string) error {
	var urlPath string
	if namespace != "" {
		urlPath = fmt.Sprintf("/k8s/clusters/%s%s/namespaces/%s/%s/%s", clusterID, path.basePath(), namespace, path.resource, name)
	} else {
		urlPath = fmt.Sprintf("/k8s/clusters/%s%s/%s/%s", clusterID, path.basePath(), path.resource, name)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.baseURL+urlPath, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("k8s delete request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("k8s delete %s: %s", resp.Status, string(body))
	}
	return nil
}

// SteveType returns the Steve API resource type for the given apiVersion and kind.
// Rancher Steve uses "core" as the API group for core/v1 resources, so "v1", "Pod" -> "core.v1.pods".
// Other groups: "apps/v1", "Deployment" -> "apps.v1.deployments".
func SteveType(apiVersion, kind string) string {
	k := strings.ToLower(kind)
	// Irregular plurals (K8s resource names)
	switch k {
	case "endpoints", "events":
		// already plural
	case "ingress":
		k = "ingresses"
	default:
		if !strings.HasSuffix(k, "s") {
			k += "s"
		}
	}
	// Core API group (no group in K8s) is exposed as "core" in Rancher Steve.
	if apiVersion == "" || apiVersion == "v1" {
		return "core.v1." + k
	}
	groupVersion := strings.ReplaceAll(apiVersion, "/", ".")
	return groupVersion + "." + k
}
