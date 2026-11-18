import React from "react";
import { useTranslation } from "react-i18next";
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
  ChevronLeft,
  ChevronRight,
  ShieldAlert,
} from "lucide-react";
import type { UserRole } from "../types";

interface SidebarProps {
  currentTab: string;
  onSelectTab: (tab: string) => void;
  role: UserRole;
  isCollapsed?: boolean;
  onToggleCollapse?: () => void;
}

interface NavItem {
  id: string;
  label: string;
  icon: React.ComponentType<{ className?: string }>;
  roles: UserRole[];
  badge?: string;
  badgeVariant?: "accent" | "success" | "warning";
}

interface NavGroup {
  id: string;
  category: string;
  items: NavItem[];
}

export const Sidebar: React.FC<SidebarProps> = ({
  currentTab,
  onSelectTab,
  role,
  isCollapsed = false,
  onToggleCollapse,
}) => {
  const { t } = useTranslation();

  const navGroups: NavGroup[] = [
    {
      id: "core",
      category: t("sidebar.core", "Core Platform"),
      items: [
        {
          id: "overview",
          label: t("nav.overview", "Server Overview"),
          icon: LayoutDashboard,
          roles: ["superadmin", "reseller", "tenant"],
        },
      ],
    },
    {
      id: "management",
      category: t("sidebar.management", "Account Management"),
      items: [
        {
          id: "accounts",
          label: t("nav.accounts", "Hosting Accounts"),
          icon: Users,
          roles: ["superadmin", "reseller"],
        },
        {
          id: "packages",
          label: t("nav.packages", "Packages & Feature Lists"),
          icon: Package,
          roles: ["superadmin"],
          badge: "Gated",
          badgeVariant: "accent",
        },
      ],
    },
    {
      id: "hosting",
      category: t("sidebar.hosting", "Web & Data Services"),
      items: [
        {
          id: "domains",
          label: t("nav.domains", "Domains & Multi-PHP"),
          icon: Globe,
          roles: ["superadmin", "reseller", "tenant"],
          badge: "PHP 8.5",
          badgeVariant: "success",
        },
        {
          id: "databases",
          label: t("nav.databases", "Databases (MariaDB / Postgres)"),
          icon: Database,
          roles: ["superadmin", "reseller", "tenant"],
          badge: "MariaDB",
          badgeVariant: "accent",
        },
        {
          id: "dns",
          label: t("nav.dns", "DNS Zones & DNS Clustering"),
          icon: Network,
          roles: ["superadmin", "reseller", "tenant"],
          badge: "AXFR",
          badgeVariant: "accent",
        },
        {
          id: "apps",
          label: t("nav.apps", "App Installer & Staging"),
          icon: FolderArchive,
          roles: ["superadmin", "reseller", "tenant"],
          badge: "1-Click",
          badgeVariant: "warning",
        },
        {
          id: "files",
          label: t("nav.files", "File Manager"),
          icon: FolderOpen,
          roles: ["superadmin", "reseller", "tenant"],
        },
      ],
    },
    {
      id: "communication",
      category: t("sidebar.communication", "Access & Mail"),
      items: [
        {
          id: "email",
          label: t("nav.email", "Email Hosting & Deliverability"),
          icon: Mail,
          roles: ["superadmin", "reseller", "tenant"],
          badge: "RBL Clean",
          badgeVariant: "success",
        },
        {
          id: "terminal",
          label: t("nav.terminal", "Web SSH Terminal"),
          icon: Terminal,
          roles: ["superadmin", "tenant"],
        },
      ],
    },
    {
      id: "security_ops",
      category: t("sidebar.security_ops", "Security & Operations"),
      items: [
        {
          id: "security",
          label: t("nav.security", "Coraza WAF & CrowdSec Firewall"),
          icon: ShieldCheck,
          roles: ["superadmin"],
          badge: "CRS v4",
          badgeVariant: "success",
        },
        {
          id: "backups",
          label: t("nav.backups", "Restic Backups & DR"),
          icon: HardDriveDownload,
          roles: ["superadmin", "reseller", "tenant"],
          badge: "AES-256",
          badgeVariant: "accent",
        },
        {
          id: "config-editor",
          label: "Service Config Editor",
          icon: Sliders,
          roles: ["superadmin"],
        },
        {
          id: "audit-logs",
          label: "Immutable Audit Log",
          icon: History,
          roles: ["superadmin"],
        },
      ],
    },
  ];

  const getRoleTitle = () => {
    switch (role) {
      case "superadmin":
        return "WHM Administration";
      case "reseller":
        return "Reseller Panel";
      default:
        return "cPanel Hosting";
    }
  };

  return (
    <aside
      className={`border-r border-border-subtle bg-bg-card/60 backdrop-blur-md flex flex-col shrink-0 transition-all duration-300 ease-in-out select-none ${
        isCollapsed ? "w-16" : "w-64"
      }`}
    >
      {/* Scope / Role Header */}
      <div className="p-3 border-b border-border-subtle/60 flex items-center justify-between">
        {!isCollapsed ? (
          <div className="px-2">
            <span className="text-[10px] font-extrabold uppercase tracking-widest text-text-secondary/80">
              {getRoleTitle()}
            </span>
            <div className="flex items-center gap-1.5 mt-0.5">
              <span className="h-1.5 w-1.5 rounded-full bg-status-success animate-pulse" />
              <span className="text-[10px] font-mono text-text-secondary">
                Linux Node #1
              </span>
            </div>
          </div>
        ) : (
          <div className="mx-auto" title={getRoleTitle()}>
            <ShieldAlert className="h-4 w-4 text-accent-primary" />
          </div>
        )}
      </div>

      {/* Navigation List */}
      <div className="flex-1 overflow-y-auto px-2 py-3 space-y-4">
        {navGroups.map((group) => {
          const visibleItems = group.items.filter((item) =>
            item.roles.includes(role),
          );
          if (visibleItems.length === 0) return null;

          return (
            <div key={group.id} className="space-y-1">
              {/* Category Heading (shown only when expanded) */}
              {!isCollapsed && (
                <div className="px-3 py-1 text-[9px] font-bold uppercase tracking-wider text-text-secondary/60">
                  {group.category}
                </div>
              )}

              {/* Group Items */}
              {visibleItems.map((item) => {
                const Icon = item.icon;
                const isActive = currentTab === item.id;

                const getBadgeClass = (
                  variant?: "accent" | "success" | "warning",
                ) => {
                  switch (variant) {
                    case "success":
                      return "bg-status-success/15 text-status-success border-status-success/30";
                    case "warning":
                      return "bg-status-warning/15 text-status-warning border-status-warning/30";
                    default:
                      return "bg-accent-primary/15 text-accent-primary border-accent-primary/30";
                  }
                };

                return (
                  <button
                    key={item.id}
                    onClick={() => onSelectTab(item.id)}
                    title={isCollapsed ? item.label : undefined}
                    className={`w-full group relative flex items-center gap-3 px-3 py-2 rounded-lg text-xs font-semibold transition-all duration-150 ${
                      isActive
                        ? "bg-accent-primary text-white shadow-sm font-bold"
                        : "text-text-secondary hover:text-text-primary hover:bg-bg-elevated"
                    } ${isCollapsed ? "justify-center px-0" : ""}`}
                  >
                    <Icon
                      className={`h-4 w-4 shrink-0 transition-transform group-hover:scale-105 ${
                        isActive
                          ? "text-white"
                          : "text-text-secondary group-hover:text-text-primary"
                      }`}
                    />

                    {!isCollapsed && (
                      <>
                        <span className="truncate flex-1 text-left">
                          {item.label}
                        </span>
                        {item.badge && (
                          <span
                            className={`text-[9px] font-mono font-semibold px-1.5 py-0.2 rounded border ${
                              isActive
                                ? "bg-white/20 text-white border-white/30"
                                : getBadgeClass(item.badgeVariant)
                            }`}
                          >
                            {item.badge}
                          </span>
                        )}
                      </>
                    )}

                    {/* Active Indicator Bar on Collapsed Mode */}
                    {isCollapsed && isActive && (
                      <span className="absolute left-0 top-1.5 bottom-1.5 w-1 bg-accent-primary rounded-r" />
                    )}
                  </button>
                );
              })}
            </div>
          );
        })}
      </div>

      {/* Collapse / Expand Toggle Footer */}
      {onToggleCollapse && (
        <div className="p-2 border-t border-border-subtle/60 flex items-center justify-between">
          {!isCollapsed && (
            <div className="px-2 text-[10px] text-text-secondary/70 font-mono">
              Debian 13 / Ubuntu 26
            </div>
          )}
          <button
            onClick={onToggleCollapse}
            title={isCollapsed ? "Expand Sidebar" : "Collapse Sidebar"}
            className={`flex items-center justify-center h-8 w-8 rounded-lg border border-border-subtle bg-bg-elevated text-text-secondary hover:text-text-primary hover:border-accent-primary/50 transition-colors ${
              isCollapsed ? "mx-auto" : ""
            }`}
          >
            {isCollapsed ? (
              <ChevronRight className="h-4 w-4" />
            ) : (
              <ChevronLeft className="h-4 w-4" />
            )}
          </button>
        </div>
      )}
    </aside>
  );
};
