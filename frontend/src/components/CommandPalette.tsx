import React, { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { Search, X, Globe, Database, Network, Mail, FolderArchive, Terminal, Shield } from 'lucide-react';

interface CommandPaletteProps {
  isOpen: boolean;
  onClose: () => void;
  onSelectAction: (tab: string) => void;
}

export const CommandPalette: React.FC<CommandPaletteProps> = ({ isOpen, onClose, onSelectAction }) => {
  const { t } = useTranslation();
  const [query, setQuery] = useState('');

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault();
        if (isOpen) {
          onClose();
        } else {
          // open handled by parent
        }
      }
      if (e.key === 'Escape' && isOpen) {
        onClose();
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [isOpen, onClose]);

  if (!isOpen) return null;

  const actions = [
    { id: 'domains', label: t('nav.domains', 'Domains & Multi-PHP Switcher'), icon: Globe },
    { id: 'databases', label: t('nav.databases', 'Databases (MariaDB 12.3 / PostgreSQL 17)'), icon: Database },
    { id: 'dns', label: t('nav.dns', 'DNS Zones & WHM DNS Only Clustering'), icon: Network },
    { id: 'email', label: t('nav.email', 'Email Hosting, DKIM & Deliverability'), icon: Mail },
    { id: 'apps', label: t('nav.apps', '1-Click App Installer & Staging'), icon: FolderArchive },
    { id: 'terminal', label: t('nav.terminal', 'Web SSH Terminal (PTY)'), icon: Terminal },
    { id: 'security', label: t('nav.security', 'Coraza WAF & CrowdSec Firewall'), icon: Shield },
  ];

  const filtered = actions.filter((a) =>
    a.label.toLowerCase().includes(query.toLowerCase())
  );

  return (
    <div className="fixed inset-0 z-50 flex items-start justify-center pt-24 bg-black/60 backdrop-blur-sm p-4">
      <div className="w-full max-w-lg rounded-xl border border-border-subtle bg-bg-card shadow-2xl overflow-hidden animate-in fade-in zoom-in-95 duration-100">
        <div className="flex items-center gap-3 border-b border-border-subtle px-4 py-3 bg-bg-elevated">
          <Search className="h-4 w-4 text-text-secondary" />
          <input
            autoFocus
            type="text"
            placeholder={t('app.command_placeholder', 'Search or run a command (Cmd+K)...')}
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            className="w-full bg-transparent text-sm text-text-primary placeholder:text-text-secondary focus:outline-none"
          />
          <button onClick={onClose} className="text-text-secondary hover:text-text-primary">
            <X className="h-4 w-4" />
          </button>
        </div>

        <div className="p-2 max-h-80 overflow-y-auto space-y-1">
          {filtered.length === 0 ? (
            <div className="px-4 py-6 text-center text-xs text-text-secondary">
              No matching commands found.
            </div>
          ) : (
            filtered.map((item) => {
              const Icon = item.icon;
              return (
                <button
                  key={item.id}
                  onClick={() => {
                    onSelectAction(item.id);
                    onClose();
                  }}
                  className="w-full flex items-center gap-3 px-3 py-2.5 rounded-lg text-xs font-semibold text-text-secondary hover:text-text-primary hover:bg-bg-elevated transition-colors text-left"
                >
                  <Icon className="h-4 w-4 text-accent-primary" />
                  <span>{item.label}</span>
                </button>
              );
            })
          )}
        </div>
      </div>
    </div>
  );
};
