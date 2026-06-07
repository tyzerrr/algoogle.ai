import { CheckCircle2, XCircle } from "lucide-react";
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

  return (
    <div className="reviewBlock">
      <div className="resultMeta">
        <span className={review.is_correct ? "status" : "status statusNeeds"}>
          {review.is_correct ? <CheckCircle2 size={16} /> : <XCircle size={16} />}
          {review.is_correct ? "正しそう" : "修正ポイントあり"}
        </span>
        <span className="tag">Time {review.complexity.time}</span>
        <span className="tag">Space {review.complexity.space}</span>
      </div>
      <p>{review.summary}</p>
      <section className="readiness">
        <h3>Google readiness</h3>
        <p>{review.google_readiness}</p>
      </section>
      <ListBlock title="良かった点" items={good.concat(review.readability_feedback)} />
      <ListBlock title="間違っていた点" items={review.bugs.concat(review.edge_cases)} />
      <ListBlock title="計算量で詰める質問" items={review.complexity_questions} />
      <ListBlock title="求める複数解法" items={review.alternative_approaches} />
      <ListBlock title="フォローアップ質問" items={review.follow_up_questions} />
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
