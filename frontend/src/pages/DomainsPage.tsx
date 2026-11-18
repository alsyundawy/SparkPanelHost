import React, { useState, useEffect } from "react";
import { useTranslation } from "react-i18next";
import { Globe, Plus, Shield, Lock, AlertTriangle, Trash2 } from "lucide-react";
import type { Domain } from "../types";

export const DomainsPage: React.FC = () => {
  const { t } = useTranslation();
  const [domains, setDomains] = useState<Domain[]>([]);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [newDomain, setNewDomain] = useState({
    domain_name: "",
    php_version: "8.5",
    waf_mode: "Blocking",
  });
  const [loading, setLoading] = useState(false);

  const fetchDomains = async () => {
    try {
      const res = await fetch("/api/v1/domains");
      if (res.ok) {
        const data = await res.json();
        setDomains(data.domains || []);
      }
    } catch {
      // Fallback sample data
      setDomains([
        {
          id: "dom-1",
          user_id: "usr-1",
          domain_name: "example.com",
          document_root: "/home/tenant/public_html/example.com",
          php_version: "8.5",
          ssl_enabled: true,
          force_https: true,
          waf_mode: "Blocking",
          waf_profile: "default",
          nginx_template: "default",
          created_at: new Date().toISOString(),
        },
      ]);
    }
  };

  useEffect(() => {
    fetchDomains();
  }, []);

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    try {
      const res = await fetch("/api/v1/domains", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(newDomain),
      });
      if (res.ok) {
        setIsModalOpen(false);
        setNewDomain({
          domain_name: "",
          php_version: "8.5",
          waf_mode: "Blocking",
        });
        fetchDomains();
      }
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm("Are you sure you want to delete this domain?")) return;
    await fetch(`/api/v1/domains/${id}`, { method: "DELETE" });
    fetchDomains();
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-xl font-bold tracking-tight text-text-primary">
            {t("nav.domains", "Domains & Multi-PHP")}
          </h1>
          <p className="text-xs text-text-secondary mt-0.5">
            Manage Virtual Hosts, Multi-PHP Switcher (8.2–8.6), SSL & Coraza WAF
          </p>
        </div>
        <button
          onClick={() => setIsModalOpen(true)}
          className="flex items-center gap-2 rounded-lg bg-accent-primary px-3.5 py-2 text-xs font-semibold text-white shadow-xs hover:bg-accent-hover transition-colors self-start sm:self-auto"
        >
          <Plus className="h-4 w-4" />
          <span>Add Domain</span>
        </button>
      </div>

      {/* PHP 8.2 EOL Warning Banner */}
      <div className="rounded-xl border border-status-warning/30 bg-status-warning/10 p-4 flex items-start gap-3">
        <AlertTriangle className="h-5 w-5 text-status-warning shrink-0 mt-0.5" />
        <div className="text-xs text-text-primary">
          <div className="font-bold text-status-warning">
            {t("php.eol_badge", "PHP 8.2 Security EOL: 31 Dec 2026")}
          </div>
          <p className="text-text-secondary mt-0.5">
            {t(
              "php.eol_warning",
              "PHP 8.2 reaches Security End-of-Life on Dec 31, 2026. Upgrading to PHP 8.5 (Recommended Default) or PHP 8.6 is strongly advised.",
            )}
          </p>
        </div>
      </div>

      {/* Domains Table */}
      <div className="rounded-xl border border-border-subtle bg-bg-card overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-left text-xs">
            <thead className="bg-bg-elevated border-b border-border-subtle text-text-secondary font-semibold">
              <tr>
                <th className="px-4 py-3">Domain Name</th>
                <th className="px-4 py-3">Document Root</th>
                <th className="px-4 py-3">PHP Version</th>
                <th className="px-4 py-3">SSL Certificate</th>
                <th className="px-4 py-3">Coraza WAF</th>
                <th className="px-4 py-3 text-right">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border-subtle text-text-primary">
              {domains.length === 0 ? (
                <tr>
                  <td
                    colSpan={6}
                    className="px-4 py-8 text-center text-text-secondary"
                  >
                    No domains configured yet. Click "Add Domain" to provision
                    one.
                  </td>
                </tr>
              ) : (
                domains.map((d) => (
                  <tr
                    key={d.id}
                    className="hover:bg-bg-elevated/50 transition-colors"
                  >
                    <td className="px-4 py-3 font-semibold flex items-center gap-2">
                      <Globe className="h-3.5 w-3.5 text-accent-primary" />
                      <span>{d.domain_name}</span>
                    </td>
                    <td className="px-4 py-3 text-text-secondary font-mono text-[11px]">
                      {d.document_root}
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-1.5 font-mono font-semibold">
                        <span>PHP {d.php_version}</span>
                        {d.php_version === "8.2" && (
                          <span className="rounded bg-status-warning/20 px-1 py-0.2 text-[9px] font-bold text-status-warning">
                            EOL 2026
                          </span>
                        )}
                        {d.php_version === "8.5" && (
                          <span className="rounded bg-status-success/20 px-1 py-0.2 text-[9px] font-bold text-status-success">
                            Default
                          </span>
                        )}
                        {d.php_version === "8.6" && (
                          <span className="rounded bg-accent-primary/20 px-1 py-0.2 text-[9px] font-bold text-accent-primary">
                            Latest
                          </span>
                        )}
                      </div>
                    </td>
                    <td className="px-4 py-3">
                      {d.ssl_enabled ? (
                        <span className="flex items-center gap-1 text-status-success font-semibold text-[11px]">
                          <Lock className="h-3 w-3" /> Let's Encrypt (ARI)
                        </span>
                      ) : (
                        <span className="text-text-secondary text-[11px]">
                          Disabled
                        </span>
                      )}
                    </td>
                    <td className="px-4 py-3">
                      <span className="rounded bg-accent-primary/10 px-2 py-0.5 text-[10px] font-bold text-accent-primary flex items-center gap-1 w-max">
                        <Shield className="h-3 w-3" /> {d.waf_mode}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-right">
                      <button
                        onClick={() => handleDelete(d.id)}
                        className="text-text-secondary hover:text-status-danger p-1 rounded transition-colors"
                        title="Delete Domain"
                      >
                        <Trash2 className="h-4 w-4" />
                      </button>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* Add Domain Modal */}
      {isModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
          <div className="w-full max-w-md rounded-xl border border-border-subtle bg-bg-card p-6 shadow-2xl">
            <h2 className="text-base font-bold text-text-primary mb-4 flex items-center gap-2">
              <Globe className="h-5 w-5 text-accent-primary" />
              <span>Add New Virtual Host</span>
            </h2>

            <form onSubmit={handleCreate} className="space-y-4 text-xs">
              <div>
                <label className="block font-semibold text-text-secondary mb-1">
                  Domain Name (FQDN)
                </label>
                <input
                  type="text"
                  required
                  placeholder="example.com"
                  value={newDomain.domain_name}
                  onChange={(e) =>
                    setNewDomain({ ...newDomain, domain_name: e.target.value })
                  }
                  className="w-full rounded-lg border border-border-subtle bg-bg-elevated px-3 py-2 text-text-primary focus:outline-none focus:border-accent-primary font-mono"
                />
              </div>

              <div>
                <label className="block font-semibold text-text-secondary mb-1">
                  PHP Runtime Version
                </label>
                <select
                  value={newDomain.php_version}
                  onChange={(e) =>
                    setNewDomain({ ...newDomain, php_version: e.target.value })
                  }
                  className="w-full rounded-lg border border-border-subtle bg-bg-elevated px-3 py-2 text-text-primary focus:outline-none focus:border-accent-primary"
                >
                  <option value="8.6">
                    PHP 8.6 (GA 19 Nov 2026 - Maximum Performance)
                  </option>
                  <option value="8.5">PHP 8.5 (Default Rekomendasi)</option>
                  <option value="8.4">PHP 8.4</option>
                  <option value="8.3">PHP 8.3</option>
                  <option value="8.2">
                    PHP 8.2 (Legacy - EOL 31 Des 2026)
                  </option>
                  <option value="static">Static HTML / Frontend SPA</option>
                  <option value="proxy">
                    Reverse Proxy (Node.js / Python)
                  </option>
                </select>
              </div>

              <div>
                <label className="block font-semibold text-text-secondary mb-1">
                  Coraza WAF Mode (OWASP CRS v4)
                </label>
                <select
                  value={newDomain.waf_mode}
                  onChange={(e) =>
                    setNewDomain({ ...newDomain, waf_mode: e.target.value })
                  }
                  className="w-full rounded-lg border border-border-subtle bg-bg-elevated px-3 py-2 text-text-primary focus:outline-none focus:border-accent-primary"
                >
                  <option value="Blocking">
                    Blocking (Full Active Defense)
                  </option>
                  <option value="DetectionOnly">
                    DetectionOnly (Log Alerts Without Dropping)
                  </option>
                  <option value="Off">Off (Disabled)</option>
                </select>
              </div>

              <div className="flex items-center justify-end gap-2 pt-3">
                <button
                  type="button"
                  onClick={() => setIsModalOpen(false)}
                  className="rounded-lg border border-border-subtle px-3 py-1.5 font-semibold text-text-secondary hover:bg-bg-elevated"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={loading}
                  className="rounded-lg bg-accent-primary px-3 py-1.5 font-semibold text-white hover:bg-accent-hover transition-colors"
                >
                  {loading ? "Creating..." : "Create Virtual Host"}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};
