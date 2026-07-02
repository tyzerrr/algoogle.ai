"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { api } from "@/lib/api";
import ErrorNotice from "@/components/ui/ErrorNotice";
import { SkeletonBlock } from "@/components/ui/Skeleton";
import type { ReviewDashboard } from "@/lib/types";

export default function ReviewPage() {
  const [dashboard, setDashboard] = useState<ReviewDashboard | null>(null);
  const [error, setError] = useState("");
  const [reloadTick, setReloadTick] = useState(0);

  useEffect(() => {
    setError("");
    api
      .reviewDashboard()
      .then(setDashboard)
      .catch((err: Error) => setError(err.message));
  }, [reloadTick]);

  return (
    <main className="page">
      <div className="pageHeader">
        <div>
          <p className="eyebrow">Review loop</p>
          <h1>復習</h1>
          <p className="muted">提出履歴と覚えておくべきミスを、次の演習につなげます。</p>
        </div>
      </div>

      {error ? (
        <ErrorNotice
          message="復習データを読み込めませんでした"
          onRetry={() => setReloadTick((tick) => tick + 1)}
          retryLabel="再取得"
        />
      ) : null}
      {!dashboard && !error ? <SkeletonBlock lines={6} /> : null}

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
            <h2>最近のフォローアップ質問</h2>
            {dashboard.recent_follow_ups.length === 0 ? (
              <p className="muted">面接官やSubmit後レビューで出た質問がここに残ります。</p>
            ) : (
              <ul className="constraints">
                {dashboard.recent_follow_ups.map((followUp) => (
                  <li key={followUp.id}>
                    <strong>{followUp.source}:</strong> {followUp.question}
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
            <h2>弱点グラフ</h2>
            {dashboard.weakness_graph.length === 0 ? (
              <p className="muted">AIレビューで検出した弱点signalがここに蓄積されます。</p>
            ) : (
              <div className="weaknessGrid">
                {dashboard.weakness_graph.map((weakness) => (
                  <article className="weaknessCard" key={weakness.category}>
                    <div className="scorecardTop">
                      <strong>{weakness.category}</strong>
                      <span className="tag">{weakness.count} signals</span>
                    </div>
                    <div className="barMeter" aria-label={`${weakness.category} severity`}>
                      <span
                        className="barFill"
                        style={{ width: `${Math.min(100, weakness.average_severity * 20)}%` }}
                      />
                    </div>
                    <p className="muted">平均 severity {weakness.average_severity}/5</p>
                    <p className="actionText">{weakness.suggested_drill}</p>
                    <p className="muted">{new Date(weakness.last_seen_at).toLocaleString()}</p>
                  </article>
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
