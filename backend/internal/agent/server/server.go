package server

import (
	"context"
	"fmt"
	"time"

	agentv1 "github.com/alsyundawy/SparkPanelHost/backend/gen/agent/v1"
	"github.com/alsyundawy/SparkPanelHost/backend/internal/agent/drivers"
	"github.com/alsyundawy/SparkPanelHost/backend/internal/agent/fileops"
)

type AgentServer struct {
	agentv1.UnimplementedSystemAgentServiceServer
}

func NewAgentServer() *AgentServer {
	return &AgentServer{}
}

func opResponse(err error, successMsg string) (*agentv1.OperationResponse, error) {
	if err != nil {
		return &agentv1.OperationResponse{
			Success: false,
			Message: err.Error(),
			AuditId: fmt.Sprintf("audit-%d", time.Now().UnixNano()),
		}, nil
	}
	return &agentv1.OperationResponse{
		Success: true,
		Message: successMsg,
		AuditId: fmt.Sprintf("audit-%d", time.Now().UnixNano()),
	}, nil
}

func (s *AgentServer) ProvisionTenant(ctx context.Context, req *agentv1.ProvisionTenantRequest) (*agentv1.OperationResponse, error) {
	limits := drivers.TenantLimits{
		DiskQuotaMB:     req.GetDiskQuotaMb(),
		InodeQuota:      req.GetInodeQuota(),
		CPUQuotaPercent: req.GetCpuQuotaPercent(),
		MemoryLimitMB:   req.GetMemoryLimitMb(),
		IOWeight:        req.GetIoWeight(),
		TasksMax:        req.GetTasksMax(),
		ShellAccess:     req.GetShellAccess(),
	}
	err := drivers.ProvisionTenant(ctx, req.GetUsername(), req.GetUid(), req.GetGid(), limits)
	return opResponse(err, fmt.Sprintf("Tenant %s provisioned successfully", req.GetUsername()))
}

func (s *AgentServer) UpdateTenantLimits(ctx context.Context, req *agentv1.UpdateTenantLimitsRequest) (*agentv1.OperationResponse, error) {
	limits := drivers.TenantLimits{
		DiskQuotaMB:     req.GetDiskQuotaMb(),
		InodeQuota:      req.GetInodeQuota(),
		CPUQuotaPercent: req.GetCpuQuotaPercent(),
		MemoryLimitMB:   req.GetMemoryLimitMb(),
		IOWeight:        req.GetIoWeight(),
		TasksMax:        req.GetTasksMax(),
	}
	err := drivers.UpdateTenantLimits(ctx, req.GetUsername(), 0, limits)
	return opResponse(err, fmt.Sprintf("Limits updated for tenant %s", req.GetUsername()))
}

func (s *AgentServer) SuspendTenant(ctx context.Context, req *agentv1.SuspendTenantRequest) (*agentv1.OperationResponse, error) {
	err := drivers.SuspendTenant(ctx, req.GetUsername(), req.GetSuspended())
	action := "suspended"
	if !req.GetSuspended() {
		action = "unsuspended"
	}
	return opResponse(err, fmt.Sprintf("Tenant %s %s successfully", req.GetUsername(), action))
}

func (s *AgentServer) SyncVirtualHost(ctx context.Context, req *agentv1.SyncVirtualHostRequest) (*agentv1.OperationResponse, error) {
	cfg := drivers.VHostConfig{
		Username:      req.GetUsername(),
		DomainName:    req.GetDomainName(),
		DocumentRoot:  req.GetDocumentRoot(),
		PHPVersion:    req.GetPhpVersion(),
		ProxyPort:     req.GetProxyPort(),
		SSLEnabled:    req.GetSslEnabled(),
		ForceHTTPS:    req.GetForceHttps(),
		WAFMode:       req.GetWafMode(),
		WAFProfile:    req.GetWafProfile(),
		NginxTemplate: req.GetNginxTemplate(),
	}
	err := drivers.SyncVirtualHost(ctx, cfg)
	return opResponse(err, fmt.Sprintf("Virtual host for %s synced successfully", req.GetDomainName()))
}

func (s *AgentServer) DeleteVirtualHost(ctx context.Context, req *agentv1.DeleteVirtualHostRequest) (*agentv1.OperationResponse, error) {
	err := drivers.DeleteVirtualHost(ctx, req.GetUsername(), req.GetDomainName())
	return opResponse(err, fmt.Sprintf("Virtual host for %s deleted successfully", req.GetDomainName()))
}

func (s *AgentServer) IssueSSLCertificate(ctx context.Context, req *agentv1.IssueSSLCertificateRequest) (*agentv1.OperationResponse, error) {
	err := drivers.IssueSSLCertificate(ctx, req.GetDomainName(), req.GetChallengeType(), req.GetWildcard())
	return opResponse(err, fmt.Sprintf("SSL certificate issued for %s", req.GetDomainName()))
}

func (s *AgentServer) ManageWAFExclusion(ctx context.Context, req *agentv1.ManageWAFExclusionRequest) (*agentv1.OperationResponse, error) {
	err := drivers.ManageWAFExclusion(ctx, req.GetDomainName(), req.GetAction(), req.GetRuleId(), req.GetUriPrefix())
	return opResponse(err, fmt.Sprintf("WAF exclusion %s for domain %s", req.GetAction(), req.GetDomainName()))
}

func (s *AgentServer) SyncPHPPool(ctx context.Context, req *agentv1.SyncPHPPoolRequest) (*agentv1.OperationResponse, error) {
	cfg := drivers.PHPPoolConfig{
		Username:    req.GetUsername(),
		PHPVersion:  req.GetPhpVersion(),
		MaxChildren: req.GetMaxChildren(),
		MemoryLimit: req.GetMemoryLimit(),
		CustomINI:   req.GetCustomIni(),
	}
	err := drivers.SyncPHPPool(ctx, cfg)
	return opResponse(err, fmt.Sprintf("PHP-FPM pool for %s (PHP %s) synced successfully", req.GetUsername(), req.GetPhpVersion()))
}

func (s *AgentServer) CreateDatabase(ctx context.Context, req *agentv1.CreateDatabaseRequest) (*agentv1.OperationResponse, error) {
	err := drivers.CreateDatabase(ctx, req.GetUsername(), req.GetEngine(), req.GetDbName(), req.GetDbUser(), req.GetDbPassword())
	return opResponse(err, fmt.Sprintf("Database %s created successfully on %s", req.GetDbName(), req.GetEngine()))
}

func (s *AgentServer) BackupDatabase(ctx context.Context, req *agentv1.BackupDatabaseRequest) (*agentv1.BackupDatabaseResponse, error) {
	archivePath, digest, err := drivers.BackupDatabase(ctx, req.GetUsername(), req.GetEngine(), req.GetDbName(), req.GetOutputDir())
	if err != nil {
		return &agentv1.BackupDatabaseResponse{
			Success: false,
		}, err
	}
	return &agentv1.BackupDatabaseResponse{
		Success:     true,
		ArchivePath: archivePath,
		Sha256:      digest,
	}, nil
}

func (s *AgentServer) SyncDNSZone(ctx context.Context, req *agentv1.SyncDNSZoneRequest) (*agentv1.OperationResponse, error) {
	var records []drivers.DNSRecord
	for _, r := range req.GetRecords() {
		records = append(records, drivers.DNSRecord{
			Name:     r.GetName(),
			Type:     r.GetType(),
			Content:  r.GetContent(),
			TTL:      r.GetTtl(),
			Priority: r.GetPriority(),
		})
	}
	err := drivers.SyncDNSZone(ctx, req.GetZoneName(), records, req.GetEnableDnssec())
	return opResponse(err, fmt.Sprintf("DNS zone %s synced to PowerDNS", req.GetZoneName()))
}

func (s *AgentServer) RegisterDNSClusterNode(ctx context.Context, req *agentv1.RegisterDNSClusterNodeRequest) (*agentv1.OperationResponse, error) {
	err := drivers.RegisterDNSClusterNode(ctx, req.GetClusterName(), req.GetHostname(), req.GetApiEndpoint(), req.GetRole(), req.GetApiToken())
	return opResponse(err, fmt.Sprintf("DNS cluster node %s registered in cluster %s", req.GetHostname(), req.GetClusterName()))
}

func (s *AgentServer) SyncDNSZoneToCluster(ctx context.Context, req *agentv1.SyncDNSZoneToClusterRequest) (*agentv1.OperationResponse, error) {
	err := drivers.SyncDNSZoneToCluster(ctx, req.GetZoneName(), req.GetClusterName(), req.GetForce())
	return opResponse(err, fmt.Sprintf("DNS zone %s synced to cluster %s", req.GetZoneName(), req.GetClusterName()))
}

func (s *AgentServer) GetDNSClusterStatus(ctx context.Context, req *agentv1.GetDNSClusterStatusRequest) (*agentv1.GetDNSClusterStatusResponse, error) {
	nodes, err := drivers.GetDNSClusterStatus(ctx, req.GetClusterName())
	if err != nil {
		return &agentv1.GetDNSClusterStatusResponse{Success: false}, err
	}
	var resNodes []*agentv1.DNSClusterNodeStatus
	for _, n := range nodes {
		resNodes = append(resNodes, &agentv1.DNSClusterNodeStatus{
			Hostname:     n.Hostname,
			Role:         n.Role,
			Status:       n.Status,
			LastSeenUnix: n.LastSeenUnix,
			ZonesSynced:  n.ZonesSynced,
		})
	}
	return &agentv1.GetDNSClusterStatusResponse{
		Success: true,
		Nodes:   resNodes,
	}, nil
}

func (s *AgentServer) ProvisionMailDomain(ctx context.Context, req *agentv1.ProvisionMailDomainRequest) (*agentv1.ProvisionMailDomainResponse, error) {
	res, err := drivers.ProvisionMailDomain(ctx, req.GetDomainName(), req.GetDkimSelector(), req.GetMtaStsMode())
	if err != nil {
		return &agentv1.ProvisionMailDomainResponse{Success: false}, err
	}
	return &agentv1.ProvisionMailDomainResponse{
		Success:              true,
		DkimRsaDnsRecord:     res.DKIMRSADNS,
		DkimEd25519DnsRecord: res.DKIMEd25519DNS,
	}, nil
}

func (s *AgentServer) SyncMailbox(ctx context.Context, req *agentv1.SyncMailboxRequest) (*agentv1.OperationResponse, error) {
	err := drivers.SyncMailbox(ctx, req.GetDomainName(), req.GetLocalPart(), req.GetPasswordHash(), req.GetQuotaMb(), req.GetActive())
	return opResponse(err, fmt.Sprintf("Mailbox %s@%s synced", req.GetLocalPart(), req.GetDomainName()))
}

func (s *AgentServer) SyncMailAlias(ctx context.Context, req *agentv1.SyncMailAliasRequest) (*agentv1.OperationResponse, error) {
	err := drivers.SyncMailAlias(ctx, req.GetDomainName(), req.GetSourceLocalPart(), req.GetDestinationEmails(), req.GetActive())
	return opResponse(err, fmt.Sprintf("Mail alias %s@%s synced", req.GetSourceLocalPart(), req.GetDomainName()))
}

func (s *AgentServer) CheckRBLStatus(ctx context.Context, req *agentv1.CheckRBLStatusRequest) (*agentv1.CheckRBLStatusResponse, error) {
	clean, results, err := drivers.CheckRBLStatus(ctx, req.GetServerIp())
	if err != nil {
		return &agentv1.CheckRBLStatusResponse{
			ServerIp: req.GetServerIp(),
			Clean:    false,
		}, err
	}
	var resList []*agentv1.RBLListing
	for _, r := range results {
		resList = append(resList, &agentv1.RBLListing{
			Provider: r.Provider,
			Listed:   r.Listed,
			Detail:   r.Detail,
		})
	}
	return &agentv1.CheckRBLStatusResponse{
		ServerIp: req.GetServerIp(),
		Clean:    clean,
		Results:  resList,
	}, nil
}

func (s *AgentServer) ApplyFirewallRule(ctx context.Context, req *agentv1.ApplyFirewallRuleRequest) (*agentv1.OperationResponse, error) {
	rule := drivers.FirewallRule{
		Action:   req.GetAction(),
		CIDRorIP: req.GetCidrOrIp(),
		Port:     req.GetPort(),
		Protocol: req.GetProtocol(),
		Comment:  req.GetComment(),
	}
	err := drivers.ApplyFirewallRule(ctx, rule)
	return opResponse(err, fmt.Sprintf("Firewall rule (%s %s:%d) applied", req.GetAction(), req.GetCidrOrIp(), req.GetPort()))
}

func (s *AgentServer) InstallAppArchive(ctx context.Context, req *agentv1.InstallAppArchiveRequest) (*agentv1.OperationResponse, error) {
	err := fileops.InstallAppArchive(ctx, req.GetUsername(), req.GetArchiveUrl(), req.GetExpectedSha256(), req.GetTargetDir(), req.GetStripComponents())
	return opResponse(err, fmt.Sprintf("App archive extracted into %s", req.GetTargetDir()))
}

func (s *AgentServer) CloneAppStaging(ctx context.Context, req *agentv1.CloneAppStagingRequest) (*agentv1.OperationResponse, error) {
	err := fileops.CloneAppStaging(ctx, req.GetUsername(), req.GetSourceDir(), req.GetTargetDir(), req.GetSourceDb(), req.GetTargetDb(), req.GetOldUrl(), req.GetNewUrl())
	return opResponse(err, fmt.Sprintf("App cloned to staging %s", req.GetTargetDir()))
}

func (s *AgentServer) DeployGitRepository(ctx context.Context, req *agentv1.DeployGitRepositoryRequest) (*agentv1.OperationResponse, error) {
	err := drivers.DeployGitRepository(ctx, req.GetUsername(), req.GetRepoUrl(), req.GetBranch(), req.GetDeployDir())
	return opResponse(err, fmt.Sprintf("Git repository deployed to %s", req.GetDeployDir()))
}

func (s *AgentServer) CreateBackup(ctx context.Context, req *agentv1.CreateBackupRequest) (*agentv1.OperationResponse, error) {
	err := drivers.CreateBackup(ctx, req.GetUsername(), req.GetDestinationId(), req.GetIncludeDatabases(), req.GetIncludeMail())
	return opResponse(err, fmt.Sprintf("Backup initiated for user %s", req.GetUsername()))
}

func (s *AgentServer) RestoreBackup(ctx context.Context, req *agentv1.RestoreBackupRequest) (*agentv1.OperationResponse, error) {
	err := drivers.RestoreBackup(ctx, req.GetUsername(), req.GetSnapshotId(), req.GetDestinationId())
	return opResponse(err, fmt.Sprintf("Snapshot %s restored for user %s", req.GetSnapshotId(), req.GetUsername()))
}

func (s *AgentServer) TestRestoreBackup(ctx context.Context, req *agentv1.TestRestoreBackupRequest) (*agentv1.OperationResponse, error) {
	err := drivers.TestRestoreBackup(ctx, req.GetSnapshotId(), req.GetDestinationId())
	return opResponse(err, fmt.Sprintf("Snapshot %s verified via automated test-restore", req.GetSnapshotId()))
}

func (s *AgentServer) ManageAppRuntime(ctx context.Context, req *agentv1.ManageAppRuntimeRequest) (*agentv1.OperationResponse, error) {
	cfg := drivers.AppRuntimeConfig{
		Username:    req.GetUsername(),
		RuntimeType: req.GetRuntimeType(),
		Action:      req.GetAction(),
		AppName:     req.GetAppName(),
		WorkingDir:  req.GetWorkingDir(),
		Entrypoint:  req.GetEntrypoint(),
		Port:        req.GetPort(),
	}
	err := drivers.ManageAppRuntime(ctx, cfg)
	return opResponse(err, fmt.Sprintf("App runtime %s %s action completed", req.GetAppName(), req.GetAction()))
}

func (s *AgentServer) SyncCronJob(ctx context.Context, req *agentv1.SyncCronJobRequest) (*agentv1.OperationResponse, error) {
	err := drivers.SyncCronJob(ctx, req.GetUsername(), req.GetAction(), req.GetSchedule(), req.GetCommand())
	return opResponse(err, fmt.Sprintf("Cron job %s for user %s", req.GetAction(), req.GetUsername()))
}

func (s *AgentServer) ProvisionFTPAccount(ctx context.Context, req *agentv1.ProvisionFTPAccountRequest) (*agentv1.OperationResponse, error) {
	err := drivers.ProvisionFTPAccount(ctx, req.GetUsername(), req.GetFtpUser(), req.GetPasswordHash(), req.GetHomeDir())
	return opResponse(err, fmt.Sprintf("FTP account %s provisioned", req.GetFtpUser()))
}

func (s *AgentServer) GetSystemTelemetry(ctx context.Context, req *agentv1.GetSystemTelemetryRequest) (*agentv1.GetSystemTelemetryResponse, error) {
	telemetry, err := drivers.GetSystemTelemetry(ctx, req.GetUsername())
	if err != nil {
		return &agentv1.GetSystemTelemetryResponse{}, err
	}
	return &agentv1.GetSystemTelemetryResponse{
		CpuPercent:         telemetry.CPUPercent,
		RamUsedMb:          telemetry.RAMUsedMB,
		RamTotalMb:         telemetry.RAMTotalMB,
		DiskUsedMb:         telemetry.DiskUsedMB,
		DiskTotalMb:        telemetry.DiskTotalMB,
		WafBlocked_24H:     telemetry.WAFBlocked24h,
		CrowdsecActiveBans: telemetry.CrowdSecActiveBans,
	}, nil
}

func (s *AgentServer) ReadServiceConfig(ctx context.Context, req *agentv1.ReadServiceConfigRequest) (*agentv1.ReadServiceConfigResponse, error) {
	filePath, content, err := drivers.ReadServiceConfig(req.GetService(), req.GetSubPath())
	if err != nil {
		return &agentv1.ReadServiceConfigResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}
	return &agentv1.ReadServiceConfigResponse{
		Success:  true,
		Message:  "Configuration loaded successfully",
		FilePath: filePath,
		Content:  content,
	}, nil
}

func (s *AgentServer) WriteServiceConfig(ctx context.Context, req *agentv1.WriteServiceConfigRequest) (*agentv1.WriteServiceConfigResponse, error) {
	backupPath, err := drivers.WriteServiceConfig(ctx, req.GetService(), req.GetSubPath(), req.GetContent())
	if err != nil {
		return &agentv1.WriteServiceConfigResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}
	return &agentv1.WriteServiceConfigResponse{
		Success:    true,
		Message:    "Configuration validated, written atomically, and service reloaded",
		BackupPath: backupPath,
	}, nil
}

func (s *AgentServer) RollbackServiceConfig(ctx context.Context, req *agentv1.RollbackServiceConfigRequest) (*agentv1.OperationResponse, error) {
	err := drivers.RollbackServiceConfig(ctx, req.GetService(), req.GetSubPath(), req.GetBackupTimestamp())
	return opResponse(err, fmt.Sprintf("Service %s configuration rolled back to backup %s", req.GetService(), req.GetBackupTimestamp()))
}

func (s *AgentServer) ListServiceConfigBackups(ctx context.Context, req *agentv1.ListServiceConfigBackupsRequest) (*agentv1.ListServiceConfigBackupsResponse, error) {
	backups, err := drivers.ListServiceConfigBackups(req.GetService(), req.GetSubPath())
	if err != nil {
		return &agentv1.ListServiceConfigBackupsResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	var pbBackups []*agentv1.ConfigBackupEntry
	for _, b := range backups {
		pbBackups = append(pbBackups, &agentv1.ConfigBackupEntry{
			BackupFile: b.BackupFile,
			Timestamp:  b.Timestamp,
			SizeBytes:  b.SizeBytes,
		})
	}

	return &agentv1.ListServiceConfigBackupsResponse{
		Success: true,
		Message: fmt.Sprintf("Found %d backups", len(pbBackups)),
		Backups: pbBackups,
	}, nil
}
