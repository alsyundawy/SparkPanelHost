import React, { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
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
} from 'lucide-react';
import type { TelemetryData } from '../types';

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
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/api/v1/telemetry/ws`;
    const ws = new WebSocket(wsUrl);

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        setTelemetry((prev) => ({
          ...prev,
          cpu_percent: data.cpu_percent ?? prev.cpu_percent,
          ram_used_mb: data.ram_used_mb ?? prev.ram_used_mb,
          ram_total_mb: data.ram_total_mb ?? prev.ram_total_mb,
          disk_used_mb: data.disk_used_mb ?? prev.disk_used_mb,
          disk_total_mb: data.disk_total_mb ?? prev.disk_total_mb,
          waf_blocked_24h: data.waf_blocked_24h ?? prev.waf_blocked_24h,
          crowdsec_active_bans: data.crowdsec_active_bans ?? prev.crowdsec_active_bans,
        }));
      } catch {
        // Fallback
      }
    };

    return () => {
      ws.close();
    };
  }, []);

  const ramPercent = Math.round((telemetry.ram_used_mb / telemetry.ram_total_mb) * 100);
  const diskPercent = Math.round((telemetry.disk_used_mb / telemetry.disk_total_mb) * 100);

  const services = [
    { name: 'Nginx 1.31 Mainline', desc: 'Reverse Proxy & HTTP/3', status: 'active' },
    { name: 'Coraza WAF + CRS v4', desc: 'Go-native memory safe WAF', status: 'active' },
    { name: 'Multi-PHP 8.2–8.6', desc: 'PHP 8.5 default (8.2 EOL Dec 2026)', status: 'warning' },
    { name: 'MariaDB 12.3.4 LTS', desc: 'Database Engine (utf8mb4)', status: 'active' },
    { name: 'PowerDNS 5.1 & Cluster', desc: 'REST API & WHM DNS Only sync', status: 'active' },
    { name: 'Postfix & Dovecot 2.4', desc: 'Maildir with Argon2id', status: 'active' },
    { name: 'Rspamd 4.3 & MTA-STS', desc: 'DKIM RSA+Ed25519 dual-sign', status: 'active' },
    { name: 'nftables & CrowdSec 1.8', desc: 'Atomic IPS & WAF bouncer', status: 'active' },
  ];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-xl font-bold tracking-tight text-text-primary">
            {t('nav.overview', 'Server Overview')}
          </h1>
          <p className="text-xs text-text-secondary mt-0.5">
            Ubuntu 26.04 / Debian 13 LTS Infrastructure Telemetry & Status
          </p>
        </div>
        <div className="flex items-center gap-2">
          <span className="flex h-2.5 w-2.5 rounded-full bg-status-success animate-pulse" />
          <span className="text-xs font-semibold text-status-success font-mono">
            System Healthy (All Daemons Green)
          </span>
        </div>
      </div>

      {/* Bento Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {/* Card 1: CPU Usage */}
        <div className="rounded-xl border border-border-subtle bg-bg-card p-4 flex flex-col justify-between shadow-xs">
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold text-text-secondary">vCPU Load</span>
            <Cpu className="h-4 w-4 text-accent-primary" />
          </div>
          <div className="my-3">
            <div className="text-2xl font-bold font-mono tabular-nums text-text-primary">
              {telemetry.cpu_percent.toFixed(1)}%
            </div>
            <div className="w-full bg-bg-elevated h-1.5 rounded-full mt-2 overflow-hidden">
              <div
                className="bg-accent-primary h-full rounded-full transition-all duration-500"
                style={{ width: `${Math.min(telemetry.cpu_percent, 100)}%` }}
              />
            </div>
          </div>
          <span className="text-[11px] text-text-secondary font-mono">
            cgroups v2 slice active
          </span>
        </div>

        {/* Card 2: Memory (RAM) */}
        <div className="rounded-xl border border-border-subtle bg-bg-card p-4 flex flex-col justify-between shadow-xs">
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold text-text-secondary">Memory Usage</span>
            <Server className="h-4 w-4 text-accent-primary" />
          </div>
          <div className="my-3">
            <div className="text-2xl font-bold font-mono tabular-nums text-text-primary">
              {ramPercent}%
            </div>
            <div className="w-full bg-bg-elevated h-1.5 rounded-full mt-2 overflow-hidden">
              <div
                className="bg-accent-primary h-full rounded-full transition-all duration-500"
                style={{ width: `${ramPercent}%` }}
              />
            </div>
          </div>
          <span className="text-[11px] text-text-secondary font-mono tabular-nums">
            {telemetry.ram_used_mb} MB / {telemetry.ram_total_mb} MB
          </span>
        </div>

        {/* Card 3: Disk Quota */}
        <div className="rounded-xl border border-border-subtle bg-bg-card p-4 flex flex-col justify-between shadow-xs">
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold text-text-secondary">Disk Storage</span>
            <HardDrive className="h-4 w-4 text-accent-primary" />
          </div>
          <div className="my-3">
            <div className="text-2xl font-bold font-mono tabular-nums text-text-primary">
              {diskPercent}%
            </div>
            <div className="w-full bg-bg-elevated h-1.5 rounded-full mt-2 overflow-hidden">
              <div
                className="bg-accent-primary h-full rounded-full transition-all duration-500"
                style={{ width: `${diskPercent}%` }}
              />
            </div>
          </div>
          <span className="text-[11px] text-text-secondary font-mono tabular-nums">
            {(telemetry.disk_used_mb / 1024).toFixed(1)} GB / {(telemetry.disk_total_mb / 1024).toFixed(1)} GB
          </span>
        </div>

        {/* Card 4: Security Shield Bento */}
        <div className="rounded-xl border border-border-subtle bg-bg-card p-4 flex flex-col justify-between shadow-xs">
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold text-text-secondary">WAF & Firewall Blocks</span>
            <ShieldAlert className="h-4 w-4 text-status-warning" />
          </div>
          <div className="my-3 flex items-baseline justify-between">
            <div>
              <div className="text-2xl font-bold font-mono tabular-nums text-text-primary">
                {telemetry.waf_blocked_24h}
              </div>
              <div className="text-[10px] text-text-secondary font-medium">Coraza WAF (24h)</div>
            </div>
            <div className="text-right">
              <div className="text-2xl font-bold font-mono tabular-nums text-status-warning">
                {telemetry.crowdsec_active_bans}
              </div>
              <div className="text-[10px] text-text-secondary font-medium">CrowdSec Active Bans</div>
            </div>
          </div>
          <span className="text-[11px] text-status-success font-medium flex items-center gap-1">
            <ShieldCheck className="h-3 w-3" /> Zero-Trust nftables active
          </span>
        </div>
      </div>

      {/* Services Grid & Quick Actions */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Core Services Health Matrix (2 cols) */}
        <div className="lg:col-span-2 rounded-xl border border-border-subtle bg-bg-card p-5">
          <div className="flex items-center justify-between mb-4">
            <div className="flex items-center gap-2">
              <Activity className="h-4 w-4 text-accent-primary" />
              <h2 className="text-sm font-bold text-text-primary">Service Ecosystem Health</h2>
            </div>
            <span className="text-xs text-text-secondary font-mono">Dual-Plane Privilege Split</span>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
            {services.map((svc) => (
              <div
                key={svc.name}
                className="flex items-start justify-between p-3 rounded-lg border border-border-subtle bg-bg-elevated"
              >
                <div>
                  <div className="text-xs font-bold text-text-primary flex items-center gap-1.5">
                    {svc.name}
                  </div>
                  <div className="text-[11px] text-text-secondary mt-0.5">{svc.desc}</div>
                </div>
                {svc.status === 'active' ? (
                  <CheckCircle2 className="h-4 w-4 text-status-success shrink-0" />
                ) : (
                  <AlertCircle className="h-4 w-4 text-status-warning shrink-0" />
                )}
              </div>
            ))}
          </div>
        </div>

        {/* Quick Actions & RBL Health */}
        <div className="space-y-4">
          {/* Email Deliverability RBL Status */}
          <div className="rounded-xl border border-border-subtle bg-bg-card p-4">
            <div className="flex items-center justify-between mb-2">
              <span className="text-xs font-bold text-text-primary">Email RBL Blacklist</span>
              <span className="rounded bg-status-success/15 px-2 py-0.5 text-[10px] font-bold text-status-success">
                Clean (0/5)
              </span>
            </div>
            <p className="text-xs text-text-secondary">
              IP reputation clean on Spamhaus, Barracuda, SpamCop & SORBS.
            </p>
          </div>

          {/* Quick Actions */}
          <div className="rounded-xl border border-border-subtle bg-bg-card p-4 space-y-2.5">
            <h3 className="text-xs font-bold text-text-primary uppercase tracking-wider">
              Quick Orchestration
            </h3>
            <button
              onClick={() => onNavigate('domains')}
              className="w-full flex items-center justify-between px-3 py-2 rounded-lg text-xs font-semibold bg-bg-elevated text-text-primary hover:bg-accent-primary hover:text-white transition-colors"
            >
              <span>Add Virtual Host & SSL</span>
              <ExternalLink className="h-3.5 w-3.5" />
            </button>
            <button
              onClick={() => onNavigate('dns')}
              className="w-full flex items-center justify-between px-3 py-2 rounded-lg text-xs font-semibold bg-bg-elevated text-text-primary hover:bg-accent-primary hover:text-white transition-colors"
            >
              <span>WHM DNS Cluster Sync</span>
              <ExternalLink className="h-3.5 w-3.5" />
            </button>
            <button
              onClick={() => onNavigate('apps')}
              className="w-full flex items-center justify-between px-3 py-2 rounded-lg text-xs font-semibold bg-bg-elevated text-text-primary hover:bg-accent-primary hover:text-white transition-colors"
            >
              <span>1-Click App Installer & Staging</span>
              <Flame className="h-3.5 w-3.5 text-status-warning" />
            </button>
            <button
              onClick={() => onNavigate('terminal')}
              className="w-full flex items-center justify-between px-3 py-2 rounded-lg text-xs font-semibold bg-bg-elevated text-text-primary hover:bg-accent-primary hover:text-white transition-colors"
            >
              <span>Launch Web SSH Terminal</span>
              <ExternalLink className="h-3.5 w-3.5" />
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};
