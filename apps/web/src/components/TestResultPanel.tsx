import { CheckCircle2, XCircle, Timer } from "lucide-react";
import type { RunResult } from "@/lib/types";

function renderValue(value: unknown) {
  if (value === undefined) return "";
  return JSON.stringify(value);
}

export default function TestResultPanel({ result }: { result?: RunResult }) {
  if (!result) {
    return <div className="empty">テストを実行すると結果がここに表示されます。</div>;
  }
  const notConfigured = result.status === "not_configured";

  return (
    <div className="reviewBlock">
      <div className="resultMeta">
        <span className={result.passed || notConfigured ? "status" : "status statusNeeds"}>
          {result.passed || notConfigured ? <CheckCircle2 size={16} /> : <XCircle size={16} />}
          {notConfigured ? "テスト未登録" : result.passed ? "全テスト通過" : "要確認"}
        </span>
        <span className="tag">
          <Timer size={15} />
          {result.duration_ms}ms
        </span>
      </div>

      {result.error ? <div className={notConfigured ? "empty" : "error"}>{result.error}</div> : null}

      <div className="resultList">
        {result.results.map((item) => (
          <div
            className={`resultCase ${
              item.passed ? "resultCasePassed" : "resultCaseFailed"
            }`}
            key={item.name}
          >
            <strong>{item.name}</strong>
            <p className="muted">{item.passed ? "passed" : "failed"}</p>
            {!item.passed ? (
              <>
                <p>
                  Expected: <span className="resultValue">{renderValue(item.expected)}</span>
                </p>
                <p>
                  Actual: <span className="resultValue">{renderValue(item.actual)}</span>
                </p>
                {item.error ? <pre className="error">{item.error}</pre> : null}
              </>
            ) : null}
          </div>
        ))}
      </div>

      {result.stderr ? <pre className="error">{result.stderr}</pre> : null}
    </div>
  );
}
