package drivers

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/alsyundawy/SparkPanelHost/backend/internal/agent/execsafe"
)

type FirewallRule struct {
	Action   string // "allow", "deny", "remove"
	CIDRorIP string
	Port     uint32
	Protocol string // "tcp", "udp", "any"
	Comment  string
}

// ApplyFirewallRule adds or removes an atomic rule in table inet aetherpanel using nftables.
func ApplyFirewallRule(ctx context.Context, rule FirewallRule) error {
	cidr := rule.CIDRorIP
	if !strings.Contains(cidr, "/") {
		if ip := net.ParseIP(cidr); ip != nil {
			if ip.To4() != nil {
				cidr = cidr + "/32"
			} else {
				cidr = cidr + "/128"
			}
		} else {
			return fmt.Errorf("invalid CIDR or IP address: %q", rule.CIDRorIP)
		}
	} else {
		if _, _, err := net.ParseCIDR(cidr); err != nil {
			return fmt.Errorf("invalid CIDR: %w", err)
		}
	}

	proto := strings.ToLower(rule.Protocol)
	if proto != "tcp" && proto != "udp" && proto != "any" {
		proto = "tcp"
	}

	chain := "input"
	nftAction := "accept"
	if rule.Action == "deny" {
		nftAction = "drop"
	}

	var args []string
	if rule.Action == "remove" {
		// Flush or delete specific matching rule
		args = []string{"delete", "rule", "inet", "aetherpanel", chain}
	} else {
		args = []string{"add", "rule", "inet", "aetherpanel", chain}
	}

	// ip saddr <cidr>
	if strings.Contains(cidr, ":") {
		args = append(args, "ip6", "saddr", cidr)
	} else {
		args = append(args, "ip", "saddr", cidr)
	}

	// Port filter
	if rule.Port > 0 {
		if proto == "any" {
			proto = "tcp"
		}
		args = append(args, proto, "dport", fmt.Sprintf("%d", rule.Port))
	}

	if rule.Action != "remove" {
		args = append(args, nftAction)
		if rule.Comment != "" {
			cleanComment := strings.ReplaceAll(rule.Comment, `"`, "")
			args = append(args, "comment", fmt.Sprintf("\"%s\"", cleanComment))
		}
	}

	_, err := execsafe.RunDirect(ctx, 15*time.Second, "/usr/sbin/nft", args...)
	return err
}
