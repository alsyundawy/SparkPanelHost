import React from 'react';
import { useTranslation } from 'react-i18next';
import { Sun, Moon, Languages, Search, Shield, LogOut, User } from 'lucide-react';
import type { User as UserType } from '../types';

interface NavbarProps {
  user: UserType | null;
  onLogout: () => void;
  onOpenCommandPalette: () => void;
  isDark: boolean;
  onToggleTheme: () => void;
}

export const Navbar: React.FC<NavbarProps> = ({
  user,
  onLogout,
  onOpenCommandPalette,
  isDark,
  onToggleTheme,
}) => {
  const { t, i18n } = useTranslation();

  const toggleLanguage = () => {
    const nextLang = i18n.language === 'id' ? 'en' : 'id';
    i18n.changeLanguage(nextLang);
  };

  return (
    <header className="sticky top-0 z-40 w-full border-b border-border-subtle bg-bg-card/80 backdrop-blur-md">
      <div className="flex h-14 items-center justify-between px-4 sm:px-6">
        {/* Brand */}
        <div className="flex items-center gap-3">
          <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-accent-primary text-white shadow-sm font-bold">
            <Shield className="h-5 w-5" />
          </div>
          <div>
            <div className="flex items-center gap-2">
              <span className="font-bold tracking-tight text-text-primary text-lg">
                SparkPanelHost
              </span>
              <span className="rounded bg-accent-primary/10 px-1.5 py-0.5 text-[10px] font-semibold text-accent-primary font-mono">
                v8.1
              </span>
            </div>
          </div>
        </div>

        {/* Center: Command Palette Trigger */}
        <button
          onClick={onOpenCommandPalette}
          className="hidden md:flex items-center gap-2 rounded-lg border border-border-subtle bg-bg-elevated px-3 py-1.5 text-xs text-text-secondary hover:border-accent-primary/50 transition-colors w-72 justify-between"
        >
          <span className="flex items-center gap-2">
            <Search className="h-3.5 w-3.5" />
            <span className="truncate">{t('app.command_placeholder', 'Search or run command...')}</span>
          </span>
          <kbd className="rounded border border-border-subtle bg-bg-canvas px-1.5 py-0.5 font-mono text-[10px]">
            ⌘K
          </kbd>
        </button>

        {/* Right Tools */}
        <div className="flex items-center gap-2 sm:gap-3">
          {/* Language Switcher */}
          <button
            onClick={toggleLanguage}
            title="Switch Language"
            className="flex items-center gap-1.5 rounded-lg border border-border-subtle bg-bg-elevated px-2.5 py-1.5 text-xs font-semibold text-text-secondary hover:text-text-primary hover:border-accent-primary/50 transition-colors"
          >
            <Languages className="h-3.5 w-3.5" />
            <span className="uppercase font-mono">{i18n.language}</span>
          </button>

          {/* Theme Toggle */}
          <button
            onClick={onToggleTheme}
            title={isDark ? 'Light Mode (WCAG AAA)' : 'OLED Dark Mode'}
            className="flex h-8 w-8 items-center justify-center rounded-lg border border-border-subtle bg-bg-elevated text-text-secondary hover:text-text-primary hover:border-accent-primary/50 transition-colors"
          >
            {isDark ? <Sun className="h-4 w-4" /> : <Moon className="h-4 w-4" />}
          </button>

          {/* User Profile Badge */}
          {user && (
            <div className="flex items-center gap-2 border-l border-border-subtle pl-3">
              <div className="flex h-8 w-8 items-center justify-center rounded-full bg-accent-primary/10 text-accent-primary">
                <User className="h-4 w-4" />
              </div>
              <div className="hidden sm:block text-left text-xs">
                <div className="font-semibold text-text-primary flex items-center gap-1.5">
                  {user.username}
                  <span className="rounded bg-accent-primary/10 px-1 py-0.2 text-[9px] font-bold uppercase text-accent-primary">
                    {user.role}
                  </span>
                </div>
                <div className="text-[10px] text-text-secondary truncate max-w-[120px]">
                  {user.email}
                </div>
              </div>

              <button
                onClick={onLogout}
                title="Logout"
                className="ml-1 p-1 text-text-secondary hover:text-status-danger transition-colors"
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
