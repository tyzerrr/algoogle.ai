import { CheckCircle2, ClipboardList, Gauge, UserCheck, XCircle } from "lucide-react";
import type { ReviewResponse } from "@/lib/types";

function ListBlock({ title, items }: { title: string; items: string[] }) {
  return (
    <section>
      <h3>{title}</h3>
      {items.length === 0 ? (
        <p className="muted">なし</p>
      ) : (
        <ul>
          {items.map((item) => (
            <li key={item}>{item}</li>
          ))}
        </ul>
      )}
    </section>
  );
}

export default function ReviewPanel({ review }: { review?: ReviewResponse }) {
  if (!review) {
    return <div className="empty">Submitすると、正誤だけでなくフォローアップと計算量の詰めが表示されます。</div>;
  }

  const good = review.is_correct ? ["ローカルテストとレビュー上は正しさが確認できています。"] : [];
  const scorecard = review.scorecard ?? [];
  const miniRounds = review.mini_rounds ?? [];
  const weaknesses = review.detected_weaknesses ?? [];
  const hireClass = review.hire_recommendation.includes("No")
    ? "status statusNeeds"
    : review.hire_recommendation.includes("Lean")
      ? "status statusTodo"
      : "status";

  return (
    <div className="reviewBlock">
      <div className="resultMeta">
        <span className={review.is_correct ? "status" : "status statusNeeds"}>
          {review.is_correct ? <CheckCircle2 size={16} /> : <XCircle size={16} />}
          {review.is_correct ? "正しそう" : "修正ポイントあり"}
        </span>
        <span className={hireClass}>
          <UserCheck size={16} />
          {review.hire_recommendation}
        </span>
        <span className="tag">Time {review.complexity.time}</span>
        <span className="tag">Space {review.complexity.space}</span>
      </div>
      <p>{review.summary}</p>
      <section className="readiness">
        <h3>Google readiness</h3>
        <p>{review.google_readiness}</p>
      </section>
      <section>
        <h3>
          <Gauge size={17} /> Interview scorecard
        </h3>
        {scorecard.length === 0 ? (
          <p className="muted">scorecardはまだありません。</p>
        ) : (
          <div className="scorecardGrid">
            {scorecard.map((item) => (
              <article className="scorecardItem" key={`${item.area}-${item.signal}`}>
                <div className="scorecardTop">
                  <strong>{item.area}</strong>
                  <span className="scoreNumber">{item.score}/4</span>
                </div>
                <p>{item.signal}</p>
                <p className="muted">{item.evidence}</p>
                <p className="actionText">{item.action}</p>
              </article>
            ))}
          </div>
        )}
      </section>
      <ListBlock title="Shadow evaluator notes" items={review.shadow_notes ?? []} />
      <ListBlock title="良かった点" items={good.concat(review.readability_feedback)} />
      <ListBlock title="間違っていた点" items={review.bugs.concat(review.edge_cases)} />
      <ListBlock title="計算量で詰める質問" items={review.complexity_questions} />
      <ListBlock title="求める複数解法" items={review.alternative_approaches} />
      <ListBlock title="フォローアップ質問" items={review.follow_up_questions} />
      <section>
        <h3>
          <ClipboardList size={17} /> Mini rounds
        </h3>
        {miniRounds.length === 0 ? (
          <p className="muted">追加ラウンドはまだありません。</p>
        ) : (
          <div className="miniRoundGrid">
            {miniRounds.map((round) => (
              <article className="miniRound" key={`${round.kind}-${round.question}`}>
                <span className="tag">{round.kind}</span>
                <p>{round.question}</p>
                <p className="muted">{round.bar}</p>
              </article>
            ))}
          </div>
        )}
      </section>
      <section>
        <h3>検出された弱点</h3>
        {weaknesses.length === 0 ? (
          <p className="muted">弱点signalはまだありません。</p>
        ) : (
          <div className="weaknessGrid">
            {weaknesses.map((weakness) => (
              <article className="weaknessCard" key={`${weakness.category}-${weakness.signal}`}>
                <div className="scorecardTop">
                  <strong>{weakness.category}</strong>
                  <span className="tag">severity {weakness.severity}/5</span>
                </div>
                <p>{weakness.signal}</p>
                <p className="muted">{weakness.evidence}</p>
                <p className="actionText">{weakness.drill}</p>
              </article>
            ))}
          </div>
        )}
      </section>
      <ListBlock title="覚えておくべきこと" items={review.mistakes_to_remember} />
      <ListBlock title="面接で使える説明フレーズ" items={review.interview_feedback} />
      <ListBlock title="次のディスカッション計画" items={review.discussion_plan} />
      <section>
        <h3>次の復習</h3>
        <p>{review.next_review_recommendation}</p>
      </section>
    </div>
  );
}
