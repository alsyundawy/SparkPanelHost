package drivers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/alsyundawy/SparkPanelHost/backend/internal/agent/execsafe"
)

type DNSRecord struct {
	Name     string
	Type     string
	Content  string
	TTL      uint32
	Priority uint32
}

type DNSClusterNodeInfo struct {
	ClusterName  string
	Hostname     string
	APIEndpoint  string
	Role         string // "write_only", "synchronize", "standalone", "dns_only"
	APIToken     string
	LastSeenUnix int64
	ZonesSynced  uint32
	Status       string // "online", "offline", "drift"
}

var (
	clusterNodesLock sync.RWMutex
	clusterNodes     = make(map[string][]DNSClusterNodeInfo) // clusterName -> []DNSClusterNodeInfo
)

func getPDNSAPIKey() string {
	if key := os.Getenv("PDNS_API_KEY"); key != "" {
		return key
	}
	// Fallback check in /etc/aetherpanel/secrets.env
	if data, err := os.ReadFile("/etc/aetherpanel/secrets.env"); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(line, "PDNS_API_KEY=") {
				return strings.Trim(strings.TrimPrefix(line, "PDNS_API_KEY="), `"'`)
			}
		}
	}
	return "aetherpanel-pdns-secret"
}

// SyncDNSZone sends zone and RRsets to local PowerDNS 5.1 via REST API.
func SyncDNSZone(ctx context.Context, zoneName string, records []DNSRecord, enableDNSSEC bool) error {
	zoneCanonical := strings.TrimSuffix(zoneName, ".") + "."
	apiKey := getPDNSAPIKey()

	// PowerDNS RRsets payload
	type rrsetRecord struct {
		Content  string `json:"content"`
		Disabled bool   `json:"disabled"`
	}
	type rrset struct {
		Name       string        `json:"name"`
		Type       string        `json:"type"`
		TTL        uint32        `json:"ttl"`
		ChangeType string        `json:"changetype"`
		Records    []rrsetRecord `json:"records"`
	}
	type patchPayload struct {
		RRsets []rrset `json:"rrsets"`
	}

	// Group records by (name, type)
	grouped := make(map[string]map[string][]string)
	ttls := make(map[string]uint32)

	for _, r := range records {
		recName := strings.TrimSuffix(r.Name, ".") + "."
		if !strings.HasSuffix(recName, zoneCanonical) {
			recName = recName + zoneCanonical
		}
		if grouped[recName] == nil {
			grouped[recName] = make(map[string][]string)
		}
		content := r.Content
		if r.Type == "MX" && r.Priority > 0 {
			content = fmt.Sprintf("%d %s", r.Priority, r.Content)
		}
		grouped[recName][r.Type] = append(grouped[recName][r.Type], content)
		ttl := r.TTL
		if ttl == 0 {
			ttl = 300
		}
		ttls[recName+":"+r.Type] = ttl
	}

	var payload patchPayload
	for recName, typeMap := range grouped {
		for recType, contents := range typeMap {
			var recs []rrsetRecord
			for _, c := range contents {
				recs = append(recs, rrsetRecord{Content: c, Disabled: false})
			}
			payload.RRsets = append(payload.RRsets, rrset{
				Name:       recName,
				Type:       recType,
				TTL:        ttls[recName+":"+recType],
				ChangeType: "REPLACE",
				Records:    recs,
			})
		}
	}

	// First ensure zone exists in PowerDNS
	client := &http.Client{Timeout: 10 * time.Second}
	zoneURL := fmt.Sprintf("http://127.0.0.1:8081/api/v1/servers/localhost/zones/%s", zoneCanonical)

	// Check if zone exists
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, zoneURL, nil)
	req.Header.Set("X-API-Key", apiKey)
	resp, err := client.Do(req)

	if err == nil && resp.StatusCode == http.StatusNotFound {
		_ = resp.Body.Close()
		// Create zone
		createPayload := map[string]interface{}{
			"name":         zoneCanonical,
			"kind":         "Native",
			"nameservers":  []string{"ns1." + zoneCanonical, "ns2." + zoneCanonical},
			"dnssec":       enableDNSSEC,
			"soa_edit_api": "DEFAULT",
		}
		b, _ := json.Marshal(createPayload)
		cReq, _ := http.NewRequestWithContext(ctx, http.MethodPost, "http://127.0.0.1:8081/api/v1/servers/localhost/zones", bytes.NewReader(b))
		cReq.Header.Set("X-API-Key", apiKey)
		cReq.Header.Set("Content-Type", "application/json")
		cResp, cErr := client.Do(cReq)
		if cErr == nil {
			_ = cResp.Body.Close()
		}
	} else if resp != nil {
		_ = resp.Body.Close()
	}

	// Apply RRset updates
	patchBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	pReq, err := http.NewRequestWithContext(ctx, http.MethodPatch, zoneURL, bytes.NewReader(patchBytes))
	if err != nil {
		return err
	}
	pReq.Header.Set("X-API-Key", apiKey)
	pReq.Header.Set("Content-Type", "application/json")

	pResp, err := client.Do(pReq)
	if err != nil {
		return fmt.Errorf("powerdns API communication error: %w", err)
	}
	defer pResp.Body.Close()

	if pResp.StatusCode != http.StatusOK && pResp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(pResp.Body)
		return fmt.Errorf("powerdns API error (status %d): %s", pResp.StatusCode, string(body))
	}

	// Configure DNSSEC via pdnsutil if requested
	if enableDNSSEC {
		_, _ = execsafe.RunDirect(ctx, 15*time.Second, "/usr/bin/pdnsutil", "secure-zone", zoneCanonical)
	}

	return nil
}

// RegisterDNSClusterNode registers a remote node for DNS clustering (WHM DNS Only style).
func RegisterDNSClusterNode(ctx context.Context, clusterName, hostname, apiEndpoint, role, apiToken string) error {
	clusterNodesLock.Lock()
	defer clusterNodesLock.Unlock()

	// Check if already registered
	list := clusterNodes[clusterName]
	for i, n := range list {
		if n.Hostname == hostname {
			list[i].APIEndpoint = apiEndpoint
			list[i].Role = role
			list[i].APIToken = apiToken
			list[i].LastSeenUnix = time.Now().Unix()
			list[i].Status = "online"
			clusterNodes[clusterName] = list
			return nil
		}
	}

	newNode := DNSClusterNodeInfo{
		ClusterName:  clusterName,
		Hostname:     hostname,
		APIEndpoint:  apiEndpoint,
		Role:         role,
		APIToken:     apiToken,
		LastSeenUnix: time.Now().Unix(),
		ZonesSynced:  0,
		Status:       "online",
	}

	clusterNodes[clusterName] = append(clusterNodes[clusterName], newNode)
	return nil
}

// SyncDNSZoneToCluster pushes zone records to cluster members (Star topology: push to DNS-Only nodes).
func SyncDNSZoneToCluster(ctx context.Context, zoneName, clusterName string, force bool) error {
	clusterNodesLock.RLock()
	nodes := clusterNodes[clusterName]
	clusterNodesLock.RUnlock()

	if len(nodes) == 0 {
		return nil // No remote cluster nodes configured
	}

	zoneCanonical := strings.TrimSuffix(zoneName, ".") + "."
	apiKey := getPDNSAPIKey()

	// Read zone from local PowerDNS
	client := &http.Client{Timeout: 10 * time.Second}
	zoneURL := fmt.Sprintf("http://127.0.0.1:8081/api/v1/servers/localhost/zones/%s", zoneCanonical)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, zoneURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-API-Key", apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch local zone for cluster sync: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("local zone not found in PowerDNS (status %d)", resp.StatusCode)
	}

	zoneData, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Push zone to all remote DNS-Only and Synchronize nodes
	var syncErrors []string
	for idx, node := range nodes {
		if node.Role == "standalone" {
			continue
		}

		remoteURL := strings.TrimRight(node.APIEndpoint, "/") + "/api/v1/servers/localhost/zones/" + zoneCanonical
		pushReq, err := http.NewRequestWithContext(ctx, http.MethodPut, remoteURL, bytes.NewReader(zoneData))
		if err != nil {
			syncErrors = append(syncErrors, fmt.Sprintf("%s: %v", node.Hostname, err))
			continue
		}
		pushReq.Header.Set("X-API-Key", node.APIToken)
		pushReq.Header.Set("Content-Type", "application/json")

		pResp, pErr := client.Do(pushReq)
		if pErr != nil {
			syncErrors = append(syncErrors, fmt.Sprintf("%s: %v", node.Hostname, pErr))
			clusterNodesLock.Lock()
			nodes[idx].Status = "offline"
			clusterNodesLock.Unlock()
			continue
		}
		_ = pResp.Body.Close()

		if pResp.StatusCode == http.StatusOK || pResp.StatusCode == http.StatusCreated || pResp.StatusCode == http.StatusNoContent {
			clusterNodesLock.Lock()
			nodes[idx].LastSeenUnix = time.Now().Unix()
			nodes[idx].ZonesSynced++
			nodes[idx].Status = "online"
			clusterNodesLock.Unlock()
		} else {
			syncErrors = append(syncErrors, fmt.Sprintf("%s: HTTP status %d", node.Hostname, pResp.StatusCode))
		}
	}

	if len(syncErrors) > 0 {
		return fmt.Errorf("partial DNS cluster sync failures: %s", strings.Join(syncErrors, "; "))
	}
	return nil
}

// GetDNSClusterStatus returns live status of all cluster nodes.
func GetDNSClusterStatus(ctx context.Context, clusterName string) ([]DNSClusterNodeInfo, error) {
	clusterNodesLock.RLock()
	nodes, exists := clusterNodes[clusterName]
	clusterNodesLock.RUnlock()

	if !exists {
		return []DNSClusterNodeInfo{}, nil
	}

	result := make([]DNSClusterNodeInfo, len(nodes))
	copy(result, nodes)
	return result, nil
}
