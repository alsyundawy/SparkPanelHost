import React, { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { Users, Plus, UserCheck, ShieldAlert } from 'lucide-react';
import type { User } from '../types';

interface AccountsPageProps {
  onImpersonate: (user: User) => void;
}

export const AccountsPage: React.FC<AccountsPageProps> = ({ onImpersonate }) => {
  const { t } = useTranslation();
  const [users, setUsers] = useState<User[]>([]);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [newUser, setNewUser] = useState({
    username: '',
    email: '',
    password: '',
    role: 'tenant',
  });

  const fetchUsers = async () => {
    try {
      const res = await fetch('/api/v1/users');
      if (res.ok) {
        const data = await res.json();
        setUsers(data.users || []);
      }
    } catch {
      // Fallback
    }

    setUsers((prev) =>
      prev.length > 0
        ? prev
        : [
            {
              id: 'usr-1',
              username: 'alpha',
              email: 'alpha@example.com',
              role: 'tenant',
              linux_uid: 2001,
              linux_gid: 2001,
              locale: 'id',
              theme: 'dark',
              status: 'active',
              created_at: new Date().toISOString(),
            },
            {
              id: 'usr-2',
              username: 'beta',
              email: 'beta@example.com',
              role: 'tenant',
              linux_uid: 2002,
              linux_gid: 2002,
              locale: 'en',
              theme: 'dark',
              status: 'suspended',
              created_at: new Date().toISOString(),
            },
          ]
    );
  };

  useEffect(() => {
    fetchUsers();
  }, []);

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    const created: User = {
      id: `usr-${Date.now()}`,
      username: newUser.username,
      email: newUser.email,
      role: newUser.role as any,
      linux_uid: 2000 + Math.floor(Math.random() * 5000),
      linux_gid: 2000 + Math.floor(Math.random() * 5000),
      locale: 'id',
      theme: 'dark',
      status: 'active',
      created_at: new Date().toISOString(),
    };
    setUsers([...users, created]);
    setIsModalOpen(false);
    setNewUser({ username: '', email: '', password: '', role: 'tenant' });
  };

  const toggleSuspend = (id: string) => {
    setUsers(
      users.map((u) => {
        if (u.id === id) {
          return {
            ...u,
            status: u.status === 'active' ? 'suspended' : 'active',
          };
        }
        return u;
      })
    );
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-xl font-bold tracking-tight text-text-primary">
            {t('nav.accounts', 'Hosting Accounts')}
          </h1>
          <p className="text-xs text-text-secondary mt-0.5">
            Admin WHM Account Orchestration & Impersonation (Zero-Trust Isolated Tenants)
          </p>
        </div>
        <button
          onClick={() => setIsModalOpen(true)}
          className="flex items-center gap-2 rounded-lg bg-accent-primary px-3.5 py-2 text-xs font-semibold text-white shadow-xs hover:bg-accent-hover transition-colors self-start sm:self-auto"
        >
          <Plus className="h-4 w-4" />
          <span>Create Account</span>
        </button>
      </div>

      {/* Accounts Table */}
      <div className="rounded-xl border border-border-subtle bg-bg-card overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-left text-xs">
            <thead className="bg-bg-elevated border-b border-border-subtle text-text-secondary font-semibold">
              <tr>
                <th className="px-4 py-3">Username</th>
                <th className="px-4 py-3">Contact Email</th>
                <th className="px-4 py-3">Linux UID/GID</th>
                <th className="px-4 py-3">Role</th>
                <th className="px-4 py-3">Status</th>
                <th className="px-4 py-3 text-right">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border-subtle text-text-primary">
              {users.map((u) => (
                <tr key={u.id} className="hover:bg-bg-elevated/50 transition-colors">
                  <td className="px-4 py-3 font-semibold flex items-center gap-2">
                    <Users className="h-3.5 w-3.5 text-accent-primary" />
                    <span>{u.username}</span>
                  </td>
                  <td className="px-4 py-3 text-text-secondary">{u.email}</td>
                  <td className="px-4 py-3 font-mono text-[11px] text-text-secondary">
                    {u.linux_uid}:{u.linux_gid}
                  </td>
                  <td className="px-4 py-3">
                    <span className="rounded bg-accent-primary/10 px-2 py-0.5 text-[10px] font-bold text-accent-primary uppercase font-mono">
                      {u.role}
                    </span>
                  </td>
                  <td className="px-4 py-3">
                    <span
                      className={`rounded px-2 py-0.5 text-[10px] font-bold ${
                        u.status === 'active'
                          ? 'bg-status-success/15 text-status-success'
                          : 'bg-status-danger/15 text-status-danger'
                      }`}
                    >
                      {u.status}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-right">
                    <div className="flex items-center justify-end gap-1.5">
                      {/* Impersonate button */}
                      <button
                        onClick={() => onImpersonate(u)}
                        className="flex items-center gap-1 bg-accent-primary/10 hover:bg-accent-primary text-accent-primary hover:text-white px-2 py-1 rounded text-[11px] font-semibold transition-colors"
                        title="Login as User (Impersonate)"
                      >
                        <UserCheck className="h-3 w-3" />
                        <span>Login as User</span>
                      </button>
                      {/* Suspend button */}
                      <button
                        onClick={() => toggleSuspend(u.id)}
                        className="p-1 text-text-secondary hover:text-status-danger rounded transition-colors"
                        title={u.status === 'active' ? 'Suspend Account' : 'Unsuspend Account'}
                      >
                        <ShieldAlert className="h-4 w-4" />
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      {/* Create Account Modal */}
      {isModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
          <div className="w-full max-w-md rounded-xl border border-border-subtle bg-bg-card p-6 shadow-2xl">
            <h2 className="text-base font-bold text-text-primary mb-4 flex items-center gap-2">
              <Users className="h-5 w-5 text-accent-primary" />
              <span>Create New Hosting Account</span>
            </h2>

            <form onSubmit={handleCreate} className="space-y-4 text-xs">
              <div>
                <label className="block font-semibold text-text-secondary mb-1">
                  Linux Username (3-16 chars lowercase)
                </label>
                <input
                  type="text"
                  required
                  placeholder="acmeclient"
                  pattern="^[a-z][a-z0-9]{2,15}$"
                  value={newUser.username}
                  onChange={(e) => setNewUser({ ...newUser, username: e.target.value.toLowerCase() })}
                  className="w-full rounded-lg border border-border-subtle bg-bg-elevated px-3 py-2 text-text-primary focus:outline-none focus:border-accent-primary font-mono"
                />
              </div>

              <div>
                <label className="block font-semibold text-text-secondary mb-1">Contact Email</label>
                <input
                  type="email"
                  required
                  placeholder="client@example.com"
                  value={newUser.email}
                  onChange={(e) => setNewUser({ ...newUser, email: e.target.value })}
                  className="w-full rounded-lg border border-border-subtle bg-bg-elevated px-3 py-2 text-text-primary focus:outline-none focus:border-accent-primary"
                />
              </div>

              <div>
                <label className="block font-semibold text-text-secondary mb-1">Password</label>
                <input
                  type="password"
                  required
                  placeholder="Strong account password..."
                  value={newUser.password}
                  onChange={(e) => setNewUser({ ...newUser, password: e.target.value })}
                  className="w-full rounded-lg border border-border-subtle bg-bg-elevated px-3 py-2 text-text-primary focus:outline-none focus:border-accent-primary font-mono"
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
                  Create Account
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};
