"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { ArrowRight, ClipboardCheck, Target } from "lucide-react";
import { api } from "@/lib/api";
import type { DailyResponse } from "@/lib/types";

export default function HomePage() {
  const [daily, setDaily] = useState<DailyResponse | null>(null);
  const [error, setError] = useState("");

  useEffect(() => {
    api
      .daily()
      .then(setDaily)
      .catch((err: Error) => setError(err.message));
  }, []);

  return (
    <main className="page">
      <div className="pageHeader">
        <div>
          <p className="eyebrow">Today&apos;s interview</p>
          <h1>今日の1問を、面接官と詰め切る。</h1>
          <p className="muted">
            解法暗記ではなく、制約、探索方針、エッジケース、計算量を会話で固める練習場です。
          </p>
        </div>
        <div className="buttonRow">
          <Link className="secondaryButton" href="/problems">
            <ClipboardCheck size={18} />
            問題一覧
          </Link>
          <Link className="secondaryButton" href="/review">
            <Target size={18} />
            復習
          </Link>
        </div>
      </div>

      {error ? <div className="error">{error}</div> : null}

      {!daily && !error ? (
        <div className="empty">今日の問題を選んでいます...</div>
      ) : daily ? (
        <div className="dailyLayout">
          <section className="band">
            <div className="resultMeta">
              <span className="difficulty">{daily.problem.difficulty}</span>
              <span className="tag">{daily.problem.pattern}</span>
            </div>
            <h2>{daily.problem.title}</h2>
            <p>{daily.reason}</p>
            <div className="tagRow">
              {daily.problem.tags.map((tag) => (
                <span className="tag" key={tag}>
                  {tag}
                </span>
              ))}
            </div>
            <div className="buttonRow" style={{ marginTop: 18 }}>
              <Link className="button" href={daily.recommended_url}>
                開始
                <ArrowRight size={18} />
              </Link>
            </div>
          </section>

          <section className="band">
            <h2>今日意識すること</h2>
            <ul className="constraints">
              {daily.focus.map((item) => (
                <li key={item}>{item}</li>
              ))}
            </ul>
          </section>
        </div>
      ) : null}
    </main>
  );
}
