export type UserRole = "superadmin" | "reseller" | "tenant";

export type FeatureKey =
  | "domains"
  | "apps"
  | "nodejs"
  | "python"
  | "git_deploy"
  | "databases"
  | "postgresql"
  | "dns"
  | "dnssec"
  | "email"
  | "file_manager"
  | "web_terminal"
  | "cron_jobs"
  | "ftp_accounts"
  | "waf_exclusions"
  | "backups";

export interface SessionContext {
  role: UserRole;
  allowedFeatures: FeatureKey[];
  shellAccess: boolean;
}

export function canAccessFeature(
  session: SessionContext,
  feature: FeatureKey,
): boolean {
  if (session.role === "superadmin") {
    return true;
  }
  if (feature === "web_terminal" && !session.shellAccess) {
    return false;
  }
  return session.allowedFeatures.includes(feature);
}

export function isPhpVersionEOL(phpVersion: string): boolean {
  return phpVersion === "8.2";
}

export function useFeatureGate(session: SessionContext) {
  return {
    canAccess: (feature: FeatureKey) => canAccessFeature(session, feature),
    isPHP82EOL: (version: string) => isPhpVersionEOL(version),
    isSuperAdmin: session.role === "superadmin",
    isReseller: session.role === "reseller",
    isTenant: session.role === "tenant",
  };
}
