import { ExternalLink } from "lucide-react";
import type { OfficialProblemContent, Problem } from "@/lib/types";
import { resolveStatementView } from "@/lib/statementView";
import ErrorNotice from "@/components/ui/ErrorNotice";
import { SkeletonBlock } from "@/components/ui/Skeleton";

export default function ProblemStatement({
  problem,
  officialContent,
  officialLoading,
  officialError,
  language,
  onLanguageChange,
  onRetryOfficial,
}: {
  problem: Problem;
  officialContent?: OfficialProblemContent | null;
  officialLoading?: boolean;
  officialError?: string;
  language: "ja" | "en";
  onLanguageChange: (language: "ja" | "en") => void;
  onRetryOfficial?: () => void;
}) {
  const view = resolveStatementView({
    problem,
    official: officialContent,
    officialLoading,
    officialError,
    language,
  });

  return (
    <div className="statement">
      <div className="statementTop">
        <div className="resultMeta">
          {problem.order_index ? <span className="difficulty">#{problem.order_index}</span> : null}
          <span className="difficulty">{problem.difficulty}</span>
          <span className="tag">{problem.pattern}</span>
          {problem.list_name ? <span className="tag">{problem.list_name}</span> : null}
          <span className="tag">{view.sourceBadge}</span>
        </div>
        {view.showLanguageToggle ? (
          <div className="segmented segmentedTwo statementLanguage" aria-label="問題文の言語">
            <button
              className={`segmentButton ${language === "ja" ? "segmentButtonActive" : ""}`}
              type="button"
              onClick={() => onLanguageChange("ja")}
            >
              日本語
            </button>
            <button
              className={`segmentButton ${language === "en" ? "segmentButtonActive" : ""}`}
              type="button"
              onClick={() => onLanguageChange("en")}
            >
              English
            </button>
          </div>
        ) : null}
      </div>

      {view.mode === "official-fallback" ? (
        <ErrorNotice
          message="LeetCodeから問題文を取得できませんでした"
          onRetry={onRetryOfficial}
          retryLabel="再取得"
        />
      ) : null}

      {view.mode === "official-loading" ? (
        <SkeletonBlock lines={5} />
      ) : (
        <p className="statementText">{view.statement}</p>
      )}

      {view.images.length ? (
        <div className="officialFigures">
          {view.images.map((image) => (
            <figure className="officialFigure" key={image.url}>
              <img
                src={image.url}
                alt={image.alt || `${problem.title} diagram`}
                loading="lazy"
                referrerPolicy="no-referrer"
              />
              {image.alt ? <figcaption>{image.alt}</figcaption> : null}
            </figure>
          ))}
        </div>
      ) : null}

      {view.constraints.length ? (
        <>
          <h3>制約</h3>
          <ul className="constraints">
            {view.constraints.map((constraint, index) => (
              <li key={`${constraint}-${index}`}>{constraint}</li>
            ))}
          </ul>
        </>
      ) : null}

      {view.examples.length ? (
        <>
          <h3>例</h3>
          {view.examples.map((example, index) => (
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
        </>
      ) : null}

      {view.guidance.length ? (
        <>
          <h3>進め方</h3>
          <ul className="constraints guidanceList">
            {view.guidance.map((item, index) => (
              <li key={`${item}-${index}`}>{item}</li>
            ))}
          </ul>
        </>
      ) : null}

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
    </div>
  );
}
