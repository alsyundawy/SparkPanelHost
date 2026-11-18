import React, { useState } from "react";
import { useTranslation } from "react-i18next";
import {
  Package,
  Plus,
  Globe,
  Database,
  Mail,
  Terminal,
  Check,
  X,
  Sliders,
} from "lucide-react";
import type { HostingPackage } from "../types";

export const PackagesPage: React.FC = () => {
  const { t } = useTranslation();

  const [packages, setPackages] = useState<HostingPackage[]>([
    {
      id: "pkg-starter",
      name: "Starter Cloud",
      quota_mb: 10240, // 10 GB
      bandwidth_mb: 102400, // 100 GB
      max_domains: 3,
      max_databases: 5,
      max_mailboxes: 10,
      shell_access: false,
      php_versions: ["8.4", "8.5"],
      features: ["cron", "file_manager", "auto_ssl", "coraza_waf"],
      created_at: "2026-01-15T08:00:00Z",
    },
    {
      id: "pkg-pro",
      name: "Professional Business",
      quota_mb: 51200, // 50 GB
      bandwidth_mb: 512000, // 500 GB
      max_domains: 15,
      max_databases: 25,
      max_mailboxes: 50,
      shell_access: true,
      php_versions: ["8.3", "8.4", "8.5", "8.6"],
      features: [
        "cron",
        "file_manager",
        "auto_ssl",
        "coraza_waf",
        "dns_cluster",
        "git_deploy",
        "staging",
      ],
      created_at: "2026-02-01T10:00:00Z",
    },
    {
      id: "pkg-enterprise",
      name: "Enterprise Dedicated",
      quota_mb: 204800, // 200 GB
      bandwidth_mb: 2048000, // 2 TB
      max_domains: 100,
      max_databases: 200,
      max_mailboxes: 500,
      shell_access: true,
      php_versions: ["8.2", "8.3", "8.4", "8.5", "8.6"],
      features: [
        "cron",
        "file_manager",
        "auto_ssl",
        "coraza_waf",
        "dns_cluster",
        "git_deploy",
        "staging",
        "web_terminal",
        "restic_backup",
      ],
      created_at: "2026-02-10T12:00:00Z",
    },
  ]);

  const [isModalOpen, setIsModalOpen] = useState(false);
  const [newPkgName, setNewPkgName] = useState("");
  const [newQuotaGb, setNewQuotaGb] = useState("25");
  const [newMaxDomains, setNewMaxDomains] = useState("10");
  const [newShellAccess, setNewShellAccess] = useState(false);

  const handleCreatePackage = (e: React.FormEvent) => {
    e.preventDefault();
    if (!newPkgName) return;

    const newPkg: HostingPackage = {
      id: `pkg-${Date.now()}`,
      name: newPkgName,
      quota_mb: parseInt(newQuotaGb, 10) * 1024,
      bandwidth_mb: parseInt(newQuotaGb, 10) * 10240,
      max_domains: parseInt(newMaxDomains, 10),
      max_databases: parseInt(newMaxDomains, 10) * 2,
      max_mailboxes: parseInt(newMaxDomains, 10) * 5,
      shell_access: newShellAccess,
      php_versions: ["8.4", "8.5", "8.6"],
      features: ["cron", "file_manager", "auto_ssl", "coraza_waf", "staging"],
      created_at: new Date().toISOString(),
    };

    setPackages([...packages, newPkg]);
    setIsModalOpen(false);
    setNewPkgName("");
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-xl font-bold tracking-tight text-text-primary flex items-center gap-2">
            <Package className="h-5 w-5 text-accent-primary" />
            {t("nav.packages", "Packages & Feature Lists")}
          </h1>
          <p className="text-xs text-text-secondary mt-0.5">
            cgroups v2 Resource Slicing, Quota Allocation & WHM Feature Gating
            Profiles
          </p>
        </div>
        <button
          onClick={() => setIsModalOpen(true)}
          className="flex items-center gap-2 px-3 py-2 rounded-lg bg-accent-primary text-white text-xs font-semibold hover:bg-accent-hover transition-colors shadow-sm self-start sm:self-auto"
        >
          <Plus className="h-4 w-4" />
          <span>Create Hosting Package</span>
        </button>
      </div>

      {/* Package Grid (AdminLTE / Bento style cards) */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-5">
        {packages.map((pkg) => (
          <div
            key={pkg.id}
            className="rounded-xl border border-border-subtle bg-bg-card p-5 flex flex-col justify-between shadow-xs hover:border-accent-primary/40 transition-colors"
          >
            <div>
              <div className="flex items-center justify-between">
                <span className="text-sm font-bold text-text-primary">
                  {pkg.name}
                </span>
                <span className="rounded bg-accent-primary/10 px-2 py-0.5 text-[10px] font-bold text-accent-primary font-mono">
                  {(pkg.quota_mb / 1024).toFixed(0)} GB Quota
                </span>
              </div>

              <div className="mt-4 space-y-2.5 text-xs text-text-secondary border-t border-border-subtle pt-3">
                <div className="flex items-center justify-between">
                  <span className="flex items-center gap-2">
                    <Globe className="h-3.5 w-3.5 text-text-secondary/80" /> Max
                    Domains
                  </span>
                  <span className="font-mono font-semibold text-text-primary">
                    {pkg.max_domains}
                  </span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="flex items-center gap-2">
                    <Database className="h-3.5 w-3.5 text-text-secondary/80" />{" "}
                    Databases
                  </span>
                  <span className="font-mono font-semibold text-text-primary">
                    {pkg.max_databases}
                  </span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="flex items-center gap-2">
                    <Mail className="h-3.5 w-3.5 text-text-secondary/80" />{" "}
                    Mailboxes
                  </span>
                  <span className="font-mono font-semibold text-text-primary">
                    {pkg.max_mailboxes}
                  </span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="flex items-center gap-2">
                    <Terminal className="h-3.5 w-3.5 text-text-secondary/80" />{" "}
                    SSH Shell Access
                  </span>
                  {pkg.shell_access ? (
                    <span className="text-[11px] font-semibold text-status-success flex items-center gap-1">
                      <Check className="h-3.5 w-3.5" /> Allowed
                    </span>
                  ) : (
                    <span className="text-[11px] font-semibold text-text-secondary/60 flex items-center gap-1">
                      <X className="h-3.5 w-3.5" /> Restricted
                    </span>
                  )}
                </div>
              </div>

              <div className="mt-4 pt-3 border-t border-border-subtle">
                <span className="text-[10px] font-bold uppercase tracking-wider text-text-secondary/70">
                  PHP Multi-Runtime
                </span>
                <div className="flex flex-wrap gap-1.5 mt-1.5">
                  {pkg.php_versions.map((v) => (
                    <span
                      key={v}
                      className="rounded bg-bg-elevated px-1.5 py-0.5 text-[10px] font-mono font-semibold text-text-primary border border-border-subtle"
                    >
                      PHP {v}
                    </span>
                  ))}
                </div>
              </div>

              <div className="mt-4 pt-3 border-t border-border-subtle">
                <span className="text-[10px] font-bold uppercase tracking-wider text-text-secondary/70">
                  Active Feature Gates
                </span>
                <div className="flex flex-wrap gap-1 mt-1.5">
                  {pkg.features.map((feat) => (
                    <span
                      key={feat}
                      className="rounded bg-accent-primary/10 px-1.5 py-0.2 text-[9px] font-mono text-accent-primary"
                    >
                      {feat}
                    </span>
                  ))}
                </div>
              </div>
            </div>

            <div className="mt-5 pt-3 border-t border-border-subtle flex items-center justify-between">
              <span className="text-[10px] font-mono text-text-secondary">
                systemd slice enforced
              </span>
              <button className="text-xs font-semibold text-accent-primary hover:underline flex items-center gap-1">
                <Sliders className="h-3 w-3" /> Edit Limits
              </button>
            </div>
          </div>
        ))}
      </div>

      {/* Create Modal */}
      {isModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-xs p-4">
          <div className="w-full max-w-md rounded-xl border border-border-subtle bg-bg-card p-6 shadow-xl space-y-4">
            <div className="flex items-center justify-between border-b border-border-subtle pb-3">
              <h2 className="text-base font-bold text-text-primary">
                New Hosting Package
              </h2>
              <button
                onClick={() => setIsModalOpen(false)}
                className="text-text-secondary hover:text-text-primary"
              >
                <X className="h-4 w-4" />
              </button>
            </div>

            <form onSubmit={handleCreatePackage} className="space-y-4 text-xs">
              <div>
                <label className="block font-semibold text-text-secondary mb-1">
                  Package Name
                </label>
                <input
                  type="text"
                  required
                  placeholder="e.g. Gold Cloud Hosting"
                  value={newPkgName}
                  onChange={(e) => setNewPkgName(e.target.value)}
                  className="w-full rounded-lg border border-border-subtle bg-bg-elevated px-3 py-2 text-text-primary focus:outline-none focus:border-accent-primary"
                />
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block font-semibold text-text-secondary mb-1">
                    Disk Quota (GB)
                  </label>
                  <input
                    type="number"
                    min="1"
                    value={newQuotaGb}
                    onChange={(e) => setNewQuotaGb(e.target.value)}
                    className="w-full rounded-lg border border-border-subtle bg-bg-elevated px-3 py-2 text-text-primary font-mono focus:outline-none focus:border-accent-primary"
                  />
                </div>
                <div>
                  <label className="block font-semibold text-text-secondary mb-1">
                    Max Domains
                  </label>
                  <input
                    type="number"
                    min="1"
                    value={newMaxDomains}
                    onChange={(e) => setNewMaxDomains(e.target.value)}
                    className="w-full rounded-lg border border-border-subtle bg-bg-elevated px-3 py-2 text-text-primary font-mono focus:outline-none focus:border-accent-primary"
                  />
                </div>
              </div>

              <div className="flex items-center gap-2 pt-1">
                <input
                  type="checkbox"
                  id="shellAccess"
                  checked={newShellAccess}
                  onChange={(e) => setNewShellAccess(e.target.checked)}
                  className="rounded border-border-subtle text-accent-primary focus:ring-accent-primary"
                />
                <label
                  htmlFor="shellAccess"
                  className="text-text-primary font-semibold"
                >
                  Enable Jailed SSH Terminal Access
                </label>
              </div>

              <div className="flex justify-end gap-2 pt-3 border-t border-border-subtle">
                <button
                  type="button"
                  onClick={() => setIsModalOpen(false)}
                  className="px-3 py-2 rounded-lg border border-border-subtle text-text-secondary hover:bg-bg-elevated font-semibold"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 rounded-lg bg-accent-primary text-white font-semibold hover:bg-accent-hover shadow-sm"
                >
                  Save Package
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};
