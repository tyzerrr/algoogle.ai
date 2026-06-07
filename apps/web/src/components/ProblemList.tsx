"use client";

import Link from "next/link";
import { useMemo, useState } from "react";
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

function masteryLabel(problem: ProblemListItem) {
  const labels = {
    not_started: "未着手",
    in_progress: "演習中",
    needs_review: "要復習",
    first_try_clean: "一発OK",
    first_try_with_followups: "フォローアップ込みOK",
    solved_after_retry: "再挑戦OK",
  };
  return labels[problem.mastery_status] ?? statusLabel(problem.status);
}

export default function ProblemList({ problems }: { problems: ProblemListItem[] }) {
  const pageSize = 10;
  const [page, setPage] = useState(1);
  const pageCount = Math.max(1, Math.ceil(problems.length / pageSize));
  const safePage = Math.min(page, pageCount);
  const visibleProblems = useMemo(
    () => problems.slice((safePage - 1) * pageSize, safePage * pageSize),
    [problems, safePage],
  );

  if (problems.length === 0) {
    return <div className="empty">問題がまだ登録されていません。</div>;
  }

  return (
    <>
      <div className="listToolbar">
        <p className="muted">
          Arai60 {problems.length}問中 {(safePage - 1) * pageSize + 1}-
          {Math.min(safePage * pageSize, problems.length)}問目
        </p>
        <div className="pager">
          <button
            className="secondaryButton"
            type="button"
            onClick={() => setPage((current) => Math.max(1, current - 1))}
            disabled={safePage === 1}
          >
            前へ
          </button>
          <span className="tag">
            {safePage} / {pageCount}
          </span>
          <button
            className="secondaryButton"
            type="button"
            onClick={() => setPage((current) => Math.min(pageCount, current + 1))}
            disabled={safePage === pageCount}
          >
            次へ
          </button>
        </div>
      </div>

      <div className="problemGrid">
        {visibleProblems.map((problem) => (
          <Link key={problem.id} href={`/problems/${problem.id}`} className="card problemCard">
            <div className="resultMeta">
              <span className="difficulty">#{problem.order_index}</span>
              <span className="difficulty">{problem.difficulty}</span>
              <span className={statusClass(problem.status)}>{masteryLabel(problem)}</span>
            </div>
            <h3>{problem.title}</h3>
            <p className="muted">{problem.pattern}</p>
            <div className="cardStats">
              <span>attempt {problem.attempt_count}</span>
              <span>follow-up {problem.follow_up_count}</span>
            </div>
            <div className="tagRow">
              {problem.tags.slice(0, 3).map((tag) => (
                <span className="tag" key={tag}>
                  {tag}
                </span>
              ))}
            </div>
          </Link>
        ))}
      </div>
    </>
  );
}
