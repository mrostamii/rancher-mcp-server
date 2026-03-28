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

// NormanClient calls the Rancher Norman management API (HTTP path prefix /v3).
type NormanClient struct {
	baseURL    string
	token      string
	insecure   bool
	httpClient *http.Client
}

// NewNormanClient creates a Norman API client. baseURL is the Rancher server URL (e.g. https://rancher.example.com).
func NewNormanClient(baseURL, token string, insecure bool) *NormanClient {
	u, _ := url.Parse(baseURL)
	if u.Scheme == "" {
		u.Scheme = "https"
	}
	base := strings.TrimSuffix(u.String(), "/")
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure},
	}
	return &NormanClient{
		baseURL:    base,
		token:      token,
		insecure:   insecure,
		httpClient: &http.Client{Transport: tr},
	}
}

// Do issues an HTTP request under /v3. path may be "schemas", "users", "users/u-abc", etc. (no leading slash) or "/v3/...".
func (c *NormanClient) Do(ctx context.Context, method, path string, query url.Values, body []byte) ([]byte, int, error) {
	p := strings.TrimPrefix(path, "/")
	if !strings.HasPrefix(p, "v3/") {
		p = "v3/" + strings.TrimPrefix(p, "/")
	}
	u, err := url.Parse(c.baseURL + "/" + p)
	if err != nil {
		return nil, 0, fmt.Errorf("norman parse url: %w", err)
	}
	if len(query) > 0 {
		u.RawQuery = query.Encode()
	}
	var rdr io.Reader
	if body != nil {
		rdr = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, u.String(), rdr)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("norman request: %w", err)
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("norman read body: %w", err)
	}
	return b, resp.StatusCode, nil
}

// RedactNormanSecrets removes sensitive fields from decoded JSON when showSensitive is false.
func RedactNormanSecrets(showSensitive bool, raw []byte) ([]byte, error) {
	if showSensitive {
		return raw, nil
	}
	var v interface{}
	if err := json.Unmarshal(raw, &v); err != nil {
		return raw, nil
	}
	redactValue(v)
	return json.MarshalIndent(v, "", "  ")
}

func redactValue(v interface{}) {
	switch x := v.(type) {
	case map[string]interface{}:
		for k, val := range x {
			lk := strings.ToLower(k)
			if lk == "token" || lk == "accesskey" || lk == "secretkey" || lk == "password" || lk == "serviceaccounttoken" {
				x[k] = "<redacted>"
				continue
			}
			if lk == "kubeconfig" || lk == "amazonec2config" || lk == "azureadconfig" {
				x[k] = "<redacted>"
				continue
			}
			redactValue(val)
		}
	case []interface{}:
		for _, el := range x {
			redactValue(el)
		}
	}
}
