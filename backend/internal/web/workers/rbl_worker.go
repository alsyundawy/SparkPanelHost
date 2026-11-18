package workers

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/alsyundawy/SparkPanelHost/backend/internal/web/store"
)

var defaultRBLProviders = []string{
	"zen.spamhaus.org",
	"b.barracudacentral.org",
	"bl.spamcop.net",
	"dnsbl.sorbs.net",
}

// RunRBLWorker periodically queries major RBL/DNSBL providers to check if the server is blacklisted.
func RunRBLWorker(ctx context.Context, s *store.Store, checkIP string, interval time.Duration) {
	if checkIP == "" {
		checkIP = "127.0.0.1"
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Initial check
	checkRBL(ctx, s, checkIP)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			checkRBL(ctx, s, checkIP)
		}
	}
}

func checkRBL(ctx context.Context, s *store.Store, ipStr string) {
	parsedIP := net.ParseIP(ipStr)
	if parsedIP == nil || parsedIP.To4() == nil {
		return
	}

	ipv4 := parsedIP.To4()
	reversedIP := fmt.Sprintf("%d.%d.%d.%d", ipv4[3], ipv4[2], ipv4[1], ipv4[0])

	for _, provider := range defaultRBLProviders {
		lookupHost := fmt.Sprintf("%s.%s", reversedIP, provider)
		ips, err := net.LookupHost(lookupHost)
		if err == nil && len(ips) > 0 {
			log.Printf("[RBL-Monitor] WARNING: Server IP %s is listed on RBL %s (response: %v)", ipStr, provider, ips)
			_ = s.AppendAuditLog(ctx, "system:rbl_worker", nil, "rbl.blacklisted", fmt.Sprintf("provider:%s,ip:%s", provider, ipStr), "127.0.0.1")
		} else if err != nil && strings.Contains(err.Error(), "no such host") {
			// Clean (not listed)
		}
	}
}
