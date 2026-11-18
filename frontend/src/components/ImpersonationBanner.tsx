import React from "react";
import { useTranslation } from "react-i18next";
import { AlertTriangle, LogOut } from "lucide-react";

interface ImpersonationBannerProps {
  username: string;
  onExit: () => void;
}

export const ImpersonationBanner: React.FC<ImpersonationBannerProps> = ({
  username,
  onExit,
}) => {
  const { t } = useTranslation();

  return (
    <div className="bg-status-warning/15 border-b border-status-warning/30 px-4 py-2 text-xs text-status-warning flex items-center justify-between">
      <div className="flex items-center gap-2">
        <AlertTriangle className="h-4 w-4 shrink-0" />
        <span>
          {t("impersonation.banner", {
            username,
            defaultValue: `You are currently impersonating tenant account ${username}. All actions are recorded in the immutable audit log.`,
          })}
        </span>
      </div>
      <button
        onClick={onExit}
        className="flex items-center gap-1.5 bg-status-warning text-white px-2.5 py-1 rounded font-semibold text-[11px] hover:bg-status-warning/90 transition-colors"
      >
        <LogOut className="h-3 w-3" />
        <span>{t("impersonation.exit", "Exit Impersonation")}</span>
      </button>
    </div>
  );
};
