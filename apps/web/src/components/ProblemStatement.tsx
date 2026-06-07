import type { Problem } from "@/lib/types";

export default function ProblemStatement({ problem }: { problem: Problem }) {
  return (
    <div className="statement">
      <div className="resultMeta">
        {problem.order_index ? <span className="difficulty">#{problem.order_index}</span> : null}
        <span className="difficulty">{problem.difficulty}</span>
        <span className="tag">{problem.pattern}</span>
        {problem.list_name ? <span className="tag">{problem.list_name}</span> : null}
      </div>
      <p>{problem.statement}</p>
      {problem.source_url ? (
        <p>
          <a href={problem.source_url} target="_blank" rel="noreferrer">
            公式問題を開く
          </a>
        </p>
      ) : null}

      <h3>例</h3>
      {problem.examples.map((example, index) => (
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

      <h3>制約</h3>
      <ul className="constraints">
        {problem.constraints.map((constraint) => (
          <li key={constraint}>{constraint}</li>
        ))}
      </ul>
    </div>
  );
}
