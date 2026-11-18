import React, { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import {
  Cpu,
  Server,
  HardDrive,
  ShieldAlert,
  ShieldCheck,
  Activity,
  CheckCircle2,
  AlertCircle,
  ExternalLink,
  Flame,
  Globe,
  Database,
  Terminal,
  Network,
  Clock,
  Layers,
  TrendingDown,
  Lock,
} from "lucide-react";
import type { TelemetryData } from "../types";

interface OverviewPageProps {
  onNavigate: (tab: string) => void;
}

export const OverviewPage: React.FC<OverviewPageProps> = ({ onNavigate }) => {
  const { t } = useTranslation();
  const [telemetry, setTelemetry] = useState<TelemetryData>({
    cpu_percent: 14.2,
    ram_used_mb: 2150,
    ram_total_mb: 8192,
    disk_used_mb: 18400,
    disk_total_mb: 102400,
    waf_blocked_24h: 318,
    crowdsec_active_bans: 24,
  });

  useEffect(() => {
    // Attempt WebSocket connection for live telemetry stream
    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    const wsUrl = `${protocol}//${window.location.host}/api/v1/telemetry/ws`;
    const ws = new WebSocket(wsUrl);

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        setTelemetry((prev) => ({
          ...prev,
          cpu_percent:
            typeof data.cpu_percent === "number"
              ? data.cpu_percent
              : prev.cpu_percent,
          ram_used_mb:
            typeof data.ram_used_mb === "number"
              ? data.ram_used_mb
              : prev.ram_used_mb,
          ram_total_mb:
            typeof data.ram_total_mb === "number"
              ? data.ram_total_mb
              : prev.ram_total_mb,
          disk_used_mb:
            typeof data.disk_used_mb === "number"
              ? data.disk_used_mb
              : prev.disk_used_mb,
          disk_total_mb:
            typeof data.disk_total_mb === "number"
              ? data.disk_total_mb
              : prev.disk_total_mb,
          waf_blocked_24h:
            typeof data.waf_blocked_24h === "number"
              ? data.waf_blocked_24h
              : prev.waf_blocked_24h,
          crowdsec_active_bans:
            typeof data.crowdsec_active_bans === "number"
              ? data.crowdsec_active_bans
              : prev.crowdsec_active_bans,
        }));
      } catch {
        // Fallback default telemetry preserved
      }
    };

    return () => {
      ws.close();
    };
  }, []);

  const ramPercent = Math.min(
    Math.round((telemetry.ram_used_mb / telemetry.ram_total_mb) * 100),
    100,
  );
  const diskPercent = Math.min(
    Math.round((telemetry.disk_used_mb / telemetry.disk_total_mb) * 100),
    100,
  );

  const services = [
    {
      name: "Nginx 1.31 Mainline",
      desc: "Reverse Proxy & HTTP/3 (:80, :443)",
      status: "active",
      port: "80/443",
    },
    {
      name: "Coraza WAF + CRS v4",
      desc: "Go-native memory-safe WAF",
      status: "active",
      port: "In-process",
    },
    {
      name: "Multi-PHP 8.2–8.6",
      desc: "PHP 8.5 default (8.2 EOL Dec 2026)",
      status: "warning",
      port: "unix:fpm",
    },
    {
      name: "MariaDB 12.3.4 LTS",
      desc: "Database Engine (utf8mb4)",
      status: "active",
      port: "3306",
    },
    {
      name: "PowerDNS 5.1 & Cluster",
      desc: "REST API & WHM DNS Only sync",
      status: "active",
      port: "53/UDP",
    },
    {
      name: "Postfix & Dovecot 2.4",
      desc: "Maildir with Argon2id",
      status: "active",
      port: "25/587/993",
    },
    {
      name: "Rspamd 4.3 & MTA-STS",
      desc: "DKIM RSA+Ed25519 dual-sign",
      status: "active",
      port: "11334",
    },
    {
      name: "nftables & CrowdSec 1.8",
      desc: "Atomic IPS & WAF bouncer",
      status: "active",
      port: "inet filter",
    },
  ];

  const rblProviders = [
    { name: "Spamhaus SBL/XBL/PBL", status: "Clean (0 Listings)" },
    { name: "Barracuda Reputation Network", status: "Clean (0 Listings)" },
    { name: "SpamCop BL", status: "Clean (0 Listings)" },
    { name: "SORBS Aggregate", status: "Clean (0 Listings)" },
    { name: "UCEPROTECT Network", status: "Clean (0 Listings)" },
  ];

  const recentEvents = [
    {
      time: "12m ago",
      title: "Coraza WAF blocked SQLi attack",
      ip: "198.51.100.42",
      badge: "CRS 942100",
      type: "warning",
    },
    {
      time: "34m ago",
      title: "Let's Encrypt SSL Auto-Renewed",
      ip: "*.sparkpanel.net",
      badge: "DNS-01 ACME",
      type: "success",
    },
    {
      time: "1h 15m ago",
      title: "MariaDB physical snapshot verified",
      ip: "/var/panel/backups",
      badge: "SHA-256 OK",
      type: "info",
    },
    {
      time: "2h 40m ago",
      title: "CrowdSec banned IP for SSH brute-force",
      ip: "203.0.113.195",
      badge: "24h nftables ban",
      type: "warning",
    },
  ];

  return (
    <div className="space-y-6">
      {/* Node Identity & Health Strip (AdminLTE header pattern) */}
      <div className="rounded-xl border border-border-subtle bg-bg-card p-4 shadow-xs">
        <div className="flex flex-col lg:flex-row lg:items-center lg:justify-between gap-4">
          <div className="flex items-start gap-3">
            <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-accent-primary/10 text-accent-primary border border-accent-primary/20 shrink-0">
              <Server className="h-5 w-5" />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <h1 className="text-base font-bold tracking-tight text-text-primary">
                  {t("nav.overview", "Server Overview")} —
                  node-01.sparkpanel.net
                </h1>
                <span className="rounded bg-status-success/15 px-2 py-0.2 text-[10px] font-bold text-status-success font-mono border border-status-success/30">
                  Primary Node
                </span>
              </div>
              <p className="text-xs text-text-secondary mt-0.5">
                Ubuntu 26.04 / Debian 13 LTS • Linux 6.8.0-45-generic (x86_64) •
                Dual-Plane Privilege Split
              </p>
            </div>
          </div>

          <div className="flex flex-wrap items-center gap-3 sm:gap-6 text-xs text-text-secondary border-t lg:border-t-0 border-border-subtle pt-3 lg:pt-0">
            <div className="flex items-center gap-1.5 font-mono">
              <Clock className="h-3.5 w-3.5 text-accent-primary" />
              <span>
                Uptime:{" "}
                <strong className="text-text-primary">14d 6h 32m</strong>
              </span>
            </div>
            <div className="flex items-center gap-1.5 font-mono">
              <Layers className="h-3.5 w-3.5 text-status-success" />
              <span>
                Load:{" "}
                <strong className="text-text-primary">0.14, 0.22, 0.18</strong>
              </span>
            </div>
            <div className="flex items-center gap-1.5">
              <span className="h-2 w-2 rounded-full bg-status-success animate-pulse" />
              <span className="font-semibold text-status-success font-mono">
                8/8 Daemons Active
              </span>
            </div>
          </div>
        </div>
      </div>

      {/* KPI Metric Small Boxes & Bento Grid (AdminLTE v4 + Modern Bento) */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {/* Card 1: vCPU Load */}
        <div className="rounded-xl border border-border-subtle bg-bg-card p-4.5 flex flex-col justify-between shadow-xs hover:border-accent-primary/40 transition-colors">
          <div className="flex items-center justify-between">
            <span className="text-xs font-bold text-text-secondary">
              vCPU Load
            </span>
            <div className="flex items-center gap-1 text-[10px] font-semibold text-status-success bg-status-success/10 px-1.5 py-0.5 rounded">
              <TrendingDown className="h-3 w-3" /> -1.4%
            </div>
          </div>
          <div className="my-3">
            <div className="flex items-baseline justify-between">
              <div className="text-2xl font-bold font-mono tabular-nums text-text-primary">
                {telemetry.cpu_percent.toFixed(1)}%
              </div>
              <Cpu className="h-5 w-5 text-accent-primary/60" />
            </div>
            <div className="w-full bg-bg-elevated h-2 rounded-full mt-2 overflow-hidden border border-border-subtle/50">
              <div
                className={`h-full rounded-full transition-all duration-500 ${
                  telemetry.cpu_percent > 80
                    ? "bg-status-danger"
                    : telemetry.cpu_percent > 60
                      ? "bg-status-warning"
                      : "bg-accent-primary"
                }`}
                style={{ width: `${Math.min(telemetry.cpu_percent, 100)}%` }}
              />
            </div>
          </div>
          <div className="flex items-center justify-between text-[11px] text-text-secondary font-mono">
            <span>cgroups v2 slice active</span>
            <span>4 Cores @ 3.4GHz</span>
          </div>
        </div>

        {/* Card 2: Memory (RAM) */}
        <div className="rounded-xl border border-border-subtle bg-bg-card p-4.5 flex flex-col justify-between shadow-xs hover:border-accent-primary/40 transition-colors">
          <div className="flex items-center justify-between">
            <span className="text-xs font-bold text-text-secondary">
              Memory Allocation
            </span>
            <Server className="h-4 w-4 text-accent-primary" />
          </div>
          <div className="my-3">
            <div className="flex items-baseline justify-between">
              <div className="text-2xl font-bold font-mono tabular-nums text-text-primary">
                {ramPercent}%
              </div>
              <span className="text-xs font-mono text-text-secondary tabular-nums">
                {(telemetry.ram_used_mb / 1024).toFixed(1)} /{" "}
                {(telemetry.ram_total_mb / 1024).toFixed(0)} GB
              </span>
            </div>
            <div className="w-full bg-bg-elevated h-2 rounded-full mt-2 overflow-hidden border border-border-subtle/50">
              <div
                className={`h-full rounded-full transition-all duration-500 ${
                  ramPercent > 85
                    ? "bg-status-danger"
                    : ramPercent > 70
                      ? "bg-status-warning"
                      : "bg-accent-primary"
                }`}
                style={{ width: `${ramPercent}%` }}
              />
            </div>
          </div>
          <div className="flex items-center justify-between text-[11px] text-text-secondary font-mono tabular-nums">
            <span>Active: 1,420 MB</span>
            <span>Cache: 730 MB</span>
          </div>
        </div>

        {/* Card 3: Disk Quota */}
        <div className="rounded-xl border border-border-subtle bg-bg-card p-4.5 flex flex-col justify-between shadow-xs hover:border-accent-primary/40 transition-colors">
          <div className="flex items-center justify-between">
            <span className="text-xs font-bold text-text-secondary">
              NVMe Storage Quota
            </span>
            <HardDrive className="h-4 w-4 text-accent-primary" />
          </div>
          <div className="my-3">
            <div className="flex items-baseline justify-between">
              <div className="text-2xl font-bold font-mono tabular-nums text-text-primary">
                {diskPercent}%
              </div>
              <span className="text-xs font-mono text-text-secondary tabular-nums">
                {(telemetry.disk_used_mb / 1024).toFixed(1)} /{" "}
                {(telemetry.disk_total_mb / 1024).toFixed(0)} GB
              </span>
            </div>
            <div className="w-full bg-bg-elevated h-2 rounded-full mt-2 overflow-hidden border border-border-subtle/50">
              <div
                className={`h-full rounded-full transition-all duration-500 ${
                  diskPercent > 85
                    ? "bg-status-danger"
                    : diskPercent > 70
                      ? "bg-status-warning"
                      : "bg-accent-primary"
                }`}
                style={{ width: `${diskPercent}%` }}
              />
            </div>
          </div>
          <div className="flex items-center justify-between text-[11px] text-text-secondary font-mono tabular-nums">
            <span>/var/panel ext4</span>
            <span className="text-status-success font-semibold">Quota OK</span>
          </div>
        </div>

        {/* Card 4: Security Shield Bento */}
        <div className="rounded-xl border border-border-subtle bg-bg-card p-4.5 flex flex-col justify-between shadow-xs hover:border-accent-primary/40 transition-colors">
          <div className="flex items-center justify-between">
            <span className="text-xs font-bold text-text-secondary">
              Coraza WAF & Firewall
            </span>
            <ShieldAlert className="h-4 w-4 text-status-warning" />
          </div>
          <div className="my-3 flex items-baseline justify-between">
            <div>
              <div className="text-2xl font-bold font-mono tabular-nums text-text-primary">
                {telemetry.waf_blocked_24h}
              </div>
              <div className="text-[10px] text-text-secondary font-medium">
                Coraza WAF (24h)
              </div>
            </div>
            <div className="text-right">
              <div className="text-2xl font-bold font-mono tabular-nums text-status-warning">
                {telemetry.crowdsec_active_bans}
              </div>
              <div className="text-[10px] text-text-secondary font-medium">
                CrowdSec Bans
              </div>
            </div>
          </div>
          <div className="flex items-center justify-between text-[11px] text-status-success font-medium">
            <span className="flex items-center gap-1">
              <ShieldCheck className="h-3.5 w-3.5" /> Zero-Trust Active
            </span>
            <span className="text-[10px] text-text-secondary font-mono">
              CRS v4.30
            </span>
          </div>
        </div>
      </div>

      {/* Quick Action Ribbon (AdminLTE Toolstrip) */}
      <div className="rounded-xl border border-border-subtle bg-bg-card p-3 shadow-xs">
        <div className="flex items-center justify-between flex-wrap gap-2">
          <span className="text-xs font-bold uppercase tracking-wider text-text-secondary px-2">
            Quick Actions:
          </span>
          <div className="flex flex-wrap items-center gap-2 flex-1 justify-end">
            <button
              onClick={() => onNavigate("domains")}
              className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-semibold bg-bg-elevated text-text-primary hover:bg-accent-primary hover:text-white transition-colors border border-border-subtle"
            >
              <Globe className="h-3.5 w-3.5 text-accent-primary" />
              <span>+ Add Domain</span>
            </button>
            <button
              onClick={() => onNavigate("databases")}
              className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-semibold bg-bg-elevated text-text-primary hover:bg-accent-primary hover:text-white transition-colors border border-border-subtle"
            >
              <Database className="h-3.5 w-3.5 text-accent-primary" />
              <span>+ New Database</span>
            </button>
            <button
              onClick={() => onNavigate("dns")}
              className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-semibold bg-bg-elevated text-text-primary hover:bg-accent-primary hover:text-white transition-colors border border-border-subtle"
            >
              <Network className="h-3.5 w-3.5 text-accent-primary" />
              <span>DNS Cluster Sync</span>
            </button>
            <button
              onClick={() => onNavigate("apps")}
              className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-semibold bg-bg-elevated text-text-primary hover:bg-accent-primary hover:text-white transition-colors border border-border-subtle"
            >
              <Flame className="h-3.5 w-3.5 text-status-warning" />
              <span>1-Click App Installer</span>
            </button>
            <button
              onClick={() => onNavigate("terminal")}
              className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-semibold bg-bg-elevated text-text-primary hover:bg-accent-primary hover:text-white transition-colors border border-border-subtle"
            >
              <Terminal className="h-3.5 w-3.5 text-accent-primary" />
              <span>Launch Terminal</span>
            </button>
          </div>
        </div>
      </div>

      {/* Main Grid: Service Health Matrix (Left 2/3) + Security/Deliverability Feed (Right 1/3) */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Core Services Health Matrix (2 cols) */}
        <div className="lg:col-span-2 rounded-xl border border-border-subtle bg-bg-card p-5 space-y-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Activity className="h-4 w-4 text-accent-primary" />
              <h2 className="text-sm font-bold text-text-primary">
                Daemon & Service Health Matrix
              </h2>
            </div>
            <span className="text-xs text-text-secondary font-mono">
              Dual-Plane Privilege Split
            </span>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
            {services.map((svc) => (
              <div
                key={svc.name}
                className="flex items-start justify-between p-3.5 rounded-xl border border-border-subtle bg-bg-elevated/60 hover:bg-bg-elevated transition-colors"
              >
                <div>
                  <div className="text-xs font-bold text-text-primary flex items-center gap-1.5">
                    {svc.name}
                  </div>
                  <div className="text-[11px] text-text-secondary mt-0.5">
                    {svc.desc}
                  </div>
                  <div className="text-[10px] text-text-secondary/70 font-mono mt-1">
                    Socket / Port: {svc.port}
                  </div>
                </div>
                <div className="flex flex-col items-end gap-1 shrink-0">
                  {svc.status === "active" ? (
                    <span className="flex items-center gap-1 text-[11px] font-semibold text-status-success">
                      <CheckCircle2 className="h-3.5 w-3.5" /> Active
                    </span>
                  ) : (
                    <span className="flex items-center gap-1 text-[11px] font-semibold text-status-warning">
                      <AlertCircle className="h-3.5 w-3.5" /> Warning
                    </span>
                  )}
                  <span className="text-[9px] font-mono text-text-secondary/70">
                    systemd
                  </span>
                </div>
              </div>
            ))}
          </div>

          <div className="p-3 rounded-lg border border-border-subtle bg-bg-elevated/40 flex items-center justify-between text-xs text-text-secondary">
            <span className="flex items-center gap-1.5">
              <Lock className="h-3.5 w-3.5 text-accent-primary" />
              Agent Unix Domain Socket:{" "}
              <strong className="font-mono text-text-primary">
                /run/panel/agent.sock (0660)
              </strong>
            </span>
            <button
              onClick={() => onNavigate("config-editor")}
              className="text-xs font-semibold text-accent-primary hover:underline flex items-center gap-1"
            >
              <span>Edit Configurations</span>
              <ExternalLink className="h-3 w-3" />
            </button>
          </div>
        </div>

        {/* Right Column: Email Deliverability RBL Health & Live Security Feed */}
        <div className="space-y-6">
          {/* Email Deliverability RBL Status */}
          <div className="rounded-xl border border-border-subtle bg-bg-card p-5 space-y-3">
            <div className="flex items-center justify-between">
              <span className="text-xs font-bold text-text-primary">
                Real-Time Email RBL Status
              </span>
              <span className="rounded bg-status-success/15 px-2 py-0.5 text-[10px] font-bold text-status-success border border-status-success/30">
                100% Clean
              </span>
            </div>
            <p className="text-xs text-text-secondary">
              Authoritative real-time anti-spam blackhole list reputation
              monitoring:
            </p>

            <div className="space-y-1.5 pt-1">
              {rblProviders.map((rbl) => (
                <div
                  key={rbl.name}
                  className="flex items-center justify-between text-[11px] py-1 border-b border-border-subtle/50 last:border-0"
                >
                  <span className="text-text-secondary truncate max-w-[170px]">
                    {rbl.name}
                  </span>
                  <span className="font-mono font-semibold text-status-success flex items-center gap-1">
                    <CheckCircle2 className="h-3 w-3" /> Clean
                  </span>
                </div>
              ))}
            </div>

            <div className="pt-2 border-t border-border-subtle flex items-center justify-between text-[11px]">
              <span className="text-text-secondary">DKIM / DMARC:</span>
              <span className="font-mono font-semibold text-status-success">
                RSA + Ed25519 Active
              </span>
            </div>
          </div>

          {/* Live Security Threat & Event Feed (AdminLTE inspiration) */}
          <div className="rounded-xl border border-border-subtle bg-bg-card p-5 space-y-3">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-1.5">
                <ShieldCheck className="h-4 w-4 text-accent-primary" />
                <h3 className="text-xs font-bold text-text-primary">
                  Recent Security Events
                </h3>
              </div>
              <button
                onClick={() => onNavigate("audit-logs")}
                className="text-[10px] font-semibold text-accent-primary hover:underline"
              >
                View Audit Log
              </button>
            </div>

            <div className="space-y-2.5 pt-1">
              {recentEvents.map((ev, idx) => (
                <div
                  key={idx}
                  className="p-2.5 rounded-lg border border-border-subtle bg-bg-elevated/40 text-xs space-y-1"
                >
                  <div className="flex items-center justify-between gap-2">
                    <span className="font-semibold text-text-primary truncate">
                      {ev.title}
                    </span>
                    <span className="text-[10px] font-mono text-text-secondary/70 shrink-0">
                      {ev.time}
                    </span>
                  </div>
                  <div className="flex items-center justify-between text-[11px] text-text-secondary">
                    <span className="font-mono text-[10px] truncate max-w-[150px]">
                      {ev.ip}
                    </span>
                    <span
                      className={`px-1.5 py-0.2 rounded text-[9px] font-mono font-bold ${
                        ev.type === "warning"
                          ? "bg-status-warning/15 text-status-warning"
                          : ev.type === "success"
                            ? "bg-status-success/15 text-status-success"
                            : "bg-accent-primary/15 text-accent-primary"
                      }`}
                    >
                      {ev.badge}
                    </span>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
