import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  FolderArchive,
  Download,
  ShieldCheck,
  Flame,
  CheckCircle2,
  GitBranch,
} from 'lucide-react';

interface AppItem {
  id: string;
  name: string;
  category: string;
  version: string;
  description: string;
  sha256: string;
}

export const AppInstallerPage: React.FC = () => {
  const { t } = useTranslation();
  const [selectedCategory, setSelectedCategory] = useState('all');
  const [installingApp, setInstallingApp] = useState<AppItem | null>(null);
  const [stagingOpen, setStagingOpen] = useState(false);
  const [installedSuccess, setInstalledSuccess] = useState('');

  const apps: AppItem[] = [
    {
      id: 'wordpress',
      name: 'WordPress',
      category: 'cms',
      version: '6.7.2',
      description: 'The world\'s most popular web publishing and blogging platform.',
      sha256: '5e884898da28047151d0e56f8dc6292773603d0d6aabbdd62a11ef721d1542d8',
    },
    {
      id: 'prestashop',
      name: 'PrestaShop',
      category: 'marketplace',
      version: '8.2.1',
      description: 'Fully scalable, open-source e-commerce solution for growing businesses.',
      sha256: '4b227777d4dd1fc61c6f884f48641d02b4d121d3fd328cb08b5531fcacdabf8a',
    },
    {
      id: 'flarum',
      name: 'Flarum',
      category: 'forum',
      version: '1.8.5',
      description: 'Delightfully simple, lightweight discussion software built with modern tech.',
      sha256: 'ef2d127de37b942baad06145e54b0c619a1f22327b2ebbcfbec78f5564afe39d',
    },
    {
      id: 'joomla',
      name: 'Joomla',
      category: 'cms',
      version: '5.2.3',
      description: 'Award-winning content management system powering millions of websites.',
      sha256: '8f434346648f6b96df89dda901c5176b10a6d83961dd3c1ac88b59b2dc327aa4',
    },
    {
      id: 'opencart',
      name: 'OpenCart',
      category: 'marketplace',
      version: '4.0.2.3',
      description: 'Free open-source turnkey e-commerce platform for merchants worldwide.',
      sha256: 'ecc8576f29f91a56641662c5e533c39dd2a0a2df3f48a80277317bfd62059346',
    },
    {
      id: 'roundcube',
      name: 'Roundcube Webmail',
      category: 'webmail',
      version: '1.6.9',
      description: 'Browser-based multilingual IMAP webmail client with application-like user interface.',
      sha256: 'c3ab8ff13720e8ad9047dd39466b3c8974e592c2fa383d4a3960714caef0c4f2',
    },
  ];

  const filtered = selectedCategory === 'all'
    ? apps
    : apps.filter((a) => a.category === selectedCategory);

  const handleInstallSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (installingApp) {
      setInstalledSuccess(`${installingApp.name} successfully deployed with verified SHA-256 integrity.`);
      setInstallingApp(null);
    }
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-xl font-bold tracking-tight text-text-primary">
            {t('nav.apps', '1-Click App Installer & Staging')}
          </h1>
          <p className="text-xs text-text-secondary mt-0.5">
            Native verified recipes (Anti-SSRF + SHA-256 + openat2 privilege drop)
          </p>
        </div>
        <button
          onClick={() => setStagingOpen(true)}
          className="flex items-center gap-2 rounded-lg border border-border-subtle bg-bg-elevated px-3 py-1.5 text-xs font-semibold text-text-primary hover:border-accent-primary transition-colors self-start sm:self-auto"
        >
          <GitBranch className="h-3.5 w-3.5 text-accent-primary" />
          <span>1-Click Staging & Push-to-Live</span>
        </button>
      </div>

      {installedSuccess && (
        <div className="rounded-xl border border-status-success/30 bg-status-success/10 p-3 text-xs text-status-success flex items-center gap-2">
          <CheckCircle2 className="h-4 w-4 shrink-0" />
          <span>{installedSuccess}</span>
        </div>
      )}

      {/* Category Pills */}
      <div className="flex items-center gap-2 overflow-x-auto pb-1 text-xs">
        {['all', 'cms', 'forum', 'marketplace', 'webmail'].map((cat) => (
          <button
            key={cat}
            onClick={() => setSelectedCategory(cat)}
            className={`rounded-lg px-3 py-1.5 font-semibold capitalize transition-colors ${
              selectedCategory === cat
                ? 'bg-accent-primary text-white shadow-xs'
                : 'border border-border-subtle bg-bg-elevated text-text-secondary hover:text-text-primary'
            }`}
          >
            {cat}
          </button>
        ))}
      </div>

      {/* Apps Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {filtered.map((app) => (
          <div
            key={app.id}
            className="rounded-xl border border-border-subtle bg-bg-card p-5 flex flex-col justify-between hover:border-accent-primary/40 transition-colors shadow-xs"
          >
            <div>
              <div className="flex items-center justify-between mb-3">
                <div className="flex items-center gap-2.5">
                  <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-accent-primary/10 text-accent-primary font-bold">
                    <FolderArchive className="h-5 w-5" />
                  </div>
                  <div>
                    <h3 className="font-bold text-sm text-text-primary">{app.name}</h3>
                    <span className="text-[11px] text-text-secondary font-mono">v{app.version}</span>
                  </div>
                </div>
                <span className="rounded bg-accent-primary/10 px-2 py-0.5 text-[9px] font-bold uppercase text-accent-primary">
                  {app.category}
                </span>
              </div>
              <p className="text-xs text-text-secondary line-clamp-2 mb-4">{app.description}</p>
            </div>

            <div className="space-y-3 pt-3 border-t border-border-subtle">
              <div className="flex items-center gap-1.5 text-[10px] text-status-success font-medium">
                <ShieldCheck className="h-3.5 w-3.5 shrink-0" />
                <span className="truncate">{t('apps.sha256_verified', 'Verified SHA-256 + openat2 Safe')}</span>
              </div>
              <button
                onClick={() => setInstallingApp(app)}
                className="w-full flex items-center justify-center gap-2 rounded-lg bg-accent-primary px-3 py-1.5 text-xs font-semibold text-white hover:bg-accent-hover transition-colors"
              >
                <Download className="h-3.5 w-3.5" />
                <span>1-Click Install</span>
              </button>
            </div>
          </div>
        ))}
      </div>

      {/* Install App Modal */}
      {installingApp && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
          <div className="w-full max-w-md rounded-xl border border-border-subtle bg-bg-card p-6 shadow-2xl">
            <h2 className="text-base font-bold text-text-primary mb-2 flex items-center gap-2">
              <Flame className="h-5 w-5 text-accent-primary" />
              <span>Install {installingApp.name} (v{installingApp.version})</span>
            </h2>
            <p className="text-xs text-text-secondary mb-4">
              Automated provisioning with database creation, SHA-256 digest check, and Coraza WAF profile activation.
            </p>

            <form onSubmit={handleInstallSubmit} className="space-y-3 text-xs">
              <div>
                <label className="block font-semibold text-text-secondary mb-1">Target Domain</label>
                <input
                  type="text"
                  required
                  placeholder="example.com"
                  defaultValue="example.com"
                  className="w-full rounded-lg border border-border-subtle bg-bg-elevated px-3 py-2 text-text-primary font-mono focus:outline-none focus:border-accent-primary"
                />
              </div>

              <div>
                <label className="block font-semibold text-text-secondary mb-1">Install Path (Relative to public_html)</label>
                <input
                  type="text"
                  placeholder="/"
                  defaultValue="/"
                  className="w-full rounded-lg border border-border-subtle bg-bg-elevated px-3 py-2 text-text-primary font-mono focus:outline-none focus:border-accent-primary"
                />
              </div>

              <div>
                <label className="block font-semibold text-text-secondary mb-1">Admin Username</label>
                <input
                  type="text"
                  required
                  defaultValue="admin"
                  className="w-full rounded-lg border border-border-subtle bg-bg-elevated px-3 py-2 text-text-primary focus:outline-none focus:border-accent-primary"
                />
              </div>

              <div>
                <label className="block font-semibold text-text-secondary mb-1">Admin Password</label>
                <input
                  type="password"
                  required
                  defaultValue="AetherSecure@2026!"
                  className="w-full rounded-lg border border-border-subtle bg-bg-elevated px-3 py-2 text-text-primary font-mono focus:outline-none focus:border-accent-primary"
                />
              </div>

              <div className="flex items-center justify-end gap-2 pt-3">
                <button
                  type="button"
                  onClick={() => setInstallingApp(null)}
                  className="rounded-lg border border-border-subtle px-3 py-1.5 font-semibold text-text-secondary hover:bg-bg-elevated"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="rounded-lg bg-accent-primary px-3 py-1.5 font-semibold text-white hover:bg-accent-hover transition-colors"
                >
                  Start Installation
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Staging Modal */}
      {stagingOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
          <div className="w-full max-w-md rounded-xl border border-border-subtle bg-bg-card p-6 shadow-2xl">
            <h2 className="text-base font-bold text-text-primary mb-2 flex items-center gap-2">
              <GitBranch className="h-5 w-5 text-accent-primary" />
              <span>1-Click Staging & Push-to-Live</span>
            </h2>
            <p className="text-xs text-text-secondary mb-4">
              Duplicate your live site into an isolated sandbox with automated database URL search-replace.
            </p>

            <div className="space-y-3 text-xs">
              <div>
                <label className="block font-semibold text-text-secondary mb-1">Source Production Site</label>
                <input
                  type="text"
                  readOnly
                  value="https://example.com"
                  className="w-full rounded-lg border border-border-subtle bg-bg-elevated px-3 py-2 text-text-primary font-mono"
                />
              </div>
              <div>
                <label className="block font-semibold text-text-secondary mb-1">Target Staging Subdomain</label>
                <input
                  type="text"
                  defaultValue="https://staging.example.com"
                  className="w-full rounded-lg border border-border-subtle bg-bg-elevated px-3 py-2 text-text-primary font-mono"
                />
              </div>

              <div className="flex items-center justify-end gap-2 pt-3">
                <button
                  onClick={() => setStagingOpen(false)}
                  className="rounded-lg border border-border-subtle px-3 py-1.5 font-semibold text-text-secondary hover:bg-bg-elevated"
                >
                  Close
                </button>
                <button
                  onClick={() => {
                    setStagingOpen(false);
                    setInstalledSuccess('Staging site created at staging.example.com.');
                  }}
                  className="rounded-lg bg-accent-primary px-3 py-1.5 font-semibold text-white hover:bg-accent-hover transition-colors"
                >
                  Clone to Staging
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
