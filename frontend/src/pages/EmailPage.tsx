import React, { useState } from "react";
import { useTranslation } from "react-i18next";
import { Mail, Plus, ShieldCheck, Key, CheckCircle2, Copy } from "lucide-react";

interface MailboxItem {
  id: string;
  email: string;
  quota_mb: number;
  active: boolean;
}

export const EmailPage: React.FC = () => {
  const { t } = useTranslation();
  const [activeTab, setActiveTab] = useState<"boxes" | "dkim" | "rbl">("boxes");
  const [mailboxes, setMailboxes] = useState<MailboxItem[]>([
    { id: "mb-1", email: "admin@example.com", quota_mb: 2048, active: true },
    { id: "mb-2", email: "support@example.com", quota_mb: 1024, active: true },
  ]);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [newMail, setNewMail] = useState({
    localPart: "",
    password: "",
    quota: 2048,
  });
  const [copiedKey, setCopiedKey] = useState("");

  const rsaKey =
    "v=DKIM1; k=rsa; p=MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA3...";
  const ed25519Key =
    "v=DKIM1; k=ed25519; p=11qYAYKxCrfVS/7Ty4x8V1v8K3tK9yL1a9Z0...";

  const copyToClipboard = (text: string, label: string) => {
    navigator.clipboard.writeText(text);
    setCopiedKey(label);
    setTimeout(() => setCopiedKey(""), 2000);
  };

  const handleCreate = (e: React.FormEvent) => {
    e.preventDefault();
    const created: MailboxItem = {
      id: `mb-${Date.now()}`,
      email: `${newMail.localPart}@example.com`,
      quota_mb: newMail.quota,
      active: true,
    };
    setMailboxes([...mailboxes, created]);
    setIsModalOpen(false);
    setNewMail({ localPart: "", password: "", quota: 2048 });
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-xl font-bold tracking-tight text-text-primary">
            {t("nav.email", "Email Hosting & Deliverability")}
          </h1>
          <p className="text-xs text-text-secondary mt-0.5">
            Postfix + Dovecot 2.4.5 + Rspamd 4.3 (RSA + Ed25519 Dual-Signing,
            MTA-STS, RBL Monitor)
          </p>
        </div>
        <button
          onClick={() => setIsModalOpen(true)}
          className="flex items-center gap-2 rounded-lg bg-accent-primary px-3.5 py-2 text-xs font-semibold text-white shadow-xs hover:bg-accent-hover transition-colors self-start sm:self-auto"
        >
          <Plus className="h-4 w-4" />
          <span>Add Mailbox</span>
        </button>
      </div>

      {/* Tabs */}
      <div className="flex items-center gap-2 border-b border-border-subtle text-xs pb-1">
        <button
          onClick={() => setActiveTab("boxes")}
          className={`px-3 py-1.5 font-semibold transition-colors ${
            activeTab === "boxes"
              ? "border-b-2 border-accent-primary text-accent-primary"
              : "text-text-secondary hover:text-text-primary"
          }`}
        >
          Mailboxes & Quotas
        </button>
        <button
          onClick={() => setActiveTab("dkim")}
          className={`px-3 py-1.5 font-semibold transition-colors ${
            activeTab === "dkim"
              ? "border-b-2 border-accent-primary text-accent-primary"
              : "text-text-secondary hover:text-text-primary"
          }`}
        >
          DKIM Dual-Signing & MTA-STS
        </button>
        <button
          onClick={() => setActiveTab("rbl")}
          className={`px-3 py-1.5 font-semibold transition-colors ${
            activeTab === "rbl"
              ? "border-b-2 border-accent-primary text-accent-primary"
              : "text-text-secondary hover:text-text-primary"
          }`}
        >
          RBL Reputation Monitor
        </button>
      </div>

      {/* Tab: Mailboxes */}
      {activeTab === "boxes" && (
        <div className="rounded-xl border border-border-subtle bg-bg-card overflow-hidden">
          <table className="w-full text-left text-xs">
            <thead className="bg-bg-elevated border-b border-border-subtle text-text-secondary font-semibold">
              <tr>
                <th className="px-4 py-3">Email Address</th>
                <th className="px-4 py-3">Storage Quota</th>
                <th className="px-4 py-3">Password Cipher</th>
                <th className="px-4 py-3">Status</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border-subtle text-text-primary">
              {mailboxes.map((mb) => (
                <tr
                  key={mb.id}
                  className="hover:bg-bg-elevated/50 transition-colors"
                >
                  <td className="px-4 py-3 font-semibold flex items-center gap-2">
                    <Mail className="h-3.5 w-3.5 text-accent-primary" />
                    <span>{mb.email}</span>
                  </td>
                  <td className="px-4 py-3 font-mono tabular-nums text-[11px]">
                    {mb.quota_mb} MB
                  </td>
                  <td className="px-4 py-3 font-mono text-text-secondary text-[10px]">
                    ARGON2ID (Dovecot 2.4)
                  </td>
                  <td className="px-4 py-3">
                    <span className="rounded bg-status-success/15 px-2 py-0.5 text-[10px] font-bold text-status-success">
                      Active
                    </span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Tab: DKIM & Deliverability */}
      {activeTab === "dkim" && (
        <div className="space-y-4">
          <div className="rounded-xl border border-border-subtle bg-bg-card p-5">
            <h2 className="text-sm font-bold text-text-primary flex items-center gap-2 mb-3">
              <Key className="h-4 w-4 text-accent-primary" />
              <span>Dual-Signing DKIM Records (Rspamd 4.3)</span>
            </h2>
            <div className="space-y-3 text-xs">
              <div>
                <div className="flex items-center justify-between mb-1">
                  <span className="font-semibold text-text-primary">
                    Selector: default._domainkey (RSA-2048)
                  </span>
                  <button
                    onClick={() => copyToClipboard(rsaKey, "rsa")}
                    className="flex items-center gap-1 text-[11px] text-accent-primary hover:underline"
                  >
                    <Copy className="h-3 w-3" />
                    <span>
                      {copiedKey === "rsa" ? "Copied!" : "Copy TXT Value"}
                    </span>
                  </button>
                </div>
                <div className="rounded-lg bg-bg-elevated border border-border-subtle p-2 font-mono text-[11px] break-all text-text-secondary">
                  {rsaKey}
                </div>
              </div>

              <div>
                <div className="flex items-center justify-between mb-1">
                  <span className="font-semibold text-text-primary">
                    Selector: ed25519._domainkey (Ed25519 Modern)
                  </span>
                  <button
                    onClick={() => copyToClipboard(ed25519Key, "ed")}
                    className="flex items-center gap-1 text-[11px] text-accent-primary hover:underline"
                  >
                    <Copy className="h-3 w-3" />
                    <span>
                      {copiedKey === "ed" ? "Copied!" : "Copy TXT Value"}
                    </span>
                  </button>
                </div>
                <div className="rounded-lg bg-bg-elevated border border-border-subtle p-2 font-mono text-[11px] break-all text-text-secondary">
                  {ed25519Key}
                </div>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Tab: RBL Monitor */}
      {activeTab === "rbl" && (
        <div className="rounded-xl border border-border-subtle bg-bg-card p-5">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-sm font-bold text-text-primary flex items-center gap-2">
              <ShieldCheck className="h-4 w-4 text-status-success" />
              <span>IP Blacklist / DNSBL Health Status</span>
            </h2>
            <span className="rounded bg-status-success/15 px-2.5 py-0.5 text-xs font-bold text-status-success">
              100% Clean
            </span>
          </div>

          <div className="space-y-2 text-xs">
            {[
              "zen.spamhaus.org",
              "bl.spamcop.net",
              "b.barracudacentral.org",
              "dnsbl.sorbs.net",
              "psbl.surriel.com",
            ].map((rbl) => (
              <div
                key={rbl}
                className="flex items-center justify-between p-3 rounded-lg border border-border-subtle bg-bg-elevated"
              >
                <span className="font-mono text-text-primary">{rbl}</span>
                <span className="flex items-center gap-1.5 text-status-success font-semibold text-[11px]">
                  <CheckCircle2 className="h-3.5 w-3.5" /> Clean (Not Listed)
                </span>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Add Mailbox Modal */}
      {isModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
          <div className="w-full max-w-md rounded-xl border border-border-subtle bg-bg-card p-6 shadow-2xl">
            <h2 className="text-base font-bold text-text-primary mb-4 flex items-center gap-2">
              <Mail className="h-5 w-5 text-accent-primary" />
              <span>Add New Mailbox</span>
            </h2>

            <form onSubmit={handleCreate} className="space-y-4 text-xs">
              <div>
                <label className="block font-semibold text-text-secondary mb-1">
                  Email Address
                </label>
                <div className="flex items-center">
                  <input
                    type="text"
                    required
                    placeholder="user"
                    value={newMail.localPart}
                    onChange={(e) =>
                      setNewMail({ ...newMail, localPart: e.target.value })
                    }
                    className="w-full rounded-l-lg border border-border-subtle bg-bg-elevated px-3 py-2 text-text-primary font-mono focus:outline-none focus:border-accent-primary"
                  />
                  <span className="px-3 py-2 rounded-r-lg border border-l-0 border-border-subtle bg-bg-elevated text-text-secondary font-mono">
                    @example.com
                  </span>
                </div>
              </div>

              <div>
                <label className="block font-semibold text-text-secondary mb-1">
                  Password
                </label>
                <input
                  type="password"
                  required
                  placeholder="Strong password..."
                  value={newMail.password}
                  onChange={(e) =>
                    setNewMail({ ...newMail, password: e.target.value })
                  }
                  className="w-full rounded-lg border border-border-subtle bg-bg-elevated px-3 py-2 text-text-primary font-mono focus:outline-none focus:border-accent-primary"
                />
              </div>

              <div>
                <label className="block font-semibold text-text-secondary mb-1">
                  Quota (MB)
                </label>
                <input
                  type="number"
                  required
                  value={newMail.quota}
                  onChange={(e) =>
                    setNewMail({
                      ...newMail,
                      quota: parseInt(e.target.value) || 1024,
                    })
                  }
                  className="w-full rounded-lg border border-border-subtle bg-bg-elevated px-3 py-2 text-text-primary font-mono focus:outline-none focus:border-accent-primary"
                />
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
                  className="rounded-lg bg-accent-primary px-3 py-1.5 font-semibold text-white hover:bg-accent-hover transition-colors"
                >
                  Save Mailbox
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};
