"use client";

import { useEffect, useRef, type MouseEvent } from "react";

export default function ConfirmDialog({
  open,
  title,
  description,
  confirmLabel,
  cancelLabel = "キャンセル",
  danger = false,
  onConfirm,
  onCancel,
}: {
  open: boolean;
  title: string;
  description?: string;
  confirmLabel: string;
  cancelLabel?: string;
  danger?: boolean;
  onConfirm: () => void;
  onCancel: () => void;
}) {
  const cancelRef = useRef<HTMLButtonElement | null>(null);

  useEffect(() => {
    if (!open) return undefined;
    cancelRef.current?.focus();
    function handleKey(event: KeyboardEvent) {
      if (event.key === "Escape") {
        event.preventDefault();
        onCancel();
      }
    }
    document.addEventListener("keydown", handleKey);
    return () => document.removeEventListener("keydown", handleKey);
  }, [open, onCancel]);

  if (!open) return null;

  function handleOverlayClick(event: MouseEvent<HTMLDivElement>) {
    if (event.target === event.currentTarget) onCancel();
  }

  return (
    <div className="dialogOverlay" onMouseDown={handleOverlayClick}>
      <div className="dialogCard" role="alertdialog" aria-modal="true" aria-label={title}>
        <h2 className="dialogTitle">{title}</h2>
        {description ? <p className="dialogDescription">{description}</p> : null}
        <div className="dialogActions">
          <button className="secondaryButton" type="button" ref={cancelRef} onClick={onCancel}>
            {cancelLabel}
          </button>
          <button
            className={danger ? "button dangerButton" : "button"}
            type="button"
            onClick={onConfirm}
          >
            {confirmLabel}
          </button>
        </div>
      </div>
    </div>
  );
}
