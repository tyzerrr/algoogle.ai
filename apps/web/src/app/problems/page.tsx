"use client";

import { useEffect, useState } from "react";
import ProblemList from "@/components/ProblemList";
import { api } from "@/lib/api";
import type { ProblemListItem } from "@/lib/types";

export default function ProblemsPage() {
  const [problems, setProblems] = useState<ProblemListItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    api
      .problems()
      .then(setProblems)
      .catch((err: Error) => setError(err.message))
      .finally(() => setLoading(false));
  }, []);

  return (
    <main className="page">
      <div className="pageHeader">
        <div>
          <p className="eyebrow">Problem set</p>
          <h1>問題一覧</h1>
          <p className="muted">典型パターンごとに、面接で説明し切れる状態を作ります。</p>
        </div>
      </div>
      {error ? <div className="error">{error}</div> : null}
      {loading ? <div className="empty">問題を読み込んでいます...</div> : <ProblemList problems={problems} />}
    </main>
  );
}
