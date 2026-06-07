import type { Problem } from "@/lib/types";

export default function ProblemStatement({ problem }: { problem: Problem }) {
  return (
    <div className="statement">
      <div className="resultMeta">
        <span className="difficulty">{problem.difficulty}</span>
        <span className="tag">{problem.pattern}</span>
      </div>
      <p>{problem.statement}</p>

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
