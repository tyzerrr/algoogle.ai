import { ExternalLink } from "lucide-react";
import type { OfficialProblemContent, Problem } from "@/lib/types";

export default function ProblemStatement({
  problem,
  officialContent,
  officialLoading,
  officialError,
}: {
  problem: Problem;
  officialContent?: OfficialProblemContent | null;
  officialLoading?: boolean;
  officialError?: string;
}) {
  const statement = officialContent?.statement || problem.statement;
  const examples = officialContent?.examples.length ? officialContent.examples : problem.examples;
  const sourceLabel = officialContent ? `${officialContent.source}本文` : "Arai60要約";

  return (
    <div className="statement">
      <div className="resultMeta">
        {problem.order_index ? <span className="difficulty">#{problem.order_index}</span> : null}
        <span className="difficulty">{problem.difficulty}</span>
        <span className="tag">{problem.pattern}</span>
        {problem.list_name ? <span className="tag">{problem.list_name}</span> : null}
        <span className="tag">{sourceLabel}</span>
      </div>
      {officialLoading ? <p className="muted">本家問題文を取得中...</p> : null}
      {officialError ? <p className="muted">本家問題文を取得できないため、Arai60要約を表示しています。</p> : null}
      <p className="statementText">{statement}</p>
      {problem.source_url ? (
        <p className="sourceLinkRow">
          <a
            className="sourceLink"
            href={problem.source_url}
            target="_blank"
            rel="noreferrer"
            aria-label="LeetCodeの公式問題を新しいタブで開く"
          >
            公式問題を開く
            <ExternalLink size={16} />
          </a>
        </p>
      ) : null}

      <h3>例</h3>
      {examples.map((example, index) => (
        <div className="example" key={`${example.input}-${index}`}>
          <p>
            <strong>Input:</strong> <code>{example.input}</code>
          </p>
          <p>
            <strong>Output:</strong> <code>{example.output}</code>
          </p>
          {example.explanation ? <p className="muted">{example.explanation}</p> : null}
        </div>
      ))}
    </div>
  );
}
