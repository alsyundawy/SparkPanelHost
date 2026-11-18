import React, { useState, useRef, useEffect } from "react";
import { useTranslation } from "react-i18next";
import {
  Sun,
  Moon,
  Languages,
  Search,
  Shield,
  LogOut,
  User,
  PanelLeft,
  Bell,
  CheckCircle2,
  AlertTriangle,
  Info,
  ChevronRight,
  Activity,
} from "lucide-react";
import type { User as UserType, NotificationItem } from "../types";

interface NavbarProps {
  user: UserType | null;
  onLogout: () => void;
  onOpenCommandPalette: () => void;
  isDark: boolean;
  onToggleTheme: () => void;
  currentTab: string;
  isSidebarCollapsed?: boolean;
  onToggleSidebar?: () => void;
}

export const Navbar: React.FC<NavbarProps> = ({
  user,
  onLogout,
  onOpenCommandPalette,
  isDark,
  onToggleTheme,
  currentTab,
  isSidebarCollapsed = false,
  onToggleSidebar,
}) => {
  const { t, i18n } = useTranslation();
  const [notificationsOpen, setNotificationsOpen] = useState(false);
  const notifRef = useRef<HTMLDivElement>(null);

  const [notifications, setNotifications] = useState<NotificationItem[]>([
    {
      id: "notif-1",
      title: "Coraza WAF Security Event",
      description:
        "Blocked SQLi attack from 198.51.100.42 targeting /wp-admin (CRS 942100)",
      timestamp: "5m ago",
      read: false,
      type: "security",
    },
    {
      id: "notif-2",
      title: "SSL Certificate Auto-Renewed",
      description:
        "Let's Encrypt wildcard certificate for *.sparkpanel.net renewed via ACME DNS-01",
      timestamp: "42m ago",
      read: false,
      type: "system",
    },
    {
      id: "notif-3",
      title: "MariaDB Snapshot Verified",
      description:
        "Physical hot backup to /var/panel/backups completed & SHA-256 verified",
      timestamp: "2h ago",
      read: true,
      type: "backup",
    },
  ]);

  const unreadCount = notifications.filter((n) => !n.read).length;

  // Close notifications dropdown on outside click
  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (notifRef.current && !notifRef.current.contains(e.target as Node)) {
        setNotificationsOpen(false);
      }
    };
    if (notificationsOpen) {
      document.addEventListener("mousedown", handleClickOutside);
    }
    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, [notificationsOpen]);

  const toggleLanguage = () => {
    const nextLang = i18n.language === "id" ? "en" : "id";
    i18n.changeLanguage(nextLang);
  };

  const getBreadcrumbLabel = (tab: string) => {
    switch (tab) {
      case "overview":
        return {
          category: "Core",
          title: t("nav.overview", "Server Overview"),
        };
      case "accounts":
        return {
          category: "Management",
          title: t("nav.accounts", "Hosting Accounts"),
        };
      case "packages":
        return {
          category: "Management",
          title: t("nav.packages", "Packages & Feature Lists"),
        };
      case "domains":
        return {
          category: "Web Services",
          title: t("nav.domains", "Domains & Multi-PHP"),
        };
      case "databases":
        return {
          category: "Web Services",
          title: t("nav.databases", "Databases (MariaDB / Postgres)"),
        };
      case "dns":
        return {
          category: "Web Services",
          title: t("nav.dns", "DNS Zones & Clustering"),
        };
      case "apps":
        return {
          category: "Web Services",
          title: t("nav.apps", "1-Click App Installer"),
        };
      case "files":
        return {
          category: "Web Services",
          title: t("nav.files", "File Manager"),
        };
      case "email":
        return {
          category: "Communication",
          title: t("nav.email", "Email Hosting & Deliverability"),
        };
      case "terminal":
        return {
          category: "Access",
          title: t("nav.terminal", "Web SSH Terminal"),
        };
      case "security":
        return {
          category: "Security & Ops",
          title: t("nav.security", "Coraza WAF & CrowdSec"),
        };
      case "backups":
        return {
          category: "Security & Ops",
          title: t("nav.backups", "Restic Backups & DR"),
        };
      case "config-editor":
        return { category: "Security & Ops", title: "Service Config Editor" };
      case "audit-logs":
        return { category: "Security & Ops", title: "Immutable Audit Log" };
      default:
        return { category: "Platform", title: tab };
    }
  };

  const breadcrumb = getBreadcrumbLabel(currentTab);

  return (
    <header className="sticky top-0 z-40 w-full border-b border-border-subtle bg-bg-card/85 backdrop-blur-md">
      <div className="flex h-14 items-center justify-between px-3 sm:px-6">
        {/* Left Section: Sidebar Toggle & Brand & Breadcrumb */}
        <div className="flex items-center gap-3">
          {/* Sidebar Toggle Button */}
          {onToggleSidebar && (
            <button
              onClick={onToggleSidebar}
              title={isSidebarCollapsed ? "Expand Sidebar" : "Collapse Sidebar"}
              className="flex h-8 w-8 items-center justify-center rounded-lg border border-border-subtle bg-bg-elevated text-text-secondary hover:text-text-primary hover:border-accent-primary/50 transition-colors"
            >
              <PanelLeft className="h-4 w-4" />
            </button>
          )}

          {/* Brand Logo & Title */}
          <div className="flex items-center gap-2.5">
            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-accent-primary text-white shadow-sm font-bold">
              <Shield className="h-4 w-4" />
            </div>
            <div className="hidden sm:flex items-center gap-2">
              <span className="font-bold tracking-tight text-text-primary text-base">
                SparkPanelHost
              </span>
              <span className="rounded bg-accent-primary/10 px-1.5 py-0.2 text-[9px] font-semibold text-accent-primary font-mono border border-accent-primary/20">
                v8.1
              </span>
            </div>
          </div>

          {/* Breadcrumbs (AdminLTE / Bootstrap inspiration) */}
          <div className="hidden md:flex items-center gap-1.5 pl-3 border-l border-border-subtle text-xs text-text-secondary">
            <span className="text-text-secondary/70">
              {breadcrumb.category}
            </span>
            <ChevronRight className="h-3 w-3 text-text-secondary/50" />
            <span className="font-semibold text-text-primary truncate max-w-[180px]">
              {breadcrumb.title}
            </span>
          </div>
        </div>

        {/* Center: Command Palette Trigger */}
        <button
          onClick={onOpenCommandPalette}
          className="hidden lg:flex items-center gap-2 rounded-lg border border-border-subtle bg-bg-elevated px-3 py-1.5 text-xs text-text-secondary hover:border-accent-primary/50 transition-colors w-64 xl:w-80 justify-between shadow-2xs"
        >
          <span className="flex items-center gap-2">
            <Search className="h-3.5 w-3.5 text-text-secondary/80" />
            <span className="truncate">
              {t("app.command_placeholder", "Search or run command...")}
            </span>
          </span>
          <kbd className="rounded border border-border-subtle bg-bg-canvas px-1.5 py-0.5 font-mono text-[10px] text-text-secondary">
            ⌘K
          </kbd>
        </button>

        {/* Right Section: System Status, Notifications, Language, Theme, User Profile */}
        <div className="flex items-center gap-2 sm:gap-2.5">
          {/* Live Server Health Pill */}
          <div className="hidden sm:flex items-center gap-1.5 rounded-full border border-status-success/30 bg-status-success/10 px-2.5 py-1 text-[11px] font-medium text-status-success">
            <span className="h-2 w-2 rounded-full bg-status-success animate-pulse" />
            <span className="font-mono font-semibold">Healthy</span>
            <span className="text-text-secondary/60 text-[10px]">|</span>
            <span className="text-[10px] font-mono text-text-secondary">
              Load 0.14
            </span>
          </div>

          {/* Notifications Center (AdminLTE inspiration) */}
          <div className="relative" ref={notifRef}>
            <button
              onClick={() => setNotificationsOpen(!notificationsOpen)}
              title="System & Security Alerts"
              className="relative flex h-8 w-8 items-center justify-center rounded-lg border border-border-subtle bg-bg-elevated text-text-secondary hover:text-text-primary hover:border-accent-primary/50 transition-colors"
            >
              <Bell className="h-4 w-4" />
              {unreadCount > 0 && (
                <span className="absolute -top-1 -right-1 flex h-4 w-4 items-center justify-center rounded-full bg-status-danger text-[9px] font-bold text-white shadow-xs">
                  {unreadCount}
                </span>
              )}
            </button>

            {/* Notifications Dropdown Popover */}
            {notificationsOpen && (
              <div className="absolute right-0 mt-2 w-80 sm:w-96 rounded-xl border border-border-subtle bg-bg-card p-3 shadow-xl z-50 animate-in fade-in slide-in-from-top-1 duration-150">
                <div className="flex items-center justify-between pb-2 border-b border-border-subtle">
                  <div className="flex items-center gap-1.5">
                    <Activity className="h-4 w-4 text-accent-primary" />
                    <span className="text-xs font-bold text-text-primary">
                      System Notifications
                    </span>
                    <span className="rounded bg-accent-primary/10 px-1.5 py-0.2 text-[9px] font-semibold text-accent-primary">
                      {unreadCount} new
                    </span>
                  </div>
                  <button
                    onClick={() => {
                      setNotifications((prev) =>
                        prev.map((n) => ({ ...n, read: true })),
                      );
                    }}
                    className="text-[10px] text-accent-primary hover:underline font-semibold"
                  >
                    Mark all read
                  </button>
                </div>

                <div className="py-2 space-y-2 max-h-72 overflow-y-auto">
                  {notifications.map((notif) => (
                    <div
                      key={notif.id}
                      className={`p-2.5 rounded-lg border text-xs transition-colors ${
                        notif.read
                          ? "border-border-subtle/50 bg-bg-elevated/40"
                          : "border-accent-primary/30 bg-accent-primary/5"
                      }`}
                    >
                      <div className="flex items-start justify-between gap-2">
                        <div className="flex items-center gap-1.5 font-bold text-text-primary">
                          {notif.type === "security" ? (
                            <AlertTriangle className="h-3.5 w-3.5 text-status-warning shrink-0" />
                          ) : notif.type === "backup" ? (
                            <CheckCircle2 className="h-3.5 w-3.5 text-status-success shrink-0" />
                          ) : (
                            <Info className="h-3.5 w-3.5 text-accent-primary shrink-0" />
                          )}
                          <span className="truncate">{notif.title}</span>
                        </div>
                        <span className="text-[10px] text-text-secondary/70 font-mono shrink-0">
                          {notif.timestamp}
                        </span>
                      </div>
                      <p className="text-[11px] text-text-secondary mt-1 leading-relaxed">
                        {notif.description}
                      </p>
                    </div>
                  ))}
                </div>

                <div className="pt-2 border-t border-border-subtle text-center">
                  <span className="text-[10px] text-text-secondary font-mono">
                    Audit Log Verified • Zero-Trust Mode
                  </span>
                </div>
              </div>
            )}
          </div>

          {/* Language Switcher */}
          <button
            onClick={toggleLanguage}
            title="Switch Language (ID / EN)"
            className="flex items-center gap-1.5 rounded-lg border border-border-subtle bg-bg-elevated px-2 py-1.5 text-xs font-semibold text-text-secondary hover:text-text-primary hover:border-accent-primary/50 transition-colors"
          >
            <Languages className="h-3.5 w-3.5" />
            <span className="uppercase font-mono text-[11px]">
              {i18n.language}
            </span>
          </button>

          {/* Theme Toggle (Dark / Light WCAG AAA) */}
          <button
            onClick={onToggleTheme}
            title={
              isDark ? "Light Mode (WCAG AAA High Contrast)" : "OLED Dark Mode"
            }
            className="flex h-8 w-8 items-center justify-center rounded-lg border border-border-subtle bg-bg-elevated text-text-secondary hover:text-text-primary hover:border-accent-primary/50 transition-colors"
          >
            {isDark ? (
              <Sun className="h-4 w-4" />
            ) : (
              <Moon className="h-4 w-4" />
            )}
          </button>

          {/* User Profile Badge */}
          {user && (
            <div className="flex items-center gap-2 border-l border-border-subtle pl-2 sm:pl-3">
              <div className="flex h-8 w-8 items-center justify-center rounded-full bg-accent-primary/10 text-accent-primary border border-accent-primary/20">
                <User className="h-4 w-4" />
              </div>
              <div className="hidden sm:block text-left text-xs">
                <div className="font-semibold text-text-primary flex items-center gap-1.5">
                  {user.username}
                  <span className="rounded bg-accent-primary/15 px-1 py-0.2 text-[9px] font-bold uppercase text-accent-primary border border-accent-primary/30">
                    {user.role}
                  </span>
                </div>
                <div className="text-[10px] text-text-secondary truncate max-w-[120px]">
                  {user.email}
                </div>
              </div>

              <button
                onClick={onLogout}
                title="Sign Out"
                className="ml-1 p-1.5 rounded-md text-text-secondary hover:text-status-danger hover:bg-status-danger/10 transition-colors"
              >
                <LogOut className="h-4 w-4" />
              </button>
            </div>
          )}
        </div>
      </div>
    </header>
  );
};
