import React from 'react';
import { useTranslation } from 'react-i18next';
import {
  LayoutDashboard,
  Users,
  Package,
  Globe,
  Database,
  Network,
  Mail,
  FolderArchive,
  FolderOpen,
  Terminal,
  ShieldCheck,
  History,
  HardDriveDownload,
  Sliders,
} from 'lucide-react';
import type { UserRole } from '../types';

interface SidebarProps {
  currentTab: string;
  onSelectTab: (tab: string) => void;
  role: UserRole;
}

export const Sidebar: React.FC<SidebarProps> = ({ currentTab, onSelectTab, role }) => {
  const { t } = useTranslation();

  const navItems = [
    {
      id: 'overview',
      label: t('nav.overview', 'Server Overview'),
      icon: LayoutDashboard,
      roles: ['superadmin', 'reseller', 'tenant'],
    },
    {
      id: 'accounts',
      label: t('nav.accounts', 'Hosting Accounts'),
      icon: Users,
      roles: ['superadmin', 'reseller'],
    },
    {
      id: 'packages',
      label: t('nav.packages', 'Packages & Feature Lists'),
      icon: Package,
      roles: ['superadmin'],
    },
    {
      id: 'domains',
      label: t('nav.domains', 'Domains & Multi-PHP'),
      icon: Globe,
      roles: ['superadmin', 'reseller', 'tenant'],
    },
    {
      id: 'databases',
      label: t('nav.databases', 'Databases (MariaDB / Postgres)'),
      icon: Database,
      roles: ['superadmin', 'reseller', 'tenant'],
    },
    {
      id: 'dns',
      label: t('nav.dns', 'DNS Zones & DNS Clustering'),
      icon: Network,
      roles: ['superadmin', 'reseller', 'tenant'],
    },
    {
      id: 'email',
      label: t('nav.email', 'Email Hosting & Deliverability'),
      icon: Mail,
      roles: ['superadmin', 'reseller', 'tenant'],
    },
    {
      id: 'apps',
      label: t('nav.apps', 'App Installer & Staging'),
      icon: FolderArchive,
      roles: ['superadmin', 'reseller', 'tenant'],
    },
    {
      id: 'files',
      label: t('nav.files', 'File Manager'),
      icon: FolderOpen,
      roles: ['superadmin', 'reseller', 'tenant'],
    },
    {
      id: 'terminal',
      label: t('nav.terminal', 'Web SSH Terminal'),
      icon: Terminal,
      roles: ['superadmin', 'tenant'],
    },
    {
      id: 'security',
      label: t('nav.security', 'Coraza WAF & CrowdSec Firewall'),
      icon: ShieldCheck,
      roles: ['superadmin'],
    },
    {
      id: 'backups',
      label: t('nav.backups', 'Restic Backups & DR'),
      icon: HardDriveDownload,
      roles: ['superadmin', 'reseller', 'tenant'],
    },
    {
      id: 'config-editor',
      label: 'Service Config Editor',
      icon: Sliders,
      roles: ['superadmin'],
    },
    {
      id: 'audit-logs',
      label: 'Immutable Audit Log',
      icon: History,
      roles: ['superadmin'],
    },
  ];

  const visibleItems = navItems.filter((item) => item.roles.includes(role));

  return (
    <aside className="w-64 border-r border-border-subtle bg-bg-card/50 flex flex-col shrink-0">
      <div className="p-4 space-y-1">
        <div className="px-3 py-2 text-[10px] font-bold uppercase tracking-wider text-text-secondary">
          {role === 'superadmin' ? 'WHM Administration' : role === 'reseller' ? 'Reseller Panel' : 'cPanel Hosting'}
        </div>
        {visibleItems.map((item) => {
          const Icon = item.icon;
          const isActive = currentTab === item.id;
          return (
            <button
              key={item.id}
              onClick={() => onSelectTab(item.id)}
              className={`w-full flex items-center gap-3 px-3 py-2 rounded-lg text-xs font-semibold transition-colors ${
                isActive
                  ? 'bg-accent-primary text-white shadow-sm'
                  : 'text-text-secondary hover:text-text-primary hover:bg-bg-elevated'
              }`}
            >
              <Icon className="h-4 w-4 shrink-0" />
              <span className="truncate">{item.label}</span>
            </button>
          );
        })}
      </div>
    </aside>
  );
};
