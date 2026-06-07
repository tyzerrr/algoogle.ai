import Link from "next/link";
import type { ProblemListItem } from "@/lib/types";

function statusClass(status: ProblemListItem["status"]) {
  if (status === "needs_review") return "status statusNeeds";
  if (status === "not_started") return "status statusTodo";
  return "status";
}

function statusLabel(status: ProblemListItem["status"]) {
  const labels = {
    not_started: "未着手",
    in_progress: "演習中",
    needs_review: "復習",
    solved: "解決済み",
  };
  return labels[status];
}

export default function ProblemList({ problems }: { problems: ProblemListItem[] }) {
  if (problems.length === 0) {
    return <div className="empty">問題がまだ登録されていません。</div>;
  }

  return (
    <div className="problemGrid">
      {problems.map((problem) => (
        <Link key={problem.id} href={`/problems/${problem.id}`} className="card">
          <div className="resultMeta">
            <span className="difficulty">{problem.difficulty}</span>
            <span className={statusClass(problem.status)}>{statusLabel(problem.status)}</span>
          </div>
          <h3>{problem.title}</h3>
          <p className="muted">{problem.pattern}</p>
          <div className="tagRow">
            {problem.tags.map((tag) => (
              <span className="tag" key={tag}>
                {tag}
              </span>
            ))}
          </div>
        </Link>
      ))}
    </div>
  );
}
