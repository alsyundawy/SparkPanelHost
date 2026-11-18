export type UserRole = 'superadmin' | 'reseller' | 'tenant';

export interface User {
  id: string;
  username: string;
  email: string;
  role: UserRole;
  reseller_id?: string;
  package_id?: string;
  linux_uid: number;
  linux_gid: number;
  locale: 'id' | 'en';
  theme: 'dark' | 'light';
  status: 'active' | 'suspended';
  created_at: string;
}

export interface Domain {
  id: string;
  user_id: string;
  domain_name: string;
  document_root: string;
  php_version: string;
  ssl_enabled: boolean;
  ssl_provider?: string;
  force_https: boolean;
  waf_mode: 'Blocking' | 'DetectionOnly' | 'Off';
  waf_profile: string;
  nginx_template: string;
  created_at: string;
}

export interface TenantDatabase {
  id: string;
  user_id: string;
  engine: 'mariadb' | 'postgresql';
  db_name: string;
  db_user: string;
  charset: string;
  size_mb: number;
  created_at: string;
}

export interface DNSCluster {
  id: string;
  name: string;
  topology: 'star' | 'primary_secondary' | 'galera' | 'anycast';
  primary_node_id?: string;
  created_at: string;
}

export interface DNSClusterNode {
  id: string;
  cluster_id: string;
  hostname: string;
  api_endpoint: string;
  role: 'write_only' | 'synchronize' | 'standalone' | 'dns_only';
  is_active: boolean;
  last_seen_at?: string;
  created_at: string;
}

export interface MailDomain {
  id: string;
  domain_id: string;
  dkim_selector: string;
  dkim_rsa_public: string;
  dkim_ed25519_public?: string;
  mta_sts_mode: 'none' | 'testing' | 'enforce';
  catchall_address?: string;
  outbound_limit_hour: number;
  created_at: string;
}

export interface AppRecipe {
  id: string;
  name: string;
  category: string;
  version: string;
  description: string;
  archive_url: string;
  sha256: string;
}

export interface TelemetryData {
  cpu_percent: number;
  ram_used_mb: number;
  ram_total_mb: number;
  disk_used_mb: number;
  disk_total_mb: number;
  waf_blocked_24h: number;
  crowdsec_active_bans: number;
}
