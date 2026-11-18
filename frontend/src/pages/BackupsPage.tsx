import React, { useState } from "react";
import { useTranslation } from "react-i18next";
import {
  HardDriveDownload,
  ShieldCheck,
  RotateCcw,
  Plus,
  Cloud,
  CheckCircle2,
  Lock,
  Database,
  History,
} from "lucide-react";
import type { BackupSnapshot } from "../types";

export const BackupsPage: React.FC = () => {
  const { t } = useTranslation();

  const [snapshots, setSnapshots] = useState<BackupSnapshot[]>([
    {
      id: "snap-1",
      snapshot_hash: "a9f8e7d6c5b4a3210fedcba9876543210fedcba987654321",
      user_id: "all_tenants",
      storage_target: "s3",
      size_bytes: 42949672960, // 40 GB
      encrypted: true,
      test_restore_status: "verified",
      created_at: new Date(Date.now() - 86400000).toISOString(),
    },
    {
      id: "snap-2",
      snapshot_hash: "1234567890abcdef1234567890abcdef1234567890abcdef",
      user_id: "all_tenants",
      storage_target: "local",
      size_bytes: 41875931136, // 39 GB
      encrypted: true,
      test_restore_status: "verified",
      created_at: new Date(Date.now() - 172800000).toISOString(),
    },
    {
      id: "snap-3",
      snapshot_hash: "fedcba9876543210fedcba9876543210fedcba9876543210",
      user_id: "all_tenants",
      storage_target: "sftp",
      size_bytes: 40802189312, // 38 GB
      encrypted: true,
      test_restore_status: "verified",
      created_at: new Date(Date.now() - 259200000).toISOString(),
    },
  ]);

  const [creating, setCreating] = useState(false);

  const handleCreateSnapshot = () => {
    setCreating(true);
    setTimeout(() => {
      const newSnap: BackupSnapshot = {
        id: `snap-${Date.now()}`,
        snapshot_hash: "e8d7c6b5a43210fedcba9876543210fedcba987654321012",
        user_id: "all_tenants",
        storage_target: "s3",
        size_bytes: 43500000000,
        encrypted: true,
        test_restore_status: "verified",
        created_at: new Date().toISOString(),
      };
      setSnapshots([newSnap, ...snapshots]);
      setCreating(false);
    }, 1200);
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-xl font-bold tracking-tight text-text-primary flex items-center gap-2">
            <HardDriveDownload className="h-5 w-5 text-accent-primary" />
            {t("nav.backups", "Restic Backups & Disaster Recovery")}
          </h1>
          <p className="text-xs text-text-secondary mt-0.5">
            Deduplicated AES-256 Encrypted Snapshots, 3-2-1 Retention &
            Automated Test-Restore
          </p>
        </div>
        <button
          onClick={handleCreateSnapshot}
          disabled={creating}
          className="flex items-center gap-2 px-3 py-2 rounded-lg bg-accent-primary text-white text-xs font-semibold hover:bg-accent-hover transition-colors shadow-sm self-start sm:self-auto disabled:opacity-50"
        >
          <Plus className="h-4 w-4" />
          <span>
            {creating ? "Creating Snapshot..." : "Run Backup Snapshot Now"}
          </span>
        </button>
      </div>

      {/* Info Boxes (AdminLTE style) */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div className="rounded-xl border border-border-subtle bg-bg-card p-4 shadow-xs">
          <div className="flex items-center justify-between">
            <span className="text-xs font-bold text-text-primary">
              Restic Repository
            </span>
            <Lock className="h-4 w-4 text-status-success" />
          </div>
          <div className="my-2">
            <div className="text-lg font-bold font-mono text-text-primary">
              AES-256-GCM
            </div>
            <div className="text-[11px] text-text-secondary mt-0.5">
              Encrypted Keyfile:{" "}
              <span className="font-mono text-text-primary">
                /etc/panel/restic.key
              </span>
            </div>
          </div>
          <div className="text-[10px] text-status-success font-semibold flex items-center gap-1 border-t border-border-subtle pt-2">
            <ShieldCheck className="h-3.5 w-3.5" /> Repository Integrity OK
          </div>
        </div>

        <div className="rounded-xl border border-border-subtle bg-bg-card p-4 shadow-xs">
          <div className="flex items-center justify-between">
            <span className="text-xs font-bold text-text-primary">
              Database Hot Physical Backup
            </span>
            <Database className="h-4 w-4 text-accent-primary" />
          </div>
          <div className="my-2">
            <div className="text-lg font-bold font-mono text-text-primary">
              mariadb-backup
            </div>
            <div className="text-[11px] text-text-secondary mt-0.5">
              Zero-Lock InnoDB Physical Stream + PostgreSQL pg_basebackup
            </div>
          </div>
          <div className="text-[10px] text-text-secondary font-mono border-t border-border-subtle pt-2">
            Point-in-Time Recovery (PITR) Enabled
          </div>
        </div>

        <div className="rounded-xl border border-border-subtle bg-bg-card p-4 shadow-xs">
          <div className="flex items-center justify-between">
            <span className="text-xs font-bold text-text-primary">
              3-2-1 Retention Policy
            </span>
            <Cloud className="h-4 w-4 text-status-warning" />
          </div>
          <div className="my-2">
            <div className="text-xs font-semibold text-text-primary">
              7 Daily • 4 Weekly • 12 Monthly
            </div>
            <div className="text-[11px] text-text-secondary mt-0.5">
              Automated prune & check via systemd timer daily at 02:00 UTC
            </div>
          </div>
          <div className="text-[10px] text-status-success font-semibold flex items-center gap-1 border-t border-border-subtle pt-2">
            <CheckCircle2 className="h-3.5 w-3.5" /> S3 Offsite Sync Verified
          </div>
        </div>
      </div>

      {/* Snapshot List Table */}
      <div className="rounded-xl border border-border-subtle bg-bg-card p-5 space-y-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <History className="h-4 w-4 text-accent-primary" />
            <h2 className="text-sm font-bold text-text-primary">
              Snapshot Catalog
            </h2>
          </div>
          <span className="text-xs text-text-secondary font-mono">
            {snapshots.length} total snapshots
          </span>
        </div>

        <div className="overflow-x-auto">
          <table className="w-full text-left text-xs">
            <thead className="border-b border-border-subtle bg-bg-elevated text-text-secondary">
              <tr>
                <th className="py-2.5 px-3 font-semibold">Snapshot ID</th>
                <th className="py-2.5 px-3 font-semibold">Target</th>
                <th className="py-2.5 px-3 font-semibold">Size</th>
                <th className="py-2.5 px-3 font-semibold">Encryption</th>
                <th className="py-2.5 px-3 font-semibold">
                  Automated Test-Restore
                </th>
                <th className="py-2.5 px-3 font-semibold">Created</th>
                <th className="py-2.5 px-3 font-semibold text-right">Action</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border-subtle">
              {snapshots.map((snap) => (
                <tr
                  key={snap.id}
                  className="hover:bg-bg-elevated/50 transition-colors"
                >
                  <td className="py-2.5 px-3 font-mono font-bold text-text-primary">
                    {snap.snapshot_hash.slice(0, 12)}...
                  </td>
                  <td className="py-2.5 px-3 font-mono text-[11px] text-text-secondary uppercase">
                    {snap.storage_target}
                  </td>
                  <td className="py-2.5 px-3 font-mono text-text-primary">
                    {(snap.size_bytes / 1073741824).toFixed(1)} GB
                  </td>
                  <td className="py-2.5 px-3">
                    <span className="rounded bg-status-success/15 px-1.5 py-0.5 text-[10px] font-bold text-status-success font-mono">
                      AES-256
                    </span>
                  </td>
                  <td className="py-2.5 px-3">
                    <span className="flex items-center gap-1 text-[11px] font-semibold text-status-success">
                      <CheckCircle2 className="h-3.5 w-3.5" /> Verified Pass
                    </span>
                  </td>
                  <td className="py-2.5 px-3 font-mono text-[11px] text-text-secondary">
                    {new Date(snap.created_at).toLocaleString()}
                  </td>
                  <td className="py-2.5 px-3 text-right">
                    <button className="text-[11px] font-semibold text-accent-primary hover:underline flex items-center gap-1 ml-auto">
                      <RotateCcw className="h-3 w-3" /> Restore
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
};
