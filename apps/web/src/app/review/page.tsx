"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { api } from "@/lib/api";
import type { ReviewDashboard } from "@/lib/types";

export default function ReviewPage() {
  const [dashboard, setDashboard] = useState<ReviewDashboard | null>(null);
  const [error, setError] = useState("");

  useEffect(() => {
    api
      .reviewDashboard()
      .then(setDashboard)
      .catch((err: Error) => setError(err.message));
  }, []);

  return (
    <main className="page">
      <div className="pageHeader">
        <div>
          <p className="eyebrow">Review loop</p>
          <h1>復習</h1>
          <p className="muted">提出履歴と覚えておくべきミスを、次の演習につなげます。</p>
        </div>
      </div>

      {error ? <div className="error">{error}</div> : null}
      {!dashboard && !error ? <div className="empty">復習データを読み込んでいます...</div> : null}

      {dashboard ? (
        <div className="reviewBlock">
          <section className="band">
            <h2>最近の提出</h2>
            {dashboard.recent_attempts.length === 0 ? (
              <p className="muted">まだ提出はありません。</p>
            ) : (
              <div className="reviewGrid">
                {dashboard.recent_attempts.map((attempt) => (
                  <Link href={`/problems/${attempt.problem_id}`} className="card" key={attempt.id}>
                    <div className="resultMeta">
                      <span className={attempt.status === "passed" ? "status" : "status statusNeeds"}>
                        {attempt.status}
                      </span>
                    </div>
                    <h3>{attempt.problem_title}</h3>
                    <p className="muted">{new Date(attempt.created_at).toLocaleString()}</p>
                  </Link>
                ))}
              </div>
            )}
          </section>

          <section className="band">
            <h2>覚えておくべきミス</h2>
            {dashboard.mistakes_to_remember.length === 0 ? (
              <p className="muted">AIレビュー後にここへ蓄積されます。</p>
            ) : (
              <ul className="constraints">
                {dashboard.mistakes_to_remember.map((note) => (
                  <li key={note.id}>
                    <strong>{note.problem_title}:</strong> {note.note}
                  </li>
                ))}
              </ul>
            )}
          </section>

          <section className="band">
            <h2>苦手パターン</h2>
            {dashboard.weak_patterns.length === 0 ? (
              <p className="muted">失敗した提出が増えると傾向が見えます。</p>
            ) : (
              <div className="tagRow">
                {dashboard.weak_patterns.map((pattern) => (
                  <span className="tag" key={pattern.pattern}>
                    {pattern.pattern} x {pattern.count}
                  </span>
                ))}
              </div>
            )}
          </section>

          <section className="band">
            <h2>今日復習すべき問題</h2>
            {dashboard.problems_for_today.length === 0 ? (
              <p className="muted">今のところ復習候補はありません。</p>
            ) : (
              <div className="reviewGrid">
                {dashboard.problems_for_today.map((problem) => (
                  <Link href={`/problems/${problem.id}`} className="card" key={problem.id}>
                    <h3>{problem.title}</h3>
                    <p className="muted">{problem.pattern}</p>
                  </Link>
                ))}
              </div>
            )}
          </section>
        </div>
      ) : null}
    </main>
  );
}
