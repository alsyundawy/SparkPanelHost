import React, { useState } from "react";
import { useTranslation } from "react-i18next";
import { ShieldCheck, ShieldAlert, Flame, Radio } from "lucide-react";

interface CrowdSecDecision {
  id: number;
  ip: string;
  scenario: string;
  duration: string;
  country: string;
  source: string;
}

export const SecurityPage: React.FC = () => {
  const { t } = useTranslation();

  const [wafMode, setWafMode] = useState<"Blocking" | "DetectionOnly">(
    "Blocking",
  );
  const [crsVersion] = useState("OWASP CRS v4.30.0");
  const [activeProfiles, setActiveProfiles] = useState<string[]>([
    "wordpress",
    "nextcloud",
    "laravel",
  ]);

  const [decisions, setDecisions] = useState<CrowdSecDecision[]>([
    {
      id: 1,
      ip: "198.51.100.42",
      scenario: "crowdsecurity/http-crawl-non_statics",
      duration: "3h 45m",
      country: "NL",
      source: "Coraza WAF Bouncer",
    },
    {
      id: 2,
      ip: "203.0.113.195",
      scenario: "crowdsecurity/ssh-bf",
      duration: "23h 12m",
      country: "RU",
      source: "nftables bouncer",
    },
    {
      id: 3,
      ip: "192.0.2.88",
      scenario: "crowdsecurity/http-probing",
      duration: "1h 10m",
      country: "US",
      source: "Coraza WAF Bouncer",
    },
    {
      id: 4,
      ip: "198.51.100.77",
      scenario: "crowdsecurity/mariadb-bf",
      duration: "12h 00m",
      country: "CN",
      source: "nftables bouncer",
    },
  ]);

  const toggleProfile = (profile: string) => {
    setActiveProfiles((prev) =>
      prev.includes(profile)
        ? prev.filter((p) => p !== profile)
        : [...prev, profile],
    );
  };

  const handleUnban = (id: number) => {
    setDecisions((prev) => prev.filter((d) => d.id !== id));
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-xl font-bold tracking-tight text-text-primary flex items-center gap-2">
            <ShieldCheck className="h-5 w-5 text-status-success" />
            {t("nav.security", "Coraza WAF & CrowdSec Firewall")}
          </h1>
          <p className="text-xs text-text-secondary mt-0.5">
            Zero-Trust Network Defense, OWASP CRS Rulesets & Behavioral
            Intrusion Prevention
          </p>
        </div>
        <div className="flex items-center gap-2">
          <span className="flex h-2.5 w-2.5 rounded-full bg-status-success animate-pulse" />
          <span className="text-xs font-semibold text-status-success font-mono">
            nftables + Coraza Enforcing
          </span>
        </div>
      </div>

      {/* Top Bento Security Status Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {/* Card 1: Coraza Engine */}
        <div className="rounded-xl border border-border-subtle bg-bg-card p-4 shadow-xs flex flex-col justify-between">
          <div className="flex items-center justify-between">
            <span className="text-xs font-bold text-text-primary">
              Coraza Go-Native WAF
            </span>
            <span className="rounded bg-status-success/15 px-2 py-0.5 text-[10px] font-bold text-status-success font-mono">
              Memory-Safe
            </span>
          </div>
          <div className="my-3">
            <div className="text-xs text-text-secondary">
              Current Engine Mode:
            </div>
            <div className="text-lg font-bold text-accent-primary font-mono mt-0.5">
              {wafMode} Mode
            </div>
            <div className="text-[11px] text-text-secondary mt-1">
              Ruleset:{" "}
              <span className="font-semibold text-text-primary">
                {crsVersion}
              </span>
            </div>
          </div>
          <div className="flex items-center gap-2 pt-2 border-t border-border-subtle">
            <button
              onClick={() => setWafMode("Blocking")}
              className={`px-2.5 py-1 rounded text-[11px] font-semibold transition-colors ${
                wafMode === "Blocking"
                  ? "bg-status-success text-white"
                  : "bg-bg-elevated text-text-secondary hover:text-text-primary"
              }`}
            >
              Enforce Blocking
            </button>
            <button
              onClick={() => setWafMode("DetectionOnly")}
              className={`px-2.5 py-1 rounded text-[11px] font-semibold transition-colors ${
                wafMode === "DetectionOnly"
                  ? "bg-status-warning text-white"
                  : "bg-bg-elevated text-text-secondary hover:text-text-primary"
              }`}
            >
              Detection Only
            </button>
          </div>
        </div>

        {/* Card 2: CrowdSec Threat Intelligence */}
        <div className="rounded-xl border border-border-subtle bg-bg-card p-4 shadow-xs flex flex-col justify-between">
          <div className="flex items-center justify-between">
            <span className="text-xs font-bold text-text-primary">
              CrowdSec 1.8.2 Engine
            </span>
            <span className="rounded bg-accent-primary/15 px-2 py-0.5 text-[10px] font-bold text-accent-primary font-mono">
              LAPI Connected
            </span>
          </div>
          <div className="my-3">
            <div className="text-2xl font-bold font-mono text-text-primary">
              {decisions.length} Active Bans
            </div>
            <div className="text-[11px] text-text-secondary mt-1">
              Community Consensus:{" "}
              <span className="text-status-success font-semibold">Synced</span>
            </div>
          </div>
          <div className="text-[10px] text-text-secondary font-mono border-t border-border-subtle pt-2 flex items-center gap-1">
            <Radio className="h-3 w-3 text-status-success" />
            Bouncers: nftables + Coraza AppSec
          </div>
        </div>

        {/* Card 3: Dual-Plane Kernel Hardening */}
        <div className="rounded-xl border border-border-subtle bg-bg-card p-4 shadow-xs flex flex-col justify-between">
          <div className="flex items-center justify-between">
            <span className="text-xs font-bold text-text-primary">
              Linux Kernel Hardening
            </span>
            <span className="rounded bg-status-success/15 px-2 py-0.5 text-[10px] font-bold text-status-success font-mono">
              CIS L2 Compliant
            </span>
          </div>
          <div className="my-3 space-y-1 text-xs">
            <div className="flex justify-between">
              <span className="text-text-secondary">nftables Table:</span>
              <span className="font-mono font-semibold text-text-primary">
                inet aetherpanel
              </span>
            </div>
            <div className="flex justify-between">
              <span className="text-text-secondary">Agent Socket:</span>
              <span className="font-mono font-semibold text-text-primary">
                SO_PEERCRED 0660
              </span>
            </div>
            <div className="flex justify-between">
              <span className="text-text-secondary">Symlink Immunity:</span>
              <span className="font-mono font-semibold text-status-success">
                Openat2 RESOLVE_BENEATH
              </span>
            </div>
          </div>
          <div className="text-[10px] text-text-secondary font-mono border-t border-border-subtle pt-2">
            Zero-Shell execution verified
          </div>
        </div>
      </div>

      {/* Main Content: CMS Profiles & Live Decisions Table */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Left: CMS Exclusion Profiles */}
        <div className="rounded-xl border border-border-subtle bg-bg-card p-5 space-y-4">
          <div className="flex items-center gap-2">
            <Flame className="h-4 w-4 text-status-warning" />
            <h2 className="text-sm font-bold text-text-primary">
              OWASP CRS Per-CMS Profiles
            </h2>
          </div>
          <p className="text-xs text-text-secondary leading-relaxed">
            Eliminates false positives by enabling curated exclusion rule tags
            for modern content management systems and frameworks.
          </p>

          <div className="space-y-2.5 pt-2">
            {[
              {
                id: "wordpress",
                name: "WordPress & WooCommerce",
                tag: "crs_rule_wordpress",
              },
              {
                id: "nextcloud",
                name: "Nextcloud Hub",
                tag: "crs_rule_nextcloud",
              },
              {
                id: "laravel",
                name: "Laravel & Livewire",
                tag: "crs_rule_laravel",
              },
              { id: "drupal", name: "Drupal CMS", tag: "crs_rule_drupal" },
            ].map((prof) => {
              const isEnabled = activeProfiles.includes(prof.id);
              return (
                <div
                  key={prof.id}
                  onClick={() => toggleProfile(prof.id)}
                  className={`flex items-center justify-between p-3 rounded-lg border cursor-pointer transition-colors ${
                    isEnabled
                      ? "border-accent-primary/40 bg-accent-primary/5"
                      : "border-border-subtle bg-bg-elevated opacity-70"
                  }`}
                >
                  <div>
                    <div className="text-xs font-bold text-text-primary">
                      {prof.name}
                    </div>
                    <div className="text-[10px] text-text-secondary font-mono mt-0.5">
                      {prof.tag}
                    </div>
                  </div>
                  <span
                    className={`h-4 w-4 rounded-full flex items-center justify-center text-[10px] font-bold ${
                      isEnabled
                        ? "bg-status-success text-white"
                        : "bg-bg-canvas text-text-secondary"
                    }`}
                  >
                    {isEnabled ? "✓" : ""}
                  </span>
                </div>
              );
            })}
          </div>
        </div>

        {/* Right: Active Bans & Decisions Table */}
        <div className="lg:col-span-2 rounded-xl border border-border-subtle bg-bg-card p-5 space-y-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <ShieldAlert className="h-4 w-4 text-status-danger" />
              <h2 className="text-sm font-bold text-text-primary">
                Active Intrusion Bans (CrowdSec)
              </h2>
            </div>
            <span className="text-xs text-text-secondary font-mono">
              Auto-sync every 60s
            </span>
          </div>

          <div className="overflow-x-auto">
            <table className="w-full text-left text-xs">
              <thead className="border-b border-border-subtle bg-bg-elevated text-text-secondary">
                <tr>
                  <th className="py-2.5 px-3 font-semibold">Attacker IP</th>
                  <th className="py-2.5 px-3 font-semibold">
                    Detection Scenario
                  </th>
                  <th className="py-2.5 px-3 font-semibold">Origin</th>
                  <th className="py-2.5 px-3 font-semibold">Duration</th>
                  <th className="py-2.5 px-3 font-semibold text-right">
                    Action
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border-subtle">
                {decisions.map((d) => (
                  <tr
                    key={d.id}
                    className="hover:bg-bg-elevated/50 transition-colors"
                  >
                    <td className="py-2.5 px-3 font-mono font-bold text-text-primary">
                      {d.ip}
                    </td>
                    <td className="py-2.5 px-3 text-text-secondary">
                      <span className="rounded bg-status-danger/10 px-1.5 py-0.5 text-[10px] font-mono text-status-danger">
                        {d.scenario}
                      </span>
                    </td>
                    <td className="py-2.5 px-3 font-mono text-[11px] text-text-secondary">
                      {d.country} • {d.source}
                    </td>
                    <td className="py-2.5 px-3 font-mono text-[11px] text-text-secondary">
                      {d.duration}
                    </td>
                    <td className="py-2.5 px-3 text-right">
                      <button
                        onClick={() => handleUnban(d.id)}
                        className="text-[11px] font-semibold text-accent-primary hover:underline"
                      >
                        Unban IP
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  );
};
