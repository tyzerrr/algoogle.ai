"use client";

import { useEffect, useState } from "react";
import ProblemList from "@/components/ProblemList";
import ErrorNotice from "@/components/ui/ErrorNotice";
import { SkeletonBlock } from "@/components/ui/Skeleton";
import { api } from "@/lib/api";
import type { ProblemListItem } from "@/lib/types";

export default function ProblemsPage() {
  const [problems, setProblems] = useState<ProblemListItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [reloadTick, setReloadTick] = useState(0);

  useEffect(() => {
    setLoading(true);
    setError("");
    api
      .problems()
      .then(setProblems)
      .catch((err: Error) => setError(err.message))
      .finally(() => setLoading(false));
  }, [reloadTick]);

  return (
    <main className="page">
      <div className="pageHeader">
        <div>
          <p className="eyebrow">Problem set</p>
          <h1>Arai60</h1>
          <p className="muted">
            新井康平氏のArai60を、解け方とフォローアップ履歴まで含めて管理します。
          </p>
        </div>
      </div>
      {error ? (
        <ErrorNotice
          message="問題一覧を読み込めませんでした"
          onRetry={() => setReloadTick((tick) => tick + 1)}
          retryLabel="再取得"
        />
      ) : loading ? (
        <SkeletonBlock lines={6} />
      ) : (
        <ProblemList problems={problems} />
      )}
    </main>
  );
}
