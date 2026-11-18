import { describe, it, expect } from 'vitest';
import { canAccessFeature, isPhpVersionEOL, useFeatureGate, type SessionContext } from './useFeatureGate';

describe('useFeatureGate', () => {
  it('allows superadmin to access any feature regardless of allowedFeatures list', () => {
    const adminSession: SessionContext = {
      role: 'superadmin',
      allowedFeatures: [],
      shellAccess: false,
    };
    expect(canAccessFeature(adminSession, 'databases')).toBe(true);
    expect(canAccessFeature(adminSession, 'web_terminal')).toBe(true);
  });

  it('denies web_terminal to tenants without shellAccess', () => {
    const tenantSession: SessionContext = {
      role: 'tenant',
      allowedFeatures: ['web_terminal'],
      shellAccess: false,
    };
    expect(canAccessFeature(tenantSession, 'web_terminal')).toBe(false);
  });

  it('allows web_terminal to tenants with shellAccess and feature enabled', () => {
    const tenantSession: SessionContext = {
      role: 'tenant',
      allowedFeatures: ['web_terminal'],
      shellAccess: true,
    };
    expect(canAccessFeature(tenantSession, 'web_terminal')).toBe(true);
  });

  it('checks feature presence for tenants', () => {
    const tenantSession: SessionContext = {
      role: 'tenant',
      allowedFeatures: ['domains', 'email'],
      shellAccess: false,
    };
    expect(canAccessFeature(tenantSession, 'domains')).toBe(true);
    expect(canAccessFeature(tenantSession, 'email')).toBe(true);
    expect(canAccessFeature(tenantSession, 'postgresql')).toBe(false);
  });

  it('identifies PHP 8.2 as EOL', () => {
    expect(isPhpVersionEOL('8.2')).toBe(true);
    expect(isPhpVersionEOL('8.3')).toBe(false);
    expect(isPhpVersionEOL('8.4')).toBe(false);
    expect(isPhpVersionEOL('8.5')).toBe(false);
    expect(isPhpVersionEOL('8.6')).toBe(false);
  });

  it('provides helper role checks in useFeatureGate hook', () => {
    const session: SessionContext = {
      role: 'reseller',
      allowedFeatures: ['domains'],
      shellAccess: true,
    };
    const gate = useFeatureGate(session);
    expect(gate.isSuperAdmin).toBe(false);
    expect(gate.isReseller).toBe(true);
    expect(gate.isTenant).toBe(false);
    expect(gate.canAccess('domains')).toBe(true);
    expect(gate.canAccess('postgresql')).toBe(false);
  });
});
