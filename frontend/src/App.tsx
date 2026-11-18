import React, { useState, useEffect } from 'react';
import { Navbar } from './components/Navbar';
import { Sidebar } from './components/Sidebar';
import { ImpersonationBanner } from './components/ImpersonationBanner';
import { CommandPalette } from './components/CommandPalette';
import { OverviewPage } from './pages/OverviewPage';
import { DomainsPage } from './pages/DomainsPage';
import { DNSClusterPage } from './pages/DNSClusterPage';
import { AppInstallerPage } from './pages/AppInstallerPage';
import { DatabasesPage } from './pages/DatabasesPage';
import { EmailPage } from './pages/EmailPage';
import { AccountsPage } from './pages/AccountsPage';
import { FileManagerPage } from './pages/FileManagerPage';
import { WebTerminalPage } from './pages/WebTerminalPage';
import { AuditLogsPage } from './pages/AuditLogsPage';
import { ConfigEditorPage } from './pages/ConfigEditorPage';
import type { User } from './types';
import { Shield, Lock } from 'lucide-react';

export const App: React.FC = () => {
  const [user, setUser] = useState<User | null>(null);
  const [impersonated, setImpersonated] = useState(false);
  const [currentTab, setCurrentTab] = useState('overview');
  const [isDark, setIsDark] = useState(true);
  const [isCommandPaletteOpen, setIsCommandPaletteOpen] = useState(false);

  // Login form state
  const [loginUsername, setLoginUsername] = useState('admin');
  const [loginPassword, setLoginPassword] = useState('AetherAdmin@2026!');
  const [loginError, setLoginError] = useState('');
  const [loginLoading, setLoginLoading] = useState(false);

  useEffect(() => {
    // Check dark mode
    if (isDark) {
      document.documentElement.classList.add('dark');
    } else {
      document.documentElement.classList.remove('dark');
    }
  }, [isDark]);

  useEffect(() => {
    // Check existing session
    fetch('/api/v1/auth/me')
      .then((res) => {
        if (res.ok) return res.json();
        throw new Error('Not logged in');
      })
      .then((data) => {
        if (data.user) {
          setUser(data.user);
          setImpersonated(data.impersonated || false);
        }
      })
      .catch(() => {
        // Fallback default demo user
        setUser({
          id: 'usr-admin',
          username: 'admin',
          email: 'admin@aetherpanel.local',
          role: 'superadmin',
          linux_uid: 0,
          linux_gid: 0,
          locale: 'id',
          theme: 'dark',
          status: 'active',
          created_at: new Date().toISOString(),
        });
      });
  }, []);

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoginLoading(true);
    setLoginError('');
    try {
      const res = await fetch('/api/v1/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username: loginUsername, password: loginPassword }),
      });
      const data = await res.json();
      if (res.ok && data.user) {
        setUser(data.user);
      } else {
        setLoginError(data.error || 'Invalid credentials');
      }
    } catch {
      setLoginError('Could not connect to backend server');
    } finally {
      setLoginLoading(false);
    }
  };

  const handleLogout = async () => {
    await fetch('/api/v1/auth/logout', { method: 'POST' });
    setUser(null);
  };

  const handleImpersonate = async (targetUser: User) => {
    try {
      const res = await fetch('/api/v1/auth/impersonate', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ target_user_id: targetUser.id }),
      });
      if (res.ok) {
        const data = await res.json();
        setUser(data.user);
        setImpersonated(true);
      }
    } catch {
      // Demo fallback
      setUser(targetUser);
      setImpersonated(true);
    }
  };

  const handleExitImpersonation = () => {
    setImpersonated(false);
    // Restore admin
    setUser({
      id: 'usr-admin',
      username: 'admin',
      email: 'admin@aetherpanel.local',
      role: 'superadmin',
      linux_uid: 0,
      linux_gid: 0,
      locale: 'id',
      theme: 'dark',
      status: 'active',
      created_at: new Date().toISOString(),
    });
  };

  if (!user) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-bg-canvas p-4">
        <div className="w-full max-w-sm rounded-2xl border border-border-subtle bg-bg-card p-8 shadow-2xl space-y-6">
          <div className="text-center space-y-2">
            <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-xl bg-accent-primary text-white shadow-md">
              <Shield className="h-6 w-6" />
            </div>
            <h1 className="text-xl font-bold tracking-tight text-text-primary">
              SparkPanelHost v0.1.0
            </h1>
            <p className="text-xs text-text-secondary">
              All-in-One Linux Web Hosting Control Panel
            </p>
          </div>

          {loginError && (
            <div className="rounded-lg bg-status-danger/10 border border-status-danger/20 p-2.5 text-xs text-status-danger text-center font-medium">
              {loginError}
            </div>
          )}

          <form onSubmit={handleLogin} className="space-y-4 text-xs">
            <div>
              <label className="block font-semibold text-text-secondary mb-1">Username</label>
              <input
                type="text"
                required
                value={loginUsername}
                onChange={(e) => setLoginUsername(e.target.value)}
                className="w-full rounded-lg border border-border-subtle bg-bg-elevated px-3 py-2 text-text-primary focus:outline-none focus:border-accent-primary font-mono"
              />
            </div>

            <div>
              <label className="block font-semibold text-text-secondary mb-1">Password</label>
              <input
                type="password"
                required
                value={loginPassword}
                onChange={(e) => setLoginPassword(e.target.value)}
                className="w-full rounded-lg border border-border-subtle bg-bg-elevated px-3 py-2 text-text-primary focus:outline-none focus:border-accent-primary font-mono"
              />
            </div>

            <button
              type="submit"
              disabled={loginLoading}
              className="w-full flex items-center justify-center gap-2 rounded-lg bg-accent-primary py-2.5 font-semibold text-white hover:bg-accent-hover transition-colors shadow-sm"
            >
              <Lock className="h-3.5 w-3.5" />
              <span>{loginLoading ? 'Authenticating...' : 'Sign In to Panel'}</span>
            </button>
          </form>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen flex flex-col bg-bg-canvas text-text-primary">
      {/* Top Navigation */}
      <Navbar
        user={user}
        onLogout={handleLogout}
        onOpenCommandPalette={() => setIsCommandPaletteOpen(true)}
        isDark={isDark}
        onToggleTheme={() => setIsDark(!isDark)}
      />

      {/* Impersonation Banner */}
      {impersonated && (
        <ImpersonationBanner
          username={user.username}
          onExit={handleExitImpersonation}
        />
      )}

      {/* Main Container */}
      <div className="flex-1 flex overflow-hidden">
        {/* Sidebar */}
        <Sidebar
          currentTab={currentTab}
          onSelectTab={setCurrentTab}
          role={user.role}
        />

        {/* Content View */}
        <main className="flex-1 overflow-y-auto p-4 sm:p-6 lg:p-8">
          {currentTab === 'overview' && <OverviewPage onNavigate={setCurrentTab} />}
          {currentTab === 'accounts' && <AccountsPage onImpersonate={handleImpersonate} />}
          {currentTab === 'domains' && <DomainsPage />}
          {currentTab === 'dns' && <DNSClusterPage />}
          {currentTab === 'apps' && <AppInstallerPage />}
          {currentTab === 'databases' && <DatabasesPage />}
          {currentTab === 'email' && <EmailPage />}
          {currentTab === 'files' && <FileManagerPage />}
          {currentTab === 'terminal' && <WebTerminalPage />}
          {currentTab === 'config-editor' && <ConfigEditorPage />}
          {currentTab === 'audit-logs' && <AuditLogsPage />}
        </main>
      </div>

      {/* Command Palette */}
      <CommandPalette
        isOpen={isCommandPaletteOpen}
        onClose={() => setIsCommandPaletteOpen(false)}
        onSelectAction={setCurrentTab}
      />
    </div>
  );
};
