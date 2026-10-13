package harvester

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mrostamii/rancher-mcp-server/internal/security"
	"github.com/mrostamii/rancher-mcp-server/pkg/client/rancher"
)

func resultText(t *testing.T, result *mcp.CallToolResult) string {
	t.Helper()
	if result == nil || len(result.Content) == 0 {
		t.Fatal("empty tool result")
	}
	tc, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}
	return tc.Text
}

func decodeBody(t *testing.T, r *http.Request) map[string]interface{} {
	t.Helper()
	b, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("decode body %s: %v", b, err)
	}
	return m
}

func TestSnapshotRestore_HaltAfterRestore(t *testing.T) {
	var got map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && strings.Contains(r.URL.Path, "virtualmachinerestores") {
			got = decodeBody(t, r)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"metadata":{"name":"restore-snap1"}}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	toolset := NewToolset(rancher.NewSteveClient(srv.URL, "token", true), &security.Policy{})
	result, err := toolset.vmSnapshotHandler(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{Arguments: map[string]interface{}{
			"cluster": "c-xxx", "action": "restore", "namespace": "default",
			"snapshot_name": "snap1", "vm_name": "vm1", "halt_after_restore": "true",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatal(resultText(t, result))
	}
	spec, _ := got["spec"].(map[string]interface{})
	if spec["haltAfterRestore"] != true {
		t.Fatalf("spec.haltAfterRestore = %#v, want true; body %#v", spec["haltAfterRestore"], got)
	}
}

func TestSnapshotRestore_OmitsHaltByDefault(t *testing.T) {
	var got map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && strings.Contains(r.URL.Path, "virtualmachinerestores") {
			got = decodeBody(t, r)
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"metadata":{"name":"restore-snap1"}}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	toolset := NewToolset(rancher.NewSteveClient(srv.URL, "token", true), &security.Policy{})
	result, err := toolset.vmSnapshotHandler(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{Arguments: map[string]interface{}{
			"cluster": "c-xxx", "action": "restore", "namespace": "default",
			"snapshot_name": "snap1", "vm_name": "vm1",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatal(resultText(t, result))
	}
	spec, _ := got["spec"].(map[string]interface{})
	if _, ok := spec["haltAfterRestore"]; ok {
		t.Fatalf("haltAfterRestore should be omitted by default, spec=%#v", spec)
	}
}

func TestBackupRestore_HaltAfterRestore(t *testing.T) {
	var got map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && strings.Contains(r.URL.Path, "virtualmachinerestores") {
			got = decodeBody(t, r)
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"metadata":{"name":"restore-bak1"}}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	toolset := NewToolset(rancher.NewSteveClient(srv.URL, "token", true), &security.Policy{})
	result, err := toolset.vmBackupHandler(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{Arguments: map[string]interface{}{
			"cluster": "c-xxx", "action": "restore", "namespace": "default",
			"backup_name": "bak1", "vm_name": "vm1", "halt_after_restore": true,
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatal(resultText(t, result))
	}
	spec, _ := got["spec"].(map[string]interface{})
	if spec["haltAfterRestore"] != true {
		t.Fatalf("spec.haltAfterRestore = %#v, want true", spec["haltAfterRestore"])
	}
	if spec["virtualMachineBackupName"] != "bak1" {
		t.Fatalf("backup name = %#v", spec["virtualMachineBackupName"])
	}
}

func TestVMCreate_VirtioFSAndSysprep(t *testing.T) {
	var vm map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "virtualmachineimages"):
			_, _ = w.Write([]byte(`{"metadata":{"name":"img"},"status":{"storageClassName":"longhorn-image-img"}}`))
		case r.Method == http.MethodPost && strings.Contains(r.URL.Path, "persistentvolumeclaims"):
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"metadata":{"name":"vm-disk-0"}}`))
		case r.Method == http.MethodPost && strings.Contains(r.URL.Path, "virtualmachines"):
			vm = decodeBody(t, r)
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"metadata":{"name":"win"}}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	toolset := NewToolset(rancher.NewSteveClient(srv.URL, "token", true), &security.Policy{})
	result, err := toolset.vmCreateHandler(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{Arguments: map[string]interface{}{
			"cluster": "c-xxx", "namespace": "default", "name": "win", "image": "img",
			"virtiofs_service_account": "builder",
			"virtiofs_secret":          "app-secret",
			"virtiofs_config_map":      "app-cm",
			"sysprep_secret":           "unattend",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatal(resultText(t, result))
	}
	if vm == nil {
		t.Fatal("VM was not created")
	}
	spec := vm["spec"].(map[string]interface{})
	tmpl := spec["template"].(map[string]interface{})["spec"].(map[string]interface{})
	domain := tmpl["domain"].(map[string]interface{})
	devices := domain["devices"].(map[string]interface{})
	filesystems, _ := devices["filesystems"].([]interface{})
	if len(filesystems) != 3 {
		t.Fatalf("filesystems = %#v, want 3 virtiofs mounts", devices["filesystems"])
	}
	for _, item := range filesystems {
		fs := item.(map[string]interface{})
		if _, ok := fs["virtiofs"]; !ok {
			t.Fatalf("filesystem missing virtiofs: %#v", fs)
		}
	}
	disks := devices["disks"].([]interface{})
	var sysprepDisk map[string]interface{}
	for _, item := range disks {
		d := item.(map[string]interface{})
		if d["name"] == "sysprep" {
			sysprepDisk = d
		}
	}
	if sysprepDisk == nil {
		t.Fatalf("sysprep disk missing: %#v", disks)
	}
	cdrom, _ := sysprepDisk["cdrom"].(map[string]interface{})
	if cdrom["bus"] != "sata" {
		t.Fatalf("sysprep bus = %#v, want sata", cdrom["bus"])
	}
	volumes := tmpl["volumes"].([]interface{})
	found := map[string]bool{}
	for _, item := range volumes {
		vol := item.(map[string]interface{})
		switch vol["name"] {
		case "virtiofs-serviceaccount":
			sa := vol["serviceAccount"].(map[string]interface{})
			if sa["serviceAccountName"] != "builder" {
				t.Fatalf("service account volume %#v", vol)
			}
			found["sa"] = true
		case "virtiofs-secret":
			sec := vol["secret"].(map[string]interface{})
			if sec["secretName"] != "app-secret" {
				t.Fatalf("secret volume %#v", vol)
			}
			found["secret"] = true
		case "virtiofs-configmap":
			cm := vol["configMap"].(map[string]interface{})
			if cm["name"] != "app-cm" {
				t.Fatalf("configmap volume %#v", vol)
			}
			found["cm"] = true
		case "sysprep":
			sp := vol["sysprep"].(map[string]interface{})["secret"].(map[string]interface{})
			if sp["name"] != "unattend" {
				t.Fatalf("sysprep volume %#v", vol)
			}
			found["sysprep"] = true
		}
	}
	if len(found) != 4 {
		t.Fatalf("volumes missing virtiofs/sysprep entries: %#v", volumes)
	}
}

func TestVMCreate_RejectsInvalidVirtioFSName(t *testing.T) {
	toolset := NewToolset(rancher.NewSteveClient("http://127.0.0.1", "token", true), &security.Policy{})
	result, err := toolset.vmCreateHandler(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{Arguments: map[string]interface{}{
			"cluster": "c-xxx", "namespace": "default", "name": "win", "image": "img",
			"virtiofs_secret": "not a name",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Fatal("expected invalid virtiofs name to fail before any API call")
	}
}

func TestVpcNatGatewayCreate(t *testing.T) {
	var got map[string]interface{}
	var path string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && strings.Contains(r.URL.Path, "vpc-nat-gateways") {
			path = r.URL.Path
			got = decodeBody(t, r)
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"metadata":{"name":"gw1"}}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	toolset := NewToolset(rancher.NewSteveClient(srv.URL, "token", true), &security.Policy{})
	result, err := toolset.vpcNatGatewayHandler(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{Arguments: map[string]interface{}{
			"cluster": "c-xxx", "action": "create", "kind": "gateway",
			"name": "gw1", "vpc": "tenant", "subnet": "tenant-subnet",
			"lan_ip": "10.16.0.254", "external_subnets": "ovn-vpc-external-network",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatal(resultText(t, result))
	}
	if !strings.Contains(path, "/apis/kubeovn.io/v1/vpc-nat-gateways") {
		t.Fatalf("path %s", path)
	}
	if got["kind"] != "VpcNatGateway" {
		t.Fatalf("kind %#v", got["kind"])
	}
	spec := got["spec"].(map[string]interface{})
	if spec["vpc"] != "tenant" || spec["subnet"] != "tenant-subnet" || spec["lanIp"] != "10.16.0.254" {
		t.Fatalf("spec %#v", spec)
	}
	ext, _ := spec["externalSubnets"].([]interface{})
	if len(ext) != 1 || ext[0] != "ovn-vpc-external-network" {
		t.Fatalf("externalSubnets %#v", spec["externalSubnets"])
	}
}

func TestVpcNatRuleCreateDNATAndSNAT(t *testing.T) {
	var dnat, snat, eip map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		body := decodeBody(t, r)
		switch {
		case strings.Contains(r.URL.Path, "iptables-dnat-rules"):
			dnat = body
		case strings.Contains(r.URL.Path, "iptables-snat-rules"):
			snat = body
		case strings.Contains(r.URL.Path, "iptables-eips"):
			eip = body
		default:
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"metadata":{"name":"ok"}}`))
	}))
	defer srv.Close()

	toolset := NewToolset(rancher.NewSteveClient(srv.URL, "token", true), &security.Policy{})
	calls := []map[string]interface{}{
		{"cluster": "c-xxx", "action": "create", "kind": "eip", "name": "eip1", "nat_gw": "gw1"},
		{"cluster": "c-xxx", "action": "create", "kind": "dnat", "name": "dnat1", "eip": "eip1", "external_port": "80", "internal_ip": "10.16.0.5", "internal_port": "8080", "protocol": "tcp"},
		{"cluster": "c-xxx", "action": "create", "kind": "snat", "name": "snat1", "eip": "eip1", "internal_cidr": "10.16.0.0/24"},
	}
	for _, args := range calls {
		result, err := toolset.vpcNatGatewayHandler(context.Background(), mcp.CallToolRequest{
			Params: mcp.CallToolParams{Arguments: args},
		})
		if err != nil {
			t.Fatal(err)
		}
		if result.IsError {
			t.Fatal(resultText(t, result))
		}
	}
	if eip["kind"] != "IptablesEIP" || eip["spec"].(map[string]interface{})["natGwDp"] != "gw1" {
		t.Fatalf("eip %#v", eip)
	}
	dspec := dnat["spec"].(map[string]interface{})
	if dnat["kind"] != "IptablesDnatRule" || dspec["eip"] != "eip1" || dspec["externalPort"] != "80" || dspec["internalIp"] != "10.16.0.5" || dspec["internalPort"] != "8080" || dspec["protocol"] != "tcp" {
		t.Fatalf("dnat %#v", dnat)
	}
	sspec := snat["spec"].(map[string]interface{})
	if snat["kind"] != "IptablesSnatRule" || sspec["eip"] != "eip1" || sspec["internalCIDR"] != "10.16.0.0/24" {
		t.Fatalf("snat %#v", snat)
	}
}

func TestVpcNatGatewayDelete_RequiresDestructive(t *testing.T) {
	toolset := NewToolset(rancher.NewSteveClient("http://127.0.0.1", "token", true), &security.Policy{DisableDestructive: true})
	result, err := toolset.vpcNatGatewayHandler(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{Arguments: map[string]interface{}{
			"cluster": "c-xxx", "action": "delete", "kind": "gateway", "name": "gw1",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Fatal("delete should be blocked when destructive ops are disabled")
	}
}

func TestSettingsGet_ReadsTopLevelValue(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && strings.Contains(r.URL.Path, "cluster-pod-security-standard") {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"apiVersion":"harvesterhci.io/v1beta1","kind":"Setting","metadata":{"name":"cluster-pod-security-standard","resourceVersion":"7"},"value":"{\"enabled\":false}"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	toolset := NewToolset(rancher.NewSteveClient(srv.URL, "token", true), &security.Policy{})
	result, err := toolset.settingsHandler(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{Arguments: map[string]interface{}{
			"cluster": "c-xxx", "name": "cluster-pod-security-standard",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatal(resultText(t, result))
	}
	text := resultText(t, result)
	if !strings.Contains(text, `"enabled":false`) && !strings.Contains(text, `\"enabled\":false`) {
		t.Fatalf("setting value not returned: %s", text)
	}
}

func TestSettingsUpdate_PodSecurityStandard(t *testing.T) {
	var got map[string]interface{}
	var method, path string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet && strings.Contains(r.URL.Path, "cluster-pod-security-standard") {
			_, _ = w.Write([]byte(`{"apiVersion":"harvesterhci.io/v1beta1","kind":"Setting","metadata":{"name":"cluster-pod-security-standard","resourceVersion":"9"},"value":"{\"enabled\":false}"}`))
			return
		}
		if r.Method == http.MethodPut && strings.Contains(r.URL.Path, "cluster-pod-security-standard") {
			method, path = r.Method, r.URL.Path
			got = decodeBody(t, r)
			_, _ = w.Write([]byte(`{"metadata":{"name":"cluster-pod-security-standard"}}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	toolset := NewToolset(rancher.NewSteveClient(srv.URL, "token", true), &security.Policy{})
	result, err := toolset.settingsHandler(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{Arguments: map[string]interface{}{
			"cluster": "c-xxx", "name": "cluster-pod-security-standard",
			"pss_enabled": "true", "pss_whitelisted": "default, user-ns",
			"pss_privileged": "kube-system", "pss_restricted": "apps",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatal(resultText(t, result))
	}
	if method != http.MethodPut || !strings.Contains(path, "/apis/harvesterhci.io/v1beta1/settings/cluster-pod-security-standard") {
		t.Fatalf("update %s %s", method, path)
	}
	meta := got["metadata"].(map[string]interface{})
	if meta["resourceVersion"] != "9" {
		t.Fatalf("resourceVersion %#v", meta["resourceVersion"])
	}
	var pss map[string]interface{}
	if err := json.Unmarshal([]byte(got["value"].(string)), &pss); err != nil {
		t.Fatal(err)
	}
	if pss["enabled"] != true {
		t.Fatalf("enabled %#v", pss["enabled"])
	}
	if pss["whitelistedNamespacesList"] != "default,user-ns" && pss["whitelistedNamespacesList"] != "default, user-ns" {
		t.Fatalf("whitelist %#v", pss["whitelistedNamespacesList"])
	}
	if pss["privilegedNamespacesList"] != "kube-system" || pss["restrictedNamespacesList"] != "apps" {
		t.Fatalf("lists %#v", pss)
	}
}

func TestSettingsUpdate_GenericValue(t *testing.T) {
	var got map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet && strings.Contains(r.URL.Path, "backup-target") {
			_, _ = w.Write([]byte(`{"apiVersion":"harvesterhci.io/v1beta1","kind":"Setting","metadata":{"name":"backup-target","resourceVersion":"3"},"value":"old","default":"{\"type\":\"nfs\"}"}`))
			return
		}
		if r.Method == http.MethodPut && strings.Contains(r.URL.Path, "backup-target") {
			got = decodeBody(t, r)
			_, _ = w.Write([]byte(`{"metadata":{"name":"backup-target"}}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	toolset := NewToolset(rancher.NewSteveClient(srv.URL, "token", true), &security.Policy{})
	result, err := toolset.settingsHandler(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{Arguments: map[string]interface{}{
			"cluster": "c-xxx", "name": "backup-target", "value": `{"type":"nfs"}`,
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatal(resultText(t, result))
	}
	if got["value"] != `{"type":"nfs"}` {
		t.Fatalf("value %#v", got["value"])
	}
	if got["metadata"].(map[string]interface{})["resourceVersion"] != "3" {
		t.Fatalf("metadata %#v", got["metadata"])
	}
	if got["default"] != `{"type":"nfs"}` {
		t.Fatalf("default was dropped: %#v", got["default"])
	}

	result, err = toolset.settingsHandler(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{Arguments: map[string]interface{}{
			"cluster": "c-xxx", "name": "backup-target", "value": "",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatal(resultText(t, result))
	}
	if got["value"] != "" {
		t.Fatalf("empty value was not written: %#v", got["value"])
	}
}

func TestSettingsUpdate_RejectsReadOnlyAndOverlap(t *testing.T) {
	ro := NewToolset(rancher.NewSteveClient("http://127.0.0.1", "token", true), &security.Policy{ReadOnly: true})
	result, err := ro.settingsHandler(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{Arguments: map[string]interface{}{
			"cluster": "c-xxx", "name": "backup-target", "value": `{"type":"nfs"}`,
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Fatal("read-only must reject setting updates")
	}

	overlap := NewToolset(rancher.NewSteveClient("http://127.0.0.1", "token", true), &security.Policy{})
	result, err = overlap.settingsHandler(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{Arguments: map[string]interface{}{
			"cluster": "c-xxx", "pss_enabled": "true",
			"pss_privileged": "apps", "pss_restricted": "apps",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Fatal("overlapping PSS namespaces must be rejected")
	}
}
