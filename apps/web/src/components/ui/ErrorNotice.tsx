import { AlertTriangle, RefreshCw } from "lucide-react";

export default function ErrorNotice({
  message,
  onRetry,
  retryLabel = "再試行",
}: {
  message: string;
  onRetry?: () => void;
  retryLabel?: string;
}) {
  return (
    <div className="errorNotice" role="alert">
      <span className="errorNoticeText">
        <AlertTriangle size={17} aria-hidden="true" />
        {message}
      </span>
      {onRetry ? (
        <button className="secondaryButton" type="button" onClick={onRetry}>
          <RefreshCw size={16} aria-hidden="true" />
          {retryLabel}
        </button>
      ) : null}
    </div>
  );
}
